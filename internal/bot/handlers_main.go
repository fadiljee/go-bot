package bot

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log"
	"strings"
	"time"

	"bot-istimewa/internal/locale"
	"bot-istimewa/internal/whatsapp"
)

func getRootMenuItems() []whatsapp.ListRow {
	return []whatsapp.ListRow{
		{ID: "1", Title: "Layanan Data", Description: "Statistik Sosial, Ekonomi & Pertanian"},
		{ID: "2", Title: "Indikator Strategis", Description: "Kemiskinan, Inflasi, IPM, PDRB, dll"},
		{ID: "3", Title: "BRS & Publikasi Rilis", Description: "Rilis Berita Resmi Statistik & Publikasi"},
		{ID: "4", Title: "Video Edukasi Literasi", Description: "Video edukasi dan literasi statistik"},
		{ID: "5", Title: "Galeri Infografis Daerah", Description: "Visualisasi ringkas data BPS Bangka Barat"},
		{ID: "6", Title: "Tanya Pakar", Description: "Konsultasi online dengan pakar BPS"},
		{ID: "7", Title: "Channel WhatsApp BPS Bangka Barat", Description: "Saluran informasi resmi BPS Bangka Barat"},
		{ID: "8", Title: "Pengaduan & Saran", Description: "Sampaikan kritik dan saran layanan"},
		{ID: "9", Title: "Bantuan & FAQ", Description: "Pertanyaan yang sering diajukan"},
		{ID: "10", Title: "Tentang BPS", Description: "Profil BPS Kabupaten Bangka Barat"},
	}
}

func getTimeGreeting() string {
	loc, err := time.LoadLocation("Asia/Jakarta")
	now := time.Now()
	if err == nil {
		now = now.In(loc)
	}
	hour := now.Hour()
	switch {
	case hour >= 4 && hour < 11:
		return "Selamat Pagi"
	case hour >= 11 && hour < 15:
		return "Selamat Siang"
	case hour >= 15 && hour < 18:
		return "Selamat Sore"
	default:
		return "Selamat Malam"
	}
}

