package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"bot-istimewa/internal/config"
	"bot-istimewa/internal/scraper"
)

func main() {
	kindFlag := flag.String("kind", "all", "Kind to scrape: all, pressrelease, publication, infographic, statictable, news")
	flag.Parse()

	cfg := config.LoadConfig()

	store, err := scraper.NewStore(cfg.ScraperDBPath)
	if err != nil {
		log.Fatalf("Error opening scraper store: %v", err)
	}
	defer store.Close()

	var enabledKinds []scraper.Kind
	if *kindFlag == "all" {
		for _, k := range cfg.ScraperEnabledKinds {
			enabledKinds = append(enabledKinds, scraper.Kind(k))
		}
	} else {
		enabledKinds = append(enabledKinds, scraper.Kind(*kindFlag))
	}

	selectorsCfg, err := scraper.LoadSelectors(cfg.ScraperSelectorsPath, enabledKinds)
	if err != nil {
		log.Fatalf("Error loading selectors: %v", err)
	}

	scraperCfg := &scraper.Config{
		Enabled:               true,
		BaseURL:               cfg.ScraperBaseURL,
		AllowedHosts:          cfg.ScraperAllowedHosts,
		UserAgent:             cfg.ScraperUserAgent,
		DBPath:                cfg.ScraperDBPath,
		SelectorsPath:         cfg.ScraperSelectorsPath,
		CronSpec:              cfg.ScraperCronSpec,
		Timezone:              cfg.ScraperTimezone,
		RunTimeout:            15 * time.Minute,
		EnabledKinds:          enabledKinds,
		MinDelayMs:            cfg.ScraperMinDelayMs,
		JitterMs:              cfg.ScraperJitterMs,
		HTTPTimeoutSeconds:    cfg.ScraperHTTPTimeoutSeconds,
		MaxPageBytes:          cfg.ScraperMaxPageBytes,
		MaxListPages:          cfg.ScraperMaxListPages,
		MaxRequestsPerRun:     cfg.ScraperMaxRequestsPerRun,
		RespectRobots:         cfg.ScraperRespectRobots,
		AppEnv:                cfg.AppEnv,
	}

	breaker := scraper.NewCircuitBreaker(store)
	robotsMgr := scraper.NewRobotsManager(&http.Client{Timeout: 10 * time.Second}, cfg.ScraperBaseURL, cfg.ScraperUserAgent, cfg.ScraperRespectRobots)
	fetcher := scraper.NewFetcher(scraperCfg, robotsMgr)
	runner := scraper.NewRunner(scraperCfg, store, fetcher, breaker, selectorsCfg, nil)

	ctx := context.Background()
	log.Printf("🚀 [SCRAPER-CLI] Memulai live scraping (kind=%s, BaseURL=%s)...", *kindFlag, cfg.ScraperBaseURL)

	res, err := runner.RunAll(ctx, scraper.TriggerManual)
	if err != nil {
		log.Printf("❌ Scraping RunAll selesai dengan error: %v", err)
	} else {
		log.Printf("✅ Scraping RunAll selesai: status=%s, reqs=%d, found=%d, new=%d, updated=%d, skipped=%d",
			res.Status, res.Requests, res.ItemsFound, res.ItemsNew, res.ItemsUpdated, res.ItemsSkipped)
	}

	counts, _ := store.GetCountsByKind(ctx)
	fmt.Printf("\n📊 [DATABASE HASIL SCRAPING]:\n")
	if len(counts) == 0 {
		fmt.Println(" (Database masih kosong / 0 items aktif)")
	}
	for k, count := range counts {
		fmt.Printf(" - %-15s: %d items\n", k, count)
	}

	runs, _ := store.GetRuns(ctx, 5)
	fmt.Printf("\n📜 [RIWAYAT SCRAPE RUNS TERBARU]:\n")
	for _, r := range runs {
		var dur string
		if r.FinishedAt != nil {
			dur = r.FinishedAt.Sub(r.StartedAt).Round(time.Millisecond).String()
		}
		fmt.Printf(" - Run #%d: status=%s trigger=%s found=%d new=%d dur=%s err=%q\n",
			r.ID, r.Status, r.Trigger, r.ItemsFound, r.ItemsNew, dur, r.ErrorMessage)
	}
}
