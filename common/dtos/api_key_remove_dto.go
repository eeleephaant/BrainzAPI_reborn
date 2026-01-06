package dtos

import "github.com/google/uuid"

type ApiKeyCreateDto struct {
	Name        string    `json:"api_key,required"`
	DeveloperID uuid.UUID `json:"developer_id,required"`
}
