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

	"github.com/cloudwego/hertz/pkg/app/server"
	"go.uber.org/zap"
)

func Run(ctx context.Context, cfg *config.Config) error {
	h := server.Default(server.WithHostPorts(":8080"))

	psqlPool, err := storage.ConnectPostgres(ctx, &cfg.Postgres)
	if err != nil {
		return err
	}
	utils.LoadCfg(&cfg.Auth)

	akr := repositories.NewApiKeysRepository(psqlPool)
	aks := services.NewApiKeysService(akr)

	akur := repositories.NewApiKeyUsageRepository(psqlPool)
	akus := services.NewApiKeysUsageService(akur, akr)

	ah := handlers.NewAuthHandler(aks, akus)
	akmh := handlers.NewApiKeysManagementHandler(aks)

	router.Register(h, ah, akmh)

	h.OnShutdown = append(h.OnShutdown, func(ctx context.Context) {
		zap.L().Info("Stopping Server gracefully...")
		psqlPool.Close()
		<-ctx.Done()
		zap.L().Info("Server Stopped gracefully!")
	})

	h.Spin()
	return nil
}
