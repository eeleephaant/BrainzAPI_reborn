package app

import (
	"brainz/developersapi/internal/config"
	"brainz/developersapi/internal/handler"
	"brainz/developersapi/internal/repository"
	"brainz/developersapi/internal/router"
	"brainz/developersapi/internal/security"
	"brainz/developersapi/internal/services"
	"brainz/developersapi/internal/storage"
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/server"
	"go.uber.org/zap"
)

func Run(ctx context.Context, cfg *config.Config) error {
	hostPort := fmt.Sprintf("%s:%d", cfg.App.Address, cfg.App.Port)
	h := server.Default(server.WithHostPorts(hostPort))

	psqlPool, err := storage.ConnectPostgres(ctx, &cfg.Postgres)
	if err != nil {
		return err
	}

	security.LoadCfg(&cfg.Auth)

	hc, err := client.NewClient()
	if err != nil {
		return err
	}

	sr := repository.NewSessionRepository(psqlPool)
	ur := repository.NewUserRepository(psqlPool)
	etr := repository.NewEmailTokensRepository(psqlPool)

	ss := services.NewSessionService(sr)
	es := services.NewEmailConfirmService(etr)
	us := services.NewUserService(ur, es)
	aks := services.NewApiKeysService(hc, "http://brainz-auth:8080", "")

	ah := handler.NewAuthHandler(us, ss)
	eh := handler.NewEmailHandler(es, us)
	kmh := handler.NewKeysManagementHandler(aks, ss, us)

	router.Register(h, ah, eh, kmh)

	h.OnShutdown = append(h.OnShutdown, func(ctx context.Context) {
		zap.L().Info("Stopping Server gracefully...")
		psqlPool.Close()
		<-ctx.Done()
		zap.L().Info("Server Stopped gracefully!")
	})

	h.Spin()
	return nil
}
