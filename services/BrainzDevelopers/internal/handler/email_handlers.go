package handler

import (
	"brainz/developersapi/internal/dtos"
	"brainz/developersapi/internal/entity"
	"brainz/developersapi/internal/services"
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"
)

type EmailHandler struct {
	Es *services.EmailService
	Us *services.UserService
}

func NewEmailHandler(es *services.EmailService, us *services.UserService) *EmailHandler {
	return &EmailHandler{Es: es, Us: us}
}

func (h *EmailHandler) GetConfirmEmailCode(ctx context.Context, c *app.RequestContext) {
	op := "EmailHandler.GetConfirmEmailCode"

	emailConfirmData := dtos.ConfirmEmailCode{}

	if err := c.BindAndValidate(&emailConfirmData); err != nil {
		c.String(400, err.Error())
		return
	}

	ect, err := h.Es.ConfirmEmailCode(ctx, emailConfirmData.Token, emailConfirmData.Code)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrEmailCodeAlreadyUsed):
			c.JSON(401, map[string]string{"error": "code is already used"})
		case errors.Is(err, entity.ErrEmailCodeExpired):
			c.JSON(401, map[string]string{"error": "code is expired"})
		default:
			c.JSON(500, map[string]string{"error": "internal server error while h.Es.ConfirmEmailCode(ctx, confirmEmailDto.Token, confirmEmailDto.Code)"})
			zap.L().Error(op,
				zap.String("details", err.Error()),
			)
		}
		return
	}
	err = h.Us.ConfirmEmail(ctx, ect.DeveloperID)
	if err != nil {
		c.JSON(500, map[string]string{"error": "internal error while h.Us.ConfirmEmail(ctx, ect.DeveloperID)"})
		return
	}
	c.JSON(200, map[string]string{"status": "success"})
}
