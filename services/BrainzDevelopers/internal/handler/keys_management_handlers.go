package handler

import (
	"brainz/common/dtos"
	"brainz/common/permissions"
	"brainz/developersapi/internal/entity"
	"brainz/developersapi/internal/services"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type KeysUserService interface {
	GetById(ctx context.Context, devId uuid.UUID) (*entity.DeveloperAccount, error)
}

// keysApiKeysService is the subset of ApiKeysService used by KeysManagementHandler (for testing).
type keysApiKeysService interface {
	GetApiKeys(ctx context.Context, developerID uuid.UUID) ([]dtos.ApiKeyShareModel, error)
	GetApiKeysUsage(ctx context.Context, developerID uuid.UUID) ([]dtos.ApiKeyUsageStats, error)
	CreateApiKey(ctx context.Context, issuer uuid.UUID, developerID uuid.UUID, name string, perms []permissions.Permission, ipWhitelist []string) (*dtos.ApiKeyCreateResponse, error)
	RemoveApiKey(ctx context.Context, apiKey string, devID string) error
}

// keysSessionValidator is the subset of SessionService used by KeysManagementHandler (for testing).
type keysSessionValidator interface {
	ValidateToken(ctx context.Context, token string, ipAddr string) (*entity.Session, error)
}

type KeysManagementHandler struct {
	ks keysApiKeysService
	ss keysSessionValidator
	us KeysUserService
	sc SessionHTTPConfig
}

func NewKeysManagementHandler(ks *services.ApiKeysService, ss *services.SessionService, us KeysUserService, sc SessionHTTPConfig) *KeysManagementHandler {
	return &KeysManagementHandler{ks: ks, ss: ss, us: us, sc: sc}
}

var (
	_ KeysUserService       = (*services.UserService)(nil)
	_ keysApiKeysService    = (*services.ApiKeysService)(nil)
	_ keysSessionValidator  = (*services.SessionService)(nil)
)

// Роли разработчика (role_id в БД) задают, какие права получит созданный API-ключ.
// Права не передаются в теле запроса — они вычисляются из аккаунта по session.
//
// Примеры:
//
//   role_id 0 — только чтение: ключ с правами [read] (без привязки к институту).
//   role_id 1 — чтение и запись: ключ с правами [read, write].
//
// Любой другой role_id → 403 "unsupported role".
const (
	DeveloperRoleReadOnly  uint = 0 // read
	DeveloperRoleReadWrite uint = 1 // read + write
)

func (k *KeysManagementHandler) GetApiKeys(ctx context.Context, c *app.RequestContext) {
	sessionToken := k.sc.TokenFromRequest(c)
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
	if keys == nil {
		keys = []dtos.ApiKeyShareModel{}
	}
	c.JSON(200, map[string][]dtos.ApiKeyShareModel{"keys": keys})
}

func (k *KeysManagementHandler) GetApiKeysUsage(ctx context.Context, c *app.RequestContext) {
	// Stats are scoped strictly to the session user; never accept developer_id from the client.
	if strings.TrimSpace(c.Query("developer_id")) != "" {
		c.JSON(400, map[string]string{"error": "developer_id query parameter is not allowed; session cookie or X-Session-Token only"})
		return
	}

	sessionToken := strings.TrimSpace(k.sc.TokenFromRequest(c))
	if sessionToken == "" {
		c.JSON(401, map[string]string{"error": "missing session token"})
		return
	}

	session, err := k.ss.ValidateToken(ctx, sessionToken, c.ClientIP())
	if err != nil || session == nil {
		c.JSON(401, map[string]string{"error": "invalid session token"})
		// Log invalid attempts at debug level; avoid leaking details to client.
		if err != nil {
			zap.L().Debug("session validation failed", zap.Error(err))
		}
		return
	}

	stats, err := k.ks.GetApiKeysUsage(ctx, session.DeveloperID)
	if err != nil {
		c.JSON(500, map[string]string{"error": "internal server error"})
		zap.L().Error("GetApiKeysUsage", zap.Error(err))
		return
	}
	if stats == nil {
		stats = []dtos.ApiKeyUsageStats{}
	}
	c.JSON(200, dtos.ApiKeysUsageResponse{Keys: stats})
}

func (k *KeysManagementHandler) CreateApiKey(ctx context.Context, c *app.RequestContext) {
	sessionToken := k.sc.TokenFromRequest(c)
	session, err := k.ss.ValidateToken(ctx, sessionToken, c.ClientIP())
	if err != nil || session == nil {
		c.JSON(400, map[string]string{"error": "invalid session token"})
		zap.L().Error("Invalid session token", zap.Error(err))
		return
	}
	apiCreateRequest := dtos.ApiKeyCreateSessionRequest{}
	err = c.BindAndValidate(&apiCreateRequest)
	if err != nil {
		c.JSON(400, map[string]string{"error": "invalid request"})
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
		session.DeveloperID,
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
	case DeveloperRoleReadOnly:
		return []permissions.Permission{
			{Action: permissions.ActionRead},
		}, nil
	case DeveloperRoleReadWrite:
		return []permissions.Permission{
			{Action: permissions.ActionRead},
			{Action: permissions.ActionWrite},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported role_id: %d", roleID)
	}
}

func (k *KeysManagementHandler) DeleteApiKey(ctx context.Context, c *app.RequestContext) {
	sessionToken := k.sc.TokenFromRequest(c)
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
