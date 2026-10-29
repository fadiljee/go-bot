package whatsapp

import (
	"sync"
	"time"
)

// MessageDeduplicator keeps track of processed WhatsApp Message IDs with a TTL
// to prevent duplicate processing caused by WhatsApp server retries or network resends.
type MessageDeduplicator struct {
	mu      sync.Mutex
	seen    map[string]time.Time
	ttl     time.Duration
	maxSize int
}

// NewMessageDeduplicator creates a new deduplicator with the specified TTL and optional max size.
func NewMessageDeduplicator(ttl time.Duration, maxSize ...int) *MessageDeduplicator {
	if ttl <= 0 {
		ttl = 60 * time.Second
	}
	limit := 5000
	if len(maxSize) > 0 && maxSize[0] > 0 {
		limit = maxSize[0]
	}
	return &MessageDeduplicator{
		seen:    make(map[string]time.Time),
		ttl:     ttl,
		maxSize: limit,
	}
}

// IsDuplicate returns true if the message ID has already been seen within the TTL window.
// If the ID has not been seen, it records the ID and returns false.
func (d *MessageDeduplicator) IsDuplicate(id string) bool {
	if id == "" {
		return false
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()

	// Periodic cleanup if map exceeds maxSize
	if len(d.seen) > d.maxSize {
		for k, t := range d.seen {
			if now.Sub(t) > d.ttl {
				delete(d.seen, k)
			}
		}
	}

	if exp, exists := d.seen[id]; exists {
		if now.Sub(exp) < d.ttl {
			return true
		}
	}

	d.seen[id] = now
	return false
}
