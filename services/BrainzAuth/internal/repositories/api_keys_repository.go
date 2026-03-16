package repositories

import (
	"brainz/auth/internal/models"
	"brainz/common/permissions"
	"context"
	"fmt"
	"net/netip"

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

func (akr *ApiKeysRepository) GetWhitelistedIPs(ctx context.Context, keyID uuid.UUID) ([]netip.Addr, error) {
	op := "ApiKeysRepository.GetWhitelistedIPs"
	conn, err := akr.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: acquire connection: %w", op, err)
	}
	defer conn.Release()
	queryBuilder := sq.Select("ip_address").
		From("api_key_whitelist").
		Where(sq.Eq{"api_key_id": keyID}).
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
	var ips []netip.Addr
	for rows.Next() {
		var ip netip.Addr
		if err := rows.Scan(&ip); err != nil {
			return nil, fmt.Errorf("%s: scan: %w", op, err)
		}
		ips = append(ips, ip)
	}
	return ips, nil
}

func (akr *ApiKeysRepository) GetPermissions(
	ctx context.Context,
	keyID uuid.UUID,
) ([]permissions.Permission, error) {
	const op = "ApiKeysRepository.GetPermissions"

	conn, err := akr.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: acquire connection: %w", op, err)
	}
	defer conn.Release()

	queryBuilder := sq.
		Select(
			"p.title",
			"g.institution_id",
		).
		From("api_key_permission_grant g").
		Join("api_key_permission p ON p.id = g.permission_id").
		Where(sq.Eq{"g.api_key_id": keyID}).
		OrderBy("p.title", "g.institution_id").
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

	var result []permissions.Permission

	for rows.Next() {
		var (
			title         string
			institutionID *int64
		)

		if err := rows.Scan(&title, &institutionID); err != nil {
			return nil, fmt.Errorf("%s: scan row: %w", op, err)
		}

		result = append(result, permissions.Permission{
			Action:        permissions.Action(title),
			InstitutionID: institutionID,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: iterate rows: %w", op, err)
	}

	return result, nil
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
		"expire_at",
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
		"expire_at",
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
		Columns("id", "name", "developer_id", "key_hash", "salt", "created_at", "revoked_at", "expire_at", "prefix_raw", "suffix_raw").
		Values(key.ID, key.Name, key.DeveloperID, key.KeyHash, key.Salt, key.CreatedAt, key.RevokedAt, key.ExpireAt, key.PrefixRaw, key.SuffixRaw).
		Suffix(`
			RETURNING
				id, name, developer_id, key_hash, salt, created_at, revoked_at, expire_at, prefix_raw, suffix_raw
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

func (akr *ApiKeysRepository) CreateWithPermissions(
	ctx context.Context,
	key *models.ApiKey,
	grants []permissions.Permission,
) (*models.ApiKey, error) {
	op := "ApiKeysRepository.CreateWithPermissions"

	tx, err := akr.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: begin tx: %w", op, err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	insertKey := sq.
		Insert("api_key").
		Columns("id", "name", "developer_id", "key_hash", "salt", "created_at", "revoked_at", "expire_at", "prefix_raw", "suffix_raw").
		Values(key.ID, key.Name, key.DeveloperID, key.KeyHash, key.Salt, key.CreatedAt, key.RevokedAt, key.ExpireAt, key.PrefixRaw, key.SuffixRaw).
		Suffix(`
			RETURNING
				id, name, developer_id, key_hash, salt, created_at, revoked_at, expire_at, prefix_raw, suffix_raw
		`).
		PlaceholderFormat(sq.Dollar)

	sqlQuery, args, err := insertKey.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: build insert key query: %w", op, err)
	}

	var created models.ApiKey
	row := tx.QueryRow(ctx, sqlQuery, args...)
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
		return nil, fmt.Errorf("%s: scan inserted key: %w", op, err)
	}

	for _, grant := range grants {
		tag, execErr := tx.Exec(
			ctx,
			`
			INSERT INTO api_key_permission_grant (api_key_id, permission_id, institution_id)
			SELECT $1, p.id, $2
			FROM api_key_permission p
			WHERE p.title = $3
			`,
			created.ID,
			grant.InstitutionID,
			string(grant.Action),
		)
		if execErr != nil {
			err = fmt.Errorf("%s: insert permission grant: %w", op, execErr)
			return nil, err
		}
		if tag.RowsAffected() == 0 {
			err = fmt.Errorf("%s: unknown permission action %q", op, grant.Action)
			return nil, err
		}
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return nil, fmt.Errorf("%s: commit tx: %w", op, commitErr)
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
		Set("expire_at", key.ExpireAt).
		Set("prefix_raw", key.PrefixRaw).
		Set("suffix_raw", key.SuffixRaw).
		Where(sq.Eq{"id": key.ID}).
		Suffix(`
			RETURNING
				id, name, developer_id, key_hash, salt, created_at, revoked_at, expire_at, prefix_raw, suffix_raw
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
func (akr *ApiKeysRepository) GetByName(ctx context.Context, keyName string) (*models.ApiKey, error) {
	op := "ApiKeysRepository.GetIdByName"
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
		"expire_at",
		"prefix_raw",
		"suffix_raw",
	).
		From("api_key").
		Where(sq.Eq{"name": keyName}).
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
