package scraper

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProvider_QueriesAndCache(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "prov_test_*")
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	store, _ := NewStore(dbPath)
	defer store.Close()

	ctx := context.Background()
	now := time.Now()

	// Seed item
	item := Item{
		Kind:        KindPublication,
		Lang:        "id",
		SourceKey:   "pub-1",
		Title:       "Bangka Barat Dalam Angka 2024",
		URL:         "https://bangkabaratkab.bps.go.id/pub/1.html",
		ReleaseDate: &now,
		ContentHash: "h1",
	}
	_, _, _ = store.ApplyItems(ctx, KindPublication, "id", []Item{item}, true)

	prov := NewProvider(store, 48)

	// First read (DB -> cache)
	items, f, err := prov.LatestPublications(ctx, "id", 5)
	if err != nil || len(items) != 1 {
		t.Fatalf("expected 1 item, got %d (err: %v)", len(items), err)
	}
	if f.Source != "scraper" || f.Stale {
		t.Errorf("expected fresh scraper source, got %+v", f)
	}

	// Second read (from in-memory cache)
	cachedItems, _, err := prov.LatestPublications(ctx, "id", 5)
	if err != nil || len(cachedItems) != 1 {
		t.Fatalf("expected 1 cached item, got %d", len(cachedItems))
	}

	// Invalidate cache
	prov.InvalidateCache()
	if len(prov.cache) != 0 {
		t.Errorf("expected empty cache after InvalidateCache()")
	}
}
