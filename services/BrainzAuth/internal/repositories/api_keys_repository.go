package repositories

import (
	"brainz/auth/internal/models"
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ApiKeysRepository struct {
	db *pgxpool.Pool
}

func NewApiKeysRepository(db *pgxpool.Pool) *ApiKeysRepository {
	return &ApiKeysRepository{db}
}

func (akr *ApiKeysRepository) GetAllByDeveloperId(ctx context.Context, developerId uuid.UUID) ([]*models.ApiKey, error) {
	op := "ApiKeysRepository.GetAllByDeveloperId"
	conn, err := akr.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: acquire connection: %w", op, err)
	}
	defer conn.Release()
	queryBuilder := sq.Select(
		"id",
		"name",
		"developer_id",
		"key_hash",
		"salt",
		"created_at",
		"revoked_at",
		"expired_at",
		"prefix_raw",
		"suffix_raw",
	).
		From("api_key").
		Where(sq.Eq{"developer_id": developerId}).
		PlaceholderFormat(sq.Dollar)
	sqlQuery, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: build query: %w", op, err)
	}
	rows, err := conn.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: execute query: %w", op, err)
	}
	defer rows.Close()
	var keys []*models.ApiKey
	for rows.Next() {
		var key models.ApiKey
		err := rows.Scan(
			&key.ID,
			&key.Name,
			&key.DeveloperID,
			&key.KeyHash,
			&key.Salt,
			&key.CreatedAt,
			&key.RevokedAt,
			&key.ExpireAt,
			&key.PrefixRaw,
			&key.SuffixRaw,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: scan row: %w", op, err)
		}
		keys = append(keys, &key)
	}
	return keys, nil
}

func (akr *ApiKeysRepository) GetById(ctx context.Context, id uuid.UUID) (*models.ApiKey, error) {
	op := "ApiKeysRepository.GetById"
	conn, err := akr.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: acquire connection: %w", op, err)
	}
	defer conn.Release()

	queryBuilder := sq.Select(
		"id",
		"name",
		"developer_id",
		"key_hash",
		"salt",
		"created_at",
		"revoked_at",
		"expired_at",
		"prefix_raw",
		"suffix_raw",
	).
		From("api_key").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)
	sqlQuery, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: build query: %w", op, err)
	}
	var key models.ApiKey
	row := conn.QueryRow(ctx, sqlQuery, args...)
	err = row.Scan(
		&key.ID,
		&key.Name,
		&key.DeveloperID,
		&key.KeyHash,
		&key.Salt,
		&key.CreatedAt,
		&key.RevokedAt,
		&key.ExpireAt,
		&key.PrefixRaw,
		&key.SuffixRaw,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: scan row: %w", op, err)
	}
	return &key, nil
}

func (akr *ApiKeysRepository) Create(ctx context.Context, key *models.ApiKey) (*models.ApiKey, error) {
	op := "ApiKeysRepository.Create"
	conn, err := akr.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: acquire connection: %w", op, err)
	}
	defer conn.Release()
	queryBuilder := sq.
		Insert("api_key").
		Columns("id", "name", "developer_id", "key_hash", "salt", "created_at", "revoked_at", "expired_at", "prefix_raw", "suffix_raw").
		Values(key.ID, key.Name, key.DeveloperID, key.KeyHash, key.Salt, key.CreatedAt, key.RevokedAt, key.ExpireAt, key.PrefixRaw, key.SuffixRaw).
		Suffix(`
			RETURNING
				id, name, developer_id, key_hash, salt, created_at, revoked_at, expired_at, prefix_raw, suffix_raw
		`).
		PlaceholderFormat(sq.Dollar)

	sqlQuery, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: build query: %w", op, err)
	}

	var created models.ApiKey
	row := conn.QueryRow(ctx, sqlQuery, args...)
	err = row.Scan(
		&created.ID,
		&created.Name,
		&created.DeveloperID,
		&created.KeyHash,
		&created.Salt,
		&created.CreatedAt,
		&created.RevokedAt,
		&created.ExpireAt,
		&created.PrefixRaw,
		&created.SuffixRaw,
	)
	if err != nil {
		return nil, err
	}
	return &created, nil
}
func (akr *ApiKeysRepository) Update(ctx context.Context, key *models.ApiKey) (*models.ApiKey, error) {
	op := "ApiKeysRepository.Update"
	conn, err := akr.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: acquire connection: %w", op, err)
	}
	defer conn.Release()
	queryBuilder := sq.
		Update("api_key").
		Set("name", key.Name).
		Set("developer_id", key.DeveloperID).
		Set("key_hash", key.KeyHash).
		Set("salt", key.Salt).
		Set("created_at", key.CreatedAt).
		Set("revoked_at", key.RevokedAt).
		Set("expired_at", key.ExpireAt).
		Set("prefix_raw", key.PrefixRaw).
		Set("suffix_raw", key.SuffixRaw).
		Where(sq.Eq{"id": key.ID}).
		Suffix(`
			RETURNING
				id, name, developer_id, key_hash, salt, created_at, revoked_at, expired_at, prefix_raw, suffix_raw
		`).
		PlaceholderFormat(sq.Dollar)

	sqlQuery, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: build query: %w", op, err)
	}
	var updated models.ApiKey
	row := conn.QueryRow(ctx, sqlQuery, args...)
	err = row.Scan(
		&updated.ID,
		&updated.Name,
		&updated.DeveloperID,
		&updated.KeyHash,
		&updated.Salt,
		&updated.CreatedAt,
		&updated.RevokedAt,
		&updated.ExpireAt,
		&updated.PrefixRaw,
		&updated.SuffixRaw,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: scan: %w", op, err)
	}
	return &updated, nil
}

func (akr *ApiKeysRepository) Remove(ctx context.Context, key *models.ApiKey) error {
	op := "ApiKeysRepository.Remove"
	conn, err := akr.db.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("%s: acquire connection: %w", op, err)
	}
	defer conn.Release()
	queryBuilder := sq.
		Delete("api_key").
		Where(sq.Eq{"id": key.ID}).
		PlaceholderFormat(sq.Dollar)
	sqlQuery, args, err := queryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("%s: build query: %w", op, err)
	}
	_, err = conn.Exec(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("%s: exec: %w", op, err)
	}
	return nil
}
