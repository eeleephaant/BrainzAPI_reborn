package dtos

import "github.com/google/uuid"

type ApiKeyData struct {
	ID     uuid.UUID
	Secret string
}
