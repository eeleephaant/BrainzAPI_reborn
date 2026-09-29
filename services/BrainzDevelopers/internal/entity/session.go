package entity

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID          uuid.UUID
	DeveloperID uuid.UUID
	UserAgent   string
	IpAddress   string
	TokenHash   []byte
	Salt        []byte
	ExpiresAt   time.Time
	Revoked_at  *time.Time
}
