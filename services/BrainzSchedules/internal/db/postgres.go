package db

import (
	"brainz-api/internal/config"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectPostgres(ctx context.Context, cfg *config.PostgresConfig) (pool *pgxpool.Pool, err error) {
	dsnSched := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

	poolConfig, err := pgxpool.ParseConfig(dsnSched)

	if err != nil {
		log.Fatalf("failed to parse pgx config (devs): %v", err)
	}
	poolConfig.MaxConns = cfg.PoolMax
	poolConfig.MinConns = 3

	for i := range 3 {
		devsPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			log.Printf("failed to create pgx pool (devs), attempt %d: %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}
		return devsPool, nil
	}
	return nil, fmt.Errorf("failed to create pgx pool (devs) after 3 attempts: %v", err)
}
