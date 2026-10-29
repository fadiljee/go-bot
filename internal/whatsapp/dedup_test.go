package whatsapp

import (
	"testing"
	"time"
)

func TestMessageDeduplicator(t *testing.T) {
	d := NewMessageDeduplicator(100 * time.Millisecond)

	// Empty ID should not be considered duplicate
	if d.IsDuplicate("") {
		t.Error("expected empty ID to return false")
	}

	// First time seeing ID1 -> false
	if d.IsDuplicate("msg-001") {
		t.Error("expected first message ID to return false")
	}

	// Second time seeing ID1 within TTL -> true
	if !d.IsDuplicate("msg-001") {
		t.Error("expected duplicate message ID to return true")
	}

	// Seeing ID2 -> false
	if d.IsDuplicate("msg-002") {
		t.Error("expected new message ID to return false")
	}

	// Wait for TTL expiration
	time.Sleep(150 * time.Millisecond)

	// After TTL expiration, ID1 should no longer be considered duplicate
	if d.IsDuplicate("msg-001") {
		t.Error("expected expired message ID to return false after TTL")
	}
}
