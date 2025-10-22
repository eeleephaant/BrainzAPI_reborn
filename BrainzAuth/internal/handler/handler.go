package handler

import (
	"brainz/auth/internal/db"
	"brainz/auth/internal/models"
	"brainz/auth/internal/models/dtos"
	"brainz/auth/utils"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gofrs/uuid/v5"
)

func Auth(ctx context.Context, c *app.RequestContext) {
	apiKey := c.Query("key")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, dtos.AuthResult{
			Status:       false,
			ErrorMessage: "missing apiKey",
		})
		return
	}

	parts := strings.Split(apiKey, ":")
	if len(parts) != 2 {
		c.JSON(http.StatusUnauthorized, dtos.AuthResult{
			Status:       false,
			ErrorMessage: "invalid apiKey format",
		})
		return
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		c.JSON(http.StatusUnauthorized, dtos.AuthResult{
			Status:       false,
			ErrorMessage: "invalid apiKey ID",
		})
		return
	}

	rawKey := parts[1]

	var key models.ApiKey
	if err := db.DB.First(&key, id).Error; err != nil {
		c.JSON(http.StatusUnauthorized, dtos.AuthResult{
			Status:       false,
			ErrorMessage: "api key not found",
		})
		return
	}

	hashed := utils.GetHashArgon2(rawKey, []byte(key.Salt))
	if subtle.ConstantTimeCompare(hashed, []byte(key.ApiKeyHash)) != 1 {
		c.JSON(http.StatusUnauthorized, dtos.AuthResult{
			Status:       false,
			ErrorMessage: "invalid apiKey",
		})
		return
	}

	now := time.Now()
	if err := db.DB.Model(&key).Update("last_used_at", now).Error; err != nil {
		fmt.Println("failed to update last_used_at:", err)
	}

	c.JSON(http.StatusOK, dtos.AuthResult{
		Status:       true,
		ErrorMessage: "success",
	})
}

func CreateApiKey(ctx context.Context, c *app.RequestContext) {
	requestData := dtos.ApiKeyCreateRequest{}

	if err := c.BindAndValidate(&requestData); err != nil {
		c.String(400, err.Error())
		return
	}

	salt := make([]byte, 16)
	rand.Read(salt)

	newKey := models.ApiKey{}
	db.DB.Create(&newKey)

	apiKeyRaw, rawBytes, err := utils.GenerateRawApiKey(newKey.ID)
	if err != nil {
		c.String(500, "failed to generate api key")
		return
	}

	hashed := utils.GetHashArgon2(string(rawBytes), salt)

	expiresAt := time.Now().AddDate(0, 2, 0)

	newKey.ApiKeyHash = string(hashed)
	newKey.Salt = string(salt)
	newKey.ExpiresAt = &expiresAt
	newKey.DeveloperId = requestData.DeveloperId
	newKey.CreatedAt = time.Now()

	db.DB.Save(&newKey)

	response := dtos.ApiKeyCreateResponse{
		ApiKey:    apiKeyRaw,
		ExpiresAt: expiresAt,
	}

	c.JSON(http.StatusOK, response)
}

func GetDeveloperApiKeys(ctx context.Context, c *app.RequestContext) {
	dev_id_str := c.Query("developer_id")

	devId := uuid.FromStringOrNil(dev_id_str)
	if devId == uuid.Nil {
		c.JSON(http.StatusBadRequest, map[string]string{"message": "missing or invalid query param developer_id"})
	}

	var apiKeys []models.ApiKey

	if err := db.DB.Where("developer_id = ?", devId).Find(&apiKeys).Error; err != nil {
		c.String(500, "Database error: "+err.Error())
		return
	}
	keysResp := make([]dtos.ApiKeyShareModel, 0, len(apiKeys))
	for _, k := range apiKeys {
		keysResp = append(keysResp, dtos.ApiKeyShareModel{
			Id:        k.ID,
			Title:     k.Title,
			ExpiresAt: *k.ExpiresAt,
			LastUsed:  *k.LastUsedAt,
		})
	}

	c.JSON(200, dtos.DevApiKeysResponse{
		ApiKeys: keysResp,
	})
}
