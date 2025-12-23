package router

import (
	"brainz/developersapi/internal/handler"
	"brainz/developersapi/internal/services"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func Register(h *server.Hertz, ss *services.SessionService, us *services.UserService) {
	authHandler := &handler.AuthHandler{Us: us, Ss: ss}
	h.POST("/register", authHandler.Register)
	h.POST("/login", authHandler.Login)
	h.GET("/key", nil)
	h.GET("/keys", nil)
	h.DELETE("/key", nil)

}
