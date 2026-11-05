package scraper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
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
	cfg       *Config
	store     *Store
	fetcher   *Fetcher
	breaker   *CircuitBreaker
	selectors *SelectorsConfig
	parsers   map[Kind]Parser
	notifier  AdminNotifier
	provider  *Provider

	mu         sync.Mutex
	lastManual time.Time
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
			_ = r.store.FinishRun(context.WithoutCancel(ctx), run)
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
			_ = r.store.FinishRun(context.WithoutCancel(ctx), run)
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
			switch {
			case errors.Is(err, ErrBlocked) || errors.Is(err, ErrRateLimited):
				cooldown := r.breaker.Trip(ctx, err.Error())
				run.Status = "blocked"
				if errors.Is(err, ErrRateLimited) {
					run.Status = "rate_limited"
				}
				run.ErrorMessage = err.Error()
				r.sendAdminNotification(ctx, fmt.Sprintf("Scraper BPS Bangka Barat: %v. Cooldown %v. Bot beralih ke data cadangan.", err, cooldown))
				allKindsSuccess = false
			case errors.Is(err, ErrAnomalyLayoutChanged) || errors.Is(err, ErrAnomalyCountDrop) || errors.Is(err, ErrAnomalyMarkerMissing):
				hasPartial = true
				allKindsSuccess = false
				r.sendAdminNotification(ctx, fmt.Sprintf("Scraper BPS Bangka Barat anomali: %v. Data lama tetap dipertahankan.", err))
				continue
			default:
				hasPartial = true
				allKindsSuccess = false
				continue
			}
			break // Halt cycle immediately (§10.4)
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
		r.breaker.Reset(ctx)
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
	funcName    string
	arrayPrefix string
}

var (
	serverActionsMu sync.RWMutex
	serverActions   = map[Kind]serverActionDef{
		KindPressRelease: {
			path:        "/id/pressrelease",
			actionID:    "608f5e41e89f2fc560866fb094a63610f7cca1b54d",
			funcName:    "getListPressrelease",
			arrayPrefix: `[{"brs_id":`,
		},
		KindPublication: {
			path:        "/id/publication",
			actionID:    "6064c6e4dcc79c5ddc0bc6a4ecb0d75a4b9cc67152",
			funcName:    "getListPublication",
			arrayPrefix: `[{"pub_id":`,
		},
		KindNews: {
			path:        "/id/news",
			actionID:    "60dde7c73a6806a8d58be2b5aee1fb05f1fbd30851",
			funcName:    "getListNews",
			arrayPrefix: `[{"news_id":`,
		},
		KindStaticTable: {
			path:        "/id/statistics-table",
			actionID:    "40b4dff2adfeae98b2f67226e3b8d32fd9fedfa6de",
			funcName:    "getListTable",
			arrayPrefix: `[{"id":`,
		},
		KindInfographic: {
			path:        "/id/infographic",
			actionID:    "70c04b524c74e17902c74fbcef5075443a6ac33c49",
			funcName:    "getListInfografis",
			arrayPrefix: `[{"inf_id":`,
		},
		KindARC: {
			path:        "/id/arc",
			actionID:    "60b0a7e5b2c5b76cf70945849c082abfb9f345e84f",
			funcName:    "getArcBrs",
			arrayPrefix: `[{"`,
		},
	}
	arcPubActionID = "7059cffa2cfc3db9b02729cb1906e796b5e9e9d239"
	reNextChunk    = regexp.MustCompile(`/_next/static/chunks/[^"'>\s]+\.js`)
)

func (r *Runner) discoverActionIDFromHTML(ctx context.Context, htmlBody []byte, baseParsed *url.URL, funcName string) string {
	if len(htmlBody) == 0 || funcName == "" {
		return ""
	}
	reRef := regexp.MustCompile(`createServerReference\)\("([0-9a-f]{40,44})"[^)]*"` + regexp.QuoteMeta(funcName) + `"\)`)
	chunks := reNextChunk.FindAllString(string(htmlBody), -1)
	seen := make(map[string]bool, len(chunks))
	for _, chunkPath := range chunks {
		if seen[chunkPath] {
			continue
		}
		seen[chunkPath] = true
		chunkURL := fmt.Sprintf("%s://%s%s", baseParsed.Scheme, baseParsed.Host, chunkPath)
		jsBody, err := r.fetcher.GetAsset(ctx, chunkURL)
		if err != nil {
			continue
		}
		if m := reRef.FindSubmatch(jsBody); len(m) == 2 {
			discovered := string(m[1])
			slog.Info("scraper.discovered_action_id", "func", funcName, "action_id", discovered)
			return discovered
		}
	}
	return ""
}

