package storage

import (
	"brainz/auth/internal/config"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectPostgres(ctx context.Context, cfg *config.PostgresConfig) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse pgx config: %w", err)
	}
	poolConfig.MaxConns = cfg.PoolMax
	poolConfig.MinConns = 3

	var lastErr error
	for i := range 3 {
		pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			lastErr = err
			log.Printf("failed to create pgx pool, attempt %d: %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}
		return pool, nil
	}
	return nil, fmt.Errorf("failed to create pgx pool after 3 attempts: %w", lastErr)
}
