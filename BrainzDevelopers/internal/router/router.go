package router

import (
	"brainz/developersapi/internal/handler"
	"brainz/developersapi/internal/services"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func Register(h *server.Hertz, ss *services.SessionService, us *services.UserService, es *services.EmailService) {
	authHandler := &handler.AuthHandler{Us: us, Ss: ss}
	emailHandler := &handler.EmailHandler{Us: us, Es: es}
	h.POST("/register", authHandler.Register)
	h.POST("/login", authHandler.Login)
	h.POST("/confirm-email", emailHandler.GetConfirmEmailCode)
	h.GET("/key", nil)
	h.GET("/keys", nil)
	h.DELETE("/key", nil)

}
