package router

import (
	"brainz/auth/internal/handlers"
	"brainz/auth/internal/services"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func Register(h *server.Hertz, aks *services.ApiKeysService) {
	ah := handlers.NewAuthHandler(aks)
	akmh := handlers.NewApiKeysManagmentHandler(aks)
	h.GET("/auth", ah.Auth)
	h.POST("/create_key", akmh.CreateApiKey)
	h.GET("/api_keys", akmh.GetDeveloperApiKeys)
}
