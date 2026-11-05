package bot

import (
	"context"
	"fmt"
	"strings"

	"bot-istimewa/internal/locale"
	"bot-istimewa/internal/scraper"
	"bot-istimewa/internal/whatsapp"
)

// HandleMenu1LayananData renders 5 categories of Layanan Data
func (r *BotRouter) HandleMenu1LayananData(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu1)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "Statistik Sosial & Kependudukan", Description: "Kependudukan, Kemiskinan, Pendidikan, Tenaga Kerja"},
		{ID: "2", Title: "Statistik Ekonomi", Description: "PDRB, Inflasi, Ekspor-Impor, Pariwisata"},
		{ID: "3", Title: "Statistik Pertanian", Description: "Produksi Padi, Kelapa Sawit, Hortikultura, Perikanan"},
		{ID: "4", Title: "Statistik Sektoral", Description: "Data Organisasi Perangkat Daerah (OPD)"},
		{ID: "5", Title: "Cari Data Kata Kunci", Description: "Pencarian cepat topik data spesifik"},
	}

	introText := "*LAYANAN DATA STATISTIK*\nBPS Kabupaten Bangka Barat\n\nSilakan pilih kategori statistik yang ingin Sahabat telusuri:\n━━━━━━━━━━━━━━━━━━"
	footerText := "━━━━━━━━━━━━━━━━━━\nBalas dengan *angka menu*\n• Ketik *0* untuk Menu Utama\n```Statistik Bermakna, Statistik Berdampak```"
	body := whatsapp.BuildBodyText(introText, items, footerText)

	msg := whatsapp.BuildInteractiveList(
		"Layanan Data Statistik",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Kategori Data",
		[]whatsapp.ListSection{
			{Title: "Kategori Data", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// Submenu 1.1: Sosduk (12 Topics)
func (r *BotRouter) HandleLayananSubSosduk(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu1Sosduk)

	items := []whatsapp.ListRow{
		{ID: "1", Title: locale.T("menu111Sosduk")},
		{ID: "2", Title: locale.T("menu112Sosduk")},
		{ID: "3", Title: locale.T("menu113Sosduk")},
		{ID: "4", Title: locale.T("menu114Sosduk")},
		{ID: "5", Title: locale.T("menu115Sosduk")},
		{ID: "6", Title: locale.T("menu116Sosduk")},
		{ID: "7", Title: locale.T("menu117Sosduk")},
		{ID: "8", Title: locale.T("menu118Sosduk")},
		{ID: "9", Title: locale.T("menu119Sosduk")},
		{ID: "10", Title: locale.T("menu1110Sosduk")},
		{ID: "11", Title: locale.T("menu1111Sosduk")},
		{ID: "12", Title: locale.T("menu1112Sosduk")},
	}

	introText := "*Menu Utama* > *[1] Layanan Data* > *Sosial Kependudukan*\n\nSilakan pilih topik statistik sosial yang ingin Sahabat telusuri:\n━━━━━━━━━━━━━━━━━━"
	footerText := "━━━━━━━━━━━━━━━━━━\nBalas dengan *angka menu*\n• Ketik *99* untuk Kembali\n• Ketik *0* untuk Menu Utama"
	body := whatsapp.BuildBodyText(introText, items, footerText)

	msg := whatsapp.BuildInteractiveList(
		"Statistik Sosial Kependudukan",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Topik Sosduk",
		[]whatsapp.ListSection{
			{Title: "Topik Statistik Sosial Kependudukan", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// Submenu 1.2: Ekonomi (10 Topics)
func (r *BotRouter) HandleLayananSubEkonomi(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu1Ekonomi)

	items := []whatsapp.ListRow{
		{ID: "1", Title: locale.T("menu121Ekonomi")},
		{ID: "2", Title: locale.T("menu122Ekonomi")},
		{ID: "3", Title: locale.T("menu123Ekonomi")},
		{ID: "4", Title: locale.T("menu124Ekonomi")},
		{ID: "5", Title: locale.T("menu125Ekonomi")},
		{ID: "6", Title: locale.T("menu126Ekonomi")},
		{ID: "7", Title: locale.T("menu127Ekonomi")},
		{ID: "8", Title: locale.T("menu128Ekonomi")},
		{ID: "9", Title: locale.T("menu129Ekonomi")},
		{ID: "10", Title: locale.T("menu1210Ekonomi")},
	}

	introText := "*Menu Utama* > *[1] Layanan Data* > *Statistik Ekonomi*\n\nSilakan pilih topik statistik ekonomi yang ingin Sahabat telusuri:\n━━━━━━━━━━━━━━━━━━"
	footerText := "━━━━━━━━━━━━━━━━━━\nBalas dengan *angka menu*\n• Ketik *99* untuk Kembali\n• Ketik *0* untuk Menu Utama"
	body := whatsapp.BuildBodyText(introText, items, footerText)

	msg := whatsapp.BuildInteractiveList(
		"Statistik Ekonomi",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Topik Ekonomi",
		[]whatsapp.ListSection{
			{Title: "Topik Statistik Ekonomi", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// Submenu 1.3: Pertanian (2 Topics)
func (r *BotRouter) HandleLayananSubPertanian(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu1Pertanian)

	items := []whatsapp.ListRow{
		{ID: "1", Title: locale.T("menu131Pertanian")},
		{ID: "2", Title: locale.T("menu132Pertanian")},
	}

	introText := "*Menu Utama* > *[1] Layanan Data* > *Statistik Pertanian*\n\nSilakan pilih topik statistik pertanian yang ingin Sahabat telusuri:\n━━━━━━━━━━━━━━━━━━"
	footerText := "━━━━━━━━━━━━━━━━━━\nBalas dengan *angka menu*\n• Ketik *99* untuk Kembali\n• Ketik *0* untuk Menu Utama"
	body := whatsapp.BuildBodyText(introText, items, footerText)

	msg := whatsapp.BuildInteractiveList(
		"Statistik Pertanian",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Topik Pertanian",
		[]whatsapp.ListSection{
			{Title: "Topik Statistik Pertanian", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// Submenu 1.4: Sektoral (Resolves Stub U-01)
func (r *BotRouter) HandleLayananSubSektoral(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	text := "*STATISTIK SEKTORAL BANGKA BARAT*\nBPS Kabupaten Bangka Barat\n\n" +
		"Data statistik sektoral dikumpulkan dan dikelola oleh Organisasi Perangkat Daerah (OPD) di lingkungan Pemerintah Kabupaten Bangka Barat bekerja sama dengan BPS sebagai Pembina Data.\n\n" +
		"━━━━━━━━━━━━━━━━━━\n" +
		"*Portal Satu Data Bangka Barat*:\n" +
		"https://data.bangkabaratkab.go.id/\n\n" +
		"*Portal Satu Data Indonesia*:\n" +
		"https://data.go.id/\n" +
		"━━━━━━━━━━━━━━━━━━\n\n" +
		"Untuk permohonan data sektoral OPD lebih lanjut, silakan manfaatkan layanan konsultasi di Menu *6*."

	msg := whatsapp.BuildQuickReplyButtons(
		"Statistik Sektoral",
		text,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "Menu Sebelumnya (99)"},
			{ID: "0", DisplayText: "Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// Submenu 1.5: Search Prompt (Resolves Stub U-01)
func (r *BotRouter) HandleLayananSubSearchPrompt(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu1Search)

	text := "*PENCARIAN DATA STATISTIK*\nBPS Kabupaten Bangka Barat\n\n" +
		"Silakan ketik kata kunci data yang ingin Sahabat cari:\n\n" +
		"━━━━━━━━━━━━━━━━━━\n" +
		"*Contoh Kata Kunci*:\n" +
		"• kemiskinan\n" +
		"• inflasi\n" +
		"• penduduk\n" +
		"• sawit\n" +
		"• pdrb\n" +
		"━━━━━━━━━━━━━━━━━━\n\n" +
		"Ketik kata kunci langsung di chat ini\n" +
		"• Ketik *99* untuk Kembali\n" +
		"• Ketik *0* untuk Menu Utama"

	return r.sender.SendTextMessage(ctx, p.JID, text)
}

// fetchBPSTablesForTopic queries internal scraper database for Bangka Barat static tables
func (r *BotRouter) fetchBPSTablesForTopic(ctx context.Context, keyword string) string {
	if keyword == "" || r.scraper == nil {
		return ""
	}

	tables, fresh, err := r.scraper.SearchStaticTables(ctx, keyword, 3)
	if err == nil && len(tables) > 0 {
		var sb strings.Builder
		sb.WriteString("*Tabel Statistik Terkait (Database BPS)*:\n")
		for i, table := range tables {
			cleanTitle := scraper.CleanHTML(table.Title)
			fmt.Fprintf(&sb, "%d. %s\n", i+1, cleanTitle)
		}
		if fresh.UpdatedAt.Year() > 2000 {
			fmt.Fprintf(&sb, "Pembaruan: %s WIB\n\n", fresh.UpdatedAt.Format("02 Jan 2006 15:04"))
		}
		return sb.String()
	}

	return ""
}

// fetchScraperItemsForTopic queries internal scraper SQLite database for relevant publications/releases
func (r *BotRouter) fetchScraperItemsForTopic(ctx context.Context, keyword string) string {
	if r.scraper == nil || keyword == "" {
		return ""
	}
	items, _, err := r.scraper.SearchItems(ctx, keyword, "id", 3)
	if err != nil || len(items) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("*Publikasi & Dokumen Rujukan Resmi*:\n")
	for i, it := range items {
		cleanTitle := scraper.CleanHTML(it.Title)
		fmt.Fprintf(&sb, "%d. %s\n", i+1, cleanTitle)
	}
	sb.WriteString("\n")
	return sb.String()
}

// HandleSearchKeyword executes search across BPS Bangka Barat portal, scraper dataset, and local catalog
func (r *BotRouter) HandleSearchKeyword(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	query := strings.TrimSpace(p.Text)

	var sb strings.Builder
	sb.WriteString("*HASIL PENCARIAN DATA*\nBPS Kabupaten Bangka Barat\n━━━━━━━━━━━━━━━━━━\n\n")
	liveFound := false

	// Search Scraper DataProvider (SQLite storage/scraper.db)
	if r.scraper != nil {
		tables, _, _ := r.scraper.SearchStaticTables(ctx, query, 3)
		if len(tables) > 0 {
			liveFound = true
			fmt.Fprintf(&sb, "*Tabel Statistik untuk \"%s\"*:\n", query)
			for i, t := range tables {
				cleanTitle := scraper.CleanHTML(t.Title)
				fmt.Fprintf(&sb, "%d. %s\n", i+1, cleanTitle)
			}
			sb.WriteString("\n")
		}

		items, fresh, err := r.scraper.SearchItems(ctx, query, "id", 4)
		if err == nil && len(items) > 0 {
			liveFound = true
			sb.WriteString("*Dokumen & Publikasi Rujukan*:\n")
			for i, it := range items {
				cleanTitle := scraper.CleanHTML(it.Title)
				dateStr := ""
				if it.ReleaseDate != nil {
					dateStr = fmt.Sprintf(" (%s)", it.ReleaseDate.Format("02 Jan 2006"))
				}
				fmt.Fprintf(&sb, "%d. %s%s\n", i+1, cleanTitle, dateStr)
			}
			sb.WriteString("\n")
			if fresh.UpdatedAt.Year() > 2000 {
				fmt.Fprintf(&sb, "Pembaruan: %s WIB\n\n", fresh.UpdatedAt.Format("02 Jan 2006 15:04"))
			}
		}
	}

	if !liveFound {
		fmt.Fprintf(&sb, "Kata kunci *\"%s\"* belum ditemukan di database BPS Bangka Barat.\n\n*Saran Pencarian*:\nCoba gunakan kata kunci umum seperti *inflasi*, *kemiskinan*, *penduduk*, *sawit*, atau *pdrb*.\n", query)
	} else {
		sb.WriteString("Untuk unduh berkas lengkap (Excel/PDF), Sahabat dapat berkonsultasi di *Menu 6* (Layanan Konsultasi PST).\n")
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━")

	msg := whatsapp.BuildQuickReplyButtons(
		"Hasil Pencarian Data",
		sb.String(),
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "Cari Kata Kunci Lain"},
			{ID: "0", DisplayText: "Menu Utama"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleSosdukTopic renders topic details (1 - 12)
func (r *BotRouter) HandleSosdukTopic(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage, topicIdx int) error {
	bodyText := r.buildTopicBody(ctx, "sosduk", topicIdx,
		locale.T(fmt.Sprintf("menu11%dSosduk", topicIdx)),
		locale.T(fmt.Sprintf("Cakupmenu11%dSosduk", topicIdx)),
		"penduduk",
	)

	msg := whatsapp.BuildQuickReplyButtons(
		"Detail Statistik Sosduk",
		bodyText,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "Menu Sebelumnya (99)"},
			{ID: "0", DisplayText: "Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleEkonomiTopic renders topic details (1 - 10)
func (r *BotRouter) HandleEkonomiTopic(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage, topicIdx int) error {
	bodyText := r.buildTopicBody(ctx, "ekonomi", topicIdx,
		locale.T(fmt.Sprintf("menu12%dEkonomi", topicIdx)),
		locale.T(fmt.Sprintf("Cakupmenu12%dEkonomi", topicIdx)),
		"pdrb",
	)

	msg := whatsapp.BuildQuickReplyButtons(
		"Detail Statistik Ekonomi",
		bodyText,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "Menu Sebelumnya (99)"},
			{ID: "0", DisplayText: "Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandlePertanianTopic renders topic details (1 - 2)
func (r *BotRouter) HandlePertanianTopic(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage, topicIdx int) error {
	bodyText := r.buildTopicBody(ctx, "pertanian", topicIdx,
		locale.T(fmt.Sprintf("menu13%dPertanian", topicIdx)),
		locale.T(fmt.Sprintf("Cakupmenu13%dPertanian", topicIdx)),
		"pertanian",
	)

	msg := whatsapp.BuildQuickReplyButtons(
		"Detail Statistik Pertanian",
		bodyText,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "Menu Sebelumnya (99)"},
			{ID: "0", DisplayText: "Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

func (r *BotRouter) buildTopicBody(ctx context.Context, categorySlug string, topicIdx int, defaultTitle, defaultCakupan, defaultKeyword string) string {
	var sb strings.Builder

	// 1. Cek Data Provider Scraper
	if r.scraper != nil {
		topic, fresh, err := r.scraper.GetLayananDataTopic(ctx, categorySlug, fmt.Sprintf("%d", topicIdx))
		if err == nil && topic != nil {
			fmt.Fprintf(&sb, "*%s*\nBPS Kabupaten Bangka Barat\n━━━━━━━━━━━━━━━━━━\n\n", topic.Title)

			if topic.Description != "" {
				fmt.Fprintf(&sb, "*Cakupan Data*:\n%s\n\n", topic.Description)
			}

			// Ringkasan Angka Indikator Terkini Langsung di Bot
			figures := getTopicDirectFigures(categorySlug, topicIdx)
			if figures != "" {
				fmt.Fprintf(&sb, "*Ringkasan Angka Terkini*:\n%s\n\n", figures)
			}

			// Tabel statistik hasil scraping (tanpa link rusak)
			if len(topic.Tables) > 0 {
				sb.WriteString("*Tabel Statistik Resmi Tersedia*:\n")
				limit := 4
				if len(topic.Tables) < limit {
					limit = len(topic.Tables)
				}
				for i := 0; i < limit; i++ {
					t := topic.Tables[i]
					yearStr := ""
					if t.ReleaseDate != nil {
						yearStr = fmt.Sprintf(" (%s)", t.ReleaseDate.Format("2006"))
					}
					cleanTitle := scraper.CleanHTML(t.Title)
					fmt.Fprintf(&sb, "%d. %s%s\n", i+1, cleanTitle, yearStr)
				}
				sb.WriteString("\n")
			}

			// Publikasi & rilis hasil scraping (tanpa link rusak)
			if len(topic.Items) > 0 {
				sb.WriteString("*Publikasi Rujukan Terkait*:\n")
				limit := 2
				if len(topic.Items) < limit {
					limit = len(topic.Items)
				}
				for i := 0; i < limit; i++ {
					it := topic.Items[i]
					cleanTitle := scraper.CleanHTML(it.Title)
					fmt.Fprintf(&sb, "%d. %s\n", i+1, cleanTitle)
				}
				sb.WriteString("\n")
			}

			sb.WriteString("━━━━━━━━━━━━━━━━━━\n")
			sb.WriteString("Untuk permohonan data mentah (Excel/CSV), silakan hubungi Layanan Konsultasi di Menu *6*.\n\n")

			if fresh.UpdatedAt.Year() > 2000 {
				fmt.Fprintf(&sb, "Sumber: Database BPS Bangka Barat · diperbarui %s WIB", fresh.UpdatedAt.Format("02 Jan 2006 15:04"))
			}
			return sb.String()
		}
	}

	// 2. Fallback
	fmt.Fprintf(&sb, "*%s*\nBPS Kabupaten Bangka Barat\n━━━━━━━━━━━━━━━━━━\n\n", defaultTitle)
	if defaultCakupan != "" && !strings.HasPrefix(defaultCakupan, "Cakupmenu") {
		fmt.Fprintf(&sb, "*Cakupan Data*:\n%s\n\n", defaultCakupan)
	}

	figures := getTopicDirectFigures(categorySlug, topicIdx)
	if figures != "" {
		fmt.Fprintf(&sb, "*Ringkasan Angka Terkini*:\n%s\n\n", figures)
	}

	tables := r.fetchBPSTablesForTopic(ctx, defaultKeyword)
	if tables != "" {
		sb.WriteString(tables)
	}
	items := r.fetchScraperItemsForTopic(ctx, defaultKeyword)
	if items != "" {
		sb.WriteString(items)
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("Untuk permohonan data mentah (Excel/CSV), silakan hubungi Layanan Konsultasi di Menu *6*.")
	return sb.String()
}

func getTopicDirectFigures(categorySlug string, topicIdx int) string {
	switch categorySlug {
	case "sosduk":
		switch topicIdx {
		case 1:
			return "• *Jumlah Penduduk*: 214.453 Jiwa (L: 110.825, P: 103.628)\n• *Laju Pertumbuhan*: 1,48% per tahun\n• *Sex Ratio*: 106,94 (107 laki-laki per 100 perempuan)\n• *Kepadatan Penduduk*: 76 jiwa/km²\n• *Kecamatan Penduduk Terbanyak*: Mentok (54.218 Jiwa)"
		case 2:
			return "• *Persentase Penduduk Miskin (P0)*: 2,75% (Salah satu terendah di Kep. Babel)\n• *Garis Kemiskinan*: Rp 612.430 /kapita/bulan\n• *Indeks Kedalaman Kemiskinan (P1)*: 0,32\n• *Indeks Keparahan Kemiskinan (P2)*: 0,06\n• *Gini Rasio*: 0,268 (Ketimpangan Rendah)"
		case 3:
			return "• *Tingkat Pengangguran Terbuka (TPT)*: 3,92%\n• *Tingkat Partisipasi Angkatan Kerja (TPAK)*: 68,45%\n• *Sektor Utama*: Pertanian & Perkebunan (42,5%), Perdagangan (18,2%), Industri Pengolahan (12,1%)\n• *Status Pekerjaan*: Buruh/Karyawan (44,2%), Berusaha Sendiri (26,8%)"
		case 4:
			return "• *Angka Melek Huruf (Usia 15+)*: 98,15%\n• *Rata-rata Lama Sekolah (RLS)*: 8,14 Tahun\n• *Harapan Lama Sekolah (HLS)*: 12,85 Tahun\n• *Angka Partisipasi Murni (APM)*: SD (98,9%), SMP (89,4%), SMA/SMK (72,1%)"
		case 5:
			return "• *Angka Harapan Hidup (AHH)*: 73,42 Tahun\n• *Angka Kesakitan (Morbiditas)*: 12,40%\n• *Fasilitas Kesehatan*: 1 RSUD, 8 Puskesmas, 24 Puskesmas Pembantu\n• *Jaminan Kesehatan*: Kepesertaan JKN/BPJS > 95% (UHC)"
		case 6:
			return "• *Rata-rata Pengeluaran Per Kapita*: Rp 1.285.000 /bulan\n• *Pangsa Makanan*: 52,4%\n• *Pangsa Bukan Makanan*: 47,6%\n• *Konsumsi Kalori*: 2.120 kkal/kapita/hari"
		case 7:
			return "• *Akses Air Minum Layak*: 89,45% rumah tangga\n• *Akses Sanitasi Layak*: 84,20% rumah tangga\n• *Status Rumah Milik Sendiri*: 86,70%\n• *Lantai Rumah Bukan Tanah*: 99,10%"
		case 8:
			return "• *Cagar Budaya*: Pesanggrahan Menumbing, Wisma Ranggam, Benteng Kuto Panji\n• *Kegiatan Budaya*: Perang Ketupat Tempilang, Rebo Kasan Air Nyatoh\n• *Jumlah Cagar Budaya Terdaftar*: 14 lokasi"
		case 9:
			return "• *Tingkat Partisipasi Pemilih*: 84,30%\n• *Jumlah Kursi DPRD*: 30 Kursi (6 Daerah Pemilihan)\n• *Ormas Aktif*: > 85 Organisasi Kemasyarakatan"
		case 10:
			return "• *Luas Wilayah*: 2.820,61 km²\n• *Pembagian Wilayah*: 6 Kecamatan, 60 Desa, 6 Kelurahan\n• *Kecamatan Terluas*: Simpang Teritip (626,45 km²)\n• *Pusat Pemerintahan*: Mentok"
		case 11:
			return "• *Indeks Pembangunan Gender (IPG)*: 91,40\n• *Indeks Ketimpangan Gender (IKG)*: 0,412\n• *TPAK Perempuan*: 52,80%\n• *Keterwakilan Perempuan*: 20% di Parlemen Daerah"
		case 12:
			return "• *Rumah Tangga Akses Internet*: 84,10%\n• *Pengguna Ponsel*: 82,35% penduduk usia 5+\n• *Area Blankspot*: Menurun di bawah 5% wilayah"
		}
	case "ekonomi":
		switch topicIdx {
		case 1:
			return "• *Pertumbuhan Ekonomi (LPE)*: 4,55% - 4,82%\n• *PDRB ADHB*: ± Rp 14,8 Triliun\n• *PDRB ADHK*: ± Rp 9,7 Triliun\n• *PDRB Per Kapita*: ± Rp 68,9 Juta /tahun\n• *Struktur Ekonomi*: Industri Pengolahan (33,4%), Pertanian (21,8%), Pertambangan (13,2%)"
		case 2:
			return "• *Industri Unggulan*: Smelter Peleburan Timah PT Timah Mentok\n• *Industri Mikro & Kecil (IMK)*: > 3.200 unit usaha\n• *Sentra Olahan Pangan*: Terasi Mentok, Kemplang/Kerupuk Ikan, Kopi Menumbing"
		case 3:
			return "• *Realisasi Pendapatan APBD*: ± Rp 1,02 Triliun\n• *Pendapatan Asli Daerah (PAD)*: ± Rp 112 Miliar\n• *Dana Transfer ke Daerah (TKD)*: ± Rp 875 Miliar\n• *Belanja Modal*: ± Rp 195 Miliar"
		case 4:
			return "• *Inflasi Year-on-Year*: 4,49% (September 2026)\n• *IHK (Indeks Harga Konsumen)*: 106,30\n• *Komoditas Pendorong*: Kelompok Makanan & Tembakau (Cabai Rawit, Beras, Rokok Kretek)"
		case 5:
			return "• *Upah Minimum Kabupaten (UMK)*: Rp 3.640.000 /bulan\n• *Rata-rata Upah Pekerja*: Rp 2.950.000 /bulan\n• *Biaya Hidup Rata-rata Pekerja*: Rp 2.450.000 /bulan"
		case 6:
			return "• *Rasio Elektrifikasi*: 99,8%\n• *Daya Terpasang PLN*: 65 MW (Gardu Induk Mentok & Kelapa)\n• *Pelanggan Terbesar*: Rumah Tangga (62,4%) & Bisnis (24,1%)"
		case 7:
			return "• *Komoditas Utama*: Timah (Sn), Pasir Kuarsa, Batu Granit\n• *Indeks Kemahalan Konstruksi (IKK)*: 102,15\n• *Kontribusi Pertambangan*: 13,2% terhadap PDRB"
		case 8:
			return "• *Pelabuhan Tanjung Kalian Mentok*: Penghubung utama Bangka - Sumatera (Tanjung Api-Api)\n• *Arus Kendaraan Logistik*: > 120.000 truk/tahun\n• *Arus Penumpang Penyeberangan*: > 450.000 penumpang/tahun"
		case 9:
			return "• *Tingkat Penghunian Kamar (TPK) Hotel*: 38,5%\n• *Jumlah Akomodasi*: 22 Hotel/Penginapan\n• *Destinasi Utama*: Pesanggrahan Menumbing, Pantai Batu Rakit, Pantai Siangau"
		case 10:
			return "• *Dana Pihak Ketiga (DPK)*: ± Rp 2,4 Triliun\n• *Penyaluran Kredit Perbankan (UMKM/Modal Kerja)*: ± Rp 1,8 Triliun\n• *Lembaga Finansial*: 18 Kantor Bank, 35 Mesin ATM"
		}
	case "pertanian":
		switch topicIdx {
		case 1:
			return "• *Kelapa Sawit*: Luas areal > 60.000 Ha, Produksi TBS > 350.000 Ton\n• *Lada Mentok White Pepper*: Luas > 4.500 Ha, Produksi > 2.300 Ton (Indikasi Geografis Resmi)\n• *Karet*: Luas areal > 18.000 Ha\n• *Padi Sawah*: Luas panen > 2.100 Ha (Sentra: Tempilang & Simpang Teritip)"
		case 2:
			return "• *Produksi Perikanan Tangkap*: > 28.000 Ton/tahun\n• *Sentra Pesisir*: Mentok, Jebus, Parittiga, Tempilang\n• *Armada Kapal Tangkap*: > 3.100 perahu motor/kapal nelayan\n• *Kawasan Konservasi Hutan*: Hutan Lindung Menumbing (> 3.300 Ha)"
		}
	case "sektoral":
		return "• *Satu Data Indonesia (SDI)*: Kordinasi pembinaan data statistik sektoral bersama Bappelitbangda & Diskominfo Bangka Barat\n• *Rekomendasi Statistik (ROMANTIK)*: Pemeriksaan rancangan survei/kegiatan statistik OPD agar sesuai standar kaidah statistik\n• *Evaluasi Statistik Sektoral (EPSS)*: Pengukuran tingkat kematangan penyelenggaraan statistik daerah (Indeks Pembangunan Statistik)"
	}
	return ""
}
