package utils

import (
	"brainz/auth/internal/config"
	"brainz/auth/internal/dtos"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

var cfg *config.AuthConfig

func LoadCfg(c *config.AuthConfig) {
	cfg = c
}

func CheckPassword(rawPassword string, salt []byte, hashedPassword []byte) bool {
	hashedUserInput := GetHashArgon2(rawPassword, salt)
	return subtle.ConstantTimeCompare(hashedUserInput, hashedPassword) == 1
}

func GetRandomSalt() []byte {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		panic(fmt.Sprintf("failed to generate salt: %v", err))
	}
	return salt
}

func GetHashArgon2(apiKey string, salt []byte) []byte {
	hash := argon2.Key([]byte(apiKey+cfg.Pepper), salt, uint32(cfg.Argon2Time), uint32(cfg.Argon2Memory), uint8(cfg.Argon2Threads), 64)
	return hash
}

func GenerateSecretKey(length int) (string, error) {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}

func ExtractDataFromKey(rawKey string) (*dtos.ApiKeyData, error) {
	op := "crypto.ExtractDataFromKey"
	parts := strings.SplitN(rawKey, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("%s: invalid apiKey format", op)
	}
	cleanUUID := strings.TrimPrefix(parts[0], "brainz_")
	id, err := uuid.Parse(cleanUUID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid apiKey ID: %w", op, err)
	}
	var extractData *dtos.ApiKeyData = &dtos.ApiKeyData{
		ID:     id,
		Secret: parts[1],
	}
	return extractData, nil

}
