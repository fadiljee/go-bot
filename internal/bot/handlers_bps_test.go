package bot

import (
	"context"
	"testing"

	"bot-istimewa/internal/config"
	"bot-istimewa/internal/scraper"
)

func TestBotScraperTablesForTopic(t *testing.T) {
	ctx := context.Background()
	store, err := scraper.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create memory store: %v", err)
	}
	defer store.Close()

	items := []scraper.Item{
		{
			Kind:      scraper.KindStaticTable,
			Lang:      "id",
			SourceKey: "table-1",
			Title:     "Jumlah Penduduk Menurut Kecamatan 2023",
			IsActive:  true,
		},
	}
	_, _, _ = store.ApplyItems(ctx, scraper.KindStaticTable, "id", items, true)
	provider := scraper.NewProvider(store, 24)

	cfg := &config.Config{
		IdleTimeoutMinutes: 30,
	}

	router := NewBotRouter(cfg, nil, nil)
	router.SetScraper(provider)

	tablesText := router.fetchBPSTablesForTopic(ctx, "Penduduk")
	if tablesText == "" {
		t.Errorf("expected non-empty tablesText from scraper")
	}
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
