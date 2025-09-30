package main

import (
	"brainz-api/internal/db"
	"brainz-api/internal/middleware"
	"brainz-api/internal/models"
	"brainz-api/internal/router"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/limiter"
)

func main() {
	db.Connect()
	db.ConnectRedis()
	db.DB.AutoMigrate(&models.Institution{}, &models.Group{}, &models.Lesson{})

	db.DB.FirstOrCreate(&models.Institution{
		Name: "ТТСИиГХ",
		Site: "http://tci72.ru",
	})

	h := server.Default(server.WithHostPorts(":8080"))

	h.Use(limiter.AdaptiveLimit())
	h.Use(middleware.AuthMiddleware("http://brainz-auth:8080/auth"))
	h.Use(middleware.RecoveryMiddleware())

	router.Register(h)

	h.Spin()
}
