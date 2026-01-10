package handler

import (
	"brainz-api/internal/services"
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"
)

type GroupHandler struct {
	gs *services.GroupService
}

func NewGroupHandler(gs *services.GroupService) *GroupHandler {
	return &GroupHandler{gs}
}

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
	groups, err := gh.gs.GetGroupsForInstitution(ctx, int64(intID))
	if err != nil {
		c.JSON(500, map[string]string{
			"error": "internal server error",
		})
		zap.L().Error("GetGroups", zap.Error(err))
		return
	}

	if len(groups) == 0 {
		c.JSON(404, map[string]string{
			"error": "groups not found",
		})
		return
	}

	c.JSON(200, groups)
}
