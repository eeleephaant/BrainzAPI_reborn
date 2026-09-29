package dtos

import (
	"time"

	"github.com/google/uuid"
)

// ApiKeyUsageDaily is request counts aggregated by calendar day (UTC).
type ApiKeyUsageDaily struct {
	Date  time.Time `json:"date"`
	Count uint32    `json:"count"`
}

// ApiKeyUsageStats is per-key usage for developer dashboards.
type ApiKeyUsageStats struct {
	ApiKeyID   uuid.UUID          `json:"api_key_id"`
	Name       string             `json:"name"`
	CreatedAt  time.Time          `json:"created_at"`
	LastUsedAt *time.Time         `json:"last_used_at,omitempty"`
	TotalUsage uint64             `json:"total_usage"`
	UsageByDay []ApiKeyUsageDaily `json:"usage_by_day"`
}

// ApiKeysUsageResponse is the JSON body for GET .../keys/usage.
type ApiKeysUsageResponse struct {
	Keys []ApiKeyUsageStats `json:"keys"`
}
