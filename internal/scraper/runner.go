package scraper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

const (
	TriggerCron    = "cron"
	TriggerManual  = "manual"
	TriggerStartup = "startup"
)

// AdminNotifier is an interface for sending alert notifications to admins (§14.3).
type AdminNotifier interface {
	NotifyAdmin(ctx context.Context, message string) error
}

// Runner coordinates scrape runs across kinds with lock, timeout, and anomaly detection (§10.4, Lampiran A).
type Runner struct {
	cfg         *Config
	store       *Store
	fetcher     *Fetcher
	breaker     *CircuitBreaker
	selectors   *SelectorsConfig
	parsers     map[Kind]Parser
	notifier    AdminNotifier
	provider    *Provider

	mu          sync.Mutex
	lastManual  time.Time
}

// NewRunner creates a new scraper runner.
func NewRunner(
	cfg *Config,
	store *Store,
	fetcher *Fetcher,
	breaker *CircuitBreaker,
	selectors *SelectorsConfig,
	notifier AdminNotifier,
) *Runner {
	r := &Runner{
		cfg:       cfg,
		store:     store,
		fetcher:   fetcher,
		breaker:   breaker,
		selectors: selectors,
		parsers:   make(map[Kind]Parser),
		notifier:  notifier,
	}

	for _, k := range cfg.EnabledKinds {
		if ks, err := selectors.ForKind(k); err == nil {
			r.parsers[k] = NewGenericKindParser(k, ks)
		}
	}

	return r
}

// SetProvider binds the provider for in-memory cache invalidation.
func (r *Runner) SetProvider(p *Provider) {
	r.provider = p
}

// Store returns the underlying scraper store.
func (r *Runner) Store() *Store {
	return r.store
}

// RunAll executes a full scraping cycle for all enabled kinds (§10.4).
func (r *Runner) RunAll(ctx context.Context, trigger string) (*RunResult, error) {
	if !r.mu.TryLock() {
		r.store.LogSkipped(ctx, trigger, "skipped_overlap")
		return nil, errors.New("scraper: cycle already running (skipped overlap)")
	}
	defer r.mu.Unlock()

	if trigger == TriggerManual {
		cooldown := time.Duration(r.cfg.ManualCooldownMinutes) * time.Minute
		if !r.lastManual.IsZero() && time.Since(r.lastManual) < cooldown {
			return nil, fmt.Errorf("scraper: manual trigger on cooldown, try again in %v", cooldown-time.Since(r.lastManual))
		}
		r.lastManual = time.Now()
	}

	if isOpen, until := r.breaker.IsOpen(); isOpen {
		r.store.LogSkipped(ctx, trigger, "skipped_breaker")
		slog.Warn("scrape.skipped_breaker", "open_until", until)
		return nil, fmt.Errorf("scraper: circuit breaker open until %v", until)
	}

	runTimeout := r.cfg.RunTimeout
	if runTimeout <= 0 {
		runTimeout = 10 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, runTimeout)
	defer cancel()

	run, err := r.store.StartRun(ctx, trigger, "all")
	if err != nil {
		return nil, fmt.Errorf("store: failed to start run: %w", err)
	}

	defer func() {
		if rec := recover(); rec != nil {
			stack := string(debug.Stack())
			slog.Error("scraper panic recovered", "panic", rec, "stack", stack)
			run.Status = "error"
			run.ErrorMessage = fmt.Sprintf("panic: %v", rec)
			_ = r.store.FinishRun(context.Background(), run)
		}
	}()

	r.fetcher.ResetBudget()
	allKindsSuccess := true
	hasPartial := false

	baseParsed, err := url.Parse(r.cfg.BaseURL)
	if err != nil {
		run.Status = "error"
		run.ErrorMessage = err.Error()
		_ = r.store.FinishRun(ctx, run)
		return run, err
	}

	for _, kind := range r.cfg.EnabledKinds {
		select {
		case <-ctx.Done():
			run.Status = "timeout"
			run.ErrorMessage = "run timed out"
			_ = r.store.FinishRun(context.Background(), run)
			return run, ctx.Err()
		default:
		}

		parser, ok := r.parsers[kind]
		if !ok {
			continue
		}

		items, skipped, err := r.scrapeKind(ctx, kind, parser, baseParsed)
		run.ItemsSkipped += skipped

		if err != nil {
			slog.Warn("scraper.kind_failed", "kind", kind, "error", err)
			if errors.Is(err, ErrBlocked) || errors.Is(err, ErrRateLimited) {
				cooldown := r.breaker.Trip(err.Error())
				run.Status = "blocked"
				if errors.Is(err, ErrRateLimited) {
					run.Status = "rate_limited"
				}
				run.ErrorMessage = err.Error()
				r.sendAdminNotification(ctx, fmt.Sprintf("Scraper BPS Bangka Barat: %v. Cooldown %v. Bot beralih ke data cadangan.", err, cooldown))
				allKindsSuccess = false
				break // Halt cycle immediately (§10.4)
			} else if errors.Is(err, ErrAnomalyLayoutChanged) || errors.Is(err, ErrAnomalyCountDrop) || errors.Is(err, ErrAnomalyMarkerMissing) {
				hasPartial = true
				allKindsSuccess = false
				r.sendAdminNotification(ctx, fmt.Sprintf("Scraper BPS Bangka Barat anomali: %v. Data lama tetap dipertahankan.", err))
				continue
			} else {
				hasPartial = true
				allKindsSuccess = false
				continue
			}
		}

		// Save items in transaction
		newCount, updatedCount, applyErr := r.store.ApplyItems(ctx, kind, "id", items, true)
		if applyErr != nil {
			slog.Error("store.apply_failed", "kind", kind, "error", applyErr)
			hasPartial = true
			allKindsSuccess = false
		} else {
			run.ItemsFound += len(items)
			run.ItemsNew += newCount
			run.ItemsUpdated += updatedCount
		}
	}

	run.Requests = r.fetcher.RequestCount()

	if allKindsSuccess {
		r.breaker.Reset()
		run.Status = "success"
	} else if hasPartial {
		run.Status = "partial"
	}

	_ = r.store.FinishRun(ctx, run)

	// Invalidate DataProvider memory cache (§9.5)
	if r.provider != nil {
		r.provider.InvalidateCache()
	}

	return run, nil
}

