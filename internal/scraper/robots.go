package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/temoto/robotstxt"
)

// RobotsManager handles fetching, caching (24h), and evaluating robots.txt (§7.1, §7.3).
type RobotsManager struct {
	client    *http.Client
	baseURL   string
	userAgent string
	respect   bool

	mu        sync.RWMutex
	cachedAt  time.Time
	data      *robotstxt.RobotsData
	fetchErr  error
}

// NewRobotsManager creates a new robots.txt evaluator.
func NewRobotsManager(client *http.Client, baseURL, userAgent string, respect bool) *RobotsManager {
	return &RobotsManager{
		client:    client,
		baseURL:   baseURL,
		userAgent: userAgent,
		respect:   respect,
	}
}

// SetStaticData allows setting mock or fixture robots.txt data for tests.
func (r *RobotsManager) SetStaticData(body []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	data, err := robotstxt.FromBytes(body)
	if err != nil {
		return err
	}
	r.data = data
	r.cachedAt = time.Now()
	r.fetchErr = nil
	return nil
}

// Allowed evaluates if a given URL path is allowed for this scraper's user-agent.
func (r *RobotsManager) Allowed(ctx context.Context, targetURL *url.URL) bool {
	if !r.respect {
		return true
	}

	r.mu.RLock()
	isFresh := !r.cachedAt.IsZero() && time.Since(r.cachedAt) < 24*time.Hour
	data := r.data
	lastErr := r.fetchErr
	r.mu.RUnlock()

	if !isFresh {
		if err := r.refresh(ctx); err != nil {
			r.mu.RLock()
			data = r.data
			lastErr = r.fetchErr
			r.mu.RUnlock()
		} else {
			r.mu.RLock()
			data = r.data
			lastErr = nil
			r.mu.RUnlock()
		}
	}

	// Rule: If robots.txt resulted in 5xx/network error and we have no cached data,
	// treat as conservative/disallowed (§15 Matriks Penanganan Kegagalan).
	if lastErr != nil && data == nil {
		return false
	}

	// If no data (e.g. 404), assume allowed.
	if data == nil {
		return true
	}

	group := data.FindGroup(r.userAgent)
	if group == nil {
		group = data.FindGroup("*")
	}
	if group == nil {
		return true
	}

	return group.Test(targetURL.Path)
}

func (r *RobotsManager) refresh(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double check cache freshness
	if !r.cachedAt.IsZero() && time.Since(r.cachedAt) < 24*time.Hour {
		return nil
	}

	robotsURL, err := url.Parse(r.baseURL)
	if err != nil {
		r.fetchErr = err
		return err
	}
	robotsURL.Path = "/robots.txt"
	robotsURL.RawQuery = ""

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, robotsURL.String(), nil)
	if err != nil {
		r.fetchErr = err
		return err
	}
	req.Header.Set("User-Agent", r.userAgent)

	resp, err := r.client.Do(req)
	if err != nil {
		r.fetchErr = err
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// 404 = no restrictions
		r.data = nil
		r.fetchErr = nil
		r.cachedAt = time.Now()
		return nil
	}

	if resp.StatusCode >= 500 {
		r.fetchErr = fmt.Errorf("robots: server returned status %d", resp.StatusCode)
		return r.fetchErr
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		r.fetchErr = err
		return err
	}

	data, err := robotstxt.FromBytes(body)
	if err != nil {
		r.fetchErr = err
		return err
	}

	r.data = data
	r.cachedAt = time.Now()
	r.fetchErr = nil
	return nil
}
