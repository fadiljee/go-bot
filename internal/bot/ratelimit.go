package bot

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type UserLimiter struct {
	limiter       *rate.Limiter
	cooldownUntil time.Time
	lastSeen      time.Time
}

type RateLimiter struct {
	mu           sync.Mutex
	users        map[string]*UserLimiter
	rateLimit10s int
	cooldownSec  int
}

func NewRateLimiter(rateLimit10s, cooldownSec int) *RateLimiter {
	if rateLimit10s <= 0 {
		rateLimit10s = 5
	}
	if cooldownSec <= 0 {
		cooldownSec = 60
	}

	rl := &RateLimiter{
		users:        make(map[string]*UserLimiter),
		rateLimit10s: rateLimit10s,
		cooldownSec:  cooldownSec,
	}

	// Periodically clean up stale user entries
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			rl.cleanup()
		}
	}()

	return rl
}

// Allow checks if the message is allowed under the rate limit.
// Returns true if allowed, false if rejected or user is in cooldown.
func (rl *RateLimiter) Allow(userPhone string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	ul, exists := rl.users[userPhone]
	if !exists {
		// rate: rateLimit10s per 10 seconds = rate.Every(10s / count)
		interval := (10 * time.Second) / time.Duration(rl.rateLimit10s)
		ul = &UserLimiter{
			limiter:  rate.NewLimiter(rate.Every(interval), rl.rateLimit10s),
			lastSeen: now,
		}
		rl.users[userPhone] = ul
	}

	ul.lastSeen = now

	// 1. Check if currently in cooldown
	if now.Before(ul.cooldownUntil) {
		return false
	}

	// 2. Check token bucket rate limiter
	if !ul.limiter.Allow() {
		// User flooded messages! Trigger cooldown
		ul.cooldownUntil = now.Add(time.Duration(rl.cooldownSec) * time.Second)
		return false
	}

	return true
}

// InCooldown checks if a user is currently locked out by rate limiter
func (rl *RateLimiter) InCooldown(userPhone string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	ul, exists := rl.users[userPhone]
	if !exists {
		return false
	}
	return time.Now().Before(ul.cooldownUntil)
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-15 * time.Minute)
	for phone, ul := range rl.users {
		if ul.lastSeen.Before(cutoff) && time.Now().After(ul.cooldownUntil) {
			delete(rl.users, phone)
		}
	}
}
