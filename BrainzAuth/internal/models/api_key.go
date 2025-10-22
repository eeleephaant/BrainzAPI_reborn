package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
)

type ApiKey struct {
	gorm.Model
	ID          uuid.UUID `gorm:"primarykey"`
	Title       string    `json:"title" gorm:"size:100"`
	ApiKeyHash  string    `gorm:"size:64;uniqueIndex;unique"`
	Salt        string
	DeveloperId uuid.UUID  `gorm:"index"`
	ExpiresAt   *time.Time `gorm:"index"`
	LastUsedAt  *time.Time `gorm:"index"`
}
