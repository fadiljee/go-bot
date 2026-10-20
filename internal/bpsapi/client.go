package bpsapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"bot-istimewa/internal/config"
)

const (
	DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36 Bot-Istimewa/3.0"
	MaxResponseBytes = 16 * 1024 * 1024 // 16 MB Limit
)

type CacheEntry struct {
	Data      []byte
	ExpiresAt time.Time
}

type Client struct {
	cfg        *config.Config
	httpClient *http.Client
	cacheMu    sync.RWMutex
	cache      map[string]*CacheEntry
}

func NewClient(cfg *config.Config) *Client {
	timeout := 12 * time.Second
	if cfg != nil && cfg.HTTPClientTimeoutSeconds > 0 {
		timeout = time.Duration(cfg.HTTPClientTimeoutSeconds) * time.Second
	}

	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		cache: make(map[string]*CacheEntry),
	}
}

// SetHTTPClient allows overriding the http client (useful in unit tests)
func (c *Client) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

// BuildURL constructs the full BPS Web API hierarchical endpoint
func (c *Client) BuildURL(params ListParams) string {
	baseURL := "https://webapi.bps.go.id/v1/api/list"
	apiKey := ""
	defaultDomain := "1903"
	defaultLang := "ind"

	if c.cfg != nil {
		if c.cfg.BPSAPIBaseURL != "" {
			baseURL = strings.TrimRight(c.cfg.BPSAPIBaseURL, "/")
		}
		if c.cfg.BPSAPIKey != "" {
			apiKey = c.cfg.BPSAPIKey
		}
		if c.cfg.BPSDefaultDomain != "" {
			defaultDomain = c.cfg.BPSDefaultDomain
		}
		if c.cfg.BPSDefaultLang != "" {
			defaultLang = c.cfg.BPSDefaultLang
		}
	}

	model := params.Model
	if model == "" {
		model = "subject"
	}

	domain := params.Domain
	if domain == "" {
		domain = defaultDomain
	}

	lang := params.Lang
	if lang == "" {
		lang = defaultLang
	}

	var sb strings.Builder
	sb.WriteString(baseURL)
	sb.WriteString("/model/")
	sb.WriteString(url.PathEscape(model))
	sb.WriteString("/lang/")
	sb.WriteString(url.PathEscape(lang))
	sb.WriteString("/domain/")
	sb.WriteString(url.PathEscape(domain))
	sb.WriteString("/key/")
	sb.WriteString(apiKey)

	// Optional query/filter segments
	if params.Page > 0 {
		sb.WriteString(fmt.Sprintf("/page/%d", params.Page))
	}
	if params.Subject != "" {
		sb.WriteString("/subject/")
		sb.WriteString(url.PathEscape(params.Subject))
	}
	if params.Subcat != "" {
		sb.WriteString("/subcat/")
		sb.WriteString(url.PathEscape(params.Subcat))
	}
	if params.Year != "" {
		sb.WriteString("/year/")
		sb.WriteString(url.PathEscape(params.Year))
	}
	if params.Month != "" {
		sb.WriteString("/month/")
		sb.WriteString(url.PathEscape(params.Month))
	}
	if params.Keyword != "" {
		sb.WriteString("/keyword/")
		sb.WriteString(url.PathEscape(params.Keyword))
	}
	if params.Var != "" {
		sb.WriteString("/var/")
		sb.WriteString(url.PathEscape(params.Var))
	}
	if params.Turvar != "" {
		sb.WriteString("/turvar/")
		sb.WriteString(url.PathEscape(params.Turvar))
	}
	if params.Vervar != "" {
		sb.WriteString("/vervar/")
		sb.WriteString(url.PathEscape(params.Vervar))
	}
	if params.Th != "" {
		sb.WriteString("/th/")
		sb.WriteString(url.PathEscape(params.Th))
	}
	if params.Turth != "" {
		sb.WriteString("/turth/")
		sb.WriteString(url.PathEscape(params.Turth))
	}

	// BPS Web API requires trailing slash
	if !strings.HasSuffix(sb.String(), "/") {
		sb.WriteString("/")
	}

	return sb.String()
}

// FetchRaw performs cached HTTP GET request with stale-if-error fallback
func (c *Client) FetchRaw(ctx context.Context, targetURL string) ([]byte, bool, error) {
	// 1. Check in-memory cache
	c.cacheMu.RLock()
	entry, found := c.cache[targetURL]
	c.cacheMu.RUnlock()

	if found && time.Now().Before(entry.ExpiresAt) {
		return entry.Data, false, nil
	}

	// 2. Fetch fresh data over HTTP
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		if found {
			return entry.Data, true, nil
		}
		return nil, false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", DefaultUserAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if found {
			slog.Warn("[BPS API] Request failed, returning stale cache", "url", targetURL, "err", err)
			return entry.Data, true, nil
		}
		return nil, false, fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if found {
			slog.Warn("[BPS API] Non-200 status, returning stale cache", "status", resp.StatusCode, "url", targetURL)
			return entry.Data, true, nil
		}
		return nil, false, fmt.Errorf("unexpected http status %d from BPS Web API", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxResponseBytes))
	if err != nil {
		if found {
			return entry.Data, true, nil
		}
		return nil, false, fmt.Errorf("failed to read response body: %w", err)
	}

	// 3. Store into cache
	ttlMinutes := 15
	if c.cfg != nil && c.cfg.BPSCacheTTLMinutes > 0 {
		ttlMinutes = c.cfg.BPSCacheTTLMinutes
	}

	c.cacheMu.Lock()
	c.cache[targetURL] = &CacheEntry{
		Data:      body,
		ExpiresAt: time.Now().Add(time.Duration(ttlMinutes) * time.Minute),
	}
	c.cacheMu.Unlock()

	return body, false, nil
}

