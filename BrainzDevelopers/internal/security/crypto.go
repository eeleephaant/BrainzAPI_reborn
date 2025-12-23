package security

import (
	"brainz/developersapi/internal/config"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"

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
