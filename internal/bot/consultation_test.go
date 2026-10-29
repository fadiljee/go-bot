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

func TestParseConsultationBookingInput_UserReportedCase(t *testing.T) {
	input := `_*Pengajuan Jadwal Konsultasi Statistik*
  
  Silakan kirimkan data permohonan konsultasi
  Sahabat dengan membalas pesan ini menggunakan
  format berikut:
  
  ` + "```" + `Nama:Fadil
  Instansi / Asal: Bps Babar
  No. WhatsApp / HP: 08216262626
  Topik / Kebutuhan Data: Kemiskinan
  Rencana Hari / Jam:` + "```" + `Kamis, 20 Oktober 2026
  10.00 WIB
  
  *Contoh Pengisian*:
  Nama: Ahmad Fauzi
  Instansi: Dinas Pertanian Bangka Barat
  No. WhatsApp / HP: 082164302643
  Topik: Data Produksi Padi & Sawit 2025
  Rencana Hari / Jam: Kamis, 29 Okt jam 10.00 WIB
  
  *(Atau cukup ketik informasi di atas secara
  bebas)*
  
  Ketik *0* untuk batal dan kembali ke Menu
  Utama._`

	details := parseConsultationBookingInput(input)

	if details.Name != "Fadil" {
		t.Errorf("expected Name 'Fadil', got '%s'", details.Name)
	}
	if details.Instansi != "Bps Babar" {
		t.Errorf("expected Instansi 'Bps Babar', got '%s'", details.Instansi)
	}
	if details.Phone != "628216262626" {
		t.Errorf("expected Phone '628216262626', got '%s'", details.Phone)
	}
	if details.Topic != "Kemiskinan" {
		t.Errorf("expected Topic 'Kemiskinan', got '%s'", details.Topic)
	}
	if !strings.Contains(details.PreferredTime, "Kamis, 20 Oktober 2026") {
		t.Errorf("expected PreferredTime to contain 'Kamis, 20 Oktober 2026', got '%s'", details.PreferredTime)
	}
	if strings.Contains(details.CleanNotes, "Ahmad Fauzi") {
		t.Errorf("expected CleanNotes to NOT contain 'Ahmad Fauzi', got: %s", details.CleanNotes)
	}
	if strings.Contains(details.CleanNotes, "format berikut") {
		t.Errorf("expected CleanNotes to NOT contain 'format berikut', got: %s", details.CleanNotes)
	}
}

func TestParseConsultationBookingInput_MultilineAndNumbered(t *testing.T) {
	// Case with numbered list
	inputNumbered := `1. Nama: Fadil
2. Instansi / Asal: BPS Bangka Barat
3. No. WhatsApp: 08216262626
4. Topik: Data Inflasi 2026
5. Jadwal: Jumat, 30 Okt jam 09.00 WIB

Contoh:
Nama: Ahmad Fauzi`

	details := parseConsultationBookingInput(inputNumbered)
	if details.Name != "Fadil" {
		t.Errorf("expected Name 'Fadil', got '%s'", details.Name)
	}
	if details.Instansi != "BPS Bangka Barat" {
		t.Errorf("expected Instansi 'BPS Bangka Barat', got '%s'", details.Instansi)
	}
	if details.Phone != "628216262626" {
		t.Errorf("expected Phone '628216262626', got '%s'", details.Phone)
	}
	if details.Topic != "Data Inflasi 2026" {
		t.Errorf("expected Topic 'Data Inflasi 2026', got '%s'", details.Topic)
	}
	if strings.Contains(details.CleanNotes, "Ahmad Fauzi") {
		t.Errorf("expected CleanNotes to NOT contain 'Ahmad Fauzi', got: %s", details.CleanNotes)
	}

	// Case with multiline values (value on next line)
	inputMultiline := `Nama:
Fadil
Instansi:
BPS Bangka Barat
No HP:
08216262626
Topik:
Data PDRB
Jadwal:
Kamis 10.00 WIB`

	detailsMulti := parseConsultationBookingInput(inputMultiline)
	if detailsMulti.Name != "Fadil" {
		t.Errorf("expected Name 'Fadil', got '%s'", detailsMulti.Name)
	}
	if detailsMulti.Instansi != "BPS Bangka Barat" {
		t.Errorf("expected Instansi 'BPS Bangka Barat', got '%s'", detailsMulti.Instansi)
	}
	if detailsMulti.Phone != "628216262626" {
		t.Errorf("expected Phone '628216262626', got '%s'", detailsMulti.Phone)
	}
	if detailsMulti.Topic != "Data PDRB" {
		t.Errorf("expected Topic 'Data PDRB', got '%s'", detailsMulti.Topic)
	}
}

