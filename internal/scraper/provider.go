package scraper

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type cacheEntry struct {
	items     []Item
	freshness Freshness
	expiresAt time.Time
}

// Provider implements DataProvider backed by SQLite and in-memory cache (§6.3, §9.5).
type Provider struct {
	store         *Store
	maxStaleHours int
	cacheTTL      time.Duration

	mu    sync.RWMutex
	cache map[string]cacheEntry
}

// NewProvider creates a new DataProvider backed by the scraper Store.
func NewProvider(store *Store, maxStaleHours int) *Provider {
	if maxStaleHours <= 0 {
		maxStaleHours = 48
	}
	return &Provider{
		store:         store,
		maxStaleHours: maxStaleHours,
		cacheTTL:      5 * time.Minute,
		cache:         make(map[string]cacheEntry),
	}
}

// InvalidateCache clears the in-memory cache upon successful new scrape data (§9.5).
func (p *Provider) InvalidateCache() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cache = make(map[string]cacheEntry)
}

func (p *Provider) getCached(key string) ([]Item, Freshness, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	entry, ok := p.cache[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, Freshness{}, false
	}
	return entry.items, entry.freshness, true
}

func (p *Provider) setCached(key string, items []Item, f Freshness) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.cache[key] = cacheEntry{
		items:     items,
		freshness: f,
		expiresAt: time.Now().Add(p.cacheTTL),
	}
}

func (p *Provider) calculateFreshness(ctx context.Context, items []Item) Freshness {
	f := Freshness{
		Source:    "scraper",
		UpdatedAt: time.Now(),
		Stale:     false,
	}

	ageHours, err := p.store.GetDataAgeHours(ctx)
	if err == nil && ageHours > 0 {
		f.UpdatedAt = time.Now().Add(-time.Duration(ageHours * float64(time.Hour)))
		if ageHours > float64(p.maxStaleHours) {
			f.Stale = true
		}
	} else if len(items) > 0 && !items[0].LastSeenAt.IsZero() {
		f.UpdatedAt = items[0].LastSeenAt
		if time.Since(items[0].LastSeenAt).Hours() > float64(p.maxStaleHours) {
			f.Stale = true
		}
	}

	return f
}

func (p *Provider) LatestPressReleases(ctx context.Context, lang string, limit int) ([]Item, Freshness, error) {
	cacheKey := fmt.Sprintf("pressrelease:%s:%d", lang, limit)
	if items, f, ok := p.getCached(cacheKey); ok {
		return items, f, nil
	}

	items, err := p.store.GetLatestItems(ctx, KindPressRelease, lang, limit)
	if err != nil {
		return nil, Freshness{}, err
	}

	f := p.calculateFreshness(ctx, items)
	p.setCached(cacheKey, items, f)
	return items, f, nil
}

func (p *Provider) LatestPublications(ctx context.Context, lang string, limit int) ([]Item, Freshness, error) {
	cacheKey := fmt.Sprintf("publication:%s:%d", lang, limit)
	if items, f, ok := p.getCached(cacheKey); ok {
		return items, f, nil
	}

	items, err := p.store.GetLatestItems(ctx, KindPublication, lang, limit)
	if err != nil {
		return nil, Freshness{}, err
	}

	f := p.calculateFreshness(ctx, items)
	p.setCached(cacheKey, items, f)
	return items, f, nil
}

func (p *Provider) LatestInfographics(ctx context.Context, limit int) ([]Item, Freshness, error) {
	cacheKey := fmt.Sprintf("infographic:%d", limit)
	if items, f, ok := p.getCached(cacheKey); ok {
		return items, f, nil
	}

	items, err := p.store.GetLatestItems(ctx, KindInfographic, "id", limit)
	if err != nil {
		return nil, Freshness{}, err
	}

	f := p.calculateFreshness(ctx, items)
	p.setCached(cacheKey, items, f)
	return items, f, nil
}

func (p *Provider) SearchStaticTables(ctx context.Context, keyword string, limit int) ([]Item, Freshness, error) {
	// Keyword search is dynamic; can query Store directly
	items, err := p.store.SearchItems(ctx, KindStaticTable, keyword, limit)
	if err != nil {
		return nil, Freshness{}, err
	}

	f := p.calculateFreshness(ctx, items)
	return items, f, nil
}

func (p *Provider) SearchItems(ctx context.Context, keyword string, lang string, limit int) ([]Item, Freshness, error) {
	items, err := p.store.SearchAllItems(ctx, keyword, limit)
	if err != nil {
		return nil, Freshness{}, err
	}
	f := p.calculateFreshness(ctx, items)
	return items, f, nil
}

func (p *Provider) CreateConsultationBooking(ctx context.Context, b *ConsultationBooking) error {
	return p.store.CreateConsultationBooking(ctx, b)
}
