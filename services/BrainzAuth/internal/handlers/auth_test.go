package handlers

import (
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
)

func TestExtractApiKey_FromAuthorizationBearer(t *testing.T) {
	ctx := app.NewContext(0)
	ctx.Request.Header.Set("Authorization", "Bearer key-from-bearer")

	got := extractApiKey(ctx)
	if got != "key-from-bearer" {
		t.Fatalf("expected %q, got %q", "key-from-bearer", got)
	}
}

func TestExtractApiKey_FromXAPIKey(t *testing.T) {
	ctx := app.NewContext(0)
	ctx.Request.Header.Set("X-API-Key", "key-from-x-api-key")

	got := extractApiKey(ctx)
	if got != "key-from-x-api-key" {
		t.Fatalf("expected %q, got %q", "key-from-x-api-key", got)
	}
}

func TestExtractApiKey_FromXApiKey(t *testing.T) {
	ctx := app.NewContext(0)
	ctx.Request.Header.Set("X-Api-Key", "key-from-x-api-key-camel")

	got := extractApiKey(ctx)
	if got != "key-from-x-api-key-camel" {
		t.Fatalf("expected %q, got %q", "key-from-x-api-key-camel", got)
	}
}

func TestExtractApiKey_AuthorizationHasPriority(t *testing.T) {
	ctx := app.NewContext(0)
	ctx.Request.Header.Set("Authorization", "Bearer key-from-bearer")
	ctx.Request.Header.Set("X-API-Key", "key-from-header")

	got := extractApiKey(ctx)
	if got != "key-from-bearer" {
		t.Fatalf("expected %q, got %q", "key-from-bearer", got)
	}
}

func TestExtractApiKey_EmptyWhenMissing(t *testing.T) {
	ctx := app.NewContext(0)

	got := extractApiKey(ctx)
	if got != "" {
		t.Fatalf("expected empty key, got %q", got)
	}
}
