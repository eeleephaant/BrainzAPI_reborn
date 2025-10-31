package app

import (
	"brainz/developersapi/internal/config"
	"brainz/developersapi/internal/router"
	"brainz/developersapi/internal/security"
	"brainz/developersapi/internal/storage/postgres"
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/hertz/pkg/app/server"
)

func Run(ctx context.Context, log *logger.Logger, cfg *config.Config) error {
	h := server.Default(server.WithHostPorts(":8080"))

	postgres.Connect(ctx, &cfg.Postgres)
	security.LoadCfg(&cfg.Auth)

	router.Register(h)

	go func() {
		h.Spin()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop
	h.Shutdown(ctx)
	postgres.Close()

	// log.Info("Server Stopped")

	return nil
}
