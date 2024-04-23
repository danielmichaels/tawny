package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/danielmichaels/tawny/design"
	"github.com/danielmichaels/tawny/gen/identity"
	"github.com/danielmichaels/tawny/internal/auth"
	"github.com/danielmichaels/tawny/internal/logger"
	"github.com/danielmichaels/tawny/internal/ptr"
	"github.com/danielmichaels/tawny/internal/store"
	"github.com/jackc/pgx/v5/pgconn"
	"goa.design/goa/v3/security"
	"math"
	"strings"
)

// identity service example implementation.
// The example methods log the requests and return zero values.
type identitysrvc struct {
	logger *logger.Logger
	db     *store.Queries
}

// NewIdentity returns the identity service implementation.
func NewIdentity(logger *logger.Logger, db *store.Queries) identity.Service {
	return &identitysrvc{logger, db}
}

// APIKeyAuth implements the authorization logic for service "identity" for the
// "api_key" security scheme.
func (s *identitysrvc) APIKeyAuth(ctx context.Context, key string, scheme *security.APIKeyScheme) (context.Context, error) {
	ak := auth.NewApiKey()
	ctx, err := ak.Validate(ctx, key, scheme, s.db)
	if err != nil {
		s.logger.Error().Err(err).Msg("token invalid")
		return ctx, &identity.Unauthorized{Message: "token invalid"}
	}
	return ctx, nil
}

// Create a new user. This will also generate a new team for that user.
func (s *identitysrvc) CreateUser(ctx context.Context, p *identity.CreateUserPayload) (res *identity.UserResult, err error) {
	res = &identity.UserResult{}
	s.logger.Print("identity.createUser")
	return
}

// Retrieve a single user. Can only retrieve users from an associated team.
func (s *identitysrvc) RetrieveUser(ctx context.Context, p *identity.RetrieveUserPayload) (res *identity.UserResult, err error) {
	u, err := s.db.GetUserByID(ctx, p.ID)
	if err != nil {
		return nil, &identity.NotFound{
			Name:    "not found",
			Message: "resource not found",
			Detail:  "resource not found",
		}
	}
	user := &identity.UserResult{
		UserUUID: &u.Uuid,
		Name:     u.Name.String,
		Email:    u.Email.String,
		//Role:      u.,
		//Verified:  nil,
		CreatedAt: ptr.Ptr(u.CreatedAt.Time.String()),
		UpdatedAt: ptr.Ptr(u.UpdatedAt.Time.String()),
	}
	return user, nil
}

// Retrieve all users that this user can see from associated teams.
func (s *identitysrvc) ListUsers(ctx context.Context, p *identity.ListUsersPayload) (res *identity.Users, err error) {
	ut := auth.CtxAuthInfo(ctx)
	ps, pn := design.PaginationQueryParams(p.PageSize, p.PageNumber)
	u, err := s.db.ListUsers(ctx, store.ListUsersParams{
		Token:  ut.User,
		Limit:  ps,
		Offset: pn,
	})
	if err != nil {
		return nil, &identity.NotFound{
			Name:    "not found",
			Message: "resource not found",
			Detail:  "resource not found",
		}
	}
	count, err := s.db.CountUsers(ctx, ut.User)
	if err != nil {
		count = 0
	}
	var users = &identity.Users{}
	for _, user := range u {
		users.Users = append(users.Users, &identity.UserResult{
			Name:      user.Name.String,
			Email:     user.Email.String,
			Role:      string(user.Role),
			CreatedAt: ptr.Ptr(user.CreatedAt.Time.String()),
			UpdatedAt: ptr.Ptr(user.UpdatedAt.Time.String()),
		})
	}
	users.Metadata = CalculateIdentityMetadata(int(count), p.PageNumber, p.PageSize)
	return users, nil
}

// Create a new team
func (s *identitysrvc) CreateTeam(ctx context.Context, p *identity.CreateTeamPayload) (res *identity.Team, err error) {
	ut := auth.CtxAuthInfo(ctx)
	t, err := s.db.CreateTeam(ctx, store.CreateTeamParams{
		TeamName:      p.Team.Name,
		TeamEmail:     p.Team.Email,
		CurrentUserID: ut.User,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr) && pgErr.Code == "P0001":
			return nil, &identity.Unauthorized{Message: "user does not have permission to create team"}
		case errors.As(err, &pgErr) && pgErr.Code == "23505":
			s.logger.Error().Err(err).Msg("error creating team")
			return nil, &identity.BadRequest{
				Name:    "bad request",
				Message: "team name or email already exists",
			}
		case errors.As(err, &pgErr):
			s.logger.Error().Err(err).Interface("pgError", pgErr).Msg("error creating team")
			return nil, &identity.ServerError{
				Name:    "internal server error",
				Message: "an unknown error occurred",
			}
		default:
			s.logger.Error().Err(err).Msg("error creating team")
			return nil, &identity.ServerError{
				Name:    "internal server error",
				Message: "an unknown error occurred",
			}
		}
	}
	team, err := parseTeam(t)
	if err != nil {
		s.logger.Error().Err(err).Msg("error parsing team")
		return nil, &identity.ServerError{
			Name:    "internal server error",
			Message: "an unknown error occurred",
		}
	}

	return &identity.Team{
		TeamUUID:  team.TeamUUID,
		Name:      team.Name,
		Email:     team.Email,
		CreatedAt: team.CreatedAt,
		UpdatedAt: team.UpdatedAt,
	}, nil
}

func (s *identitysrvc) AddTeamMember(ctx context.Context, payload *identity.AddTeamMemberPayload) (res *identity.Team, err error) {
	//TODO implement me
	panic("implement me")
}

func (s *identitysrvc) RemoveTeamMember(ctx context.Context, payload *identity.RemoveTeamMemberPayload) (res *identity.Team, err error) {
	//TODO implement me
	panic("implement me")
}

// parseTeam is a workaround for the 'create_team' stored procedure which returns a string literal
// instead of Go values. If it cannot be type cast to string it will error.
func parseTeam(input interface{}) (*identity.Team, error) {
	s, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("unable to parse team")
	}
	// Remove parenthesis
	s = strings.TrimLeft(s, "(")
	s = strings.TrimRight(s, ")")

	// Split by comma
	parts := strings.Split(s, ",")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid format")
	}

	// Remove quotes
	for i := range parts {
		parts[i] = strings.Trim(parts[i], "\"")
	}

	// Create Team and assign values
	t := &identity.Team{
		TeamUUID:  parts[0],
		Name:      parts[1],
		Email:     parts[2],
		CreatedAt: &parts[3], // or use a function to parse into datetime if necessary
		UpdatedAt: nil,       // assuming that updated_at is not provided in the original string
	}

	return t, nil
}
func CalculateIdentityMetadata(totalRecords, page, pageSize int) *identity.PaginationMetadata {
	if totalRecords == 0 {
		return &identity.PaginationMetadata{}
	}
	return &identity.PaginationMetadata{
		CurrentPage: int32(page),
		PageSize:    int32(pageSize),
		FirstPage:   1,
		LastPage:    int32(int(math.Ceil(float64(totalRecords) / float64(pageSize)))),
		Total:       int32(totalRecords),
	}
}
