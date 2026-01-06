package models

import (
	"time"
)

type ApiKeyUsage struct {
	ID        uint
	ApiKeyID  uint
	Endpoint  string
	Method    string
	Status    int
	LatencyMs int
	IP        string
	CreatedAt time.Time
}
