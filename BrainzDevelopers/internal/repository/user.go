package repository

import (
	"brainz/developersapi/internal/entity"
	"brainz/developersapi/internal/models"
	"brainz/developersapi/internal/security"
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db}
}

func (ur *UserRepository) Create(ctx context.Context, createDevRequest *models.AccountCreateRequestData) (*entity.DeveloperAccount, error) {
	op := "UserRepository.Create"
	conn, err := ur.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: acquire connection: %w", op, err)
	}
	defer conn.Release()

	salt := security.GetRandomSalt()
	passwordHash := security.GetHashArgon2(createDevRequest.Password, salt)

	queryBuilder := sq.
		Insert("developer_accounts").
		Columns("email", "password_hash", "salt", "role_id").
		Values(createDevRequest.Email, passwordHash, salt, 0).
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
		From("developer_account").
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
