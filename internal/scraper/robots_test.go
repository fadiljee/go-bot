package scraper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestRobotsManager(t *testing.T) {
	robotsContent := `User-agent: *
Disallow: /admin/
Disallow: /private/
Allow: /id/pressrelease.html
`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(robotsContent))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	mgr := NewRobotsManager(ts.Client(), ts.URL, "BotIstimewa/3.1", true)
	ctx := context.Background()

	uAllowed, _ := url.Parse(ts.URL + "/id/pressrelease.html")
	if !mgr.Allowed(ctx, uAllowed) {
		t.Errorf("expected %s to be allowed", uAllowed.String())
	}

	uDisallowed, _ := url.Parse(ts.URL + "/admin/login.html")
	if mgr.Allowed(ctx, uDisallowed) {
		t.Errorf("expected %s to be disallowed", uDisallowed.String())
	}
}

func TestRobotsManager_404(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer ts.Close()

	mgr := NewRobotsManager(ts.Client(), ts.URL, "BotIstimewa/3.1", true)
	ctx := context.Background()

	u, _ := url.Parse(ts.URL + "/anything.html")
	if !mgr.Allowed(ctx, u) {
		t.Errorf("expected 404 robots.txt to allow everything, got disallowed")
	}
}

func TestRobotsManager_500(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	mgr := NewRobotsManager(ts.Client(), ts.URL, "BotIstimewa/3.1", true)
	ctx := context.Background()

	u, _ := url.Parse(ts.URL + "/id/pressrelease.html")
	if mgr.Allowed(ctx, u) {
		t.Errorf("expected 500 robots.txt to be conservative (disallow), got allowed")
	}
}
