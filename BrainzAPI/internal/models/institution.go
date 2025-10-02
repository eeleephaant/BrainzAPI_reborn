package models

import "gorm.io/gorm"

type Institution struct {
	gorm.Model
	Name string `json:"name" gorm:"size:100;index"`
	Site string `json:"site_link" gorm:"size:255"`
}
