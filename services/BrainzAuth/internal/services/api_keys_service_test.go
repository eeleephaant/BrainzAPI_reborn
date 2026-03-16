package services

import "testing"

func TestResolveAPIKeyID_UUID(t *testing.T) {
	rawID := "123e4567-e89b-12d3-a456-426614174000"

	id, err := resolveAPIKeyID(rawID)
	if err != nil {
		t.Fatalf("resolveAPIKeyID returned error: %v", err)
	}

	if id.String() != rawID {
		t.Fatalf("expected id %q, got %q", rawID, id.String())
	}
}

func TestResolveAPIKeyID_RawKeyFormat(t *testing.T) {
	rawID := "123e4567-e89b-12d3-a456-426614174000"
	rawKey := "brainz_" + rawID + ":supersecret"

	id, err := resolveAPIKeyID(rawKey)
	if err != nil {
		t.Fatalf("resolveAPIKeyID returned error: %v", err)
	}

	if id.String() != rawID {
		t.Fatalf("expected id %q, got %q", rawID, id.String())
	}
}

func TestResolveAPIKeyID_Invalid(t *testing.T) {
	_, err := resolveAPIKeyID("not-a-key")
	if err == nil {
		t.Fatal("expected error for invalid key, got nil")
	}
}
