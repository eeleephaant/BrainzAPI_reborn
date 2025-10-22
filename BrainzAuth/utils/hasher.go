package utils

import (
	"os"

	"golang.org/x/crypto/argon2"
)

func GetHashArgon2(apiKey string, salt []byte) []byte {
	pepper := os.Getenv("API_KEY_PEPPER")
	hash := argon2.Key([]byte(apiKey+pepper), salt, 3, 32*1024, 4, 64)
	return hash
}
