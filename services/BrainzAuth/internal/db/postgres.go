package db

import (
	"brainz/auth/internal/config"
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var AuthPool *pgxpool.Pool

func Connect(ctx context.Context, cfg *config.PostgresConfig) {
	dsnAuth := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

	poolConfigAuth, err := pgxpool.ParseConfig(dsnAuth)

	if err != nil {
		log.Fatalf("failed to parse pgx config (devs): %v", err)
	}
	poolConfigAuth.MaxConns = cfg.PoolMax
	poolConfigAuth.MinConns = 3

	authsPoool, err := pgxpool.NewWithConfig(ctx, poolConfigAuth)
	if err != nil {
		log.Fatalf("failed to create pgx pool (devs): %v", err)
	}
	AuthPool = authsPoool
}

func Close() {
	if AuthPool != nil {
		AuthPool.Close()
	}
}
