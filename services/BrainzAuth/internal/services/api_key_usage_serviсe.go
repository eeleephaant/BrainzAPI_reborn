package services

import (
	"brainz/auth/internal/models"
	"brainz/auth/internal/repositories"
	"context"
	"time"

	"github.com/google/uuid"
)

type ApiKeyUsageService struct {
	akur *repositories.ApiKeyUsageRepository
}

func NewApiKeysUsageService(akur *repositories.ApiKeyUsageRepository) *ApiKeyUsageService {
	return &ApiKeyUsageService{akur: akur}
}

func (akus *ApiKeyUsageService) CreateUsage(ctx context.Context, APIKeyID uuid.UUID, endpoint string, responseCode string) (*models.ApiKeyUsage, error) {
	// op := "ApiKeyUsageService.CreateUsage"
	uuid := uuid.New()
	newAPIKeyUsage := models.ApiKeyUsage{
		ID:           uuid,
		ApiKeyID:     APIKeyID,
		Endpoint:     endpoint,
		ResponseCode: responseCode,
		UsageAt:      time.Now(),
	}
	createdApiKeyUsage, err := akus.akur.Create(ctx, &newAPIKeyUsage)
	if err != nil {
		return nil, err
	}

	return createdApiKeyUsage, nil
}
