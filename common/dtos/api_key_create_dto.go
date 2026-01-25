package dtos

import "github.com/google/uuid"

type ApiKeyRemoveDto struct {
	Key     string    `json:"key,required"`
	DevUUID uuid.UUID `json:"dev_uuid,required"`
}
