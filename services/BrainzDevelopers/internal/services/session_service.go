package services

import (
	"brainz/developersapi/internal/entity"
	"brainz/developersapi/internal/repository"
	"brainz/developersapi/internal/security"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SessionService struct {
	sr *repository.SessionRepository
}

func NewSessionService(sr *repository.SessionRepository) *SessionService {
	return &SessionService{sr: sr}
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
