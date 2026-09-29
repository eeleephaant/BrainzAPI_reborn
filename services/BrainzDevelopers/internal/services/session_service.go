package services

import (
	"brainz/developersapi/internal/entity"
	"brainz/developersapi/internal/security"
	"context"
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SessionRepository defines persistence for sessions (for testing and DI).
type SessionRepository interface {
	GetByID(ctx context.Context, sessionID uuid.UUID) (*entity.Session, error)
	CreateSession(ctx context.Context, s *entity.Session) (*entity.Session, error)
}

type SessionService struct {
	sr SessionRepository
}

func NewSessionService(sr SessionRepository) *SessionService {
	return &SessionService{sr: sr}
}

func (ss *SessionService) ValidateToken(ctx context.Context, token string, ipAddr string) (*entity.Session, error) {
	tokenUUID, tokenRaw, err := security.ExtractDataFromToken(token)
	if err != nil {
		return nil, fmt.Errorf("extract data from token: %w", err)
	}

	session, err := ss.sr.GetByID(ctx, tokenUUID)
	if err != nil {
		return nil, fmt.Errorf("get session by ID: %w", err)
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	if ipAddr != session.IpAddress {
		return nil, fmt.Errorf("ip address mismatch")
	}

	userKeyHash := security.GetHashArgon2(tokenRaw, session.Salt)
	if subtle.ConstantTimeCompare(session.TokenHash, userKeyHash) != 1 {
		return nil, fmt.Errorf("invalid api key")
	}
	return session, nil
}

func (ss *SessionService) CreateNew(ctx context.Context, user *entity.DeveloperAccount, userAgent string, ipAddr string) (*entity.Session, *string, error) {
	tokenRaw, err := security.GenerateSecretKey(32)
	if err != nil {
		return nil, nil, err
	}

	salt := security.GetRandomSalt()
	tokenHash := security.GetHashArgon2(tokenRaw, salt)

	session := entity.Session{
		ID:          uuid.New(),
		DeveloperID: user.ID,
		UserAgent:   userAgent,
		IpAddress:   ipAddr,
		ExpiresAt:   time.Now().Add(time.Hour * 24 * 14),
		TokenHash:   tokenHash,
		Salt:        salt,
	}
	createdSession, err := ss.sr.CreateSession(ctx, &session)
	if err != nil {
		return nil, nil, err
	}
	displayedToken := fmt.Sprintf("%s:%s", session.ID.String(), tokenRaw)

	return createdSession, &displayedToken, nil
}
