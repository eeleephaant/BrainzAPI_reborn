package handler

import (
	"brainz-api/internal/dtos"
	"brainz-api/internal/services"
	"brainz/common/permissions"
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"go.uber.org/zap"
)

type InstitutionHandler struct {
	is *services.InstitutionService
	as *services.AuthService
}

func NewInstitutionHandler(is *services.InstitutionService, as *services.AuthService) *InstitutionHandler {
	return &InstitutionHandler{is, as}
}

func (ih *InstitutionHandler) CreateInstitution(ctx context.Context, c *app.RequestContext) {
	isLegit, err := ih.as.Authorize(ctx, c, c.Request.Header.Get("X-Api-Key"), &permissions.Permission{Action: permissions.ActionAdmin, InstitutionID: nil})
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": "internal server error"})
		zap.L().Error("Failed authorize token", zap.Error(err))
		return
	}
	if !isLegit {
		c.Status(consts.StatusForbidden)
		return
	}
	institutionCreateDto := dtos.InstitutionCreateDto{}

	if err := c.BindAndValidate(&institutionCreateDto); err != nil {
		c.String(400, err.Error())
		return
	}
	createdID, err := ih.is.CreateNew(ctx, institutionCreateDto)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": "internal server error"})
		zap.L().Error("Failed to create institution", zap.Error(err))
		return
	}
	c.JSON(consts.StatusCreated, map[string]int64{"id": createdID})
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
