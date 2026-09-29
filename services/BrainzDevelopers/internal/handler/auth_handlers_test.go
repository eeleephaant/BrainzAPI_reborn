package handler

import (
	"brainz/developersapi/internal/dtos"
	"brainz/developersapi/internal/entity"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/google/uuid"
)

type mockAuthUserService struct {
	authenticate   func(ctx context.Context, email, password string) (*entity.DeveloperAccount, error)
	registrateUser func(ctx context.Context, registerDto *dtos.RegisterDto) (*entity.EmailConfirmationToken, error)
	getByID        func(ctx context.Context, devID uuid.UUID) (*entity.DeveloperAccount, error)
}

func (m *mockAuthUserService) Authenticate(ctx context.Context, email, password string) (*entity.DeveloperAccount, error) {
	if m.authenticate != nil {
		return m.authenticate(ctx, email, password)
	}
	return nil, entity.ErrWrongCredentials
}

func (m *mockAuthUserService) RegistrateUser(ctx context.Context, registerDto *dtos.RegisterDto) (*entity.EmailConfirmationToken, error) {
	if m.registrateUser != nil {
		return m.registrateUser(ctx, registerDto)
	}
	return nil, errors.New("not implemented")
}

func (m *mockAuthUserService) GetById(ctx context.Context, devID uuid.UUID) (*entity.DeveloperAccount, error) {
	if m.getByID != nil {
		return m.getByID(ctx, devID)
	}
	return nil, errors.New("not implemented")
}

type mockAuthSessionService struct {
	createNew     func(ctx context.Context, user *entity.DeveloperAccount, userAgent, ipAddr string) (*entity.Session, *string, error)
	validateToken func(ctx context.Context, token string, ipAddr string) (*entity.Session, error)
}

func (m *mockAuthSessionService) CreateNew(ctx context.Context, user *entity.DeveloperAccount, userAgent, ipAddr string) (*entity.Session, *string, error) {
	if m.createNew != nil {
		return m.createNew(ctx, user, userAgent, ipAddr)
	}
	tok := "test-token"
	return &entity.Session{ID: uuid.New(), DeveloperID: user.ID}, &tok, nil
}

func (m *mockAuthSessionService) ValidateToken(ctx context.Context, token string, ipAddr string) (*entity.Session, error) {
	if m.validateToken != nil {
		return m.validateToken(ctx, token, ipAddr)
	}
	return nil, errors.New("invalid session")
}