func (r *Runner) scrapeKind(ctx context.Context, kind Kind, parser Parser, baseParsed *url.URL) ([]Item, int, error) {
	// 1. Try Next.js Server Action first (for modern Next.js BPS portal)
	if saItems, saSkipped, err := r.scrapeServerAction(ctx, kind, baseParsed); err == nil && len(saItems) > 0 {
		return saItems, saSkipped, nil
	}

	maxPages := r.cfg.MaxListPages
	if maxPages <= 0 {
		maxPages = 5
	}

	var allItems []Item
	skippedCount := 0
	var lastBody []byte

	// Get selector for markers
	ks, _ := r.selectors.ForKind(kind)

	for page := 1; page <= maxPages; page++ {
		listURL := parser.ListURL(baseParsed, "id", page)

		etag, lastMod, _ := r.store.GetPageCache(ctx, listURL)
		cond := &Conditional{ETag: etag, LastModified: lastMod}

		resp, err := r.fetcher.Get(ctx, listURL, cond)
		if err != nil {
			if errors.Is(err, ErrNotModified) {
				// 304 Not Modified: Early-stop (§10.5)
				break
			}
			return nil, skippedCount, err
		}

		lastBody = resp.Body
		_ = r.store.SetPageCache(ctx, listURL, resp.ETag, resp.LastModified, "")

		pageU, _ := url.Parse(listURL)
		items, pagination, parseErr := parser.ParseList(resp.Body, pageU)
		if parseErr != nil {
			return nil, skippedCount, parseErr
		}

		for i := range items {
			it := &items[i]
			// Fetch detail if needed (§10.5)
			if it.URL != "" && (it.Abstract == "" || it.FileURL == "" || it.ImageURL == "") {
				dResp, dErr := r.fetcher.Get(ctx, it.URL, nil)
				if dErr == nil {
					_ = parser.ParseDetail(dResp.Body, pageU, it)
				}
			}

			// Validate
			if valErr := ValidateItem(it, r.cfg.AllowedHosts, r.cfg.AppEnv); valErr != nil {
				skippedCount++
				continue
			}
			allItems = append(allItems, *it)
		}

		if !pagination.HasNext {
			break
		}
	}

	// Anomaly Guard check (§8.7)
	// Check against previous counts in DB
	prevItems, _ := r.store.GetLatestItems(ctx, kind, "id", 100)
	prevCount := len(prevItems)
	if anomalyErr := AnomalyGuard(kind, len(allItems), prevCount, lastBody, ks.RequiredMarkers); anomalyErr != nil {
		return nil, skippedCount, anomalyErr
	}

	return allItems, skippedCount, nil
}

type serverActionDef struct {
	path        string
	actionID    string
	arrayPrefix string
}

