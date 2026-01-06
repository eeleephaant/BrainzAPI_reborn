package postgres

import (
	"brainz/developersapi/internal/config"
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DevsPool *pgxpool.Pool

func Connect(ctx context.Context, cfg *config.PostgresConfig) {
	dsnDevs := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

	poolConfigDevs, err := pgxpool.ParseConfig(dsnDevs)

	if err != nil {
		log.Fatalf("failed to parse pgx config (devs): %v", err)
	}
	poolConfigDevs.MaxConns = cfg.PoolMax
	poolConfigDevs.MinConns = 3

	devsPool, err := pgxpool.NewWithConfig(ctx, poolConfigDevs)
	if err != nil {
		log.Fatalf("failed to create pgx pool (devs): %v", err)
	}
	DevsPool = devsPool
}

func Close() {
	if DevsPool != nil {
		DevsPool.Close()
	}
}
