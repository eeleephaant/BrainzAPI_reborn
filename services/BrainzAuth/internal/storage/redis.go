package storage

import (
	"brainz/auth/internal/config"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(ctx context.Context, cfg *config.RedisConfig) (*redis.Client, error) {
	rc := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + cfg.Port,
		Password: cfg.Password,
		DB:       0,
	})

	if err := rc.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return rc, nil
}
