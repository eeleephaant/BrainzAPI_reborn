package services

import (
	"brainz/auth/internal/models"
	"brainz/auth/internal/repositories"
	"brainz/auth/internal/utils"
	"context"
	"time"

	"github.com/google/uuid"
)

type ApiKeyUsageService struct {
	akur *repositories.ApiKeyUsageRepository
	akr  *repositories.ApiKeysRepository
}

func NewApiKeysUsageService(akur *repositories.ApiKeyUsageRepository, akr *repositories.ApiKeysRepository) *ApiKeyUsageService {
	return &ApiKeyUsageService{akur: akur,
		akr: akr}
}

func (akus *ApiKeyUsageService) RecordUsage(ctx context.Context, APIKey string, endpoint string, method string) (*models.ApiKeyUsage, error) {
	// op := "ApiKeyUsageService.CreateUsage"
	uuid := uuid.New()
	APIKeyID, err := utils.ExtractDataFromKey(APIKey)
	if err != nil {
		return nil, err
	}
	newAPIKeyUsage := models.ApiKeyUsage{
		ID:           uuid,
		ApiKeyID:     APIKeyID.ID,
		Endpoint:     endpoint,
		Method:       method,
		ResponseCode: "200",
		UsageAt:      time.Now(),
	}
	createdApiKeyUsage, err := akus.akur.Create(ctx, &newAPIKeyUsage)
	if err != nil {
		return nil, err
	}

	return createdApiKeyUsage, nil
}
