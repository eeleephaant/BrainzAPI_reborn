package services

import (
	"brainz/auth/internal/models"
	"brainz/auth/internal/repositories"
	"brainz/auth/internal/utils"
	"brainz/common/permissions"
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

func (aks *ApiKeysService) UserCanModifyKey(ctx context.Context, apiKey string, devUUID uuid.UUID) (bool, error) {
	op := "ApiKeysService.UserCanModifyKey"
	keyID, err := resolveAPIKeyID(apiKey)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	key, err := aks.akr.GetById(ctx, keyID)
	if err != nil {
		return false, err
	}
	if key.DeveloperID != devUUID {
		return false, nil
	}
	return true, nil
}

func (aks *ApiKeysService) CheckPermission(ctx context.Context, apiKey string, perm permissions.Permission) (bool, error) {
	//op := "ApiKeysService.CheckPermission"
	extractData, err := utils.ExtractDataFromKey(apiKey)
	if err != nil {
		return false, err
	}

	perms, err := aks.akr.GetPermissions(ctx, extractData.ID)
	if err != nil {
		return false, err
	}
	for _, granted := range perms {
		if granted.Allows(perm) {
			return true, nil
		}
	}
	return false, nil
}

func (aks *ApiKeysService) ValidateKey(ctx context.Context, rawKey string, ipAddr string) error {
	op := "ApiKeysService.ValidateKey"
	extractData, err := utils.ExtractDataFromKey(rawKey)
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

	list, err := aks.akr.GetWhitelistedIPs(ctx, extractData.ID)
	if err != nil {
		return err
	}

	if len(list) == 0 {
		return nil
	}

	for _, ip := range list {
		if ip.String() == ipAddr {
			return nil
		}

	}

	return fmt.Errorf("%s: ip address %s is not whitelisted", op, ipAddr)
}

func (aks *ApiKeysService) CreateApiKey(ctx context.Context, developerId uuid.UUID, name string, grants []permissions.Permission) (*models.ApiKey, *string, error) {
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

	createdKey, err := aks.akr.CreateWithPermissions(ctx, &newKey, grants)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: failed to create api key in db: %w", op, err)
	}

	return createdKey, &displayedKey, nil
}

func (aks *ApiKeysService) RemoveApiKey(ctx context.Context, apiKey string) error {
	op := "ApiKeysService.RemoveApiKey"
	keyUUID, err := resolveAPIKeyID(apiKey)
	if err != nil {
		return fmt.Errorf("%s: failed to resolve api key id: %w", op, err)
	}

	key, err := aks.akr.GetById(ctx, keyUUID)
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

func resolveAPIKeyID(apiKey string) (uuid.UUID, error) {
	if keyID, err := uuid.Parse(apiKey); err == nil {
		return keyID, nil
	}

	extracted, err := utils.ExtractDataFromKey(apiKey)
	if err != nil {
		return uuid.Nil, err
	}
	return extracted.ID, nil
}