var serverActions = map[Kind]serverActionDef{
	KindPressRelease: {
		path:        "/id/pressrelease",
		actionID:    "60faa5fc36a0363681ae70a99e757ca04fc85e382f",
		arrayPrefix: `[{"brs_id":`,
	},
	KindPublication: {
		path:        "/id/publication",
		actionID:    "60624e3a7c825a0b98392f2120a4870bd26b8f90f2",
		arrayPrefix: `[{"pub_id":`,
	},
	KindNews: {
		path:        "/id/news",
		actionID:    "608404231d97eefc336168319bc8c592791e348afc",
		arrayPrefix: `[{"news_id":`,
	},
	KindStaticTable: {
		path:        "/id/statistics-table",
		actionID:    "40043348a3c6d33ba8e1e4554e4e5c37b872433488",
		arrayPrefix: `[{"id":`,
	},
}

func (r *Runner) scrapeServerAction(ctx context.Context, kind Kind, baseParsed *url.URL) ([]Item, int, error) {
	def, ok := serverActions[kind]
	if !ok {
		return nil, 0, errors.New("no server action defined for kind")
	}

	pageURL := fmt.Sprintf("%s://%s%s", baseParsed.Scheme, baseParsed.Host, def.path)
	// Initial GET to acquire session cookies
	_, _ = r.fetcher.Get(ctx, pageURL, nil)

	var allItems []Item
	skippedCount := 0

	maxPages := r.cfg.MaxListPages
	if maxPages <= 0 {
		maxPages = 2
	}

	for page := 1; page <= maxPages; page++ {
		payload := fmt.Sprintf(`["id",{"page":%d}]`, page)
		if kind == KindStaticTable {
			payload = fmt.Sprintf(`[{"locale":"id","page":%d,"perpage":20}]`, page)
		}
		respBody, err := r.fetcher.PostAction(ctx, pageURL, def.actionID, payload)
		if err != nil {
			if len(allItems) > 0 {
				break
			}
			return nil, skippedCount, err
		}

		jsonBlock := extractJSONBlock(string(respBody), def.arrayPrefix)
		if jsonBlock == nil {
			break
		}

		pageItems, err := r.parseServerActionJSON(kind, jsonBlock, baseParsed)
		if err != nil || len(pageItems) == 0 {
			break
		}

		for i := range pageItems {
			it := &pageItems[i]
			if valErr := ValidateItem(it, r.cfg.AllowedHosts, r.cfg.AppEnv); valErr != nil {
				skippedCount++
				continue
			}
			allItems = append(allItems, *it)
		}
	}

	return allItems, skippedCount, nil
}

