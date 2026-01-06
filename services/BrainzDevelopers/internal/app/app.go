package app

import (
	"brainz/developersapi/internal/config"
	"brainz/developersapi/internal/repository"
	"brainz/developersapi/internal/router"
	"brainz/developersapi/internal/security"
	"brainz/developersapi/internal/services"
	"brainz/developersapi/internal/storage/postgres"
	"context"

	"github.com/cloudwego/hertz/pkg/app/server"
	"go.uber.org/zap"
)

func Run(ctx context.Context, cfg *config.Config) error {
	h := server.Default(server.WithHostPorts(":8080"))

	postgres.Connect(ctx, &cfg.Postgres)
	security.LoadCfg(&cfg.Auth)

	sr := repository.NewSessionRepository(postgres.DevsPool)
	ur := repository.NewUserRepository(postgres.DevsPool)
	etr := repository.NewEmailTokensRepository(postgres.DevsPool)

	ss := services.NewSessionService(sr)
	es := services.NewEmailConfirmService(etr)
	us := services.NewUserService(ur, es)

	router.Register(h, ss, us, es)

	h.OnShutdown = append(h.OnShutdown, func(ctx context.Context) {
		zap.L().Info("Stopping Server gracefully...")
		postgres.Close()
		<-ctx.Done()
		zap.L().Info("Server Stopped gracefully!")
	})

	h.Spin()
	return nil
}
