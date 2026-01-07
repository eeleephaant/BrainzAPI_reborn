package repositories

import (
	"brainz/auth/internal/models"
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ApiKeyUsageRepository struct {
	db *pgxpool.Pool
}

func NewApiKeyUsageRepository(db *pgxpool.Pool) *ApiKeyUsageRepository{
	return &ApiKeyUsageRepository{db}
}

func (akur *ApiKeyUsageRepository) Create(ctx context.Context, key *models.ApiKeyUsage) (*models.ApiKeyUsage, error){
	op := "ApiKeyUsageRepository.Create"
	conn, err = akur.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: acquire connection: %w", op, err)
	}
	defer conn.Release()
	queryBuilder := sq.Insert("api_key_usage").
	Columns("id", "api_key_id", "endpoint", "usage_at", "response_code").
	Values(key.ID, key.ApiKeyID, key.Endpoint, key.CreatedAt, key.Status).
	Suffix(`
			RETURNING
				id, api_key_id, developer_id, endpoint, usage_at, created_at, response_code
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
		&created.ApiKeyID,
		&created.Endpoint,
		&created.CreatedAt,
		&created.Status
	)
	if err != nil {
		return nil, err
	}
	return &created, nil
}