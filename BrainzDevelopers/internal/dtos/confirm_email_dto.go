package dtos

type ConfirmEmailCode struct {
	Token string `json:"token" validate:"required,len=32"`
	Code  string `json:"code" validate:"required,numeric,len=6"`
}