// ParseResponse unmarshals the two-element array format of BPS Web API data
func ParseResponse[T any](rawBytes []byte, isStale bool) (*ListResult[T], error) {
	var envelope RawBPSResponse
	if err := json.Unmarshal(rawBytes, &envelope); err != nil {
		return nil, fmt.Errorf("failed to unmarshal BPS envelope: %w", err)
	}

	result := &ListResult[T]{
		Status:           envelope.Status,
		DataAvailability: envelope.DataAvailability,
		Items:            make([]T, 0),
		IsStale:          isStale,
	}

	if len(envelope.Data) == 0 || string(envelope.Data) == "null" || string(envelope.Data) == "[]" {
		return result, nil
	}

	var rawParts []json.RawMessage
	if err := json.Unmarshal(envelope.Data, &rawParts); err != nil {
		// Sometimes data might be returned directly as an object or array in non-standard errors
		return result, nil
	}

	// Extract Pagination from first element
	if len(rawParts) > 0 {
		var p Pagination
		if err := json.Unmarshal(rawParts[0], &p); err == nil {
			result.Pagination = &p
		}
	}

	// Extract Items list from second element
	if len(rawParts) > 1 {
		var items []T
		if err := json.Unmarshal(rawParts[1], &items); err == nil {
			result.Items = items
		}
	}

	return result, nil
}

// ClearCache flushes all cached BPS Web API responses
func (c *Client) ClearCache() {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	c.cache = make(map[string]*CacheEntry)
}

// GetSubjects retrieves the list of statistical subjects
func (c *Client) GetSubjects(ctx context.Context, params ListParams) (*ListResult[Subject], error) {
	params.Model = "subject"
	targetURL := c.BuildURL(params)
	data, isStale, err := c.FetchRaw(ctx, targetURL)
	if err != nil {
		return nil, err
	}
	return ParseResponse[Subject](data, isStale)
}

// GetSubCats retrieves the list of statistical categories
func (c *Client) GetSubCats(ctx context.Context, params ListParams) (*ListResult[SubCat], error) {
	params.Model = "subcat"
	targetURL := c.BuildURL(params)
	data, isStale, err := c.FetchRaw(ctx, targetURL)
	if err != nil {
		return nil, err
	}
	return ParseResponse[SubCat](data, isStale)
}

// GetVariables retrieves the list of statistical variables
func (c *Client) GetVariables(ctx context.Context, params ListParams) (*ListResult[Variable], error) {
	params.Model = "var"
	targetURL := c.BuildURL(params)
	data, isStale, err := c.FetchRaw(ctx, targetURL)
	if err != nil {
		return nil, err
	}
	return ParseResponse[Variable](data, isStale)
}

// GetStaticTables retrieves statistical static tables
func (c *Client) GetStaticTables(ctx context.Context, params ListParams) (*ListResult[StaticTable], error) {
	params.Model = "statictable"
	targetURL := c.BuildURL(params)
	data, isStale, err := c.FetchRaw(ctx, targetURL)
	if err != nil {
		return nil, err
	}
	return ParseResponse[StaticTable](data, isStale)
}

// GetPublications retrieves publication catalog
func (c *Client) GetPublications(ctx context.Context, params ListParams) (*ListResult[Publication], error) {
	params.Model = "publication"
	targetURL := c.BuildURL(params)
	data, isStale, err := c.FetchRaw(ctx, targetURL)
	if err != nil {
		return nil, err
	}
	return ParseResponse[Publication](data, isStale)
}

// GetPressReleases retrieves Berita Resmi Statistik (BRS)
func (c *Client) GetPressReleases(ctx context.Context, params ListParams) (*ListResult[PressRelease], error) {
	params.Model = "pressrelease"
	targetURL := c.BuildURL(params)
	data, isStale, err := c.FetchRaw(ctx, targetURL)
	if err != nil {
		return nil, err
	}
	return ParseResponse[PressRelease](data, isStale)
}

// GetInfographics retrieves statistical infographics
func (c *Client) GetInfographics(ctx context.Context, params ListParams) (*ListResult[Infographic], error) {
	params.Model = "infographic"
	targetURL := c.BuildURL(params)
	data, isStale, err := c.FetchRaw(ctx, targetURL)
	if err != nil {
		return nil, err
	}
	return ParseResponse[Infographic](data, isStale)
}

// GetData retrieves dynamic statistical variable data values
func (c *Client) GetData(ctx context.Context, params ListParams) (*ListResult[map[string]any], error) {
	params.Model = "data"
	targetURL := c.BuildURL(params)
	data, isStale, err := c.FetchRaw(ctx, targetURL)
	if err != nil {
		return nil, err
	}
	return ParseResponse[map[string]any](data, isStale)
}

// FetchGenericList executes query for any arbitrary model string
func (c *Client) FetchGenericList(ctx context.Context, params ListParams) (*ListResult[map[string]any], error) {
	targetURL := c.BuildURL(params)
	data, isStale, err := c.FetchRaw(ctx, targetURL)
	if err != nil {
		return nil, err
	}
	return ParseResponse[map[string]any](data, isStale)
}
