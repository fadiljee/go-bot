package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"bot-istimewa/internal/config"
	"bot-istimewa/internal/scraper"
	"bot-istimewa/internal/whatsapp"
)

func TestAPIIndicators(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api_ind_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := scraper.NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	now := time.Now()
	// Seed a press release item
	brsItem := scraper.Item{
		Kind:        scraper.KindPressRelease,
		Lang:        "id",
		SourceKey:   "brs-inflasi-1",
		Title:       "September 2026, Inflasi Year on Year (y-on-y) Kabupaten Bangka Barat sebesar 4,49 persen.",
		URL:         "https://bangkabaratkab.bps.go.id/brs/1.html",
		FileURL:     "https://webapi.bps.go.id/download.php?f=pdf1",
		ReleaseDate: &now,
		ContentHash: "hash-brs-1",
	}
	_, _, _ = store.ApplyItems(ctx, scraper.KindPressRelease, "id", []scraper.Item{brsItem}, true)

	cfg := &config.Config{
		AppPort:              "3000",
		APISecretKey:         "secret_123",
		ScraperMaxStaleHours: 48,
	}

	clientMgr := &whatsapp.ClientManager{IsLoggedIn: true}
	server := NewAPIServer(cfg, clientMgr, nil)

	provider := scraper.NewProvider(store, 48)
	server.SetScraper(nil, store, nil, nil, provider)

	// 1. Without auth header -> 401 Unauthorized
	reqNoAuth := httptest.NewRequest("GET", "/api/v1/indicators", nil)
	respNoAuth, err := server.app.Test(reqNoAuth)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respNoAuth.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", respNoAuth.StatusCode)
	}

	// 2. With auth header -> 200 OK, returns 7 indicators
	reqAuth := httptest.NewRequest("GET", "/api/v1/indicators", nil)
	reqAuth.Header.Set("X-API-Key", "secret_123")
	respAuth, err := server.app.Test(reqAuth)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respAuth.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", respAuth.StatusCode)
	}

	var jsonBody struct {
		Success bool                         `json:"success"`
		Total   int                          `json:"total"`
		Data    []scraper.StrategicIndicator `json:"data"`
	}
	if err := json.NewDecoder(respAuth.Body).Decode(&jsonBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !jsonBody.Success || jsonBody.Total != 7 {
		t.Errorf("expected 7 indicators, got total=%d, success=%v", jsonBody.Total, jsonBody.Success)
	}

	// 3. GET /api/v1/indicators/inflasi -> 200 OK
	reqInflasi := httptest.NewRequest("GET", "/api/v1/indicators/inflasi", nil)
	reqInflasi.Header.Set("X-API-Key", "secret_123")
	respInflasi, err := server.app.Test(reqInflasi)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respInflasi.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for inflasi, got %d", respInflasi.StatusCode)
	}

	var singleBody struct {
		Success bool                       `json:"success"`
		Data    scraper.StrategicIndicator `json:"data"`
	}
	if err := json.NewDecoder(respInflasi.Body).Decode(&singleBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if singleBody.Data.ID != 6 || singleBody.Data.Slug != "inflasi" {
		t.Errorf("expected ID 6 and slug inflasi, got ID=%d slug=%s", singleBody.Data.ID, singleBody.Data.Slug)
	}

	// 4. GET /api/v1/indikator/1 (Indonesian alias) -> 200 OK
	reqAlias := httptest.NewRequest("GET", "/api/v1/indikator/1", nil)
	reqAlias.Header.Set("X-API-Key", "secret_123")
	respAlias, err := server.app.Test(reqAlias)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respAlias.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for alias /api/v1/indikator/1, got %d", respAlias.StatusCode)
	}

	// 5. GET /api/v1/indicators?format=text -> 200 OK text/plain
	reqText := httptest.NewRequest("GET", "/api/v1/indicators?format=text", nil)
	reqText.Header.Set("X-API-Key", "secret_123")
	respText, err := server.app.Test(reqText)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respText.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", respText.StatusCode)
	}

	// 6. Unknown indicator -> 404 Not Found
	reqNotFound := httptest.NewRequest("GET", "/api/v1/indicators/999", nil)
	reqNotFound.Header.Set("X-API-Key", "secret_123")
	respNotFound, err := server.app.Test(reqNotFound)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respNotFound.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 for unknown indicator, got %d", respNotFound.StatusCode)
	}
}
