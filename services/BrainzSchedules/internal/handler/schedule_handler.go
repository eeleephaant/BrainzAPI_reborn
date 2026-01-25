package handler

import (
	"brainz-api/internal/dtos"
	"brainz-api/internal/services"
	"brainz/common/permissions"
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type ScheduleHandler struct {
	rc *redis.Client
	ls *services.LessonService
	as *services.AuthService
}

func NewScheduleHandler(rc *redis.Client, ls *services.LessonService, as *services.AuthService) *ScheduleHandler {
	return &ScheduleHandler{rc, ls, as}
}

func (sh *ScheduleHandler) AddLessons(ctx context.Context, c *app.RequestContext) {
	lsnCreateDTO := dtos.LessonsCreateDTO{}

	if err := c.BindAndValidate(&lsnCreateDTO); err != nil {
		c.String(400, err.Error())
		return
	}
	isLegit, err := sh.as.Authorize(ctx, c, c.Request.Header.Get("X-Api-Key"), &permissions.Permission{Action: permissions.ActionWrite, InstitutionID: &lsnCreateDTO.InstitutionID})

	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"message": "internal server error"})
		zap.L().Error("AddLessons failed", zap.Error(err))
		return
	}
	if !isLegit {
		c.Status(consts.StatusForbidden)
		return
	}

	err = sh.ls.AddLessons(ctx, lsnCreateDTO)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"message": "internal server error"})
		zap.L().Error("AddLessons failed", zap.Error(err))
		return
	}
	c.Status(consts.StatusCreated)

}

func (sh *ScheduleHandler) GetLessons(ctx context.Context, c *app.RequestContext) {
	instIDStr := c.Query("institution_id")
	dateStr := c.Query("date")
	if instIDStr == "" {
		c.JSON(http.StatusBadRequest, map[string]string{
			"error": "institution_id is required",
		})
		return
	}

	institutionID, err := strconv.ParseUint(instIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid institution_id format",
		})
		return
	}
	instID64 := int64(institutionID)
	isLegit, err := sh.as.Authorize(ctx, c, c.Request.Header.Get("X-Api-Key"), &permissions.Permission{Action: permissions.ActionRead, InstitutionID: &instID64})
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
		return
	}
	if !isLegit {
		c.Status(http.StatusForbidden)
		return
	}

	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid date format, expected YYYY-MM-DD",
		})
		return
	}

	lessons, err := sh.ls.GetForDateAndInstitution(ctx, targetDate, institutionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, lessons)
}
