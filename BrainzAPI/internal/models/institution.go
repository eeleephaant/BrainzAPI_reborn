package models

import "gorm.io/gorm"

type Institution struct {
	gorm.Model
	Name string `json:"name"`
	Site string `json:"site_link"`
}
