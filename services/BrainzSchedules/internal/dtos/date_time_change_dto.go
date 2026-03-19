package dtos

import "time"

type DateTimeChangeDTO struct {
	InstitutionID uint64         `json:"institution_id"`
	Date          time.Time      `json:"date"`
	LessonTimings LessonsTimings `json:"lesson_timings"`
}
