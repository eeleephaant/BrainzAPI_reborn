package dtos

type AuthResult struct {
	Status       bool   `json:"status"`
	ErrorMessage string `json:"error_message"`
}
