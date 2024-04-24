package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/danielmichaels/tawny/internal/config"
	"github.com/danielmichaels/tawny/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDatabasePool(ctx context.Context, cfg *config.Conf) (*pgxpool.Pool, error) {
	minConns := 2
	dbUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Db.User,
		cfg.Db.Password,
		cfg.Db.Host,
		cfg.Db.Port,
		cfg.Db.Db,
		cfg.Db.SSLMode,
	)
	// DATABASE_URL is a commonly used paradigm
	if os.Getenv("DATABASE_URL") != "" {
		dbUrl = os.Getenv("DATABASE_URL")
	}
	dbPool := fmt.Sprintf(
		"%s&pool_max_conns=%d&pool_min_conns=%d",
		dbUrl,
		cfg.Db.MaxConns,
		minConns,
	)
	c, err := pgxpool.ParseConfig(dbPool)
	if err != nil {
		return nil, err
	}

	// Setting the build statement cache to nil helps this work with pgbouncer
	c.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	c.MaxConnLifetime = 1 * time.Hour
	c.MaxConnIdleTime = 30 * time.Second
	return pgxpool.NewWithConfig(ctx, c)
}

func (store *Queries) BoostrapAdminIfNotExists(ctx context.Context, logger *logger.Logger) error {
	exists, err := store.DoesAdminExist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	logger.Info().Msg("admin user does not exist, creating...")
	pw := os.Getenv("ADMIN_PASSWORD")
	if pw == "" {
		return errors.New("ADMIN_PASSWORD environment variable not set")
	}
	hpw, err := HashPassword(pw)
	if err != nil {
		return err
	}
	admin, err := store.CreateUserWithNewTeam(ctx, CreateUserWithNewTeamParams{
		Column1:  pgtype.Text{String: "admin", Valid: true},
		Name:     pgtype.Text{String: "admin", Valid: true},
		Email:    pgtype.Text{String: "admin@tawny.internal", Valid: true},
		Password: pgtype.Text{String: hpw, Valid: true},
	})
	if err != nil {
		return err
	}
	err = store.UpdateUserRole(ctx, UpdateUserRoleParams{
		Role:  UserRoleAdmin,
		Token: admin.PersonalAccessToken,
	})
	if err != nil {
		return err
	}
	logger.Info().Interface("admin", admin).Msg("admin created")
	return nil
}
