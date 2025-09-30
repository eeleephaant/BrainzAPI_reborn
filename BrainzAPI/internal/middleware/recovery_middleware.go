package middleware

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
)

func RecoveryMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if r := recover(); r != nil {
				c.JSON(500, map[string]any{
					"error": fmt.Sprintf("internal error: %v", r),
				})
			}
		}()
		c.Next(ctx)
	}
}
