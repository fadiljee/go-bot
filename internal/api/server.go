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
	app       *fiber.App
	cfg       *config.Config
	clientMgr *whatsapp.ClientManager
	gsheet    *gsheet.Client
	bps       *bpsapi.Client
	startTime time.Time
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

	return c.JSON(fiber.Map{
		"app":       "BOT ISTIMEWA",
		"version":   "3.0.0 (Go Native Hardened)",
		"agency":    "BPS Kabupaten Bangka Barat",
		"connected": s.clientMgr.IsLoggedIn,
		"uptime":    time.Since(s.startTime).String(),
		"memory": fiber.Map{
			"alloc_mb":       fmt.Sprintf("%.2f MB", float64(m.Alloc)/(1024*1024)),
			"sys_mb":         fmt.Sprintf("%.2f MB", float64(m.Sys)/(1024*1024)),
			"num_goroutines": runtime.NumGoroutine(),
		},
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
					<h2 style="color:#d32f2f;">🔒 Akses QR Terkunci (403 Forbidden)</h2>
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
					<h2>⏳ Menunggu QR Code WhatsApp...</h2>
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
		{ID: "berhenti_langganan", DisplayText: "🔕 Berhenti Berlangganan"},
	}
	blastMsg := whatsapp.BuildQuickReplyButtons("📢 Rilis Data BPS Bangka Barat", req.Message, "PST BPS Kabupaten Bangka Barat", optOutButtons)

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

func (s *APIServer) Listen(addr string) error {
	return s.app.Listen(addr)
}

func (s *APIServer) Shutdown() error {
	return s.app.Shutdown()
}
