package handler

import (
	"brainz/developersapi/internal/dtos"
	"brainz/developersapi/internal/entity"
	"brainz/developersapi/internal/services"
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AuthUserService is the subset of UserService used by AuthHandler (for testing).
type AuthUserService interface {
	Authenticate(ctx context.Context, email, password string) (*entity.DeveloperAccount, error)
	RegistrateUser(ctx context.Context, registerDto *dtos.RegisterDto) (*entity.EmailConfirmationToken, error)
	GetById(ctx context.Context, devId uuid.UUID) (*entity.DeveloperAccount, error)
}

// AuthSessionService is the subset of SessionService used by AuthHandler (for testing).
type AuthSessionService interface {
	CreateNew(ctx context.Context, user *entity.DeveloperAccount, userAgent string, ipAddr string) (*entity.Session, *string, error)
	ValidateToken(ctx context.Context, token string, ipAddr string) (*entity.Session, error)
}

type AuthHandler struct {
	Us AuthUserService
	Ss AuthSessionService
	SC SessionHTTPConfig
}

func NewAuthHandler(us AuthUserService, ss AuthSessionService, sc SessionHTTPConfig) *AuthHandler {
	return &AuthHandler{Us: us, Ss: ss, SC: sc}
}

// Ensure concrete types satisfy interfaces.
var (
	_ AuthUserService    = (*services.UserService)(nil)
	_ AuthSessionService = (*services.SessionService)(nil)
)

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
		return
	}

	h.SC.SetSessionCookie(c, *token)
	c.JSON(200, map[string]string{"token": *token})
}

func (h *AuthHandler) Register(ctx context.Context, c *app.RequestContext) {
	op := "handlers.Register"
	registerData := dtos.RegisterDto{}

	if err := c.BindAndValidate(&registerData); err != nil {
		c.String(400, err.Error())
		return
	}

	ect, err := h.Us.RegistrateUser(ctx, &registerData)
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

	_ = ect // email confirmation disabled
	c.JSON(201, map[string]string{"message": "User created"})
}

func (h *AuthHandler) GetProfile(ctx context.Context, c *app.RequestContext) {
	sessionToken := h.SC.TokenFromRequest(c)
	session, err := h.Ss.ValidateToken(ctx, sessionToken, c.ClientIP())
	if err != nil || session == nil {
		c.JSON(400, map[string]string{"error": "invalid session token"})
		zap.L().Error("Invalid session token", zap.Error(err))
		return
	}
	user, err := h.Us.GetById(ctx, session.DeveloperID)
	if err != nil {
		c.JSON(500, map[string]string{"error": "internal server error"})
		zap.L().Error("GetProfile load user", zap.Error(err))
		return
	}
	c.JSON(200, dtos.NewProfileResponse(user))
}
