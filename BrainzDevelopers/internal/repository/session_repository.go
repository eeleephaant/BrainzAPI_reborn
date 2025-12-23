package repository

import (
	"brainz/developersapi/internal/entity"
	"brainz/developersapi/internal/models"
	"context"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct {
	db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{db}
}

func (r *SessionRepository) CreateSession(
	ctx context.Context, s *entity.Session,
) (*entity.Session, error) {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("session repository: acquire: %w", err)
	}
	defer conn.Release()

	queryBuilder := sq.
		Insert("developer_session").
		Columns(
			"id",
			"developer_id",
			"user_agent",
			"ip_address",
			"token_hash",
			"salt",
			"expires_at",
		).
		Values(
			s.ID,
			s.DeveloperID,
			s.UserAgent,
			s.IpAddress,
			s.TokenHash,
			s.Salt,
			s.ExpiresAt,
		).
		Suffix(`
			RETURNING
				id,
				developer_id,
				user_agent,
				ip_address,
				token_hash,
				salt,
				expires_at
		`)

	sqlStr, args, err := queryBuilder.
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("session repository: build sql: %w", err)
	}

	row := conn.QueryRow(ctx, sqlStr, args...)

	created := &entity.Session{}
	err = row.Scan(
		&created.ID,
		&created.DeveloperID,
		&created.UserAgent,
		&created.IpAddress,
		&created.TokenHash,
		&created.Salt,
		&created.ExpiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("session repository: scan: %w", err)
	}

	return created, nil
}

func (r *SessionRepository) Update(ctx context.Context, session *entity.Session) (*entity.Session, error) {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("session repository: acquire: %w", err)
	}
	defer conn.Release()

	queryBuilder := sq.
		Update("developer_session").
		Set("developer_id", session.DeveloperID).
		Set("user_agent", session.UserAgent).
		Set("ip_address", session.IpAddress).
		Set("token_hash", session.TokenHash).
		Set("salt", session.Salt).
		Set("expires_at", session.ExpiresAt).
		Set("revoked_at", session.Revoked_at).
		Where(sq.Eq{"id": session.ID}).
		Suffix(`
			RETURNING
				id,
				developer_id,
				user_agent,
				ip_address,
				token_hash,
				salt,
				expires_at,
				revoked_at
		`)

	sqlStr, args, err := queryBuilder.
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("session repository: build sql: %w", err)
	}

	row := conn.QueryRow(ctx, sqlStr, args...)

	updated := &entity.Session{}
	if err := row.Scan(
		&updated.ID,
		&updated.DeveloperID,
		&updated.UserAgent,
		&updated.IpAddress,
		&updated.TokenHash,
		&updated.Salt,
		&updated.ExpiresAt,
		&updated.Revoked_at,
	); err != nil {
		return nil, fmt.Errorf("session repository: scan: %w", err)
	}

	return updated, nil
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
