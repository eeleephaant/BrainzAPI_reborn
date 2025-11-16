package handler

import (
	"brainz/developersapi/internal/dtos"
	"brainz/developersapi/internal/entity"
	"brainz/developersapi/internal/services"
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"
)

type AuthHandler struct {
	us *services.UserService
	ss *services.SessionService
}

func (h *AuthHandler) Login(ctx context.Context, c *app.RequestContext) {
	op := "handlers.Login"
	loginData := dtos.LoginDto{}

	if err := c.Bind(&loginData); err != nil {
		c.JSON(400, map[string]string{"error": "invalid request"})
		return
	}

	if errs, err := loginData.Validate(); err != nil {
		c.JSON(400, map[string]any{
			"message": "validation failed",
			"errors":  errs,
		})
		return
	}

	account, err := h.us.Authenticate(ctx, loginData.Email, loginData.Password)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrWrongCredentials):
			c.JSON(401, map[string]string{"error": "wrong credentials"})
		case errors.Is(err, entity.ErrUserBanned):
			c.JSON(403, map[string]string{"error": "user banned"})
		case errors.Is(err, entity.ErrNeed2FA):
			c.JSON(401, map[string]string{"error": "2FA required"})
		default:
			c.JSON(500, map[string]string{"error": "internal server error"})
			zap.L().Error(op,
				zap.String("details", err.Error()),
			)
		}
		return
	}

	_, token, err := h.ss.CreateNew(ctx, account, string(c.UserAgent()), c.ClientIP())
	if err != nil {
		c.JSON(500, map[string]string{"error": "internal server error"})
	}

	c.JSON(200, map[string]string{"token": *token})
}

func (h *AuthHandler) Register(ctx context.Context, c *app.RequestContext) {
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

	_, err := h.us.RegistrateUser(ctx, &registerData)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrEmailAlreadyExists):
			c.JSON(409, map[string]string{"error": "email already exists"})
		default:
			c.JSON(500, map[string]string{"error": "internal server error"})
			zap.L().Error(op,
				zap.String("details", err.Error()),
			)
			return
		}
	}

	c.JSON(201, map[string]string{"message": "user created"})
}
