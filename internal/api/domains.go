package api

import (
	"context"

	"github.com/danielmichaels/tawny/gen/domains"
	"github.com/danielmichaels/tawny/gen/identity"
	"github.com/danielmichaels/tawny/internal/auth"
	"github.com/danielmichaels/tawny/internal/k8sclient"
	"github.com/danielmichaels/tawny/internal/logger"
	"github.com/danielmichaels/tawny/internal/store"
	"goa.design/goa/v3/security"
)

// domains service example implementation.
// The example methods log the requests and return zero values.
type domainssrvc struct {
	logger  *logger.Logger
	db      *store.Queries
	kclient *k8sclient.K8sClient
}

// NewDomains returns the domains service implementation.
func NewDomains(
	logger *logger.Logger,
	db *store.Queries,
	kclient *k8sclient.K8sClient,
) domains.Service {
	return &domainssrvc{logger, db, kclient}
}

// APIKeyAuth implements the authorization logic for service "identity" for the
// "api_key" security scheme.
func (s *domainssrvc) APIKeyAuth(
	ctx context.Context,
	key string,
	scheme *security.APIKeyScheme,
) (context.Context, error) {
	ak := auth.NewApiKey()
	ctx, err := ak.Validate(ctx, key, scheme, s.db)
	if err != nil {
		s.logger.Error().Err(err).Msg("token invalid")
		return ctx, &identity.Unauthorized{Message: "token invalid"}
	}
	return ctx, nil
}

func (s *domainssrvc) ListDomains(
	ctx context.Context,
	payload *domains.ListDomainsPayload,
) (res *domains.DomainsResult, err error) {
	//TODO implement me
	panic("implement me")
}

func (s *domainssrvc) CreateDomain(
	ctx context.Context,
	payload *domains.CreateDomainPayload,
) (res *domains.DomainResult, err error) {
	//TODO implement me
	panic("implement me")
}
