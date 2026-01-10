package handler

import (
	"brainz-api/internal/services"
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"go.uber.org/zap"
)

type InstitutionHandler struct {
	is *services.InstitutionService
}

func NewInstitutionHandler(is *services.InstitutionService) *InstitutionHandler {
	return &InstitutionHandler{is}
}

func (ih *InstitutionHandler) GetInstitutions(ctx context.Context, c *app.RequestContext) {
	instDTOs, err := ih.is.GetInstitutions(ctx)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": "internal server error"})
		zap.L().Error("Failed to get institutions", zap.Error(err))
		return
	}
	c.JSON(consts.StatusOK, instDTOs)
}
