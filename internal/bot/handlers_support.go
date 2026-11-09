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

// HandleMenu6Konsultasi renders Konsultasi Statistik menu
func (r *BotRouter) HandleMenu6Konsultasi(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu6)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "Ajukan Jadwal Konsultasi", Description: "Reservasi jadwal konsultasi tatap muka/online"},
		{ID: "2", Title: "Tanya Para Pakar", Description: "Hubungi pakar statistik BPS langsung"},
		{ID: "3", Title: "Informasi PST & Jam Layanan", Description: "Jadwal operasional & kontak petugas PST"},
	}

	introText := "*TANYA PAKAR & KONSULTASI ONLINE*\n_PST BPS Kabupaten Bangka Barat_\n────────────\n\nAda yang ingin Sahabat tanyakan seputar data & statistik? Pilih jenis layanan:"
	footerText := "────────────\nBalas dengan *angka menu (1 - 3)*\n• Ketik *0* untuk Menu Utama\n_Statistik Bermakna, Statistik Berdampak_"
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

func formatPhoneDisplay(raw string) string {
	norm := normalizePhone(raw)
	local := norm
	if strings.HasPrefix(norm, "62") {
		local = "0" + norm[2:]
	}
	if len(local) >= 10 {
		p1 := local[:4]
		p2 := local[4:8]
		p3 := local[8:]
		return fmt.Sprintf("%s-%s-%s (+62 %s-%s-%s)", p1, p2, p3, local[1:4], p2, p3)
	}
	return local
}

// HandleMenu6BookingPrompt prompts user to submit consultation booking details
func (r *BotRouter) HandleMenu6BookingPrompt(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu6Booking)

	prompt := "*PENGAJUAN JADWAL KONSULTASI*\n" +
		"_PST BPS Kabupaten Bangka Barat_\n" +
		"────────────\n\n" +
		"Silakan salin dan lengkapi formulir permohonan konsultasi berikut:\n\n" +
		"Nama: \n" +
		"Instansi / Asal: \n" +
		"No. WhatsApp / HP: \n" +
		"Topik / Kebutuhan Data: \n" +
		"Rencana Jadwal (Hari, Tanggal, Jam): \n\n" +
		"*Contoh Pengisian:*\n" +
		"Nama: Ahmad Fauzi\n" +
		"Instansi / Asal: Dinas Pertanian Bangka Barat\n" +
		"No. WhatsApp / HP: 082164302643\n" +
		"Topik / Kebutuhan Data: Data Produksi Padi & Sawit 2025\n" +
		"Rencana Jadwal (Hari, Tanggal, Jam): Kamis, 12 Nov 2026, Pukul 10.00 WIB\n\n" +
		"_Catatan: Jam layanan PST Senin–Kamis (08.00–15.30 WIB) & Jumat (08.00–15.00 WIB)._\n\n" +
		"────────────\n" +
		"• Ketik *99* untuk Kembali ke Menu Konsultasi\n" +
		"• Ketik *0* untuk Batal & Menu Utama"

	return r.sender.SendTextMessage(ctx, p.JID, prompt)
}

