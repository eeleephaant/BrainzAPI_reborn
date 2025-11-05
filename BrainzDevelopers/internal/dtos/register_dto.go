package dtos

import (
	"github.com/go-playground/validator/v10"
)

type RegisterDto struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=12,max=256"`
}

func (r *RegisterDto) Validate() ([]ValidationErrorResponse, error) {
	validate := validator.New()
	err := validate.Struct(r)
	if err == nil {
		return nil, nil
	}

	var errs []ValidationErrorResponse
	for _, e := range err.(validator.ValidationErrors) {
		var msg string
		switch e.Tag() {
		case "required":
			msg = "This field is required"
		case "email":
			msg = "Invalid email format"
		case "min":
			msg = "Value is too short"
		case "max":
			msg = "Value is too long"
		default:
			msg = "Invalid value"
		}

		errs = append(errs, ValidationErrorResponse{
			Field:   e.Field(),
			Message: msg,
		})
	}
	return errs, err
}
