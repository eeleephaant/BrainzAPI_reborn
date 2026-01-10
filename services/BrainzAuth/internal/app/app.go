package app

import (
	"brainz/auth/internal/config"
	"brainz/auth/internal/db"
	"brainz/auth/internal/repositories"
	"brainz/auth/internal/router"
	"brainz/auth/internal/services"
	"brainz/auth/internal/utils"
	"context"

	"github.com/cloudwego/hertz/pkg/app/server"
	"go.uber.org/zap"
)

func Run(ctx context.Context, cfg *config.Config) error {
	h := server.Default(server.WithHostPorts(":8080"))

	db.Connect(ctx, &cfg.Postgres)
	utils.LoadCfg(&cfg.Auth)

	akr := repositories.NewApiKeysRepository(db.AuthPool)
	aks := services.NewApiKeysService(akr)
	akur := repositories.ApiKeyUsageRepository(db.AuthPool)
	akus := services.NewApiKeysUsageService(akur)

	router.Register(h, aks, akus)

	h.OnShutdown = append(h.OnShutdown, func(ctx context.Context) {
		zap.L().Info("Stopping Server gracefully...")
		db.Close()
		<-ctx.Done()
		zap.L().Info("Server Stopped gracefully!")
	})

	h.Spin()
	return nil
}
