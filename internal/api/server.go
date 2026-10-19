package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"strings"
	"time"

	"bot-istimewa/internal/config"
	"bot-istimewa/internal/gsheet"
	"bot-istimewa/internal/whatsapp"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.mau.fi/whatsmeow/types"
)

type APIServer struct {
	app       *fiber.App
	cfg       *config.Config
	clientMgr *whatsapp.ClientManager
	gsheet    *gsheet.Client
	startTime time.Time
}

func NewAPIServer(cfg *config.Config, clientMgr *whatsapp.ClientManager, gsheetClient *gsheet.Client) *APIServer {
	app := fiber.New(fiber.Config{
		AppName:               "BOT ISTIMEWA v3.0 (Go Edition)",
		DisableStartupMessage: true,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	srv := &APIServer{
		app:       app,
		cfg:       cfg,
		clientMgr: clientMgr,
		gsheet:    gsheetClient,
		startTime: time.Now(),
	}

	srv.setupRoutes()
	return srv
}

func (s *APIServer) setupRoutes() {
	// Public Health Check
	s.app.Get("/health", s.handleHealth)

	// Web Browser QR Scanner
	s.app.Get("/qr", s.handleQRPage)

	// API Group
	api := s.app.Group("/api", s.authMiddleware)

	// Cache Invalidation
	api.Post("/cache/refresh", s.handleCacheRefresh)

	// Asynchronous Blast (FR-07)
	api.Post("/blast", s.handleBlast)
}

func (s *APIServer) authMiddleware(c *fiber.Ctx) error {
	apiKey := c.Get("X-API-Key")
	if apiKey == "" {
		authHeader := c.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			apiKey = strings.TrimPrefix(authHeader, "Bearer ")
		}
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
		"version":   "3.0.0 (Go Native)",
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

func (s *APIServer) handleQRPage(c *fiber.Ctx) error {
	if s.clientMgr.IsLoggedIn {
		return c.Type("html").SendString(`
			<!DOCTYPE html>
			<html>
			<head><title>BOT ISTIMEWA - QR Code</title><meta name="viewport" content="width=device-width, initial-scale=1"></head>
			<body style="font-family:sans-serif; text-align:center; padding: 40px; background:#f4f6f9;">
				<div style="max-width:450px; margin:auto; background:white; padding:30px; border-radius:12px; box-shadow:0 4px 10px rgba(0,0,0,0.1);">
					<h2 style="color:#2e7d32;">✅ Bot Sudah Terhubung!</h2>
					<p>BOT ISTIMEWA WhatsApp BPS Kabupaten Bangka Barat sedang aktif dan siap melayani.</p>
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

	// Run blast asynchronously in background goroutine with random delay
	go func(numbers []string, msg string) {
		log.Printf("[BLAST] Memulai pengiriman pesan ke %d nomor subscriber...", len(numbers))
		for i, num := range numbers {
			cleanNum := strings.TrimSpace(num)
			cleanNum = strings.TrimPrefix(cleanNum, "+")
			if strings.HasPrefix(cleanNum, "08") {
				cleanNum = "62" + cleanNum[1:]
			}

			jid := types.NewJID(cleanNum, types.DefaultUserServer)
			err := s.clientMgr.Sender.SendTextMessage(context.Background(), jid, msg)
			if err != nil {
				log.Printf("[BLAST] Gagal mengirim ke %s: %v", cleanNum, err)
			} else {
				log.Printf("[BLAST] (%d/%d) Berhasil terkirim ke %s", i+1, len(numbers), cleanNum)
			}

			// Random delay 3 to 7 seconds to prevent ban
			delay := time.Duration(3000+rand.Intn(4000)) * time.Millisecond
			time.Sleep(delay)
		}
		log.Println("[BLAST] Seluruh pengiriman blast selesai.")
	}(req.Numbers, req.Message)

	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Pengiriman blast ke %d nomor sedang diproses di latar belakang.", len(req.Numbers)),
	})
}

func (s *APIServer) Listen(addr string) error {
	return s.app.Listen(addr)
}

func (s *APIServer) Shutdown() error {
	return s.app.Shutdown()
}

