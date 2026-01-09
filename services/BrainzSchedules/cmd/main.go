package main

import (
	"brainz-api/internal/db"
	"brainz-api/internal/middleware"
	"brainz-api/internal/router"
	"net/http"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/limiter"
)

func main() {
	db.ConnectPostgres()
	db.ConnectRedis()
	db.DB.AutoMigrate(&models.Institution{}, &models.Group{}, &models.Lesson{})

	db.DB.FirstOrCreate(&models.Institution{}, models.Institution{Name: "ТТСИиГХ"}, models.Institution{Site: "http://tci72.ru"})

	h := server.Default(server.WithHostPorts(":8080"))
	h.NoHijackConnPool = true

	h.Use(middleware.RecoveryMiddleware())
	h.Use(limiter.AdaptiveLimit())
	h.Use(middleware.AuthMiddleware("http://brainz-auth:8080/auth", &http.Client{Timeout: 3 * time.Second}))

	router.Register(h)

	h.Spin()
}
