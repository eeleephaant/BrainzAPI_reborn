package app

import (
	"brainz/auth/internal/config"
	"brainz/auth/internal/handlers"
	"brainz/auth/internal/repositories"
	"brainz/auth/internal/router"
	"brainz/auth/internal/services"
	"brainz/auth/internal/storage"
	"brainz/auth/internal/utils"
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func Run(ctx context.Context, cfg *config.Config) error {
	hostPort := fmt.Sprintf("%s:%d", cfg.App.Address, cfg.App.Port)
	h := server.Default(server.WithHostPorts(hostPort))

	psqlPool, err := storage.ConnectPostgres(ctx, &cfg.Postgres)
	if err != nil {
		return err
	}
	utils.LoadCfg(&cfg.Auth)

	var (
		rc  *redis.Client
		rls *services.RateLimiterService
	)

	if cfg.RateLimit.Enabled {
		redisClient, redisErr := storage.ConnectRedis(ctx, &cfg.Redis)
		if redisErr != nil {
			if cfg.RateLimit.FailOpen {
				zap.L().Warn("redis unavailable, rate limiter disabled", zap.Error(redisErr))
			} else {
				return redisErr
			}
		} else {
			rc = redisClient
			rls = services.NewRateLimiterService(redisClient, cfg.RateLimit)
		}
	}

	akr := repositories.NewApiKeysRepository(psqlPool)
	aks := services.NewApiKeysService(akr)

	akur := repositories.NewApiKeyUsageRepository(psqlPool)
	akus := services.NewApiKeysUsageService(akur, akr)

	ah := handlers.NewAuthHandler(aks, akus, rls)
	akmh := handlers.NewApiKeysManagementHandler(aks)

	router.Register(h, ah, akmh)

	h.OnShutdown = append(h.OnShutdown, func(ctx context.Context) {
		zap.L().Info("Stopping Server gracefully...")
		psqlPool.Close()
		if rc != nil {
			rc.Close()
		}
		zap.L().Info("Server Stopped gracefully!")
	})

	h.Spin()
	return nil
}
