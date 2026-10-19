package bot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"bot-istimewa/internal/locale"
	"bot-istimewa/internal/whatsapp"
)

func getRootMenuItems() []whatsapp.ListRow {
	return []whatsapp.ListRow{
		{ID: "1", Title: "1️⃣  Layanan Data", Description: "Statistik Sosial, Ekonomi & Pertanian"},
		{ID: "2", Title: "2️⃣  Indikator Strategis", Description: "Kemiskinan, Inflasi, IPM, PDRB, dll"},
		{ID: "3", Title: "3️⃣  Produk Statistik", Description: "BRS, Publikasi, Data Mikro"},
		{ID: "4", Title: "4️⃣  Video Literasi Statistik", Description: "Video edukasi dan pemahaman statistik"},
		{ID: "5", Title: "5️⃣  Infografik", Description: "Visualisasi ringkas data BPS"},
		{ID: "6", Title: "6️⃣  Konsultasi Statistik", Description: "Konsultasi online dengan pakar BPS"},
		{ID: "7", Title: "7️⃣  Berlangganan Gratis", Description: "Langganan rilis publikasi berkala"},
		{ID: "8", Title: "8️⃣  Pengaduan & Aspirasi", Description: "Sampaikan kritik dan saran layanan"},
		{ID: "9", Title: "9️⃣  Hubungi Developer", Description: "Kontak pengembang bot"},
		{ID: "10", Title: "🔟  FAQs", Description: "Pertanyaan yang sering diajukan"},
		{ID: "11", Title: "1️⃣1️⃣  Tentang BPS", Description: "Profil BPS Kabupaten Bangka Barat"},
	}
}

