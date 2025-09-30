package handler

import (
	"brainz-api/internal/db"
	"brainz-api/internal/models"
	"brainz-api/internal/models/mappers"
	"brainz-api/internal/models/requests"
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func GetInstitutions(ctx context.Context, c *app.RequestContext) {
	var institutions []*models.Institution

	if err := db.DB.Find(&institutions).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if len(institutions) < 1 {
		c.NotFound()
		return
	}

	institutionsDto := mappers.InstitutionsToDTOs(institutions)

	c.JSON(consts.StatusOK, institutionsDto)
}

func AddInstitution(ctx context.Context, c *app.RequestContext) {
	requestData := requests.AddInstitutionRequest{}

	if err := c.BindAndValidate(&requestData); err != nil {
		c.String(400, err.Error())
		return
	}

	institution := models.Institution{Name: requestData.Name, Site: requestData.SiteLink}

	if err := db.DB.Create(&institution).Error; err != nil {
		c.String(500, "failed to create institution: %v", err)
		return
	}

	c.JSON(201, map[string]interface{}{
		"id":   institution.ID,
		"name": institution.Name,
	})
}
