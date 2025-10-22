package models

import "gorm.io/gorm"

type Developer struct {
	gorm.Model
	Username      string `gorm:"unique;not null;type:varchar(64)"`
	Email         string `gorm:"unique;not null;type:varchar(128)"`
	EmailVerified bool   `gorm:"not null;default:false"`
	PasswordHash  string `gorm:"type:varchar(128);not null"`
}