func isPhoneKey(key string) bool {
	if isScheduleKey(key) {
		return false
	}
	if key == "wa" ||
		strings.Contains(key, "whatsapp") ||
		strings.Contains(key, "no. wa") ||
		strings.Contains(key, "no wa") ||
		strings.Contains(key, "nomor wa") ||
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

func isScheduleKey(key string) bool {
	return strings.Contains(key, "rencana") ||
		strings.Contains(key, "hari") ||
		strings.Contains(key, "jam") ||
		strings.Contains(key, "jadwal") ||
		strings.Contains(key, "waktu") ||
		strings.Contains(key, "tanggal") ||
		strings.Contains(key, "tgl") ||
		strings.Contains(key, "pukul")
}

func shouldStopBookingParse(cleanLower string) bool {
	return strings.Contains(cleanLower, "contoh") ||
		strings.Contains(cleanLower, "catatan:") ||
		strings.Contains(cleanLower, "jam layanan pst") ||
		strings.Contains(cleanLower, "ketik 0") ||
		strings.Contains(cleanLower, "ketik *0*") ||
		strings.Contains(cleanLower, "ketik 99") ||
		strings.Contains(cleanLower, "ketik *99*") ||
		strings.Contains(cleanLower, "batal dan kembali") ||
		strings.Contains(cleanLower, "batal & menu utama") ||
		strings.Contains(cleanLower, "kembali ke menu konsultasi") ||
		strings.Contains(cleanLower, "menu utama") ||
		strings.Contains(cleanLower, "secara bebas")
}

func shouldSkipBookingLine(cleanTrimmed, cleanLower string) bool {
	return strings.Contains(cleanTrimmed, "───") ||
		strings.Contains(cleanTrimmed, "━━━") ||
		strings.Contains(cleanLower, "pst bps kabupaten bangka barat") ||
		strings.Contains(cleanLower, "pengajuan jadwal konsultasi") ||
		strings.Contains(cleanLower, "silakan kirimkan data") ||
		strings.Contains(cleanLower, "silakan salin dan lengkapi") ||
		strings.Contains(cleanLower, "formulir permohonan") ||
		strings.Contains(cleanLower, "permohonan konsultasi") ||
		strings.Contains(cleanLower, "membalas pesan ini") ||
		strings.Contains(cleanLower, "format berikut") ||
		strings.Contains(cleanLower, "menggunakan format") ||
		strings.Contains(cleanLower, "cukup ketik") ||
		cleanLower == "```"
}

func normalizeBookingKey(rawKey string) string {
	key := strings.ToLower(strings.TrimSpace(rawKey))
	key = strings.TrimLeft(key, "0123456789.-)•* \t[]_`~")
	key = strings.Trim(key, "*_`~ \t-•[]()")
	return key
}

func isKnownBookingKey(key string) bool {
	if key == "" {
		return false
	}
	if strings.HasPrefix(key, "nama") || key == "pemohon" || key == "klien" {
		return true
	}
	if strings.HasPrefix(key, "instansi") || strings.HasPrefix(key, "asal") || key == "lembaga" || key == "dinas" || key == "kantor" || key == "perusahaan" || key == "kampus" || key == "universitas" || key == "sekolah" || key == "pekerjaan" || key == "profesi" {
		return true
	}
	if isPhoneKey(key) {
		return true
	}
	if strings.HasPrefix(key, "topik") || strings.HasPrefix(key, "kebutuhan") || strings.HasPrefix(key, "keperluan") || strings.HasPrefix(key, "perihal") || strings.HasPrefix(key, "judul") || strings.HasPrefix(key, "konsultasi") || key == "data" {
		return true
	}
	return isScheduleKey(key)
}

func isBookingFormKeyLine(line string) bool {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return false
	}
	return isKnownBookingKey(normalizeBookingKey(parts[0]))
}

