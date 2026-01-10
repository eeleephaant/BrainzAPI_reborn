package handler

import (
	"brainz-api/internal/services"
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/redis/go-redis/v9"
)

type ScheduleHandler struct {
	rc *redis.Client
	ls *services.LessonService
}

func NewScheduleHandler(rc *redis.Client, ls *services.LessonService) *ScheduleHandler {
	return &ScheduleHandler{rc, ls}
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
