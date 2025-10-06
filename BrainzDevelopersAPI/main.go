package main

import (
	"brainz/developersapi/internal/db"
	"brainz/developersapi/internal/models"
	"brainz/developersapi/internal/router"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func main() {
	db.Connect()
	db.DevsDB.AutoMigrate(&models.Developer{})

	h := server.Default(server.WithHostPorts(":8080"))

	router.Register(h)

	h.Spin()
}
