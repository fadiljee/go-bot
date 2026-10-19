package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bot-istimewa/internal/api"
	"bot-istimewa/internal/bot"
	"bot-istimewa/internal/config"
	"bot-istimewa/internal/gsheet"
	"bot-istimewa/internal/locale"
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

	// 5. Initialize Bot Message Router
	botRouter := bot.NewBotRouter(cfg, clientMgr.Sender, gsheetClient)

	botStartTime := time.Now()

	// 6. Register Event Listener
	clientMgr.Client.AddEventHandler(func(rawEvt interface{}) {
		switch evt := rawEvt.(type) {
		case *events.Message:
			// Strict Source Filtering (Lapis 2): ignore self, groups, status/broadcasts
			if evt.Info.IsFromMe || evt.Info.IsGroup || evt.Info.Chat.Server == "broadcast" || evt.Info.Chat.Server == "newsletter" {
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
	apiServer := api.NewAPIServer(cfg, clientMgr, gsheetClient)
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

	if err := apiServer.Shutdown(); err != nil {
		log.Printf("⚠️ Gagal mematikan API server: %v", err)
	}

	clientMgr.Client.Disconnect()
	log.Println("🔌 WhatsApp client terputus.")

	select {
	case <-shutdownCtx.Done():
		log.Println("⏱️ Graceful shutdown timeout.")
	default:
		log.Println("👋 BOT ISTIMEWA v3.0 berhasil dimatikan dengan bersih.")
	}
}
