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
	if respHealth.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK for /health, got %d", respHealth.StatusCode)
	}

	// 2. QR page should return 403 Forbidden because bot is already logged in (Auto-Lock)
	reqQR := httptest.NewRequest("GET", "/qr", nil)
	respQR, err := server.app.Test(reqQR)
	if err != nil {
		t.Fatalf("qr test error: %v", err)
	}
	if respQR.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for /qr when logged in, got %d", respQR.StatusCode)
	}

	// 3. API route without key should return 401 Unauthorized
	reqAPI := httptest.NewRequest("POST", "/api/cache/refresh", nil)
	respAPI, err := server.app.Test(reqAPI)
	if err != nil {
		t.Fatalf("api test error: %v", err)
	}
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
	if respAPIQR.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for /api/v1/qr when logged in, got %d", respAPIQR.StatusCode)
	}
}
