package services

import (
	"brainz/common/dtos"
	"brainz/common/permissions"

	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ApiKeysService struct {
	hc         *client.Client
	authURL    string
	authHeader string
}

var ErrAccessDenied = errors.New("access denied")

func NewApiKeysService(hc *client.Client, authURL, authHeader string) *ApiKeysService {
	return &ApiKeysService{hc: hc, authURL: authURL, authHeader: authHeader}
}

func (aks *ApiKeysService) CreateApiKey(ctx context.Context, issuer uuid.UUID, developerID uuid.UUID, name string, perms []permissions.Permission, ipWhitelist []string) (*dtos.ApiKeyCreateResponse, error) {
	if issuer != developerID {
		return nil, fmt.Errorf("you cannot create keys for other developers")
	}

	reqBody := dtos.ApiKeyCreateDto{
		DevUUID:     developerID,
		ApiKeyName:  name,
		Permissions: perms,
		IPWhitelist: ipWhitelist,
	}
	bodyBytes, err := sonic.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req := protocol.AcquireRequest()
	defer protocol.ReleaseRequest(req)
	req.SetMethod(http.MethodPost)
	req.SetRequestURI(aks.authURL + "/key")
	req.Header.Set("Content-Type", "application/json")
	if aks.authHeader != "" {
		req.Header.Set("Authorization", aks.authHeader)
	}
	req.SetBody(bodyBytes)

	resp := protocol.AcquireResponse()
	defer protocol.ReleaseResponse(resp)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

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

func (aks *ApiKeysService) GetApiKeys(ctx context.Context, developerID uuid.UUID) ([]dtos.ApiKeyShareModel, error) {
	base, err := url.Parse(aks.authURL)
	if err != nil {
		return nil, fmt.Errorf("invalid auth URL: %w", err)
	}
	base.Path = "/keys"
	base.RawQuery = url.Values{"developer_id": {developerID.String()}}.Encode()
	fullURL := base.String()

	req := protocol.AcquireRequest()
	defer protocol.ReleaseRequest(req)

	req.SetMethod(http.MethodGet)
	req.SetRequestURI(fullURL) // полный URL — иначе клиент не знает куда подключаться и идёт на localhost (404 от другого сервиса)
	if aks.authHeader != "" {
		req.Header.Set("Authorization", aks.authHeader)
	}
	resp := protocol.AcquireResponse()
	defer protocol.ReleaseResponse(resp)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = aks.hc.Do(ctx, req, resp)
	if err != nil {
		zap.L().Error("HTTP request failed", zap.Error(err))
		return nil, fmt.Errorf("request to auth service: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		zap.L().Error("Auth service returned error", zap.Int("status_code", resp.StatusCode()), zap.String("body", string(resp.Body())))
		return nil, fmt.Errorf("auth service error: %s", string(resp.Body()))
	}

	var wrapper dtos.ApiKeysGetResponse
	if err := json.Unmarshal(resp.Body(), &wrapper); err != nil {
		zap.L().Error("Failed to unmarshal response", zap.Error(err))
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return wrapper.ApiKeys, nil
}

func (aks *ApiKeysService) GetApiKeysUsage(ctx context.Context, developerID uuid.UUID) ([]dtos.ApiKeyUsageStats, error) {
	base, err := url.Parse(aks.authURL)
	if err != nil {
		return nil, fmt.Errorf("invalid auth URL: %w", err)
	}
	base.Path = "/keys/usage"
	base.RawQuery = url.Values{"developer_id": {developerID.String()}}.Encode()
	fullURL := base.String()

	req := protocol.AcquireRequest()
	defer protocol.ReleaseRequest(req)

	req.SetMethod(http.MethodGet)
	req.SetRequestURI(fullURL)
	if aks.authHeader != "" {
		req.Header.Set("Authorization", aks.authHeader)
	}
	resp := protocol.AcquireResponse()
	defer protocol.ReleaseResponse(resp)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = aks.hc.Do(ctx, req, resp)
	if err != nil {
		zap.L().Error("HTTP request failed", zap.Error(err))
		return nil, fmt.Errorf("request to auth service: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		zap.L().Error("Auth service returned error", zap.Int("status_code", resp.StatusCode()), zap.String("body", string(resp.Body())))
		return nil, fmt.Errorf("auth service error: %s", string(resp.Body()))
	}

	var wrapper dtos.ApiKeysUsageResponse
	if err := json.Unmarshal(resp.Body(), &wrapper); err != nil {
		zap.L().Error("Failed to unmarshal usage response", zap.Error(err))
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return wrapper.Keys, nil
}

func (aks *ApiKeysService) RemoveApiKey(ctx context.Context, apiKey string, devID string) error {
	req := protocol.AcquireRequest()
	defer protocol.ReleaseRequest(req)
	req.SetMethod(http.MethodDelete)
	req.SetRequestURI(aks.authURL + "/key")
	req.Header.Set("Content-Type", "application/json")
	if aks.authHeader != "" {
		req.Header.Set("Authorization", aks.authHeader)
	}

	devUUID, err := uuid.Parse(devID)
	if err != nil {
		return err
	}

	body := dtos.ApiKeyRemoveDto{
		Key:     apiKey,
		DevUUID: devUUID,
	}

	data, err := sonic.Marshal(body)
	if err != nil {
		return err
	}

	req.SetBody(data)

	resp := protocol.AcquireResponse()
	defer protocol.ReleaseResponse(resp)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = aks.hc.Do(ctx, req, resp)
	if err != nil {
		return fmt.Errorf("request to auth service: %w", err)
	}

	if resp.StatusCode() == http.StatusForbidden {
		return ErrAccessDenied
	}
	if bytes.Contains(bytes.ToLower(resp.Body()), []byte("access denied")) {
		return ErrAccessDenied
	}

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusNoContent {
		return fmt.Errorf("auth service error: %s", string(resp.Body()))
	}

	return nil
}
