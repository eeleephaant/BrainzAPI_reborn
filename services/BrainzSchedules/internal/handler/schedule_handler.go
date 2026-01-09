package handler

import (
	"brainz-api/internal/db"
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/redis/go-redis/v9"
)

func GetLessonsCount(ctx context.Context, c *app.RequestContext) {

	c.JSON(200, "")
}

func GetAvailableScheduleDays(ctx context.Context, c *app.RequestContext) {
	startDateStr := string(c.Query("start_date"))
	endDateStr := string(c.Query("end_date"))
	institution_id := string(c.Query("institution_id"))

	if startDateStr == "" || endDateStr == "" {
		c.JSON(400, map[string]string{"error": "start_date and end_date are required"})
		return
	}

	startDate, err1 := time.Parse("2006-01-02", startDateStr)
	endDate, err2 := time.Parse("2006-01-02", endDateStr)
	if err1 != nil || err2 != nil {
		c.JSON(400, map[string]string{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	endDate = endDate.AddDate(0, 0, 1)

	c.JSON(200, "")
}

func GetLessons(ctx context.Context, c *app.RequestContext) {
	instID := c.Query("institution_id")
	dateStr := c.Query("date")

	var institutionID uint
	if instID != "" {
		_, err := fmt.Sscan(instID, &institutionID)
		if err != nil {
			c.JSON(500, map[string]string{"error": err.Error()})
		}
	}

	var startOfDay, endOfDay time.Time
	if dateStr != "" {
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(500, map[string]string{"error": err.Error()})
		}
		startOfDay = t
		endOfDay = t.Add(24 * time.Hour)
	}

	// c.Data(200, "application/json", "")

	// db.RedisClient.Set(ctx, key, jsonData, time.Hour*1)
}
