package services

import (
	"brainz-api/internal/dtos"
	cmnDtos "brainz/common/dtos"
	"brainz/common/permissions"
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type AuthService struct {
	baseURL *url.URL
}

func NewAuthService(baseURL *url.URL) *AuthService {
	return &AuthService{
		baseURL: baseURL,
	}
}

func (ps *AuthService) Authorize(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
	if apiKey == "" {
		// Missing API key: treat as unauthorized without reporting internal error.
		return false, nil
	}

	u := *ps.baseURL
	// Base URL is often configured as http://brainz-auth:8080 without path; Auth expects POST /auth.
	if u.Path == "" || u.Path == "/" {
		u.Path = "/auth"
	}
	req := &protocol.Request{}
	req.SetRequestURI(u.String())
	req.SetMethod(consts.MethodPost)

	bodyBytes, err := sonic.Marshal(cmnDtos.RequestedPermissionDTO{Permission: *perm})
	if err != nil {
		return false, fmt.Errorf("cannot marshal permissions: %w", err)
	}

	req.SetBody(bodyBytes)
	req.Header.SetContentTypeBytes([]byte("application/json"))
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("X-Original-IP", reqCtx.ClientIP())
	req.Header.Set("X-Original-Path", string(reqCtx.FullPath()))
	req.Header.Set("X-Original-Method", string(reqCtx.Method()))

	resp := &protocol.Response{}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = client.Do(ctx, req, resp)
	if err != nil {
		return false, err
	}
	defer func() {
		protocol.ReleaseResponse(resp)
		protocol.ReleaseRequest(req)
	}()

	if resp.StatusCode() != consts.StatusOK {
		return false, nil
	}

	body := resp.Body()
	if len(body) == 0 {
		return false, fmt.Errorf("failed to decode auth response: empty body")
	}

	var authResp dtos.AuthResult
	if err := sonic.Unmarshal(body, &authResp); err != nil {
		return false, fmt.Errorf("failed to decode auth response: %w", err)
	}

	return authResp.Status, nil
}
