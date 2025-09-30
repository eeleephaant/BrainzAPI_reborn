package models

import (
	"time"

	"gorm.io/gorm"
)

type Lesson struct {
	gorm.Model
	Name          string      `json:"name"`
	CabNum        string      `json:"cab_num"`
	TeacherName   string      `json:"teacher_name"`
	StartTime     time.Time   `json:"start_time"`
	EndTime       time.Time   `json:"end_time"`
	Num           uint        `json:"num"`
	GroupID       uint        `json:"group_id"`
	Group         Group       `gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	InstitutionID uint        `json:"institution_id"`
	Institution   Institution `gorm:"foreignKey:InstitutionID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
