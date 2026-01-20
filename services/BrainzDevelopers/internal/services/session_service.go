package services

import (
	"brainz/developersapi/internal/entity"
	"brainz/developersapi/internal/repository"
	"brainz/developersapi/internal/security"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SessionService struct {
	sr *repository.SessionRepository
}

func NewSessionService(sr *repository.SessionRepository) *SessionService {
	return &SessionService{sr: sr}
}

func (ss *SessionService) ValidateToken(ctx context.Context, token string, ipAddr string) (*entity.Session, error) {
	devUUID, tokenRaw, err := security.ExtractDataFromToken(token)
	if err != nil {
		return nil, fmt.Errorf("extract data from token: %w", err)
	}

	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid api key format")
	}

	sessionID, err := uuid.Parse(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid api key")
	}

	session, err := ss.sr.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session by ID: %w", err)
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, nil
	}

	if ipAddr != session.IpAddress {
		return nil, fmt.Errorf("ip address mismatch")
	}
	return nil, nil
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
