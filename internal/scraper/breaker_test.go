package scraper

import (
	"context"
	"testing"
	"time"
)

type mockBreakerStore struct {
	fail   int
	until  *time.Time
	reason string
}

func (m *mockBreakerStore) GetBreakerState(ctx context.Context) (int, *time.Time, string, error) {
	return m.fail, m.until, m.reason, nil
}

func (m *mockBreakerStore) SaveBreakerState(ctx context.Context, fail int, until *time.Time, reason string) error {
	m.fail = fail
	m.until = until
	m.reason = reason
	return nil
}

func TestCircuitBreaker_Escalation(t *testing.T) {
	store := &mockBreakerStore{}
	cb := NewCircuitBreaker(store)

	expectedCooldowns := []time.Duration{
		1 * time.Hour,
		2 * time.Hour,
		4 * time.Hour,
		8 * time.Hour,
		24 * time.Hour,
		24 * time.Hour, // capped at 24h
	}

	for i, expected := range expectedCooldowns {
		dur := cb.Trip("test block")
		if dur != expected {
			t.Errorf("trip #%d: expected cooldown %v, got %v", i+1, expected, dur)
		}
		isOpen, until := cb.IsOpen()
		if !isOpen || until == nil {
			t.Errorf("trip #%d: expected breaker to be open", i+1)
		}
		if store.fail != i+1 {
			t.Errorf("trip #%d: store fail count %d != %d", i+1, store.fail, i+1)
		}
	}

	// Test Reset
	cb.Reset()
	isOpen, until := cb.IsOpen()
	if isOpen || until != nil {
		t.Errorf("expected breaker to be closed after reset")
	}
	if store.fail != 0 {
		t.Errorf("expected store fail count to be 0 after reset, got %d", store.fail)
	}
}

func TestCircuitBreaker_Persistence(t *testing.T) {
	now := time.Now().Add(2 * time.Hour)
	store := &mockBreakerStore{
		fail:   2,
		until:  &now,
		reason: "blocked earlier",
	}

	// Load existing state
	cb := NewCircuitBreaker(store)
	isOpen, until := cb.IsOpen()
	if !isOpen || until == nil {
		t.Fatalf("expected restored breaker to be open")
	}

	open, fails, _, reason := cb.Status()
	if !open || fails != 2 || reason != "blocked earlier" {
		t.Errorf("status mismatch: open=%v, fails=%d, reason=%s", open, fails, reason)
	}
}
