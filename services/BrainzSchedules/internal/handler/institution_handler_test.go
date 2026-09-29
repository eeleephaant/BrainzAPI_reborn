package handler

import (
	"brainz-api/internal/dtos"
	"brainz/common/permissions"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/ut"
)

type mockInstitutionService struct {
	createNewFn       func(ctx context.Context, instDto dtos.InstitutionCreateDto) (int64, error)
	getInstitutionsFn func(ctx context.Context) ([]dtos.InstitutionDto, error)
}

func (m *mockInstitutionService) CreateNew(ctx context.Context, instDto dtos.InstitutionCreateDto) (int64, error) {
	return m.createNewFn(ctx, instDto)
}

func (m *mockInstitutionService) GetInstitutions(ctx context.Context) ([]dtos.InstitutionDto, error) {
	return m.getInstitutionsFn(ctx)
}

func TestInstitutionHandler_CreateInstitution(t *testing.T) {
	ctx := context.Background()

	makeCtx := func(body any) *app.RequestContext {
		payload, _ := json.Marshal(body)
		return ut.CreateUtRequestContext("POST", "/institution", &ut.Body{Body: bytes.NewReader(payload), Len: len(payload)},
			ut.Header{Key: "Content-Type", Value: "application/json"})
	}

	t.Run("authorize_error", func(t *testing.T) {
		h := &InstitutionHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return false, errors.New("auth failed")
				},
			},
		}
		c := makeCtx(map[string]string{"name": "Brainz", "site_link": "https://brainz.dev"})
		h.CreateInstitution(ctx, c)
		if c.Response.StatusCode() != 500 {
			t.Fatalf("status = %d, want 500", c.Response.StatusCode())
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		h := &InstitutionHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return false, nil
				},
			},
		}
		c := makeCtx(map[string]string{"name": "Brainz", "site_link": "https://brainz.dev"})
		h.CreateInstitution(ctx, c)
		if c.Response.StatusCode() != 403 {
			t.Fatalf("status = %d, want 403", c.Response.StatusCode())
		}
	})

	t.Run("invalid_body", func(t *testing.T) {
		h := &InstitutionHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return true, nil
				},
			},
		}
		c := makeCtx(map[string]string{"name": string(bytes.Repeat([]byte("a"), 101))})
		h.CreateInstitution(ctx, c)
		if c.Response.StatusCode() != 400 {
			t.Fatalf("status = %d, want 400", c.Response.StatusCode())
		}
	})

	t.Run("service_error", func(t *testing.T) {
		h := &InstitutionHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return true, nil
				},
			},
			is: &mockInstitutionService{
				createNewFn: func(ctx context.Context, instDto dtos.InstitutionCreateDto) (int64, error) {
					return 0, errors.New("create failed")
				},
			},
		}
		c := makeCtx(map[string]string{"name": "Brainz", "site_link": "https://brainz.dev"})
		h.CreateInstitution(ctx, c)
		if c.Response.StatusCode() != 500 {
			t.Fatalf("status = %d, want 500", c.Response.StatusCode())
		}
	})

	t.Run("success", func(t *testing.T) {
		h := &InstitutionHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return true, nil
				},
			},
			is: &mockInstitutionService{
				createNewFn: func(ctx context.Context, instDto dtos.InstitutionCreateDto) (int64, error) {
					return 99, nil
				},
			},
		}
		c := makeCtx(map[string]string{"name": "Brainz", "site_link": "https://brainz.dev"})
		h.CreateInstitution(ctx, c)
		if c.Response.StatusCode() != 201 {
			t.Fatalf("status = %d, want 201", c.Response.StatusCode())
		}
	})
}

func TestInstitutionHandler_GetInstitutions(t *testing.T) {
	ctx := context.Background()

	t.Run("service_error", func(t *testing.T) {
		h := &InstitutionHandler{
			is: &mockInstitutionService{
				getInstitutionsFn: func(ctx context.Context) ([]dtos.InstitutionDto, error) {
					return nil, errors.New("list failed")
				},
			},
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return true, nil
				},
			},
		}
		c := ut.CreateUtRequestContext("GET", "/institution", nil)
		h.GetInstitutions(ctx, c)
		if c.Response.StatusCode() != 500 {
			t.Fatalf("status = %d, want 500", c.Response.StatusCode())
		}
	})

	t.Run("success", func(t *testing.T) {
		h := &InstitutionHandler{
			is: &mockInstitutionService{
				getInstitutionsFn: func(ctx context.Context) ([]dtos.InstitutionDto, error) {
					return []dtos.InstitutionDto{{Id: 1, Name: "Brainz", Site: "https://brainz.dev"}}, nil
				},
			},
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return true, nil
				},
			},
		}
		c := ut.CreateUtRequestContext("GET", "/institution", nil)
		h.GetInstitutions(ctx, c)
		if c.Response.StatusCode() != 200 {
			t.Fatalf("status = %d, want 200", c.Response.StatusCode())
		}
		if !bytes.Contains(c.Response.Body(), []byte("Brainz")) {
			t.Fatalf("body = %s, want Brainz", c.Response.Body())
		}
	})
}
