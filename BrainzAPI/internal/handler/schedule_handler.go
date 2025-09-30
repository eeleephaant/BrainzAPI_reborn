package handler

import (
	"brainz-api/internal/db"
	"brainz-api/internal/models"
	"brainz-api/internal/models/dtos"
	"brainz-api/internal/models/mappers"
	"context"
	"encoding/json"
	"fmt"
	"time"

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

	key := "lessons:" + dateStr + ":" + instID
	val, err := db.RedisClient.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		c.JSON(500, map[string]string{"error": err.Error()})
	}

	if err == nil {
		if err := json.Unmarshal([]byte(val), &lessonDTOs); err != nil {
			c.JSON(500, map[string]string{"error": err.Error()})
		}
		c.JSON(200, lessonDTOs)
		return
	}

	query := db.DB.Preload("Group").Preload("Institution").Model(&models.Lesson{})

	if institutionID != 0 {
		query = query.Where("institution_id = ?", institutionID)
	}

	if !startOfDay.IsZero() {
		query = query.Where("start_time >= ? AND start_time < ?", startOfDay, endOfDay)
	}

	if err := query.Find(&lessons).Error; err != nil {
		c.JSON(500, map[string]string{"error": err.Error()})
	}

	if len(lessons) < 1 {
		c.NotFound()
		return
	}

	lessonDTOs = mappers.LessonsToDTOs(lessons)

	jsonData, _ := json.Marshal(lessonDTOs)
	db.RedisClient.Set(ctx, key, jsonData, time.Hour*1)

	c.JSON(200, lessonDTOs)
}