func TestAuthHandler_Login(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid_request_400", func(t *testing.T) {
		us := &mockAuthUserService{}
		ss := &mockAuthSessionService{}
		h := NewAuthHandler(us, ss, SessionHTTPConfig{})
		body := []byte(`{}`)
		c := ut.CreateUtRequestContext("POST", "/login", &ut.Body{Body: bytes.NewReader(body), Len: len(body)},
			ut.Header{Key: "Content-Type", Value: "application/json"})
		h.Login(ctx, c)
		if c.Response.StatusCode() != 400 {
			t.Errorf("status = %d, want 400", c.Response.StatusCode())
		}
	})

	t.Run("wrong_credentials_401", func(t *testing.T) {
		us := &mockAuthUserService{
			authenticate: func(ctx context.Context, email, password string) (*entity.DeveloperAccount, error) {
				return nil, entity.ErrWrongCredentials
			},
		}
		ss := &mockAuthSessionService{}
		h := NewAuthHandler(us, ss, SessionHTTPConfig{})
		body, _ := json.Marshal(map[string]string{"email": "a@b.com", "password": "password123456"})
		c := ut.CreateUtRequestContext("POST", "/login", &ut.Body{Body: bytes.NewReader(body), Len: len(body)},
			ut.Header{Key: "Content-Type", Value: "application/json"})
		h.Login(ctx, c)
		if c.Response.StatusCode() != 401 {
			t.Errorf("status = %d, want 401", c.Response.StatusCode())
		}
	})

	t.Run("email_not_confirmed_401", func(t *testing.T) {
		us := &mockAuthUserService{
			authenticate: func(ctx context.Context, email, password string) (*entity.DeveloperAccount, error) {
				return nil, entity.ErrEmailNotConfirmed
			},
		}
		ss := &mockAuthSessionService{}
		h := NewAuthHandler(us, ss, SessionHTTPConfig{})
		body, _ := json.Marshal(map[string]string{"email": "a@b.com", "password": "password123456"})
		c := ut.CreateUtRequestContext("POST", "/login", &ut.Body{Body: bytes.NewReader(body), Len: len(body)},
			ut.Header{Key: "Content-Type", Value: "application/json"})
		h.Login(ctx, c)
		if c.Response.StatusCode() != 401 {
			t.Errorf("status = %d, want 401", c.Response.StatusCode())
		}
	})

	t.Run("success_200", func(t *testing.T) {
		acc := &entity.DeveloperAccount{ID: uuid.New(), Email: "ok@test.com"}
		us := &mockAuthUserService{
			authenticate: func(ctx context.Context, email, password string) (*entity.DeveloperAccount, error) {
				return acc, nil
			},
		}
		tok := "session-token-123"
		ss := &mockAuthSessionService{
			createNew: func(ctx context.Context, user *entity.DeveloperAccount, userAgent, ipAddr string) (*entity.Session, *string, error) {
				return &entity.Session{ID: uuid.New(), DeveloperID: user.ID}, &tok, nil
			},
		}
		h := NewAuthHandler(us, ss, SessionHTTPConfig{})
		body, _ := json.Marshal(map[string]string{"email": "ok@test.com", "password": "password123456"})
		c := ut.CreateUtRequestContext("POST", "/login", &ut.Body{Body: bytes.NewReader(body), Len: len(body)},
			ut.Header{Key: "Content-Type", Value: "application/json"})
		h.Login(ctx, c)
		if c.Response.StatusCode() != 200 {
			t.Errorf("status = %d, want 200", c.Response.StatusCode())
		}
		if !bytes.Contains(c.Response.Body(), []byte("session-token-123")) {
			t.Errorf("body should contain token, got %s", c.Response.Body())
		}
		setCookie := string(c.Response.Header.Peek("Set-Cookie"))
		if !strings.Contains(setCookie, "brainz_session=") || !strings.Contains(setCookie, "session-token-123") {
			t.Errorf("Set-Cookie should include HttpOnly session cookie, got %q", setCookie)
		}
	})
}

func TestAuthHandler_GetProfile(t *testing.T) {
	ctx := context.Background()
	devID := uuid.New()
	created := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)

	t.Run("invalid_session_400", func(t *testing.T) {
		h := NewAuthHandler(&mockAuthUserService{}, &mockAuthSessionService{
			validateToken: func(ctx context.Context, token string, ipAddr string) (*entity.Session, error) {
				return nil, errors.New("bad")
			},
		}, SessionHTTPConfig{})
		c := ut.CreateUtRequestContext("GET", "/profile", nil,
			ut.Header{Key: "X-Forwarded-For", Value: "127.0.0.1"})
		c.Request.Header.Set("X-Session-Token", "bad")
		h.GetProfile(ctx, c)
		if c.Response.StatusCode() != 400 {
			t.Fatalf("status = %d, want 400", c.Response.StatusCode())
		}
	})

	t.Run("get_user_error_500", func(t *testing.T) {
		h := NewAuthHandler(&mockAuthUserService{
			getByID: func(ctx context.Context, id uuid.UUID) (*entity.DeveloperAccount, error) {
				return nil, errors.New("db")
			},
		}, &mockAuthSessionService{
			validateToken: func(ctx context.Context, token string, ipAddr string) (*entity.Session, error) {
				return &entity.Session{DeveloperID: devID}, nil
			},
		}, SessionHTTPConfig{})
		c := ut.CreateUtRequestContext("GET", "/profile", nil,
			ut.Header{Key: "X-Forwarded-For", Value: "127.0.0.1"})
		c.Request.Header.Set("X-Session-Token", "ok")
		h.GetProfile(ctx, c)
		if c.Response.StatusCode() != 500 {
			t.Fatalf("status = %d, want 500", c.Response.StatusCode())
		}
	})

	t.Run("success_200", func(t *testing.T) {
		secret := []byte{1}
		acc := &entity.DeveloperAccount{
			ID:               devID,
			Email:            "p@test.com",
			EmailConfirmedAt: nil,
			CreatedAt:        created,
			RoleId:           1,
			BannedAt:         nil,
			TwoFactorSecret:  &secret,
		}
		h := NewAuthHandler(&mockAuthUserService{
			getByID: func(ctx context.Context, id uuid.UUID) (*entity.DeveloperAccount, error) {
				if id != devID {
					t.Fatalf("dev id = %v, want %v", id, devID)
				}
				return acc, nil
			},
		}, &mockAuthSessionService{
			validateToken: func(ctx context.Context, token string, ipAddr string) (*entity.Session, error) {
				return &entity.Session{DeveloperID: devID}, nil
			},
		}, SessionHTTPConfig{})
		c := ut.CreateUtRequestContext("GET", "/profile", nil,
			ut.Header{Key: "X-Forwarded-For", Value: "127.0.0.1"})
		c.Request.Header.Set("X-Session-Token", "ok")
		h.GetProfile(ctx, c)
		if c.Response.StatusCode() != 200 {
			t.Fatalf("status = %d, want 200", c.Response.StatusCode())
		}
		var out struct {
			ID               string `json:"id"`
			Email            string `json:"email"`
			RoleId           uint   `json:"role_id"`
			TwoFactorEnabled bool   `json:"two_factor_enabled"`
			CreatedAt        string `json:"created_at"`
		}
		if err := json.Unmarshal(c.Response.Body(), &out); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if out.ID != devID.String() || out.Email != "p@test.com" || out.RoleId != 1 || !out.TwoFactorEnabled {
			t.Fatalf("body = %+v", out)
		}
	})
}

