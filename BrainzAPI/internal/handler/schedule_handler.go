package handler

import (
	"brainz-api/internal/db"
	"brainz-api/internal/models"
	"brainz-api/internal/models/dtos"
	"brainz-api/internal/models/mappers"
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/redis/go-redis/v9"
)

func GetLessonsCount(ctx context.Context, c *app.RequestContext) {
	var count int64

	if err := db.DB.Model(&models.Lesson{}).Count(&count).Error; err != nil {
		c.JSON(500, map[string]any{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, map[string]int64{
		"count": count,
	})
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

	var days []time.Time
	if err := db.DB.Model(&models.Lesson{}).
		Select("DISTINCT DATE(start_time) as day").
		Where("start_time >= ? AND start_time < ? AND institution_id = ?", startDate, endDate, institution_id).
		Order("day").
		Pluck("day", &days).Error; err != nil {
		c.JSON(500, map[string]string{"error": err.Error()})
		return
	}

	result := make([]string, 0, len(days))
	for _, d := range days {
		result = append(result, d.Format("2006-01-02"))
	}

	c.JSON(200, result)
}

func GetLessons(ctx context.Context, c *app.RequestContext) {
	var lessons []*models.Lesson
	var lessonDTOs []dtos.LessonDto

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

	key := "lsn:" + dateStr + ":" + instID

	val, err := db.RedisClient.Get(ctx, key).Bytes()
	if err != nil && err != redis.Nil {
		c.JSON(500, map[string]string{"error": err.Error()})
	}

	if err == nil {
		println("Redis hit schedule")
		c.Data(200, "application/json", val)
		return
	}

	println("Redis miss")
	query := db.DB.Preload("Group").Preload("Institution").Model(&models.Lesson{})

	if institutionID != 0 {
		query = query.Where("institution_id = ?", institutionID)
	}

	if !startOfDay.IsZero() {
		query = query.Where("start_time >= ? AND start_time < ?", startOfDay, endOfDay)
	}

	if err := query.Find(&lessons).Error; err != nil {
		c.JSON(500, map[string]string{"error": err.Error()})
		return
	}

	if len(lessons) < 1 {
		c.NotFound()
		return
	}

	lessonDTOs = mappers.LessonsToDTOs(lessons)
	jsonData, _ := sonic.Marshal(lessonDTOs)

	c.Data(200, "application/json", jsonData)

	db.RedisClient.Set(ctx, key, jsonData, time.Hour*1)
}
