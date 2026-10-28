package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"bot-istimewa/internal/bpsapi"
	"bot-istimewa/internal/scraper"
	"bot-istimewa/internal/whatsapp"
)

// HandleMenu3_Produk renders Produk Statistik menu
func (r *BotRouter) HandleMenu3_Produk(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu3)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "1️⃣  Berita Rilis Statistik (BRS)", Description: "Rilis data terbaru BPS Bangka Barat"},
		{ID: "2", Title: "2️⃣  Publikasi", Description: "Buku & laporan statistik resmi BPS"},
		{ID: "3", Title: "3️⃣  Data Mikro", Description: "Akses data unit level via SILASTIK"},
	}

	introText := "📚 *Produk Statistik BPS*\nSilakan pilih jenis produk statistik yang Sahabat butuhkan 💁‍♀️\n\n●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●\n\n📋 *Pilih nomor produk:*"
	body := whatsapp.BuildBodyText(introText, items, "\n*0.* 📑 Menu Awal\n\n◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇\n```Statistik Bermakna, Statistik Berdampak```")

	msg := whatsapp.BuildInteractiveList(
		"📚 Produk Statistik",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Produk Statistik 📕",
		[]whatsapp.ListSection{
			{Title: "Jenis Produk Statistik", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

func (r *BotRouter) HandleProdukDetail(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage, idx int) error {
	var title, content string
	switch idx {
	case 1:
		title = "📰 Berita Resmi Statistik (BRS)"
		var sb strings.Builder
		sb.WriteString("📢 *Berita Resmi Statistik (BRS) — BPS Bangka Barat*\n\n")

		// Tarik data rilis BRS live dari Web API BPS Bangka Barat (domain: 1903)
		itemsFound := false
		if r.bps != nil {
			brsRes, err := r.bps.GetPressReleases(ctx, bpsapi.ListParams{
				Domain: "1903",
				Page:   1,
			})
			if err == nil && len(brsRes.Items) > 0 {
				itemsFound = true
				sb.WriteString("✨ *Daftar Rilis BRS Terbaru (Live Portal)*:\n\n")
				limit := 4
				if len(brsRes.Items) < limit {
					limit = len(brsRes.Items)
				}
				for i := 0; i < limit; i++ {
					item := brsRes.Items[i]
					sb.WriteString(fmt.Sprintf("%d. 📄 *%s*\n", i+1, item.Title))
					if item.RlDate != "" {
						sb.WriteString(fmt.Sprintf("   🗓️ Rilis: %s\n", item.RlDate))
					}
					if item.Pdf != "" {
						sb.WriteString(fmt.Sprintf("   📥 Unduh PDF: %s\n", item.Pdf))
					}
					sb.WriteString("\n")
				}
			}
		}

		// Fallback ke Scraper DataProvider (Priority 2) (§4.2)
		if !itemsFound && r.scraper != nil {
			scrapedItems, fresh, err := r.scraper.LatestPressReleases(ctx, "id", 4)
			if err == nil && len(scrapedItems) > 0 {
				itemsFound = true
				sb.WriteString("✨ *Daftar Rilis BRS Terbaru (Portal BPS)*:\n\n")
				for i, item := range scrapedItems {
					sb.WriteString(fmt.Sprintf("%d. 📄 *%s*\n", i+1, item.Title))
					if item.ReleaseDate != nil {
						sb.WriteString(fmt.Sprintf("   🗓️ Rilis: %s\n", item.ReleaseDate.Format("02 Jan 2006")))
					}
					if item.FileURL != "" {
						sb.WriteString(fmt.Sprintf("   📥 Unduh PDF: %s\n", item.FileURL))
					} else if item.URL != "" {
						sb.WriteString(fmt.Sprintf("   🔗 Baca: %s\n", item.URL))
					}
					sb.WriteString("\n")
				}
				sb.WriteString(formatFreshnessFooter(fresh) + "\n\n")
			}
		}

		sb.WriteString("🌐 *Portal BRS Bangka Barat*:\nhttps://bangkabaratkab.bps.go.id/id/pressrelease")
		content = sb.String()

	case 2:
		title = "📚 Publikasi BPS"
		var sb strings.Builder
		sb.WriteString("📖 *Publikasi Resmi BPS Kabupaten Bangka Barat*\n\n")

		// Tarik katalog buku publikasi live dari Web API BPS Bangka Barat (domain: 1903)
		pubFound := false
		if r.bps != nil {
			pubRes, err := r.bps.GetPublications(ctx, bpsapi.ListParams{
				Domain: "1903",
				Page:   1,
			})
			if err == nil && len(pubRes.Items) > 0 {
				pubFound = true
				sb.WriteString("✨ *Publikasi Terbaru (Live Portal)*:\n\n")
				limit := 4
				if len(pubRes.Items) < limit {
					limit = len(pubRes.Items)
				}
				for i := 0; i < limit; i++ {
					item := pubRes.Items[i]
					sb.WriteString(fmt.Sprintf("%d. 📕 *%s*\n", i+1, item.Title))
					if item.RlDate != "" {
						sb.WriteString(fmt.Sprintf("   🗓️ Rilis: %s\n", item.RlDate))
					}
					if item.Pdf != "" {
						sb.WriteString(fmt.Sprintf("   📥 Unduh: %s\n", item.Pdf))
					}
					sb.WriteString("\n")
				}
			}
		}

		// Fallback ke Scraper DataProvider (Priority 2) (§4.2)
		if !pubFound && r.scraper != nil {
			scrapedPubs, fresh, err := r.scraper.LatestPublications(ctx, "id", 4)
			if err == nil && len(scrapedPubs) > 0 {
				pubFound = true
				sb.WriteString("✨ *Publikasi Terbaru (Portal BPS)*:\n\n")
				for i, item := range scrapedPubs {
					sb.WriteString(fmt.Sprintf("%d. 📕 *%s*\n", i+1, item.Title))
					if item.ReleaseDate != nil {
						sb.WriteString(fmt.Sprintf("   🗓️ Rilis: %s\n", item.ReleaseDate.Format("02 Jan 2006")))
					}
					if item.FileURL != "" {
						sb.WriteString(fmt.Sprintf("   📥 Unduh: %s\n", item.FileURL))
					} else if item.URL != "" {
						sb.WriteString(fmt.Sprintf("   🔗 Lihat: %s\n", item.URL))
					}
					sb.WriteString("\n")
				}
				sb.WriteString(formatFreshnessFooter(fresh) + "\n\n")
			}
		}

		sb.WriteString("🌐 *Katalog Lengkap Publikasi*:\nhttps://bangkabaratkab.bps.go.id/id/publication")
		content = sb.String()

	case 3:
		title = "💾 Layanan Data Mikro"
		content = "🔬 *Akses Data Mikro BPS*\n\nBPS menyediakan data mikro (raw data) level individu/rumah tangga/usaha yang telah di-anonymize untuk keperluan penelitian akademis dan perumusan kebijakan.\n\n🌐 *Akses Portal SILASTIK BPS*:\nhttps://silastik.bps.go.id/\n🌐 *Website BPS Bangka Barat*:\nhttps://bangkabaratkab.bps.go.id/"
	}

	msg := whatsapp.BuildQuickReplyButtons(
		title,
		content,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "↩️ Menu Sebelumnya (99)"},
			{ID: "0", DisplayText: "🏠 Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleMenu4_Video renders Video Literasi menu
func (r *BotRouter) HandleMenu4_Video(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu4)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "1️⃣  Statistik Sosial Kependudukan", Description: "Video edukasi data sosial & kependudukan"},
		{ID: "2", Title: "2️⃣  Statistik Pertanian", Description: "Video edukasi data sektor pertanian"},
		{ID: "3", Title: "3️⃣  Statistik Ekonomi", Description: "Video edukasi PDRB, inflasi & ekonomi"},
		{ID: "4", Title: "4️⃣  Sensus", Description: "Video Sensus Penduduk & Sensus Pertanian"},
		{ID: "5", Title: "5️⃣  Statistik Sektoral", Description: "Video data lintas sektor OPD"},
		{ID: "6", Title: "6️⃣  Statistik Khusus", Description: "Video topik statistik khusus & tematik"},
		{ID: "7", Title: "7️⃣  Metodologi Statistik", Description: "Video konsep & metodologi pengumpulan data"},
		{ID: "8", Title: "8️⃣  Tips & Trik Statistik", Description: "Video cara baca & manfaatkan data"},
	}

	introText := "📽️ *Video Literasi Statistik*\nNonton video edukasi BPS untuk memahami data lebih mudah 🎬\n\n●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●\n\n📋 *Pilih topik video:*"
	body := whatsapp.BuildBodyText(introText, items, "\n*0.* 📑 Menu Awal\n\n◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇\n```Statistik Bermakna, Statistik Berdampak```")

	msg := whatsapp.BuildInteractiveList(
		"📽️ Video Literasi Statistik",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Topik Video 🎬",
		[]whatsapp.ListSection{
			{Title: "Topik Video", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

func (r *BotRouter) HandleVideoDetail(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage, idx int) error {
	noStr := strconv.Itoa(idx)
	videoData, err := r.gsheet.GetVideoStat(noStr)
	if err != nil || videoData == "" {
		videoData = "🎥 Video literasi untuk topik ini sedang disiapkan. Silakan kunjungi kanal YouTube resmi BPS Bangka Barat."
	}

	msg := whatsapp.BuildQuickReplyButtons(
		"🎬 Video Literasi",
		videoData,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "↩️ Menu Sebelumnya (99)"},
			{ID: "0", DisplayText: "🏠 Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleMenu5_Infografik renders Infografik menu
func (r *BotRouter) HandleMenu5_Infografik(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu5)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "1️⃣  Infografik Sosial Kependudukan", Description: "Visualisasi data kependudukan, kemiskinan, pendidikan"},
		{ID: "2", Title: "2️⃣  Infografik Pertanian", Description: "Visualisasi produksi & lahan pertanian"},
		{ID: "3", Title: "3️⃣  Infografik Ekonomi", Description: "Visualisasi PDRB, inflasi & ekonomi"},
		{ID: "4", Title: "4️⃣  Infografik Sensus", Description: "Visualisasi data hasil sensus BPS"},
	}

	introText := "🏞️ *Infografik Statistik*\nNikmati visualisasi ringkas data BPS Bangka Barat 📊\n\n●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●\n\n📋 *Pilih kategori infografik:*"
	body := whatsapp.BuildBodyText(introText, items, "\n*0.* 📑 Menu Awal\n\n◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇\n```Statistik Bermakna, Statistik Berdampak```")

	msg := whatsapp.BuildInteractiveList(
		"🏞️ Infografik Statistik",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Kategori Infografik 🖼️",
		[]whatsapp.ListSection{
			{Title: "Kategori Infografik", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

func (r *BotRouter) HandleInfografikDetail(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage, idx int) error {
	// 1. Coba tarik infografik live dari Web API BPS Bangka Barat (domain: 1903)
	if r.bps != nil {
		infoRes, err := r.bps.GetInfographics(ctx, bpsapi.ListParams{
			Domain: "1903",
			Page:   1,
		})
		if err == nil && len(infoRes.Items) > 0 {
			_ = r.sender.SendTextMessage(ctx, p.JID, "🖼️ Menyiapkan infografik terbaru langsung dari portal BPS Bangka Barat... 🚀")
			limit := 2
			if len(infoRes.Items) < limit {
				limit = len(infoRes.Items)
			}
			for i := 0; i < limit; i++ {
				item := infoRes.Items[i]
				caption := fmt.Sprintf("📊 *%s*\n🌐 https://bangkabaratkab.bps.go.id/id/infographic", item.Title)
				if item.Img != "" {
					_ = r.sender.SendImageFromURL(ctx, p.JID, item.Img, caption)
				}
			}

			navMsg := whatsapp.BuildQuickReplyButtons(
				"Navigasi Infografik",
				"Selesai menampilkan infografik dari portal BPS Bangka Barat. Silakan pilih menu berikutnya:",
				"PST BPS Kabupaten Bangka Barat",
				[]whatsapp.QuickReplyButton{
					{ID: "99", DisplayText: "↩️ Menu Sebelumnya (99)"},
					{ID: "0", DisplayText: "🏠 Menu Utama (0)"},
				},
			)
			return r.sender.SendInteractiveMessage(ctx, p.JID, navMsg)
		}
	}

	// 2. Fallback ke Scraper DataProvider (Priority 2) (§4.2)
	if r.scraper != nil {
		infoItems, fresh, err := r.scraper.LatestInfographics(ctx, 2)
		if err == nil && len(infoItems) > 0 {
			_ = r.sender.SendTextMessage(ctx, p.JID, "🖼️ Menyiapkan infografik terbaru dari portal BPS Bangka Barat... 🚀")
			for _, item := range infoItems {
				caption := fmt.Sprintf("📊 *%s*\n%s\n🌐 https://bangkabaratkab.bps.go.id/id/infographic", item.Title, formatFreshnessFooter(fresh))
				if item.ImageURL != "" {
					_ = r.sender.SendImageFromURL(ctx, p.JID, item.ImageURL, caption)
				}
			}

			navMsg := whatsapp.BuildQuickReplyButtons(
				"Navigasi Infografik",
				"Selesai menampilkan infografik dari portal BPS Bangka Barat. Silakan pilih menu berikutnya:",
				"PST BPS Kabupaten Bangka Barat",
				[]whatsapp.QuickReplyButton{
					{ID: "99", DisplayText: "↩️ Menu Sebelumnya (99)"},
					{ID: "0", DisplayText: "🏠 Menu Utama (0)"},
				},
			)
			return r.sender.SendInteractiveMessage(ctx, p.JID, navMsg)
		}
	}

	// 3. Fallback ke Google Sheets jika API BPS dan Scraper belum tersedia
	noStr := strconv.Itoa(idx)
	titles, urls, err := r.gsheet.GetInfografik(noStr)
	if err != nil || len(urls) == 0 {
		return r.sender.SendTextMessage(ctx, p.JID, "🖼️ Belum ada infografik terbaru untuk kategori ini.\n🌐 Cek infografik lengkap di: https://bangkabaratkab.bps.go.id/id/infographic\n\nKetik *99* untuk kembali atau *0* ke menu utama.")
	}

	_ = r.sender.SendTextMessage(ctx, p.JID, fmt.Sprintf("🖼️ Menampilkan %d infografik...", len(urls)))

	for i, u := range urls {
		caption := "📊 Infografik BPS Bangka Barat"
		if i < len(titles) && titles[i] != "" {
			caption = "📊 " + titles[i]
		}
		_ = r.sender.SendImageFromURL(ctx, p.JID, u, caption)
	}

	navMsg := whatsapp.BuildQuickReplyButtons(
		"Navigasi Infografik",
		"Selesai menampilkan infografik. Silakan pilih menu berikutnya:",
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "↩️ Menu Sebelumnya (99)"},
			{ID: "0", DisplayText: "🏠 Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, navMsg)
}

// HandleMenu11_TentangBPS handles Tentang BPS (Resolves TD-01 & TD-06)
func (r *BotRouter) HandleMenu11_TentangBPS(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu11_About)

	text := fmt.Sprintf("🏢 *Mengenal BPS Lebih Dekat*\n\nSetelah menjelajahi banyak data di menu sebelumnya, apakah Sahabat Data sudah mengenal lebih dekat Badan Pusat Statistik (BPS)? 😊\n\nSeperti kata pepatah _'Tak Kenal Maka Tak Sayang'_. Yuk kita mengenal BPS kita tercinta lebih dalam.\n\n🎥 *Tonton Video Profil BPS*:\n%s\n\n📍 *Alamat Kantor*:\nKomplek Perkantoran Pemerintah Daerah Terpadu\nKabupaten Bangka Barat, Kepulauan Bangka Belitung", r.cfg.TentangBPSVideoURL)

	msg := whatsapp.BuildQuickReplyButtons(
		"🏢 Profil BPS Bangka Barat",
		text,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "0", DisplayText: "🏠 Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

func formatFreshnessFooter(f scraper.Freshness) string {
	msg := fmt.Sprintf("ℹ️ Sumber: bangkabaratkab.bps.go.id · diperbarui %s WIB", f.UpdatedAt.Format("02 Jan 2006 15:04"))
	if f.Stale {
		msg += " *(data mungkin belum terbaru, cek langsung di portal)*"
	}
	return msg
}

