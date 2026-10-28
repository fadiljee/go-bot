package bot

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/types"

	"bot-istimewa/internal/gsheet"
	"bot-istimewa/internal/scraper"
	"bot-istimewa/internal/whatsapp"
)

// HandleMenu6_Konsultasi renders Konsultasi Statistik menu
func (r *BotRouter) HandleMenu6_Konsultasi(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu6)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "1️⃣  Ajukan Jadwal Konsultasi", Description: "Reservasi jadwal konsultasi tatap muka/online"},
		{ID: "2", Title: "2️⃣  Tanya Para Pakar", Description: "Hubungi pakar statistik BPS langsung"},
		{ID: "3", Title: "3️⃣  Informasi PST & Jam Layanan", Description: "Jadwal operasional & kontak petugas PST"},
	}

	introText := "👩‍🏫 *Konsultasi Statistik*\nAda yang ingin Sahabat tanyakan seputar data & statistik? 🙋‍♀️\n\n●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●\n\n📋 *Pilih jenis layanan konsultasi:*"
	body := whatsapp.BuildBodyText(introText, items, "\n*0.* 📑 Menu Awal\n\n◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇\n```Statistik Bermakna, Statistik Berdampak```")

	msg := whatsapp.BuildInteractiveList(
		"👩‍🏫 Konsultasi Statistik",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Layanan Konsultasi 💬",
		[]whatsapp.ListSection{
			{Title: "Layanan Konsultasi", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleMenu6_BookingPrompt prompts user to submit consultation booking details
func (r *BotRouter) HandleMenu6_BookingPrompt(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu6_Booking)

	prompt := "📅 *Pengajuan Jadwal Konsultasi Statistik*\n\n" +
		"Silakan kirimkan data permohonan konsultasi Sahabat dengan membalas pesan ini menggunakan format berikut:\n\n" +
		"```Nama:\nInstansi / Asal:\nTopik / Kebutuhan Data:\nRencana Hari / Jam:```\n\n" +
		"📌 *Contoh Pengisian*:\n" +
		"Nama: Ahmad Fauzi\n" +
		"Instansi: Dinas Pertanian Bangka Barat\n" +
		"Topik: Data Produksi Padi & Sawit 2025\n" +
		"Rencana Hari / Jam: Kamis, 29 Okt jam 10.00 WIB\n\n" +
		"*(Atau cukup ketik informasi di atas secara bebas)*\n\n" +
		"Ketik *0* untuk batal dan kembali ke Menu Utama."

	return r.sender.SendTextMessage(ctx, p.JID, prompt)
}

// HandleProcessBooking processes user booking input, stores to SQLite, notifies user, and alerts office group
func (r *BotRouter) HandleProcessBooking(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	input := strings.TrimSpace(p.Text)
	if input == "" {
		return r.sender.SendTextMessage(ctx, p.JID, "⚠️ Mohon ketikkan data permohonan konsultasi Sahabat, atau ketik *0* untuk kembali.")
	}

	// Simple parser for key-value or lines
	lines := strings.Split(input, "\n")
	var name, instansi, topic, preferredTime string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "nama:") {
			name = strings.TrimSpace(trimmed[5:])
		} else if strings.HasPrefix(lower, "instansi:") || strings.HasPrefix(lower, "asal:") || strings.HasPrefix(lower, "instansi / asal:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				instansi = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(lower, "topik:") || strings.HasPrefix(lower, "kebutuhan data:") || strings.HasPrefix(lower, "topik / kebutuhan data:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				topic = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(lower, "rencana hari / jam:") || strings.HasPrefix(lower, "rencana hari/jam:") || strings.HasPrefix(lower, "jadwal:") || strings.HasPrefix(lower, "waktu:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				preferredTime = strings.TrimSpace(parts[1])
			}
		}
	}

	// Fallback if user submitted freeform single line or text without labels
	if name == "" {
		if p.Name != "" {
			name = p.Name
		} else {
			name = "Sahabat Data (" + p.JID.User + ")"
		}
	}
	if instansi == "" {
		instansi = "Masyarakat / Umum"
	}
	if topic == "" {
		if len(lines) > 0 && lines[0] != "" {
			topic = lines[0]
		} else {
			topic = input
		}
	}
	if preferredTime == "" {
		preferredTime = "Hari & jam kerja PST (08.00 - 15.30 WIB)"
	}

	// Generate ticket number: e.g. PST-261028-123
	ticketNo := fmt.Sprintf("PST-%s-%03d", time.Now().Format("060102"), rand.Intn(900)+100)

	booking := &scraper.ConsultationBooking{
		TicketNo:      ticketNo,
		UserJID:       p.JID.String(),
		UserName:      name,
		Institution:   instansi,
		Topic:         topic,
		PreferredTime: preferredTime,
		Notes:         input,
		CreatedAt:     time.Now(),
	}

	// 1. Save to SQLite DB
	if r.scraper != nil {
		if err := r.scraper.CreateConsultationBooking(ctx, booking); err != nil {
			log.Printf("⚠️ [BOOKING] Gagal menyimpan ke SQLite: %v", err)
		}
	}

	// 2. Reply confirmation to user
	userReply := fmt.Sprintf("✅ *Pengajuan Jadwal Konsultasi Diterima!*\n\n"+
		"Terima kasih Bapak/Ibu *%s*, permohonan konsultasi statistik Anda telah berhasil dicatat dan diteruskan ke Tim Pelayanan Statistik Terpadu (PST) BPS Kabupaten Bangka Barat.\n\n"+
		"📋 *Rincian Tiket Konsultasi*:\n"+
		"• 🎫 *No. Tiket*: `%s`\n"+
		"• 👤 *Nama*: %s\n"+
		"• 🏢 *Instansi*: %s\n"+
		"• 📊 *Topik Data*: %s\n"+
		"• ⏰ *Rencana Jadwal*: %s\n\n"+
		"Petugas PST atau pakar terkait kami akan menghubungi Anda pada kontak WhatsApp ini untuk konfirmasi persiapan data. Sampai jumpa di PST BPS Bangka Barat! 🙏\n\n"+
		"Ketik *0* untuk kembali ke Menu Utama.",
		name, ticketNo, name, instansi, topic, preferredTime)

	_ = r.sender.SendTextMessage(ctx, p.JID, userReply)

	// 3. Auto-Forward Notification to Office Internal Group
	if r.cfg.InternalPSTGroupJID != "" {
		groupJID, err := types.ParseJID(r.cfg.InternalPSTGroupJID)
		if err == nil {
			groupMsg := fmt.Sprintf("🔔 *PEMBERITAHUAN JADWAL KONSULTASI BARU (PST BPS)*\n\n"+
				"Halo Tim PST & Para Pakar BPS Bangka Barat, ada permohonan konsultasi baru masuk melalui WhatsApp Bot:\n\n"+
				"🎫 *No. Tiket*: `%s`\n"+
				"👤 *Nama Pemohon*: %s\n"+
				"🏢 *Instansi / Asal*: %s\n"+
				"📱 *Kontak WhatsApp*: https://wa.me/%s\n"+
				"📊 *Bidang / Topik*: %s\n"+
				"⏰ *Rencana Jadwal*: %s\n"+
				"🗓️ *Waktu Masuk*: %s WIB\n\n"+
				"📝 *Pesan Lengkap*:\n_%s_\n\n"+
				"📌 *Dimohon petugas piket PST dan pakar terkait untuk menindaklanjuti dan mempersiapkan bahan data. Terima kasih!* 🤝",
				ticketNo, name, instansi, p.JID.User, topic, preferredTime, time.Now().Format("02 Jan 2006 15:04"), input)

			if sendErr := r.sender.SendTextMessage(ctx, groupJID, groupMsg); sendErr != nil {
				log.Printf("❌ [PST GROUP] Gagal mengirim notifikasi ke grup %s: %v", r.cfg.InternalPSTGroupJID, sendErr)
			} else {
				log.Printf("📢 [PST GROUP] Sukses meneruskan jadwal %s ke grup internal kantor %s", ticketNo, r.cfg.InternalPSTGroupJID)
			}
		} else {
			log.Printf("⚠️ [PST GROUP] Format INTERNAL_PST_GROUP_JID (%s) tidak valid: %v", r.cfg.InternalPSTGroupJID, err)
		}
	} else {
		log.Printf("ℹ️ [PST GROUP] INTERNAL_PST_GROUP_JID belum diisi di .env; notifikasi grup dilewati.")
	}

	// Reset state back to root after booking
	s.ResetToRoot()
	return nil
}

