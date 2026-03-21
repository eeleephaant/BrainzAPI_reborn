package handler

import (
	"brainz/common/dtos"
	"brainz/common/permissions"
	"brainz/developersapi/internal/entity"
	"brainz/developersapi/internal/services"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/google/uuid"
)

type mockKeysSession struct {
	validateFn func(ctx context.Context, token string, ipAddr string) (*entity.Session, error)
}

func (m *mockKeysSession) ValidateToken(ctx context.Context, token string, ipAddr string) (*entity.Session, error) {
	if m.validateFn == nil {
		return nil, errors.New("no session")
	}
	return m.validateFn(ctx, token, ipAddr)
}

type mockKeysAPI struct {
	getFn    func(ctx context.Context, developerID uuid.UUID) ([]dtos.ApiKeyShareModel, error)
	createFn func(ctx context.Context, issuer, developerID uuid.UUID, name string, perms []permissions.Permission, ipWhitelist []string) (*dtos.ApiKeyCreateResponse, error)
	removeFn func(ctx context.Context, apiKey, devID string) error
}

func (m *mockKeysAPI) GetApiKeys(ctx context.Context, developerID uuid.UUID) ([]dtos.ApiKeyShareModel, error) {
	if m.getFn == nil {
		return nil, nil
	}
	return m.getFn(ctx, developerID)
}

func (m *mockKeysAPI) CreateApiKey(ctx context.Context, issuer, developerID uuid.UUID, name string, perms []permissions.Permission, ipWhitelist []string) (*dtos.ApiKeyCreateResponse, error) {
	if m.createFn == nil {
		return nil, errors.New("not implemented")
	}
	return m.createFn(ctx, issuer, developerID, name, perms, ipWhitelist)
}

func (m *mockKeysAPI) RemoveApiKey(ctx context.Context, apiKey, devID string) error {
	if m.removeFn == nil {
		return nil
	}
	return m.removeFn(ctx, apiKey, devID)
}

type mockKeysUser struct {
	getByIDFn func(ctx context.Context, devID uuid.UUID) (*entity.DeveloperAccount, error)
}

func (m *mockKeysUser) GetById(ctx context.Context, devId uuid.UUID) (*entity.DeveloperAccount, error) {
	if m.getByIDFn == nil {
		return nil, errors.New("not implemented")
	}
	return m.getByIDFn(ctx, devId)
}

func newSessionCtx() *app.RequestContext {
	return ut.CreateUtRequestContext("GET", "/keys", nil, ut.Header{Key: "X-Forwarded-For", Value: "127.0.0.1"})
}

func TestKeysManagementHandler_GetApiKeys(t *testing.T) {
	ctx := context.Background()
	devID := uuid.New()

	t.Run("invalid_session_400", func(t *testing.T) {
		h := &KeysManagementHandler{
			ss: &mockKeysSession{
				validateFn: func(ctx context.Context, token string, ipAddr string) (*entity.Session, error) {
					return nil, errors.New("bad")
				},
			},
		}
		c := newSessionCtx()
		c.Request.Header.Set("X-Session-Token", "bad")
		h.GetApiKeys(ctx, c)
		if c.Response.StatusCode() != 400 {
			t.Fatalf("status = %d, want 400", c.Response.StatusCode())
		}
	})

	t.Run("service_error_500", func(t *testing.T) {
		h := &KeysManagementHandler{
			ss: &mockKeysSession{
				validateFn: func(ctx context.Context, token string, ipAddr string) (*entity.Session, error) {
					return &entity.Session{DeveloperID: devID}, nil
				},
			},
			ks: &mockKeysAPI{
				getFn: func(ctx context.Context, developerID uuid.UUID) ([]dtos.ApiKeyShareModel, error) {
					return nil, errors.New("db failed")
				},
			},
		}
		c := newSessionCtx()
		c.Request.Header.Set("X-Session-Token", "ok")
		h.GetApiKeys(ctx, c)
		if c.Response.StatusCode() != 500 {
			t.Fatalf("status = %d, want 500", c.Response.StatusCode())
		}
	})

	t.Run("success_200", func(t *testing.T) {
		h := &KeysManagementHandler{
			ss: &mockKeysSession{
				validateFn: func(ctx context.Context, token string, ipAddr string) (*entity.Session, error) {
					return &entity.Session{DeveloperID: devID}, nil
				},
			},
			ks: &mockKeysAPI{
				getFn: func(ctx context.Context, developerID uuid.UUID) ([]dtos.ApiKeyShareModel, error) {
					if developerID != devID {
						t.Fatalf("developerID = %v, want %v", developerID, devID)
					}
					return []dtos.ApiKeyShareModel{{Id: uuid.New()}}, nil
				},
			},
		}
		c := newSessionCtx()
		c.Request.Header.Set("X-Session-Token", "ok")
		h.GetApiKeys(ctx, c)
		if c.Response.StatusCode() != 200 {
			t.Fatalf("status = %d, want 200", c.Response.StatusCode())
		}
	})
}

