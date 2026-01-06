package dtos

import "time"

type OnceApiKeyDto struct {
	ApiKeyRaw string    `json:"api_key_raw"`
	ExpireAt  time.Time `json:"expire_at"`
}
