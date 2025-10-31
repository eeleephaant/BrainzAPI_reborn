package db

import (
	wsmodels "brainz-api/internal/models/ws_models"
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var Ctx = context.Background()

func ConnectRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: "",
		DB:       0,
	})

	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Fatal("failed to connect to Redis:", err)
	}
}

func RedisEventsListener() {
	pubsub := RedisClient.PSubscribe(Ctx, "info_stream:*")
	for msg := range pubsub.Channel() {
		wsmodels.MainHub.Broadcast(msg.Channel, []byte(msg.Payload))
	}
}
