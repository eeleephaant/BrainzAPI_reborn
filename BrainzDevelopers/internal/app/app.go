package app

import (
	"brainz/developersapi/internal/config"
	"brainz/developersapi/internal/repository"
	"brainz/developersapi/internal/router"
	"brainz/developersapi/internal/security"
	"brainz/developersapi/internal/services"
	"brainz/developersapi/internal/storage/postgres"
	"context"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/hertz/pkg/app/server"
	"go.uber.org/zap"
)

func Run(ctx context.Context, log *logger.Logger, cfg *config.Config) error {
	h := server.Default(server.WithHostPorts(":8080"))

	postgres.Connect(ctx, &cfg.Postgres)
	security.LoadCfg(&cfg.Auth)

	sr := repository.NewSessionRepository(postgres.DevsPool)
	ur := repository.NewUserRepository(postgres.DevsPool)

	ss := services.NewSessionService(sr)
	us := services.NewUserService(ur)

	router.Register(h, ss, us)

	h.OnShutdown = append(h.OnShutdown, func(ctx context.Context) {
		zap.L().Info("Stopping Server gracefully...")
		postgres.Close()
		<-ctx.Done()
		zap.L().Info("Server Stopped gracefully!")
	})

	h.Spin()
	return nil
}
