package repository

import (
	"brainz/developersapi/internal/entity"
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db}
}

func (ur *UserRepository) Update(ctx context.Context, user *entity.DeveloperAccount) (*entity.DeveloperAccount, error) {
	op := "UserRepository.Update"
	conn, err := ur.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: acquire connection: %w", op, err)
	}
	defer conn.Release()
	queryBuilder := sq.
		Update("developer_accounts").
		Set("email", user.Email).
		Set("email_confirmed_at", user.EmailConfirmedAt).
		Set("password_hash", user.PasswordHash).
		Set("salt", user.Salt).
		Set("two_factor_secret", user.TwoFactorSecret).
		Set("created_at", user.CreatedAt).
		Set("banned_at", user.BannedAt).
		Set("role_id", user.RoleId).
		Where(sq.Eq{"id": user.ID}).
		Suffix(`
			RETURNING
				id, email, email_confirmed_at, password_hash,
				salt, two_factor_secret, created_at, banned_at, role_id
		`).
		PlaceholderFormat(sq.Dollar)

	sqlQuery, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: build query: %w", op, err)
	}

	var updated entity.DeveloperAccount
	row := conn.QueryRow(ctx, sqlQuery, args...)
	err = row.Scan(
		&updated.ID,
		&updated.Email,
		&updated.EmailConfirmedAt,
		&updated.PasswordHash,
		&updated.Salt,
		&updated.TwoFactorSecret,
		&updated.CreatedAt,
		&updated.BannedAt,
		&updated.RoleId,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: scan: %w", op, err)
	}
	return &updated, nil
}

func (ur *UserRepository) Create(ctx context.Context, developer_account *entity.DeveloperAccount) (*entity.DeveloperAccount, error) {
	op := "UserRepository.Create"
	conn, err := ur.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: acquire connection: %w", op, err)
	}
	defer conn.Release()

	queryBuilder := sq.
		Insert("developer_accounts").
		Columns("email", "email_confirmed_at", "password_hash", "salt", "role_id").
		Values(developer_account.Email, developer_account.EmailConfirmedAt, developer_account.PasswordHash, developer_account.Salt, developer_account.RoleId).
		Suffix("RETURNING id, email, email_confirmed_at, password_hash, salt, two_factor_secret, created_at, banned_at, role_id").
		PlaceholderFormat(sq.Dollar)

	sqlQuery, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: build query: %w", op, err)
	}

	var newAccount entity.DeveloperAccount
	row := conn.QueryRow(ctx, sqlQuery, args...)

	if err := row.Scan(
		&newAccount.ID,
		&newAccount.Email,
		&newAccount.EmailConfirmedAt,
		&newAccount.PasswordHash,
		&newAccount.Salt,
		&newAccount.TwoFactorSecret,
		&newAccount.CreatedAt,
		&newAccount.BannedAt,
		&newAccount.RoleId,
	); err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok && pgErr.Code == "23505" {
			return nil, entity.ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("%s: execute query: %w", op, err)
	}

	return &newAccount, nil
}

func (ur *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.DeveloperAccount, error) {
	op := "UserRepository.GetByEmail"
	conn, err := ur.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: acquire connection: %w", op, err)
	}
	defer conn.Release()

	queryBuilder := sq.
		Select("id", "email", "password_hash", "email_confirmed_at", "salt", "two_factor_secret", "created_at", "banned_at", "role_id").
		From("developer_accounts").
		Where(sq.Eq{"email": email}).
		PlaceholderFormat(sq.Dollar)
	sqlQuery, args, err := queryBuilder.ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: build query: %w", op, err)
	}

	var account entity.DeveloperAccount
	row := conn.QueryRow(ctx, sqlQuery, args...)

	if err := row.Scan(
		&account.ID,
		&account.Email,
		&account.PasswordHash,
		&account.EmailConfirmedAt,
		&account.Salt,
		&account.TwoFactorSecret,
		&account.CreatedAt,
		&account.BannedAt,
		&account.RoleId,
	); err != nil {
		return nil, fmt.Errorf("%s: execute query: %w", op, err)
	}

	return &account, nil
}

func (ur *UserRepository) GetById(ctx context.Context, devId uuid.UUID) (*entity.DeveloperAccount, error) {
	op := "UserRepository.GetById"
	conn, err := ur.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: acquire connection: %w", op, err)
	}
	defer conn.Release()

	queryBuilder := sq.
		Select("id", "email", "password_hash", "email_confirmed_at", "salt", "two_factor_secret", "created_at", "banned_at", "role_id").
		From("developer_accounts").
		Where(sq.Eq{"id": devId}).
		PlaceholderFormat(sq.Dollar)
	sqlQuery, args, err := queryBuilder.ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: build query: %w", op, err)
	}

	var account entity.DeveloperAccount
	row := conn.QueryRow(ctx, sqlQuery, args...)

	if err := row.Scan(
		&account.ID,
		&account.Email,
		&account.PasswordHash,
		&account.EmailConfirmedAt,
		&account.Salt,
		&account.TwoFactorSecret,
		&account.CreatedAt,
		&account.BannedAt,
		&account.RoleId,
	); err != nil {
		return nil, fmt.Errorf("%s: execute query: %w", op, err)
	}

	return &account, nil
}
