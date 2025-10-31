package dtos

import (
	"github.com/go-playground/validator/v10"
)

type LoginDto struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=12,max=256"`
}

func (r *LoginDto) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
