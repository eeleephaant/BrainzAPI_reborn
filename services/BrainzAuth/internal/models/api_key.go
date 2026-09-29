package models

import (
	"time"

	"github.com/google/uuid"
)

type ApiKey struct {
	ID          uuid.UUID
	Name        string
	ExpireAt    time.Time
	RevokedAt   *time.Time
	DeveloperID uuid.UUID
	KeyHash     []byte
	Salt        []byte
	PrefixRaw   string
	SuffixRaw   string
	CreatedAt   time.Time
}
