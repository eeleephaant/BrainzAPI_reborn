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
	"errors"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type AuthHandler struct {
	svc *services.
}

func Login(ctx context.Context, c *app.RequestContext) {
	op := "handlers.Login"
	loginData := dtos.LoginDto{}

	if err := c.Bind(&loginData); err != nil {
		c.JSON(400, map[string]string{"error": "invalid request"})
		return
	}

	if errs, err := loginData.Validate(); err != nil {
		c.JSON(400, map[string]any{
			"message": "Validation failed",
			"errors":  errs,
		})
		return
	}

	conn, err := postgres.DevsPool.Acquire(ctx)
	if err != nil {
		c.JSON(500, map[string]string{"error": "internal error"})
		zap.L().Error(op,
			zap.String("message", "error acquiring database connection"),
			zap.String("details", err.Error()),
		)
		return
	}
	defer conn.Release()
	var (
		accountSalt         []byte
		accountPasswordHash []byte
		twoFactorSecret     *string
		userId              uuid.UUID
		bannedAt            *time.Time
		emailConfirmedAt    *time.Time
	)
	if err := conn.QueryRow(ctx,
		"SELECT id, salt, password_hash, two_factor_secret, banned_at, email_confirmed_at FROM developer_account WHERE email=$1",
		loginData.Email,
	).Scan(&userId, &accountSalt, &accountPasswordHash, &twoFactorSecret, &bannedAt, &emailConfirmedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(401, map[string]string{"message": "invalid password or email"})
			return
		}
		c.JSON(500, map[string]string{"error": "internal error"})
		zap.L().Error(op,
			zap.String("message", "error SELECT querying database"),
			zap.String("details", err.Error()),
		)
		return
	}

	passwordHash := security.GetHashArgon2(loginData.Password, accountSalt)

	if !bytes.Equal(passwordHash, accountPasswordHash) {
		c.JSON(401, map[string]string{"message": "invalid email or password"})
		return
	}

	if emailConfirmedAt == nil {
		c.JSON(401, map[string]string{"message": "need to confirm email"})
		return
	}

	if bannedAt != nil {
		c.JSON(401, map[string]string{"message": "whooops... your account is banned"})
		return
	}

	if twoFactorSecret != nil {
		c.JSON(200, map[string]string{"message": "Need for 2fa"})
		return
	}

	requestData := models.SessionRequestData{IpAddress: c.ClientIP(), UserAgent: string(c.UserAgent()), DeveloperId: userId}

	key, err := services.GetNewSession(ctx, conn, &requestData)
	if err != nil {
		c.JSON(500, map[string]string{"error": "internal error"})
		zap.L().Error(op,
			zap.String("message", "error while creating session"),
			zap.String("details", err.Error()),
		)
		return
	}
	sessionKeyStr := hex.EncodeToString(key)

	c.JSON(200, map[string]string{"session_key": sessionKeyStr})
}

func Register(ctx context.Context, c *app.RequestContext) {
	op := "handlers.Register"
	registerData := dtos.RegisterDto{}

	if err := c.Bind(&registerData); err != nil {
		c.String(400, err.Error())
		return
	}

	if errs, err := registerData.Validate(); err != nil {
		c.JSON(400, map[string]any{
			"message": "validation failed",
			"errors":  errs,
		})
		return
	}

	zap.L().Error(op,
		zap.String("message", "error acquiring database connection"),
		zap.String("details", err.Error()),
	)

	c.JSON(201, map[string]string{"message": "user created"})
}
