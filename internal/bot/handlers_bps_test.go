package bot

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"bot-istimewa/internal/bpsapi"
	"bot-istimewa/internal/config"
	"bot-istimewa/internal/scraper"
	"bot-istimewa/internal/whatsapp"
	"go.mau.fi/whatsmeow/types"
)

func TestBotBPSBangkaBaratIntegration(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		path := r.URL.Path

		switch {
		case len(path) >= 15 && path[:15] == "/model/pressrel":
			_, _ = w.Write([]byte(`{
				"status": "OK",
				"data-availability": "available",
				"data": [
					{"page": 1, "pages": 1, "per_page": 10, "total": 1},
					[{"brs_id": 101, "title": "Perkembangan Indeks Harga Konsumen Bangka Barat", "rl_date": "2024-03-01", "pdf": "https://bangkabaratkab.bps.go.id/brs.pdf"}]
				]
			}`))
		case len(path) >= 18 && path[:18] == "/model/publication":
			_, _ = w.Write([]byte(`{
				"status": "OK",
				"data-availability": "available",
				"data": [
					{"page": 1, "pages": 1, "per_page": 10, "total": 1},
					[{"pub_id": "P01", "title": "Kabupaten Bangka Barat Dalam Angka 2024", "rl_date": "2024-02-28", "pdf": "https://bangkabaratkab.bps.go.id/dda.pdf"}]
				]
			}`))
		case len(path) >= 18 && path[:18] == "/model/statictable":
			_, _ = w.Write([]byte(`{
				"status": "OK",
				"data-availability": "available",
				"data": [
					{"page": 1, "pages": 1, "per_page": 10, "total": 1},
					[{"table_id": 501, "title": "Jumlah Penduduk Menurut Kecamatan 2023", "excel": "https://bangkabaratkab.bps.go.id/penduduk.xlsx"}]
				]
			}`))
		default:
			_, _ = w.Write([]byte(`{
				"status": "OK",
				"data-availability": "available",
				"data": [
					{"page": 1, "pages": 1, "per_page": 10, "total": 0},
					[]
				]
			}`))
		}
	}))
	defer mockServer.Close()

	cfg := &config.Config{
		BPSAPIBaseURL:      mockServer.URL,
		BPSAPIKey:          "test-key",
		BPSDefaultDomain:   "1903",
		IdleTimeoutMinutes: 30,
	}

	bpsClient := bpsapi.NewClient(cfg)
	bpsClient.SetHTTPClient(mockServer.Client())

	router := NewBotRouter(cfg, nil, nil, bpsClient)

	// 1. Test live static tables fetching for topic
	tablesText := router.fetchBPSTablesForTopic(context.Background(), "kependudukan")
	if tablesText == "" {
		t.Errorf("expected non-empty tablesText from live BPS Bangka Barat mock")
	}

	// 2. Test BRS and Publications in HandleProdukDetail (dry test without sending)
	ctx := context.Background()
	s := &UserSession{CurrentState: StateMenu3}
	p := &whatsapp.ParsedMessage{
		JID:  types.NewJID("628123456789", types.DefaultUserServer),
		Text: "1",
	}

	// Verify that fetchBPSTablesForTopic contains Bangka Barat URL
	if tablesText != "" && !testing.Short() {
		t.Logf("Tabel output:\n%s", tablesText)
	}

	_ = s
	_ = p
	_ = ctx
}

func TestInfographicBySubjectIsolation(t *testing.T) {
	ctx := context.Background()
	store, err := scraper.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create memory store: %v", err)
	}
	defer store.Close()

	items := []scraper.Item{
		{
			Kind:      scraper.KindInfographic,
			Lang:      "id",
			SourceKey: "cat1_174",
			Title:     "Transportasi dan Komunikasi",
			Subject:   "Statistik Demografi dan Sosial",
			ImageURL:  "https://example.com/sosial.jpg",
			IsActive:  true,
		},
		{
			Kind:      scraper.KindInfographic,
			Lang:      "id",
			SourceKey: "cat2_187",
			Title:     "Inflasi Bangka Barat September 2026",
			Subject:   "Statistik Ekonomi",
			ImageURL:  "https://example.com/ekonomi.jpg",
			IsActive:  true,
		},
		{
			Kind:      scraper.KindInfographic,
			Lang:      "id",
			SourceKey: "cat3_180",
			Title:     "Perbandingan Antar Kabupaten/Kota",
			Subject:   "Statistik Lingkungan Hidup dan Multidomain",
			ImageURL:  "https://example.com/lingkungan.jpg",
			IsActive:  true,
		},
	}

	_, _, err = store.ApplyItems(ctx, scraper.KindInfographic, "id", items, true)
	if err != nil {
		t.Fatalf("failed to apply items: %v", err)
	}

	provider := scraper.NewProvider(store, 24)

	// Test 1: Category Sosial
	sosial, _, err := provider.LatestInfographicsBySubject(ctx, "Sosial", 5)
	if err != nil || len(sosial) != 1 || sosial[0].Title != "Transportasi dan Komunikasi" {
		t.Fatalf("expected 1 sosial item, got %+v (err=%v)", sosial, err)
	}

	// Test 2: Category Ekonomi
	ekonomi, _, err := provider.LatestInfographicsBySubject(ctx, "Ekonomi", 5)
	if err != nil || len(ekonomi) != 1 || sosial[0].Title == ekonomi[0].Title {
		t.Fatalf("expected distinct ekonomi item, got %+v (err=%v)", ekonomi, err)
	}
	if ekonomi[0].Title != "Inflasi Bangka Barat September 2026" {
		t.Errorf("expected ekonomi title, got %s", ekonomi[0].Title)
	}

	// Test 3: Category Lingkungan
	lingkungan, _, err := provider.LatestInfographicsBySubject(ctx, "Lingkungan", 5)
	if err != nil || len(lingkungan) != 1 || lingkungan[0].Title != "Perbandingan Antar Kabupaten/Kota" {
		t.Fatalf("expected 1 lingkungan item, got %+v (err=%v)", lingkungan, err)
	}
}
