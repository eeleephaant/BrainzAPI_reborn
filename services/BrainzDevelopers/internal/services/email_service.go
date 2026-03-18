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
	etr    *repository.EmailTokensRepository
	users  *repository.UserRepository
	mailer Mailer
}

func NewEmailConfirmService(etr *repository.EmailTokensRepository, users *repository.UserRepository, mailer Mailer) *EmailService {
	return &EmailService{
		etr:    etr,
		users:  users,
		mailer: mailer,
	}
}

func (ecs *EmailService) SendConfirmationEmail(ctx context.Context, developerId uuid.UUID) (*entity.EmailConfirmationToken, error) {
	op := "EmailService.SendConfirmationEmail"

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
		return nil, err
	}

	// Load developer email to send code.
	account, err := ecs.users.GetById(ctx, developerId)
	if err != nil {
		return nil, fmt.Errorf("%s: load developer: %w", op, err)
	}

	subject := "Email confirmation code"
	body := fmt.Sprintf("Your confirmation code is: %s\nIt will expire in 30 minutes.", emailConfToken.NumbericCode)

	if err := ecs.mailer.Send(ctx, account.Email, subject, body); err != nil {
		return nil, fmt.Errorf("%s: send email: %w", op, err)
	}

	return emailConfToken, nil
}

func (ecs *EmailService) ConfirmEmailCode(ctx context.Context, token string, code string) (*entity.EmailConfirmationToken, error) {
	ect, err := ecs.etr.GetByTokenAndCode(ctx, token, code)
	if err != nil {
		return nil, err
	}

	if time.Now().After(ect.ExpiresAt) {
		return nil, entity.ErrEmailCodeExpired
	}

	if ect.UsedAt != nil {
		return nil, entity.ErrEmailCodeAlreadyUsed
	}

	now := time.Now()
	ect.UsedAt = &now
	ect, err = ecs.etr.Update(ctx, ect)
	if err != nil {
		return nil, err
	}

	return ect, nil

}
