package api

import (
	"context"
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

func TestAPILayananData(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "api_layanan_test_*")
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

	// Seed static tables and publications
	tableItem := scraper.Item{
		Kind:      scraper.KindStaticTable,
		Lang:      "id",
		SourceKey: "table-1",
		Title:     "Angka Kemiskinan Kabupaten Bangka Barat 2010-2014",
		Subject:   "Kondisi Tempat Tinggal, Kemiskinan, dan Permasalahan Sosial Lintas Sektor",
		URL:       "https://bangkabaratkab.bps.go.id/id/statistics-table/1.html",
	}
	_, _, _ = store.ApplyItems(ctx, scraper.KindStaticTable, "id", []scraper.Item{tableItem}, true)

	pubItem := scraper.Item{
		Kind:      scraper.KindPublication,
		Lang:      "id",
		SourceKey: "pub-1",
		Title:     "Profil Kemiskinan dan Kesejahteraan Rakyat",
		URL:       "https://bangkabaratkab.bps.go.id/id/publication/1.html",
		FileURL:   "https://webapi.bps.go.id/download.php?f=pdf1",
	}
	_, _, _ = store.ApplyItems(ctx, scraper.KindPublication, "id", []scraper.Item{pubItem}, true)

	cfg := &config.Config{
		AppPort:              "3000",
		APISecretKey:         "secret_123",
		ScraperMaxStaleHours: 48,
	}

	clientMgr := &whatsapp.ClientManager{IsLoggedIn: true}
	server := NewAPIServer(cfg, clientMgr, nil)

	provider := scraper.NewProvider(store, 48)
	server.SetScraper(nil, store, nil, nil, provider)

	// 1. GET /api/v1/layanan-data without auth -> 401
	reqNoAuth := httptest.NewRequest("GET", "/api/v1/layanan-data", nil)
	respNoAuth, err := server.app.Test(reqNoAuth)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respNoAuth.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", respNoAuth.StatusCode)
	}

	// 2. GET /api/v1/layanan-data with auth -> 200, returns 4 categories
	reqAuth := httptest.NewRequest("GET", "/api/v1/layanan-data", nil)
	reqAuth.Header.Set("X-API-Key", "secret_123")
	respAuth, err := server.app.Test(reqAuth)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respAuth.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", respAuth.StatusCode)
	}

	var catBody struct {
		Success         bool                   `json:"success"`
		TotalCategories int                    `json:"total_categories"`
		Data            []scraper.DataCategory `json:"data"`
	}
	if err := json.NewDecoder(respAuth.Body).Decode(&catBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !catBody.Success || catBody.TotalCategories != 4 {
		t.Errorf("expected 4 categories, got %d", catBody.TotalCategories)
	}

	// 3. GET /api/v1/layanan-data/sosduk -> 200, returns 12 topics
	reqSosduk := httptest.NewRequest("GET", "/api/v1/layanan-data/sosduk", nil)
	reqSosduk.Header.Set("X-API-Key", "secret_123")
	respSosduk, err := server.app.Test(reqSosduk)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respSosduk.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for sosduk category, got %d", respSosduk.StatusCode)
	}

	var singleCat struct {
		Success bool                 `json:"success"`
		Data    scraper.DataCategory `json:"data"`
	}
	if err := json.NewDecoder(respSosduk.Body).Decode(&singleCat); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(singleCat.Data.Topics) != 12 {
		t.Errorf("expected 12 topics in sosduk, got %d", len(singleCat.Data.Topics))
	}

	// 4. GET /api/v1/layanan-data/sosduk/kemiskinan -> 200, returns kemiskinan topic with tables
	reqTopic := httptest.NewRequest("GET", "/api/v1/layanan-data/sosduk/kemiskinan", nil)
	reqTopic.Header.Set("X-API-Key", "secret_123")
	respTopic, err := server.app.Test(reqTopic)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respTopic.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for kemiskinan topic, got %d", respTopic.StatusCode)
	}

	var topicBody struct {
		Success bool              `json:"success"`
		Data    scraper.DataTopic `json:"data"`
	}
	if err := json.NewDecoder(respTopic.Body).Decode(&topicBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if topicBody.Data.ID != 2 || len(topicBody.Data.Tables) == 0 {
		t.Errorf("expected topic ID 2 with enriched tables, got ID=%d tables=%d", topicBody.Data.ID, len(topicBody.Data.Tables))
	}

	// 5. GET /api/v1/layanan-data?q=kemiskinan -> 200, searches tables and items
	reqSearch := httptest.NewRequest("GET", "/api/v1/layanan-data?q=kemiskinan", nil)
	reqSearch.Header.Set("X-API-Key", "secret_123")
	respSearch, err := server.app.Test(reqSearch)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respSearch.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for search query, got %d", respSearch.StatusCode)
	}

	var searchBody struct {
		Success     bool           `json:"success"`
		TotalTables int            `json:"total_tables"`
		TotalItems  int            `json:"total_items"`
		Tables      []scraper.Item `json:"tables"`
		Items       []scraper.Item `json:"items"`
	}
	if err := json.NewDecoder(respSearch.Body).Decode(&searchBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !searchBody.Success || searchBody.TotalTables == 0 || searchBody.TotalItems == 0 {
		t.Errorf("expected search results for kemiskinan, got tables=%d items=%d", searchBody.TotalTables, searchBody.TotalItems)
	}

	// 6. GET /api/v1/layanan-data/sosduk/1?format=text -> 200, returns plaintext
	reqText := httptest.NewRequest("GET", "/api/v1/layanan-data/sosduk/1?format=text", nil)
	reqText.Header.Set("X-API-Key", "secret_123")
	respText, err := server.app.Test(reqText)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respText.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for plaintext format, got %d", respText.StatusCode)
	}
}
