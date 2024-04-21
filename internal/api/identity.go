package api

import (
	"context"
	"github.com/danielmichaels/tawny/gen/identity"
	"github.com/danielmichaels/tawny/internal/auth"
	"github.com/danielmichaels/tawny/internal/logger"
	"github.com/danielmichaels/tawny/internal/store"
	"goa.design/goa/v3/security"
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
	res = &identity.UserResult{}
	s.logger.Print("identity.retrieveUser")
	return
}

// Retrieve all users that this user can see from associated teams.
func (s *identitysrvc) ListUsers(ctx context.Context, p *identity.ListUsersPayload) (res *identity.Users, err error) {
	res = &identity.Users{}
	s.logger.Print("identity.listUsers")
	return
}

// Create a new team
func (s *identitysrvc) CreateTeam(ctx context.Context, p *identity.CreateTeamPayload) (res *identity.Team, err error) {
	res = &identity.Team{}
	s.logger.Print("identity.createTeam")
	return
}
