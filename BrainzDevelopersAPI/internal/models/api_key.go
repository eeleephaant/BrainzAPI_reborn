package models

import "gorm.io/gorm"

type ApiKey struct {
	gorm.Model
	Title      string `json:"title" gorm:"size:100"`
	ApiKeyHash string `gorm:"size:64;uniqueIndex"`
	IpAddress  string `gorm:"size:45"`
}
