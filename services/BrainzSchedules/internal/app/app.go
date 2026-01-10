package app

import (
	"brainz-api/internal/config"
	"brainz-api/internal/db"
	"brainz-api/internal/handler"
	"brainz-api/internal/repositories"
	"brainz-api/internal/router"
	"brainz-api/internal/services"
	"context"

	"github.com/cloudwego/hertz/pkg/app/server"
	"go.uber.org/zap"
)

func Run(ctx context.Context, cfg *config.Config) error {
	psqlPool, err := db.ConnectPostgres(ctx, &cfg.Postgres)
	if err != nil {
		panic(err)
	}
	rc, err := db.ConnectRedis(ctx, &cfg.Redis)
	if err != nil {
		panic(err)
	}

	gr := repositories.NewGroupRepository(psqlPool)
	ir := repositories.NewInstitutionRepository(psqlPool)
	lr := repositories.NewLessonRepository(psqlPool)

	ls := services.NewLessonService(lr)
	is := services.NewInstitutionService(ir)
	gs := services.NewGroupService(gr)

	sh := handler.NewScheduleHandler(rc, ls)
	ih := handler.NewInstitutionHandler(is)
	gh := handler.NewGroupHandler(gs)

	h := server.Default(server.WithHostPorts(":8080"))

	router.Register(ctx, h, sh, ih, gh)

	go db.RedisEventsListener(ctx, rc)

	h.OnShutdown = append(h.OnShutdown, func(ctx context.Context) {
		zap.L().Info("Stopping Server gracefully...")
		psqlPool.Close()
		rc.Close()
		<-ctx.Done()
		zap.L().Info("Server Stopped gracefully!")
	})

	h.Spin()
	return nil
}
