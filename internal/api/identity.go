package api

import (
	"context"
	"github.com/danielmichaels/tawny/gen/identity"
	"github.com/danielmichaels/tawny/internal/auth"
	"github.com/danielmichaels/tawny/internal/logger"
	"github.com/danielmichaels/tawny/internal/ptr"
	"github.com/danielmichaels/tawny/internal/store"
	"goa.design/goa/v3/security"
	"math"
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
		ID:        nil,
		Username:  u.Username,
		Email:     u.Email,
		Verified:  &u.Verified,
		CreatedAt: ptr.Ptr(u.CreatedAt.Time.String()),
		UpdatedAt: ptr.Ptr(u.UpdatedAt.Time.String()),
	}
	return user, nil
}

// Retrieve all users that this user can see from associated teams.
func (s *identitysrvc) ListUsers(ctx context.Context, p *identity.ListUsersPayload) (res *identity.Users, err error) {
	ut := auth.CtxAuthInfo(ctx)
	res = &identity.Users{}
	u, err := s.db.ListUsers(ctx, ut.User)
	if err != nil {
		return nil, &identity.NotFound{
			Name:    "not found",
			Message: "resource not found",
			Detail:  "resource not found",
		}
	}
	users := &identity.Users{}
	for _, user := range u {
		users.Users = append(users.Users, &identity.UserResult{
			Username:  user.Username,
			Email:     user.Email,
			Role:      string(user.Role),
			Verified:  &user.Verified,
			CreatedAt: ptr.Ptr(user.CreatedAt.Time.String()),
			UpdatedAt: ptr.Ptr(user.UpdatedAt.Time.String()),
		})
	}
	users.Metadata = CalculateIdentityMetadata(len(users.Users), p.PageNumber, p.PageSize)

	return users, nil
}

// Create a new team
func (s *identitysrvc) CreateTeam(ctx context.Context, p *identity.CreateTeamPayload) (res *identity.Team, err error) {
	res = &identity.Team{}
	s.logger.Print("identity.createTeam")
	return
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
