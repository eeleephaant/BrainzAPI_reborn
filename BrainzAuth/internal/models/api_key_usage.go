package models

import (
	"time"

	"gorm.io/gorm"
)

type ApiKeyUsage struct {
	gorm.Model
	ID        uint   `gorm:"primaryKey"`
	ApiKeyID  uint   `gorm:"index"`
	Endpoint  string `gorm:"size:255"`
	Method    string `gorm:"size:10"`
	Status    int
	LatencyMs int
	IP        string `gorm:"size:45"`
	CreatedAt time.Time
}