func TestAuthHandler_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid_request_400", func(t *testing.T) {
		us := &mockAuthUserService{}
		ss := &mockAuthSessionService{}
		h := NewAuthHandler(us, ss, SessionHTTPConfig{})
		body := []byte(`{"email":"bad"}`) // short password
		c := ut.CreateUtRequestContext("POST", "/register", &ut.Body{Body: bytes.NewReader(body), Len: len(body)},
			ut.Header{Key: "Content-Type", Value: "application/json"})
		h.Register(ctx, c)
		if c.Response.StatusCode() != 400 {
			t.Errorf("status = %d, want 400", c.Response.StatusCode())
		}
	})

	t.Run("email_already_exists_409", func(t *testing.T) {
		us := &mockAuthUserService{
			registrateUser: func(ctx context.Context, dto *dtos.RegisterDto) (*entity.EmailConfirmationToken, error) {
				return nil, entity.ErrEmailAlreadyExists
			},
		}
		ss := &mockAuthSessionService{}
		h := NewAuthHandler(us, ss, SessionHTTPConfig{})
		body, _ := json.Marshal(map[string]string{"email": "exists@test.com", "password": "securepassword123"})
		c := ut.CreateUtRequestContext("POST", "/register", &ut.Body{Body: bytes.NewReader(body), Len: len(body)},
			ut.Header{Key: "Content-Type", Value: "application/json"})
		h.Register(ctx, c)
		if c.Response.StatusCode() != 409 {
			t.Errorf("status = %d, want 409", c.Response.StatusCode())
		}
	})

	t.Run("success_201", func(t *testing.T) {
		us := &mockAuthUserService{
			registrateUser: func(ctx context.Context, dto *dtos.RegisterDto) (*entity.EmailConfirmationToken, error) {
				return &entity.EmailConfirmationToken{
					ID: uuid.New(), DeveloperID: uuid.New(), Token: "confirm-token",
					NumbericCode: "123456", ExpiresAt: time.Now().Add(30 * time.Minute), CreatedAt: time.Now(),
				}, nil
			},
		}
		ss := &mockAuthSessionService{}
		h := NewAuthHandler(us, ss, SessionHTTPConfig{})
		body, _ := json.Marshal(map[string]string{"email": "new@test.com", "password": "securepassword123"})
		c := ut.CreateUtRequestContext("POST", "/register", &ut.Body{Body: bytes.NewReader(body), Len: len(body)},
			ut.Header{Key: "Content-Type", Value: "application/json"})
		h.Register(ctx, c)
		if c.Response.StatusCode() != 201 {
			t.Errorf("status = %d, want 201", c.Response.StatusCode())
		}
	})
}
