package scraper

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	ErrBlocked          = errors.New("scraper: diblokir oleh situs (WAF/bot detection)")
	ErrRateLimited      = errors.New("scraper: rate limited (429)")
	ErrNotModified      = errors.New("scraper: 304 not modified")
	ErrRobotsDisallowed = errors.New("scraper: dilarang oleh robots.txt")
	ErrHostNotAllowed   = errors.New("scraper: host tidak diizinkan")
	ErrBodyTooLarge     = errors.New("scraper: body melebihi batas")
	ErrNotHTML          = errors.New("scraper: content-type bukan html")
	ErrBudgetExceeded   = errors.New("scraper: batas request per siklus terlampaui")
	ErrSSRF             = errors.New("scraper: url melanggar aturan anti-SSRF")
)

var blockMarkers = []string{
	"captcha", "just a moment", "attention required", "access denied",
	"checking your browser", "cf-chl", "request blocked", "bot detection",
}

// Conditional carries HTTP caching headers for conditional GET (§7.4).
type Conditional struct {
	ETag         string
	LastModified string
}

// Response holds the fetched HTTP response data.
type Response struct {
	URL          string
	Status       int
	Body         []byte
	ETag         string
	LastModified string
	FetchedAt    time.Time
}

// Fetcher performs polite, rate-limited HTTP GET requests with SSRF & anti-bot protection (§7).
type Fetcher struct {
	cfg        *Config
	client     *http.Client
	robots     *RobotsManager
	requestCount int
	countMu    sync.Mutex
	lastReqMu  sync.Mutex
	lastReqTime time.Time
}

// NewFetcher creates a Fetcher with polite defaults and custom transport.
func NewFetcher(cfg *Config, robots *RobotsManager, customClient ...*http.Client) *Fetcher {
	var client *http.Client
	if len(customClient) > 0 && customClient[0] != nil {
		client = customClient[0]
	} else {
		dialer := &net.Dialer{
			Timeout: 8 * time.Second,
		}
		transport := &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, port, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, err
				}
				if cfg.AppEnv != "test" {
					if err := checkIPAntiSSRF(host); err != nil {
						return nil, err
					}
				}
				return dialer.DialContext(ctx, network, net.JoinHostPort(host, port))
			},
			TLSHandshakeTimeout: 8 * time.Second,
			MaxIdleConns:        5,
			IdleConnTimeout:     30 * time.Second,
		}

		jar, _ := cookiejar.New(nil)
		client = &http.Client{
			Jar:       jar,
			Transport: transport,
			Timeout:   time.Duration(cfg.HTTPTimeoutSeconds) * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 3 {
					return errors.New("scraper: too many redirects (max 3)")
				}
				// Verify redirect host is in allowed hosts
				return checkHostAllowed(req.URL.Hostname(), cfg.AllowedHosts)
			},
		}
	}

	return &Fetcher{
		cfg:    cfg,
		client: client,
		robots: robots,
	}
}

// ResetBudget resets the per-run request counter.
func (f *Fetcher) ResetBudget() {
	f.countMu.Lock()
	f.requestCount = 0
	f.countMu.Unlock()
}

// RequestCount returns the total requests made in the current run.
func (f *Fetcher) RequestCount() int {
	f.countMu.Lock()
	defer f.countMu.Unlock()
	return f.requestCount
}

// ValidateURL checks host allowlist and SSRF rules (§7.8).
func (f *Fetcher) ValidateURL(rawURL string) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid URL %v", ErrSSRF, err)
	}

	if f.cfg.AppEnv != "test" && u.Scheme != "https" {
		return nil, fmt.Errorf("%w: scheme must be https, got %s", ErrSSRF, u.Scheme)
	}
	if u.User != nil {
		return nil, fmt.Errorf("%w: credentials in URL are prohibited", ErrSSRF)
	}

	hostname := u.Hostname()
	if hostname == "" {
		return nil, fmt.Errorf("%w: empty hostname", ErrSSRF)
	}

	if err := checkHostAllowed(hostname, f.cfg.AllowedHosts); err != nil {
		return nil, err
	}

	if f.cfg.AppEnv != "test" {
		if err := checkIPAntiSSRF(hostname); err != nil {
			return nil, err
		}
	}

	return u, nil
}

// Get executes a polite HTTP GET request (§7.10).
func (f *Fetcher) Get(ctx context.Context, rawURL string, cond *Conditional) (*Response, error) {
	u, err := f.ValidateURL(rawURL)
	if err != nil {
		return nil, err
	}

	// 1. Robots.txt check
	if f.robots != nil && !f.robots.Allowed(ctx, u) {
		return nil, ErrRobotsDisallowed
	}

	// 2. Budget check
	f.countMu.Lock()
	if f.requestCount >= f.cfg.MaxRequestsPerRun {
		f.countMu.Unlock()
		return nil, ErrBudgetExceeded
	}
	f.requestCount++
	f.countMu.Unlock()

	// 3. Polite Rate Limiting & Jitter Delay (§7.1)
	f.applyRateLimit(ctx)

	// 4. Build Request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", f.cfg.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml;q=0.9,*/*;q=0.5")
	req.Header.Set("Accept-Language", "id-ID,id;q=0.9,en;q=0.5")
	if cond != nil {
		if cond.ETag != "" {
			req.Header.Set("If-None-Match", cond.ETag)
		}
		if cond.LastModified != "" {
			req.Header.Set("If-Modified-Since", cond.LastModified)
		}
	}

	// 5. Execute with Retry (§7.3)
	resp, err := f.doWithRetry(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return nil, ErrNotModified
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrRateLimited
	}
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrBlocked
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("scraper: http %d untuk %s", resp.StatusCode, u.String())
	}

	contentType := resp.Header.Get("Content-Type")
	if !isHTML(contentType) {
		return nil, ErrNotHTML
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, f.cfg.MaxPageBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > f.cfg.MaxPageBytes {
		return nil, ErrBodyTooLarge
	}

	if looksBlocked(body) {
		return nil, ErrBlocked
	}

	return &Response{
		URL:          u.String(),
		Status:       resp.StatusCode,
		Body:         body,
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
		FetchedAt:    time.Now(),
	}, nil
}