func TestKeysManagementHandler_CreateApiKey(t *testing.T) {
	ctx := context.Background()
	devID := uuid.New()

	makePost := func(body any) *app.RequestContext {
		payload, _ := json.Marshal(body)
		return ut.CreateUtRequestContext("POST", "/key", &ut.Body{Body: bytes.NewReader(payload), Len: len(payload)},
			ut.Header{Key: "Content-Type", Value: "application/json"})
	}

	t.Run("invalid_session_400", func(t *testing.T) {
		h := &KeysManagementHandler{
			ss: &mockKeysSession{
				validateFn: func(ctx context.Context, token string, ipAddr string) (*entity.Session, error) {
					return nil, errors.New("bad")
				},
			},
		}
		c := makePost(map[string]string{"api_key_name": "k"})
		c.Request.Header.Set("X-Session-Token", "bad")
		h.CreateApiKey(ctx, c)
		if c.Response.StatusCode() != 400 {
			t.Fatalf("status = %d, want 400", c.Response.StatusCode())
		}
	})

	t.Run("invalid_body_400", func(t *testing.T) {
		h := &KeysManagementHandler{
			ss: &mockKeysSession{
				validateFn: func(ctx context.Context, token string, ipAddr string) (*entity.Session, error) {
					return &entity.Session{DeveloperID: devID}, nil
				},
			},
		}
		c := makePost(map[string]string{})
		c.Request.Header.Set("X-Session-Token", "ok")
		h.CreateApiKey(ctx, c)
		if c.Response.StatusCode() != 400 {
			t.Fatalf("status = %d, want 400", c.Response.StatusCode())
		}
	})

	t.Run("get_user_error_500", func(t *testing.T) {
		h := &KeysManagementHandler{
			ss: &mockKeysSession{
				validateFn: func(ctx context.Context, token string, ipAddr string) (*entity.Session, error) {
					return &entity.Session{DeveloperID: devID}, nil
				},
			},
			us: &mockKeysUser{
				getByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.DeveloperAccount, error) {
					return nil, errors.New("not found")
				},
			},
		}
		c := makePost(map[string]string{"api_key_name": "k"})
		c.Request.Header.Set("X-Session-Token", "ok")
		h.CreateApiKey(ctx, c)
		if c.Response.StatusCode() != 500 {
			t.Fatalf("status = %d, want 500", c.Response.StatusCode())
		}
	})

	t.Run("unsupported_role_403", func(t *testing.T) {
		h := &KeysManagementHandler{
			ss: &mockKeysSession{
				validateFn: func(ctx context.Context, token string, ipAddr string) (*entity.Session, error) {
					return &entity.Session{DeveloperID: devID}, nil
				},
			},
			us: &mockKeysUser{
				getByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.DeveloperAccount, error) {
					return &entity.DeveloperAccount{ID: devID, RoleId: 99}, nil
				},
			},
		}
		c := makePost(map[string]string{"api_key_name": "k"})
		c.Request.Header.Set("X-Session-Token", "ok")
		h.CreateApiKey(ctx, c)
		if c.Response.StatusCode() != 403 {
			t.Fatalf("status = %d, want 403", c.Response.StatusCode())
		}
	})

	t.Run("create_service_error_500", func(t *testing.T) {
		h := &KeysManagementHandler{
			ss: &mockKeysSession{
				validateFn: func(ctx context.Context, token string, ipAddr string) (*entity.Session, error) {
					return &entity.Session{DeveloperID: devID}, nil
				},
			},
			us: &mockKeysUser{
				getByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.DeveloperAccount, error) {
					return &entity.DeveloperAccount{ID: devID, RoleId: DeveloperRoleReadOnly}, nil
				},
			},
			ks: &mockKeysAPI{
				createFn: func(ctx context.Context, issuer, developerID uuid.UUID, name string, perms []permissions.Permission, ipWhitelist []string) (*dtos.ApiKeyCreateResponse, error) {
					return nil, errors.New("auth failed")
				},
			},
		}
		c := makePost(map[string]string{"api_key_name": "k"})
		c.Request.Header.Set("X-Session-Token", "ok")
		h.CreateApiKey(ctx, c)
		if c.Response.StatusCode() != 500 {
			t.Fatalf("status = %d, want 500", c.Response.StatusCode())
		}
	})

	t.Run("success_201", func(t *testing.T) {
		var gotIssuer, gotDev uuid.UUID
		var gotPerms []permissions.Permission
		h := &KeysManagementHandler{
			ss: &mockKeysSession{
				validateFn: func(ctx context.Context, token string, ipAddr string) (*entity.Session, error) {
					return &entity.Session{DeveloperID: devID}, nil
				},
			},
			us: &mockKeysUser{
				getByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.DeveloperAccount, error) {
					return &entity.DeveloperAccount{ID: devID, RoleId: DeveloperRoleReadWrite}, nil
				},
			},
			ks: &mockKeysAPI{
				createFn: func(ctx context.Context, issuer, developerID uuid.UUID, name string, perms []permissions.Permission, ipWhitelist []string) (*dtos.ApiKeyCreateResponse, error) {
					gotIssuer, gotDev = issuer, developerID
					gotPerms = perms
					return &dtos.ApiKeyCreateResponse{Name: name, ApiKey: "secret", ExpiresAt: time.Now().Add(24 * time.Hour)}, nil
				},
			},
		}
		payload := []byte(`{"api_key_name":"my-key","ip_whitelist":["1.2.3.4"]}`)
		c2 := ut.CreateUtRequestContext("POST", "/key", &ut.Body{Body: bytes.NewReader(payload), Len: len(payload)},
			ut.Header{Key: "Content-Type", Value: "application/json"})
		c2.Request.Header.Set("X-Session-Token", "ok")
		h.CreateApiKey(ctx, c2)
		if c2.Response.StatusCode() != 201 {
			t.Fatalf("status = %d, want 201", c2.Response.StatusCode())
		}
		if gotIssuer != devID || gotDev != devID {
			t.Fatalf("issuer=%v dev=%v, want both %v", gotIssuer, gotDev, devID)
		}
		if len(gotPerms) != 2 {
			t.Fatalf("perms = %d, want 2", len(gotPerms))
		}
	})
}

