package bot

import (
	"context"
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
		{ID: "7", Title: "Langganan Publikasi", Description: "Langganan rilis berkala gratis"},
		{ID: "8", Title: "Pengaduan & Saran", Description: "Sampaikan kritik dan saran layanan"},
		{ID: "9", Title: "Kontak Developer", Description: "Kontak pengembang bot"},
		{ID: "10", Title: "Bantuan & FAQ", Description: "Pertanyaan yang sering diajukan"},
		{ID: "11", Title: "Tentang BPS", Description: "Profil BPS Kabupaten Bangka Barat"},
	}
}

func getTimeGreeting() string {
	loc, err := time.LoadLocation("Asia/Jakarta")
	now := time.Now()
	if err == nil {
		now = now.In(loc)
	}
	hour := now.Hour()
	if hour >= 4 && hour < 11 {
		return "Selamat Pagi"
	} else if hour >= 11 && hour < 15 {
		return "Selamat Siang"
	} else if hour >= 15 && hour < 18 {
		return "Selamat Sore"
	}
	return "Selamat Malam"
}

func buildWelcomeMenuBody(greetingLine, introSubtitle string) string {
	var sb strings.Builder
	sb.WriteString("🟢 *BOT ISTIMEWA*\n")
	sb.WriteString("_BPS Kabupaten Bangka Barat_\n\n")
	sb.WriteString(greetingLine + "\n")
	sb.WriteString(introSubtitle + "\n\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("📊 *DATA & INDIKATOR*\n")
	sb.WriteString("1️⃣  Layanan Data\n")
	sb.WriteString("2️⃣  Indikator Strategis\n")
	sb.WriteString("3️⃣  BRS & Publikasi Rilis\n\n")
	sb.WriteString("🎨 *KONTEN & VISUAL*\n")
	sb.WriteString("4️⃣  Video Edukasi Literasi\n")
	sb.WriteString("5️⃣  Galeri Infografis Daerah\n\n")
	sb.WriteString("💬 *KONSULTASI & LAYANAN*\n")
	sb.WriteString("6️⃣  Tanya Pakar\n")
	sb.WriteString("7️⃣  Langganan Publikasi\n")
	sb.WriteString("8️⃣  Pengaduan & Saran\n\n")
	sb.WriteString("⚙️ *INFORMASI*\n")
	sb.WriteString("9️⃣  Kontak Developer\n")
	sb.WriteString("🔟  Bantuan & FAQ\n")
	sb.WriteString("🏛️  Ketik *11* untuk Tentang BPS\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━\n\n")
	sb.WriteString("⌨️ Balas dengan *angka menu*\n")
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

	// 1. Send sticker selamat datang
	_ = r.sender.SendSticker(ctx, p.JID, "StikerChatbot/IKON ISTIMEWA - selamat datang new.webp")

	// 2. Format personalized greeting with time & name (PRD §6.1 F-01 & §14)
	userName := strings.TrimSpace(p.Name)
	timeGreeting := getTimeGreeting()
	var greetingLine string
	if userName != "" {
		greetingLine = fmt.Sprintf("%s, *%s*! 👋", timeGreeting, userName)
	} else {
		greetingLine = fmt.Sprintf("%s, Sahabat Data! 👋", timeGreeting)
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

	// Send sticker siap
	_ = r.sender.SendSticker(ctx, p.JID, "StikerChatbot/IKON ISTIMEWA - siap new.webp")

	userName := strings.TrimSpace(p.Name)
	var greetingLine string
	if userName != "" {
		greetingLine = fmt.Sprintf("Hai kembali, *%s*! 👋", userName)
	} else {
		greetingLine = "Hai kembali, Sahabat Data! 👋"
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
	log.Printf("[NAVIGASI] Tombol Kembali (99) ditekan oleh %s -> Kembali ke state: %s", p.JID.User, prevState)

	switch prevState {
	case StateAdminLoggedIn, StateAdminManagePakar:
		s.CurrentState = StateAdminLoggedIn
		return r.SendAdminDashboard(ctx, s, p)
	case StateRoot:
		return r.HandleWelcomeShort(ctx, s, p)
	case StateMenu1:
		return r.HandleMenu1_LayananData(ctx, s, p)
	case StateMenu1_Sosduk:
		return r.HandleLayananSub_Sosduk(ctx, s, p)
	case StateMenu1_Ekonomi:
		return r.HandleLayananSub_Ekonomi(ctx, s, p)
	case StateMenu1_Pertanian:
		return r.HandleLayananSub_Pertanian(ctx, s, p)
	case StateMenu2:
		return r.HandleMenu2_Indikator(ctx, s, p)
	case StateMenu3:
		return r.HandleMenu3_Produk(ctx, s, p)
	case StateMenu4, StateMenu4_Sosduk, StateMenu4_Ekonomi:
		return r.HandleMenu4_Video(ctx, s, p)
	case StateMenu5:
		return r.HandleMenu5_Infografik(ctx, s, p)
	case StateMenu6, StateMenu6_Booking:
		return r.HandleMenu6_Konsultasi(ctx, s, p)
	case StateMenu6_Pakar:
		return r.HandleKonsultasiPakarList(ctx, s, p)
	default:
		return r.HandleWelcomeShort(ctx, s, p)
	}
}

// HandleAdminPrompt asks admin password
func (r *BotRouter) HandleAdminPrompt(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	return r.sender.SendTextMessage(ctx, p.JID, "Silakan masukkan *Password Admin* untuk melanjutkan:")
}

// HandleAdminAuth verifies admin password from .env, applies brute-force lockout, and auto-expires admin session
func (r *BotRouter) HandleAdminAuth(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	phone := p.JID.User

	// Check lockout
	if locked, until := r.isAdminLockedOut(phone); locked {
		s.ResetToRoot()
		msg := fmt.Sprintf("Nomor Anda sedang dikunci dari percobaan admin hingga %s.\nSilakan coba lagi nanti atau ketik *0* untuk ke menu utama.", until.Format("15:04:05"))
		return r.sender.SendTextMessage(ctx, p.JID, msg)
	}

	enteredPass := strings.TrimSpace(p.Text)
	if enteredPass == r.cfg.AdminPassword {
		r.adminMu.Lock()
		r.adminFailedAttempts[phone] = 0
		r.adminMu.Unlock()

		log.Printf("[ADMIN AUTH] Admin %s berhasil login", phone)
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

		log.Printf("[ADMIN LOCKOUT] Nomor %s dikunci selama %v karena gagal password %d kali", phone, lockoutDuration, attempts)
		s.ResetToRoot()
		return r.sender.SendTextMessage(ctx, p.JID, fmt.Sprintf("Akses ditolak! Anda telah gagal %d kali berturut-turut.\nNomor Anda dikunci selama %d menit dari percobaan login admin.\nKetik *0* untuk kembali ke menu utama.", attempts, r.cfg.AdminLockoutMinutes))
	}
	r.adminMu.Unlock()

	log.Printf("[ADMIN AUTH FAIL] Nomor %s salah memasukkan password (%d/%d)", phone, attempts, maxRetry)
	return r.sender.SendTextMessage(ctx, p.JID, fmt.Sprintf("Password admin salah! Percobaan %d/%d.\nKetik *0* untuk membatalkan dan kembali ke menu utama.", attempts, maxRetry))
}

// SendAdminDashboard displays the full admin panel options
func (r *BotRouter) SendAdminDashboard(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.CurrentState = StateAdminLoggedIn
	adminText := "*PANEL KENDALI ADMIN — BOT ISTIMEWA BPS*\n\n" +
		"Halo Admin BPS Kabupaten Bangka Barat. Silakan pilih menu pengelolaan:\n\n" +
		"*1.* 👥 *Kelola Narahubung Para Pakar (SQLite)*\n" +
		"     _Lihat & ubah nomor WhatsApp / nama penanggung jawab 8 bidang pakar._\n\n" +
		"*2.* ⚙ *Status Scraper & Usia Data*\n" +
		"     _Cek jumlah data di database, usia data scraper, dan circuit breaker._\n\n" +
		"*3.* 🔄 *Jalankan Scraping Manual BPS Sekarang*\n" +
		"     _Memicu scraping on-demand pembaruan data publikasi, rilis & tabel._\n\n" +
		"*4.* 📊 *Tautan Database Google Sheets Legacy*\n" +
		"     _Akses link spreadsheet internal (1-8)._\n\n" +
		"───────────────────────────\n" +
		"*Ketik angka (1-4)* untuk memilih menu.\n" +
		"*Ketik 0* untuk keluar dari Mode Admin dan kembali ke Menu Utama."
	return r.sender.SendTextMessage(ctx, p.JID, adminText)
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

// HandleIdleTimeout sends SKD survey outro and closing stickers after 15 min idle
func (r *BotRouter) HandleIdleTimeout(ctx context.Context, s *UserSession) {
	log.Printf("[TIMEOUT] Sesi idle %d menit berakhir untuk %s. Mengirim Survei SKD dan penutup.", r.cfg.IdleTimeoutMinutes, s.JID.User)

	outro := locale.T("outro")
	if outro != "" {
		_ = r.sender.SendTextMessage(ctx, s.JID, outro)
	}

	_ = r.sender.SendSticker(ctx, s.JID, "StikerChatbot/IKON ISTIMEWA - terimakasih.webp")

	outro2 := locale.T("outro2")
	if outro2 != "" {
		_ = r.sender.SendTextMessage(ctx, s.JID, outro2)
	}

	_ = r.sender.SendSticker(ctx, s.JID, "StikerChatbot/IKON ISTIMEWA-love data.webp")

	r.sessions.Delete(s.JID)
}
