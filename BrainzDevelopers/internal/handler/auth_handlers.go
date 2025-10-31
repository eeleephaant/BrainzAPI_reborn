package handler

import (
	"brainz/developersapi/internal/dtos"
	"brainz/developersapi/internal/models"
	"brainz/developersapi/internal/security"
	"brainz/developersapi/internal/services"
	"brainz/developersapi/internal/storage/postgres"
	"bytes"
	"context"
	"encoding/hex"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func Login(ctx context.Context, c *app.RequestContext) {
	loginData := dtos.LoginDto{}

	if err := c.Bind(&loginData); err != nil {
		c.String(400, err.Error())
		return
	}

	if err := loginData.Validate(); err != nil {
		c.String(400, err.Error())
		return
	}

	conn, err := postgres.DevsPool.Acquire(ctx)
	if err != nil {
		c.String(500, "internal error")
		return
	}
	defer conn.Release()
	var (
		accountSalt         []byte
		accountPasswordHash []byte
		twoFactorSecret     *string
		userId              uuid.UUID
	)

	err = conn.QueryRow(ctx,
		"SELECT id, salt, password_hash, two_factor_secret FROM developer_account WHERE email=$1", loginData.Email).Scan(&userId, &accountSalt, &accountPasswordHash, &twoFactorSecret)
	if err != nil {
		c.String(500, "internal error")
		return
	}

	passwordHash := security.GetHashArgon2(loginData.Password, accountSalt)

	if !bytes.Equal(passwordHash, accountPasswordHash) {
		c.String(401, "invalid email or password")
		return
	}

	if twoFactorSecret != nil {
		c.String(200, "")
		return
	}

	requestData := models.SessionRequestData{IpAddress: c.ClientIP(), UserAgent: string(c.UserAgent()), DeveloperId: userId}

	key, err := services.GetNewSession(ctx, conn, &requestData)
	if err != nil {
		c.String(500, "internal error")
		return
	}
	sessionKeyStr := hex.EncodeToString(key)

	c.JSON(200, map[string]string{"session_key": sessionKeyStr})
}

func Register(ctx context.Context, c *app.RequestContext) {
	registerData := dtos.RegisterDto{}

	if err := c.Bind(&registerData); err != nil {
		c.String(400, err.Error())
		return
	}

	if err := registerData.Validate(); err != nil {
		c.String(400, err.Error())
		return
	}

	salt := security.GetRandomSalt()
	password_hash := security.GetHashArgon2(registerData.Password, salt)

	conn, err := postgres.DevsPool.Acquire(ctx)
	if err != nil {
		c.String(500, "internal error")
		return
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		c.String(500, "internal error")
		return
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx,
		`INSERT INTO "users" (email, password_hash, created_at, role_id, salt) 
		VALUES ($1, $2, now(), $3, $4)`, registerData.Email, password_hash, 0, salt)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" { // SQLSTATE code: unique_violation (email)
			c.String(409, "email already registered")
			return
		}
		c.String(500, "internal error")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.String(500, "internal error")
		return
	}
	c.String(201, "user created")
}
