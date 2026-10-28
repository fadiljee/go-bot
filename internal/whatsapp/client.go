package whatsapp

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"bot-istimewa/internal/config"

	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	_ "modernc.org/sqlite"
)

type ClientManager struct {
	Client     *whatsmeow.Client
	Sender     *Sender
	cfg        *config.Config
	qrCodePNG  []byte
	qrCodeStr  string
	qrMu       sync.RWMutex
	IsLoggedIn bool
}

func NewClientManager(cfg *config.Config) (*ClientManager, error) {
	_ = os.MkdirAll("storage", 0755)

	ctx := context.Background()
	dbLog := waLog.Stdout("Database", "ERROR", true)
	container, err := sqlstore.New(ctx, "sqlite", "file:storage/whatsmeow.db?_pragma=foreign_keys(1)&_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)", dbLog)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite store: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get first device: %w", err)
	}

	// Set custom device name for WhatsApp Web identity
	store.DeviceProps.Os = &cfg.WABrowserName

	clientLog := waLog.Stdout("WhatsApp", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	mq := NewMessageQueue(client, cfg.QueueInteractiveMinDelayMs, cfg.QueueInteractiveMaxDelayMs, cfg.QueueBlastMinDelayMs, cfg.QueueBlastMaxDelayMs)

	mgr := &ClientManager{
		Client: client,
		Sender: NewSender(client, mq),
		cfg:    cfg,
	}

	return mgr, nil
}

// Start connects to WhatsApp and manages QR codes
func (m *ClientManager) Start(ctx context.Context) error {
	if m.Client.Store.ID == nil {
		// New login required
		qrChan, _ := m.Client.GetQRChannel(ctx)
		err := m.Client.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}

		go func() {
			for evt := range qrChan {
				if evt.Event == "code" {
					log.Printf("[WHATSAPP] Kode QR baru diterima. Silakan scan di terminal atau buka http://localhost:%s/qr", m.cfg.AppPort)
					m.setQR(evt.Code)

					// Print terminal QR
					q, err := qrcode.New(evt.Code, qrcode.Medium)
					if err == nil {
						fmt.Println(q.ToSmallString(false))
					}
				} else {
					log.Printf("[WHATSAPP] QR Event: %s", evt.Event)
					if evt.Event == "success" {
						m.clearQR()
						m.IsLoggedIn = true
						log.Println("[STATUS] Bot WhatsApp BERHASIL TERHUBUNG!")
					}
				}
			}
		}()
	} else {
		// Already logged in
		err := m.Client.Connect()
		if err != nil {
			return fmt.Errorf("failed to reconnect: %w", err)
		}
		m.IsLoggedIn = true
		log.Println("[STATUS] Bot WhatsApp BERHASIL TERHUBUNG menggunakan sesi tersimpan!")
	}

	return nil
}

func (m *ClientManager) setQR(code string) {
	png, err := qrcode.Encode(code, qrcode.Medium, 256)
	m.qrMu.Lock()
	defer m.qrMu.Unlock()
	m.qrCodeStr = code
	if err == nil {
		m.qrCodePNG = png
	}
}

func (m *ClientManager) clearQR() {
	m.qrMu.Lock()
	defer m.qrMu.Unlock()
	m.qrCodeStr = ""
	m.qrCodePNG = nil
}

// GetQR returns the current QR code PNG bytes if waiting for scan
func (m *ClientManager) GetQR() ([]byte, string) {
	m.qrMu.RLock()
	defer m.qrMu.RUnlock()
	return m.qrCodePNG, m.qrCodeStr
}