func buildWelcomeMenuBody(greetingLine, introSubtitle string) string {
	var sb strings.Builder
	sb.WriteString("*BOT ISTIMEWA*\n")
	sb.WriteString("BPS Kabupaten Bangka Barat\n\n")
	sb.WriteString(greetingLine + "\n")
	sb.WriteString(introSubtitle + "\n\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("*DATA DAN INDIKATOR*\n")
	sb.WriteString("1. Layanan Data\n")
	sb.WriteString("2. Indikator Strategis\n")
	sb.WriteString("3. BRS & Publikasi Rilis\n\n")
	sb.WriteString("*KONTEN DAN VISUAL*\n")
	sb.WriteString("4. Video Edukasi Literasi\n")
	sb.WriteString("5. Galeri Infografis Daerah\n\n")
	sb.WriteString("*KONSULTASI DAN LAYANAN*\n")
	sb.WriteString("6. Tanya Pakar\n")
	sb.WriteString("7. Channel WhatsApp BPS Bangka Barat\n")
	sb.WriteString("8. Pengaduan & Saran\n\n")
	sb.WriteString("*INFORMASI*\n")
	sb.WriteString("9. Bantuan & FAQ\n")
	sb.WriteString("10. Tentang BPS\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━\n\n")
	sb.WriteString("Balas dengan *angka menu (1 - 10)*\n")
	sb.WriteString("```Statistik Bermakna, Statistik Berdampak```")
	return sb.String()
}

// HandleWelcomeFull handles first message / keyword 100
func (r *BotRouter) HandleWelcomeFull(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.mu.Lock()
	if !s.LastWelcomeSent.IsZero() && time.Since(s.LastWelcomeSent) < 3*time.Second {
		s.mu.Unlock()
		log.Printf("[ROUTER] Mengabaikan HandleWelcomeFull duplikat untuk %s (baru dikirim %v lalu)", p.JID.String(), time.Since(s.LastWelcomeSent).Round(time.Millisecond))
		return nil
	}
	s.LastWelcomeSent = time.Now()
	s.CurrentState = StateRoot
	s.HistoryStack = nil
	s.mu.Unlock()

	// Format personalized greeting with time & name (PRD §6.1 F-01 & §14)
	userName := strings.TrimSpace(p.Name)
	timeGreeting := getTimeGreeting()
	var greetingLine string
	if userName != "" {
		greetingLine = fmt.Sprintf("%s, *%s*!", timeGreeting, userName)
	} else {
		greetingLine = fmt.Sprintf("%s, Sahabat Data!", timeGreeting)
	}

	body := buildWelcomeMenuBody(greetingLine, "Saya *Ak Isti*, asisten virtual statistik.")
	items := getRootMenuItems()

	msg := whatsapp.BuildInteractiveList(
		"",
		body,
		"",
		"Pilih Nomor Layanan",
		[]whatsapp.ListSection{
			{Title: "Menu Layanan Utama", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleWelcomeShort handles returning to main menu (keyword 0)
func (r *BotRouter) HandleWelcomeShort(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.mu.Lock()
	if !s.LastWelcomeSent.IsZero() && time.Since(s.LastWelcomeSent) < 2*time.Second {
		s.mu.Unlock()
		log.Printf("[ROUTER] Mengabaikan HandleWelcomeShort duplikat untuk %s (baru dikirim %v lalu)", p.JID.String(), time.Since(s.LastWelcomeSent).Round(time.Millisecond))
		return nil
	}
	s.LastWelcomeSent = time.Now()
	s.CurrentState = StateRoot
	s.HistoryStack = nil
	s.mu.Unlock()

	userName := strings.TrimSpace(p.Name)
	var greetingLine string
	if userName != "" {
		greetingLine = fmt.Sprintf("Hai kembali, *%s*!", userName)
	} else {
		greetingLine = "Hai kembali, Sahabat Data!"
	}

	body := buildWelcomeMenuBody(greetingLine, "Yuk lanjut berselancar menjelajahi data Bangka Barat.")
	items := getRootMenuItems()

	msg := whatsapp.BuildInteractiveList(
		"",
		body,
		"",
		"Pilih Nomor Layanan",
		[]whatsapp.ListSection{
			{Title: "Menu Layanan Utama", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleReset clears the user session state
func (r *BotRouter) HandleReset(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	sessionKey := p.JID.ToNonAD().String()
	if p.PhoneNumber != "" {
		sessionKey = p.PhoneNumber
	}
	r.sessions.Delete(p.JID, sessionKey)
	text := "Sesi percakapan Anda telah berhasil direset.\nKetik pesan apa saja atau ketik *0* untuk kembali ke menu utama ISTIMEWA."
	return r.sender.SendTextMessage(ctx, p.JID, text)
}

// HandleBack pops state from stack and re-executes previous menu
func (r *BotRouter) HandleBack(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	prevState := s.PopState()
	log.Printf("[NAVIGASI] Tombol Kembali (99) ditekan oleh %s -> Kembali ke state: %s", whatsapp.MaskJID(p.JID.User), prevState)

	switch prevState {
	case StateAdminLoggedIn, StateAdminManagePakar:
		s.CurrentState = StateAdminLoggedIn
		return r.SendAdminDashboard(ctx, s, p)
	case StateRoot:
		return r.HandleWelcomeShort(ctx, s, p)
	case StateMenu1:
		return r.HandleMenu1LayananData(ctx, s, p)
	case StateMenu1Sosduk:
		return r.HandleLayananSubSosduk(ctx, s, p)
	case StateMenu1Ekonomi:
		return r.HandleLayananSubEkonomi(ctx, s, p)
	case StateMenu1Pertanian:
		return r.HandleLayananSubPertanian(ctx, s, p)
	case StateMenu2:
		return r.HandleMenu2Indikator(ctx, s, p)
	case StateMenu3:
		return r.HandleMenu3Produk(ctx, s, p)
	case StateMenu4, StateMenu4Sosduk, StateMenu4Ekonomi:
		return r.HandleMenu4Video(ctx, s, p)
	case StateMenu5:
		return r.HandleMenu5Infografik(ctx, s, p)
	case StateMenu6, StateMenu6Booking:
		return r.HandleMenu6Konsultasi(ctx, s, p)
	case StateMenu6Pakar:
		return r.HandleKonsultasiPakarList(ctx, s, p)
	case StateMenu10:
		return r.HandleMenu10FAQ(ctx, s, p)
	default:
		return r.HandleWelcomeShort(ctx, s, p)
	}
}

func getAdminMenuItems() []whatsapp.ListRow {
	return []whatsapp.ListRow{
		{ID: "1", Title: "Kelola Kontak Pakar", Description: "Lihat & ubah nomor WA pakar"},
		{ID: "2", Title: "Status Scraper & Data", Description: "Cek statistik data & circuit breaker"},
		{ID: "3", Title: "Scraping Manual BPS", Description: "Pemicu pembaruan on-demand portal"},
	}
}

// HandleAdminPrompt asks admin password
func (r *BotRouter) HandleAdminPrompt(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	msg := "*LOGIN PANEL ADMIN*\n" +
		"BPS Kabupaten Bangka Barat\n" +
		"━━━━━━━━━━━━━━━━━━\n\n" +
		"Silakan masukkan *Password Admin* untuk melanjutkan verifikasi akses:\n\n" +
		"━━━━━━━━━━━━━━━━━━\n" +
		"• Ketik *0* untuk membatalkan dan kembali ke Menu Utama"
	return r.sender.SendTextMessage(ctx, p.JID, msg)
}

// HandleAdminAuth verifies admin password from .env, applies brute-force lockout, and auto-expires admin session
func (r *BotRouter) HandleAdminAuth(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	phone := p.JID.User

	// Check lockout
	if locked, until := r.isAdminLockedOut(phone); locked {
		s.ResetToRoot()
		msg := fmt.Sprintf("*AKSES DIKUNCI*\n"+
			"BPS Kabupaten Bangka Barat\n"+
			"━━━━━━━━━━━━━━━━━━\n\n"+
			"Nomor Anda sedang dikunci dari percobaan login admin hingga *%s WIB*.\n"+
			"Silakan coba lagi nanti.\n\n"+
			"━━━━━━━━━━━━━━━━━━\n"+
			"• Ketik *0* untuk kembali ke Menu Utama", until.Format("15:04:05"))
		return r.sender.SendTextMessage(ctx, p.JID, msg)
	}

	enteredPass := strings.TrimSpace(p.Text)
	if r.cfg.AdminPassword != "" && subtle.ConstantTimeCompare([]byte(enteredPass), []byte(r.cfg.AdminPassword)) == 1 {
		r.adminMu.Lock()
		r.adminFailedAttempts[phone] = 0
		r.adminMu.Unlock()

		log.Printf("[ADMIN AUTH] Admin %s berhasil login", whatsapp.MaskJID(phone))
		s.CurrentState = StateAdminLoggedIn
		return r.SendAdminDashboard(ctx, s, p)
	}

	// Password wrong
	r.adminMu.Lock()
	r.adminFailedAttempts[phone]++
	attempts := r.adminFailedAttempts[phone]
	maxRetry := r.cfg.AdminMaxRetry
	if maxRetry <= 0 {
		maxRetry = 3
	}

	if attempts >= maxRetry {
		lockoutDuration := time.Duration(r.cfg.AdminLockoutMinutes) * time.Minute
		if lockoutDuration <= 0 {
			lockoutDuration = 60 * time.Minute
		}
		r.adminLockedUntil[phone] = time.Now().Add(lockoutDuration)
		r.adminFailedAttempts[phone] = 0
		r.adminMu.Unlock()

		log.Printf("[ADMIN LOCKOUT] Nomor %s dikunci selama %v karena gagal password %d kali", whatsapp.MaskJID(phone), lockoutDuration, attempts)
		s.ResetToRoot()
		msg := fmt.Sprintf("*AKSES DITOLAK*\n"+
			"BPS Kabupaten Bangka Barat\n"+
			"━━━━━━━━━━━━━━━━━━\n\n"+
			"Anda telah gagal memasukkan password sebanyak *%d kali* berturut-turut.\n"+
			"Nomor Anda dikunci selama *%d menit* demi keamanan sistem.\n\n"+
			"━━━━━━━━━━━━━━━━━━\n"+
			"• Ketik *0* untuk kembali ke Menu Utama", attempts, r.cfg.AdminLockoutMinutes)
		return r.sender.SendTextMessage(ctx, p.JID, msg)
	}
	r.adminMu.Unlock()

	log.Printf("[ADMIN AUTH FAIL] Nomor %s salah memasukkan password (%d/%d)", whatsapp.MaskJID(phone), attempts, maxRetry)
	msg := fmt.Sprintf("*PASSWORD ADMIN SALAH*\n"+
		"BPS Kabupaten Bangka Barat\n"+
		"━━━━━━━━━━━━━━━━━━\n\n"+
		"Password yang Anda masukkan tidak sesuai.\n"+
		"Sisa percobaan: *%d/%d*.\n\n"+
		"━━━━━━━━━━━━━━━━━━\n"+
		"• Masukkan kembali password yang benar\n"+
		"• Ketik *0* untuk membatalkan dan kembali ke Menu Utama", attempts, maxRetry)
	return r.sender.SendTextMessage(ctx, p.JID, msg)
}

// SendAdminDashboard displays the full admin panel options
func (r *BotRouter) SendAdminDashboard(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.CurrentState = StateAdminLoggedIn

	var sb strings.Builder
	sb.WriteString("*PANEL KENDALI ADMIN*\n")
	sb.WriteString("BPS Kabupaten Bangka Barat\n\n")
	sb.WriteString("Halo Admin BPS Kabupaten Bangka Barat.\n")
	sb.WriteString("Silakan pilih menu pengelolaan sistem bot:\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━\n\n")
	sb.WriteString("1. *Kelola Narahubung Para Pakar (SQLite)*\n")
	sb.WriteString("   • Lihat & ubah kontak WhatsApp 8 bidang pakar\n\n")
	sb.WriteString("2. *Status Scraper & Usia Data*\n")
	sb.WriteString("   • Cek volume data tersimpan, jadwal & circuit breaker\n\n")
	sb.WriteString("3. *Jalankan Scraping Manual BPS*\n")
	sb.WriteString("   • Memicu scraping on-demand portal Bangka Barat\n\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━\n\n")
	sb.WriteString("Balas dengan *angka menu (1 - 3)*\n")
	sb.WriteString("• Ketik *0* untuk keluar dari Mode Admin")

	items := getAdminMenuItems()
	msg := whatsapp.BuildInteractiveList(
		"",
		sb.String(),
		"",
		"Pilih Menu Admin",
		[]whatsapp.ListSection{
			{Title: "Menu Pengelolaan Admin", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleUnavailable sends fallback text with Quick Reply buttons
func (r *BotRouter) HandleUnavailable(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	body := "Mohon maaf Sahabat Data, pilihan yang Anda masukkan tidak tersedia atau belum dikenali.\n\nSilakan pilih salah satu tombol navigasi di bawah atau ketik angka menu:\n*0.* Menu Utama\n*99.* Menu Sebelumnya"

	msg := whatsapp.BuildQuickReplyButtons(
		"Pilihan Tidak Ditemukan",
		body,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "Menu Sebelumnya (99)"},
			{ID: "0", DisplayText: "Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleIdleTimeout sends SKD survey outro after 15 min idle
func (r *BotRouter) HandleIdleTimeout(ctx context.Context, s *UserSession) {
	log.Printf("[TIMEOUT] Sesi idle %d menit berakhir untuk %s. Mengirim Survei SKD dan penutup.", r.cfg.IdleTimeoutMinutes, whatsapp.MaskJID(s.JID.User))

	outro := locale.T("outro")
	if outro != "" {
		_ = r.sender.SendTextMessage(ctx, s.JID, outro)
	}

	outro2 := locale.T("outro2")
	if outro2 != "" {
		_ = r.sender.SendTextMessage(ctx, s.JID, outro2)
	}

	r.sessions.Delete(s.JID)
}
