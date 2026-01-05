package handlers

import (
	"brainz/auth/internal/dtos"
	"brainz/auth/internal/services"
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
)

type ApiKeysManagmentHandler struct {
	aks *services.ApiKeysService
}

func NewApiKeysManagmentHandler(aks *services.ApiKeysService) *ApiKeysManagmentHandler {
	return &ApiKeysManagmentHandler{aks: aks}
}

func (akmh *ApiKeysManagmentHandler) CreateApiKey(ctx context.Context, c *app.RequestContext) {
	requestData := dtos.ApiKeyCreateRequest{}

	if err := c.BindAndValidate(&requestData); err != nil {
		c.String(400, err.Error())
		return
	}

	err, key, displayedKey := akmh.aks.CreateApiKey(ctx, requestData.DeveloperId, requestData.Name)
	if err != nil {
		c.String(500, "Failed to create api key: "+err.Error())
		return
	}
	response := dtos.ApiKeyCreateResponse{
		ApiKey:    *displayedKey,
		ExpiresAt: key.ExpireAt,
	}
	c.JSON(200, response)

}

func (akmh *ApiKeysManagmentHandler) GetDeveloperApiKeys(ctx context.Context, c *app.RequestContext) {
	devIdStr := c.Query("developer_id")
	devId, err := uuid.Parse(devIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"message": "missing or invalid query param developer_id"})
	}

	err, apiKeys := akmh.aks.GetApiKeys(ctx, devId)

	keysResp := make([]dtos.ApiKeyShareModel, 0, len(apiKeys))
	for _, k := range apiKeys {
		keysResp = append(keysResp, dtos.ApiKeyShareModel{
			Id:        k.ID,
			Title:     k.Name,
			ExpiresAt: *&k.ExpireAt,
			RevokedAt: k.RevokedAt,
			Suffix:    k.SuffixRaw,
			Prefix:    k.PrefixRaw,
		})
	}

	c.JSON(200, dtos.DevApiKeysResponse{
		ApiKeys: keysResp,
	})
}
