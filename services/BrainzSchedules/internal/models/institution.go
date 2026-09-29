package models

import "time"

type Institution struct {
	ID        int64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Name      string
	SiteLink  *string
}
