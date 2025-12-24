package services

import (
	"brainz/developersapi/internal/dtos"
	"brainz/developersapi/internal/entity"
	"brainz/developersapi/internal/repository"
	"brainz/developersapi/internal/security"
	"context"
	"time"

	"github.com/google/uuid"
)

type UserService struct {
	ur  *repository.UserRepository
	ecs *EmailService
}

func NewUserService(ur *repository.UserRepository) *UserService {
	return &UserService{ur: ur}
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
	*account.EmailConfirmedAt = time.Now()
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

	err, ect := us.ecs.SendConfirmationEmail(ctx, acc.ID)

	if err != nil {
		return nil, err
	}

	return ect, nil
}
