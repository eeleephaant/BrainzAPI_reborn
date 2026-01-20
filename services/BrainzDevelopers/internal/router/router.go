package router

import (
	"brainz/developersapi/internal/handler"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func Register(h *server.Hertz, ah *handler.AuthHandler, eh *handler.EmailHandler, kmh *handler.KeysManagmentHandler) {
	h.POST("/register", ah.Register)
	h.POST("/login", ah.Login)
	h.POST("/confirm-email", eh.GetConfirmEmailCode)
	h.GET("/keys", kmh.GetApiKeys)
	h.DELETE("/key", nil)

}
