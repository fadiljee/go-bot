package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bot-istimewa/internal/api"
	"bot-istimewa/internal/bpsapi"
	"bot-istimewa/internal/bot"
	"bot-istimewa/internal/config"
	"bot-istimewa/internal/gsheet"
	"bot-istimewa/internal/locale"
	"bot-istimewa/internal/scraper"
	"bot-istimewa/internal/whatsapp"

	"go.mau.fi/whatsmeow/types/events"
)

func main() {
	log.Println("==================================================")
	log.Println("🤖 MEMULAI BOT ISTIMEWA v3.0 (Go Native Engine)")
	log.Println("🏛️  Badan Pusat Statistik Kabupaten Bangka Barat")
	log.Println("==================================================")

	// 1. Load Configurations
	cfg := config.LoadConfig()
	log.Printf("⚙️  Konfigurasi aktif | Port: %s | Tahun Indikator: %s",
		cfg.AppPort, cfg.ActiveIndicatorYear)

	// 2. Load Locales
	err := locale.Init("locales/id_ID.json")
	if err != nil {
		log.Fatalf("❌ Gagal memuat kamus lokalisasi id_ID: %v", err)
	}
	log.Println("📖 Kamus lokalisasi bahasa berhasil dimuat.")

	// 3. Initialize Google Sheets Client & Prewarm Cache
	gsheetClient := gsheet.NewClient(cfg)
	log.Println("🌐 Menginisialisasi koneksi data Google Sheets (GAS)...")
	go func() {
		gsheetClient.Prewarm()
		log.Println("✅ Data Google Sheets berhasil di-cache di memori.")
	}()

	// 4. Initialize WhatsApp Client Manager
	clientMgr, err := whatsapp.NewClientManager(cfg)
	if err != nil {
		log.Fatalf("❌ Gagal membuat WhatsApp Client Manager: %v", err)
	}

	// 4b. Initialize BPS Web API Client (Domain: 1903 Bangka Barat)
	bpsClient := bpsapi.NewClient(cfg)
	log.Printf("📊 BPS Web API Client aktif | Domain: %s (%s) | Base: %s",
		cfg.BPSDefaultDomain, "BPS Kabupaten Bangka Barat", cfg.BPSAPIBaseURL)

	// 5. Initialize Bot Message Router
	botRouter := bot.NewBotRouter(cfg, clientMgr.Sender, gsheetClient, bpsClient)

	// 5b. Initialize Scraper Components (§6, §10)
	var scraperStore *scraper.Store
	var scraperSched *scraper.Scheduler
	var scraperRunner *scraper.Runner
	var scraperBreaker *scraper.CircuitBreaker

	if cfg.ScraperEnabled {
		store, err := scraper.NewStore(cfg.ScraperDBPath)
		if err != nil {
			log.Printf("⚠️ Gagal menginisialisasi SQLite scraper store: %v", err)
		} else {
			scraperStore = store
			var enabledKinds []scraper.Kind
			for _, k := range cfg.ScraperEnabledKinds {
				enabledKinds = append(enabledKinds, scraper.Kind(k))
			}
			scraperCfg := &scraper.Config{
				Enabled:               cfg.ScraperEnabled,
				BaseURL:               cfg.ScraperBaseURL,
				AllowedHosts:          cfg.ScraperAllowedHosts,
				UserAgent:             cfg.ScraperUserAgent,
				DBPath:                cfg.ScraperDBPath,
				SelectorsPath:         cfg.ScraperSelectorsPath,
				CronSpec:              cfg.ScraperCronSpec,
				Timezone:              cfg.ScraperTimezone,
				RunTimeout:            time.Duration(cfg.ScraperRunTimeoutMinutes) * time.Minute,
				StartupCatchupHours:   cfg.ScraperStartupCatchupHours,
				ManualCooldownMinutes: cfg.ScraperManualCooldownMinutes,
				EnabledKinds:          enabledKinds,
				MinDelayMs:            cfg.ScraperMinDelayMs,
				JitterMs:              cfg.ScraperJitterMs,
				HTTPTimeoutSeconds:    cfg.ScraperHTTPTimeoutSeconds,
				MaxPageBytes:          cfg.ScraperMaxPageBytes,
				MaxListPages:          cfg.ScraperMaxListPages,
				MaxRequestsPerRun:     cfg.ScraperMaxRequestsPerRun,
				RespectRobots:         cfg.ScraperRespectRobots,
				MaxStaleHours:         cfg.ScraperMaxStaleHours,
				DataSourceOrder:       cfg.DataSourceOrder,
				NotifyAdmin:           cfg.ScraperNotifyAdmin,
				NotifyAfterFailures:   cfg.ScraperNotifyAfterFailures,
				AppEnv:                cfg.AppEnv,
			}
			if valErr := scraperCfg.Validate(); valErr != nil {
				log.Printf("⚠️ Validasi config scraper gagal: %v", valErr)
			} else {
				selectorsCfg, selErr := scraper.LoadSelectors(cfg.ScraperSelectorsPath, enabledKinds)
				if selErr != nil {
					log.Printf("⚠️ Gagal memuat selector scraper YAML: %v", selErr)
				} else {
					scraperBreaker = scraper.NewCircuitBreaker(scraperStore)
					robotsMgr := scraper.NewRobotsManager(&http.Client{Timeout: 10 * time.Second}, cfg.ScraperBaseURL, cfg.ScraperUserAgent, cfg.ScraperRespectRobots)
					fetcher := scraper.NewFetcher(scraperCfg, robotsMgr)
					scraperRunner = scraper.NewRunner(scraperCfg, scraperStore, fetcher, scraperBreaker, selectorsCfg, nil)
					provider := scraper.NewProvider(scraperStore, cfg.ScraperMaxStaleHours)
					scraperRunner.SetProvider(provider)
					scraperSched = scraper.NewScheduler(scraperCfg, scraperRunner, scraperStore)
					if err := scraperSched.Start(); err != nil {
						log.Printf("⚠️ Gagal menjalankan scheduler scraper: %v", err)
					}
					// Connect scraper provider to bot router
					botRouter.SetScraper(provider)
					log.Printf("🕸️ [SCRAPER] Scraper otomatis BPS Bangka Barat aktif (Cron: %s | DB: %s)",
						cfg.ScraperCronSpec, cfg.ScraperDBPath)
				}
			}
		}
	}

	botStartTime := time.Now()

	// 6. Register Event Listener
	clientMgr.Client.AddEventHandler(func(rawEvt interface{}) {
		switch evt := rawEvt.(type) {
		case *events.Message:
			// Log Group JID to console so developers can easily copy-paste it to .env
			if evt.Info.IsGroup {
				log.Printf("👥 [INFO GRUP WA] Pesan masuk di Grup JID: %s (Pengirim: %s)", evt.Info.Chat.String(), evt.Info.Sender.User)
				return
			}

			// Strict Source Filtering (Lapis 2): ignore self, status/broadcasts
			if evt.Info.IsFromMe || evt.Info.Chat.Server == "broadcast" || evt.Info.Chat.Server == "newsletter" {
				return
			}
			// Historical message filter: ignore messages sent before bot startup
			if evt.Info.Timestamp.Before(botStartTime) {
				return
			}

			parsed := whatsapp.ParseIncomingMessage(evt)
			if parsed != nil {
				go botRouter.Dispatch(context.Background(), parsed)
			}

		case *events.Connected:
			clientMgr.IsLoggedIn = true
			log.Println("📶 [WHATSAPP] Terhubung ke server WhatsApp.")

			// Auto-discover joined groups on connect (helpful for developer to find INTERNAL_PST_GROUP_JID)
			go func() {
				time.Sleep(3 * time.Second)
				groups, err := clientMgr.Client.GetJoinedGroups(context.Background())
				if err == nil && len(groups) > 0 {
					log.Printf("📋 [DAFTAR GRUP WA BOT (%d grup)]:", len(groups))
					for i, g := range groups {
						log.Printf("   %d. 🏷️ Nama: '%s' ➡️ JID: %s", i+1, g.GroupName.Name, g.JID.String())
					}
					log.Println("💡 [TIPS] Copy salah satu JID di atas dan masukkan ke .env (INTERNAL_PST_GROUP_JID=...)")
				}
			}()

		case *events.LoggedOut:
			clientMgr.IsLoggedIn = false
			log.Println("⚠️ [WHATSAPP] Sesi telah keluar (Logged Out). Harap pindai ulang QR Code.")

		case *events.StreamReplaced:
			log.Println("⚠️ [WHATSAPP] Sesi ditimpa oleh koneksi lain.")
		}
	})

	// 7. Start WhatsApp Connection
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := clientMgr.Start(ctx); err != nil {
		log.Printf("⚠️ Gagal memulai koneksi awal WhatsApp: %v", err)
	}

	// 8. Start Fiber REST API Server
	apiServer := api.NewAPIServer(cfg, clientMgr, gsheetClient, bpsClient)
	if scraperRunner != nil && scraperStore != nil {
		apiServer.SetScraper(scraperRunner, scraperStore, scraperSched, scraperBreaker)
	}
	go func() {
		addr := fmt.Sprintf("%s:%s", cfg.AppHost, cfg.AppPort)
		log.Printf("🚀 Server Web Dashboard aktif di http://%s", addr)
		log.Printf("📲 Akses http://%s/qr untuk memindai QR code via browser.", addr)
		if err := apiServer.Listen(addr); err != nil {
			log.Printf("ℹ️ Web server berhenti: %v", err)
		}
	}()

	// 9. Wait for Shutdown Signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("\n🛑 Menerima sinyal terminasi, mematikan sistem secara aman...")

	// Graceful shutdown with 5s timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if scraperSched != nil {
		scraperSched.Stop(shutdownCtx)
	}
	if scraperStore != nil {
		_ = scraperStore.Close()
	}

	if err := apiServer.Shutdown(); err != nil {
		log.Printf("⚠️ Gagal mematikan API server: %v", err)
	}

	clientMgr.Client.Disconnect()
	log.Println("🔌 WhatsApp client terputus.")

	select {
	case <-shutdownCtx.Done():
		log.Println("⏱️ Graceful shutdown timeout.")
	default:
		log.Println("👋 BOT ISTIMEWA v3.1 berhasil dimatikan dengan bersih.")
	}
}