// PostAction executes a polite POST request with Next-Action header for Next.js Server Actions.
func (f *Fetcher) PostAction(ctx context.Context, actionURL string, actionID string, payload string) ([]byte, error) {
	u, err := f.ValidateURL(actionURL)
	if err != nil {
		return nil, err
	}

	if f.robots != nil && !f.robots.Allowed(ctx, u) {
		return nil, ErrRobotsDisallowed
	}

	f.countMu.Lock()
	if f.cfg.MaxRequestsPerRun > 0 && f.requestCount >= f.cfg.MaxRequestsPerRun {
		f.countMu.Unlock()
		return nil, ErrBudgetExceeded
	}
	f.requestCount++
	f.countMu.Unlock()

	f.applyRateLimit(ctx)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), strings.NewReader(payload))
	if err != nil {
		return nil, err
	}
	ua := f.cfg.UserAgent
	if !strings.Contains(ua, "Chrome") {
		ua = "Mozilla/5.0 (compatible; BotIstimewa/3.1; Chrome/130.0.0.0)"
	}
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Next-Action", actionID)
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")

	resp, err := f.doWithRetry(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrRateLimited
	}
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrBlocked
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("scraper: http %d untuk %s", resp.StatusCode, u.String())
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, f.cfg.MaxPageBytes+1))
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (f *Fetcher) applyRateLimit(ctx context.Context) {
	f.lastReqMu.Lock()
	defer f.lastReqMu.Unlock()

	delay := time.Duration(f.cfg.MinDelayMs)*time.Millisecond
	if f.cfg.JitterMs > 0 {
		delay += time.Duration(rand.Intn(f.cfg.JitterMs))*time.Millisecond
	}

	if !f.lastReqTime.IsZero() {
		elapsed := time.Since(f.lastReqTime)
		if elapsed < delay {
			wait := delay - elapsed
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return
			}
		}
	}
	f.lastReqTime = time.Now()
}

func (f *Fetcher) doWithRetry(req *http.Request) (*http.Response, error) {
	maxRetries := 2
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := f.client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < maxRetries {
				sleepBackoff(attempt)
				continue
			}
			return nil, lastErr
		}

		// Handle status codes according to §7.3
		if resp.StatusCode == http.StatusTooManyRequests {
			return resp, nil // ErrRateLimited
		}
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
			return resp, nil // ErrBlocked
		}
		if resp.StatusCode == http.StatusServiceUnavailable {
			// HTTP 503: retry 1 time only (§7.3)
			if attempt == 0 {
				resp.Body.Close()
				sleepBackoff(0)
				continue
			}
			return resp, nil
		}
		if resp.StatusCode == 500 || resp.StatusCode == 502 || resp.StatusCode == 504 {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			if attempt < maxRetries {
				sleepBackoff(attempt)
				continue
			}
			return nil, lastErr
		}

		return resp, nil
	}

	return nil, lastErr
}

func sleepBackoff(attempt int) {
	// Attempt 0 -> 5s + jitter, Attempt 1 -> 20s + jitter
	base := 5 * time.Second
	if attempt > 0 {
		base = 20 * time.Second
	}
	jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
	time.Sleep(base + jitter)
}

func isHTML(contentType string) bool {
	ct := strings.ToLower(contentType)
	return strings.Contains(ct, "text/html") || strings.Contains(ct, "application/xhtml+xml")
}

func looksBlocked(body []byte) bool {
	if len(body) < 2048 {
		return true // Halaman sah BPS selalu puluhan KB (§7.7, Lampiran B)
	}
	headLen := len(body)
	if headLen > 16*1024 {
		headLen = 16 * 1024
	}
	head := strings.ToLower(string(body[:headLen]))
	for _, m := range blockMarkers {
		if strings.Contains(head, m) {
			return true
		}
	}
	return false
}

func checkHostAllowed(hostname string, allowed []string) error {
	target := strings.ToLower(strings.TrimSpace(hostname))
	for _, a := range allowed {
		if strings.EqualFold(target, strings.TrimSpace(a)) {
			return nil
		}
	}
	return fmt.Errorf("%w: host %s not in allowed list", ErrHostNotAllowed, hostname)
}

func checkIPAntiSSRF(hostname string) error {
	lower := strings.ToLower(hostname)
	if lower == "localhost" || lower == "127.0.0.1" || lower == "::1" {
		return fmt.Errorf("%w: loopback host not allowed: %s", ErrSSRF, hostname)
	}

	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil // If DNS fails now, dial will fail later
	}

	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return fmt.Errorf("%w: private/loopback IP address not allowed: %s (%s)", ErrSSRF, hostname, ip.String())
		}
	}
	return nil
}
