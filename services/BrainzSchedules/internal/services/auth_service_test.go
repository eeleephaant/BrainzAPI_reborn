package services

import (
	"brainz/common/permissions"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
)

func TestAuthService_Authorize(t *testing.T) {
	makeReqCtx := func() *app.RequestContext {
		c := app.NewContext(0)
		c.Request.SetRequestURI("/lessons")
		c.Request.SetMethod(http.MethodGet)
		c.Request.SetHost("brainz-schedules")
		return c
	}

	t.Run("missing_api_key", func(t *testing.T) {
		baseURL, _ := url.Parse("http://127.0.0.1:1")
		svc := NewAuthService(baseURL)

		_, err := svc.Authorize(context.Background(), makeReqCtx(), "", &permissions.Permission{Action: permissions.ActionRead})
		if err == nil || !strings.Contains(err.Error(), "api key is required") {
			t.Fatalf("error = %v, want api key is required", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body := make([]byte, r.ContentLength)
			_, _ = r.Body.Read(body)
			if !strings.Contains(string(body), `"perm"`) {
				t.Fatalf("request body should contain perm wrapper, got %s", string(body))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":true,"error_message":""}`))
		}))
		defer srv.Close()

		baseURL, _ := url.Parse(srv.URL)
		svc := NewAuthService(baseURL)

		ok, err := svc.Authorize(context.Background(), makeReqCtx(), "test-api-key", &permissions.Permission{Action: permissions.ActionRead})
		if err != nil {
			t.Fatalf("Authorize error = %v", err)
		}
		if !ok {
			t.Fatal("Authorize returned false, want true")
		}
	})

	t.Run("non_200_response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "forbidden", http.StatusForbidden)
		}))
		defer srv.Close()

		baseURL, _ := url.Parse(srv.URL)
		svc := NewAuthService(baseURL)

		ok, err := svc.Authorize(context.Background(), makeReqCtx(), "test-api-key", &permissions.Permission{Action: permissions.ActionRead})
		if err != nil {
			t.Fatalf("Authorize error = %v", err)
		}
		if ok {
			t.Fatal("Authorize returned true, want false")
		}
	})

	t.Run("invalid_json_response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{invalid json`))
		}))
		defer srv.Close()

		baseURL, _ := url.Parse(srv.URL)
		svc := NewAuthService(baseURL)

		_, err := svc.Authorize(context.Background(), makeReqCtx(), "test-api-key", &permissions.Permission{Action: permissions.ActionRead})
		if err == nil || !strings.Contains(err.Error(), "failed to decode auth response") {
			t.Fatalf("error = %v, want decode error", err)
		}
	})
}
