package dtos

import (
	"github.com/go-playground/validator/v10"
)

type RegisterDto struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=12,max=256"`
}

func (r *RegisterDto) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
