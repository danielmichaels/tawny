package auth

import (
	"context"
	"fmt"
	"github.com/danielmichaels/tawny/internal/store"

	"goa.design/goa/v3/security"
)

type ApiKey struct{}

func NewApiKey() *ApiKey {
	return &ApiKey{}
}

func (a *ApiKey) Validate(
	ctx context.Context,
	key string,
	scheme *security.APIKeyScheme,
	db *store.Queries,
) (context.Context, error) {
	u, err := db.RetrieveUserWithTeamInfoByAPIKEY(ctx, key)
	if err != nil {
		return ctx, fmt.Errorf("no user matches apikey. err: %w", err)
	}
	ctx = CtxSetAuthInfo(ctx, CtxInfo{
		UserUUID: u.Uuid,
		TeamUUID: u.TeamUuid,
	})
	return ctx, nil
}

type CtxInfo struct {
	UserUUID string
	TeamUUID string
}
type ctxValue int

const (
	ctxValueClaims ctxValue = iota
)

func CtxSetAuthInfo(ctx context.Context, auth CtxInfo) context.Context {
	return context.WithValue(ctx, ctxValueClaims, auth)
}

func CtxAuthInfo(ctx context.Context) (auth CtxInfo) {
	auth, _ = ctx.Value(ctxValueClaims).(CtxInfo)
	return
}
