package services

import (
	"brainz/developersapi/internal/entity"
	"brainz/developersapi/internal/repository"
	"brainz/developersapi/internal/security"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type EmailService struct {
	etr *repository.EmailTokensRepository
}

func NewEmailConfirmService(etr *repository.EmailTokensRepository) *EmailService {
	return &EmailService{etr: etr}
}

func (ecs *EmailService) SendConfirmationEmail(ctx context.Context, developerId uuid.UUID) (error, *entity.EmailConfirmationToken) {
	op := "EmailConfirmService.SendConfirmationEmail"

	numCode, err := security.GenerateRandomNumericCode(6)
	if err != nil {
		panic(fmt.Errorf("%s: generate NumbericCode: %w", op, err))
	}
	token, err := security.GenerateSecretKey(16)
	if err != nil {
		panic(fmt.Errorf("%s: generate token: %w", op, err))
	}

	emailConfToken := &entity.EmailConfirmationToken{
		DeveloperID:  developerId,
		Token:        token,
		NumbericCode: numCode,
		ExpiresAt:    time.Now().Add(30 * time.Minute),
	}

	emailConfToken, err = ecs.etr.Create(ctx, emailConfToken)
	if err != nil {
		return err, nil
	}
	return nil, emailConfToken
}
