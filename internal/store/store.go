package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/danielmichaels/tawny/internal/config"
	"github.com/danielmichaels/tawny/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"time"
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
	admin, err := store.CreateUserWithNewTeam(ctx, CreateUserWithNewTeamParams{
		Name:     "admin",
		Email:    "admin@tawny.internal",
		Password: pw,
	})
	if err != nil {
		return err
	}
	err = store.UpdateUserRole(ctx, UpdateUserRoleParams{
		Role:   UserRoleAdmin,
		UserID: admin.UserID,
	})
	if err != nil {
		return err
	}
	logger.Info().Interface("admin", admin).Msg("admin created")
	return nil
}
