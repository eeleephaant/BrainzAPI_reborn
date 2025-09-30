package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
