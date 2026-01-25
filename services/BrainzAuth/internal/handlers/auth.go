package handlers

import (
	"brainz/auth/internal/dtos"
	"brainz/auth/internal/services"
	cmnDtos "brainz/common/dtos"
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"
)

type AuthHandler struct {
	aks  *services.ApiKeysService
	akus *services.ApiKeyUsageService
}

func NewAuthHandler(aks *services.ApiKeysService, akus *services.ApiKeyUsageService) *AuthHandler {
	return &AuthHandler{
		aks:  aks,
		akus: akus,
	}
}

func (ah *AuthHandler) Auth(ctx context.Context, c *app.RequestContext) {
	apiKey := c.Query("key")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, dtos.AuthResult{
			Status:       false,
			ErrorMessage: "missing apiKey",
		})
		zap.L().Error("Missing apiKey in request")
		return
	}
	requestedPermissionDTO := cmnDtos.RequestedPermissionDTO{}
	if err := c.BindAndValidate(&requestedPermissionDTO); err != nil {
		c.String(400, err.Error())
		return
	}

	err := ah.aks.ValidateKey(ctx, apiKey, c.Request.Header.Get("X-Original-IP"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, dtos.AuthResult{
			Status:       false,
			ErrorMessage: err.Error(),
		})
		zap.L().Error("API key validation failed", zap.Error(err))
		return
	}
	isLegit, err := ah.aks.CheckPermission(ctx, apiKey, requestedPermissionDTO.Permission)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		zap.L().Error("API key checking permissions failed", zap.Error(err))
		return
	}
	if !isLegit {
		c.JSON(http.StatusForbidden, dtos.AuthResult{Status: false, ErrorMessage: "API key does not have permission"})
		zap.L().Error("API key does not have permission", zap.Any("requestedPermissionDTO", requestedPermissionDTO))
		return
	}
	c.JSON(http.StatusOK, dtos.AuthResult{
		Status:       true,
		ErrorMessage: "",
	})
	zap.L().Info("Auth successful")

	path := c.Request.Header.Get("X-Original-Path")
	method := c.Request.Header.Get("X-Original-Method")
	_, err = ah.akus.RecordUsage(ctx, apiKey, path, method)
	if err != nil {
		zap.L().Error("Usage record error: ", zap.Error(err))
		return
	}

}
