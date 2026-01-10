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

	"github.com/cloudwego/hertz/pkg/app/server"
	"go.uber.org/zap"
)

func Run(ctx context.Context, cfg *config.Config) error {
	h := server.Default(server.WithHostPorts(":8080"))

	psqlPool, err := storage.ConnectPostgres(ctx, &cfg.Postgres)
	if err != nil {
		panic(err)
	}

	security.LoadCfg(&cfg.Auth)

	sr := repository.NewSessionRepository(psqlPool)
	ur := repository.NewUserRepository(psqlPool)
	etr := repository.NewEmailTokensRepository(psqlPool)

	ss := services.NewSessionService(sr)
	es := services.NewEmailConfirmService(etr)
	us := services.NewUserService(ur, es)

	ah := handler.NewAuthHandler(us, ss)
	eh := handler.NewEmailHandler(es, us)

	router.Register(h, ah, eh)

	h.OnShutdown = append(h.OnShutdown, func(ctx context.Context) {
		zap.L().Info("Stopping Server gracefully...")
		psqlPool.Close()
		<-ctx.Done()
		zap.L().Info("Server Stopped gracefully!")
	})

	h.Spin()
	return nil
}
