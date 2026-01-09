package db

import (
	"brainz-api/internal/config"
	"brainz-api/internal/models"
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func ConnectRedis(ctx context.Context, cfg *config.RedisConfig) {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + cfg.Port,
		Password: "",
		DB:       0,
	})

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatal("failed to connect to Redis:", err)
	}
}

func RedisEventsListener(ctx context.Context) {
	pubsub := RedisClient.PSubscribe(ctx, "info_stream:*")
	for msg := range pubsub.Channel() {
		models.MainHub.Broadcast(msg.Channel, []byte(msg.Payload))
	}
}
