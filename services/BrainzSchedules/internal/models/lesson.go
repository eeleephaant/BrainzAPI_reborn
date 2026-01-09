package models

import "time"

type Lesson struct {
	ID            int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
	Name          string
	CabNum        *string
	TeacherName   string
	StartTime     time.Time
	EndTime       time.Time
	Num           int16
	GroupID       *int64
	InstitutionID *int64
}
