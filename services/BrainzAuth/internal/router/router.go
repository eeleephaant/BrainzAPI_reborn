package router

import (
	"brainz/auth/internal/handlers"
	"brainz/auth/internal/services"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func Register(h *server.Hertz, aks *services.ApiKeysService, akus *services.ApiKeyUsageService) {
	ah := handlers.NewAuthHandler(aks, akus)
	akmh := handlers.NewApiKeysManagmentHandler(aks)
	h.GET("/auth", ah.Auth)
	h.POST("/key", akmh.CreateApiKey)
	h.DELETE("/key", akmh.RemoveApiKey)
	h.GET("/keys", akmh.GetDeveloperApiKeys)
}
