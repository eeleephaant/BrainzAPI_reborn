package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func GetInstitutions(ctx context.Context, c *app.RequestContext) {
	c.JSON(consts.StatusOK, "")
}

func AddInstitution(ctx context.Context, c *app.RequestContext) {
	c.JSON(201, "")
}
