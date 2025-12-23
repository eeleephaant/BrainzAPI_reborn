package repository

import (
	"brainz/developersapi/internal/entity"
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailTokensRepository struct {
	db *pgxpool.Pool
}

func NewEmailTokensRepository(db *pgxpool.Pool) *EmailTokensRepository {
	return &EmailTokensRepository{db}
}

func (ur *EmailTokensRepository) Create(ctx context.Context, emailConfToken *entity.EmailConfirmationToken) (*entity.EmailConfirmationToken, error) {
	op := "EmailTokensRepository.Create"
	conn, err := ur.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: acquire connection: %w", op, err)
	}
	defer conn.Release()

	queryBuilder := sq.
		Insert("email_confirmation_token").
		Columns(
			"developer_id",
			"token",
			"numberic_code",
			"expires_at",
		).
		Values(
			emailConfToken.DeveloperID,
			emailConfToken.Token,
			emailConfToken.ExpiresAt,
		).
		Suffix(`
			RETURNING
				id, developer_id, token, expires_at, used_at, created_at
		`).
		PlaceholderFormat(sq.Dollar)

	sqlQuery, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: build query: %w", op, err)
	}
	var created entity.EmailConfirmationToken
	row := conn.QueryRow(ctx, sqlQuery, args...)
	err = row.Scan(
		&created.ID,
		&created.DeveloperID,
		&created.Token,
		&created.ExpiresAt,
		&created.UsedAt,
		&created.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: scan row: %w", op, err)
	}
	return &created, nil
}

func (ur *EmailTokensRepository) Update(ctx context.Context, emailConfToken *entity.EmailConfirmationToken) (*entity.EmailConfirmationToken, error) {
	op := "EmailTokensRepository.Update"
	conn, err := ur.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: acquire connection: %w", op, err)
	}
	defer conn.Release()
	queryBuilder := sq.
		Update("email_confirmation_token").
		Set("developer_id", emailConfToken.DeveloperID).
		Set("token", emailConfToken.Token).
		Set("expires_at", emailConfToken.ExpiresAt).
		Set("used_at", emailConfToken.UsedAt).
		Where(sq.Eq{"id": emailConfToken.ID}).
		Suffix(`
			RETURNING
				id, developer_id, token, expires_at, used_at, created_at
		`).
		PlaceholderFormat(sq.Dollar)

	sqlQuery, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: build query: %w", op, err)
	}

	var updated entity.EmailConfirmationToken
	row := conn.QueryRow(ctx, sqlQuery, args...)
	err = row.Scan(
		&updated.ID,
		&updated.DeveloperID,
		&updated.Token,
		&updated.ExpiresAt,
		&updated.UsedAt,
		&updated.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: scan row: %w", op, err)
	}
	return &updated, nil
}
