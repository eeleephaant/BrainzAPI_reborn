package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func urlQueryEscape(s string) string {
	// API keys contain ':' which is safe in query values; avoid escaping to keep
	// upstream parsing/validation consistent.
	return s
}

const (
	authBaseURL         = "http://127.0.0.1:8081"
	developersBaseURL   = "http://127.0.0.1:3333"
	schedulesBaseURL    = "http://127.0.0.1:18080"
	composeProject      = "brainz-e2e"
	stackStartupTimeout = 6 * time.Minute
	serviceReadyTimeout = 2 * time.Minute
)

var repoRoot string

type registerResponse struct {
	Message string `json:"message"`
}

type loginResponse struct {
	Token string `json:"token"`
}

type apiKeyCreateResponse struct {
	Name      string    `json:"name"`
	APIKey    string    `json:"api_key"`
	ExpiresAt time.Time `json:"expires_at"`
}

type institutionResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Site string `json:"site_link"`
}

type groupResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type lessonResponse struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	CabNum        string    `json:"cab_num"`
	TeacherName   string    `json:"teacher_name"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Num           uint8     `json:"num"`
	GroupID       uint      `json:"group_id"`
	InstitutionID uint      `json:"institution_id"`
}

type authResult struct {
	Status       bool   `json:"status"`
	ErrorMessage string `json:"error_message"`
}

type apiKeyUsageDaily struct {
	Date  string `json:"date"`
	Count uint32 `json:"count"`
}

type apiKeyUsageEntry struct {
	ApiKeyID   string             `json:"api_key_id"`
	Name       string             `json:"name"`
	CreatedAt  string             `json:"created_at"`
	LastUsedAt *string            `json:"last_used_at"`
	TotalUsage uint64             `json:"total_usage"`
	UsageByDay []apiKeyUsageDaily `json:"usage_by_day"`
}

type apiKeysUsageResponse struct {
	Keys []apiKeyUsageEntry `json:"keys"`
}

func TestMain(m *testing.M) {
	root, err := findRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "locate repo root: %v\n", err)
		os.Exit(1)
	}
	repoRoot = root

	// Host port for schedules (avoid clashing with another stack on :8080). Must match schedulesBaseURL.
	_ = os.Setenv("SCHEDULES_PORT", "18080")

	_, _ = runCompose(2*time.Minute, "down", "-v", "--remove-orphans")

	for _, service := range []string{"brainz-auth", "brainz-developers", "brainz-schedules"} {
		if _, err := runCompose(stackStartupTimeout, "build", service); err != nil {
			fmt.Fprintf(os.Stderr, "compose build for %s failed: %v\n", service, err)
			_, _ = runCompose(2*time.Minute, "down", "-v", "--remove-orphans")
			os.Exit(1)
		}
	}

	if _, err := runCompose(stackStartupTimeout, "up", "-d", "db", "redis", "brainz-auth", "brainz-developers", "brainz-schedules"); err != nil {
		fmt.Fprintf(os.Stderr, "compose up failed: %v\n", err)
		_, _ = runCompose(2*time.Minute, "down", "-v", "--remove-orphans")
		os.Exit(1)
	}

	if err := waitForTCP("127.0.0.1:3333", serviceReadyTimeout); err != nil {
		fmt.Fprintf(os.Stderr, "developers port not ready: %v\n", err)
		_, _ = runCompose(2*time.Minute, "down", "-v", "--remove-orphans")
		os.Exit(1)
	}
	if err := waitForTCP("127.0.0.1:8081", serviceReadyTimeout); err != nil {
		fmt.Fprintf(os.Stderr, "auth port not ready: %v\n", err)
		_, _ = runCompose(2*time.Minute, "down", "-v", "--remove-orphans")
		os.Exit(1)
	}
	// Must match schedulesBaseURL and SCHEDULES_PORT (root compose maps ${SCHEDULES_PORT:-8080}:8080).
	if err := waitForTCP("127.0.0.1:18080", serviceReadyTimeout); err != nil {
		fmt.Fprintf(os.Stderr, "schedules port not ready: %v\n", err)
		_, _ = runCompose(2*time.Minute, "down", "-v", "--remove-orphans")
		os.Exit(1)
	}
	if err := waitForHTTP(serviceReadyTimeout, developersBaseURL+"/register", http.MethodPost, map[string]string{}); err != nil {
		fmt.Fprintf(os.Stderr, "developers http not ready: %v\n", err)
		_, _ = runCompose(2*time.Minute, "down", "-v", "--remove-orphans")
		os.Exit(1)
	}
	if err := waitForHTTP(serviceReadyTimeout, authBaseURL+"/auth", http.MethodGet, nil); err != nil {
		fmt.Fprintf(os.Stderr, "auth http not ready: %v\n", err)
		_, _ = runCompose(2*time.Minute, "down", "-v", "--remove-orphans")
		os.Exit(1)
	}
	if err := waitForHTTP(serviceReadyTimeout, schedulesBaseURL+"/institution", http.MethodGet, nil); err != nil {
		fmt.Fprintf(os.Stderr, "schedules http not ready: %v\n", err)
		_, _ = runCompose(2*time.Minute, "down", "-v", "--remove-orphans")
		os.Exit(1)
	}

	code := m.Run()

	if _, err := runCompose(2*time.Minute, "down", "-v", "--remove-orphans"); err != nil {
		fmt.Fprintf(os.Stderr, "compose down failed: %v\n", err)
	}
	os.Exit(code)
}

func TestUserKeyHasReadOnlyAccess(t *testing.T) {
	email := uniqueEmail("user")
	password := "SuperSecretPass123!"

	registerUser(t, email, password)

	sessionToken := loginUser(t, email, password)
	developerID := fetchDeveloperID(t, email)
	apiKey := createAPIKey(t, sessionToken, developerID, []map[string]any{
		{"Action": "write"},
	})

	assertAuthorized(t, apiKey, "read", ptr(int64(1)), http.StatusOK, true)
	assertAuthorized(t, apiKey, "write", ptr(int64(1)), http.StatusForbidden, false)
}

func TestAdminKeyHasWriteAccess(t *testing.T) {
	email := uniqueEmail("admin")
	password := "SuperSecretPass123!"

	registerUser(t, email, password)
	setDeveloperRole(t, email, 1)

	sessionToken := loginUser(t, email, password)
	developerID := fetchDeveloperID(t, email)
	apiKey := createAPIKey(t, sessionToken, developerID, []map[string]any{
		{"Action": "read"},
	})

	assertAuthorized(t, apiKey, "read", ptr(int64(1)), http.StatusOK, true)
	assertAuthorized(t, apiKey, "write", ptr(int64(999)), http.StatusOK, true)
}

func TestDuplicateRegistrationRejected(t *testing.T) {
	email := uniqueEmail("duplicate")
	password := "SuperSecretPass123!"

	registerUser(t, email, password)

	body := map[string]string{
		"email":    email,
		"password": password,
	}
	var resp map[string]string
	status := doJSON(t, http.MethodPost, developersBaseURL+"/register", body, nil, &resp)
	if status != http.StatusConflict {
		t.Fatalf("duplicate register status = %d, want %d", status, http.StatusConflict)
	}
	if resp["error"] != "email already exists" {
		t.Fatalf("duplicate register error = %q, want %q", resp["error"], "email already exists")
	}
}

func TestCreateKeyPortalIgnoresForeignDevUUIDUsesSession(t *testing.T) {
	password := "SuperSecretPass123!"

	_, sessionToken, _ := registerConfirmAndLogin(t, "owner", password)
	_, _, otherDeveloperID := registerConfirmAndLogin(t, "other", password)

	// BrainzDevelopers binds ApiKeyCreateSessionRequest only; dev_uuid/permissions in body are ignored.
	// The key is always created for the session developer.
	body := map[string]any{
		"api_key_name": "foreign-key",
		"dev_uuid":     otherDeveloperID,
		"permissions":  []map[string]any{{"Action": "read"}},
		"ip_whitelist": []string{},
	}
	headers := map[string]string{
		"X-Session-Token": sessionToken,
	}

	var resp apiKeyCreateResponse
	status := doJSON(t, http.MethodPost, developersBaseURL+"/key", body, headers, &resp)
	if status != http.StatusCreated {
		t.Fatalf("create key status = %d, want %d", status, http.StatusCreated)
	}
	if resp.APIKey == "" {
		t.Fatal("expected api key in response")
	}
	// Confirm key is under owner, not otherDeveloperID (would appear in owner's list only).
	var keysResp struct {
		Keys []struct {
			ID string `json:"id"`
		} `json:"keys"`
	}
	st := doJSON(t, http.MethodGet, developersBaseURL+"/keys", nil, headers, &keysResp)
	if st != http.StatusOK {
		t.Fatalf("list keys status = %d", st)
	}
	if len(keysResp.Keys) != 1 {
		t.Fatalf("owner keys len = %d, want 1", len(keysResp.Keys))
	}
}

func TestUnsupportedRoleCannotCreateKey(t *testing.T) {
	password := "SuperSecretPass123!"

	email, sessionToken, developerID := registerConfirmAndLogin(t, "unsupported_role", password)
	ensureRoleExists(t, 99, "unsupported")
	setDeveloperRole(t, email, 99)

	body := map[string]any{
		"api_key_name": "unsupported-role-key",
		"dev_uuid":     developerID,
		"permissions":  []map[string]any{{"Action": "read"}},
		"ip_whitelist": []string{},
	}
	headers := map[string]string{
		"X-Session-Token": sessionToken,
	}

	var resp map[string]string
	status := doJSON(t, http.MethodPost, developersBaseURL+"/key", body, headers, &resp)
	if status != http.StatusForbidden {
		t.Fatalf("create key with unsupported role status = %d, want %d", status, http.StatusForbidden)
	}
	if resp["error"] != "unsupported role" {
		t.Fatalf("create key with unsupported role error = %q, want %q", resp["error"], "unsupported role")
	}
}

func TestAuthRejectsInvalidAPIKey(t *testing.T) {
	var resp authResult
	status := doJSON(t, http.MethodGet, authBaseURL+"/auth", map[string]any{
		"perm": map[string]any{
			"Action":        "read",
			"InstitutionID": 1,
		},
	}, map[string]string{
		"X-API-Key":         "brainz_invalid:not-a-real-secret",
		"X-Original-IP":     "127.0.0.1",
		"X-Original-Path":   "/e2e-check",
		"X-Original-Method": http.MethodGet,
	}, &resp)
	if status != http.StatusUnauthorized {
		t.Fatalf("auth with invalid api key status = %d, want %d", status, http.StatusUnauthorized)
	}
	if resp.Status {
		t.Fatal("auth with invalid api key unexpectedly succeeded")
	}
	if resp.ErrorMessage != "invalid apiKey" {
		t.Fatalf("auth with invalid api key error = %q, want %q", resp.ErrorMessage, "invalid apiKey")
	}
}

func TestCreateKeyRejectsInvalidSessionToken(t *testing.T) {
	email := uniqueEmail("invalid_session")
	password := "SuperSecretPass123!"

	registerUser(t, email, password)
	developerID := fetchDeveloperID(t, email)

	body := map[string]any{
		"api_key_name": "invalid-session-key",
		"dev_uuid":     developerID,
		"permissions":  []map[string]any{{"Action": "read"}},
		"ip_whitelist": []string{},
	}
	headers := map[string]string{
		"X-Session-Token": "not-a-real-session-token",
	}

	var resp map[string]string
	status := doJSON(t, http.MethodPost, developersBaseURL+"/key", body, headers, &resp)
	if status != http.StatusBadRequest {
		t.Fatalf("create key with invalid session status = %d, want %d", status, http.StatusBadRequest)
	}
	if resp["error"] != "invalid session token" {
		t.Fatalf("create key with invalid session error = %q, want %q", resp["error"], "invalid session token")
	}
}

func TestDevelopersGetKeysWithoutSessionReturns400(t *testing.T) {
	status := doJSON(t, http.MethodGet, developersBaseURL+"/keys", nil, nil, nil)
	if status != http.StatusBadRequest {
		t.Fatalf("get keys without session status = %d, want %d", status, http.StatusBadRequest)
	}
}

func TestDevelopersGetKeysUsageWithoutSessionReturns401(t *testing.T) {
	status := doJSON(t, http.MethodGet, developersBaseURL+"/keys/usage", nil, nil, nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("get keys usage without session status = %d, want %d", status, http.StatusUnauthorized)
	}
}

func TestDevelopersKeysUsageRejectsDeveloperIDQueryParam(t *testing.T) {
	password := "SuperSecretPass123!"
	_, sessionToken, otherID := registerConfirmAndLogin(t, "usage_qs", password)

	u := developersBaseURL + "/keys/usage?developer_id=" + urlQueryEscape(otherID)
	var resp map[string]string
	status := doJSON(t, http.MethodGet, u, nil, map[string]string{"X-Session-Token": sessionToken}, &resp)
	if status != http.StatusBadRequest {
		t.Fatalf("keys usage with developer_id query status = %d, want %d", status, http.StatusBadRequest)
	}
	if resp["error"] == "" {
		t.Fatalf("expected error body, got %+v", resp)
	}
}

func TestDevelopersKeyUsageStatsNoKeys(t *testing.T) {
	password := "SuperSecretPass123!"
	email := uniqueEmail("usage_no_keys")
	registerUser(t, email, password)
	sessionToken := loginUser(t, email, password)

	var out apiKeysUsageResponse
	status := doJSON(t, http.MethodGet, developersBaseURL+"/keys/usage", nil, map[string]string{
		"X-Session-Token": sessionToken,
	}, &out)
	if status != http.StatusOK {
		t.Fatalf("keys usage status = %d, want %d", status, http.StatusOK)
	}
	if len(out.Keys) != 0 {
		t.Fatalf("keys usage with no api keys: len(keys) = %d, want 0", len(out.Keys))
	}
}

func TestDevelopersKeyUsageStatsEmptyUsageRows(t *testing.T) {
	password := "SuperSecretPass123!"
	_, sessionToken, developerID := registerConfirmAndLogin(t, "usage_empty", password)
	_ = createAPIKey(t, sessionToken, developerID, []map[string]any{{"Action": "read"}})

	var out apiKeysUsageResponse
	status := doJSON(t, http.MethodGet, developersBaseURL+"/keys/usage", nil, map[string]string{
		"X-Session-Token": sessionToken,
	}, &out)
	if status != http.StatusOK {
		t.Fatalf("keys usage status = %d, want %d", status, http.StatusOK)
	}
	if len(out.Keys) != 1 {
		t.Fatalf("len(keys) = %d, want 1", len(out.Keys))
	}
	k := out.Keys[0]
	if k.TotalUsage != 0 {
		t.Fatalf("total_usage = %d, want 0", k.TotalUsage)
	}
	if len(k.UsageByDay) != 0 {
		t.Fatalf("len(usage_by_day) = %d, want 0", len(k.UsageByDay))
	}
	if k.LastUsedAt != nil {
		t.Fatalf("last_used_at = %q, want omitted when no usage rows", *k.LastUsedAt)
	}
}

func TestDevelopersKeyUsageStatsWithMockedUsageRows(t *testing.T) {
	password := "SuperSecretPass123!"
	_, sessionToken, developerID := registerConfirmAndLogin(t, "usage_mock", password)
	_ = createAPIKey(t, sessionToken, developerID, []map[string]any{{"Action": "read"}})

	apiKeyID := fetchApiKeyID(t, developerID)
	if apiKeyID == "" {
		t.Fatal("fetchApiKeyID returned empty")
	}

	// Two events on the same UTC day, one on another day (within 90d window).
	insertMockAPIKeyUsage(t, apiKeyID)

	var out apiKeysUsageResponse
	status := doJSON(t, http.MethodGet, developersBaseURL+"/keys/usage", nil, map[string]string{
		"X-Session-Token": sessionToken,
	}, &out)
	if status != http.StatusOK {
		t.Fatalf("keys usage status = %d, want %d", status, http.StatusOK)
	}
	if len(out.Keys) != 1 {
		t.Fatalf("len(keys) = %d, want 1", len(out.Keys))
	}
	k := out.Keys[0]
	if k.ApiKeyID != apiKeyID {
		t.Fatalf("api_key_id = %q, want %q", k.ApiKeyID, apiKeyID)
	}
	if k.TotalUsage != 3 {
		t.Fatalf("total_usage = %d, want 3", k.TotalUsage)
	}
	if k.LastUsedAt == nil || *k.LastUsedAt == "" {
		t.Fatal("expected last_used_at after mocked usage")
	}
	if len(k.UsageByDay) != 2 {
		t.Fatalf("len(usage_by_day) = %d, want 2 (two distinct UTC days)", len(k.UsageByDay))
	}
	var daySum uint32
	for _, d := range k.UsageByDay {
		daySum += d.Count
	}
	if daySum != 3 {
		t.Fatalf("sum(usage_by_day.count) = %d, want 3", daySum)
	}

	// BrainzAuth direct endpoint should match (same aggregates).
	var authOut apiKeysUsageResponse
	authStatus := doJSON(t, http.MethodGet, authBaseURL+"/keys/usage?developer_id="+urlQueryEscape(developerID), nil, nil, &authOut)
	if authStatus != http.StatusOK {
		t.Fatalf("auth /keys/usage status = %d, want %d", authStatus, http.StatusOK)
	}
	if len(authOut.Keys) != 1 || authOut.Keys[0].TotalUsage != 3 {
		t.Fatalf("auth usage mismatch: %+v", authOut.Keys)
	}
}

func TestDevelopersDeleteKeyForOtherDeveloperReturns403(t *testing.T) {
	password := "SuperSecretPass123!"

	_, ownerSession, ownerDevID := registerConfirmAndLogin(t, "del_owner", password)
	_, attackerSession, attackerDevID := registerConfirmAndLogin(t, "del_attacker", password)

	// owner creates a key
	ownerKey := createAPIKey(t, ownerSession, ownerDevID, []map[string]any{{"Action": "read"}})

	// attacker tries to delete owner's key
	status, body := doRaw(
		t,
		http.MethodDelete,
		developersBaseURL+"/key?api_key="+urlQueryEscape(ownerKey),
		nil,
		map[string]string{"X-Session-Token": attackerSession},
	)
	if status != http.StatusForbidden {
		t.Fatalf(
			"delete other developer key status=%d, want %d (attacker=%s) body=%s",
			status, http.StatusForbidden, attackerDevID, string(body),
		)
	}
}

func TestDevelopersDeleteOwnKeyReturns200AndRemovesKey(t *testing.T) {
	password := "SuperSecretPass123!"
	_, sessionToken, developerID := registerConfirmAndLogin(t, "del_own_key", password)
	apiKey := createAPIKey(t, sessionToken, developerID, []map[string]any{{"Action": "read"}})

	var before struct {
		Keys []struct {
			ID string `json:"id"`
		} `json:"keys"`
	}
	st := doJSON(t, http.MethodGet, developersBaseURL+"/keys", nil, map[string]string{"X-Session-Token": sessionToken}, &before)
	if st != http.StatusOK {
		t.Fatalf("list keys before delete status = %d, want %d", st, http.StatusOK)
	}
	if len(before.Keys) != 1 {
		t.Fatalf("list keys before delete: len = %d, want 1", len(before.Keys))
	}

	var delResp map[string]string
	delURL := developersBaseURL + "/key?api_key=" + urlQueryEscape(apiKey)
	delSt := doJSON(t, http.MethodDelete, delURL, nil, map[string]string{"X-Session-Token": sessionToken}, &delResp)
	if delSt != http.StatusOK {
		t.Fatalf("delete own key status = %d, want %d", delSt, http.StatusOK)
	}
	if delResp["status"] != "success" {
		t.Fatalf("delete response = %+v, want status=success", delResp)
	}

	var after struct {
		Keys []struct {
			ID string `json:"id"`
		} `json:"keys"`
	}
	st = doJSON(t, http.MethodGet, developersBaseURL+"/keys", nil, map[string]string{"X-Session-Token": sessionToken}, &after)
	if st != http.StatusOK {
		t.Fatalf("list keys after delete status = %d, want %d", st, http.StatusOK)
	}
	if len(after.Keys) != 0 {
		t.Fatalf("list keys after delete: len = %d, want 0", len(after.Keys))
	}

	var authResp authResult
	authSt := doJSON(t, http.MethodGet, authBaseURL+"/auth", map[string]any{
		"perm": map[string]any{
			"Action":        "read",
			"InstitutionID": 1,
		},
	}, map[string]string{
		"X-API-Key":         apiKey,
		"X-Original-IP":     "127.0.0.1",
		"X-Original-Path":   "/e2e-after-delete",
		"X-Original-Method": http.MethodGet,
	}, &authResp)
	if authSt != http.StatusUnauthorized {
		t.Fatalf("auth with deleted key status = %d, want %d", authSt, http.StatusUnauthorized)
	}
	if authResp.Status {
		t.Fatal("auth with deleted key unexpectedly succeeded")
	}
}

func TestSchedulesCreateInstitutionWithoutAdminReturns403(t *testing.T) {
	// read-only key should not be able to create institution (admin permission required).
	readKey := createPortalAPIKeyForRole(t, "inst_readonly", 0)

	body := map[string]any{
		"name":      "E2E Forbidden Institution",
		"site_link": "https://example.com",
	}
	status := doJSON(t, http.MethodPost, schedulesBaseURL+"/institution", body, map[string]string{
		"X-Api-Key": readKey,
	}, nil)
	if status != http.StatusForbidden {
		t.Fatalf("create institution with read key status = %d, want %d", status, http.StatusForbidden)
	}
}

func TestRegisterRejectsInvalidBody(t *testing.T) {
	body := map[string]string{
		"email":    "not-an-email",
		"password": "short",
	}

	status := doJSON(t, http.MethodPost, developersBaseURL+"/register", body, nil, nil)
	if status != http.StatusBadRequest {
		t.Fatalf("register invalid body status = %d, want %d", status, http.StatusBadRequest)
	}
}

func TestLoginRejectsInvalidBody(t *testing.T) {
	body := map[string]string{
		"email":    "broken-email",
		"password": "tiny",
	}

	status := doJSON(t, http.MethodPost, developersBaseURL+"/login", body, nil, nil)
	if status != http.StatusBadRequest {
		t.Fatalf("login invalid body status = %d, want %d", status, http.StatusBadRequest)
	}
}

func TestAuthRejectsMissingPermissionBody(t *testing.T) {
	password := "SuperSecretPass123!"
	_, sessionToken, developerID := registerConfirmAndLogin(t, "missing_perm", password)
	apiKey := createAPIKey(t, sessionToken, developerID, []map[string]any{
		{"Action": "read"},
	})

	status := doJSON(t, http.MethodGet, authBaseURL+"/auth", nil, map[string]string{
		"X-API-Key":         apiKey,
		"X-Original-IP":     "127.0.0.1",
		"X-Original-Path":   "/e2e-check",
		"X-Original-Method": http.MethodGet,
	}, nil)
	if status != http.StatusBadRequest {
		t.Fatalf("auth missing permission body status = %d, want %d", status, http.StatusBadRequest)
	}
}

func TestAuthRejectsMissingAPIKey(t *testing.T) {
	var resp authResult
	status := doJSON(t, http.MethodGet, authBaseURL+"/auth", map[string]any{
		"perm": map[string]any{
			"Action":        "read",
			"InstitutionID": 1,
		},
	}, map[string]string{
		"X-Original-IP":     "127.0.0.1",
		"X-Original-Path":   "/e2e-check",
		"X-Original-Method": http.MethodGet,
	}, &resp)
	if status != http.StatusUnauthorized {
		t.Fatalf("auth missing api key status = %d, want %d", status, http.StatusUnauthorized)
	}
	if resp.Status {
		t.Fatal("auth without api key unexpectedly succeeded")
	}
}

func TestSchedulesGetInstitutionsRequiresKey(t *testing.T) {
	readKey := createPortalAPIKeyForRole(t, "inst_list", 0)
	var resp []institutionResponse
	status := doJSON(t, http.MethodGet, schedulesBaseURL+"/institution", nil, map[string]string{
		"X-Api-Key": readKey,
	}, &resp)
	if status != http.StatusOK {
		t.Fatalf("get institutions status = %d, want %d", status, http.StatusOK)
	}
	if len(resp) == 0 {
		t.Fatal("expected seeded institutions, got empty list")
	}
	if resp[0].Name == "" {
		t.Fatalf("unexpected institution payload: %+v", resp[0])
	}
}

func TestSchedulesReadAndWriteAccessFlow(t *testing.T) {
	groupID := ensureScheduleGroup(t, "E2E-PI-101", 1)

	adminKey := createPortalAPIKeyForRole(t, "schedule_admin", 1)
	targetDate := "2026-03-15"
	body := map[string]any{
		"institutionID": 1,
		"lessons": []map[string]any{
			{
				"name":         "E2E Algebra",
				"cab_num":      "404",
				"teacher_name": "E2E Teacher",
				"start_time":   "2026-03-15T09:00:00Z",
				"end_time":     "2026-03-15T10:30:00Z",
				"num":          1,
				"group_id":     groupID,
			},
		},
	}

	status := doJSON(t, http.MethodPost, schedulesBaseURL+"/lessons", body, map[string]string{
		"X-Api-Key": adminKey,
	}, nil)
	if status != http.StatusCreated {
		t.Fatalf("admin create lesson status = %d, want %d", status, http.StatusCreated)
	}

	readKey := createPortalAPIKeyForRole(t, "schedule_user", 0)

	var groups []groupResponse
	status = doJSON(t, http.MethodGet, schedulesBaseURL+"/group?institution_id=1", nil, map[string]string{
		"X-Api-Key": readKey,
	}, &groups)
	if status != http.StatusOK {
		t.Fatalf("get groups status = %d, want %d", status, http.StatusOK)
	}
	if len(groups) == 0 {
		t.Fatal("expected at least one group")
	}

	var lessons []lessonResponse
	status = doJSON(t, http.MethodGet, schedulesBaseURL+"/lessons?institution_id=1&date="+targetDate, nil, map[string]string{
		"X-Api-Key": readKey,
	}, &lessons)
	if status != http.StatusOK {
		t.Fatalf("get lessons status = %d, want %d", status, http.StatusOK)
	}
	if len(lessons) == 0 {
		t.Fatal("expected at least one lesson")
	}

	status = doJSON(t, http.MethodPost, schedulesBaseURL+"/lessons", body, map[string]string{
		"X-Api-Key": readKey,
	}, nil)
	if status != http.StatusForbidden {
		t.Fatalf("read-only create lesson status = %d, want %d", status, http.StatusForbidden)
	}
}

func TestSchedulesGetGroupsWithoutKeyReturns403(t *testing.T) {
	status := doJSON(t, http.MethodGet, schedulesBaseURL+"/group?institution_id=1", nil, nil, nil)
	if status != http.StatusForbidden {
		t.Fatalf("get groups without api key status = %d, want %d", status, http.StatusForbidden)
	}
}

func TestSchedulesGetLessonsWithoutKeyReturns403(t *testing.T) {
	status := doJSON(t, http.MethodGet, schedulesBaseURL+"/lessons?institution_id=1&date=2026-03-15", nil, nil, nil)
	if status != http.StatusForbidden {
		t.Fatalf("get lessons without api key status = %d, want %d", status, http.StatusForbidden)
	}
}

func TestSchedulesInfoStreamWebSocketUpgrade(t *testing.T) {
	u := strings.Replace(schedulesBaseURL, "http://", "ws://", 1) + "/info-stream?institution_id=1"

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, resp, err := dialer.Dial(u, nil)
	if err != nil {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		t.Fatalf("websocket dial failed (status=%d): %v", status, err)
	}
	defer conn.Close()

	_ = conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	if err := conn.WriteMessage(websocket.TextMessage, []byte("e2e")); err != nil {
		t.Fatalf("websocket write failed: %v", err)
	}
}

func registerUser(t *testing.T, email, password string) {
	t.Helper()

	body := map[string]string{
		"email":    email,
		"password": password,
	}

	var resp registerResponse
	status := doJSON(t, http.MethodPost, developersBaseURL+"/register", body, nil, &resp)
	if status != http.StatusCreated {
		t.Fatalf("register status = %d, want %d", status, http.StatusCreated)
	}
}

func loginUser(t *testing.T, email, password string) string {
	t.Helper()

	body := map[string]string{
		"email":    email,
		"password": password,
	}

	var resp loginResponse
	status := doJSON(t, http.MethodPost, developersBaseURL+"/login", body, nil, &resp)
	if status != http.StatusOK {
		t.Fatalf("login status = %d, want %d", status, http.StatusOK)
	}
	if resp.Token == "" {
		t.Fatal("login response did not include token")
	}
	return resp.Token
}

func registerConfirmAndLogin(t *testing.T, prefix, password string) (string, string, string) {
	t.Helper()

	email := uniqueEmail(prefix)
	registerUser(t, email, password)
	sessionToken := loginUser(t, email, password)
	developerID := fetchDeveloperID(t, email)

	return email, sessionToken, developerID
}

func createAPIKey(t *testing.T, sessionToken, developerID string, requestedPerms []map[string]any) string {
	t.Helper()

	body := map[string]any{
		"api_key_name": "e2e-key",
		"dev_uuid":     developerID,
		"permissions":  requestedPerms,
		"ip_whitelist": []string{},
	}
	headers := map[string]string{
		"X-Session-Token": sessionToken,
	}

	var resp apiKeyCreateResponse
	status := doJSON(t, http.MethodPost, developersBaseURL+"/key", body, headers, &resp)
	if status != http.StatusCreated {
		t.Fatalf("create key status = %d, want %d", status, http.StatusCreated)
	}
	if resp.APIKey == "" {
		t.Fatal("create key response did not include api key")
	}
	return resp.APIKey
}

func createPortalAPIKeyForRole(t *testing.T, prefix string, roleID int) string {
	t.Helper()

	email, sessionToken, developerID := registerConfirmAndLogin(t, prefix, "SuperSecretPass123!")
	setDeveloperRole(t, email, roleID)

	return createAPIKey(t, sessionToken, developerID, []map[string]any{
		{"Action": "read"},
		{"Action": "write"},
	})
}

func assertAuthorized(t *testing.T, apiKey, action string, institutionID *int64, wantStatus int, wantAuthorized bool) {
	t.Helper()

	perm := map[string]any{
		"Action": action,
	}
	if institutionID != nil {
		perm["InstitutionID"] = *institutionID
	}

	body := map[string]any{
		"perm": perm,
	}
	headers := map[string]string{
		"X-API-Key":         apiKey,
		"X-Original-IP":     "127.0.0.1",
		"X-Original-Path":   "/e2e-check",
		"X-Original-Method": http.MethodGet,
	}

	var resp authResult
	status := doJSON(t, http.MethodGet, authBaseURL+"/auth", body, headers, &resp)
	if status != wantStatus {
		t.Fatalf("auth(%s) status = %d, want %d, body = %+v", action, status, wantStatus, resp)
	}
	if resp.Status != wantAuthorized {
		t.Fatalf("auth(%s) authorized = %v, want %v", action, resp.Status, wantAuthorized)
	}
}

func fetchDeveloperID(t *testing.T, email string) string {
	t.Helper()
	query := fmt.Sprintf(
		"SELECT id FROM developer_accounts WHERE email = '%s' ORDER BY created_at DESC LIMIT 1;",
		email,
	)
	return strings.TrimSpace(execPSQLQuery(t, "brainz_developers", query))
}

func fetchApiKeyID(t *testing.T, developerID string) string {
	t.Helper()
	query := fmt.Sprintf(
		"SELECT id FROM api_key WHERE developer_id = '%s' ORDER BY created_at DESC LIMIT 1;",
		developerID,
	)
	return strings.TrimSpace(execPSQLQuery(t, "brainz_auth", query))
}

// insertMockAPIKeyUsage inserts three rows: two share the same UTC calendar day (both NOW()-1d), one older day.
func insertMockAPIKeyUsage(t *testing.T, apiKeyID string) {
	t.Helper()
	// Single line: leading newline breaks `psql -c` when passed through `sh -lc`.
	query := fmt.Sprintf(
		"INSERT INTO api_key_usage (id, api_key_id, endpoint, method, usage_at, response_code) VALUES "+
			"(uuid_generate_v4(), '%s'::uuid, '/e2e-a', 'GET', NOW() - interval '1 day', '200'), "+
			"(uuid_generate_v4(), '%s'::uuid, '/e2e-b', 'GET', NOW() - interval '1 day', '200'), "+
			"(uuid_generate_v4(), '%s'::uuid, '/e2e-c', 'GET', NOW() - interval '4 days', '200');",
		apiKeyID, apiKeyID, apiKeyID,
	)
	out := execPSQLQuery(t, "brainz_auth", query)
	if !strings.Contains(out, "INSERT 0 3") {
		t.Fatalf("insert api_key_usage: output = %q, want INSERT 0 3", strings.TrimSpace(out))
	}
}

func setDeveloperRole(t *testing.T, email string, roleID int) {
	t.Helper()
	query := fmt.Sprintf(
		"UPDATE developer_accounts SET role_id = %d WHERE email = '%s';",
		roleID,
		email,
	)
	if out := execPSQLQuery(t, "brainz_developers", query); strings.TrimSpace(out) != "UPDATE 1" {
		t.Fatalf("set role query output = %q, want %q", strings.TrimSpace(out), "UPDATE 1")
	}
}

func ensureRoleExists(t *testing.T, roleID int, title string) {
	t.Helper()
	query := fmt.Sprintf(
		"INSERT INTO role (id, title) VALUES (%d, '%s') ON CONFLICT (id) DO NOTHING;",
		roleID,
		title,
	)
	out := strings.TrimSpace(execPSQLQuery(t, "brainz_developers", query))
	if out != "INSERT 0 1" && out != "INSERT 0 0" {
		t.Fatalf("ensure role query output = %q, want INSERT 0 1 or INSERT 0 0", out)
	}
}

func ensureScheduleGroup(t *testing.T, name string, institutionID int64) int64 {
	t.Helper()
	query := fmt.Sprintf(
		"INSERT INTO groups (name, institution_id) VALUES ('%s', %d) RETURNING id;",
		name,
		institutionID,
	)
	out := strings.TrimSpace(execPSQLQuery(t, "brainz_lessons", query))
	if idx := strings.Index(out, "\n"); idx >= 0 {
		out = out[:idx]
	}
	out = strings.TrimSpace(out)
	id, err := strconv.ParseInt(out, 10, 64)
	if err != nil {
		t.Fatalf("parse group id %q: %v", out, err)
	}
	return id
}

func execPSQLQuery(t *testing.T, dbName, query string) string {
	t.Helper()

	script := fmt.Sprintf(`psql -v ON_ERROR_STOP=1 -U "$DB_USER" -d %s -At -c %s`, dbName, strconv.Quote(query))
	out, err := runCompose(2*time.Minute, "exec", "-T", "db", "sh", "-lc", script)
	if err != nil {
		t.Fatalf("psql query failed: %v", err)
	}
	return out
}

func doJSON(t *testing.T, method, url string, body any, headers map[string]string, out any) int {
	t.Helper()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("http request failed: %v", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	if out != nil && len(bytes.TrimSpace(data)) > 0 && (resp.StatusCode >= 200 && resp.StatusCode < 300 || resp.StatusCode >= 400 && resp.StatusCode < 500) {
		if err := json.Unmarshal(data, out); err != nil {
			t.Fatalf("unmarshal response body %q: %v", string(data), err)
		}
	}

	return resp.StatusCode
}

func doRaw(t *testing.T, method, url string, body any, headers map[string]string) (int, []byte) {
	t.Helper()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("http request failed: %v", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return resp.StatusCode, data
}

func waitForTCP(address string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", address, 2*time.Second)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("timed out waiting for %s", address)
}

func waitForHTTP(timeout time.Duration, url, method string, body any) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		status, err := probeHTTP(method, url, body)
		if err == nil && status > 0 && status < 500 {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("timed out waiting for %s %s", method, url)
}

func probeHTTP(method, url string, body any) (int, error) {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return 0, err
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func runCompose(timeout time.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	composeArgs := []string{
		"compose",
		"-p", composeProject,
		"-f", filepath.Join(repoRoot, "docker-compose.yml"),
		"-f", filepath.Join(repoRoot, "tests", "e2e", "docker-compose.e2e.yml"),
	}
	composeArgs = append(composeArgs, args...)

	cmd := exec.CommandContext(ctx, "docker", composeArgs...)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "DB_NAME=postgres")
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return string(output), fmt.Errorf("docker compose timed out after %s: %s", timeout, string(output))
	}
	if err != nil {
		return string(output), fmt.Errorf("docker compose %v failed: %w\n%s", args, err, string(output))
	}
	return string(output), nil
}

func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s_%d@example.com", prefix, time.Now().UnixNano())
}

func ptr[T any](v T) *T {
	return &v
}

func findRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "docker-compose.yml")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("docker-compose.yml not found from %s", wd)
		}
		dir = parent
	}
}
