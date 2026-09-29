package models

import "time"

type Group struct {
	ID            int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
	Name          string
	InstitutionID *int64
}
