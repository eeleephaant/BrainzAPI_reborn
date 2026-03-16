package utils

import (
	"brainz/auth/internal/config"
	"encoding/hex"
	"testing"
)

func TestGenerateSecretKey_LengthAndHex(t *testing.T) {
	key, err := GenerateSecretKey(32)
	if err != nil {
		t.Fatalf("GenerateSecretKey returned error: %v", err)
	}

	if len(key) != 64 {
		t.Fatalf("expected hex key length 64, got %d", len(key))
	}

	if _, err := hex.DecodeString(key); err != nil {
		t.Fatalf("generated key is not valid hex: %v", err)
	}
}

func TestExtractDataFromKey_Valid(t *testing.T) {
	raw := "brainz_123e4567-e89b-12d3-a456-426614174000:supersecret"

	data, err := ExtractDataFromKey(raw)
	if err != nil {
		t.Fatalf("ExtractDataFromKey returned error: %v", err)
	}

	if data.Secret != "supersecret" {
		t.Fatalf("expected secret %q, got %q", "supersecret", data.Secret)
	}

	expectedID := "123e4567-e89b-12d3-a456-426614174000"
	if data.ID.String() != expectedID {
		t.Fatalf("expected id %q, got %q", expectedID, data.ID.String())
	}
}

func TestExtractDataFromKey_InvalidFormat(t *testing.T) {
	_, err := ExtractDataFromKey("invalid-key-format")
	if err == nil {
		t.Fatal("expected error for invalid key format, got nil")
	}
}

func TestCheckPassword(t *testing.T) {
	LoadCfg(&config.AuthConfig{
		Pepper:        "test-pepper",
		Argon2Memory:  64,
		Argon2Time:    1,
		Argon2Threads: 1,
	})

	salt := []byte("1234567890abcdef")
	hash := GetHashArgon2("my-secret", salt)

	if !CheckPassword("my-secret", salt, hash) {
		t.Fatal("expected CheckPassword to accept valid password")
	}

	if CheckPassword("wrong-secret", salt, hash) {
		t.Fatal("expected CheckPassword to reject invalid password")
	}
}

func TestGetRandomSalt_Length(t *testing.T) {
	salt := GetRandomSalt()
	if len(salt) != 16 {
		t.Fatalf("expected salt length 16, got %d", len(salt))
	}
}
