package bot

import (
	"context"
	"fmt"
	"strings"

	"bot-istimewa/internal/bpsapi"
	"bot-istimewa/internal/locale"
	"bot-istimewa/internal/whatsapp"
)

// HandleMenu1_LayananData renders 5 categories of Layanan Data
func (r *BotRouter) HandleMenu1_LayananData(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu1)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "1. Statistik Sosial Kependudukan", Description: "Kependudukan, Kemiskinan, Pendidikan, Tenaga Kerja"},
		{ID: "2", Title: "2. Statistik Ekonomi", Description: "PDRB, Inflasi, Ekspor-Impor, Pariwisata"},
		{ID: "3", Title: "3. Statistik Pertanian", Description: "Produksi Padi, Kelapa Sawit, Hortikultura"},
		{ID: "4", Title: "4. Statistik Sektoral", Description: "Data Organisasi Perangkat Daerah (OPD)"},
		{ID: "5", Title: "5. Cari Data Kata Kunci", Description: "Pencarian cepat topik data spesifik"},
	}

	introText := "*Layanan Data BPS*\nSilakan pilih kategori statistik yang ingin Sahabat telusuri \n\n●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●\n\n*Pilih nomor kategori data:*"
	body := whatsapp.BuildBodyText(introText, items, "\n*0.* Menu Awal\n\n◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇\n```Statistik Bermakna, Statistik Berdampak```")

	msg := whatsapp.BuildInteractiveList(
		"Layanan Data Statistik",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Kategori Data ",
		[]whatsapp.ListSection{
			{Title: "Kategori Data", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// Submenu 1.1: Sosduk (12 Topics)
func (r *BotRouter) HandleLayananSub_Sosduk(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu1_Sosduk)

	items := []whatsapp.ListRow{
		{ID: "1", Title: locale.T("menu111Sosduk"), Description: "Kependudukan & Migrasi"},
		{ID: "2", Title: locale.T("menu112Sosduk"), Description: "Kemiskinan & Sosial Lintas Sektor"},
		{ID: "3", Title: locale.T("menu113Sosduk"), Description: "Tenaga Kerja"},
		{ID: "4", Title: locale.T("menu114Sosduk"), Description: "Pendidikan"},
		{ID: "5", Title: locale.T("menu115Sosduk"), Description: "Kesehatan"},
		{ID: "6", Title: locale.T("menu116Sosduk"), Description: "Konsumsi dan Pendapatan"},
		{ID: "7", Title: locale.T("menu117Sosduk"), Description: "Pemukiman & Perumahan"},
		{ID: "8", Title: locale.T("menu118Sosduk"), Description: "Budaya"},
		{ID: "9", Title: locale.T("menu119Sosduk"), Description: "Aktivitas Politik & Komunitas"},
		{ID: "10", Title: locale.T("menu1110Sosduk"), Description: "Regional & Area Kecil"},
		{ID: "11", Title: locale.T("menu1111Sosduk"), Description: "Gender & Kelompok Khusus"},
		{ID: "12", Title: locale.T("menu1112Sosduk"), Description: "Masyarakat Informasi"},
	}

	introText := locale.T("introMenu11", p.Name)
	footerText := locale.T("footerMenu11")
	body := whatsapp.BuildBodyText(introText, items, footerText)

	msg := whatsapp.BuildInteractiveList(
		"Statistik Sosial Kependudukan",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Topik Sosduk ",
		[]whatsapp.ListSection{
			{Title: "Topik Statistik Sosial Kependudukan", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// Submenu 1.2: Ekonomi (10 Topics)
func (r *BotRouter) HandleLayananSub_Ekonomi(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu1_Ekonomi)

	items := []whatsapp.ListRow{
		{ID: "1", Title: locale.T("menu121Ekonomi"), Description: "Neraca Ekonomi"},
		{ID: "2", Title: locale.T("menu122Ekonomi"), Description: "Statistik Bisnis"},
		{ID: "3", Title: locale.T("menu123Ekonomi"), Description: "Keuangan Pemerintah & Sektor Publik"},
		{ID: "4", Title: locale.T("menu124Ekonomi"), Description: "Harga-Harga"},
		{ID: "5", Title: locale.T("menu125Ekonomi"), Description: "Biaya Tenaga Kerja"},
		{ID: "6", Title: locale.T("menu126Ekonomi"), Description: "Energi"},
		{ID: "7", Title: locale.T("menu127Ekonomi"), Description: "Pertambangan, Manufaktur & Konstruksi"},
		{ID: "8", Title: locale.T("menu128Ekonomi"), Description: "Transportasi"},
		{ID: "9", Title: locale.T("menu129Ekonomi"), Description: "Pariwisata"},
		{ID: "10", Title: locale.T("menu1210Ekonomi"), Description: "Perbankan, Asuransi & Finansial"},
	}

	introText := locale.T("introMenu12", p.Name)
	footerText := locale.T("footerMenu12")
	body := whatsapp.BuildBodyText(introText, items, footerText)

	msg := whatsapp.BuildInteractiveList(
		"Statistik Ekonomi",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Topik Ekonomi ",
		[]whatsapp.ListSection{
			{Title: "Topik Statistik Ekonomi", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// Submenu 1.3: Pertanian (2 Topics)
func (r *BotRouter) HandleLayananSub_Pertanian(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu1_Pertanian)

	items := []whatsapp.ListRow{
		{ID: "1", Title: locale.T("menu131Pertanian"), Description: "Pertanian, Kehutanan & Perikanan"},
		{ID: "2", Title: locale.T("menu132Pertanian"), Description: "Statistik Lingkungan"},
	}

	introText := locale.T("introMenu13", p.Name)
	footerText := locale.T("footerMenu13")
	body := whatsapp.BuildBodyText(introText, items, footerText)

	msg := whatsapp.BuildInteractiveList(
		"Statistik Pertanian",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Topik Pertanian ",
		[]whatsapp.ListSection{
			{Title: "Topik Statistik Pertanian", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// Submenu 1.4: Sektoral (Resolves Stub U-01)
func (r *BotRouter) HandleLayananSub_Sektoral(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	text := "*Statistik Sektoral Kabupaten Bangka Barat*\n\nData statistik sektoral dikumpulkan dan dikelola oleh Organisasi Perangkat Daerah (OPD) di lingkungan Pemerintah Kabupaten Bangka Barat bekerja sama dengan BPS sebagai Pembina Data.\n\n*Portal Satu Data Bangka Barat*:\nhttps://data.bangkabaratkab.go.id/\n\n*Portal Satu Data Indonesia*:\nhttps://data.go.id/\n\nUntuk permohonan data sektoral OPD lebih lanjut, silakan manfaatkan layanan konsultasi di Menu *6*."

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
func (r *BotRouter) HandleLayananSub_SearchPrompt(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu1_Search)

	text := "*Cari Data Berdasarkan Kata Kunci*\n\nSilakan ketik kata kunci data yang ingin Sahabat cari \n\nContoh:\n- *kemiskinan*\n- *inflasi*\n- *penduduk*\n- *pdrb*\n\n_Ketik *99* untuk kembali atau *0* untuk ke menu utama._"

	return r.sender.SendTextMessage(ctx, p.JID, text)
}

// fetchBPSTablesForTopic queries BPS Web API for Bangka Barat (domain: 1903) static tables, falling back to scraper
func (r *BotRouter) fetchBPSTablesForTopic(ctx context.Context, keyword string) string {
	if keyword == "" {
		return ""
	}

	// 1. Try BPS Web API
	if r.bps != nil {
		res, err := r.bps.GetStaticTables(ctx, bpsapi.ListParams{
			Domain:  "1903",
			Keyword: keyword,
			Page:    1,
		})
		if err == nil && len(res.Items) > 0 {
			var sb strings.Builder
			sb.WriteString("*Tabel Statistik BPS Bangka Barat (Live Portal)*:\n")
			limit := 3
			if len(res.Items) < limit {
				limit = len(res.Items)
			}
			for i := 0; i < limit; i++ {
				table := res.Items[i]
				sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, table.Title))
				if table.Excel != "" {
					sb.WriteString(fmt.Sprintf("   Unduh Excel: %s\n", table.Excel))
				}
			}
			sb.WriteString("*Katalog Tabel*: https://bangkabaratkab.bps.go.id/id/statistics-table\n\n")
			return sb.String()
		}
	}

	// 2. Fallback to Scraper DataProvider (Priority 2) (§4.2)
	if r.scraper != nil {
		tables, fresh, err := r.scraper.SearchStaticTables(ctx, keyword, 3)
		if err == nil && len(tables) > 0 {
			var sb strings.Builder
			sb.WriteString("*Tabel Statistik BPS Bangka Barat (Portal BPS)*:\n")
			for i, table := range tables {
				sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, table.Title))
				if table.FileURL != "" {
					sb.WriteString(fmt.Sprintf("   Unduh Excel: %s\n", table.FileURL))
				} else if table.URL != "" {
					sb.WriteString(fmt.Sprintf("   Lihat: %s\n", table.URL))
				}
			}
			sb.WriteString(fmt.Sprintf("ℹSumber: bangkabaratkab.bps.go.id · diperbarui %s WIB\n\n", fresh.UpdatedAt.Format("02 Jan 2006 15:04")))
			return sb.String()
		}
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
	sb.WriteString("*Publikasi & Rilis Terkait (Database Scraper BPS)*:\n")
	for i, it := range items {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, it.Title))
		if it.FileURL != "" {
			sb.WriteString(fmt.Sprintf("   Unduh: %s\n", it.FileURL))
		} else if it.URL != "" {
			sb.WriteString(fmt.Sprintf("   Link: %s\n", it.URL))
		}
	}
	sb.WriteString("\n")
	return sb.String()
}

// HandleSearchKeyword executes search across BPS Bangka Barat portal, scraper dataset, and local catalog
func (r *BotRouter) HandleSearchKeyword(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	query := strings.TrimSpace(p.Text)

	var sb strings.Builder
	liveFound := false

	// 1. Live search from BPS Bangka Barat Web API (domain: 1903)
	if r.bps != nil {
		liveTables, err := r.bps.GetStaticTables(ctx, bpsapi.ListParams{
			Domain:  "1903",
			Keyword: query,
			Page:    1,
		})
		if err == nil && len(liveTables.Items) > 0 {
			liveFound = true
			sb.WriteString("*Tabel BPS Bangka Barat (Live Portal)*:\n")
			limit := 3
			if len(liveTables.Items) < limit {
				limit = len(liveTables.Items)
			}
			for i := 0; i < limit; i++ {
				t := liveTables.Items[i]
				sb.WriteString(fmt.Sprintf("• *%s*\n", t.Title))
				if t.Excel != "" {
					sb.WriteString(fmt.Sprintf("  Unduh Excel: %s\n", t.Excel))
				}
			}
			sb.WriteString("\n")
		}
	}

	// 2. Search Scraper DataProvider (SQLite storage/scraper.db)
	if r.scraper != nil {
		items, fresh, err := r.scraper.SearchItems(ctx, query, "id", 5)
		if err == nil && len(items) > 0 {
			liveFound = true
			sb.WriteString("*Dokumen & Rilis BPS Bangka Barat*:\n\n")
			for i, it := range items {
				sb.WriteString(fmt.Sprintf("%d. *%s*\n", i+1, it.Title))
				if it.ReleaseDate != nil {
					sb.WriteString(fmt.Sprintf("   Rilis: %s\n", it.ReleaseDate.Format("02 Jan 2006")))
				}
				if it.FileURL != "" {
					sb.WriteString(fmt.Sprintf("   Unduh: %s\n", it.FileURL))
				} else if it.URL != "" {
					sb.WriteString(fmt.Sprintf("   Link: %s\n", it.URL))
				}
				sb.WriteString("\n")
			}
			sb.WriteString(fmt.Sprintf("ℹSumber: Database Scraper BPS · diperbarui %s WIB\n\n", fresh.UpdatedAt.Format("02 Jan 2006 15:04")))
		}
	}

	if !liveFound {
		sb.WriteString(fmt.Sprintf("Maaf Sahabat Data, kata kunci *\"%s\"* belum ditemukan di database rilis/publikasi BPS Bangka Barat.\n\nSilakan coba kata kunci lain (misal: *inflasi*, *kemiskinan*, *penduduk*, *kelapa*, *pdrb*) atau kunjungi portal resmi:\nhttps://bangkabaratkab.bps.go.id\n", query))
	}

	msg := whatsapp.BuildQuickReplyButtons(
		"Hasil Pencarian Data",
		sb.String(),
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "Cari Kata Kunci Lain"},
			{ID: "0", DisplayText: "Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleSosdukTopic renders topic details (1 - 12)
func (r *BotRouter) HandleSosdukTopic(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage, topicIdx int) error {
	cakupKey := fmt.Sprintf("Cakupmenu11%dSosduk", topicIdx)
	linkKey := fmt.Sprintf("Linkmenu11%dSosduk", topicIdx)
	cakupan := locale.T(cakupKey)
	link := locale.T(linkKey)

	// Keyword mapping for BPS Bangka Barat portal
	var topicKeyword string
	switch topicIdx {
	case 1:
		topicKeyword = "kependudukan"
	case 2:
		topicKeyword = "kemiskinan"
	case 3:
		topicKeyword = "tenaga kerja"
	case 4:
		topicKeyword = "pendidikan"
	case 5:
		topicKeyword = "kesehatan"
	case 6:
		topicKeyword = "konsumsi"
	case 7:
		topicKeyword = "perumahan"
	case 8:
		topicKeyword = "budaya"
	case 9:
		topicKeyword = "politik"
	case 10:
		topicKeyword = "geografi"
	case 11:
		topicKeyword = "gender"
	case 12:
		topicKeyword = "informasi"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*Detail Statistik: %s*\n\n", locale.T(fmt.Sprintf("menu11%dSosduk", topicIdx))))
	if cakupan != "" && cakupan != cakupKey {
		sb.WriteString(cakupan + "\n\n")
	}
	if link != "" && link != linkKey {
		sb.WriteString(fmt.Sprintf("*Kunjungi Website BPS*:\n%s\n\n", link))
	}

	// 1. Live tables from BPS Bangka Barat portal
	liveTables := r.fetchBPSTablesForTopic(ctx, topicKeyword)
	if liveTables != "" {
		sb.WriteString(liveTables)
	} else {
		// 2. Scraped publications & releases from local SQLite scraper
		relatedItems := r.fetchScraperItemsForTopic(ctx, topicKeyword)
		if relatedItems != "" {
			sb.WriteString(relatedItems)
		}
	}

	msg := whatsapp.BuildQuickReplyButtons(
		"Detail Statistik Sosduk",
		sb.String(),
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
	cakupKey := fmt.Sprintf("Cakupmenu12%dEkonomi", topicIdx)
	linkKey := fmt.Sprintf("Linkmenu12%dEkonomi", topicIdx)
	cakupan := locale.T(cakupKey)
	link := locale.T(linkKey)

	var topicKeyword string
	switch topicIdx {
	case 1:
		topicKeyword = "pdrb"
	case 2:
		topicKeyword = "industri"
	case 3:
		topicKeyword = "keuangan"
	case 4:
		topicKeyword = "harga"
	case 5:
		topicKeyword = "upah"
	case 6:
		topicKeyword = "energi"
	case 7:
		topicKeyword = "pertambangan"
	case 8:
		topicKeyword = "transportasi"
	case 9:
		topicKeyword = "pariwisata"
	case 10:
		topicKeyword = "keuangan"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*Detail Statistik: %s*\n\n", locale.T(fmt.Sprintf("menu12%dEkonomi", topicIdx))))
	if cakupan != "" && cakupan != cakupKey {
		sb.WriteString(cakupan + "\n\n")
	}
	if link != "" && link != linkKey {
		sb.WriteString(fmt.Sprintf("*Kunjungi Website BPS*:\n%s\n\n", link))
	}

	// 1. Live tables from BPS Bangka Barat portal
	liveTables := r.fetchBPSTablesForTopic(ctx, topicKeyword)
	if liveTables != "" {
		sb.WriteString(liveTables)
	} else {
		// 2. Scraped publications & releases from local SQLite scraper
		relatedItems := r.fetchScraperItemsForTopic(ctx, topicKeyword)
		if relatedItems != "" {
			sb.WriteString(relatedItems)
		}
	}

	msg := whatsapp.BuildQuickReplyButtons(
		"Detail Statistik Ekonomi",
		sb.String(),
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
	cakupKey := fmt.Sprintf("Cakupmenu13%dPertanian", topicIdx)
	linkKey := fmt.Sprintf("Linkmenu13%dPertanian", topicIdx)
	cakupan := locale.T(cakupKey)
	link := locale.T(linkKey)

	var topicKeyword string
	switch topicIdx {
	case 1:
		topicKeyword = "padi"
	case 2:
		topicKeyword = "hortikultura"
	case 3:
		topicKeyword = "sawit"
	default:
		topicKeyword = "pertanian"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*Detail Statistik: %s*\n\n", locale.T(fmt.Sprintf("menu13%dPertanian", topicIdx))))
	if cakupan != "" && cakupan != cakupKey {
		sb.WriteString(cakupan + "\n\n")
	}
	if link != "" && link != linkKey {
		sb.WriteString(fmt.Sprintf("*Kunjungi Website BPS*:\n%s\n\n", link))
	}

	// 1. Live tables from BPS Bangka Barat portal
	liveTables := r.fetchBPSTablesForTopic(ctx, topicKeyword)
	if liveTables != "" {
		sb.WriteString(liveTables)
	} else {
		// 2. Scraped publications & releases from local SQLite scraper
		relatedItems := r.fetchScraperItemsForTopic(ctx, topicKeyword)
		if relatedItems != "" {
			sb.WriteString(relatedItems)
		}
	}

	msg := whatsapp.BuildQuickReplyButtons(
		"Detail Statistik Pertanian",
		sb.String(),
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "Menu Sebelumnya (99)"},
			{ID: "0", DisplayText: "Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}
