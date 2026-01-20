package services

import (
	"brainz/common/dtos"

	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ApiKeysService struct {
	hc         *client.Client
	authURL    string // URL Auth-сервиса, например: http://brainz-auth:8080
	authHeader string
}

func NewApiKeysService(hc *client.Client, authURL, authHeader string) *ApiKeysService {
	return &ApiKeysService{hc: hc, authURL: authURL, authHeader: authHeader}
}

func (aks *ApiKeysService) CreateApiKey(ctx context.Context, developerID uuid.UUID, name string) (*dtos.ApiKeyCreateResponse, error) {
	reqBody := dtos.ApiKeyCreateDto{
		DeveloperID: developerID,
		Name:        name,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req := protocol.AcquireRequest()
	defer protocol.ReleaseRequest(req)
	req.SetMethod(http.MethodPost)
	req.SetRequestURI(aks.authURL + "/create_key")
	req.Header.Set("Content-Type", "application/json")
	if aks.authHeader != "" {
		req.Header.Set("Authorization", aks.authHeader)
	}
	req.SetBody(bodyBytes)

	resp := protocol.AcquireResponse()
	defer protocol.ReleaseResponse(resp)

	err = aks.hc.Do(ctx, req, resp)
	if err != nil {
		zap.L().Error("HTTP request failed", zap.Error(err))
		return nil, fmt.Errorf("request to auth service: %w", err)
	}

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		zap.L().Error("Auth service returned error", zap.Int("status_code", resp.StatusCode()), zap.String("body", string(resp.Body())))
		return nil, fmt.Errorf("auth service error: %s", string(resp.Body()))
	}

	var apiKeyResp dtos.ApiKeyCreateResponse
	if err := json.Unmarshal(resp.Body(), &apiKeyResp); err != nil {
		zap.L().Error("Failed to unmarshal response", zap.Error(err))
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &apiKeyResp, nil
}

func (aks *ApiKeysService) GetApiKeys(ctx context.Context, developerID uuid.UUID) ([]dtos.ApiKeysGetResponse, error) {
	req := protocol.AcquireRequest()
	defer protocol.ReleaseRequest(req)

	req.SetMethod(http.MethodGet)
	req.SetRequestURI(aks.authURL + "/api_keys?developer_id=" + developerID.String())
	if aks.authHeader != "" {
		req.Header.Set("Authorization", aks.authHeader)
	}
	resp := protocol.AcquireResponse()
	defer protocol.ReleaseResponse(resp)

	err := aks.hc.Do(ctx, req, resp)
	if err != nil {
		zap.L().Error("HTTP request failed", zap.Error(err))
		return nil, fmt.Errorf("request to auth service: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		zap.L().Error("Auth service returned error", zap.Int("status_code", resp.StatusCode()), zap.String("body", string(resp.Body())))
		return nil, fmt.Errorf("auth service error: %s", string(resp.Body()))
	}

	var apiKeys []dtos.ApiKeysGetResponse
	if err := json.Unmarshal(resp.Body(), &apiKeys); err != nil {
		zap.L().Error("Failed to unmarshal response", zap.Error(err))
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return apiKeys, nil
}

func (aks *ApiKeysService) RemoveApiKey(ctx context.Context, apiKeyID string) error {
	req := protocol.AcquireRequest()
	defer protocol.ReleaseRequest(req)
	req.SetMethod(http.MethodDelete)
	req.SetRequestURI(aks.authURL + "/remove_key/" + apiKeyID)
	if aks.authHeader != "" {
		req.Header.Set("Authorization", aks.authHeader)
	}

	resp := protocol.AcquireResponse()
	defer protocol.ReleaseResponse(resp)

	err := aks.hc.Do(ctx, req, resp)
	if err != nil {
		return fmt.Errorf("request to auth service: %w", err)
	}

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusNoContent {
		return fmt.Errorf("auth service error: %s", string(resp.Body()))
	}

	return nil
}
