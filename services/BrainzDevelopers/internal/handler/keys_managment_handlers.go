package handler

import (
	"brainz/common/dtos"
	"brainz/developersapi/internal/services"
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
)

type KeysManagmentHandler struct {
	ks *services.ApiKeysService
	ss *services.SessionService
}

func NewKeysManagmentHandler(ks *services.ApiKeysService) *KeysManagmentHandler {
	return &KeysManagmentHandler{ks: ks}
}

func (k *KeysManagmentHandler) GetApiKeys(ctx context.Context, c *app.RequestContext) {
	sessionToken := c.Request.Header.Get("X-Session-Token")
	if err != nil {
		c.JSON(400, map[string]string{"error": "invalid request"})
		return
	}
	keys, err := k.ks.GetApiKeys(ctx, devUUID)
	if err != nil {
		c.JSON(500, map[string]string{"error": "internal server error"})
		return
	}
	c.JSON(200, keys)
}

func (k *KeysManagmentHandler) CreateApiKey(ctx context.Context, c *app.RequestContext) {
	apiCreateRequest := dtos.ApiKeyCreateDto{}
	err := c.BindAndValidate(apiCreateRequest)
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
