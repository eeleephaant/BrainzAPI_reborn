package handler

import (
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

	c.JSON(200, "")
}
