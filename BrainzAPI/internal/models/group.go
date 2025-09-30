package models

import "gorm.io/gorm"

type Group struct {
	gorm.Model
	Name          string      `json:"name"`
	InstitutionID uint        `json:"institution_id"`
	Institution   Institution `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
