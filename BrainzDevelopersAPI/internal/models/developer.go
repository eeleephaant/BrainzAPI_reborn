package models

type Developer struct {
	Username     string `gorm:"unique;not null"`
	PasswordHash string `gorm:"length:128;not null"`
}
