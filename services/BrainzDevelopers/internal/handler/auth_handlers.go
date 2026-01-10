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
	Us *services.UserService
	Ss *services.SessionService
}

func NewAuthHandler(us *services.UserService, ss *services.SessionService) *AuthHandler {
	return &AuthHandler{Us: us, Ss: ss}
}

func (h *AuthHandler) Login(ctx context.Context, c *app.RequestContext) {
	op := "handlers.Login"
	loginData := dtos.LoginDto{}

	if err := c.BindAndValidate(&loginData); err != nil {
		c.JSON(400, map[string]string{"error": "invalid request"})
		return
	}

	account, err := h.Us.Authenticate(ctx, loginData.Email, loginData.Password)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrWrongCredentials):
			c.JSON(401, map[string]string{"error": "wrong credentials"})
		case errors.Is(err, entity.ErrUserBanned):
			c.JSON(403, map[string]string{"error": "user banned"})
		case errors.Is(err, entity.ErrNeed2FA):
			c.JSON(401, map[string]string{"error": "2FA required"})
		case errors.Is(err, entity.ErrEmailNotConfirmed):
			c.JSON(401, map[string]string{"error": "email not confirmed"})
		default:
			c.JSON(500, map[string]string{"error": "internal server error"})
			zap.L().Error(op,
				zap.String("details", err.Error()),
			)
		}
		return
	}

	_, token, err := h.Ss.CreateNew(ctx, account, string(c.UserAgent()), c.ClientIP())
	if err != nil {
		c.JSON(500, map[string]string{"error": "internal server error"})
	}

	c.JSON(200, map[string]string{"token": *token})
}

func (h *AuthHandler) Register(ctx context.Context, c *app.RequestContext) {
	op := "handlers.Register"
	registerData := dtos.RegisterDto{}

	if err := c.BindAndValidate(&registerData); err != nil {
		c.String(400, err.Error())
		return
	}

	_, err := h.Us.RegistrateUser(ctx, &registerData)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrEmailAlreadyExists):
			c.JSON(409, map[string]string{"error": "email already exists"})
			return
		default:
			c.JSON(500, map[string]string{"error": "internal server error"})
			zap.L().Error(op,
				zap.String("details", err.Error()),
			)
			return
		}
	}

	c.JSON(201, map[string]string{"message": "User created, check your email to confirm it"})
}
