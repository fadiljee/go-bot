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

func TestBPSRouteGroup(t *testing.T) {
	// Mock BPS upstream server
	mockBPS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"status": "OK",
			"data-availability": "available",
			"data": [
				{"page": 1, "pages": 1, "per_page": 10, "total": 1},
				[{"sub_id": 1, "title": "Sosial dan Kependudukan", "subcat_id": 1, "subcat": "Sosduk"}]
			]
		}`))
	}))
	defer mockBPS.Close()

	cfg := &config.Config{
		AppPort:          "3000",
		APISecretKey:     "secret_bps_route_key",
		BPSAPIBaseURL:    mockBPS.URL,
		BPSAPIKey:        "bps-test-key",
		BPSDefaultDomain: "1903",
	}

	clientMgr := &whatsapp.ClientManager{IsLoggedIn: true}
	server := NewAPIServer(cfg, clientMgr, nil)
	server.bps.SetHTTPClient(mockBPS.Client())

	// 1. Without auth -> 401
	reqUnauth := httptest.NewRequest("GET", "/api/v1/bps/subjects", nil)
	respUnauth, err := server.app.Test(reqUnauth)
	if err != nil {
		t.Fatalf("unauth test error: %v", err)
	}
	if respUnauth.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for /api/v1/bps/subjects, got %d", respUnauth.StatusCode)
	}

	// 2. With auth -> 200 OK
	reqAuth := httptest.NewRequest("GET", "/api/v1/bps/subjects?page=1", nil)
	reqAuth.Header.Set("X-API-Key", "secret_bps_route_key")
	respAuth, err := server.app.Test(reqAuth)
	if err != nil {
		t.Fatalf("auth test error: %v", err)
	}
	if respAuth.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK for /api/v1/bps/subjects, got %d", respAuth.StatusCode)
	}

	// 3. Test publications endpoint
	reqPub := httptest.NewRequest("GET", "/api/v1/bps/publications?year=2024", nil)
	reqPub.Header.Set("X-API-Key", "secret_bps_route_key")
	respPub, err := server.app.Test(reqPub)
	if err != nil {
		t.Fatalf("pub test error: %v", err)
	}
	if respPub.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK for /api/v1/bps/publications, got %d", respPub.StatusCode)
	}

	// 4. Test pressreleases endpoint
	reqBrs := httptest.NewRequest("GET", "/api/v1/bps/pressreleases", nil)
	reqBrs.Header.Set("X-API-Key", "secret_bps_route_key")
	respBrs, err := server.app.Test(reqBrs)
	if err != nil {
		t.Fatalf("brs test error: %v", err)
	}
	if respBrs.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK for /api/v1/bps/pressreleases, got %d", respBrs.StatusCode)
	}

	// 5. Test cache clear endpoint
	reqClear := httptest.NewRequest("POST", "/api/v1/bps/cache/clear", nil)
	reqClear.Header.Set("X-API-Key", "secret_bps_route_key")
	respClear, err := server.app.Test(reqClear)
	if err != nil {
		t.Fatalf("clear test error: %v", err)
	}
	if respClear.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK for /api/v1/bps/cache/clear, got %d", respClear.StatusCode)
	}
}

