package main

import (
	"brainz-api/internal/config"
	"brainz-api/internal/db"
	"brainz-api/internal/middleware"
	"context"
	"net/http"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/limiter"
)

func main() {
	ctx := context.Background()
	cfg := config.MustLoad()
	db.ConnectPostgres(ctx, &cfg.Postgres)
	db.ConnectRedis(ctx, &cfg.Redis)

	h := server.Default(server.WithHostPorts(":8080"))
	h.NoHijackConnPool = true

	h.Use(middleware.RecoveryMiddleware())
	h.Use(limiter.AdaptiveLimit())
	h.Use(middleware.AuthMiddleware("http://brainz-auth:8080/auth", &http.Client{Timeout: 3 * time.Second}))

	h.Spin()
}
