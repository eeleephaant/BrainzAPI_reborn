package services

import (
	"brainz/developersapi/internal/dtos"
	"brainz/developersapi/internal/entity"
	"brainz/developersapi/internal/security"
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UserRepository defines persistence for developer accounts (for testing and DI).
type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*entity.DeveloperAccount, error)
	GetById(ctx context.Context, devId uuid.UUID) (*entity.DeveloperAccount, error)
	Create(ctx context.Context, developer_account *entity.DeveloperAccount) (*entity.DeveloperAccount, error)
	Update(ctx context.Context, user *entity.DeveloperAccount) (*entity.DeveloperAccount, error)
}

// EmailConfirmationSender sends email confirmation (for testing and DI).
type EmailConfirmationSender interface {
	SendConfirmationEmail(ctx context.Context, developerId uuid.UUID) (*entity.EmailConfirmationToken, error)
}

type UserService struct {
	ur  UserRepository
	ecs EmailConfirmationSender
}

func NewUserService(ur UserRepository, ecs EmailConfirmationSender) *UserService {
	return &UserService{ur: ur, ecs: ecs}
}

func (us *UserService) Authenticate(ctx context.Context, email string, password string) (*entity.DeveloperAccount, error) {
	account, err := us.ur.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if !security.CheckPassword(password, account.Salt, account.PasswordHash) {
		return nil, entity.ErrWrongCredentials
	}

	if account.BannedAt != nil {
		return nil, entity.ErrUserBanned
	}

	if account.EmailConfirmedAt == nil {
		return nil, entity.ErrEmailNotConfirmed
	}

	return account, nil
}

func (us *UserService) ConfirmEmail(ctx context.Context, devId uuid.UUID) error {
	account, err := us.ur.GetById(ctx, devId)
	if err != nil {
		return err
	}
	now := time.Now()
	account.EmailConfirmedAt = &now
	_, err = us.ur.Update(ctx, account)
	if err != nil {
		return err
	}
	return nil
}

func (us *UserService) RegistrateUser(ctx context.Context, registerDto *dtos.RegisterDto) (*entity.EmailConfirmationToken, error) {
	salt := security.GetRandomSalt()
	passwordHash := security.GetHashArgon2(registerDto.Password, salt)

	var user = entity.DeveloperAccount{
		Email:        registerDto.Email,
		PasswordHash: passwordHash,
		Salt:         salt,
		RoleId:       0,
	}

	acc, err := us.ur.Create(ctx, &user)
	if err != nil {
		return nil, err
	}

	ect, err := us.ecs.SendConfirmationEmail(ctx, acc.ID)

	zap.L().Debug("Sent confirmation email",
		zap.String("token", ect.Token),
		zap.String("code", ect.NumbericCode),
	)

	if err != nil {
		return nil, err
	}

	return ect, nil
}
