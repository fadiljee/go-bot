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

	// Helper for testing GET endpoints with auth
	testEndpoint := func(name, path string) {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest("GET", path, nil)
			req.Header.Set("X-API-Key", "secret_bps_route_key")
			resp, err := server.app.Test(req)
			if err != nil {
				t.Fatalf("%s error: %v", path, err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected 200 OK for %s, got %d", path, resp.StatusCode)
			}
		})
	}

	// 2. Test all 9 GET endpoints in BPS route group
	testEndpoint("generic list", "/api/v1/bps/list?model=subject&page=1")
	testEndpoint("subjects", "/api/v1/bps/subjects?page=1")
	testEndpoint("subcats", "/api/v1/bps/subcats")
	testEndpoint("variables", "/api/v1/bps/variables?subject=1")
	testEndpoint("statictables", "/api/v1/bps/statictables?keyword=penduduk")
	testEndpoint("publications", "/api/v1/bps/publications?year=2024")
	testEndpoint("pressreleases", "/api/v1/bps/pressreleases")
	testEndpoint("infographics", "/api/v1/bps/infographics")
	testEndpoint("data", "/api/v1/bps/data?var=123")

	// 3. Test POST cache clear endpoint
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

