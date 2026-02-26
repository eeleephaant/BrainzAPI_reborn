package handler

import (
	"brainz/common/dtos"
	"brainz/developersapi/internal/services"
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"
)

type KeysManagementHandler struct {
	ks *services.ApiKeysService
	ss *services.SessionService
}

func NewKeysManagementHandler(ks *services.ApiKeysService, ss *services.SessionService) *KeysManagementHandler {
	return &KeysManagementHandler{ks: ks, ss: ss}
}

func (k *KeysManagementHandler) GetApiKeys(ctx context.Context, c *app.RequestContext) {
	sessionToken := c.Request.Header.Get("X-Session-Token")
	session, err := k.ss.ValidateToken(ctx, sessionToken, c.ClientIP())
	if err != nil {
		c.JSON(400, map[string]string{"error": "invalid session token"})
		zap.L().Error("Invalid session token", zap.Error(err))
		return
	}
	keys, err := k.ks.GetApiKeys(ctx, session.DeveloperID)
	if err != nil {
		c.JSON(500, map[string]string{"error": "internal server error"})
		zap.L().Error("Internal server error", zap.Error(err))
		return
	}
	c.JSON(200, keys)
}

func (k *KeysManagementHandler) CreateApiKey(ctx context.Context, c *app.RequestContext) {
	apiCreateRequest := dtos.ApiKeyCreateDto{}
	err := c.BindAndValidate(&apiCreateRequest)
	if err != nil {
		c.JSON(400, map[string]string{"error": "invalid request"})
		return
	}
	newKey, err := k.ks.CreateApiKey(ctx, apiCreateRequest.DeveloperID, apiCreateRequest.Name)
	if err != nil {
		c.JSON(500, map[string]string{"error": "internal server error"})
		return
	}
	c.JSON(201, newKey)
}

func (k *KeysManagementHandler) DeleteApiKey(ctx context.Context, c *app.RequestContext) {
	sessionToken := c.Request.Header.Get("X-Session-Token")
	apiKeyStr := c.Query("api_key")
	session, err := k.ss.ValidateToken(ctx, sessionToken, c.ClientIP())
	if err != nil {
		c.JSON(400, map[string]string{"error": "invalid session token"})
		zap.L().Error("Invalid session token", zap.Error(err))
		return
	}
	err = k.ks.RemoveApiKey(ctx, apiKeyStr, session.DeveloperID.String())
	if err != nil {
		c.JSON(500, map[string]string{"error": "internal server error"})
		zap.L().Error("Internal server error", zap.Error(err))
		return
	}
	c.JSON(200, map[string]string{"status": "success"})

}
