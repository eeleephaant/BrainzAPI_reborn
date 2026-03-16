package app

import (
	"brainz-api/internal/config"
	"brainz-api/internal/handler"
	"brainz-api/internal/repositories"
	"brainz-api/internal/router"
	"brainz-api/internal/services"
	"brainz-api/internal/storage"
	"context"
	"fmt"
	"net/url"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func Run(ctx context.Context, cfg *config.Config) error {
	psqlPool, err := storage.ConnectPostgres(ctx, &cfg.Postgres)
	if err != nil {
		return err
	}

	var rc *redis.Client
	if cfg.Redis.Host != "" {
		connected, err := storage.ConnectRedis(ctx, &cfg.Redis)
		if err != nil {
			zap.L().Warn("Redis unavailable, running without Redis (info_stream push disabled)", zap.Error(err))
			rc = nil
		} else {
			rc = connected
		}
	} else {
		zap.L().Info("Redis host empty, running without Redis (info_stream push disabled)")
	}

	authServiceURL := cfg.AuthServiceURL
	if authServiceURL == "" {
		authServiceURL = "http://brainz-auth:8080"
	}
	authURL, err := url.Parse(authServiceURL)
	if err != nil {
		return fmt.Errorf("parse AUTH_SERVICE_URL: %w", err)
	}

	gr := repositories.NewGroupRepository(psqlPool)
	ir := repositories.NewInstitutionRepository(psqlPool)
	lr := repositories.NewLessonRepository(psqlPool)

	ls := services.NewLessonService(lr, gr)
	is := services.NewInstitutionService(ir)
	gs := services.NewGroupService(gr)

	as := services.NewAuthService(authURL)

	sh := handler.NewScheduleHandler(rc, ls, as)
	ih := handler.NewInstitutionHandler(is, as)
	gh := handler.NewGroupHandler(gs, as)

	hostPort := fmt.Sprintf("%s:%d", cfg.App.Address, cfg.App.Port)
	h := server.Default(server.WithHostPorts(hostPort))

	router.Register(ctx, h, sh, ih, gh)

	if rc != nil {
		go storage.RedisEventsListener(ctx, rc)
	}

	h.OnShutdown = append(h.OnShutdown, func(ctx context.Context) {
		zap.L().Info("Stopping Server gracefully...")
		psqlPool.Close()
		if rc != nil {
			rc.Close()
		}
		<-ctx.Done()
		zap.L().Info("Server Stopped gracefully!")
	})

	h.Spin()
	return nil
}
