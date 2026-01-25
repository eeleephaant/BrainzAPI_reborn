package services

import (
	"brainz-api/internal/dtos"
	"brainz/common/permissions"
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type AuthService struct {
	client  *http.Client
	baseURL *url.URL
}

func NewAuthService(baseURL *url.URL) *AuthService {
	return &AuthService{
		baseURL: baseURL,
	}
}

func (ps *AuthService) Authorize(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
	if apiKey == "" {
		return false, fmt.Errorf("api key is required")
	}

	u := *ps.baseURL
	q := u.Query()
	q.Set("key", apiKey)
	u.RawQuery = q.Encode()

	req := &protocol.Request{}
	req.SetRequestURI(u.String())
	req.SetMethod(consts.MethodGet)

	bodyBytes, err := sonic.Marshal(perm)
	if err != nil {
		return false, fmt.Errorf("cannot marshal permissions: %w", err)
	}

	req.SetBody(bodyBytes)

	req.Header.Set("X-Original-IP", reqCtx.ClientIP())
	req.Header.Set("X-Original-Path", string(reqCtx.FullPath()))
	req.Header.Set("X-Original-Method", string(reqCtx.Method()))

	resp := &protocol.Response{}

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

	var authResp dtos.AuthResult
	if err := sonic.Unmarshal(body, &authResp); err != nil {
		return false, fmt.Errorf("failed to decode auth response: %w", err)
	}

	return authResp.Status, nil
}
