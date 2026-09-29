package dtos

import "time"

type ApiKeyCreateResponse struct {
	Name      string    `json:"name,required"`
	ApiKey    string    `json:"api_key,required"`
	ExpiresAt time.Time `json:"expires_at,required"`
}