func (r *Runner) postActionWithDiscovery(
	ctx context.Context,
	kind Kind,
	pageURL string,
	htmlBody []byte,
	baseParsed *url.URL,
	actionID string,
	funcName string,
	payload string,
) ([]byte, string, error) {
	respBody, err := r.fetcher.PostAction(ctx, pageURL, actionID, payload)
	if err == nil && !strings.Contains(string(respBody), "Server Action unavailable") {
		return respBody, actionID, nil
	}

	discovered := r.discoverActionIDFromHTML(ctx, htmlBody, baseParsed, funcName)
	if discovered == "" || discovered == actionID {
		if err != nil {
			return nil, actionID, err
		}
		return respBody, actionID, errors.New("server action unavailable")
	}

	serverActionsMu.Lock()
	if def, ok := serverActions[kind]; ok && def.funcName == funcName {
		def.actionID = discovered
		serverActions[kind] = def
	}
	if funcName == "getArcPublikasi" {
		arcPubActionID = discovered
	}
	serverActionsMu.Unlock()

	respBody, err = r.fetcher.PostAction(ctx, pageURL, discovered, payload)
	return respBody, discovered, err
}

func (r *Runner) scrapeServerAction(ctx context.Context, kind Kind, baseParsed *url.URL) ([]Item, int, error) {
	serverActionsMu.RLock()
	def, ok := serverActions[kind]
	serverActionsMu.RUnlock()
	if !ok {
		return nil, 0, errors.New("no server action defined for kind")
	}

	pageURL := fmt.Sprintf("%s://%s%s", baseParsed.Scheme, baseParsed.Host, def.path)
	// Initial GET to acquire session cookies and HTML for action ID discovery if needed
	initResp, _ := r.fetcher.Get(ctx, pageURL, nil)
	var htmlBody []byte
	if initResp != nil {
		htmlBody = initResp.Body
	}

	if kind == KindARC {
		return r.scrapeARC(ctx, pageURL, htmlBody, baseParsed, def)
	}

	var allItems []Item
	skippedCount := 0

	maxPages := r.cfg.MaxListPages
	if maxPages <= 0 {
		maxPages = 2
	}

	actionID := def.actionID

	if kind == KindInfographic {
		categories := []struct {
			id   int
			name string
		}{
			{id: 1, name: "Statistik Demografi dan Sosial"},
			{id: 2, name: "Statistik Ekonomi"},
			{id: 3, name: "Statistik Lingkungan Hidup dan Multidomain"},
		}

		for _, cat := range categories {
			for page := 1; page <= maxPages; page++ {
				payload := fmt.Sprintf(`["id",%d,%d]`, page, cat.id)
				respBody, updatedID, err := r.postActionWithDiscovery(ctx, kind, pageURL, htmlBody, baseParsed, actionID, def.funcName, payload)
				actionID = updatedID
				if err != nil {
					break
				}

				jsonBlock := extractJSONBlock(string(respBody), def.arrayPrefix)
				if jsonBlock == nil {
					break
				}

				pageItems, err := r.parseInfographicJSON(jsonBlock, baseParsed, cat.id, cat.name)
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

				if len(pageItems) < 10 {
					break
				}
			}
		}

		return allItems, skippedCount, nil
	}

	for page := 1; page <= maxPages; page++ {
		payload := fmt.Sprintf(`["id",{"page":%d}]`, page)
		if kind == KindStaticTable {
			payload = fmt.Sprintf(`[{"locale":"id","page":%d,"perpage":20}]`, page)
		}
		respBody, updatedID, err := r.postActionWithDiscovery(ctx, kind, pageURL, htmlBody, baseParsed, actionID, def.funcName, payload)
		actionID = updatedID
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

		if len(pageItems) < 10 {
			break
		}
	}

	return allItems, skippedCount, nil
}

func (r *Runner) scrapeARC(
	ctx context.Context,
	pageURL string,
	htmlBody []byte,
	baseParsed *url.URL,
	def serverActionDef,
) ([]Item, int, error) {
	var allItems []Item
	skippedCount := 0
	year := fmt.Sprintf("%d", time.Now().Year())

	// 1. Scrape BRS Release Schedule (getArcBrs)
	brsPayload := fmt.Sprintf(`["id",%q]`, year)
	brsBody, _, err := r.postActionWithDiscovery(ctx, KindARC, pageURL, htmlBody, baseParsed, def.actionID, def.funcName, brsPayload)
	if err == nil {
		if block := extractJSONBlock(string(brsBody), `[{"`); block != nil {
			if items, parseErr := r.parseARCJSON(block, baseParsed, year, "Berita Resmi Statistik (BRS)"); parseErr == nil {
				for i := range items {
					it := &items[i]
					if valErr := ValidateItem(it, r.cfg.AllowedHosts, r.cfg.AppEnv); valErr != nil {
						skippedCount++
						continue
					}
					allItems = append(allItems, *it)
				}
			}
		}
	}

	// 2. Scrape Publication Release Schedule (getArcPublikasi) across months 01..12
	serverActionsMu.RLock()
	pubActionID := arcPubActionID
	serverActionsMu.RUnlock()

	months := []string{"01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12"}
	for _, m := range months {
		pubPayload := fmt.Sprintf(`["id",%q,%q]`, year, m)
		pubBody, updatedID, pErr := r.postActionWithDiscovery(ctx, KindARC, pageURL, htmlBody, baseParsed, pubActionID, "getArcPublikasi", pubPayload)
		pubActionID = updatedID
		if pErr != nil {
			continue
		}
		if block := extractJSONBlock(string(pubBody), `[{"`); block != nil {
			if items, parseErr := r.parseARCJSON(block, baseParsed, year, "Publikasi Statistik"); parseErr == nil {
				for i := range items {
					it := &items[i]
					if valErr := ValidateItem(it, r.cfg.AllowedHosts, r.cfg.AppEnv); valErr != nil {
						skippedCount++
						continue
					}
					allItems = append(allItems, *it)
				}
			}
		}
	}

	return allItems, skippedCount, nil
}

