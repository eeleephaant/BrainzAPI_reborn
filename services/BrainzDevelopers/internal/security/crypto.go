package security

import (
	"brainz/developersapi/internal/config"
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

func CheckPassword(rawPassword string, salt, hashedPassword []byte) bool {
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

func GenerateRandomNumericCode(length int) (string, error) {
	const digits = "0123456789"
	bytes := make([]byte, length)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	for i := 0; i < length; i++ {
		bytes[i] = digits[bytes[i]%10]
	}

	return string(bytes), nil
}

func GenerateSecretKey(length int) (string, error) {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}

func ExtractDataFromToken(token string) (uuid.UUID, string, error) {
	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 {
		return uuid.Nil, "", fmt.Errorf("invalid token format")
	}

	sessionID, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("invalid session id: %w", err)
	}

	secretKey := parts[1]
	if secretKey == "" {
		return uuid.Nil, "", fmt.Errorf("empty secret key")
	}

	return sessionID, secretKey, nil
}
