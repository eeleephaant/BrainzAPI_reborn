package handler

import (
	"brainz/auth/internal/db"
	"brainz/auth/internal/models"
	"brainz/auth/internal/models/dtos"
	"brainz/auth/utils"
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

func Auth(ctx context.Context, c *app.RequestContext) {
	apiKey := c.Query("key")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, dtos.AuthResult{
			Status:       false,
			ErrorMessage: "missing apiKey",
		})
		return
	}

	hashedKey := utils.HashKey(apiKey)

	var key models.ApiKey

	if err := db.DB.Where("api_key_hash = ?", hashedKey).First(&key).Error; err != nil {
		c.JSON(http.StatusUnauthorized, dtos.AuthResult{
			Status:       false,
			ErrorMessage: "invalid apiKey",
		})
		return
	}

	if key.IpAddress != "" {
		clientIP := c.ClientIP()
		if clientIP != key.IpAddress {
			c.JSON(http.StatusUnauthorized, dtos.AuthResult{
				Status:       false,
				ErrorMessage: "ip mismatch",
			})
			return
		}
	}

	c.JSON(http.StatusOK, dtos.AuthResult{
		Status:       true,
		ErrorMessage: "success",
	})

}
