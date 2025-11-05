package handler

import (
	"brainz/developersapi/internal/storage/postgres"
	"context"
	"net/http"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func GetConfirmEmail(ctx context.Context, c *app.RequestContext) {
	op := "handler.GetConfirmEmail"

}

func ConfirmEmail(ctx context.Context, c *app.RequestContext) {
	op := "handler.ConfirmEmail"

	token := string(c.Query("token"))
	if token == "" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "missing token"})
		return
	}

	conn, err := postgres.DevsPool.Acquire(ctx)
	if err != nil {
		zap.L().Error(op,
			zap.String("message", "error postgres.DevsPool.Acquire"),
			zap.String("details", err.Error()),
		)
		c.JSON(500, map[string]string{"error": "internal error"})
		return
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		zap.L().Error(op,
			zap.String("message", "error conn.Begin()"),
			zap.String("details", err.Error()),
		)
		c.JSON(500, map[string]string{"error": "internal error"})
		return
	}
	defer tx.Rollback(ctx)

	var (
		developerID string
		expiresAt   time.Time
		usedAt      *time.Time
	)
	err = tx.QueryRow(ctx, `
		SELECT developer_id, expires_at, used_at
		FROM email_confirmation_token
		WHERE token = $1
	`, token).Scan(&developerID, &expiresAt, &usedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(400, map[string]string{"error": "invalid or expired token"})
			return
		}
		zap.L().Error(op,
			zap.String("message", "error UPDATE developer_account"),
			zap.String("details", err.Error()),
		)
		c.JSON(500, map[string]string{"error": "internal error"})
		return
	}

	if usedAt != nil {
		c.JSON(400, map[string]string{"error": "token already used"})
		return
	}
	if time.Now().After(expiresAt) {
		c.JSON(400, map[string]string{"error": "token expired"})
		return
	}

	_, err = tx.Exec(ctx, `
		UPDATE developer_account
		SET email_confirmed_at = now()
		WHERE id = $1 AND email_confirmed_at IS NULL
	`, developerID)
	if err != nil {
		zap.L().Error(op,
			zap.String("message", "error UPDATE developer_account"),
			zap.String("details", err.Error()),
		)
		c.JSON(500, map[string]string{"error": "internal error"})
		return
	}

	_, err = tx.Exec(ctx, `
		UPDATE email_confirmation_token
		SET used_at = now()
		WHERE token = $1
	`, token)
	if err != nil {
		zap.L().Error(op,
			zap.String("message", "error UPDATE email_confirmation_token"),
			zap.String("details", err.Error()),
		)
		c.JSON(500, map[string]string{"error": "internal error"})
		return
	}

	if err = tx.Commit(ctx); err != nil {
		zap.L().Error(op,
			zap.String("message", "error tx Commit"),
			zap.String("details", err.Error()),
		)
		c.JSON(500, map[string]string{"error": "internal error"})
		return
	}

	c.JSON(200, map[string]string{
		"message": "Email confirmed successfully",
	})
}
