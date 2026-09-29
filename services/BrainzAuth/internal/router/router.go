package router

import (
	"brainz/auth/internal/handlers"
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"go.uber.org/zap"
)

func Register(h *server.Hertz, ah *handlers.AuthHandler, akmh *handlers.ApiKeysManagementHandler) {
	h.GET("/auth", ah.Auth)
	h.POST("/auth", ah.Auth)
	h.POST("/key", akmh.CreateApiKey)
	h.DELETE("/key", akmh.RemoveApiKey)
	h.GET("/keys/usage", akmh.GetDeveloperKeysUsage)
	h.GET("/keys", akmh.GetDeveloperApiKeys)

	h.NoRoute(func(ctx context.Context, c *app.RequestContext) {
		path := c.Request.URI().Path()
		zap.L().Warn("404 NoRoute", zap.ByteString("path", path), zap.ByteString("method", c.Method()))
		c.String(consts.StatusNotFound, "not found")
	})
}
