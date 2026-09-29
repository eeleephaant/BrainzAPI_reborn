package entity

import (
	"time"

	"github.com/google/uuid"
)

type EmailConfirmationToken struct {
	ID           uuid.UUID
	DeveloperID  uuid.UUID
	Token        string
	NumbericCode string
	ExpiresAt    time.Time
	UsedAt       *time.Time
	CreatedAt    time.Time
}
