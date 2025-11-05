package services

import (
	"brainz/developersapi/internal/entity"
	"brainz/developersapi/internal/repository"
)

type UserService struct {
	ur *repository.UserRepository
}

func NewUserService(ur *repository.UserRepository) *UserService {
	return &UserService{ur: ur}
}

func (us *UserService) GetUserByEmail(ctx context.Context, email string) (entity.DeveloperAccount, error) {
	op := "UserService.GetUserByEmail"
	err, user := us.ur.GetByEmail(ctx context.Context, email string); if err != nil {
		zap.L().Error(op,
			zap.String("message", "error acquiring database connection"),
			zap.String("details", err.Error()),
		)
		return nil, err
	} else {
		return user.Email, nil
	}
}
