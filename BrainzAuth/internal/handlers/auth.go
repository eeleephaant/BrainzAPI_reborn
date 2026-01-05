package handlers

import (
	"brainz/auth/internal/dtos"
	"brainz/auth/internal/services"
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

type AuthHandler struct {
	aks *services.ApiKeysService
}

func NewAuthHandler(aks *services.ApiKeysService) *AuthHandler {
	return &AuthHandler{aks: aks}
}

func (ah *AuthHandler) Auth(ctx context.Context, c *app.RequestContext) {
	apiKey := c.Query("key")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, dtos.AuthResult{
			Status:       false,
			ErrorMessage: "missing apiKey",
		})
		return
	}
	err := ah.aks.ValidateKey(ctx, apiKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dtos.AuthResult{
			Status:       false,
			ErrorMessage: err.Error(),
		})
		return
	}

}