func (r *BotRouter) HandleKonsultasiUmum(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	piketText := "🤝 *Petugas Pelayanan Statistik Terpadu (PST) BPS Bangka Barat*\n\nPetugas PST siap membantu menjawab pertanyaan Sahabat seputar data, publikasi, dan rujukan statistik.\n\n⏰ *Jam Pelayanan*:\nSenin - Kamis: 08.00 - 15.30 WIB\nJumat: 08.00 - 15.00 WIB\n\nSilakan klik link berikut untuk terhubung langsung via WhatsApp:\n📞 https://wa.me/6282164302643?text=Halo%20Petugas%20PST%20BPS%20Bangka%20Barat"

	msg := whatsapp.BuildQuickReplyButtons(
		"👩‍💼 Petugas Konsultasi Umum",
		piketText,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "↩️ Menu Sebelumnya (99)"},
			{ID: "0", DisplayText: "🏠 Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

func (r *BotRouter) HandleKonsultasiPakarList(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu6_Pakar)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "1️⃣  Statistik Sosial Kependudukan", Description: "Kependudukan, Kemiskinan, Pendidikan"},
		{ID: "2", Title: "2️⃣  Statistik Pertanian", Description: "Perkebunan, Kehutanan, Perikanan"},
		{ID: "3", Title: "3️⃣  Statistik Ekonomi", Description: "Neraca, PDRB, Industri"},
		{ID: "4", Title: "4️⃣  Distribusi, Harga & KTIP", Description: "Inflasi, Pariwisata, Perdagangan"},
		{ID: "5", Title: "5️⃣  Statistik Sektoral", Description: "Data OPD & Rekomendasi Statistik"},
		{ID: "6", Title: "6️⃣  Statistik Khusus", Description: "Topik data tematik & khusus"},
		{ID: "7", Title: "7️⃣  Metodologi Statistik", Description: "Desain sampling & pengumpulan data"},
		{ID: "8", Title: "8️⃣  Diseminasi & Layanan", Description: "Publikasi, website & PST"},
	}

	introText := "🧑‍💼 *Pakar Statistik BPS Bangka Barat*\nSilakan pilih bidang kepakaran yang ingin Sahabat konsultasikan:"
	body := whatsapp.BuildBodyText(introText, items, "\n*0.* 📑 Menu Awal\n\n◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇\n```Statistik Bermakna, Statistik Berdampak```")

	msg := whatsapp.BuildInteractiveList(
		"🧑‍💼 Tanya Para Pakar",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Bidang Pakar 👨‍💼",
		[]whatsapp.ListSection{
			{Title: "Bidang Kepakaran", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

func (r *BotRouter) HandlePakarDetail(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage, idx int) error {
	pakars := gsheet.DefaultPakar()
	var pakar *gsheet.PakarItem
	noStr := strconv.Itoa(idx)
	for i := range pakars {
		if pakars[i].No.String() == noStr {
			pakar = &pakars[i]
			break
		}
	}

	var pakarText string
	var pakarImg string
	if pakar != nil {
		pakarText = fmt.Sprintf("🤝 Perkenalkan Sahabat Data, ini adalah *%s*\n\n%s\n\nBeliau akan membantu menjawab pertanyaan Sahabat Data.\n\nMohon diingat untuk menghubungi beliau hanya pada waktu kerja dan hari kerja yah sobat data 🙆‍♀️\n\nSilakan hubungi melalui link WhatsApp berikut:\n📞 %s", pakar.Nama, pakar.Penjelasan, pakar.Link)
		pakarImg = pakar.Foto
	} else {
		pakarText = fmt.Sprintf("🤝 Pakar bidang %d sedang tidak bertugas. Silakan hubungi admin di menu 9.", idx)
	}

	if pakarImg != "" {
		_ = r.sender.SendImageFromURL(ctx, p.JID, pakarImg, "Profil Pakar Statistik")
	}

	msg := whatsapp.BuildQuickReplyButtons(
		"👨‍🏫 Profil Pakar Statistik",
		pakarText,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "↩️ Pilih Pakar Lain (99)"},
			{ID: "0", DisplayText: "🏠 Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleMenu7_Berlangganan renders subscription menu
func (r *BotRouter) HandleMenu7_Berlangganan(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	text := "📰 *Berlangganan Gratis Rilis Data BPS*\n\nDapatkan pemberitahuan rilis publikasi terbaru, Berita Resmi Statistik (BRS), dan ringkasan data strategis Kabupaten Bangka Barat langsung ke nomor WhatsApp Sahabat!\n\n📋 *Formulir Berlangganan*:\nhttps://s.bps.go.id/LanggananDataBabar\n\n🔒 _Kerahasiaan data nomor Anda dijamin oleh Undang-Undang No. 16 Tahun 1997 tentang Statistik._"

	msg := whatsapp.BuildQuickReplyButtons(
		"📰 Layanan Berlangganan",
		text,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "0", DisplayText: "🏠 Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleMenu8_Pengaduan renders complaint & aspiration channels
func (r *BotRouter) HandleMenu8_Pengaduan(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	text := "📢 *Pengaduan & Aspirasi Layanan*\n\nBPS Kabupaten Bangka Barat berkomitmen memberikan pelayanan prima yang bersih, transparan, dan bebas dari pungutan liar (Zona Integritas WBK/WBBM).\n\nJika ada kritik, keluhan, atau laporan ketidaksesuaian pelayanan, silakan sampaikan melalui:\n\n1. 🌐 *SP4N-LAPOR!*:\nhttps://www.lapor.go.id/\n\n2. 📧 *Email Pengaduan*:\nbps1903@bps.go.id\n\n3. 🏢 *Kotak Pengaduan PST*:\nKantor BPS Kabupaten Bangka Barat\n\n_Setiap laporan yang masuk akan ditindaklanjuti secara profesional dan rahasia._"

	msg := whatsapp.BuildQuickReplyButtons(
		"📢 Layanan Pengaduan",
		text,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "0", DisplayText: "🏠 Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleMenu9_HubDev renders developer contact
func (r *BotRouter) HandleMenu9_HubDev(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	text := "⚙️ *Hubungi Pengembang Bot ISTIMEWA*\n\nJika Sahabat Data mengalami kendala teknis, bot tidak merespons, atau memiliki ide peningkatan fitur untuk BOT ISTIMEWA, silakan hubungi tim IT BPS Kabupaten Bangka Barat:\n\n📞 *Kontak Developer*:\nhttps://wa.me/6285860125305\n\n📧 *Email*:\nbps1903@bps.go.id"

	msg := whatsapp.BuildQuickReplyButtons(
		"⚙️ Kontak Pengembang",
		text,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "0", DisplayText: "🏠 Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleMenu10_FAQ renders FAQ menu (Resolves TD-08)
func (r *BotRouter) HandleMenu10_FAQ(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu10)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "1️⃣  Cara Mendapatkan Data BPS", Description: "Panduan akses data resmi"},
		{ID: "2", Title: "2️⃣  Biaya Layanan Data", Description: "Informasi tarif PST BPS"},
		{ID: "3", Title: "3️⃣  Data Tingkat Desa/Kelurahan", Description: "Ketersediaan data level mikro"},
		{ID: "4", Title: "4️⃣  Permohonan Rekomendasi Kegiatan", Description: "Rekomendasi statistik sektoral"},
		{ID: "5", Title: "5️⃣  Jadwal Rilis Data", Description: "Kalender rilis publikasi BPS"},
	}

	introText := "📒 *Pertanyaan yang Sering Diajukan (FAQs)*\nSilakan pilih topik pertanyaan yang ingin Sahabat ketahui penjelasannya:"
	body := whatsapp.BuildBodyText(introText, items, "\n*0.* 📑 Menu Awal\n\n◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇\n```Statistik Bermakna, Statistik Berdampak```")

	msg := whatsapp.BuildInteractiveList(
		"📒 Frequently Asked Questions",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Pertanyaan FAQ ❓",
		[]whatsapp.ListSection{
			{Title: "Daftar Pertanyaan", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

func (r *BotRouter) HandleFAQDetail(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage, idx int) error {
	faqs := []struct {
		Q string
		A string
	}{
		{
			Q: "Bagaimana cara memperoleh data statistik BPS Kabupaten Bangka Barat?",
			A: "Sahabat Data dapat mengunduh publikasi dan tabel data secara gratis melalui website https://bangkabaratkab.bps.go.id, melalui Bot WhatsApp ini pada Menu 1, atau datang langsung ke Pelayanan Statistik Terpadu (PST) BPS Kabupaten Bangka Barat.",
		},
		{
			Q: "Apakah layanan data di BPS Bangka Barat berbayar?",
			A: "Seluruh publikasi, tabel statistik, berita resmi statistik (BRS), dan konsultasi di PST BPS Bangka Barat adalah GRATIS (Rp 0). Hanya data mikro khusus atau peta digital beresolusi tinggi tertentu yang dikenakan tarif PNBP resmi sesuai PP No. 7 Tahun 2015.",
		},
		{
			Q: "Apakah BPS menyediakan data tingkat Desa atau Kelurahan?",
			A: "Ya! BPS menerbitkan publikasi 'Kecamatan Dalam Angka' setiap tahun yang memuat data rinci hingga tingkat desa/kelurahan untuk seluruh 6 kecamatan di Bangka Barat (Muntok, Simpang Teritip, Jebus, Kelapa, Tempilang, Parittiga).",
		},
		{
			Q: "Bagaimana cara mengajukan Rekomendasi Kegiatan Statistik (ROMANTIK)?",
			A: "Bagi instansi pemerintah/OPD yang akan menyelenggarakan survei/kompilasi data statistik sektoral, silakan mengajukan permohonan rekomendasi melalui portal https://romantik.bps.go.id atau konsultasi melalui PST BPS Bangka Barat.",
		},
		{
			Q: "Kapan jadwal rilis data terbaru BPS Bangka Barat?",
			A: "Jadwal rilis publikasi dan Berita Resmi Statistik (BRS) dirilis berkala sesuai ARC (Advance Release Calendar). BRS inflasi/ekonomi umumnya dirilis setiap awal bulan, dan publikasi 'Bangka Barat Dalam Angka' dirilis setiap akhir Februari.",
		},
	}

	var faqText string
	if idx >= 1 && idx <= len(faqs) {
		item := faqs[idx-1]
		faqText = fmt.Sprintf("*Pertanyaan %d* 🤔 :\n*%s*\n\n👩‍🏫 *Jawaban*:\n%s\n\n◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇\n```Statistik Bermakna, Statistik Berdampak```", idx, item.Q, item.A)
	} else {
		faqText = "❓ Informasi untuk pertanyaan ini sedang diperbarui."
	}

	msg := whatsapp.BuildQuickReplyButtons(
		"💡 Jawaban FAQ",
		faqText,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "↩️ Pertanyaan Lain (99)"},
			{ID: "0", DisplayText: "🏠 Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}
