package dtos

import (
	"brainz/common/permissions"

	"github.com/google/uuid"
)

type ApiKeyCreateDto struct {
	ApiKeyName  string                   `json:"api_key_name,required"`
	DevUUID     uuid.UUID                `json:"dev_uuid,required"`
	Permissions []permissions.Permission `json:"permissions"`
	IPWhitelist []string                 `json:"ip_whitelist"`
}
