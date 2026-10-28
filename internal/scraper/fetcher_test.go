package scraper

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestFetcher_Scenarios(t *testing.T) {
	validHTML := `<!DOCTYPE html><html><head><title>Badan Pusat Statistik</title></head><body><div id="content">` +
		strings.Repeat("<p>Badan Pusat Statistik Kabupaten Bangka Barat menyajikan data statistik resmi daerah.</p>", 50) +
		`</div></body></html>`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok.html":
			if r.Header.Get("If-None-Match") == "etag-123" {
				w.WriteHeader(http.StatusNotModified)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("ETag", "etag-123")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(validHTML))

		case "/403.html":
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Forbidden"))

		case "/429.html":
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("Too Many Requests"))

		case "/blocked-text.html":
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			// Status 200 but contains Cloudflare block challenge and large enough
			w.Write([]byte("<html><head><title>Just a moment...</title></head><body>" + strings.Repeat("Checking your browser before accessing bangkabaratkab.bps.go.id.", 40) + "</body></html>"))

		case "/small.html":
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<html><body>Too small</body></html>"))

		case "/json.json":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(strings.Repeat(`{"key":"value"}`, 200)))

		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	cfg := DefaultConfig()
	cfg.AppEnv = "test"
	cfg.AllowedHosts = []string{u.Hostname()}
	cfg.MinDelayMs = 0
	cfg.JitterMs = 0

	fetcher := NewFetcher(cfg, nil, ts.Client())
	ctx := context.Background()

	// 1. Test 200 OK
	resp, err := fetcher.Get(ctx, ts.URL+"/ok.html", nil)
	if err != nil {
		t.Fatalf("expected 200 OK success, got: %v", err)
	}
	if resp.ETag != "etag-123" {
		t.Errorf("expected etag-123, got %s", resp.ETag)
	}

	// 2. Test 304 Not Modified
	_, err = fetcher.Get(ctx, ts.URL+"/ok.html", &Conditional{ETag: "etag-123"})
	if err != ErrNotModified {
		t.Errorf("expected ErrNotModified, got %v", err)
	}

	// 3. Test 403 Forbidden
	_, err = fetcher.Get(ctx, ts.URL+"/403.html", nil)
	if err != ErrBlocked {
		t.Errorf("expected ErrBlocked for 403, got %v", err)
	}

	// 4. Test 429 Too Many Requests
	_, err = fetcher.Get(ctx, ts.URL+"/429.html", nil)
	if err != ErrRateLimited {
		t.Errorf("expected ErrRateLimited for 429, got %v", err)
	}

	// 5. Test Blocked Text (Status 200 with 'Just a moment')
	_, err = fetcher.Get(ctx, ts.URL+"/blocked-text.html", nil)
	if err != ErrBlocked {
		t.Errorf("expected ErrBlocked for challenge text, got %v", err)
	}

	// 6. Test Body too small (<2048 bytes)
	_, err = fetcher.Get(ctx, ts.URL+"/small.html", nil)
	if err != ErrBlocked {
		t.Errorf("expected ErrBlocked for body < 2KB, got %v", err)
	}

	// 7. Test Non-HTML content
	_, err = fetcher.Get(ctx, ts.URL+"/json.json", nil)
	if err != ErrNotHTML {
		t.Errorf("expected ErrNotHTML, got %v", err)
	}

	// 8. Test Disallowed Host
	_, err = fetcher.Get(ctx, "https://malicious.com/data.html", nil)
	if !errors.Is(err, ErrHostNotAllowed) {
		t.Errorf("expected ErrHostNotAllowed, got %v", err)
	}
}
