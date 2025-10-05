package models

import (
	"time"

	"gorm.io/gorm"
)

type Lesson struct {
	gorm.Model
	Name          string      `json:"name" gorm:"size:100"`
	CabNum        string      `json:"cab_num" gorm:"size:20"`
	TeacherName   string      `json:"teacher_name" gorm:"size:100"`
	StartTime     time.Time   `json:"start_time" gorm:"index"`
	EndTime       time.Time   `json:"end_time" gorm:"index"`
	Num           uint8       `json:"num" gorm:"size:2"`
	GroupID       uint        `json:"group_id" gorm:"index"`
	Group         Group       `gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	InstitutionID uint        `json:"institution_id" gorm:"index"`
	Institution   Institution `gorm:"foreignKey:InstitutionID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
