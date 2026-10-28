package scraper

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScheduler_Lifecycle(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "sched_test_*")
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	store, _ := NewStore(dbPath)
	defer store.Close()

	cfg := DefaultConfig()
	cfg.AppEnv = "test"
	cfg.CronSpec = "@every 1s"
	cfg.StartupCatchupHours = 1

	runner := NewRunner(cfg, store, NewFetcher(cfg, nil), NewCircuitBreaker(store), &SelectorsConfig{}, nil)
	sched := NewScheduler(cfg, runner, store)

	err := sched.Start()
	if err != nil {
		t.Fatalf("failed to start scheduler: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	next := sched.NextRun()
	if next.IsZero() {
		t.Errorf("expected non-zero next run")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	sched.Stop(ctx)
}
