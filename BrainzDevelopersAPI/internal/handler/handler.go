package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

func Register(ctx context.Context, c *app.RequestContext) {
	apiKey := c.Query("key")

}