func (r *Runner) parseARCJSON(block []byte, baseParsed *url.URL, year, categoryLabel string) ([]Item, error) {
	var raw []struct {
		ID      string `json:"id"`
		Title   string `json:"title"`
		Date    string `json:"date"`
		Status  string `json:"status"`
		Type    string `json:"type"`
		URL     string `json:"url"`
		Periode string `json:"periode"`
	}
	if err := json.Unmarshal(block, &raw); err != nil {
		return nil, err
	}

	var items []Item
	seen := make(map[string]bool, len(raw))
	for idx, it := range raw {
		cleanTitle := CleanHTML(it.Title)
		if cleanTitle == "" {
			continue
		}
		var relDate *time.Time
		if t, err := time.Parse("2006-01-02", strings.TrimSpace(it.Date)); err == nil {
			relDate = &t
		}
		rawID := strings.TrimSpace(it.ID)
		if rawID == "" {
			rawID = fmt.Sprintf("%s_%d", it.Date, idx)
		}
		typePrefix := strings.ToLower(strings.TrimSpace(it.Type))
		if typePrefix == "" {
			typePrefix = "arc"
		}
		sourceKey := fmt.Sprintf("%s_%s_%s", typePrefix, year, rawID)
		if seen[sourceKey] {
			continue
		}
		seen[sourceKey] = true

		itemURL := strings.TrimSpace(it.URL)
		switch {
		case strings.HasPrefix(itemURL, "http://") || strings.HasPrefix(itemURL, "https://"):
		case strings.HasPrefix(itemURL, "/"):
			itemURL = fmt.Sprintf("%s://%s%s", baseParsed.Scheme, baseParsed.Host, itemURL)
		default:
			itemURL = fmt.Sprintf("%s://%s/id/arc?id=%s", baseParsed.Scheme, baseParsed.Host, url.QueryEscape(sourceKey))
		}

		statusText := strings.TrimSpace(it.Status)
		if statusText == "" {
			statusText = "Terjadwal"
		}
		if strings.TrimSpace(it.Periode) != "" {
			statusText = fmt.Sprintf("%s (%s)", statusText, strings.TrimSpace(it.Periode))
		}

		item := Item{
			Kind:        KindARC,
			Lang:        "id",
			SourceKey:   sourceKey,
			URL:         itemURL,
			Title:       cleanTitle,
			Abstract:    statusText,
			ReleaseDate: relDate,
			Subject:     categoryLabel,
			IsActive:    true,
		}
		item.ContentHash = CalculateContentHash(&item)
		items = append(items, item)
	}
	return items, nil
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

func (r *Runner) parseInfographicJSON(block []byte, baseParsed *url.URL, catID int, catName string) ([]Item, error) {
	var raw []struct {
		ID       int     `json:"inf_id"`
		Title    string  `json:"title"`
		Desc     string  `json:"desc"`
		Category int     `json:"category"`
		Date     *string `json:"date"`
		Img      string  `json:"img"`
		DL       string  `json:"dl"`
	}
	if err := json.Unmarshal(block, &raw); err != nil {
		return nil, err
	}
	var items []Item
	for _, it := range raw {
		var relDate *time.Time
		if it.Date != nil && strings.TrimSpace(*it.Date) != "" {
			if t, err := time.Parse("2006-01-02", strings.TrimSpace(*it.Date)); err == nil {
				relDate = &t
			}
		}
		itemURL := fmt.Sprintf("%s://%s/id/infographic?id=%d", baseParsed.Scheme, baseParsed.Host, it.ID)
		item := Item{
			Kind:         KindInfographic,
			Lang:         "id",
			SourceKey:    fmt.Sprintf("cat%d_%d", catID, it.ID),
			URL:          itemURL,
			Title:        it.Title,
			Abstract:     cleanHTMLText(it.Desc),
			ReleaseDate:  relDate,
			FileURL:      it.DL,
			ThumbnailURL: it.Img,
			ImageURL:     it.Img,
			Subject:      catName,
			IsActive:     true,
		}
		item.ContentHash = CalculateContentHash(&item)
		items = append(items, item)
	}
	return items, nil
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
	return CleanHTML(s)
}

func (r *Runner) sendAdminNotification(ctx context.Context, msg string) {
	if !r.cfg.NotifyAdmin || r.notifier == nil {
		return
	}
	_ = r.notifier.NotifyAdmin(ctx, msg)
}
