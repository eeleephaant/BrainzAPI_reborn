package services

import (
	"brainz/developersapi/internal/config"
	"brainz/developersapi/internal/entity"
	"brainz/developersapi/internal/security"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func init() {
	security.LoadCfg(&config.AuthConfig{
		Pepper:        "test-pepper",
		Argon2Memory:  64 * 1024,
		Argon2Time:    1,
		Argon2Threads: 2,
	})
}

type mockSessionRepo struct {
	sessions map[uuid.UUID]*entity.Session
	create   func(ctx context.Context, s *entity.Session) (*entity.Session, error)
}

func (m *mockSessionRepo) GetByID(ctx context.Context, sessionID uuid.UUID) (*entity.Session, error) {
	if m.sessions != nil {
		if s, ok := m.sessions[sessionID]; ok {
			return s, nil
		}
	}
	return nil, fmt.Errorf("session not found")
}

func (m *mockSessionRepo) CreateSession(ctx context.Context, s *entity.Session) (*entity.Session, error) {
	if m.create != nil {
		return m.create(ctx, s)
	}
	if m.sessions == nil {
		m.sessions = make(map[uuid.UUID]*entity.Session)
	}
	m.sessions[s.ID] = s
	return s, nil
}

func TestSessionService_ValidateToken(t *testing.T) {
	ctx := context.Background()
	tokenRaw, _ := security.GenerateSecretKey(32)
	salt := security.GetRandomSalt()
	tokenHash := security.GetHashArgon2(tokenRaw, salt)
	sessionID := uuid.New()
	expiresAt := time.Now().Add(time.Hour)

	t.Run("valid_token", func(t *testing.T) {
		repo := &mockSessionRepo{
			sessions: map[uuid.UUID]*entity.Session{
				sessionID: {
					ID:          sessionID,
					IpAddress:   "192.168.1.1",
					ExpiresAt:   expiresAt,
					Salt:        salt,
					TokenHash:   tokenHash,
				},
			},
		}
		svc := NewSessionService(repo)
		token := sessionID.String() + ":" + tokenRaw
		sess, err := svc.ValidateToken(ctx, token, "192.168.1.1")
		if err != nil {
			t.Fatalf("ValidateToken: %v", err)
		}
		if sess == nil || sess.ID != sessionID {
			t.Errorf("expected session %v, got %+v", sessionID, sess)
		}
	})

	t.Run("wrong_ip", func(t *testing.T) {
		repo := &mockSessionRepo{
			sessions: map[uuid.UUID]*entity.Session{
				sessionID: {
					ID:        sessionID,
					IpAddress: "192.168.1.1",
					ExpiresAt: expiresAt,
					Salt:      salt,
					TokenHash: tokenHash,
				},
			},
		}
		svc := NewSessionService(repo)
		token := sessionID.String() + ":" + tokenRaw
		_, err := svc.ValidateToken(ctx, token, "10.0.0.1")
		if err == nil {
			t.Fatal("expected error on IP mismatch")
		}
	})

	t.Run("invalid_token_format", func(t *testing.T) {
		svc := NewSessionService(&mockSessionRepo{})
		_, err := svc.ValidateToken(ctx, "bad-token", "1.2.3.4")
		if err == nil {
			t.Fatal("expected error for invalid token format")
		}
	})
}

func TestSessionService_CreateNew(t *testing.T) {
	ctx := context.Background()
	user := &entity.DeveloperAccount{ID: uuid.New(), Email: "dev@test.com"}

	t.Run("success", func(t *testing.T) {
		repo := &mockSessionRepo{}
		svc := NewSessionService(repo)
		sess, tokenPtr, err := svc.CreateNew(ctx, user, "test-agent", "127.0.0.1")
		if err != nil {
			t.Fatalf("CreateNew: %v", err)
		}
		if sess == nil {
			t.Fatal("expected non-nil session")
		}
		if tokenPtr == nil || *tokenPtr == "" {
			t.Fatal("expected non-empty token")
		}
		if sess.DeveloperID != user.ID {
			t.Errorf("session.DeveloperID = %v, want %v", sess.DeveloperID, user.ID)
		}
		// Token format: uuid:raw
		if len(*tokenPtr) < 36+1+10 {
			t.Errorf("token too short: %q", *tokenPtr)
		}
	})
}
