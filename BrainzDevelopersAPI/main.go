package main

import (
	"github.com/cloudwego/hertz/pkg/app/server"
)

func main() {
	db.Connect()
	db.DB.AutoMigrate(&models.ApiKey{})

	// TODO: Ultrashit, delete when u are be super smart boy

	db.DB.FirstOrCreate(&models.ApiKey{
		Title:      "MVP key",
		ApiKeyHash: "d03058fd988b0276712c406c94130c13ce305410b9fb3f71da95d86c8dbc25a1",
		IpAddress:  "",
	})

	h := server.Default(server.WithHostPorts(":8080"))

	router.Register(h)

	h.Spin()
}