func TestKeysManagementHandler_DeleteApiKey(t *testing.T) {
	ctx := context.Background()
	devID := uuid.New()

	t.Run("invalid_session_400", func(t *testing.T) {
		h := &KeysManagementHandler{
			ss: &mockKeysSession{
				validateFn: func(ctx context.Context, token string, ipAddr string) (*entity.Session, error) {
					return nil, errors.New("bad")
				},
			},
		}
		c := ut.CreateUtRequestContext("DELETE", "/key?api_key=abc", nil)
		c.Request.Header.Set("X-Session-Token", "bad")
		h.DeleteApiKey(ctx, c)
		if c.Response.StatusCode() != 400 {
			t.Fatalf("status = %d, want 400", c.Response.StatusCode())
		}
	})

	t.Run("access_denied_403", func(t *testing.T) {
		h := &KeysManagementHandler{
			ss: &mockKeysSession{
				validateFn: func(ctx context.Context, token string, ipAddr string) (*entity.Session, error) {
					return &entity.Session{DeveloperID: devID}, nil
				},
			},
			ks: &mockKeysAPI{
				removeFn: func(ctx context.Context, apiKey, devIDStr string) error {
					return services.ErrAccessDenied
				},
			},
		}
		c := ut.CreateUtRequestContext("DELETE", "/key?api_key=abc", nil)
		c.Request.Header.Set("X-Session-Token", "ok")
		h.DeleteApiKey(ctx, c)
		if c.Response.StatusCode() != 403 {
			t.Fatalf("status = %d, want 403", c.Response.StatusCode())
		}
	})

	t.Run("success_200", func(t *testing.T) {
		h := &KeysManagementHandler{
			ss: &mockKeysSession{
				validateFn: func(ctx context.Context, token string, ipAddr string) (*entity.Session, error) {
					return &entity.Session{DeveloperID: devID}, nil
				},
			},
			ks: &mockKeysAPI{
				removeFn: func(ctx context.Context, apiKey, devIDStr string) error {
					if devIDStr != devID.String() {
						t.Fatalf("devID = %q", devIDStr)
					}
					return nil
				},
			},
		}
		c := ut.CreateUtRequestContext("DELETE", "/key?api_key=abc", nil)
		c.Request.Header.Set("X-Session-Token", "ok")
		h.DeleteApiKey(ctx, c)
		if c.Response.StatusCode() != 200 {
			t.Fatalf("status = %d, want 200", c.Response.StatusCode())
		}
	})
}
