package dtos

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type ApiKeyShareModel struct {
	Id        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	ExpiresAt time.Time `json:"expires_at"`
	LastUsed  time.Time `json:"last_used"`
}
