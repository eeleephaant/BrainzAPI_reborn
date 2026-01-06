package dtos

import "github.com/google/uuid"

type ApiKeyRemoveDto struct {
	KeyUUID uuid.UUID `json:"key_uuid,required"`
}
