package dtos

import "time"

type ApiKeyCreateResponse struct {
	ApiKey    string    `json:"api_key"`
	ExpiresAt time.Time `json:"expires_at"`
}
