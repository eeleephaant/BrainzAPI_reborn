package handler

import (
	"brainz/developersapi/internal/services"
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

type KeysManagmentHandler struct {
	ks *services.ApiKeysService
}

func NewKeysManagmentHandler(ks *services.ApiKeysService) *KeysManagmentHandler {
	return &KeysManagmentHandler{ks: ks}
}

func (k *KeysManagmentHandler) GetApiKeys(ctx context.Context, c *app.RequestContext) {
	keys, err := k.ks.GetApiKeys(ctx, developerID)
	if err != nil {
		c.JSON(500, map[string]string{"error": "internal server error"})
		return
	}
}
