package handler

import (
	"brainz-api/internal/dtos"
	"brainz-api/internal/services"
	"brainz/common/permissions"
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"go.uber.org/zap"
)

type GroupQueryService interface {
	GetGroupsForInstitution(ctx context.Context, instID int64) ([]dtos.GroupDto, error)
}

type GroupAuthService interface {
	Authorize(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error)
}

type GroupHandler struct {
	gs GroupQueryService
	as GroupAuthService
}

func NewGroupHandler(gs *services.GroupService, as *services.AuthService) *GroupHandler {
	return &GroupHandler{gs, as}
}

var (
	_ GroupQueryService = (*services.GroupService)(nil)
	_ GroupAuthService  = (*services.AuthService)(nil)
)

func (gh *GroupHandler) GetGroupsForInst(ctx context.Context, c *app.RequestContext) {
	instID := c.Query("institution_id")

	if instID == "" {
		c.JSON(400, map[string]string{
			"error": "query parameter institution_id is required",
		})
		return
	}
	intID, err := strconv.Atoi(instID)
	if err != nil {
		c.JSON(400, map[string]string{
			"error": "query parameter institution_id must be an integer",
		})
		return
	}
	institutionID := int64(intID)

	isLegit, err := gh.as.Authorize(ctx, c, c.Request.Header.Get("X-Api-Key"), &permissions.Permission{
		Action:        permissions.ActionRead,
		InstitutionID: &institutionID,
	})
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
		zap.L().Error("Authorize group access", zap.Error(err))
		return
	}
	if !isLegit {
		c.Status(consts.StatusForbidden)
		return
	}

	groups, err := gh.gs.GetGroupsForInstitution(ctx, int64(intID))
	if err != nil {
		c.JSON(500, map[string]string{
			"error": "internal server error",
		})
		zap.L().Error("GetGroups", zap.Error(err))
		return
	}

	if len(groups) == 0 {
		c.Status(404)
		return
	}

	c.JSON(200, groups)
}
