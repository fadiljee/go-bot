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
	"bot-istimewa/internal/locale"
	"bot-istimewa/internal/scraper"
	"bot-istimewa/internal/whatsapp"
)

// HandleMenu6_Konsultasi renders Konsultasi Statistik menu
func (r *BotRouter) HandleMenu6_Konsultasi(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu6)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "Ajukan Jadwal Konsultasi", Description: "Reservasi jadwal konsultasi tatap muka/online"},
		{ID: "2", Title: "Tanya Para Pakar", Description: "Hubungi pakar statistik BPS langsung"},
		{ID: "3", Title: "Informasi PST & Jam Layanan", Description: "Jadwal operasional & kontak petugas PST"},
	}

	introText := "💬 *TANYA PAKAR & KONSULTASI ONLINE*\n_PST BPS Kabupaten Bangka Barat_\n\nAda yang ingin Sahabat tanyakan seputar data & statistik? Pilih jenis layanan:\n━━━━━━━━━━━━━━━━━━"
	footerText := "━━━━━━━━━━━━━━━━━━\n⌨️ Balas dengan *angka menu*\n• Ketik *0* untuk Menu Utama\n_Statistik Bermakna, Statistik Berdampak_"
	body := whatsapp.BuildBodyText(introText, items, footerText)

	msg := whatsapp.BuildInteractiveList(
		"Konsultasi Statistik",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Layanan Konsultasi",
		[]whatsapp.ListSection{
			{Title: "Layanan Konsultasi", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

func normalizePhone(raw string) string {
	raw = strings.TrimSpace(raw)
	var digits strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	d := digits.String()
	if strings.HasPrefix(d, "08") {
		return "62" + d[1:]
	}
	if strings.HasPrefix(d, "8") && len(d) >= 9 {
		return "62" + d
	}
	return d
}

// HandleMenu6_BookingPrompt prompts user to submit consultation booking details
func (r *BotRouter) HandleMenu6_BookingPrompt(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu6_Booking)

	prompt := "📝 *PENGAJUAN JADWAL KONSULTASI*\n_PST BPS Kabupaten Bangka Barat_\n\n" +
		"Silakan kirimkan data permohonan konsultasi Sahabat dengan format berikut:\n\n" +
		"━━━━━━━━━━━━━━━━━━\n" +
		"Nama: \n" +
		"Instansi / Asal: \n" +
		"No. WhatsApp / HP: \n" +
		"Topik / Kebutuhan Data: \n" +
		"Rencana Hari / Jam: \n" +
		"━━━━━━━━━━━━━━━━━━\n\n" +
		"💡 _(Atau ketik informasi kebutuhan data Sahabat secara bebas)_\n\n" +
		"• Ketik *0* untuk Batal & Kembali ke Menu Utama"

	return r.sender.SendTextMessage(ctx, p.JID, prompt)
}

func isPhoneKey(key string) bool {
	if strings.Contains(key, "wa") ||
		strings.Contains(key, "whatsapp") ||
		strings.Contains(key, "hp") ||
		strings.Contains(key, "telepon") ||
		strings.Contains(key, "telp") ||
		strings.Contains(key, "kontak") ||
		strings.Contains(key, "phone") ||
		strings.HasPrefix(key, "no.") ||
		strings.HasPrefix(key, "no ") {
		return true
	}
	return false
}

type BookingDetails struct {
	Name          string
	Instansi      string
	Phone         string
	Topic         string
	PreferredTime string
	CleanNotes    string
}

func parseConsultationBookingInput(input string) BookingDetails {
	lines := strings.Split(input, "\n")
	var b BookingDetails
	var meaningfulLines []string

	for i := 0; i < len(lines); i++ {
		rawTrimmed := strings.TrimSpace(lines[i])
		if rawTrimmed == "" {
			continue
		}

		cleanTrimmed := strings.Trim(rawTrimmed, "*_`~ \t-•")
		cleanLower := strings.ToLower(cleanTrimmed)

		// 1. Detect stop markers (contoh pengisian, petunjuk pembatalan)
		// Users frequently copy/quote the whole bot template including "*Contoh Pengisian*:"
		// We must stop parsing immediately to prevent example names/topics from leaking or overwriting user data.
		if strings.Contains(cleanLower, "contoh") ||
			strings.Contains(cleanLower, "ketik 0") ||
			strings.Contains(cleanLower, "ketik *0*") ||
			strings.Contains(cleanLower, "batal dan kembali") ||
			strings.Contains(cleanLower, "menu utama") ||
			strings.Contains(cleanLower, "secara bebas") {
			break
		}

		// 2. Skip template headers/instructions if copied by user
		if strings.Contains(cleanLower, "pengajuan jadwal konsultasi") ||
			strings.Contains(cleanLower, "silakan kirimkan data") ||
			strings.Contains(cleanLower, "permohonan konsultasi") ||
			strings.Contains(cleanLower, "membalas pesan ini") ||
			strings.Contains(cleanLower, "format berikut") ||
			strings.Contains(cleanLower, "menggunakan format") ||
			strings.Contains(cleanLower, "cukup ketik") ||
			cleanLower == "```" {
			continue
		}

		// 3. Key-Value parsing
		parts := strings.SplitN(cleanTrimmed, ":", 2)
		if len(parts) == 2 {
			key := strings.ToLower(strings.TrimSpace(parts[0]))
			key = strings.TrimLeft(key, "0123456789.-)•* \t[]_`~")
			key = strings.Trim(key, "*_`~ \t-•[]()")

			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, "*_`~ \t")

			// If val is empty on this line, check if value is provided on the next line
			if val == "" && i+1 < len(lines) {
				nextTrimmed := strings.TrimSpace(lines[i+1])
				nextClean := strings.Trim(nextTrimmed, "*_`~ \t-•")
				nextLower := strings.ToLower(nextClean)
				if nextClean != "" && !strings.Contains(nextClean, ":") &&
					!strings.Contains(nextLower, "contoh") &&
					!strings.Contains(nextLower, "ketik 0") {
					val = nextClean
					rawTrimmed = rawTrimmed + " " + nextTrimmed
					i++
				}
			}

			isHandledKey := false

			// Match Name (handles "Nama:", "Nama Pemohon:", "Nama Lengkap:", "Pemohon:", etc.)
			if strings.HasPrefix(key, "nama") || key == "pemohon" || key == "klien" {
				if b.Name == "" && val != "" {
					b.Name = val
				}
				isHandledKey = true
			} else if strings.HasPrefix(key, "instansi") || strings.HasPrefix(key, "asal") || key == "lembaga" || key == "dinas" || key == "kantor" || key == "perusahaan" || key == "kampus" || key == "universitas" || key == "sekolah" || key == "pekerjaan" || key == "profesi" {
				if b.Instansi == "" && val != "" {
					b.Instansi = val
				}
				isHandledKey = true
			} else if isPhoneKey(key) {
				if b.Phone == "" && val != "" {
					b.Phone = normalizePhone(val)
				}
				isHandledKey = true
			} else if strings.HasPrefix(key, "topik") || strings.HasPrefix(key, "kebutuhan") || strings.HasPrefix(key, "keperluan") || strings.HasPrefix(key, "perihal") || strings.HasPrefix(key, "judul") || strings.HasPrefix(key, "konsultasi") || key == "data" {
				if b.Topic == "" && val != "" {
					b.Topic = val
				}
				isHandledKey = true
			} else if strings.Contains(key, "hari") || strings.Contains(key, "jam") || strings.Contains(key, "jadwal") || strings.Contains(key, "waktu") || strings.Contains(key, "tanggal") || strings.Contains(key, "tgl") {
				if b.PreferredTime == "" && val != "" {
					// Check if next line contains time info continuation (e.g. "10.00 WIB")
					if i+1 < len(lines) {
						nextTrimmed := strings.TrimSpace(lines[i+1])
						nextClean := strings.Trim(nextTrimmed, "*_`~ \t-•")
						nextLower := strings.ToLower(nextClean)
						if nextClean != "" && !strings.Contains(nextClean, ":") &&
							!strings.Contains(nextLower, "contoh") &&
							!strings.Contains(nextLower, "ketik 0") &&
							(strings.Contains(nextLower, "wib") || strings.Contains(nextLower, "jam") || strings.Contains(nextLower, ".")) {
							val = val + " " + nextClean
							rawTrimmed = rawTrimmed + " " + nextTrimmed
							i++
						}
					}
					b.PreferredTime = val
				}
				isHandledKey = true
			}

			if isHandledKey {
				if val != "" {
					meaningfulLines = append(meaningfulLines, rawTrimmed)
				}
				continue
			}
		}

		meaningfulLines = append(meaningfulLines, rawTrimmed)
	}

	// Safety check: Never allow dummy template example data to be treated as user data
	if strings.Contains(strings.ToLower(b.Name), "ahmad fauzi") {
		b.Name = ""
	}
	if b.Phone == "6282164302643" {
		b.Phone = ""
	}

	if len(meaningfulLines) > 0 {
		b.CleanNotes = strings.Join(meaningfulLines, "\n")
	} else {
		b.CleanNotes = ""
	}

	return b
}

// HandleProcessBooking processes user booking input, stores to SQLite, notifies user, and alerts office group
func (r *BotRouter) HandleProcessBooking(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	input := strings.TrimSpace(p.Text)
	if input == "" {
		return r.sender.SendTextMessage(ctx, p.JID, "Mohon ketikkan data permohonan konsultasi Sahabat, atau ketik *0* untuk kembali.")
	}

	details := parseConsultationBookingInput(input)
	name := details.Name
	instansi := details.Instansi
	phoneInput := details.Phone
	topic := details.Topic
	preferredTime := details.PreferredTime

	// Fallback if user submitted freeform text without structured labels
	if name == "" {
		if p.Name != "" && !strings.Contains(strings.ToLower(p.Name), "ahmad fauzi") {
			name = p.Name
		} else {
			name = "Sahabat Data"
		}
	}
	if instansi == "" {
		instansi = "Masyarakat / Umum"
	}
	if topic == "" {
		if details.CleanNotes != "" {
			topic = details.CleanNotes
		} else {
			topic = "Konsultasi Data Statistik PST"
		}
	}
	if preferredTime == "" {
		preferredTime = "Hari & jam kerja PST (08.00 - 15.30 WIB)"
	}

	// Clean notes for group notification and database
	notes := details.CleanNotes
	if notes == "" {
		notes = fmt.Sprintf("Nama: %s\nInstansi: %s\nTopik: %s\nRencana Jadwal: %s", name, instansi, topic, preferredTime)
	}

	// Determine final phone number:
	// Priority 1: User typed in the booking message
	// Priority 2: Parsed Phone Number from WhatsApp (SenderAlt / Phone JID)
	// Priority 3: Fallback if JID is phone number
	phone := phoneInput
	if phone == "" && p.PhoneNumber != "" {
		phone = normalizePhone(p.PhoneNumber)
	}
	if phone == "" && p.JID.Server == "s.whatsapp.net" {
		phone = normalizePhone(p.JID.User)
	}

	var waContact string
	if phone != "" && strings.HasPrefix(phone, "62") {
		waContact = fmt.Sprintf("https://wa.me/%s", phone)
	} else if phone != "" {
		waContact = fmt.Sprintf("+%s (https://wa.me/%s)", phone, phone)
	} else {
		// In case user is on WhatsApp LID addressing without visible phone number
		waContact = fmt.Sprintf("%s (Akun WhatsApp Pengirim)", p.JID.User)
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
		Notes:         notes,
		CreatedAt:     time.Now(),
	}

	// 1. Save to SQLite DB
	if r.scraper != nil {
		if err := r.scraper.CreateConsultationBooking(ctx, booking); err != nil {
			log.Printf("[BOOKING] Gagal menyimpan ke SQLite: %v", err)
		}
	}

	// 2. Reply confirmation to user
	userReply := fmt.Sprintf("✅ *PENGAJUAN JADWAL KONSULTASI DITERIMA*\n_PST BPS Kabupaten Bangka Barat_\n\n"+
		"Terima kasih *%s*, permohonan konsultasi statistik Sahabat telah dicatat dan diteruskan ke Tim Pelayanan Statistik Terpadu (PST) BPS Bangka Barat.\n\n"+
		"━━━━━━━━━━━━━━━━━━\n"+
		"🎫 *No. Tiket*: `%s`\n"+
		"👤 *Nama*: %s\n"+
		"🏢 *Instansi*: %s\n"+
		"📋 *Topik Data*: %s\n"+
		"🕒 *Rencana Jadwal*: %s\n"+
		"━━━━━━━━━━━━━━━━━━\n\n"+
		"Petugas PST atau pakar terkait kami akan segera menghubungi Sahabat via WhatsApp untuk konfirmasi persiapan data.\n\n"+
		"• Ketik *0* untuk kembali ke Menu Utama",
		name, ticketNo, name, instansi, topic, preferredTime)

	_ = r.sender.SendTextMessage(ctx, p.JID, userReply)

	// 3. Auto-Forward Notification to Office Internal Group
	if r.cfg.InternalPSTGroupJID != "" {
		groupJID, err := types.ParseJID(r.cfg.InternalPSTGroupJID)
		if err == nil {
			groupMsg := fmt.Sprintf("*PEMBERITAHUAN JADWAL KONSULTASI BARU (PST BPS)*\n\n"+
				"Halo Tim PST & Para Pakar BPS Bangka Barat, ada permohonan konsultasi baru masuk melalui WhatsApp Bot:\n\n"+
				"*No. Tiket*: `%s`\n"+
				"*Nama Pemohon*: %s\n"+
				"*Instansi / Asal*: %s\n"+
				"*Kontak WhatsApp*: %s\n"+
				"*Bidang / Topik*: %s\n"+
				" *Rencana Jadwal*: %s\n"+
				"*Waktu Masuk*: %s WIB\n\n"+
				"*Pesan Lengkap*:\n_%s_\n\n"+
				"*Dimohon petugas piket PST dan pakar terkait untuk menindaklanjuti dan mempersiapkan bahan data. Terima kasih!* ",
				ticketNo, name, instansi, waContact, topic, preferredTime, time.Now().Format("02 Jan 2006 15:04"), notes)

			if sendErr := r.sender.SendTextMessage(ctx, groupJID, groupMsg); sendErr != nil {
				log.Printf("[PST GROUP] Gagal mengirim notifikasi ke grup %s: %v", r.cfg.InternalPSTGroupJID, sendErr)
			} else {
				log.Printf("[PST GROUP] Sukses meneruskan jadwal %s ke grup internal kantor %s", ticketNo, r.cfg.InternalPSTGroupJID)
			}
		} else {
			log.Printf("[PST GROUP] Format INTERNAL_PST_GROUP_JID (%s) tidak valid: %v", r.cfg.InternalPSTGroupJID, err)
		}
	} else {
		log.Printf("ℹ[PST GROUP] INTERNAL_PST_GROUP_JID belum diisi di .env; notifikasi grup dilewati.")
	}

	// Reset state back to root after booking
	s.ResetToRoot()
	return nil
}

func (r *BotRouter) HandleKonsultasiUmum(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	piketText := "🕒 *JAM LAYANAN PST BPS BANGKA BARAT*\n_Pelayanan Statistik Terpadu_\n\n" +
		"Petugas PST kami siap melayani pertanyaan Sahabat seputar data, publikasi, dan rujukan statistik:\n\n" +
		"━━━━━━━━━━━━━━━━━━\n" +
		"🕒 *Jadwal Layanan*:\n" +
		"• Senin – Kamis: 08.00 – 15.30 WIB\n" +
		"• Jumat: 08.00 – 15.00 WIB\n\n" +
		"💬 *WhatsApp Petugas PST*:\n" +
		"https://wa.me/6282164302643?text=Halo%20Petugas%20PST%20BPS%20Bangka%20Barat\n" +
		"━━━━━━━━━━━━━━━━━━\n\n" +
		"_Silakan hubungi kami pada jam operasional kerja kedinasan._"

	msg := whatsapp.BuildQuickReplyButtons(
		"Petugas Konsultasi Umum",
		piketText,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "Menu Sebelumnya (99)"},
			{ID: "0", DisplayText: "Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

func (r *BotRouter) HandleKonsultasiPakarList(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu6_Pakar)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "Statistik Sosial Kependudukan", Description: "Kependudukan, Kemiskinan, Pendidikan, Kesehatan"},
		{ID: "2", Title: "Statistik Pertanian", Description: "Perkebunan Sawit, Lada, Karet, Perikanan"},
		{ID: "3", Title: "Statistik Ekonomi & PDRB", Description: "Neraca Ekonomi, PDRB, Industri Manufaktur"},
		{ID: "4", Title: "Distribusi, Harga & KTIP", Description: "Inflasi, IHK, Pariwisata, Perdagangan"},
		{ID: "5", Title: "Statistik Sektoral", Description: "Data OPD, Satu Data Indonesia, ROMANTIK"},
		{ID: "6", Title: "Statistik Khusus", Description: "Analisis Tematik Khusus & EPSS"},
		{ID: "7", Title: "Metodologi Statistik", Description: "Desain Sampling, Survei & Kualitas Data"},
		{ID: "8", Title: "Diseminasi & Layanan", Description: "Publikasi, Perpustakaan & Portal PST Online"},
	}

	introText := "📍 *Menu Utama* ❯ *[6] Konsultasi* ❯ *Tanya Pakar*\n\nSilakan pilih bidang kepakaran yang ingin Sahabat konsultasikan:\n━━━━━━━━━━━━━━━━━━"
	footerText := "━━━━━━━━━━━━━━━━━━\n⌨️ Balas dengan *angka menu*\n• Ketik *99* untuk Kembali\n• Ketik *0* untuk Menu Utama"
	body := whatsapp.BuildBodyText(introText, items, footerText)

	msg := whatsapp.BuildInteractiveList(
		"Tanya Para Pakar",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Bidang Pakar",
		[]whatsapp.ListSection{
			{Title: "Bidang Kepakaran", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

func (r *BotRouter) HandlePakarDetail(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage, idx int) error {
	var pakarName, bidangTitle, desc, phone string

	// 1. Ambil kontak pakar langsung dari database SQLite scraper
	if r.scraper != nil {
		if c, err := r.scraper.GetExpertContact(ctx, idx); err == nil && c != nil {
			pakarName = c.PakarName
			bidangTitle = c.BidangTitle
			desc = c.Description
			phone = c.Phone
		}
	}

	// 2. Fallback jika database belum siap
	if pakarName == "" {
		pakars := gsheet.DefaultPakar()
		noStr := strconv.Itoa(idx)
		for i := range pakars {
			if pakars[i].No.String() == noStr {
				pakarName = pakars[i].Nama
				bidangTitle = pakars[i].Nama
				desc = pakars[i].Penjelasan
				phone = "6282164302643"
				break
			}
		}
	}

	if phone == "" {
		phone = "6282164302643"
	}
	cleanPhone := normalizePhone(phone)

	waLink := fmt.Sprintf("https://wa.me/%s?text=Halo%%20%s,%%20saya%%20ingin%%20konsultasi%%20seputar%%20data%%20statistik%%20Bangka%%20Barat.", cleanPhone, strings.ReplaceAll(pakarName, " ", "%20"))

	pakarText := fmt.Sprintf("👤 *PROFIL PAKAR STATISTIK*\n_BPS Kabupaten Bangka Barat_\n\n"+
		"━━━━━━━━━━━━━━━━━━\n"+
		"📌 *Bidang*: %s\n"+
		"👤 *Pakar*: %s\n"+
		"📞 *No. WhatsApp*: +%s\n"+
		"📝 *Cakupan*: %s\n"+
		"━━━━━━━━━━━━━━━━━━\n\n"+
		"💬 *Hubungi via WhatsApp*:\n%s\n\n"+
		"ℹ _Mohon menghubungi beliau pada hari dan jam kerja dinas (Senin - Jumat, 08.00 - 15.30 WIB). Terima kasih!_",
		bidangTitle, pakarName, cleanPhone, desc, waLink)

	msg := whatsapp.BuildQuickReplyButtons(
		"Profil Pakar Statistik",
		pakarText,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "Pilih Pakar Lain (99)"},
			{ID: "0", DisplayText: "Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleMenu7_Berlangganan renders subscription menu
func (r *BotRouter) HandleMenu7_Berlangganan(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	text := "📬 *LANGGANAN PUBLIKASI GRATIS*\n_BPS Kabupaten Bangka Barat_\n\n" +
		"Dapatkan pemberitahuan rilis publikasi terbaru, Berita Resmi Statistik (BRS), dan ringkasan data strategis Kabupaten Bangka Barat langsung ke nomor WhatsApp Sahabat!\n\n" +
		"━━━━━━━━━━━━━━━━━━\n" +
		"🔗 *Formulir Berlangganan*:\n" +
		"https://s.bps.go.id/LanggananDataBabar\n" +
		"━━━━━━━━━━━━━━━━━━\n\n" +
		"🔒 _Kerahasiaan nomor Sahabat dijamin oleh Undang-Undang No. 16 Tahun 1997 tentang Statistik._"

	msg := whatsapp.BuildQuickReplyButtons(
		"Layanan Berlangganan",
		text,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "0", DisplayText: "Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleMenu8_Pengaduan renders complaint & aspiration channels
func (r *BotRouter) HandleMenu8_Pengaduan(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	text := "📢 *PENGADUAN & KOTAK SARAN*\n_BPS Kabupaten Bangka Barat_\n\n" +
		"BPS Kabupaten Bangka Barat berkomitmen memberikan pelayanan prima yang bersih, transparan, dan bebas dari pungutan liar (Zona Integritas WBK/WBBM).\n\n" +
		"━━━━━━━━━━━━━━━━━━\n" +
		"1️⃣ *Portal SP4N-LAPOR!*:\n" +
		"https://www.lapor.go.id/\n\n" +
		"2️⃣ *Email Pengaduan*:\n" +
		"bps1903@bps.go.id\n\n" +
		"3️⃣ *Kotak Pengaduan Fisik*:\n" +
		"Ruang PST BPS Kabupaten Bangka Barat\n" +
		"━━━━━━━━━━━━━━━━━━\n\n" +
		"🔒 _Setiap aspirasi & laporan yang masuk akan ditindaklanjuti secara profesional dan rahasia._"

	msg := whatsapp.BuildQuickReplyButtons(
		"Layanan Pengaduan",
		text,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "0", DisplayText: "Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleMenu9_HubDev renders developer contact
func (r *BotRouter) HandleMenu9_HubDev(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	devPhone := strings.TrimSpace(r.cfg.DeveloperPhone)
	if devPhone != "" {
		devPhone = normalizePhone(devPhone)
	}

	var text string
	if devPhone != "" {
		text = fmt.Sprintf("👨‍💻 *KONTAK PENGEMBANG BOT*\n_Tim TI BPS Kabupaten Bangka Barat_\n\n"+
			"Jika Sahabat Data mengalami kendala teknis, bot tidak merespons, atau memiliki ide peningkatan fitur untuk BOT ISTIMEWA, silakan hubungi tim IT BPS Kabupaten Bangka Barat:\n\n"+
			"━━━━━━━━━━━━━━━━━━\n"+
			"💬 *WhatsApp Tim IT*:\n"+
			"https://wa.me/%s\n\n"+
			"✉️ *Email Resmi*:\n"+
			"bps1903@bps.go.id\n"+
			"━━━━━━━━━━━━━━━━━━\n\n"+
			"_Kami senang menerima masukan Sahabat demi kenyamanan layanan bot ini._", devPhone)
	} else if intro := locale.T("introMenu9"); intro != "" {
		text = intro
	} else {
		text = "👨‍💻 *KONTAK PENGEMBANG BOT*\n_Tim TI BPS Kabupaten Bangka Barat_\n\n" +
			"Jika Sahabat Data mengalami kendala teknis, bot tidak merespons, atau memiliki ide peningkatan fitur untuk BOT ISTIMEWA, silakan hubungi tim IT BPS Kabupaten Bangka Barat:\n\n" +
			"━━━━━━━━━━━━━━━━━━\n" +
			"💬 *WhatsApp Tim IT*:\n" +
			"https://wa.me/6283843607098\n\n" +
			"✉️ *Email Resmi*:\n" +
			"bps1903@bps.go.id\n" +
			"━━━━━━━━━━━━━━━━━━\n\n" +
			"_Kami senang menerima masukan Sahabat demi kenyamanan layanan bot ini._"
	}

	msg := whatsapp.BuildQuickReplyButtons(
		"Kontak Pengembang",
		text,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "0", DisplayText: "Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleMenu10_FAQ renders FAQ menu (Resolves TD-08)
func (r *BotRouter) HandleMenu10_FAQ(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu10)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "Cara Mendapatkan Data BPS", Description: "Panduan akses data resmi"},
		{ID: "2", Title: "Biaya Layanan Data", Description: "Informasi tarif PST BPS"},
		{ID: "3", Title: "Data Tingkat Desa/Kelurahan", Description: "Ketersediaan data level mikro"},
		{ID: "4", Title: "Permohonan Rekomendasi Kegiatan", Description: "Rekomendasi statistik sektoral"},
		{ID: "5", Title: "Jadwal Rilis Data", Description: "Kalender rilis publikasi BPS"},
	}

	introText := "💡 *BANTUAN & FAQ*\n_BPS Kabupaten Bangka Barat_\n\nBerikut beberapa pertanyaan yang sering diajukan seputar layanan BPS Bangka Barat:\n━━━━━━━━━━━━━━━━━━"
	footerText := "━━━━━━━━━━━━━━━━━━\n⌨️ Balas dengan *angka menu*\n• Ketik *0* untuk Menu Utama\n_Statistik Bermakna, Statistik Berdampak_"
	body := whatsapp.BuildBodyText(introText, items, footerText)

	msg := whatsapp.BuildInteractiveList(
		"Frequently Asked Questions",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Pertanyaan FAQ",
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
		faqText = fmt.Sprintf("💡 *BANTUAN & FAQ #%d*\n_BPS Kabupaten Bangka Barat_\n\n"+
			"❓ *%s*\n\n"+
			"━━━━━━━━━━━━━━━━━━\n"+
			"💡 *Penjelasan*:\n"+
			"%s\n"+
			"━━━━━━━━━━━━━━━━━━\n\n"+
			"```Statistik Bermakna, Statistik Berdampak```", idx, item.Q, item.A)
	} else {
		faqText = "Informasi untuk pertanyaan ini sedang diperbarui."
	}

	msg := whatsapp.BuildQuickReplyButtons(
		"Jawaban FAQ",
		faqText,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "Pertanyaan Lain (99)"},
			{ID: "0", DisplayText: "Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}
