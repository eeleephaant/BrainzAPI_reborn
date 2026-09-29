package storage

import (
	"brainz/developersapi/internal/config"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectPostgres(ctx context.Context, cfg *config.PostgresConfig) (pool *pgxpool.Pool, err error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx config (devs): %w", err)
	}
	poolConfig.MaxConns = cfg.PoolMax
	poolConfig.MinConns = 3

	for i := 0; i < 3; i++ {
		pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
				// leave logging to caller; just retry with backoff
			time.Sleep(2 * time.Second)
			continue
		}
		return pool, nil
	}
	return nil, fmt.Errorf("failed to create pgx pool (devs) after 3 attempts: %v", err)
}
