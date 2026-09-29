package dtos

import (
	"brainz/developersapi/internal/entity"
	"time"

	"github.com/google/uuid"
)

// ProfileResponse is the public developer account shape (no secrets).
type ProfileResponse struct {
	ID               uuid.UUID  `json:"id"`
	Email            string     `json:"email"`
	EmailConfirmedAt *time.Time `json:"email_confirmed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	RoleId           uint       `json:"role_id"`
	BannedAt         *time.Time `json:"banned_at,omitempty"`
	TwoFactorEnabled bool       `json:"two_factor_enabled"`
}

func NewProfileResponse(a *entity.DeveloperAccount) ProfileResponse {
	tfa := a.TwoFactorSecret != nil && len(*a.TwoFactorSecret) > 0
	return ProfileResponse{
		ID:               a.ID,
		Email:            a.Email,
		EmailConfirmedAt: a.EmailConfirmedAt,
		CreatedAt:        a.CreatedAt,
		RoleId:           a.RoleId,
		BannedAt:         a.BannedAt,
		TwoFactorEnabled: tfa,
	}
}
