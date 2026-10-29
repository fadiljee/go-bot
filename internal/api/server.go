package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"

	"bot-istimewa/internal/bpsapi"
	"bot-istimewa/internal/config"
	"bot-istimewa/internal/gsheet"
	"bot-istimewa/internal/scraper"
	"bot-istimewa/internal/whatsapp"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

type APIServer struct {
	app              *fiber.App
	cfg              *config.Config
	clientMgr        *whatsapp.ClientManager
	gsheet           *gsheet.Client
	bps              *bpsapi.Client
	scraperRunner    *scraper.Runner
	scraperStore     *scraper.Store
	scraperScheduler *scraper.Scheduler
	scraperBreaker   *scraper.CircuitBreaker
	scraperProvider  scraper.DataProvider
	startTime        time.Time
}

func NewAPIServer(cfg *config.Config, clientMgr *whatsapp.ClientManager, gsheetClient *gsheet.Client, bpsClient ...*bpsapi.Client) *APIServer {
	app := fiber.New(fiber.Config{
		AppName:               "BOT ISTIMEWA v3.0 (Go Edition)",
		DisableStartupMessage: true,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	var bps *bpsapi.Client
	if len(bpsClient) > 0 && bpsClient[0] != nil {
		bps = bpsClient[0]
	} else {
		bps = bpsapi.NewClient(cfg)
	}

	srv := &APIServer{
		app:       app,
		cfg:       cfg,
		clientMgr: clientMgr,
		gsheet:    gsheetClient,
		bps:       bps,
		startTime: time.Now(),
	}

	srv.setupRoutes()
	return srv
}

func (s *APIServer) setupRoutes() {
	// Public Health Check
	s.app.Get("/health", s.handleHealth)

	// Web Browser QR Scanner (Auto-Locks when logged in)
	s.app.Get("/qr", s.handleQRPage)

	// API Group (Protected by X-API-Key / Bearer token)
	api := s.app.Group("/api", s.authMiddleware)

	// Legacy & V1 Cache Invalidation endpoints
	api.Post("/cache/refresh", s.handleCacheRefresh)
	api.Post("/v1/cache/clear", s.handleCacheRefresh)

	// Legacy & V1 Asynchronous Blast endpoints
	api.Post("/blast", s.handleBlast)
	api.Post("/v1/blast", s.handleBlast)

	// Protected API QR endpoint
	api.Get("/v1/qr", s.handleQRAPI)

	// BPS Web API Route Group (Protected by X-API-Key)
	bpsGroup := api.Group("/v1/bps")
	bpsGroup.Get("/list", s.handleBPSList)
	bpsGroup.Get("/subjects", s.handleBPSSubjects)
	bpsGroup.Get("/subcats", s.handleBPSSubcats)
	bpsGroup.Get("/variables", s.handleBPSVariables)
	bpsGroup.Get("/statictables", s.handleBPSStaticTables)
	bpsGroup.Get("/publications", s.handleBPSPublications)
	bpsGroup.Get("/pressreleases", s.handleBPSPressReleases)
	bpsGroup.Get("/infographics", s.handleBPSInfographics)
	bpsGroup.Get("/data", s.handleBPSData)
	bpsGroup.Post("/cache/clear", s.handleBPSCacheClear)

	// Scraper Route Group (§12)
	scraperGroup := api.Group("/v1/scraper")
	scraperGroup.Get("/status", s.handleScraperStatus)
	scraperGroup.Post("/run", s.handleScraperRun)
	scraperGroup.Get("/runs", s.handleScraperRuns)
	scraperGroup.Get("/items", s.handleScraperItems)
	scraperGroup.Post("/breaker/reset", s.handleScraperBreakerReset)

	// Strategic Indicators Route Group (Menu 2)
	indGroup := api.Group("/v1/indicators")
	indGroup.Get("/", s.handleGetIndicators)
	indGroup.Get("/:id", s.handleGetIndicatorByID)

	// Alias for Indonesian endpoint: /api/v1/indikator
	aliasGroup := api.Group("/v1/indikator")
	aliasGroup.Get("/", s.handleGetIndicators)
	aliasGroup.Get("/:id", s.handleGetIndicatorByID)

	// Layanan Data Route Group (Menu 1: Sosduk, Ekonomi, Pertanian, Sektoral)
	layananGroup := api.Group("/v1/layanan-data")
	layananGroup.Get("/", s.handleGetLayananData)
	layananGroup.Get("/:category", s.handleGetLayananCategory)
	layananGroup.Get("/:category/:topic", s.handleGetLayananTopic)

	// Alias for English endpoint: /api/v1/services/data
	servicesGroup := api.Group("/v1/services/data")
	servicesGroup.Get("/", s.handleGetLayananData)
	servicesGroup.Get("/:category", s.handleGetLayananCategory)
	servicesGroup.Get("/:category/:topic", s.handleGetLayananTopic)
}

func (s *APIServer) SetScraper(runner *scraper.Runner, store *scraper.Store, scheduler *scraper.Scheduler, breaker *scraper.CircuitBreaker, provider ...scraper.DataProvider) {
	s.scraperRunner = runner
	s.scraperStore = store
	s.scraperScheduler = scheduler
	s.scraperBreaker = breaker
	if len(provider) > 0 && provider[0] != nil {
		s.scraperProvider = provider[0]
	}
}

func (s *APIServer) getProvider() scraper.DataProvider {
	if s.scraperProvider != nil {
		return s.scraperProvider
	}
	if s.scraperStore != nil {
		s.scraperProvider = scraper.NewProvider(s.scraperStore, s.cfg.ScraperMaxStaleHours)
		return s.scraperProvider
	}
	return nil
}

func (s *APIServer) authMiddleware(c *fiber.Ctx) error {
	apiKey := c.Get("X-API-Key")
	if apiKey == "" {
		authHeader := c.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			apiKey = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}
	if apiKey == "" {
		apiKey = c.Query("key")
	}

	if apiKey != s.cfg.APISecretKey {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized: Invalid or missing API key",
		})
	}

	return c.Next()
}

