package models

import (
	"time"

	"github.com/google/uuid"
)

type ApiKeyUsage struct {
	ID           uuid.UUID
	ApiKeyID     uuid.UUID
	Endpoint     string
	ResponseCode string
	UsageAt      time.Time
}
