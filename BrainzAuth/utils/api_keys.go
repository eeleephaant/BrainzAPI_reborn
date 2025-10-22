package utils

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"

	"github.com/gofrs/uuid/v5"
)

func GenerateRawApiKey(id uuid.UUID) (string, []byte, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", nil, err
	}

	rawKey := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	fullKey := fmt.Sprintf("%s:%s", id, rawKey)

	return fullKey, randomBytes, nil
}
