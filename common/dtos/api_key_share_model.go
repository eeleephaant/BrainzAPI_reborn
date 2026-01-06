package dtos

import (
	"time"

	"github.com/google/uuid"
)

type ApiKeyShareModel struct {
	Id        uuid.UUID  `json:"id"`
	Title     string     `json:"title"`
	Suffix    string     `json:"suffix"`
	Prefix    string     `json:"prefix"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at"`
}