// HandleWelcomeFull handles first message / keyword 100
func (r *BotRouter) HandleWelcomeFull(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.CurrentState = StateRoot

	// 1. Send sticker selamat datang
	_ = r.sender.SendSticker(ctx, p.JID, "StikerChatbot/IKON ISTIMEWA - selamat datang new.webp")

	// 2. Send single-bubble interactive message
	items := getRootMenuItems()
	userName := strings.TrimSpace(p.Name)
	var greeting string
	if userName != "" {
		greeting = fmt.Sprintf("Halo Sahabat Data *%s* 👋", userName)
	} else {
		greeting = "Halo Sahabat Data 👋"
	}
	introText := fmt.Sprintf("🌟 *Selamat datang di ISTIMEWA* 🌟\nBPS Kabupaten Bangka Barat\n\n%s\nPerkenalkan saya *Ak Isti* 🙋‍♀️, bot pintar yang siap memandu Sahabat menjelajahi dunia data.\n\n📋 *Pilih nomor layanan* di bawah ini:", greeting)
	body := whatsapp.BuildBodyText(introText, items, "\n◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇\n```Statistik Bermakna, Statistik Berdampak```")

	msg := whatsapp.BuildInteractiveList(
		"🤖 BOT ISTIMEWA — BPS Bangka Barat",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Nomor Layanan 📋",
		[]whatsapp.ListSection{
			{Title: "Menu Layanan Utama", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleWelcomeShort handles returning to main menu (keyword 0)
func (r *BotRouter) HandleWelcomeShort(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.CurrentState = StateRoot

	// Send sticker siap
	_ = r.sender.SendSticker(ctx, p.JID, "StikerChatbot/IKON ISTIMEWA - siap new.webp")

	items := getRootMenuItems()
	userName := strings.TrimSpace(p.Name)
	var greeting string
	if userName != "" {
		greeting = fmt.Sprintf("👋 Hai Sahabat Data *%s*!", userName)
	} else {
		greeting = "👋 Hai Sahabat Data!"
	}
	introText := fmt.Sprintf("%s\nYuk kita lanjut berselancar menjelajahi dunia data 🏄‍♀️\n\n📋 *Pilih nomor layanan* di bawah ini:", greeting)
	body := whatsapp.BuildBodyText(introText, items, "\n◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇\n```Statistik Bermakna, Statistik Berdampak```")

	msg := whatsapp.BuildInteractiveList(
		"🤖 BOT ISTIMEWA — BPS Bangka Barat",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Nomor Layanan 📋",
		[]whatsapp.ListSection{
			{Title: "Menu Layanan Utama", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleReset clears the user session state
func (r *BotRouter) HandleReset(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	r.sessions.Delete(p.JID)
	text := "🔄 Sesi percakapan Anda telah berhasil direset.\nKetik pesan apa saja atau ketik *0* untuk kembali ke menu utama ISTIMEWA."
	return r.sender.SendTextMessage(ctx, p.JID, text)
}

// HandleBack pops state from stack and re-executes previous menu
func (r *BotRouter) HandleBack(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	prevState := s.PopState()
	log.Printf("[NAVIGASI] Tombol Kembali (99) ditekan oleh %s -> Kembali ke state: %s", p.JID.User, prevState)

	switch prevState {
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
	case StateMenu4:
		return r.HandleMenu4_Video(ctx, s, p)
	case StateMenu5:
		return r.HandleMenu5_Infografik(ctx, s, p)
	case StateMenu6:
		return r.HandleMenu6_Konsultasi(ctx, s, p)
	case StateMenu6_Pakar:
		return r.HandleKonsultasiPakarList(ctx, s, p)
	default:
		return r.HandleWelcomeShort(ctx, s, p)
	}
}

// HandleAdminPrompt asks admin password
func (r *BotRouter) HandleAdminPrompt(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	return r.sender.SendTextMessage(ctx, p.JID, "🔐 Silakan masukkan *Password Admin* untuk melanjutkan:")
}

// HandleAdminAuth verifies admin password from .env and displays admin database links
func (r *BotRouter) HandleAdminAuth(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	if strings.TrimSpace(p.Text) == r.cfg.AdminPassword {
		s.CurrentState = StateAdminLoggedIn
		return r.SendAdminDashboard(ctx, s, p)
	}

	return r.sender.SendTextMessage(ctx, p.JID, "❌ Password admin salah! Akses ditolak.\nKetik *0* untuk kembali ke menu utama.")
}

// SendAdminDashboard displays the full 8 database management links for administrators
func (r *BotRouter) SendAdminDashboard(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	adminText := fmt.Sprintf("🛠️ *PANEL ADMIN BOT ISTIMEWA*\n\n" +
		"Selamat datang di Panel Admin BPS Kabupaten Bangka Barat.\n" +
		"Berikut tautan pengelolaan database Google Sheets:\n\n" +
		"*1.* %s\n\n" +
		"*2.* %s\n\n" +
		"*3.* %s\n\n" +
		"*4.* %s\n\n" +
		"*5.* %s\n\n" +
		"*6.* %s\n\n" +
		"*7.* %s\n\n" +
		"*8.* %s\n\n" +
		"───────────────────────────\n" +
		"*Ketik angka (1-8)* untuk mengambil link database tertentu.\n" +
		"*Ketik 0* untuk keluar dari Mode Admin dan kembali ke Menu Utama.",
		locale.T("menuAdmin1"),
		locale.T("menuAdmin2"),
		locale.T("menuAdmin3"),
		locale.T("menuAdmin4"),
		locale.T("menuAdmin5"),
		locale.T("menuAdmin6"),
		locale.T("menuAdmin7"),
		locale.T("menuAdmin8"),
	)
	return r.sender.SendTextMessage(ctx, p.JID, adminText)
}

// HandleUnavailable sends fallback text with Quick Reply buttons
func (r *BotRouter) HandleUnavailable(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	body := "🙆‍♀️ Mohon maaf Sahabat Data, pilihan yang Anda masukkan tidak tersedia atau belum dikenali.\n\nSilakan pilih salah satu tombol navigasi di bawah atau ketik angka menu:\n*0.* 📑 Menu Utama\n*99.* ↩️ Menu Sebelumnya"

	msg := whatsapp.BuildQuickReplyButtons(
		"⚠️ Pilihan Tidak Ditemukan",
		body,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "↩️ Menu Sebelumnya (99)"},
			{ID: "0", DisplayText: "🏠 Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleIdleTimeout sends SKD survey outro and closing stickers after 15 min idle
func (r *BotRouter) HandleIdleTimeout(ctx context.Context, s *UserSession) {
	log.Printf("[TIMEOUT] Sesi idle 15 menit berakhir untuk %s. Mengirim Survei SKD dan penutup.", s.JID.User)

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
