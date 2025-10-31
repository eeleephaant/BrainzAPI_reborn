package models

import (
	"time"

	"gorm.io/gorm"
)

type Lesson struct {
	gorm.Model
	Name          string      `json:"name" gorm:"size:100;uniqueIndex:unique_lesson"`
	CabNum        string      `json:"cab_num" gorm:"size:20"`
	TeacherName   string      `json:"teacher_name" gorm:"size:100;uniqueIndex:unique_lesson"`
	StartTime     time.Time   `json:"start_time" gorm:"index;uniqueIndex:unique_lesson"`
	EndTime       time.Time   `json:"end_time" gorm:"index"`
	Num           uint8       `json:"num" gorm:"uniqueIndex:unique_lesson"`
	GroupID       uint        `json:"group_id" gorm:"index;uniqueIndex:unique_lesson"`
	Group         Group       `gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	InstitutionID uint        `json:"institution_id" gorm:"index"`
	Institution   Institution `gorm:"foreignKey:InstitutionID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
