package handler

import (
	"brainz-api/internal/db"
	models "brainz-api/internal/models/db_models"
	"brainz-api/internal/models/mappers"
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

func GetGroups(ctx context.Context, c *app.RequestContext) {
	instID := c.Query("institution_id")

	if instID == "" {
		c.JSON(400, map[string]string{
			"error": "query parameter institution_id is required",
		})
		return
	}

	var groups []*models.Group
	if err := db.DB.Where("institution_id = ?", instID).Find(&groups).Error; err != nil {
		c.JSON(500, map[string]any{
			"error": err.Error(),
		})
		return
	}

	groupsDto := mappers.GroupsToDTOs(groups)

	c.JSON(200, groupsDto)
}
