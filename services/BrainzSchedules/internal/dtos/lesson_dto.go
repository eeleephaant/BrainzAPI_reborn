package dtos

import "time"

type LessonDto struct {
	Id            int64     `json:"id"`
	Name          string    `json:"name"`
	CabNum        string    `json:"cab_num"`
	TeacherName   string    `json:"teacher_name"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Num           uint8     `json:"num"`
	GroupID       uint      `json:"group_id"`
	InstitutionID uint      `json:"institution_id"`
}
