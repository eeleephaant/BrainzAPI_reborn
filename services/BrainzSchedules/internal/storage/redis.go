package storage

import (
	"brainz-api/internal/config"
	"brainz-api/internal/models"
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

	_, err := rc.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return rc, nil
}

func RedisEventsListener(ctx context.Context, rc *redis.Client) {
	pubsub := rc.PSubscribe(ctx, "info_stream:*")
	for msg := range pubsub.Channel() {
		models.MainHub.Broadcast(msg.Channel, []byte(msg.Payload))
	}
}
