package handlers

import (
	"brainz/auth/internal/services"
	"brainz/common/dtos"
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ApiKeysManagementHandler struct {
	aks *services.ApiKeysService
}

func NewApiKeysManagementHandler(aks *services.ApiKeysService) *ApiKeysManagementHandler {
	return &ApiKeysManagementHandler{aks: aks}
}

func (akmh *ApiKeysManagementHandler) RemoveApiKey(ctx context.Context, c *app.RequestContext) {
	requestData := dtos.ApiKeyRemoveDto{}

	if err := c.BindAndValidate(&requestData); err != nil {
		c.String(400, err.Error())
		zap.L().Error("bind and validate failed", zap.Error(err))
		return
	}
	isLegit, err := akmh.aks.UserCanModifyKey(ctx, requestData.Key, requestData.DevUUID)
	if err != nil {
		c.String(400, err.Error())
		zap.L().Error("check user failed", zap.Error(err))
		return
	}

	if !isLegit {
		c.String(403, "Access denied")
		zap.L().Error("access denied", zap.Error(err))
		return
	}

	err = akmh.aks.RemoveApiKey(ctx, requestData.Key)
	if err != nil {
		c.String(500, "Failed to remove api key: "+err.Error())
		zap.L().Error("Failed to remove api key", zap.Error(err))
		return
	}
	c.String(200, "Api key removed successfully")
	zap.L().Info("Api key removed successfully")
}

func (akmh *ApiKeysManagementHandler) CreateApiKey(ctx context.Context, c *app.RequestContext) {
	requestData := dtos.ApiKeyCreateDto{}

	if err := c.BindAndValidate(&requestData); err != nil {
		c.String(400, err.Error())
		zap.L().Error("Invalid create api key request", zap.Error(err))
		return
	}

	key, displayedKey, err := akmh.aks.CreateApiKey(ctx, requestData.DeveloperID, requestData.Name)
	if err != nil {
		c.String(500, "Failed to create api key: "+err.Error())
		zap.L().Error("Failed to create api key", zap.Error(err))
		return
	}
	response := dtos.ApiKeyCreateResponse{
		Name:      key.Name,
		ApiKey:    *displayedKey,
		ExpiresAt: key.ExpireAt,
	}
	c.JSON(200, response)
	zap.L().Info("Api key created successfully")
}

func (akmh *ApiKeysManagementHandler) GetDeveloperApiKeys(ctx context.Context, c *app.RequestContext) {
	devIdStr := c.Query("developer_id")
	devId, err := uuid.Parse(devIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"message": "missing or invalid query param developer_id"})
		zap.L().Error("Invalid developer_id query param", zap.Error(err))
		return
	}

	apiKeys, err := akmh.aks.GetApiKeys(ctx, devId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"message": "failed to get api keys: " + err.Error()})
		zap.L().Error("Failed to get api keys", zap.Error(err))
		return
	}

	keysResp := make([]dtos.ApiKeyShareModel, 0, len(apiKeys))
	for _, k := range apiKeys {
		keysResp = append(keysResp, dtos.ApiKeyShareModel{
			Id:        k.ID,
			Title:     k.Name,
			ExpiresAt: k.ExpireAt,
			RevokedAt: k.RevokedAt,
			Suffix:    k.SuffixRaw,
			Prefix:    k.PrefixRaw,
		})
	}

	c.JSON(200, dtos.ApiKeysGetResponse{
		ApiKeys: keysResp,
	})
	zap.L().Info("Api keys retrieved successfully")
}
