package repositories

import (
	"brainz/auth/internal/models"
	"brainz/common/dtos"
	"context"
	"fmt"
	"sort"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ApiKeyUsageRepository struct {
	db *pgxpool.Pool
}

func NewApiKeyUsageRepository(db *pgxpool.Pool) *ApiKeyUsageRepository {
	return &ApiKeyUsageRepository{db}
}

func (akur *ApiKeyUsageRepository) Create(ctx context.Context, key *models.ApiKeyUsage) (*models.ApiKeyUsage, error) {
	op := "ApiKeyUsageRepository.Create"
	conn, err := akur.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: acquire connection: %w", op, err)
	}
	defer conn.Release()
	queryBuilder := sq.Insert("api_key_usage").
		Columns("id", "api_key_id", "endpoint", "method", "usage_at", "response_code").
		Values(key.ID, key.ApiKeyID, key.Endpoint, key.Method, key.UsageAt, key.ResponseCode).
		Suffix(`
		RETURNING
			id, api_key_id, endpoint, method, usage_at, response_code
	`).PlaceholderFormat(sq.Dollar)
	sqlQuery, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: build query: %w", op, err)
	}

	var created models.ApiKeyUsage
	row := conn.QueryRow(ctx, sqlQuery, args...)
	err = row.Scan(
		&created.ID,
		&created.ApiKeyID,
		&created.Endpoint,
		&created.Method,
		&created.UsageAt,
		&created.ResponseCode,
	)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

// UsageStatsForDeveloper returns per-key totals and daily breakdown (UTC days) for the last window starting at historySince.
func (akur *ApiKeyUsageRepository) UsageStatsForDeveloper(ctx context.Context, developerID uuid.UUID, historySince time.Time) ([]dtos.ApiKeyUsageStats, error) {
	op := "ApiKeyUsageRepository.UsageStatsForDeveloper"
	conn, err := akur.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: acquire: %w", op, err)
	}
	defer conn.Release()

	totalsSQL := `
SELECT k.id, k.name, k.created_at,
       COALESCE(COUNT(u.id), 0)::bigint AS total_requests,
       MAX(u.usage_at) AS last_usage
FROM api_key k
LEFT JOIN api_key_usage u ON u.api_key_id = k.id
WHERE k.developer_id = $1
GROUP BY k.id, k.name, k.created_at
ORDER BY k.created_at DESC`

	rows, err := conn.Query(ctx, totalsSQL, developerID)
	if err != nil {
		return nil, fmt.Errorf("%s: totals query: %w", op, err)
	}
	defer rows.Close()

	byID := make(map[uuid.UUID]*dtos.ApiKeyUsageStats)
	order := make([]uuid.UUID, 0)

	for rows.Next() {
		var (
			id        uuid.UUID
			name      string
			createdAt time.Time
			total     int64
			lastUsage *time.Time
		)
		if err := rows.Scan(&id, &name, &createdAt, &total, &lastUsage); err != nil {
			return nil, fmt.Errorf("%s: totals scan: %w", op, err)
		}
		if total < 0 {
			total = 0
		}
		st := dtos.ApiKeyUsageStats{
			ApiKeyID:   id,
			Name:       name,
			CreatedAt:  createdAt,
			LastUsedAt: lastUsage,
			TotalUsage: uint64(total),
			UsageByDay: []dtos.ApiKeyUsageDaily{},
		}
		byID[id] = &st
		order = append(order, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: totals rows: %w", op, err)
	}

	dailySQL := `
SELECT u.api_key_id,
       (date_trunc('day', u.usage_at AT TIME ZONE 'UTC'))::date AS day,
       COUNT(*)::bigint AS cnt
FROM api_key_usage u
INNER JOIN api_key k ON k.id = u.api_key_id AND k.developer_id = $1
WHERE u.usage_at >= $2
GROUP BY u.api_key_id, (date_trunc('day', u.usage_at AT TIME ZONE 'UTC'))::date`

	dailyRows, err := conn.Query(ctx, dailySQL, developerID, historySince)
	if err != nil {
		return nil, fmt.Errorf("%s: daily query: %w", op, err)
	}
	defer dailyRows.Close()

	for dailyRows.Next() {
		var (
			keyID uuid.UUID
			day   time.Time
			cnt   int64
		)
		if err := dailyRows.Scan(&keyID, &day, &cnt); err != nil {
			return nil, fmt.Errorf("%s: daily scan: %w", op, err)
		}
		st, ok := byID[keyID]
		if !ok {
			continue
		}
		c := cnt
		if c < 0 {
			c = 0
		}
		st.UsageByDay = append(st.UsageByDay, dtos.ApiKeyUsageDaily{
			Date:  day.UTC(),
			Count: uint32(c),
		})
	}
	if err := dailyRows.Err(); err != nil {
		return nil, fmt.Errorf("%s: daily rows: %w", op, err)
	}

	out := make([]dtos.ApiKeyUsageStats, 0, len(order))
	for _, id := range order {
		st := byID[id]
		sort.Slice(st.UsageByDay, func(i, j int) bool {
			return st.UsageByDay[i].Date.After(st.UsageByDay[j].Date)
		})
		out = append(out, *st)
	}
	return out, nil
}
