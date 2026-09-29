package middleware

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/cloudwego/hertz/pkg/app"
)

func RecoveryMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("[PANIC RECOVERY] %v\n%s\n", r, debug.Stack())

				c.JSON(500, map[string]string{
					"error": "internal server error",
				})
				c.Abort()
			}
		}()

		c.Next(ctx)
	}
}
