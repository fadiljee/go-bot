package bot

import (
	"testing"
)

func TestRateLimiter(t *testing.T) {
	// 5 messages per 10s, 2s cooldown for test
	rl := NewRateLimiter(5, 2)
	phone := "628123456789"

	// First 5 should succeed
	for i := 0; i < 5; i++ {
		if !rl.Allow(phone) {
			t.Fatalf("expected request %d to be allowed", i+1)
		}
	}

	// 6th should be denied and trigger cooldown
	if rl.Allow(phone) {
		t.Fatal("expected 6th request to be denied")
	}

	// Should still be in cooldown immediately
	if !rl.InCooldown(phone) {
		t.Fatal("expected user to be in cooldown")
	}

	if rl.Allow(phone) {
		t.Fatal("expected request during cooldown to be denied")
	}
}
