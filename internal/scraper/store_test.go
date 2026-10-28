package scraper

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStore_UpsertAndMissedRuns(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "store_test_*")
	if err != nil {
		t.Fatal(err)
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

	item1 := Item{
		Kind:        KindPressRelease,
		Lang:        "id",
		SourceKey:   "101",
		Title:       "Berita Resmi Statistik A",
		URL:         "https://bangkabaratkab.bps.go.id/brs/101.html",
		ReleaseDate: &now,
		ContentHash: "hash-a1",
	}

	item2 := Item{
		Kind:        KindPressRelease,
		Lang:        "id",
		SourceKey:   "102",
		Title:       "Berita Resmi Statistik B",
		URL:         "https://bangkabaratkab.bps.go.id/brs/102.html",
		ReleaseDate: &now,
		ContentHash: "hash-b1",
	}

	// 1. First cycle: Insert both items
	newCount, updatedCount, err := store.ApplyItems(ctx, KindPressRelease, "id", []Item{item1, item2}, true)
	if err != nil {
		t.Fatalf("cycle 1 failed: %v", err)
	}
	if newCount != 2 || updatedCount != 0 {
		t.Errorf("cycle 1: expected 2 new, 0 updated, got %d new, %d updated", newCount, updatedCount)
	}

	// 2. Second cycle: Identical items (Idempotent)
	newCount, updatedCount, err = store.ApplyItems(ctx, KindPressRelease, "id", []Item{item1, item2}, true)
	if err != nil {
		t.Fatalf("cycle 2 failed: %v", err)
	}
	if newCount != 0 || updatedCount != 0 {
		t.Errorf("cycle 2: expected 0 new, 0 updated, got %d new, %d updated", newCount, updatedCount)
	}

	// 3. Third cycle: item1 has updated hash
	item1Updated := item1
	item1Updated.Title = "Berita Resmi Statistik A (Revisi)"
	item1Updated.ContentHash = "hash-a2"
	newCount, updatedCount, err = store.ApplyItems(ctx, KindPressRelease, "id", []Item{item1Updated, item2}, true)
	if err != nil {
		t.Fatalf("cycle 3 failed: %v", err)
	}
	if newCount != 0 || updatedCount != 1 {
		t.Errorf("cycle 3: expected 0 new, 1 updated, got %d new, %d updated", newCount, updatedCount)
	}

	// 4. Test Missed Runs: item2 disappears in cycle 4, 5, 6
	// Cycle 4 (missed 1)
	_, _, _ = store.ApplyItems(ctx, KindPressRelease, "id", []Item{item1Updated}, true)
	items, _ := store.GetLatestItems(ctx, KindPressRelease, "id", 10)
	if len(items) != 2 {
		t.Errorf("expected both items still active after 1 missed run, got %d", len(items))
	}

	// Cycle 5 (missed 2)
	_, _, _ = store.ApplyItems(ctx, KindPressRelease, "id", []Item{item1Updated}, true)
	items, _ = store.GetLatestItems(ctx, KindPressRelease, "id", 10)
	if len(items) != 2 {
		t.Errorf("expected both items still active after 2 missed runs, got %d", len(items))
	}

	// Cycle 6 (missed 3 -> deactivated)
	_, _, _ = store.ApplyItems(ctx, KindPressRelease, "id", []Item{item1Updated}, true)
	items, _ = store.GetLatestItems(ctx, KindPressRelease, "id", 10)
	if len(items) != 1 {
		t.Errorf("expected 1 item active after 3 missed runs, got %d", len(items))
	}
	if items[0].SourceKey != "101" {
		t.Errorf("expected active item to be 101, got %s", items[0].SourceKey)
	}

	// 5. Test Search
	searchRes, err := store.SearchItems(ctx, KindPressRelease, "Revisi", 10)
	if err != nil || len(searchRes) != 1 {
		t.Fatalf("expected 1 search result, got %d (err: %v)", len(searchRes), err)
	}
}

func TestStore_BreakerPersistence(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "store_breaker_*")
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	store, _ := NewStore(dbPath)
	defer store.Close()

	ctx := context.Background()
	now := time.Now().Add(4 * time.Hour)

	err := store.SaveBreakerState(ctx, 3, &now, "403 blocked")
	if err != nil {
		t.Fatalf("SaveBreakerState failed: %v", err)
	}

	fails, until, reason, err := store.GetBreakerState(ctx)
	if err != nil {
		t.Fatalf("GetBreakerState failed: %v", err)
	}
	if fails != 3 || until == nil || reason != "403 blocked" {
		t.Errorf("mismatch: fails=%d, until=%v, reason=%s", fails, until, reason)
	}
}