func formatPreferredSchedule(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.Trim(s, "*_`~ \t-•,;|")
	fields := strings.Fields(s)
	for i, f := range fields {
		// Normalize HH:MM time notation (e.g. "10:00") to Indonesian standard "10.00"
		if len(f) == 5 && f[2] == ':' &&
			f[0] >= '0' && f[0] <= '2' &&
			f[1] >= '0' && f[1] <= '9' &&
			f[3] >= '0' && f[3] <= '5' &&
			f[4] >= '0' && f[4] <= '9' {
			fields[i] = f[:2] + "." + f[3:]
		}
	}
	return strings.Join(fields, " ")
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

		// 1. Detect stop markers (contoh pengisian, catatan, petunjuk navigasi)
		if shouldStopBookingParse(cleanLower) {
			break
		}

		// 2. Skip template headers/instructions/dividers if copied by user
		if shouldSkipBookingLine(cleanTrimmed, cleanLower) {
			continue
		}

		// 3. Key-Value parsing
		parts := strings.SplitN(cleanTrimmed, ":", 2)
		if len(parts) == 2 {
			key := normalizeBookingKey(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, "*_`~ \t")

			if isKnownBookingKey(key) {
				// If val is empty on this line, check if value is provided on the next line
				if val == "" && i+1 < len(lines) {
					nextTrimmed := strings.TrimSpace(lines[i+1])
					nextClean := strings.Trim(nextTrimmed, "*_`~ \t-•")
					nextLower := strings.ToLower(nextClean)
					if nextClean != "" &&
						!shouldStopBookingParse(nextLower) &&
						!shouldSkipBookingLine(nextClean, nextLower) &&
						!isBookingFormKeyLine(nextClean) {
						val = nextClean
						rawTrimmed = rawTrimmed + " " + nextTrimmed
						i++
					}
				}

				switch {
				case strings.HasPrefix(key, "nama") || key == "pemohon" || key == "klien":
					if b.Name == "" && val != "" {
						b.Name = val
					}
				case strings.HasPrefix(key, "instansi") || strings.HasPrefix(key, "asal") || key == "lembaga" || key == "dinas" || key == "kantor" || key == "perusahaan" || key == "kampus" || key == "universitas" || key == "sekolah" || key == "pekerjaan" || key == "profesi":
					if b.Instansi == "" && val != "" {
						b.Instansi = val
					}
				case isPhoneKey(key):
					if b.Phone == "" && val != "" {
						b.Phone = normalizePhone(val)
					}
				case strings.HasPrefix(key, "topik") || strings.HasPrefix(key, "kebutuhan") || strings.HasPrefix(key, "keperluan") || strings.HasPrefix(key, "perihal") || strings.HasPrefix(key, "judul") || strings.HasPrefix(key, "konsultasi") || key == "data":
					if b.Topic == "" && val != "" {
						b.Topic = val
					}
				case isScheduleKey(key):
					if val != "" {
						// Check if next line contains time info continuation (e.g. "10.00 WIB" or "10:00 WIB")
						if i+1 < len(lines) {
							nextTrimmed := strings.TrimSpace(lines[i+1])
							nextClean := strings.Trim(nextTrimmed, "*_`~ \t-•")
							nextLower := strings.ToLower(nextClean)
							if nextClean != "" &&
								!shouldStopBookingParse(nextLower) &&
								!shouldSkipBookingLine(nextClean, nextLower) &&
								!isBookingFormKeyLine(nextClean) &&
								(strings.Contains(nextLower, "wib") || strings.Contains(nextLower, "jam") || strings.Contains(nextLower, "pukul") || strings.Contains(nextLower, ".") || strings.Contains(nextLower, ":")) {
								val = val + " " + nextClean
								rawTrimmed = rawTrimmed + " " + nextTrimmed
								i++
							}
						}
						formattedSched := formatPreferredSchedule(val)
						if b.PreferredTime == "" {
							b.PreferredTime = formattedSched
						} else if !strings.Contains(strings.ToLower(b.PreferredTime), strings.ToLower(formattedSched)) {
							b.PreferredTime = b.PreferredTime + ", " + formattedSched
						}
					}
				}

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

func isMeaningfulBookingValue(val string, minRunes int) bool {
	clean := strings.Trim(val, "*_`~ \t-•,.:;/|\\()[]{}")
	if len([]rune(clean)) < minRunes {
		return false
	}
	hasLetter := false
	for _, r := range clean {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			hasLetter = true
			break
		}
	}
	return hasLetter
}

func validateBookingDetails(b BookingDetails) []string {
	var missing []string
	if !isMeaningfulBookingValue(b.Name, 2) {
		missing = append(missing, "Nama Lengkap")
	}
	if !isMeaningfulBookingValue(b.Instansi, 2) {
		missing = append(missing, "Instansi / Asal")
	}
	if len(b.Phone) < 9 {
		missing = append(missing, "No. WhatsApp / HP")
	}
	if !isMeaningfulBookingValue(b.Topic, 3) {
		missing = append(missing, "Topik / Kebutuhan Data")
	}
	if !isMeaningfulBookingValue(b.PreferredTime, 4) {
		missing = append(missing, "Rencana Jadwal (Hari, Tanggal, Jam)")
	}
	return missing
}

// HandleProcessBooking processes user booking input, validates required form fields, stores to SQLite, notifies user, and alerts office group
func (r *BotRouter) HandleProcessBooking(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	input := strings.TrimSpace(p.Text)
	if input == "" {
		return r.sender.SendTextMessage(ctx, p.JID, "Mohon lengkapi formulir permohonan konsultasi Sahabat, atau ketik *0* untuk kembali ke Menu Utama.")
	}

	details := parseConsultationBookingInput(input)
	missingFields := validateBookingDetails(details)
	if len(missingFields) > 0 {
		var sb strings.Builder
		sb.WriteString("*FORMAT PENGAJUAN BELUM SESUAI*\n")
		sb.WriteString("_PST BPS Kabupaten Bangka Barat_\n")
		sb.WriteString("────────────\n\n")
		sb.WriteString("Mohon maaf Sahabat Data, format pengajuan jadwal konsultasi yang dikirimkan belum lengkap atau belum sesuai format formulir.\n\n")
		sb.WriteString("*Data yang belum lengkap / sesuai:*\n")
		for _, field := range missingFields {
			fmt.Fprintf(&sb, "• %s\n", field)
		}
		sb.WriteString("\nSilakan salin dan lengkapi seluruh isian formulir berikut:\n\n")
		sb.WriteString("Nama: \n")
		sb.WriteString("Instansi / Asal: \n")
		sb.WriteString("No. WhatsApp / HP: \n")
		sb.WriteString("Topik / Kebutuhan Data: \n")
		sb.WriteString("Rencana Jadwal (Hari, Tanggal, Jam): \n\n")
		sb.WriteString("*Contoh Format Rencana Jadwal:*\n")
		sb.WriteString("Kamis, 12 Nov 2026, Pukul 10.00 WIB\n\n")
		sb.WriteString("────────────\n")
		sb.WriteString("• Ketik *99* untuk Kembali ke Menu Konsultasi\n")
		sb.WriteString("• Ketik *0* untuk Batal & Menu Utama")
		return r.sender.SendTextMessage(ctx, p.JID, sb.String())
	}

	name := details.Name
	instansi := details.Instansi
	phone := details.Phone
	topic := details.Topic
	preferredTime := formatPreferredSchedule(details.PreferredTime)

	// Clean structured notes for group notification and database
	notes := fmt.Sprintf("Nama: %s\nInstansi / Asal: %s\nNo. WhatsApp / HP: %s\nTopik / Kebutuhan Data: %s\nRencana Jadwal: %s",
		name, instansi, formatPhoneDisplay(phone), topic, preferredTime)

	var waContact string
	switch {
	case phone != "" && strings.HasPrefix(phone, "62"):
		waContact = fmt.Sprintf("%s (https://wa.me/%s)", formatPhoneDisplay(phone), phone)
	case phone != "":
		waContact = fmt.Sprintf("+%s (https://wa.me/%s)", phone, phone)
	default:
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
	userReply := fmt.Sprintf("*PENGAJUAN KONSULTASI DITERIMA*\n"+
		"_PST BPS Kabupaten Bangka Barat_\n"+
		"────────────\n\n"+
		"Terima kasih *%s*, permohonan konsultasi statistik Sahabat telah dicatat dan diteruskan ke Tim PST BPS Bangka Barat.\n\n"+
		"*1. No. Tiket*: `%s`\n"+
		"*2. Nama*: %s\n"+
		"*3. Instansi*: %s\n"+
		"*4. No. WhatsApp*: %s\n"+
		"*5. Topik Data*: %s\n"+
		"*6. Rencana Jadwal*: %s\n\n"+
		"Petugas PST atau pakar terkait akan segera menghubungi Sahabat via WhatsApp.\n\n"+
		"────────────\n"+
		"• Ketik *0* untuk kembali ke Menu Utama",
		name, ticketNo, name, instansi, formatPhoneDisplay(phone), topic, preferredTime)

	_ = r.sender.SendTextMessage(ctx, p.JID, userReply)

	// 3. Auto-Forward Notification to Office Internal Group
	if r.cfg.InternalPSTGroupJID != "" {
		groupJID, err := types.ParseJID(r.cfg.InternalPSTGroupJID)
		if err == nil {
			groupMsg := fmt.Sprintf("*PEMBERITAHUAN KONSULTASI BARU (PST BPS)*\n"+
				"────────────\n\n"+
				"Halo Tim PST & Para Pakar BPS Bangka Barat, ada permohonan konsultasi baru melalui WhatsApp Bot:\n\n"+
				"*1. No. Tiket*: `%s`\n"+
				"*2. Nama Pemohon*: %s\n"+
				"*3. Instansi / Asal*: %s\n"+
				"*4. Kontak WhatsApp*: %s\n"+
				"*5. Bidang / Topik*: %s\n"+
				"*6. Rencana Jadwal*: %s\n"+
				"*7. Waktu Masuk*: %s WIB\n\n"+
				"_Dimohon petugas piket PST dan pakar terkait untuk menindaklanjuti. Terima kasih!_",
				ticketNo, name, instansi, waContact, topic, preferredTime, time.Now().Format("02 Jan 2006 15:04"))

			if sendErr := r.sender.SendTextMessage(ctx, groupJID, groupMsg); sendErr != nil {
				log.Printf("[PST GROUP] Gagal mengirim notifikasi ke grup %s: %v", r.cfg.InternalPSTGroupJID, sendErr)
			} else {
				log.Printf("[PST GROUP] Sukses meneruskan jadwal %s ke grup internal kantor %s", ticketNo, r.cfg.InternalPSTGroupJID)
			}
		} else {
			log.Printf("[PST GROUP] Format INTERNAL_PST_GROUP_JID (%s) tidak valid: %v", r.cfg.InternalPSTGroupJID, err)
		}
	} else {
		log.Printf("[PST GROUP] INTERNAL_PST_GROUP_JID belum diisi di .env; notifikasi grup dilewati.")
	}

	// Reset state back to root after booking
	s.ResetToRoot()
	return nil
}

func (r *BotRouter) HandleKonsultasiUmum(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	pstPhone := "6282164302643"
	formattedPhone := formatPhoneDisplay(pstPhone)
	piketText := fmt.Sprintf("*INFORMASI PST & JAM LAYANAN*\n"+
		"_Pelayanan Statistik Terpadu BPS Bangka Barat_\n"+
		"────────────\n\n"+
		"Petugas PST kami siap melayani pertanyaan Sahabat seputar data, publikasi, dan rujukan statistik:\n\n"+
		"*1. Jadwal Layanan*\n"+
		"• Senin – Kamis: 08.00 – 15.30 WIB\n"+
		"• Jumat: 08.00 – 15.00 WIB\n\n"+
		"*2. WhatsApp Petugas PST*\n"+
		"• Nomor: %s\n"+
		"• Chat Langsung: https://wa.me/%s\n\n"+
		"Silakan hubungi kami pada jam operasional kerja kedinasan.",
		formattedPhone, pstPhone)

	msg := whatsapp.BuildQuickReplyButtons(
		"Informasi PST & Jam Layanan",
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
	s.PushState(StateMenu6Pakar)

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

	introText := "*TANYA PARA PAKAR STATISTIK*\n_BPS Kabupaten Bangka Barat_\n────────────\n\nSilakan pilih bidang kepakaran yang ingin Sahabat konsultasikan:"
	footerText := "────────────\nBalas dengan *angka menu (1 - 8)*\n• Ketik *99* untuk Kembali\n• Ketik *0* untuk Menu Utama"
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
	displayPhone := formatPhoneDisplay(cleanPhone)

	waLink := fmt.Sprintf("https://wa.me/%s", cleanPhone)

	pakarText := fmt.Sprintf("*PROFIL PAKAR STATISTIK*\n"+
		"_BPS Kabupaten Bangka Barat_\n"+
		"────────────\n\n"+
		"*1. Bidang*\n%s\n\n"+
		"*2. Nama Pakar*\n%s\n\n"+
		"*3. Nomor WhatsApp*\n%s\n\n"+
		"*4. Cakupan Konsultasi*\n%s\n\n"+
		"*Hubungi via WhatsApp*:\n%s\n\n"+
		"Mohon menghubungi beliau pada jam kerja dinas (Senin – Jumat, 08.00 – 15.30 WIB).",
		bidangTitle, pakarName, displayPhone, desc, waLink)

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

// HandleMenu7Berlangganan renders Channel WhatsApp BPS Bangka Barat menu (currently unavailable)
func (r *BotRouter) HandleMenu7Berlangganan(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	text := "*CHANNEL WHATSAPP BPS BANGKA BARAT*\n" +
		"_BPS Kabupaten Bangka Barat_\n" +
		"────────────\n\n" +
		"Mohon maaf Sahabat Data, fitur *Channel WhatsApp BPS Kabupaten Bangka Barat* saat ini *belum tersedia* dan sedang dalam tahap pengembangan.\n\n" +
		"Untuk sementara, Sahabat Data dapat memantau rilis publikasi, Berita Resmi Statistik (BRS), dan infografis terbaru melalui menu utama Bot ISTIMEWA atau portal resmi:\n" +
		"https://bangkabaratkab.bps.go.id"

	msg := whatsapp.BuildQuickReplyButtons(
		"Channel WhatsApp BPS Bangka Barat",
		text,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "0", DisplayText: "Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleMenu8Pengaduan renders complaint & aspiration channels
func (r *BotRouter) HandleMenu8Pengaduan(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	text := "*PENGADUAN & SARAN*\n" +
		"_BPS Kabupaten Bangka Barat_\n" +
		"────────────\n\n" +
		"BPS Kabupaten Bangka Barat berkomitmen memberikan pelayanan prima yang bersih, transparan, dan bebas dari pungutan liar (Zona Integritas WBK/WBBM).\n\n" +
		"Saluran resmi pengaduan dan saran:\n\n" +
		"*1. Tatap Muka*\n" +
		"Ruang PST BPS Kabupaten Bangka Barat, Komplek Perkantoran Pemda Terpadu Bangka Barat\n\n" +
		"*2. Kotak Pengaduan*\n" +
		"Kotak Pengaduan & Saran di Ruang Layanan PST BPS Kabupaten Bangka Barat\n\n" +
		"*3. SP4N-LAPOR*\n" +
		"https://www.lapor.go.id\n\n" +
		"*4. Email*\n" +
		"bps1903@bps.go.id\n\n" +
		"*5. Link Google Form*\n" +
		"https://s.bps.go.id/1903_FormPengaduan\n\n" +
		"Setiap aspirasi & laporan yang masuk akan ditindaklanjuti secara profesional dan rahasia."

	msg := whatsapp.BuildQuickReplyButtons(
		"Pengaduan & Saran",
		text,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "0", DisplayText: "Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleMenu10FAQ renders FAQ menu (Resolves TD-08)
func (r *BotRouter) HandleMenu10FAQ(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu10)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "Cara Mendapatkan Data BPS", Description: "Panduan akses data resmi"},
		{ID: "2", Title: "Biaya Layanan Data", Description: "Informasi tarif PST BPS"},
		{ID: "3", Title: "Data Tingkat Desa/Kelurahan", Description: "Ketersediaan data level mikro"},
		{ID: "4", Title: "Permohonan Rekomendasi Kegiatan", Description: "Rekomendasi statistik sektoral"},
		{ID: "5", Title: "Jadwal Rilis Data", Description: "Kalender rilis ARC BPS Bangka Barat"},
		{ID: "6", Title: "Bagaimana Data BPS Dihasilkan", Description: "Proses bisnis statistik (GSBPM) BPS"},
	}

	introText := "*BANTUAN & FAQ*\n_BPS Kabupaten Bangka Barat_\n────────────\n\nBerikut beberapa pertanyaan yang sering diajukan seputar layanan BPS Bangka Barat:"
	footerText := "────────────\nBalas dengan *angka menu (1 - 6)*\n• Ketik *0* untuk Menu Utama\n_Statistik Bermakna, Statistik Berdampak_"
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

func (r *BotRouter) buildARCScheduleFAQ(ctx context.Context) string {
	var sb strings.Builder
	sb.WriteString("*BANTUAN & FAQ #5*\n")
	sb.WriteString("_BPS Kabupaten Bangka Barat_\n")
	sb.WriteString("────────────\n\n")
	sb.WriteString("*Jadwal Rilis Data (Advance Release Calendar / ARC)*\n\n")

	if r.scraper != nil {
		arcItems, fresh, err := r.scraper.LatestARCSchedule(ctx, "id", 50)
		if err == nil && len(arcItems) > 0 {
			var pubs, brsList []scraper.Item
			for _, it := range arcItems {
				if strings.Contains(strings.ToLower(it.Subject), "publikasi") {
					pubs = append(pubs, it)
				} else {
					brsList = append(brsList, it)
				}
			}

			if len(pubs) > 0 {
				sb.WriteString("*A. Jadwal Rilis Publikasi Statistik*\n\n")
				maxPub := len(pubs)
				if maxPub > 10 {
					maxPub = 10
				}
				for i := 0; i < maxPub; i++ {
					it := pubs[i]
					dateStr := "-"
					if it.ReleaseDate != nil {
						dateStr = it.ReleaseDate.Format("02 Jan 2006")
					}
					statusStr := it.Abstract
					if statusStr == "" {
						statusStr = "Terjadwal"
					}
					fmt.Fprintf(&sb, "*%d. %s*\nJadwal: %s (%s)\n\n", i+1, it.Title, dateStr, statusStr)
				}
			}

			if len(brsList) > 0 {
				sb.WriteString("*B. Jadwal Rilis Berita Resmi Statistik (BRS)*\n\n")
				maxBRS := len(brsList)
				if maxBRS > 6 {
					maxBRS = 6
				}
				for i := 0; i < maxBRS; i++ {
					it := brsList[i]
					statusStr := strings.ReplaceAll(it.Abstract, " | ", ", ")
					if statusStr == "" {
						statusStr = "Terjadwal"
					}
					fmt.Fprintf(&sb, "*%d. %s*\nJadwal: %s\n\n", i+1, it.Title, statusStr)
				}
			}

			if fresh.UpdatedAt.Year() > 2000 {
				fmt.Fprintf(&sb, "_Diperbarui: %s WIB_\n", fresh.UpdatedAt.Format("02 Jan 2006 15:04"))
			}
			sb.WriteString("Portal ARC Resmi:\nhttps://bangkabaratkab.bps.go.id/id/arc")
			return sb.String()
		}
	}

	sb.WriteString("Jadwal rilis publikasi dan Berita Resmi Statistik (BRS) BPS Kabupaten Bangka Barat disusun dalam Advance Release Calendar (ARC).\n\n")
	sb.WriteString("Lihat kalender jadwal rilis selengkapnya di:\n")
	sb.WriteString("https://bangkabaratkab.bps.go.id/id/arc")
	return sb.String()
}

func (r *BotRouter) HandleFAQDetail(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage, idx int) error {
	if idx == 5 {
		faqText := r.buildARCScheduleFAQ(ctx)
		msg := whatsapp.BuildQuickReplyButtons(
			"Jadwal Rilis Data (ARC)",
			faqText,
			"PST BPS Kabupaten Bangka Barat",
			[]whatsapp.QuickReplyButton{
				{ID: "99", DisplayText: "Pertanyaan Lain (99)"},
				{ID: "0", DisplayText: "Menu Utama (0)"},
			},
		)
		return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
	}

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
			A: "Ya! BPS menerbitkan publikasi 'Kecamatan Dalam Angka' setiap tahun yang memuat data rinci hingga tingkat desa/kelurahan untuk seluruh 6 kecamatan di Bangka Barat (Mentok, Simpang Teritip, Jebus, Kelapa, Tempilang, Parittiga).",
		},
		{
			Q: "Bagaimana cara mengajukan Rekomendasi Kegiatan Statistik (ROMANTIK)?",
			A: "Bagi instansi pemerintah/OPD yang akan menyelenggarakan survei/kompilasi data statistik sektoral, silakan mengajukan permohonan rekomendasi melalui portal https://romantik.bps.go.id atau konsultasi melalui PST BPS Bangka Barat.",
		},
		{
			Q: "Kapan jadwal rilis data terbaru BPS Bangka Barat?",
			A: "Jadwal rilis publikasi dan Berita Resmi Statistik (BRS) dirilis berkala sesuai ARC (Advance Release Calendar) di https://bangkabaratkab.bps.go.id/id/arc.",
		},
		{
			Q: "Bagaimana data BPS dihasilkan?",
			A: "Data BPS secara umum dihasilkan melalui proses panjang yang terstruktur melalui GSBPM (Generic Statistical Business Process Model), sebuah standar internasional dalam menghasilkan data resmi. Mulai dari tahap perencanaan, perancangan, pembangunan, pengumpulan, pengolahan, penyajian, hingga evaluasi. Setiap langkah dilakukan secara cermat dengan melibatkan ekspertis di bidangnya agar data yang dihasilkan dapat terjaga kualitasnya.\n\nBPS menghasilkan banyak data dari ratusan pendataan yang dilakukan di seluruh Indonesia yang dikelola oleh beberapa direktorat di BPS. Statmin mengajak Sahabat Data untuk melihat gambaran proses bisnis yang ada di salah satu survei di BPS. Simak videonya di YouTube BPS Statistics.",
		},
	}

	var faqText string
	if idx >= 1 && idx <= len(faqs) {
		item := faqs[idx-1]
		faqText = fmt.Sprintf("*BANTUAN & FAQ #%d*\n"+
			"_BPS Kabupaten Bangka Barat_\n"+
			"────────────\n\n"+
			"*%s*\n\n"+
			"*Penjelasan*:\n"+
			"%s", idx, item.Q, item.A)
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
