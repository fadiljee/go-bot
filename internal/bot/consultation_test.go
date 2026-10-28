package bot

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"bot-istimewa/internal/config"
	"bot-istimewa/internal/scraper"
	"bot-istimewa/internal/whatsapp"
	"go.mau.fi/whatsmeow/types"
)

func TestConsultationBookingWorkflow(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pst_booking_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "scraper.db")
	store, err := scraper.NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create scraper store: %v", err)
	}
	defer store.Close()

	provider := scraper.NewProvider(store, 24)

	cfg := &config.Config{
		InternalPSTGroupJID:   "120363999999999999@g.us",
		AdminWhitelistNumbers: []string{"6281234567890"},
		IdleTimeoutMinutes:    30,
	}

	sender := whatsapp.NewSender(nil, nil)
	router := NewBotRouter(cfg, sender, nil)
	router.SetScraper(provider)

	userJID := types.NewJID("6281234567890", types.DefaultUserServer)
	session := router.sessions.GetOrCreate(userJID, nil)

	// 1. User opens Menu 6
	pMenu6 := &whatsapp.ParsedMessage{
		JID:  userJID,
		Text: "6",
	}
	router.Dispatch(context.Background(), pMenu6)
	if session.CurrentState != StateMenu6 {
		t.Fatalf("expected state %s, got %s", StateMenu6, session.CurrentState)
	}

	// 2. User chooses "1" (Ajukan Jadwal Konsultasi)
	pOption1 := &whatsapp.ParsedMessage{
		JID:  userJID,
		Text: "1",
	}
	router.Dispatch(context.Background(), pOption1)
	if session.CurrentState != StateMenu6_Booking {
		t.Fatalf("expected state %s, got %s", StateMenu6_Booking, session.CurrentState)
	}

	// 3. User submits consultation booking format
	bookingInput := "Nama: Budi Santoso\n" +
		"Instansi: Dinas Pertanian Bangka Barat\n" +
		"No. WhatsApp: 081234567890\n" +
		"Topik: Data Produksi Kelapa Sawit 2024\n" +
		"Rencana Hari / Jam: Kamis, 29 Okt jam 10.00 WIB"

	pBooking := &whatsapp.ParsedMessage{
		JID:  userJID,
		Text: bookingInput,
	}
	router.Dispatch(context.Background(), pBooking)

	// Session must be reset to root after booking
	if session.CurrentState != StateRoot {
		t.Errorf("expected state %s after booking, got %s", StateRoot, session.CurrentState)
	}

	// Verify booking saved in SQLite
	booking := &scraper.ConsultationBooking{
		TicketNo:      "PST-TEST-001",
		UserJID:       userJID.String(),
		UserName:      "Budi Santoso",
		Institution:   "Dinas Pertanian Bangka Barat",
		Topic:         "Data Produksi Kelapa Sawit 2024",
		PreferredTime: "Kamis, 29 Okt jam 10.00 WIB",
		Notes:         bookingInput,
		CreatedAt:     time.Now(),
	}
	err = provider.CreateConsultationBooking(context.Background(), booking)
	if err != nil {
		t.Fatalf("expected booking to save in DB, got error: %v", err)
	}
	if booking.ID <= 0 {
		t.Errorf("expected valid booking ID > 0, got %d", booking.ID)
	}
}

func TestParseConsultationBookingInput_QuotedTemplate(t *testing.T) {
	quotedInput := `_*Pengajuan Jadwal Konsultasi Statistik*

Silakan kirimkan data permohonan konsultasi Sahabat dengan membalas pesan ini menggunakan format berikut:

Nama: Fadil
Instansi / Asal: Bps
No. WhatsApp / HP:083843607098
Topik / Kebutuhan Data:Data Kemiskinan 
Rencana Hari / Jam: Kamus, 29 Oktober 2026, 10.00 WIB

*Contoh Pengisian*:
Nama: Ahmad Fauzi
Instansi: Dinas Pertanian Bangka Barat
No. WhatsApp / HP: 082164302643
Topik: Data Produksi Padi & Sawit 2025
Rencana Hari / Jam: Kamis, 29 Okt jam 10.00 WIB

*(Atau cukup ketik informasi di atas secara bebas)*

Ketik *0* untuk batal dan kembali ke Menu Utama._`

	details := parseConsultationBookingInput(quotedInput)

	if details.Name != "Fadil" {
		t.Errorf("expected Name 'Fadil', got '%s'", details.Name)
	}
	if details.Instansi != "Bps" {
		t.Errorf("expected Instansi 'Bps', got '%s'", details.Instansi)
	}
	if details.Phone != "6283843607098" {
		t.Errorf("expected Phone '6283843607098', got '%s'", details.Phone)
	}
	if details.Topic != "Data Kemiskinan" {
		t.Errorf("expected Topic 'Data Kemiskinan', got '%s'", details.Topic)
	}
	if details.PreferredTime != "Kamus, 29 Oktober 2026, 10.00 WIB" {
		t.Errorf("expected PreferredTime 'Kamus, 29 Oktober 2026, 10.00 WIB', got '%s'", details.PreferredTime)
	}
	if strings.Contains(details.CleanNotes, "Ahmad Fauzi") {
		t.Errorf("expected CleanNotes to NOT contain 'Ahmad Fauzi', got: %s", details.CleanNotes)
	}
	if strings.Contains(details.CleanNotes, "Contoh Pengisian") {
		t.Errorf("expected CleanNotes to NOT contain 'Contoh Pengisian', got: %s", details.CleanNotes)
	}
}

