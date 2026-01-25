package router

import (
	"brainz/auth/internal/handlers"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func Register(h *server.Hertz, ah *handlers.AuthHandler, akmh *handlers.ApiKeysManagementHandler) {
	h.GET("/auth", ah.Auth)
	h.POST("/key", akmh.CreateApiKey)
	h.DELETE("/key", akmh.RemoveApiKey)
	h.GET("/keys", akmh.GetDeveloperApiKeys)
}