func (s *APIServer) handleHealth(c *fiber.Ctx) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	var lastAgeHours float64
	var breakerOpen bool
	if s.scraperStore != nil {
		lastAgeHours, _ = s.scraperStore.GetDataAgeHours(context.Background())
	}
	if s.scraperBreaker != nil {
		breakerOpen, _ = s.scraperBreaker.IsOpen()
	}

	return c.JSON(fiber.Map{
		"app":       "BOT ISTIMEWA",
		"version":   "3.1.0 (Go Native Hardened)",
		"agency":    "BPS Kabupaten Bangka Barat",
		"connected": s.clientMgr.IsLoggedIn,
		"uptime":    time.Since(s.startTime).String(),
		"memory": fiber.Map{
			"alloc_mb":       fmt.Sprintf("%.2f MB", float64(m.Alloc)/(1024*1024)),
			"sys_mb":         fmt.Sprintf("%.2f MB", float64(m.Sys)/(1024*1024)),
			"num_goroutines": runtime.NumGoroutine(),
		},
		"scraper_last_success_age_hours": lastAgeHours,
		"scraper_breaker_open":           breakerOpen,
	})
}

// handleQRPage renders HTML QR page or returns 403 Forbidden if already logged in (Auto-Lock)
func (s *APIServer) handleQRPage(c *fiber.Ctx) error {
	if s.clientMgr.IsLoggedIn {
		return c.Status(fiber.StatusForbidden).Type("html").SendString(`
			<!DOCTYPE html>
			<html>
			<head><title>403 Forbidden - BOT ISTIMEWA</title><meta name="viewport" content="width=device-width, initial-scale=1"></head>
			<body style="font-family:sans-serif; text-align:center; padding: 40px; background:#f4f6f9;">
				<div style="max-width:450px; margin:auto; background:white; padding:30px; border-radius:12px; box-shadow:0 4px 10px rgba(0,0,0,0.1);">
					<h2 style="color:#d32f2f;">Akses QR Terkunci (403 Forbidden)</h2>
					<p>Bot WhatsApp sudah berhasil terhubung dan sesi aktif. Akses pemindaian QR ditutup demi keamanan sistem.</p>
				</div>
			</body>
			</html>
		`)
	}

	qrPNG, _ := s.clientMgr.GetQR()
	if len(qrPNG) == 0 {
		return c.Type("html").SendString(`
			<!DOCTYPE html>
			<html>
			<head><title>BOT ISTIMEWA - QR Code</title><meta http-equiv="refresh" content="3"><meta name="viewport" content="width=device-width, initial-scale=1"></head>
			<body style="font-family:sans-serif; text-align:center; padding: 40px; background:#f4f6f9;">
				<div style="max-width:450px; margin:auto; background:white; padding:30px; border-radius:12px; box-shadow:0 4px 10px rgba(0,0,0,0.1);">
					<h2> Menunggu QR Code WhatsApp...</h2>
					<p>Halaman ini akan me-refresh otomatis setiap 3 detik.</p>
				</div>
			</body>
			</html>
		`)
	}

	base64PNG := base64.StdEncoding.EncodeToString(qrPNG)
	html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>Pindai QR Code - BOT ISTIMEWA</title>
			<meta http-equiv="refresh" content="5">
			<meta name="viewport" content="width=device-width, initial-scale=1">
			<style>
				body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; text-align: center; background: #f0f2f5; padding: 20px; }
				.card { max-width: 420px; margin: 40px auto; background: white; padding: 30px; border-radius: 16px; box-shadow: 0 4px 20px rgba(0,0,0,0.08); }
				img { width: 260px; height: 260px; border-radius: 8px; border: 1px solid #e1e4e8; }
				h2 { color: #1f2937; margin-top: 0; }
				p { color: #4b5563; font-size: 14px; line-height: 1.5; }
				.badge { background: #e0f2fe; color: #0369a1; padding: 4px 12px; border-radius: 20px; font-weight: 600; font-size: 12px; display: inline-block; margin-bottom: 12px; }
			</style>
		</head>
		<body>
			<div class="card">
				<span class="badge">BPS KABUPATEN BANGKA BARAT</span>
				<h2>Pindai QR Code WhatsApp</h2>
				<p>Buka WhatsApp di ponsel bot &rarr; Perangkat Tertaut &rarr; Tautkan Perangkat, lalu scan QR di bawah ini:</p>
				<img src="data:image/png;base64,%s" alt="QR Code WhatsApp" />
				<p><small>Halaman otomatis diperbarui setiap 5 detik.</small></p>
			</div>
		</body>
		</html>
	`, base64PNG)

	return c.Type("html").SendString(html)
}

func (s *APIServer) handleQRAPI(c *fiber.Ctx) error {
	if s.clientMgr.IsLoggedIn {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "Forbidden: WhatsApp client is already logged in",
		})
	}

	qrPNG, qrStr := s.clientMgr.GetQR()
	if len(qrPNG) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "QR code is not currently generated or waiting for scan",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"qr_raw":  qrStr,
		"qr_png":  base64.StdEncoding.EncodeToString(qrPNG),
	})
}

func (s *APIServer) handleCacheRefresh(c *fiber.Ctx) error {
	log.Println("[API] Permintaan pembersihan dan pemutakhiran cache Google Sheets diterima...")
	err := s.gsheet.RefreshAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Seluruh cache Google Sheets berhasil diperbarui.",
	})
}

type BlastRequest struct {
	Numbers []string `json:"numbers"`
	Message string   `json:"message"`
}

func (s *APIServer) handleBlast(c *fiber.Ctx) error {
	var req BlastRequest
	if err := c.BodyParser(&req); err != nil || len(req.Numbers) == 0 || req.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Payload tidak valid. Pastikan 'numbers' (array) dan 'message' terisi.",
		})
	}

	// Build blast message with Opt-Out quick reply button to prevent report/spam ban
	optOutButtons := []whatsapp.QuickReplyButton{
		{ID: "berhenti_langganan", DisplayText: "Berhenti Berlangganan"},
	}
	blastMsg := whatsapp.BuildQuickReplyButtons("Rilis Data BPS Bangka Barat", req.Message, "PST BPS Kabupaten Bangka Barat", optOutButtons)

	extra := &whatsmeow.SendRequestExtra{
		AdditionalNodes: &[]waBinary.Node{
			{
				Tag: "biz",
				Content: []waBinary.Node{
					{
						Tag: "interactive",
						Attrs: waBinary.Attrs{"type": "native_flow", "v": "1"},
						Content: []waBinary.Node{
							{
								Tag: "native_flow",
								Attrs: waBinary.Attrs{"v": "9", "name": "mixed"},
							},
						},
					},
				},
			},
		},
	}

	// Dispatch to low-priority BlastQueue
	go func(numbers []string) {
		log.Printf("[BLAST QUEUE] Memasukkan %d nomor ke antrean BlastQueue (delay 5-10 detik)...", len(numbers))
		for _, num := range numbers {
			cleanNum := strings.TrimSpace(num)
			cleanNum = strings.TrimPrefix(cleanNum, "+")
			if strings.HasPrefix(cleanNum, "08") {
				cleanNum = "62" + cleanNum[1:]
			}

			jid := types.NewJID(cleanNum, types.DefaultUserServer)
			if s.clientMgr.Sender.Queue() != nil {
				s.clientMgr.Sender.Queue().EnqueueBlast(jid, blastMsg, extra)
			} else {
				_ = s.clientMgr.Sender.SendInteractiveMessage(context.Background(), jid, blastMsg)
			}
		}
		log.Println("[BLAST QUEUE] Seluruh nomor telah dijadwalkan di antrean blast.")
	}(req.Numbers)

	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Pengiriman blast ke %d nomor sedang diproses di antrean latar belakang dengan tombol opt-out.", len(req.Numbers)),
	})
}

func (s *APIServer) extractBPSParams(c *fiber.Ctx, defaultModel string) bpsapi.ListParams {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	model := c.Query("model", defaultModel)
	return bpsapi.ListParams{
		Model:   model,
		Domain:  c.Query("domain", ""),
		Lang:    c.Query("lang", ""),
		Page:    page,
		Subject: c.Query("subject", ""),
		Subcat:  c.Query("subcat", ""),
		Year:    c.Query("year", ""),
		Month:   c.Query("month", ""),
		Keyword: c.Query("keyword", ""),
		Var:     c.Query("var", ""),
		Turvar:  c.Query("turvar", ""),
		Vervar:  c.Query("vervar", ""),
		Th:      c.Query("th", ""),
		Turth:   c.Query("turth", ""),
	}
}

func (s *APIServer) handleBPSList(c *fiber.Ctx) error {
	params := s.extractBPSParams(c, "subject")
	result, err := s.bps.FetchGenericList(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"success":           true,
		"model":             params.Model,
		"status":            result.Status,
		"data_availability": result.DataAvailability,
		"pagination":        result.Pagination,
		"data":              result.Items,
		"is_stale":          result.IsStale,
	})
}

func (s *APIServer) handleBPSSubjects(c *fiber.Ctx) error {
	params := s.extractBPSParams(c, "subject")
	result, err := s.bps.GetSubjects(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"success":           true,
		"model":             "subject",
		"status":            result.Status,
		"data_availability": result.DataAvailability,
		"pagination":        result.Pagination,
		"data":              result.Items,
		"is_stale":          result.IsStale,
	})
}

func (s *APIServer) handleBPSSubcats(c *fiber.Ctx) error {
	params := s.extractBPSParams(c, "subcat")
	result, err := s.bps.GetSubCats(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"success":           true,
		"model":             "subcat",
		"status":            result.Status,
		"data_availability": result.DataAvailability,
		"pagination":        result.Pagination,
		"data":              result.Items,
		"is_stale":          result.IsStale,
	})
}

func (s *APIServer) handleBPSVariables(c *fiber.Ctx) error {
	params := s.extractBPSParams(c, "var")
	result, err := s.bps.GetVariables(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"success":           true,
		"model":             "var",
		"status":            result.Status,
		"data_availability": result.DataAvailability,
		"pagination":        result.Pagination,
		"data":              result.Items,
		"is_stale":          result.IsStale,
	})
}

func (s *APIServer) handleBPSStaticTables(c *fiber.Ctx) error {
	params := s.extractBPSParams(c, "statictable")
	result, err := s.bps.GetStaticTables(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"success":           true,
		"model":             "statictable",
		"status":            result.Status,
		"data_availability": result.DataAvailability,
		"pagination":        result.Pagination,
		"data":              result.Items,
		"is_stale":          result.IsStale,
	})
}

func (s *APIServer) handleBPSPublications(c *fiber.Ctx) error {
	params := s.extractBPSParams(c, "publication")
	result, err := s.bps.GetPublications(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"success":           true,
		"model":             "publication",
		"status":            result.Status,
		"data_availability": result.DataAvailability,
		"pagination":        result.Pagination,
		"data":              result.Items,
		"is_stale":          result.IsStale,
	})
}

func (s *APIServer) handleBPSPressReleases(c *fiber.Ctx) error {
	params := s.extractBPSParams(c, "pressrelease")
	result, err := s.bps.GetPressReleases(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"success":           true,
		"model":             "pressrelease",
		"status":            result.Status,
		"data_availability": result.DataAvailability,
		"pagination":        result.Pagination,
		"data":              result.Items,
		"is_stale":          result.IsStale,
	})
}

func (s *APIServer) handleBPSInfographics(c *fiber.Ctx) error {
	params := s.extractBPSParams(c, "infographic")
	result, err := s.bps.GetInfographics(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"success":           true,
		"model":             "infographic",
		"status":            result.Status,
		"data_availability": result.DataAvailability,
		"pagination":        result.Pagination,
		"data":              result.Items,
		"is_stale":          result.IsStale,
	})
}

func (s *APIServer) handleBPSData(c *fiber.Ctx) error {
	params := s.extractBPSParams(c, "data")
	result, err := s.bps.GetData(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"success":           true,
		"model":             "data",
		"status":            result.Status,
		"data_availability": result.DataAvailability,
		"pagination":        result.Pagination,
		"data":              result.Items,
		"is_stale":          result.IsStale,
	})
}

func (s *APIServer) handleBPSCacheClear(c *fiber.Ctx) error {
	s.bps.ClearCache()
	return c.JSON(fiber.Map{
		"success": true,
		"message": "BPS Web API in-memory cache successfully cleared",
	})
}

// Scraper Endpoints (§12)

func (s *APIServer) handleScraperStatus(c *fiber.Ctx) error {
	if s.scraperStore == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"error":   "Scraper not initialized",
		})
	}

	ctx := c.Context()
	var lastRunData interface{}
	if lr, err := s.scraperStore.GetLastRun(ctx); err == nil && lr != nil {
		lastRunData = lr
	}

	isOpen := false
	fails := 0
	var until *time.Time
	if s.scraperBreaker != nil {
		isOpen, fails, until, _ = s.scraperBreaker.Status()
	}

	counts, _ := s.scraperStore.GetCountsByKind(ctx)
	ageHours, _ := s.scraperStore.GetDataAgeHours(ctx)

	var nextRun interface{}
	if s.scraperScheduler != nil {
		nr := s.scraperScheduler.NextRun()
		if !nr.IsZero() {
			nextRun = nr.Format(time.RFC3339)
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"scheduler": fiber.Map{
			"enabled":  s.cfg.ScraperEnabled,
			"cron":     s.cfg.ScraperCronSpec,
			"timezone": s.cfg.ScraperTimezone,
			"next_run": nextRun,
		},
		"last_run": lastRunData,
		"breaker": fiber.Map{
			"open":             isOpen,
			"consecutive_fail": fails,
			"open_until":       until,
		},
		"counts":         counts,
		"data_age_hours": ageHours,
	})
}

func (s *APIServer) handleScraperRun(c *fiber.Ctx) error {
	if s.scraperRunner == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"error":   "Scraper runner not initialized",
		})
	}

	if s.scraperBreaker != nil {
		if open, until := s.scraperBreaker.IsOpen(); open {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"success":    false,
				"error":      "Circuit breaker is open",
				"open_until": until,
			})
		}
	}

	// Trigger runner asynchronously (§12)
	go func() {
		_, _ = s.scraperRunner.RunAll(context.Background(), scraper.TriggerManual)
	}()

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"success": true,
		"message": "Scraper run triggered successfully",
		"trigger": "manual",
	})
}

func (s *APIServer) handleScraperRuns(c *fiber.Ctx) error {
	if s.scraperStore == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"error":   "Scraper store not initialized",
		})
	}

	limit := c.QueryInt("limit", 20)
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	runs, err := s.scraperStore.GetRuns(c.Context(), limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    runs,
	})
}

func (s *APIServer) handleScraperItems(c *fiber.Ctx) error {
	if s.scraperStore == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"error":   "Scraper store not initialized",
		})
	}

	kindStr := c.Query("kind", string(scraper.KindPressRelease))
	lang := c.Query("lang", "id")
	q := c.Query("q")
	limit := c.QueryInt("limit", 20)
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var items []scraper.Item
	var err error

	if q != "" {
		items, err = s.scraperStore.SearchItems(c.Context(), scraper.Kind(kindStr), q, limit)
	} else {
		items, err = s.scraperStore.GetLatestItems(c.Context(), scraper.Kind(kindStr), lang, limit)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"kind":    kindStr,
		"lang":    lang,
		"data":    items,
	})
}

func (s *APIServer) handleScraperBreakerReset(c *fiber.Ctx) error {
	if s.scraperBreaker == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"error":   "Circuit breaker not initialized",
		})
	}

	s.scraperBreaker.Reset()
	log.Println("[SCRAPER] Circuit breaker manually reset by admin via REST API")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Circuit breaker reset successfully",
	})
}

// handleGetIndicators handles GET /api/v1/indicators (Menu 2 Strategic Indicators)
func (s *APIServer) handleGetIndicators(c *fiber.Ctx) error {
	prov := s.getProvider()
	if prov == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"error":   "Scraper database/provider is not initialized",
		})
	}

	indicators, freshness, err := prov.GetAllStrategicIndicators(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Filter by category or search keyword if requested
	catFilter := strings.ToLower(c.Query("category"))
	if catFilter != "" {
		filtered := make([]scraper.StrategicIndicator, 0)
		for _, ind := range indicators {
			if strings.Contains(strings.ToLower(ind.Category), catFilter) || strings.Contains(strings.ToLower(ind.Slug), catFilter) {
				filtered = append(filtered, ind)
			}
		}
		indicators = filtered
	}

	format := strings.ToLower(c.Query("format"))
	if format == "text" {
		var sb strings.Builder
		sb.WriteString("=== INDIKATOR STRATEGIS BPS KABUPATEN BANGKA BARAT ===\n\n")
		for _, ind := range indicators {
			sb.WriteString(fmt.Sprintf("[%d] %s (%s)\n", ind.ID, ind.Title, ind.Category))
			for _, m := range ind.Metrics {
				sb.WriteString(fmt.Sprintf("  • %s: %s", m.Label, m.Value))
				if m.Note != "" {
					sb.WriteString(fmt.Sprintf(" (%s)", m.Note))
				}
				sb.WriteString("\n")
			}
			if len(ind.Items) > 0 {
				sb.WriteString(fmt.Sprintf("  Dokumen: %s\n", ind.Items[0].Title))
				if ind.Items[0].FileURL != "" {
					sb.WriteString(fmt.Sprintf("  Unduh: %s\n", ind.Items[0].FileURL))
				}
			}
			sb.WriteString("\n")
		}
		return c.Type("text").SendString(sb.String())
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"total":     len(indicators),
		"freshness": freshness,
		"data":      indicators,
	})
}

// handleGetIndicatorByID handles GET /api/v1/indicators/:id
func (s *APIServer) handleGetIndicatorByID(c *fiber.Ctx) error {
	idOrSlug := c.Params("id")
	prov := s.getProvider()
	if prov == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"error":   "Scraper database/provider is not initialized",
		})
	}

	ind, freshness, err := prov.GetStrategicIndicator(c.Context(), idOrSlug)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	format := strings.ToLower(c.Query("format"))
	if format == "text" {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("*%s*\n\n", ind.Title))
		for _, m := range ind.Metrics {
			sb.WriteString(fmt.Sprintf("• *%s*: %s", m.Label, m.Value))
			if m.Note != "" {
				sb.WriteString(fmt.Sprintf(" (%s)", m.Note))
			}
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("\n*Sumber*: %s\n", ind.Source))
		if len(ind.Items) > 0 {
			sb.WriteString(fmt.Sprintf("*Dokumen Resmi*: %s\n", ind.Items[0].Title))
			if ind.Items[0].FileURL != "" {
				sb.WriteString(fmt.Sprintf("*Unduh PDF*: %s\n", ind.Items[0].FileURL))
			}
		}
		return c.Type("text").SendString(sb.String())
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"freshness": freshness,
		"data":      ind,
	})
}

// handleGetLayananData handles GET /api/v1/layanan-data (Menu 1 Layanan Data)
func (s *APIServer) handleGetLayananData(c *fiber.Ctx) error {
	prov := s.getProvider()
	if prov == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"error":   "Scraper database/provider is not initialized",
		})
	}

	// Keyword search query if passed: ?q=<search>
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		query = strings.TrimSpace(c.Query("keyword"))
	}
	if query != "" {
		items, freshness, err := prov.SearchItems(c.Context(), query, "id", 10)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		}
		tables, _, _ := prov.SearchStaticTables(c.Context(), query, 10)
		return c.JSON(fiber.Map{
			"success":      true,
			"query":        query,
			"total_tables": len(tables),
			"total_items":  len(items),
			"freshness":    freshness,
			"tables":       tables,
			"items":        items,
		})
	}

	categories, freshness, err := prov.GetLayananDataCategories(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	format := strings.ToLower(c.Query("format"))
	if format == "text" {
		var sb strings.Builder
		sb.WriteString("=== LAYANAN DATA BPS KABUPATEN BANGKA BARAT (MENU 1) ===\n\n")
		for _, cat := range categories {
			sb.WriteString(fmt.Sprintf("*[%s] %s*\n", cat.ID, cat.Title))
			sb.WriteString(cat.Description + "\n")
			for _, top := range cat.Topics {
				sb.WriteString(fmt.Sprintf("  • %d. %s (%d tabel, %d dokumen)\n", top.ID, top.Title, len(top.Tables), len(top.Items)))
			}
			sb.WriteString("\n")
		}
		return c.Type("text").SendString(sb.String())
	}

	return c.JSON(fiber.Map{
		"success":          true,
		"total_categories": len(categories),
		"freshness":        freshness,
		"data":             categories,
	})
}

// handleGetLayananCategory handles GET /api/v1/layanan-data/:category
func (s *APIServer) handleGetLayananCategory(c *fiber.Ctx) error {
	catSlug := c.Params("category")
	prov := s.getProvider()
	if prov == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"error":   "Scraper database/provider is not initialized",
		})
	}

	cat, freshness, err := prov.GetLayananDataCategory(c.Context(), catSlug)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	format := strings.ToLower(c.Query("format"))
	if format == "text" {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("*%s*\n\n", cat.Title))
		sb.WriteString(cat.Description + "\n\n")
		sb.WriteString("*Daftar Topik Statistik*:\n")
		for _, top := range cat.Topics {
			sb.WriteString(fmt.Sprintf("%d. *%s*\n", top.ID, top.Title))
			if top.Description != "" {
				sb.WriteString(fmt.Sprintf("   %s\n", top.Description))
			}
			if len(top.Tables) > 0 {
				sb.WriteString(fmt.Sprintf("   Tabel Utama: %s\n", top.Tables[0].Title))
			}
			if len(top.Items) > 0 {
				sb.WriteString(fmt.Sprintf("   Dokumen: %s\n", top.Items[0].Title))
			}
			sb.WriteString("\n")
		}
		return c.Type("text").SendString(sb.String())
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"freshness": freshness,
		"data":      cat,
	})
}

// handleGetLayananTopic handles GET /api/v1/layanan-data/:category/:topic
func (s *APIServer) handleGetLayananTopic(c *fiber.Ctx) error {
	catSlug := c.Params("category")
	topicIdOrSlug := c.Params("topic")
	prov := s.getProvider()
	if prov == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"error":   "Scraper database/provider is not initialized",
		})
	}

	topic, freshness, err := prov.GetLayananDataTopic(c.Context(), catSlug, topicIdOrSlug)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	format := strings.ToLower(c.Query("format"))
	if format == "text" {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("*%s*\n\n", topic.Title))
		if topic.Description != "" {
			sb.WriteString(topic.Description + "\n\n")
		}
		if topic.SubjectURL != "" {
			sb.WriteString(fmt.Sprintf("*Portal BPS*: %s\n\n", topic.SubjectURL))
		}
		if len(topic.Tables) > 0 {
			sb.WriteString("*Tabel Statistik Terkait*:\n")
			for i, t := range topic.Tables {
				sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, t.Title))
				if t.URL != "" {
					sb.WriteString(fmt.Sprintf("   Lihat: %s\n", t.URL))
				}
			}
			sb.WriteString("\n")
		}
		if len(topic.Items) > 0 {
			sb.WriteString("*Publikasi & Dokumen Terkait*:\n")
			for i, it := range topic.Items {
				sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, it.Title))
				if it.FileURL != "" {
					sb.WriteString(fmt.Sprintf("   Unduh: %s\n", it.FileURL))
				} else if it.URL != "" {
					sb.WriteString(fmt.Sprintf("   Link: %s\n", it.URL))
				}
			}
			sb.WriteString("\n")
		}
		if freshness.UpdatedAt.Year() > 2000 {
			sb.WriteString(fmt.Sprintf("ℹSumber: Database Scraper BPS · diperbarui %s WIB", freshness.UpdatedAt.Format("02 Jan 2006 15:04")))
		}
		return c.Type("text").SendString(sb.String())
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"freshness": freshness,
		"data":      topic,
	})
}

func (s *APIServer) Listen(addr string) error {
	return s.app.Listen(addr)
}

func (s *APIServer) Shutdown() error {
	return s.app.Shutdown()
}

