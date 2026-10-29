package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"bot-istimewa/internal/config"
	"bot-istimewa/internal/scraper"
	"bot-istimewa/internal/whatsapp"
)

func TestAPIPakar(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api_pakar_test_*")
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

	cfg := &config.Config{
		AppPort:              "3000",
		APISecretKey:         "secret_123",
		ScraperMaxStaleHours: 48,
	}

	clientMgr := &whatsapp.ClientManager{IsLoggedIn: true}
	server := NewAPIServer(cfg, clientMgr, nil)

	provider := scraper.NewProvider(store, 48)
	server.SetScraper(nil, store, nil, nil, provider)

	// 1. Without Auth -> 401
	reqNoAuth := httptest.NewRequest("GET", "/api/v1/pakar", nil)
	respNoAuth, err := server.app.Test(reqNoAuth)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respNoAuth.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", respNoAuth.StatusCode)
	}

	// 2. GET /api/v1/pakar -> 8 items
	reqList := httptest.NewRequest("GET", "/api/v1/pakar", nil)
	reqList.Header.Set("X-API-Key", "secret_123")
	respList, err := server.app.Test(reqList)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respList.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", respList.StatusCode)
	}

	var listData struct {
		Success bool                    `json:"success"`
		Data    []scraper.ExpertContact `json:"data"`
	}
	if err := json.NewDecoder(respList.Body).Decode(&listData); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !listData.Success || len(listData.Data) != 8 {
		t.Errorf("expected 8 contacts, got %d", len(listData.Data))
	}

	// 3. GET /api/v1/pakar/1
	reqOne := httptest.NewRequest("GET", "/api/v1/pakar/1", nil)
	reqOne.Header.Set("X-API-Key", "secret_123")
	respOne, err := server.app.Test(reqOne)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respOne.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", respOne.StatusCode)
	}

	var oneData struct {
		Success bool                  `json:"success"`
		Data    scraper.ExpertContact `json:"data"`
	}
	_ = json.NewDecoder(respOne.Body).Decode(&oneData)
	if oneData.Data.BidangID != 1 {
		t.Errorf("expected bidang 1, got %d", oneData.Data.BidangID)
	}

	// 4. POST /api/v1/pakar/1 -> Update
	updatePayload := map[string]string{
		"name":  "Bpk. Ahmad Fauzi, S.Tr.Stat.",
		"phone": "6281299998888",
	}
	bodyBytes, _ := json.Marshal(updatePayload)
	reqUpdate := httptest.NewRequest("POST", "/api/v1/pakar/1", bytes.NewReader(bodyBytes))
	reqUpdate.Header.Set("X-API-Key", "secret_123")
	reqUpdate.Header.Set("Content-Type", "application/json")
	respUpdate, err := server.app.Test(reqUpdate)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respUpdate.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", respUpdate.StatusCode)
	}

	// 5. GET /api/v1/experts/1 (English alias) -> check updated
	reqAlias := httptest.NewRequest("GET", "/api/v1/experts/1", nil)
	reqAlias.Header.Set("X-API-Key", "secret_123")
	respAlias, err := server.app.Test(reqAlias)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	var aliasData struct {
		Success bool                  `json:"success"`
		Data    scraper.ExpertContact `json:"data"`
	}
	_ = json.NewDecoder(respAlias.Body).Decode(&aliasData)
	if aliasData.Data.PakarName != "Bpk. Ahmad Fauzi, S.Tr.Stat." || aliasData.Data.Phone != "6281299998888" {
		t.Errorf("expected updated name/phone, got %s / %s", aliasData.Data.PakarName, aliasData.Data.Phone)
	}
}
