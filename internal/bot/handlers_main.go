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
		{ID: "3", Title: "Produk Statistik", Description: "BRS, Publikasi, Data Mikro"},
		{ID: "4", Title: "Video Literasi Statistik", Description: "Video edukasi dan pemahaman statistik"},
		{ID: "5", Title: "Infografik", Description: "Visualisasi ringkas data BPS"},
		{ID: "6", Title: "Konsultasi Statistik", Description: "Konsultasi online dengan pakar BPS"},
		{ID: "7", Title: "Berlangganan Gratis", Description: "Langganan rilis publikasi berkala"},
		{ID: "8", Title: "Pengaduan & Aspirasi", Description: "Sampaikan kritik dan saran layanan"},
		{ID: "9", Title: "Hubungi Developer", Description: "Kontak pengembang bot"},
		{ID: "10", Title: "FAQs", Description: "Pertanyaan yang sering diajukan"},
		{ID: "11", Title: "Tentang BPS", Description: "Profil BPS Kabupaten Bangka Barat"},
	}
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

	// 2. Send single-bubble interactive message
	items := getRootMenuItems()
	userName := strings.TrimSpace(p.Name)
	var greeting string
	if userName != "" {
		greeting = fmt.Sprintf("Halo Sahabat Data *%s* ", userName)
	} else {
		greeting = "Halo Sahabat Data "
	}
	introText := fmt.Sprintf("*Selamat datang di ISTIMEWA* \nBPS Kabupaten Bangka Barat\n\n%s\nPerkenalkan saya *Ak Isti* , bot pintar yang siap memandu Sahabat menjelajahi dunia data.\n\n*Pilih nomor layanan* di bawah ini:", greeting)
	body := whatsapp.BuildBodyText(introText, items, "\n◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇\n```Statistik Bermakna, Statistik Berdampak```")

	msg := whatsapp.BuildInteractiveList(
		"BOT ISTIMEWA — BPS Bangka Barat",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Nomor Layanan ",
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

	items := getRootMenuItems()
	userName := strings.TrimSpace(p.Name)
	var greeting string
	if userName != "" {
		greeting = fmt.Sprintf("Hai Sahabat Data *%s*!", userName)
	} else {
		greeting = "Hai Sahabat Data!"
	}
	introText := fmt.Sprintf("%s\nYuk kita lanjut berselancar menjelajahi dunia data \n\n*Pilih nomor layanan* di bawah ini:", greeting)
	body := whatsapp.BuildBodyText(introText, items, "\n◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇\n```Statistik Bermakna, Statistik Berdampak```")

	msg := whatsapp.BuildInteractiveList(
		"BOT ISTIMEWA — BPS Bangka Barat",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Nomor Layanan ",
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
