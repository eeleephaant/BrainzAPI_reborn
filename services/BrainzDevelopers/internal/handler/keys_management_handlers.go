package handler

import (
	"brainz/common/dtos"
	"brainz/common/permissions"
	"brainz/developersapi/internal/entity"
	"brainz/developersapi/internal/services"
	"context"
	"errors"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type KeysUserService interface {
	GetById(ctx context.Context, devId uuid.UUID) (*entity.DeveloperAccount, error)
}

type KeysManagementHandler struct {
	ks *services.ApiKeysService
	ss *services.SessionService
	us KeysUserService
}

func NewKeysManagementHandler(ks *services.ApiKeysService, ss *services.SessionService, us KeysUserService) *KeysManagementHandler {
	return &KeysManagementHandler{ks: ks, ss: ss, us: us}
}

var _ KeysUserService = (*services.UserService)(nil)

func (k *KeysManagementHandler) GetApiKeys(ctx context.Context, c *app.RequestContext) {
	sessionToken := c.Request.Header.Get("X-Session-Token")
	session, err := k.ss.ValidateToken(ctx, sessionToken, c.ClientIP())
	if err != nil || session == nil {
		c.JSON(400, map[string]string{"error": "invalid session token"})
		zap.L().Error("Invalid session token", zap.Error(err))
		return
	}
	keys, err := k.ks.GetApiKeys(ctx, session.DeveloperID)
	if err != nil {
		c.JSON(500, map[string]string{"error": "internal server error"})
		zap.L().Error("Internal server error", zap.Error(err))
		return
	}
	c.JSON(200, keys)
}

func (k *KeysManagementHandler) CreateApiKey(ctx context.Context, c *app.RequestContext) {
	sessionToken := c.Request.Header.Get("X-Session-Token")
	session, err := k.ss.ValidateToken(ctx, sessionToken, c.ClientIP())
	if err != nil || session == nil {
		c.JSON(400, map[string]string{"error": "invalid session token"})
		zap.L().Error("Invalid session token", zap.Error(err))
		return
	}
	apiCreateRequest := dtos.ApiKeyCreateDto{}
	err = c.BindAndValidate(&apiCreateRequest)
	if err != nil {
		c.JSON(400, map[string]string{"error": "invalid request"})
		return
	}

	if apiCreateRequest.DevUUID != session.DeveloperID {
		c.JSON(403, map[string]string{"error": "you cannot create keys for other developers"})
		return
	}

	user, err := k.us.GetById(ctx, session.DeveloperID)
	if err != nil {
		c.JSON(500, map[string]string{"error": "internal server error"})
		zap.L().Error("Failed to load developer account", zap.Error(err))
		return
	}

	grants, err := permissionsForRole(user.RoleId)
	if err != nil {
		c.JSON(403, map[string]string{"error": "unsupported role"})
		zap.L().Error("Unsupported developer role", zap.Uint("role_id", user.RoleId), zap.Error(err))
		return
	}

	newKey, err := k.ks.CreateApiKey(
		ctx,
		session.DeveloperID,
		apiCreateRequest.DevUUID,
		apiCreateRequest.ApiKeyName,
		grants,
		apiCreateRequest.IPWhitelist,
	)
	if err != nil {
		c.JSON(500, map[string]string{"error": "internal server error"})
		return
	}
	c.JSON(201, newKey)
}

func permissionsForRole(roleID uint) ([]permissions.Permission, error) {
	switch roleID {
	case 0:
		return []permissions.Permission{
			{Action: permissions.ActionRead},
		}, nil
	case 1:
		return []permissions.Permission{
			{Action: permissions.ActionRead},
			{Action: permissions.ActionWrite},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported role_id: %d", roleID)
	}
}

func (k *KeysManagementHandler) DeleteApiKey(ctx context.Context, c *app.RequestContext) {
	sessionToken := c.Request.Header.Get("X-Session-Token")
	apiKeyStr := c.Query("api_key")
	session, err := k.ss.ValidateToken(ctx, sessionToken, c.ClientIP())
	if err != nil || session == nil {
		c.JSON(400, map[string]string{"error": "invalid session token"})
		zap.L().Error("Invalid session token", zap.Error(err))
		return
	}
	err = k.ks.RemoveApiKey(ctx, apiKeyStr, session.DeveloperID.String())
	if err != nil {
		if errors.Is(err, services.ErrAccessDenied) {
			c.JSON(403, map[string]string{"error": "access denied"})
			return
		}
		c.JSON(500, map[string]string{"error": "internal server error"})
		zap.L().Error("Internal server error", zap.Error(err))
		return
	}
	c.JSON(200, map[string]string{"status": "success"})
}
