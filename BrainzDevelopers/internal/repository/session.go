package repository

import (
	"brainz/developersapi/internal/models"
	"brainz/developersapi/internal/security"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct {
	db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{db}
}

func (r *SessionRepository) GetNewSession(ctx context.Context, request *models.SessionRequestData) ([]byte, error) {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()
	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	sessionID := uuid.New().String()
	secretKeyRaw, err := security.GenerateSecretKey(128)
	if err != nil {
		return nil, fmt.Errorf("not able to create secret key: %w", err)
	}
	salt := security.GetRandomSalt()
	secretKeyHash := security.GetHashArgon2(secretKeyRaw, salt)
	expiresAt := time.Now().Add(24 * 14 * time.Hour)

	_, err = tx.Exec(ctx,
		`INSERT INTO developer_session
		(id, developer_id, user_agent, ip_address, token_hash, salt, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		sessionID, request.DeveloperId, request.UserAgent, request.IpAddress, secretKeyHash, salt, expiresAt)

	if err != nil {
		return nil, fmt.Errorf("not able to create secret key")
	}
	outputKey := fmt.Sprintf("%s:%s", sessionID, secretKeyRaw)
	return []byte(outputKey), nil
}

func (r *SessionRepository) ValidateSessionKey(ctx context.Context, authRequest *models.SessionAuthData) (bool, error) {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()
	tx, err := conn.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	parts := strings.SplitN(authRequest.ApiKeyRaw, ":", 2)
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid api key format")
	}

	sessionID, err := uuid.Parse(parts[0])
	if err != nil {
		return false, fmt.Errorf("invalid api key")
	}

	var expiresAt time.Time
	var IpAddress string

	err = tx.QueryRow(ctx, "SELECT expires_at, ip_address FROM developer_session WHERE id=$1", sessionID).Scan(&expiresAt, &IpAddress)
	if err != nil {
		return false, nil
	}

	if time.Now().After(expiresAt) {
		return false, nil
	}

	if IpAddress != authRequest.IpAddress {
		return false, nil
	}
	return true, nil
}
