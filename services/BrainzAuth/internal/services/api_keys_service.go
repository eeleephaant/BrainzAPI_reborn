package services

import (
	"brainz/auth/internal/models"
	"brainz/auth/internal/repositories"
	"brainz/auth/internal/utils"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ApiKeysService struct {
	akr *repositories.ApiKeysRepository
}

func NewApiKeysService(akr *repositories.ApiKeysRepository) *ApiKeysService {
	return &ApiKeysService{akr: akr}
}
func (aks *ApiKeysService) ValidateKey(ctx context.Context, raw_key string) error {
	op := "ApiKeysService.ValidateKey"
	extractData, err := utils.ExtractDataFromKey(raw_key)
	if err != nil {
		return err
	}
	key, err := aks.akr.GetById(ctx, extractData.ID)
	if err != nil {
		return err
	}

	if key.RevokedAt != nil || key.ExpireAt.Before(time.Now()) {
		return fmt.Errorf("%s: apiKey is revoked or expired", op)
	}

	if !utils.CheckPassword(extractData.Secret, key.Salt, key.KeyHash) {
		return fmt.Errorf("%s: invalid apiKey", op)
	}

	return nil
}

func (aks *ApiKeysService) CreateApiKey(ctx context.Context, developerId uuid.UUID, name string) (*models.ApiKey, *string, error) {
	op := "ApiKeysService.CreateApiKey"
	expiresAt := time.Now().AddDate(0, 2, 0)
	rawKey, err := utils.GenerateSecretKey(32)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: failed to generate raw api key: %w", op, err)
	}
	salt := utils.GetRandomSalt()
	hashedKey := utils.GetHashArgon2(rawKey, salt)
	newUUID := uuid.New()

	displayedKey := fmt.Sprintf("brainz_%s:%s", newUUID.String(), rawKey)

	newKey := models.ApiKey{
		ID:          newUUID,
		DeveloperID: developerId,
		Name:        name,
		ExpireAt:    expiresAt,
		CreatedAt:   time.Now(),
		Salt:        salt,
		KeyHash:     hashedKey,
		PrefixRaw:   displayedKey[:10],
		SuffixRaw:   displayedKey[len(displayedKey)-4:],
	}

	createdKey, err := aks.akr.Create(ctx, &newKey)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: failed to create api key in db: %w", op, err)
	}

	return createdKey, &displayedKey, nil
}

func (aks *ApiKeysService) RemoveApiKey(ctx context.Context, keyId uuid.UUID) error {
	op := "ApiKeysService.RemoveApiKey"
	key, err := aks.akr.GetById(ctx, keyId)
	if err != nil {
		return fmt.Errorf("%s: failed to get api key from db: %w", op, err)
	}
	err = aks.akr.Remove(ctx, key)
	if err != nil {
		return fmt.Errorf("%s: failed to remove api key from db: %w", op, err)
	}
	return nil
}

func (aks *ApiKeysService) GetApiKeys(ctx context.Context, developerId uuid.UUID) ([]*models.ApiKey, error) {
	op := "ApiKeysService.GetApiKeys"
	keys, err := aks.akr.GetAllByDeveloperId(ctx, developerId)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get api keys from db: %w", op, err)
	}
	return keys, nil
}
