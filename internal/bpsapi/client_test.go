package bpsapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"bot-istimewa/internal/config"
)

func TestBuildURL(t *testing.T) {
	cfg := &config.Config{
		BPSAPIBaseURL:    "https://webapi.bps.go.id/v1/api/list",
		BPSAPIKey:        "test-secret-key-1903",
		BPSDefaultDomain: "1903",
		BPSDefaultLang:   "ind",
	}

	client := NewClient(cfg)

	tests := []struct {
		name     string
		params   ListParams
		expected []string
	}{
		{
			name: "default subject url",
			params: ListParams{
				Model: "subject",
			},
			expected: []string{
				"https://webapi.bps.go.id/v1/api/list/model/subject/lang/ind/domain/1903/key/test-secret-key-1903/",
			},
		},
		{
			name: "publication with year, keyword, and page",
			params: ListParams{
				Model:   "publication",
				Domain:  "1900",
				Lang:    "eng",
				Page:    2,
				Year:    "2024",
				Keyword: "Bangka Barat Dalam Angka",
			},
			expected: []string{
				"/model/publication",
				"/lang/eng",
				"/domain/1900",
				"/key/test-secret-key-1903",
				"/page/2",
				"/year/2024",
				"/keyword/Bangka%20Barat%20Dalam%20Angka",
			},
		},
		{
			name: "variables with subject filter",
			params: ListParams{
				Model:   "var",
				Subject: "26",
			},
			expected: []string{
				"/model/var",
				"/subject/26",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.BuildURL(tt.params)
			for _, exp := range tt.expected {
				if !strings.Contains(got, exp) {
					t.Errorf("BuildURL() = %q, expected to contain %q", got, exp)
				}
			}
			if !strings.HasSuffix(got, "/") {
				t.Errorf("BuildURL() = %q, expected to have trailing slash", got)
			}
		})
	}
}

func TestParseResponse(t *testing.T) {
	jsonPayload := []byte(`{
		"status": "OK",
		"data-availability": "available",
		"data": [
			{
				"page": 1,
				"pages": 3,
				"per_page": 10,
				"total": 25
			},
			[
				{
					"sub_id": 1,
					"title": "Sosial dan Kependudukan",
					"subcat_id": 1,
					"subcat": "Sosduk",
					"ntabel": 15
				},
				{
					"sub_id": 2,
					"title": "Ekonomi dan Perdagangan",
					"subcat_id": 2,
					"subcat": "Ekonomi",
					"ntabel": 20
				}
			]
		]
	}`)

	result, err := ParseResponse[Subject](jsonPayload, false)
	if err != nil {
		t.Fatalf("ParseResponse() returned unexpected error: %v", err)
	}

	if result.Status != "OK" {
		t.Errorf("expected status 'OK', got %q", result.Status)
	}
	if result.DataAvailability != "available" {
		t.Errorf("expected data-availability 'available', got %q", result.DataAvailability)
	}
	if result.Pagination == nil {
		t.Fatalf("expected pagination not nil")
	}
	if result.Pagination.Total != 25 {
		t.Errorf("expected pagination total 25, got %d", result.Pagination.Total)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}
	if result.Items[0].Title != "Sosial dan Kependudukan" {
		t.Errorf("expected item title 'Sosial dan Kependudukan', got %q", result.Items[0].Title)
	}
}

func TestClientCachingAndStaleFallback(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"status": "OK",
				"data-availability": "available",
				"data": [
					{"page": 1, "pages": 1, "per_page": 10, "total": 1},
					[{"pub_id": "P01", "title": "Bangka Barat Dalam Angka 2024", "rl_date": "2024-02-28"}]
				]
			}`))
			return
		}
		// Subsequent calls simulate server error
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := &config.Config{
		BPSAPIBaseURL:      server.URL,
		BPSAPIKey:          "dummy-key",
		BPSCacheTTLMinutes: 5,
	}

	client := NewClient(cfg)
	client.SetHTTPClient(server.Client())

	ctx := context.Background()

	// 1st Call: hits the mock server
	res1, err := client.GetPublications(ctx, ListParams{Model: "publication"})
	if err != nil {
		t.Fatalf("1st call failed: %v", err)
	}
	if len(res1.Items) != 1 || res1.Items[0].Title != "Bangka Barat Dalam Angka 2024" {
		t.Errorf("unexpected 1st call result: %+v", res1)
	}
	if res1.IsStale {
		t.Errorf("expected fresh data, got is_stale=true")
	}

	// 2nd Call: served from memory cache (callCount remains 1)
	res2, err := client.GetPublications(ctx, ListParams{Model: "publication"})
	if err != nil {
		t.Fatalf("2nd call failed: %v", err)
	}
	if len(res2.Items) != 1 {
		t.Errorf("unexpected 2nd call result: %+v", res2)
	}
	if callCount != 1 {
		t.Errorf("expected callCount 1 (cache hit), got %d", callCount)
	}

	// Invalidate fresh cache by forcing expiration
	targetURL := client.BuildURL(ListParams{Model: "publication"})
	client.cacheMu.Lock()
	client.cache[targetURL].ExpiresAt = time.Now().Add(-1 * time.Minute)
	client.cacheMu.Unlock()

	// 3rd Call: HTTP returns 500, but client safely falls back to stale cache!
	res3, err := client.GetPublications(ctx, ListParams{Model: "publication"})
	if err != nil {
		t.Fatalf("3rd call should have succeeded via stale fallback: %v", err)
	}
	if !res3.IsStale {
		t.Errorf("expected is_stale=true on fallback")
	}
	if len(res3.Items) != 1 || res3.Items[0].Title != "Bangka Barat Dalam Angka 2024" {
		t.Errorf("unexpected fallback data: %+v", res3)
	}

	// Test ClearCache
	client.ClearCache()
	client.cacheMu.RLock()
	cacheLen := len(client.cache)
	client.cacheMu.RUnlock()
	if cacheLen != 0 {
		t.Errorf("expected 0 cache entries after ClearCache, got %d", cacheLen)
	}
}
