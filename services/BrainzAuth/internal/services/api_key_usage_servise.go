package services

import (
	"brainz/auth/internal/models"
	"brainz/auth/internal/repositories"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ApiKeyUsageServise struct {
	akur *repositories.ApiKeyUsageRepository
}

func NewApiKeysUsageService(akur *repositories.ApiKeyUsageRepository) *ApiKeyUsageServise {
	return &ApiKeyUsageServise{akur: akur}
}

func (akus *ApiKeyUsageServise) CreateUsage(ctx context.Context, apikeyid uuid.UUID, endpoint string, responsecode string) (*models.ApiKeyUsage, error) {
	op := "ApiKeyUsageServise.CreateUsage"
	uuid := uuid.New()
	Newapikeyusage := models.ApiKeyUsage{
		ID:           uuid,
		ApiKeyID:     apikeyid,
		Endpoint:     endpoint,
		ResponseCode: responsecode,
		UsageAt:      time.Now(),
	}
	createdApiKeyUsage, err := akus.akur.Create(ctx, &Newapikeyusage)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create api key usage in db: %w", op, err)
	}
	return createdApiKeyUsage, nil
}
