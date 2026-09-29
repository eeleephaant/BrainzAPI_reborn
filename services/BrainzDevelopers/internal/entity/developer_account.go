package entity

import (
	"time"

	"github.com/google/uuid"
)

type DeveloperAccount struct {
	ID               uuid.UUID
	Email            string `json:"email" validate:"required,email"`
	EmailConfirmedAt *time.Time
	PasswordHash     []byte
	Salt             []byte
	TwoFactorSecret  *[]byte
	CreatedAt        time.Time
	BannedAt         *time.Time
	RoleId           uint `json:"role_id" validate:"required,email"`
}
