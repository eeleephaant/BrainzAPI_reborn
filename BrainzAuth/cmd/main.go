package main

import (
	"brainz/auth/internal/db"
	"brainz/auth/internal/models"
	"brainz/auth/internal/router"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func main() {
	db.Connect()
	db.DB.AutoMigrate(&models.ApiKey{}, &models.ApiKeyUsage{})

	h := server.Default(server.WithHostPorts(":8080"))

	router.Register(h)

	h.Spin()
}
