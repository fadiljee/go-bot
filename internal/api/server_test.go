package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bot-istimewa/internal/config"
	"bot-istimewa/internal/whatsapp"
)

func TestAPIServerSecurity(t *testing.T) {
	cfg := &config.Config{
		AppPort:      "3000",
		APISecretKey: "test_secret_key_12345",
	}

	clientMgr := &whatsapp.ClientManager{
		IsLoggedIn: true, // Simulate already logged in
	}

	server := NewAPIServer(cfg, clientMgr, nil)

	// 1. Health check should be public
	reqHealth := httptest.NewRequest("GET", "/health", nil)
	respHealth, err := server.app.Test(reqHealth)
	if err != nil {
		t.Fatalf("health test error: %v", err)
	}
	defer func() { _ = respHealth.Body.Close() }()
	if respHealth.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK for /health, got %d", respHealth.StatusCode)
	}

	// 2. QR page should return 403 Forbidden because bot is already logged in (Auto-Lock)
	reqQR := httptest.NewRequest("GET", "/qr", nil)
	respQR, err := server.app.Test(reqQR)
	if err != nil {
		t.Fatalf("qr test error: %v", err)
	}
	defer func() { _ = respQR.Body.Close() }()
	if respQR.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for /qr when logged in, got %d", respQR.StatusCode)
	}

	// 3. API route without key should return 401 Unauthorized
	reqAPI := httptest.NewRequest("POST", "/api/cache/refresh", nil)
	respAPI, err := server.app.Test(reqAPI)
	if err != nil {
		t.Fatalf("api test error: %v", err)
	}
	defer func() { _ = respAPI.Body.Close() }()
	if respAPI.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for /api without key, got %d", respAPI.StatusCode)
	}

	// 4. API route with invalid key should return 401 Unauthorized
	reqBadKey := httptest.NewRequest("POST", "/api/cache/refresh", nil)
	reqBadKey.Header.Set("X-API-Key", "wrong_key")
	respBadKey, err := server.app.Test(reqBadKey)
	if err != nil {
		t.Fatalf("api test error: %v", err)
	}
	defer func() { _ = respBadKey.Body.Close() }()
	if respBadKey.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized with bad key, got %d", respBadKey.StatusCode)
	}

	// 5. API QR endpoint should return 403 Forbidden when logged in (even with valid key)
	reqAPIQR := httptest.NewRequest("GET", "/api/v1/qr", nil)
	reqAPIQR.Header.Set("X-API-Key", "test_secret_key_12345")
	respAPIQR, err := server.app.Test(reqAPIQR)
	if err != nil {
		t.Fatalf("api qr test error: %v", err)
	}
	defer func() { _ = respAPIQR.Body.Close() }()
	if respAPIQR.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for /api/v1/qr when logged in, got %d", respAPIQR.StatusCode)
	}

	// 6. Query parameter ?key= must be rejected to prevent secret leakage in logs
	reqQueryKey := httptest.NewRequest("POST", "/api/cache/refresh?key=test_secret_key_12345", nil)
	respQueryKey, err := server.app.Test(reqQueryKey)
	if err != nil {
		t.Fatalf("query key test error: %v", err)
	}
	defer func() { _ = respQueryKey.Body.Close() }()
	if respQueryKey.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized when passing API key via query string, got %d", respQueryKey.StatusCode)
	}

	// 7. Security headers must be present
	if got := respHealth.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("expected X-Content-Type-Options=nosniff, got %q", got)
	}
	if got := respHealth.Header.Get("X-Frame-Options"); got != "DENY" {
		t.Errorf("expected X-Frame-Options=DENY, got %q", got)
	}
}

func TestAPIQRProxyProtectionAndBlastValidation(t *testing.T) {
	cfg := &config.Config{
		AppPort:      "3000",
		APISecretKey: "test_secret_key_12345",
	}

	// Simulate logged out state
	clientMgr := &whatsapp.ClientManager{
		IsLoggedIn: false,
	}
	server := NewAPIServer(cfg, clientMgr, nil)

	// 1. Remote request via reverse proxy (X-Forwarded-For) to /qr without API key -> 401 Unauthorized
	reqRemoteQR := httptest.NewRequest("GET", "/qr", nil)
	reqRemoteQR.Header.Set("X-Forwarded-For", "203.0.113.10")
	respRemoteQR, err := server.app.Test(reqRemoteQR)
	if err != nil {
		t.Fatalf("remote qr error: %v", err)
	}
	defer func() { _ = respRemoteQR.Body.Close() }()
	if respRemoteQR.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for remote /qr without API key, got %d", respRemoteQR.StatusCode)
	}

	// 2. Local request to /qr when logged out -> 200 OK
	reqLocalQR := httptest.NewRequest("GET", "/qr", nil)
	respLocalQR, err := server.app.Test(reqLocalQR)
	if err != nil {
		t.Fatalf("local qr error: %v", err)
	}
	defer func() { _ = respLocalQR.Body.Close() }()
	if respLocalQR.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK for local /qr when logged out, got %d", respLocalQR.StatusCode)
	}

	// 3. Blast validation: reject invalid phone numbers
	if _, ok := sanitizeBlastPhone("abc-invalid"); ok {
		t.Error("expected sanitizeBlastPhone to reject non-digit phone")
	}
	if _, ok := sanitizeBlastPhone("12345"); ok {
		t.Error("expected sanitizeBlastPhone to reject too-short phone")
	}
	if clean, ok := sanitizeBlastPhone("0812-3456-7890"); !ok || clean != "6281234567890" {
		t.Errorf("expected 6281234567890, got %q (ok=%v)", clean, ok)
	}
}

func TestAPIServerScraperRoutes(t *testing.T) {
	cfg := &config.Config{
		AppPort:        "3000",
		APISecretKey:   "secret_scraper_key",
		ScraperEnabled: true,
	}

	clientMgr := &whatsapp.ClientManager{IsLoggedIn: true}
	server := NewAPIServer(cfg, clientMgr, nil)

	// 1. Unauthenticated request to /api/v1/scraper/status -> 401
	reqUnauth := httptest.NewRequest("GET", "/api/v1/scraper/status", nil)
	respUnauth, err := server.app.Test(reqUnauth)
	if err != nil {
		t.Fatalf("unauth error: %v", err)
	}
	defer func() { _ = respUnauth.Body.Close() }()
	if respUnauth.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for /api/v1/scraper/status, got %d", respUnauth.StatusCode)
	}

	// 2. Health check includes scraper fields
	reqHealth := httptest.NewRequest("GET", "/health", nil)
	respHealth, err := server.app.Test(reqHealth)
	if err != nil || respHealth.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK for /health, got %d", respHealth.StatusCode)
	}
	defer func() { _ = respHealth.Body.Close() }()
}
