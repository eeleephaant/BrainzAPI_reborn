package router

import (
	"brainz/developersapi/internal/handler"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func Register(h *server.Hertz, ah *handler.AuthHandler, kmh *handler.KeysManagementHandler) {
	h.POST("/register", ah.Register)
	h.POST("/login", ah.Login)
	h.GET("/keys", kmh.GetApiKeys)
	h.POST("/key", kmh.CreateApiKey)
	h.DELETE("/key", kmh.DeleteApiKey)
}
