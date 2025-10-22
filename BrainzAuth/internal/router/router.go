package router

import (
	"brainz/auth/internal/handler"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func Register(h *server.Hertz) {
	h.GET("/auth", handler.Auth)
	h.POST("/create_key", handler.CreateApiKey)
	h.GET("/api_keys", handler.GetDeveloperApiKeys)
}
