package dtos

import (
	"time"

	"github.com/google/uuid"
)

type ApiKeyUsage struct {
	ApiKeyId      uuid.UUID    `json:"api_key_id"`
	Title         string       `json:"name"`
	CreatedAt     time.Time    `json:"created_at"`
	LastUsage     time.Time    `json:"last_used_at"`
	TotalRequests uint64       `json:"total_usage"`
	UsageByDay    []DailyUsage `json:"usage_by_day"`
}
