package models

import "gorm.io/gorm"

type Group struct {
	gorm.Model
	Name          string      `json:"name" gorm:"size:50"`
	InstitutionID uint        `json:"institution_id" gorm:"index"`
	Institution   Institution `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
