package scraper

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProvider_Indicators(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ind_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	now := time.Now()

	// Seed an inflation press release
	brsItem := Item{
		Kind:        KindPressRelease,
		Lang:        "id",
		SourceKey:   "brs-inflasi-1",
		Title:       "September 2026, Inflasi Year on Year (y-on-y) Kabupaten Bangka Barat sebesar 4,49 persen.",
		URL:         "https://bangkabaratkab.bps.go.id/brs/1.html",
		FileURL:     "https://webapi.bps.go.id/download.php?f=pdf1",
		ReleaseDate: &now,
		ContentHash: "hash-brs-1",
	}
	_, _, err = store.ApplyItems(ctx, KindPressRelease, "id", []Item{brsItem}, true)
	if err != nil {
		t.Fatalf("failed to apply brs: %v", err)
	}

	prov := NewProvider(store, 48)

	// 1. Test GetAllStrategicIndicators
	all, f, err := prov.GetAllStrategicIndicators(ctx)
	if err != nil {
		t.Fatalf("GetAllStrategicIndicators failed: %v", err)
	}
	if len(all) != 7 {
		t.Fatalf("expected 7 indicators, got %d", len(all))
	}
	if f.Source != "scraper" {
		t.Errorf("expected source scraper, got %s", f.Source)
	}

	// 2. Test GetStrategicIndicator by ID
	ind1, _, err := prov.GetStrategicIndicator(ctx, "1")
	if err != nil {
		t.Fatalf("failed to get indicator 1: %v", err)
	}
	if ind1.Slug != "kependudukan" || ind1.ID != 1 {
		t.Errorf("expected slug 'kependudukan' and ID 1, got %+v", ind1)
	}

	// 3. Test GetStrategicIndicator by slug and check dynamic inflation parsing
	indInflasi, _, err := prov.GetStrategicIndicator(ctx, "inflasi")
	if err != nil {
		t.Fatalf("failed to get indicator 'inflasi': %v", err)
	}
	if indInflasi.ID != 6 {
		t.Errorf("expected ID 6 for inflasi, got %d", indInflasi.ID)
	}
	foundYoy := false
	for _, m := range indInflasi.Metrics {
		if m.Key == "inflasi_yoy" {
			foundYoy = true
			if m.Value != "4,49%" {
				t.Errorf("expected dynamically extracted value '4,49%%', got %q", m.Value)
			}
		}
	}
	if !foundYoy {
		t.Errorf("expected metric inflasi_yoy in indicator")
	}

	// 4. Test invalid slug
	_, _, err = prov.GetStrategicIndicator(ctx, "invalid-slug")
	if err == nil {
		t.Errorf("expected error for invalid slug, got nil")
	}

	// 5. Test cache invalidation
	prov.InvalidateCache()
	indCacheMu.RLock()
	cacheLen := len(indCache)
	indCacheMu.RUnlock()
	if cacheLen != 0 {
		t.Errorf("expected empty indCache after InvalidateCache, got len %d", cacheLen)
	}
}
