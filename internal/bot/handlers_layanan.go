package bot

import (
	"context"
	"fmt"
	"strings"

	"bot-istimewa/internal/gsheet"
	"bot-istimewa/internal/locale"
	"bot-istimewa/internal/whatsapp"
)

// HandleMenu1_LayananData renders 5 categories of Layanan Data
func (r *BotRouter) HandleMenu1_LayananData(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu1)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "1️⃣  Statistik Sosial Kependudukan", Description: "Kependudukan, Kemiskinan, Pendidikan, Tenaga Kerja"},
		{ID: "2", Title: "2️⃣  Statistik Ekonomi", Description: "PDRB, Inflasi, Ekspor-Impor, Pariwisata"},
		{ID: "3", Title: "3️⃣  Statistik Pertanian", Description: "Produksi Padi, Kelapa Sawit, Hortikultura"},
		{ID: "4", Title: "4️⃣  Statistik Sektoral", Description: "Data Organisasi Perangkat Daerah (OPD)"},
		{ID: "5", Title: "5️⃣  Cari Data Kata Kunci", Description: "Pencarian cepat topik data spesifik"},
	}

	introText := "📖 *Layanan Data BPS*\nSilakan pilih kategori statistik yang ingin Sahabat telusuri 💁‍♀️\n\n●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●\n\n📋 *Pilih nomor kategori data:*"
	body := whatsapp.BuildBodyText(introText, items, "\n*0.* 📑 Menu Awal\n\n◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇\n```Statistik Bermakna, Statistik Berdampak```")

	msg := whatsapp.BuildInteractiveList(
		"📖 Layanan Data Statistik",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Kategori Data 📊",
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
		"📂 Statistik Sosial Kependudukan",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Topik Sosduk 📋",
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
		"💰 Statistik Ekonomi",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Topik Ekonomi 📋",
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
		"🌾 Statistik Pertanian",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Topik Pertanian 📋",
		[]whatsapp.ListSection{
			{Title: "Topik Statistik Pertanian", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// Submenu 1.4: Sektoral (Resolves Stub U-01)
func (r *BotRouter) HandleLayananSub_Sektoral(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	text := "🏛️ *Statistik Sektoral Kabupaten Bangka Barat*\n\nData statistik sektoral dikumpulkan dan dikelola oleh Organisasi Perangkat Daerah (OPD) di lingkungan Pemerintah Kabupaten Bangka Barat bekerja sama dengan BPS sebagai Pembina Data.\n\n🌐 *Portal Satu Data Bangka Barat*:\nhttps://data.bangkabaratkab.go.id/\n\n🌐 *Portal Satu Data Indonesia*:\nhttps://data.go.id/\n\nUntuk permohonan data sektoral OPD lebih lanjut, silakan manfaatkan layanan konsultasi di Menu *6*."

	msg := whatsapp.BuildQuickReplyButtons(
		"🏛️ Statistik Sektoral",
		text,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "↩️ Kembali (99)"},
			{ID: "0", DisplayText: "🏠 Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// Submenu 1.5: Search Prompt (Resolves Stub U-01)
func (r *BotRouter) HandleLayananSub_SearchPrompt(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu1_Search)

	var sb strings.Builder
	sb.WriteString("🔍 *Cari Data Berdasarkan Kata Kunci*\n\n")
	sb.WriteString("Silakan ketik kata kunci data yang ingin Sahabat cari ✍️\n\n")
	sb.WriteString("Atau sentuh kata kunci populer di bawah ini untuk langsung mencari:\n")

	botNum := whatsapp.GetBotNumber()
	if botNum != "" {
		sb.WriteString(fmt.Sprintf("• *kemiskinan* 👉 https://wa.me/%s?text=kemiskinan\n", botNum))
		sb.WriteString(fmt.Sprintf("• *inflasi* 👉 https://wa.me/%s?text=inflasi\n", botNum))
		sb.WriteString(fmt.Sprintf("• *penduduk* 👉 https://wa.me/%s?text=penduduk\n", botNum))
		sb.WriteString(fmt.Sprintf("• *pdrb* 👉 https://wa.me/%s?text=pdrb\n\n", botNum))
	} else {
		sb.WriteString("- *kemiskinan*\n- *inflasi*\n- *penduduk*\n- *pdrb*\n\n")
	}

	msg := whatsapp.BuildQuickReplyButtons(
		"🔍 Cari Data",
		sb.String(),
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "↩️ Menu Sebelumnya (99)"},
			{ID: "0", DisplayText: "🏠 Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleSearchKeyword executes search across all 6 sheets
func (r *BotRouter) HandleSearchKeyword(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	query := strings.TrimSpace(p.Text)
	results, err := r.gsheet.SearchAll(query)
	if err != nil {
		return r.sender.SendTextMessage(ctx, p.JID, "❌ Maaf sobat data, terjadi kendala saat memproses pencarian. Silakan coba lagi nanti.")
	}

	formatted := gsheet.FormatSearchResults(query, results)

	msg := whatsapp.BuildQuickReplyButtons(
		"🔍 Hasil Pencarian Data",
		formatted,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "↩️ Cari Kata Kunci Lain"},
			{ID: "0", DisplayText: "🏠 Menu Utama (0)"},
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

	// Fetch dynamic tables from Sheet if available
	sheetCode := fmt.Sprintf("11%d", topicIdx)
	tableData, _ := r.gsheet.GetLayananData(sheetCode)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📖 *Detail Statistik: %s*\n\n", locale.T(fmt.Sprintf("menu11%dSosduk", topicIdx))))
	if cakupan != "" && cakupan != cakupKey {
		sb.WriteString(cakupan + "\n\n")
	}
	if link != "" && link != linkKey {
		sb.WriteString(fmt.Sprintf("🌐 *Kunjungi Website BPS*:\n%s\n\n", link))
	}
	if tableData != "" {
		sb.WriteString("📑 *Tabel Data Tersedia*:\n" + tableData + "\n")
	}

	msg := whatsapp.BuildQuickReplyButtons(
		"📂 Detail Statistik Sosduk",
		sb.String(),
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "↩️ Menu Sebelumnya (99)"},
			{ID: "0", DisplayText: "🏠 Menu Utama (0)"},
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

	sheetCode := fmt.Sprintf("12%d", topicIdx)
	tableData, _ := r.gsheet.GetLayananData(sheetCode)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("💰 *Detail Statistik: %s*\n\n", locale.T(fmt.Sprintf("menu12%dEkonomi", topicIdx))))
	if cakupan != "" && cakupan != cakupKey {
		sb.WriteString(cakupan + "\n\n")
	}
	if link != "" && link != linkKey {
		sb.WriteString(fmt.Sprintf("🌐 *Kunjungi Website BPS*:\n%s\n\n", link))
	}
	if tableData != "" {
		sb.WriteString("📑 *Tabel Data Tersedia*:\n" + tableData + "\n")
	}

	msg := whatsapp.BuildQuickReplyButtons(
		"💰 Detail Statistik Ekonomi",
		sb.String(),
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "↩️ Menu Sebelumnya (99)"},
			{ID: "0", DisplayText: "🏠 Menu Utama (0)"},
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

	sheetCode := fmt.Sprintf("13%d", topicIdx)
	tableData, _ := r.gsheet.GetLayananData(sheetCode)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🌾 *Detail Statistik: %s*\n\n", locale.T(fmt.Sprintf("menu13%dPertanian", topicIdx))))
	if cakupan != "" && cakupan != cakupKey {
		sb.WriteString(cakupan + "\n\n")
	}
	if link != "" && link != linkKey {
		sb.WriteString(fmt.Sprintf("🌐 *Kunjungi Website BPS*:\n%s\n\n", link))
	}
	if tableData != "" {
		sb.WriteString("📑 *Tabel Data Tersedia*:\n" + tableData + "\n")
	}

	msg := whatsapp.BuildQuickReplyButtons(
		"🌾 Detail Statistik Pertanian",
		sb.String(),
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "↩️ Menu Sebelumnya (99)"},
			{ID: "0", DisplayText: "🏠 Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}