func (r *Runner) parseServerActionJSON(kind Kind, block []byte, baseParsed *url.URL) ([]Item, error) {
	switch kind {
	case KindPressRelease:
		var raw []struct {
			ID        int    `json:"brs_id"`
			Title     string `json:"title"`
			Abstract  string `json:"abstract"`
			Date      string `json:"rl_date"`
			PDF       string `json:"pdf"`
			Size      string `json:"size"`
			Thumbnail string `json:"thumbnail"`
		}
		if err := json.Unmarshal(block, &raw); err != nil {
			return nil, err
		}
		var items []Item
		for _, it := range raw {
			t, _ := time.Parse("2006-01-02", it.Date)
			u := fmt.Sprintf("%s://%s/id/pressrelease/%s/%d/%s.html",
				baseParsed.Scheme, baseParsed.Host, strings.ReplaceAll(it.Date, "-", "/"), it.ID, url.PathEscape(it.Title))
			item := Item{
				Kind:         KindPressRelease,
				Lang:         "id",
				SourceKey:    fmt.Sprintf("%d", it.ID),
				URL:          u,
				Title:        it.Title,
				Abstract:     cleanHTMLText(it.Abstract),
				ReleaseDate:  &t,
				FileURL:      it.PDF,
				FileSizeText: it.Size,
				ThumbnailURL: it.Thumbnail,
				ImageURL:     it.Thumbnail,
				IsActive:     true,
			}
			item.ContentHash = CalculateContentHash(&item)
			items = append(items, item)
		}
		return items, nil

	case KindPublication:
		var raw []struct {
			ID       string `json:"pub_id"`
			Title    string `json:"title"`
			Abstract string `json:"abstract"`
			Date     string `json:"rl_date"`
			Cover    string `json:"cover"`
			PDF      string `json:"pdf"`
			Size     string `json:"size"`
		}
		if err := json.Unmarshal(block, &raw); err != nil {
			return nil, err
		}
		var items []Item
		for _, it := range raw {
			t, _ := time.Parse("2006-01-02", it.Date)
			u := fmt.Sprintf("%s://%s/id/publication/%s/%s.html",
				baseParsed.Scheme, baseParsed.Host, strings.ReplaceAll(it.Date, "-", "/"), it.ID)
			item := Item{
				Kind:         KindPublication,
				Lang:         "id",
				SourceKey:    it.ID,
				URL:          u,
				Title:        it.Title,
				Abstract:     cleanHTMLText(it.Abstract),
				ReleaseDate:  &t,
				FileURL:      it.PDF,
				FileSizeText: it.Size,
				ThumbnailURL: it.Cover,
				ImageURL:     it.Cover,
				IsActive:     true,
			}
			item.ContentHash = CalculateContentHash(&item)
			items = append(items, item)
		}
		return items, nil

	case KindNews:
		var raw []struct {
			ID      int    `json:"news_id"`
			Title   string `json:"title"`
			News    string `json:"news"`
			Date    string `json:"rl_date"`
			Picture string `json:"picture"`
		}
		if err := json.Unmarshal(block, &raw); err != nil {
			return nil, err
		}
		var items []Item
		for _, it := range raw {
			t, _ := time.Parse("2006-01-02", it.Date)
			u := fmt.Sprintf("%s://%s/id/news/%s/%d/%s.html",
				baseParsed.Scheme, baseParsed.Host, strings.ReplaceAll(it.Date, "-", "/"), it.ID, url.PathEscape(it.Title))
			item := Item{
				Kind:         KindNews,
				Lang:         "id",
				SourceKey:    fmt.Sprintf("%d", it.ID),
				URL:          u,
				Title:        it.Title,
				Abstract:     cleanHTMLText(it.News),
				ReleaseDate:  &t,
				ThumbnailURL: it.Picture,
				ImageURL:     it.Picture,
				IsActive:     true,
			}
			item.ContentHash = CalculateContentHash(&item)
			items = append(items, item)
		}
		return items, nil

	case KindStaticTable:
		var raw []struct {
			ID         string `json:"id"`
			Title      string `json:"title"`
			Subject    string `json:"subject"`
			IDSubject  int    `json:"id_subject"`
			SubcatName string `json:"subcat_name"`
			LastUpdate string `json:"last_update"`
		}
		if err := json.Unmarshal(block, &raw); err != nil {
			return nil, err
		}
		var items []Item
		for _, it := range raw {
			cleanTitle := CleanText(it.Title)
			slug := strings.ToLower(cleanTitle)
			slug = strings.ReplaceAll(slug, "%", "-persen")
			var slugClean strings.Builder
			for _, r := range slug {
				if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
					slugClean.WriteRune(r)
				} else {
					slugClean.WriteRune('-')
				}
			}
			slugStr := strings.Trim(slugClean.String(), "-")

			u := fmt.Sprintf("%s://%s/id/statistics-table/%s/%s.html",
				baseParsed.Scheme, baseParsed.Host, it.ID, slugStr)

			var relDate *time.Time
			if t, err := time.Parse("2006-01-02 15:04:05", it.LastUpdate); err == nil {
				relDate = &t
			} else if t, err := time.Parse("2006-01-02", it.LastUpdate); err == nil {
				relDate = &t
			}

			item := Item{
				Kind:        KindStaticTable,
				Lang:        "id",
				SourceKey:   it.ID,
				URL:         u,
				Title:       cleanTitle,
				Subject:     it.Subject,
				ReleaseDate: relDate,
				IsActive:    true,
			}
			item.ContentHash = CalculateContentHash(&item)
			items = append(items, item)
		}
		return items, nil
	}
	return nil, errors.New("unsupported kind for server action")
}

func extractJSONBlock(resp string, prefix string) []byte {
	start := strings.Index(resp, prefix)
	if start == -1 {
		return nil
	}
	for end := len(resp); end > start+10; end-- {
		sub := resp[start:end]
		if sub[len(sub)-1] == ']' {
			var check []interface{}
			if err := json.Unmarshal([]byte(sub), &check); err == nil {
				return []byte(sub)
			}
		}
	}
	return nil
}

func cleanHTMLText(s string) string {
	s = strings.ReplaceAll(s, "<p>", "")
	s = strings.ReplaceAll(s, "</p>", "\n")
	s = strings.ReplaceAll(s, "<br>", "\n")
	s = strings.ReplaceAll(s, "<br/>", "\n")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "\r", "")
	return strings.TrimSpace(s)
}

func (r *Runner) sendAdminNotification(ctx context.Context, msg string) {
	if !r.cfg.NotifyAdmin || r.notifier == nil {
		return
	}
	_ = r.notifier.NotifyAdmin(ctx, msg)
}
