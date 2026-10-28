package scraper

import (
	"context"
	"sync"
	"time"
)

// BreakerStore abstracts persistence of breaker state into SQLite or memory (§7.9).
type BreakerStore interface {
	GetBreakerState(ctx context.Context) (consecutiveFail int, openUntil *time.Time, lastReason string, err error)
	SaveBreakerState(ctx context.Context, consecutiveFail int, openUntil *time.Time, lastReason string) error
}

// CircuitBreaker manages exponential cooldown when blocked by WAF/bot detection (§7.9).
type CircuitBreaker struct {
	mu              sync.RWMutex
	consecutiveFail int
	openUntil       *time.Time
	lastReason      string
	store           BreakerStore
}

// NewCircuitBreaker creates a circuit breaker and loads any stored state.
func NewCircuitBreaker(store BreakerStore) *CircuitBreaker {
	cb := &CircuitBreaker{store: store}
	if store != nil {
		if fails, until, reason, err := store.GetBreakerState(context.Background()); err == nil {
			cb.consecutiveFail = fails
			cb.openUntil = until
			cb.lastReason = reason
		}
	}
	return cb
}

// CooldownForFailures returns the backoff duration based on consecutive failure count (§7.9).
func CooldownForFailures(failures int) time.Duration {
	switch {
	case failures <= 1:
		return 1 * time.Hour
	case failures == 2:
		return 2 * time.Hour
	case failures == 3:
		return 4 * time.Hour
	case failures == 4:
		return 8 * time.Hour
	default:
		return 24 * time.Hour
	}
}

// IsOpen checks if the breaker is open (tripped) and in cooldown.
func (cb *CircuitBreaker) IsOpen() (bool, *time.Time) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if cb.openUntil == nil {
		return false, nil
	}

	if time.Now().Before(*cb.openUntil) {
		return true, cb.openUntil
	}

	return false, nil
}

// Trip trips the breaker due to blocking/rate limiting.
func (cb *CircuitBreaker) Trip(reason string) time.Duration {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.consecutiveFail++
	cooldown := CooldownForFailures(cb.consecutiveFail)
	until := time.Now().Add(cooldown)
	cb.openUntil = &until
	cb.lastReason = reason

	if cb.store != nil {
		_ = cb.store.SaveBreakerState(context.Background(), cb.consecutiveFail, cb.openUntil, cb.lastReason)
	}

	return cooldown
}

// Reset clears the breaker state on successful runs or manual admin intervention.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.consecutiveFail = 0
	cb.openUntil = nil
	cb.lastReason = ""

	if cb.store != nil {
		_ = cb.store.SaveBreakerState(context.Background(), 0, nil, "")
	}
}

// Status returns the current breaker diagnostics (§12).
func (cb *CircuitBreaker) Status() (isOpen bool, consecutiveFail int, openUntil *time.Time, reason string) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	open := false
	if cb.openUntil != nil && time.Now().Before(*cb.openUntil) {
		open = true
	}
	return open, cb.consecutiveFail, cb.openUntil, cb.lastReason
}
