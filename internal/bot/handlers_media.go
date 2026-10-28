package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"bot-istimewa/internal/bpsapi"
	"bot-istimewa/internal/scraper"
	"bot-istimewa/internal/whatsapp"
	"go.mau.fi/whatsmeow/types"
)

// HandleMenu3_Produk renders Produk Statistik menu
func (r *BotRouter) HandleMenu3_Produk(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu3)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "1. Berita Rilis Statistik (BRS)", Description: "Rilis data terbaru BPS Bangka Barat"},
		{ID: "2", Title: "2. Publikasi", Description: "Buku & laporan statistik resmi BPS"},
		{ID: "3", Title: "3. Data Mikro", Description: "Akses data unit level via SILASTIK"},
	}

	introText := "*Produk Statistik BPS*\nSilakan pilih jenis produk statistik yang Sahabat butuhkan \n\n●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●\n\n*Pilih nomor produk:*"
	body := whatsapp.BuildBodyText(introText, items, "\n*0.* Menu Awal\n\n◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇\n```Statistik Bermakna, Statistik Berdampak```")

	msg := whatsapp.BuildInteractiveList(
		"Produk Statistik",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Produk Statistik ",
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
		title = "Berita Resmi Statistik (BRS)"
		var sb strings.Builder
		sb.WriteString("*Berita Resmi Statistik (BRS) — BPS Bangka Barat*\n\n")

		// Tarik data rilis BRS live dari Web API BPS Bangka Barat (domain: 1903)
		itemsFound := false
		if r.bps != nil {
			brsRes, err := r.bps.GetPressReleases(ctx, bpsapi.ListParams{
				Domain: "1903",
				Page:   1,
			})
			if err == nil && len(brsRes.Items) > 0 {
				itemsFound = true
				sb.WriteString("*Daftar Rilis BRS Terbaru (Live Portal)*:\n\n")
				limit := 4
				if len(brsRes.Items) < limit {
					limit = len(brsRes.Items)
				}
				for i := 0; i < limit; i++ {
					item := brsRes.Items[i]
					sb.WriteString(fmt.Sprintf("%d. *%s*\n", i+1, item.Title))
					if item.RlDate != "" {
						sb.WriteString(fmt.Sprintf("   Rilis: %s\n", item.RlDate))
					}
					if item.Pdf != "" {
						sb.WriteString(fmt.Sprintf("   Unduh PDF: %s\n", item.Pdf))
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
				sb.WriteString("*Daftar Rilis BRS Terbaru (Portal BPS)*:\n\n")
				for i, item := range scrapedItems {
					sb.WriteString(fmt.Sprintf("%d. *%s*\n", i+1, item.Title))
					if item.ReleaseDate != nil {
						sb.WriteString(fmt.Sprintf("   Rilis: %s\n", item.ReleaseDate.Format("02 Jan 2006")))
					}
					if item.FileURL != "" {
						sb.WriteString(fmt.Sprintf("   Unduh PDF: %s\n", item.FileURL))
					} else if item.URL != "" {
						sb.WriteString(fmt.Sprintf("   Baca: %s\n", item.URL))
					}
					sb.WriteString("\n")
				}
				sb.WriteString(formatFreshnessFooter(fresh) + "\n\n")

				// Kirim gambar rilis teratas jika tersedia
				if len(scrapedItems) > 0 && scrapedItems[0].ImageURL != "" {
					_ = r.sender.SendTextMessage(ctx, p.JID, "Menyiapkan infografik rilis terbaru ... ")
					_ = r.sender.SendImageFromURL(ctx, p.JID, scrapedItems[0].ImageURL, fmt.Sprintf("*%s*", scrapedItems[0].Title))
				}
			}
		}

		sb.WriteString("*Portal BRS Bangka Barat*:\nhttps://bangkabaratkab.bps.go.id/id/pressrelease")
		content = sb.String()

	case 2:
		title = "Publikasi BPS"
		var sb strings.Builder
		sb.WriteString("*Publikasi Resmi BPS Kabupaten Bangka Barat*\n\n")

		// Tarik katalog buku publikasi live dari Web API BPS Bangka Barat (domain: 1903)
		pubFound := false
		if r.bps != nil {
			pubRes, err := r.bps.GetPublications(ctx, bpsapi.ListParams{
				Domain: "1903",
				Page:   1,
			})
			if err == nil && len(pubRes.Items) > 0 {
				pubFound = true
				sb.WriteString("*Publikasi Terbaru (Live Portal)*:\n\n")
				limit := 4
				if len(pubRes.Items) < limit {
					limit = len(pubRes.Items)
				}
				for i := 0; i < limit; i++ {
					item := pubRes.Items[i]
					sb.WriteString(fmt.Sprintf("%d. *%s*\n", i+1, item.Title))
					if item.RlDate != "" {
						sb.WriteString(fmt.Sprintf("   Rilis: %s\n", item.RlDate))
					}
					if item.Pdf != "" {
						sb.WriteString(fmt.Sprintf("   Unduh: %s\n", item.Pdf))
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
				sb.WriteString("*Publikasi Terbaru (Portal BPS)*:\n\n")
				for i, item := range scrapedPubs {
					sb.WriteString(fmt.Sprintf("%d. *%s*\n", i+1, item.Title))
					if item.ReleaseDate != nil {
						sb.WriteString(fmt.Sprintf("   Rilis: %s\n", item.ReleaseDate.Format("02 Jan 2006")))
					}
					if item.FileURL != "" {
						sb.WriteString(fmt.Sprintf("   Unduh: %s\n", item.FileURL))
					} else if item.URL != "" {
						sb.WriteString(fmt.Sprintf("   Lihat: %s\n", item.URL))
					}
					sb.WriteString("\n")
				}
				sb.WriteString(formatFreshnessFooter(fresh) + "\n\n")

				// Kirim gambar cover publikasi teratas jika tersedia
				if len(scrapedPubs) > 0 && scrapedPubs[0].ImageURL != "" {
					_ = r.sender.SendTextMessage(ctx, p.JID, "Menyiapkan sampul buku publikasi terbaru ... ")
					_ = r.sender.SendImageFromURL(ctx, p.JID, scrapedPubs[0].ImageURL, fmt.Sprintf("*%s*", scrapedPubs[0].Title))
				}
			}
		}

		sb.WriteString("*Katalog Lengkap Publikasi*:\nhttps://bangkabaratkab.bps.go.id/id/publication")
		content = sb.String()

	case 3:
		title = "Layanan Data Mikro"
		content = "*Akses Data Mikro BPS*\n\nBPS menyediakan data mikro (raw data) level individu/rumah tangga/usaha yang telah di-anonymize untuk keperluan penelitian akademis dan perumusan kebijakan.\n\n*Akses Portal SILASTIK BPS*:\nhttps://silastik.bps.go.id/\n*Website BPS Bangka Barat*:\nhttps://bangkabaratkab.bps.go.id/"
	}

	msg := whatsapp.BuildQuickReplyButtons(
		title,
		content,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "Menu Sebelumnya (99)"},
			{ID: "0", DisplayText: "Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleMenu4_Video renders Video Literasi menu
func (r *BotRouter) HandleMenu4_Video(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu4)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "1. Statistik Sosial Kependudukan", Description: "Video edukasi data sosial & kependudukan"},
		{ID: "2", Title: "2. Statistik Pertanian", Description: "Video edukasi data sektor pertanian"},
		{ID: "3", Title: "3. Statistik Ekonomi", Description: "Video edukasi PDRB, inflasi & ekonomi"},
		{ID: "4", Title: "4. Sensus", Description: "Video Sensus Penduduk & Sensus Pertanian"},
		{ID: "5", Title: "5. Statistik Sektoral", Description: "Video data lintas sektor OPD"},
		{ID: "6", Title: "6. Statistik Khusus", Description: "Video topik statistik khusus & tematik"},
		{ID: "7", Title: "7. Metodologi Statistik", Description: "Video konsep & metodologi pengumpulan data"},
		{ID: "8", Title: "8. Tips & Trik Statistik", Description: "Video cara baca & manfaatkan data"},
	}

	introText := "*Video Literasi Statistik*\nNonton video edukasi BPS untuk memahami data lebih mudah \n\n●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●\n\n*Pilih topik video:*"
	body := whatsapp.BuildBodyText(introText, items, "\n*0.* Menu Awal\n\n◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇\n```Statistik Bermakna, Statistik Berdampak```")

	msg := whatsapp.BuildInteractiveList(
		"Video Literasi Statistik",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Topik Video ",
		[]whatsapp.ListSection{
			{Title: "Topik Video", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

func (r *BotRouter) HandleVideoDetail(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage, idx int) error {
	switch idx {
	case 1:
		// Sub-menu for Sosial Kependudukan (5 videos)
		return r.HandleVideoSub_Sosduk(ctx, s, p)
	case 2:
		// Pertanian - Belum ada video lokal
		text := "*Video Literasi: Statistik Pertanian*\n\nVideo edukasi sektor pertanian sedang dalam tahap penyusunan oleh tim Diseminasi BPS Bangka Barat.\n\nTonton video rilis lainnya di kanal resmi YouTube:\nhttps://www.youtube.com/channel/UCGAZVY1p74KlQbs1L1KhGtg"
		return r.sendVideoPlaceholder(ctx, p.JID, "Statistik Pertanian", text)
	case 3:
		// Sub-menu for Ekonomi (5 videos)
		return r.HandleVideoSub_Ekonomi(ctx, s, p)
	case 4:
		// Sensus - Belum ada video lokal
		text := "*Video Literasi: Sensus BPS*\n\nVideo edukasi Sensus Penduduk dan Sensus Pertanian sedang dipersiapkan.\n\nTonton video resmi BPS di kanal YouTube:\nhttps://www.youtube.com/channel/UCGAZVY1p74KlQbs1L1KhGtg"
		return r.sendVideoPlaceholder(ctx, p.JID, "Sensus BPS", text)
	case 5:
		// Statistik Sektoral (1 video)
		return r.sendVideoResponse(ctx, p.JID,
			"storage/videos/Statistik Sektoral - Statistik Sektoral dan Aplikasi Romantik Badan Pusat Statistik.mp4",
			"Statistik Sektoral & Aplikasi Romantik",
			"Video panduan penyelenggaraan statistik sektoral bagi OPD serta pengajuan rekomendasi kegiatan statistik lewat portal Romantik BPS.",
		)
	case 6:
		// Statistik Khusus (1 video)
		return r.sendVideoResponse(ctx, p.JID,
			"storage/videos/Statistik Khusus - Kenalan dengan Statistik Khusus.mp4",
			"Kenalan dengan Statistik Khusus",
			"Video edukasi mengenai definisi, pemanfaatan, dan tata kelola statistik khusus di lingkungan BPS.",
		)
	case 7:
		// Metodologi - Belum ada video lokal
		text := "*Video Literasi: Metodologi Statistik*\n\nVideo konsep & metodologi pengumpulan data sedang dipersiapkan.\n\nKunjungi kanal YouTube BPS Bangka Barat:\nhttps://www.youtube.com/channel/UCGAZVY1p74KlQbs1L1KhGtg"
		return r.sendVideoPlaceholder(ctx, p.JID, "Metodologi Statistik", text)
	case 8:
		// Tips & Trik - Belum ada video lokal
		text := "*Video Literasi: Tips & Trik Statistik*\n\nVideo panduan tips & trik membaca data statistik segera hadir.\n\nKunjungi kanal YouTube BPS Bangka Barat:\nhttps://www.youtube.com/channel/UCGAZVY1p74KlQbs1L1KhGtg"
		return r.sendVideoPlaceholder(ctx, p.JID, "Tips & Trik Statistik", text)
	default:
		return r.HandleMenu4_Video(ctx, s, p)
	}
}

// HandleVideoSub_Sosduk renders sub-list for Sosial Kependudukan videos
func (r *BotRouter) HandleVideoSub_Sosduk(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu4_Sosduk)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "1. Garis Kemiskinan", Description: "Konsep penentuan batas garis kemiskinan"},
		{ID: "2", Title: "2. Indeks Pembangunan Manusia", Description: "Memahami indikator IPM BPS"},
		{ID: "3", Title: "3. Ketenagakerjaan", Description: "Definisi angkatan kerja & konsep kerja"},
		{ID: "4", Title: "4. Pengangguran Kota vs Desa", Description: "Mengapa pengangguran kota lebih tinggi"},
		{ID: "5", Title: "5. Hitung Angka Kemiskinan", Description: "Metode penghitungan angka kemiskinan"},
	}

	introText := "*Video Literasi: Statistik Sosial Kependudukan*\nPilih video edukasi yang ingin Sahabat tonton langsung di WhatsApp :\n\n●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●"
	body := whatsapp.BuildBodyText(introText, items, "\n*99.* Menu Sebelumnya\n*0.* Menu Awal")

	msg := whatsapp.BuildInteractiveList(
		"Video Sosial Kependudukan",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Judul Video ",
		[]whatsapp.ListSection{
			{Title: "Daftar Video Sosduk", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

func (r *BotRouter) HandleVideoDetail_Sosduk(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage, idx int) error {
	var filePath, title, desc string
	switch idx {
	case 1:
		filePath = "storage/videos/Sosduk - Garis Kemiskinan.mp4"
		title = "Garis Kemiskinan"
		desc = "Video edukasi resmi BPS Bangka Barat tentang penentuan garis kemiskinan."
	case 2:
		filePath = "storage/videos/Sosduk - Indeks Pembangungn Manusia (IPM).mp4"
		title = "Indeks Pembangunan Manusia (IPM)"
		desc = "Video edukasi resmi BPS Bangka Barat tentang komponen IPM (kesehatan, pendidikan, pengeluaran)."
	case 3:
		filePath = "storage/videos/Sosduk - Ketenagakerjaan.mp4"
		title = "Ketenagakerjaan"
		desc = "Video edukasi resmi BPS Bangka Barat tentang ketenagakerjaan dan angkatan kerja."
	case 4:
		filePath = "storage/videos/Sosduk - Kok Bisa Pengangguran Perkotaan Lebih Tinggi dari Perdesaan.mp4"
		title = "Pengangguran Perkotaan vs Perdesaan"
		desc = "Video edukasi fenomena dan penyebab tingkat pengangguran perkotaan lebih tinggi dari perdesaan."
	case 5:
		filePath = "storage/videos/Sosduk_Memahami_Penghitungan_Angka_Kemiskinan.mp4"
		title = "Memahami Penghitungan Angka Kemiskinan"
		desc = "Video edukasi metodologi dan survei Susenas dalam menghitung angka kemiskinan."
	default:
		return r.HandleVideoSub_Sosduk(ctx, s, p)
	}

	return r.sendVideoResponse(ctx, p.JID, filePath, title, desc)
}

// HandleVideoSub_Ekonomi renders sub-list for Ekonomi videos
func (r *BotRouter) HandleVideoSub_Ekonomi(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu4_Ekonomi)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "1. Membaca PDRB Lapangan Usaha", Description: "Struktur ekonomi & sektor unggulan daerah"},
		{ID: "2", Title: "2. Memahami PDB per Kapita", Description: "Pendapatan rata-rata dan kesejahteraan"},
		{ID: "3", Title: "3. Memaknai Data Inflasi", Description: "Indeks Harga Konsumen & laju inflasi"},
		{ID: "4", Title: "4. Kualitas Hidup dengan IPM", Description: "Kaitan ekonomi dengan pembangunan manusia"},
		{ID: "5", Title: "5. Pertumbuhan Ekonomi", Description: "Konsep laju pertumbuhan ekonomi daerah"},
	}

	introText := "*Video Literasi: Statistik Ekonomi*\nPilih video edukasi yang ingin Sahabat tonton langsung di WhatsApp :\n\n●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●"
	body := whatsapp.BuildBodyText(introText, items, "\n*99.* Menu Sebelumnya\n*0.* Menu Awal")

	msg := whatsapp.BuildInteractiveList(
		"Video Statistik Ekonomi",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Judul Video ",
		[]whatsapp.ListSection{
			{Title: "Daftar Video Ekonomi", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

func (r *BotRouter) HandleVideoDetail_Ekonomi(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage, idx int) error {
	var filePath, title, desc string
	switch idx {
	case 1:
		filePath = "storage/videos/Ekonomi - Cara Membaca PDRB Menurut Lapangan Usaha.mp4"
		title = "Membaca PDRB Menurut Lapangan Usaha"
		desc = "Video edukasi resmi BPS Bangka Barat cara membaca kontribusi sektor ekonomi terhadap PDRB."
	case 2:
		filePath = "storage/videos/Ekonomi - Memahami PDB per Kapita.mp4"
		title = "Memahami PDB per Kapita"
		desc = "Video edukasi resmi BPS Bangka Barat tentang PDB/PDRB per kapita dan maknanya."
	case 3:
		filePath = "storage/videos/Ekonomi - Memaknai Data Inflasi.mp4"
		title = "Memaknai Data Inflasi"
		desc = "Video edukasi resmi BPS Bangka Barat cara memahami inflasi dan dampaknya ke masyarakat."
	case 4:
		filePath = "storage/videos/Ekonomi - Mengukur Kualitas Hidup dengan Indeks Pembangunan Manusia.mp4"
		title = "Mengukur Kualitas Hidup dengan IPM"
		desc = "Video edukasi hubungan pertumbuhan ekonomi dan kualitas hidup manusia."
	case 5:
		filePath = "storage/videos/vid4wwhatappweb Ekonomi - Pertumbuhan Ekonomi.mp4"
		title = "Pertumbuhan Ekonomi"
		desc = "Video edukasi resmi BPS tentang perhitungan dan makna pertumbuhan ekonomi."
	default:
		return r.HandleVideoSub_Ekonomi(ctx, s, p)
	}

	return r.sendVideoResponse(ctx, p.JID, filePath, title, desc)
}

func (r *BotRouter) sendVideoResponse(ctx context.Context, jid types.JID, filePath, title, desc string) error {
	caption := fmt.Sprintf("*Video Literasi: %s*\n\n%s\n\n◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇\n```Badan Pusat Statistik Kabupaten Bangka Barat```", title, desc)

	// Check if local compressed video exists
	if _, err := os.Stat(filePath); err == nil {
		_ = r.sender.SendTextMessage(ctx, jid, fmt.Sprintf("*Menyiapkan video '%s'...*\nMohon tunggu beberapa detik, video akan otomatis terkirim dan dapat diputar langsung di WhatsApp ", title))

		if err := r.sender.SendVideoFromFile(ctx, jid, filePath, caption); err != nil {
			log.Printf("[VIDEO] Gagal kirim file video: %v, fallback ke teks info", err)
			_ = r.sender.SendTextMessage(ctx, jid, caption)
		}
	} else {
		log.Printf("ℹ[VIDEO] File video %s tidak ditemukan di disk, kirim teks deskripsi", filePath)
		fallbackText := fmt.Sprintf("%s\n\n*Tonton di Kanal YouTube BPS Bangka Barat*:\nhttps://www.youtube.com/channel/UCGAZVY1p74KlQbs1L1KhGtg", caption)
		_ = r.sender.SendTextMessage(ctx, jid, fallbackText)
	}

	// Follow with interactive navigation buttons
	navMsg := whatsapp.BuildQuickReplyButtons(
		"Navigasi Video",
		"Selesai memuat video edukasi. Silakan pilih menu berikutnya:",
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "Menu Sebelumnya (99)"},
			{ID: "0", DisplayText: "Menu Utama (0)"},
		},
	)
	return r.sender.SendInteractiveMessage(ctx, jid, navMsg)
}

func (r *BotRouter) sendVideoPlaceholder(ctx context.Context, jid types.JID, title, text string) error {
	msg := whatsapp.BuildQuickReplyButtons(
		title,
		text,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "Menu Sebelumnya (99)"},
			{ID: "0", DisplayText: "Menu Utama (0)"},
		},
	)
	return r.sender.SendInteractiveMessage(ctx, jid, msg)
}

// HandleMenu5_Infografik renders Infografik menu
func (r *BotRouter) HandleMenu5_Infografik(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu5)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "1. Infografik Sosial Kependudukan", Description: "Visualisasi data kependudukan, kemiskinan, pendidikan"},
		{ID: "2", Title: "2. Infografik Pertanian", Description: "Visualisasi produksi & lahan pertanian"},
		{ID: "3", Title: "3. Infografik Ekonomi", Description: "Visualisasi PDRB, inflasi & ekonomi"},
		{ID: "4", Title: "4. Infografik Sensus", Description: "Visualisasi data hasil sensus BPS"},
	}

	introText := "*Infografik Statistik*\nNikmati visualisasi ringkas data BPS Bangka Barat \n\n●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●◌●\n\n*Pilih kategori infografik:*"
	body := whatsapp.BuildBodyText(introText, items, "\n*0.* Menu Awal\n\n◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇\n```Statistik Bermakna, Statistik Berdampak```")

	msg := whatsapp.BuildInteractiveList(
		"Infografik Statistik",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Kategori Infografik ",
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
			_ = r.sender.SendTextMessage(ctx, p.JID, "Menyiapkan infografik terbaru langsung dari portal BPS Bangka Barat... ")
			limit := 2
			if len(infoRes.Items) < limit {
				limit = len(infoRes.Items)
			}
			for i := 0; i < limit; i++ {
				item := infoRes.Items[i]
				caption := fmt.Sprintf("*%s*\nhttps://bangkabaratkab.bps.go.id/id/infographic", item.Title)
				if item.Img != "" {
					_ = r.sender.SendImageFromURL(ctx, p.JID, item.Img, caption)
				}
			}

			navMsg := whatsapp.BuildQuickReplyButtons(
				"Navigasi Infografik",
				"Selesai menampilkan infografik dari portal BPS Bangka Barat. Silakan pilih menu berikutnya:",
				"PST BPS Kabupaten Bangka Barat",
				[]whatsapp.QuickReplyButton{
					{ID: "99", DisplayText: "Menu Sebelumnya (99)"},
					{ID: "0", DisplayText: "Menu Utama (0)"},
				},
			)
			return r.sender.SendInteractiveMessage(ctx, p.JID, navMsg)
		}
	}

	// 2. Fallback ke Scraper DataProvider (Priority 2) (§4.2)
	if r.scraper != nil {
		infoItems, fresh, err := r.scraper.LatestInfographics(ctx, 2)
		if err == nil && len(infoItems) > 0 {
			_ = r.sender.SendTextMessage(ctx, p.JID, "Menyiapkan infografik terbaru dari portal BPS Bangka Barat... ")
			for _, item := range infoItems {
				caption := fmt.Sprintf("*%s*\n%s\nhttps://bangkabaratkab.bps.go.id/id/infographic", item.Title, formatFreshnessFooter(fresh))
				if item.ImageURL != "" {
					_ = r.sender.SendImageFromURL(ctx, p.JID, item.ImageURL, caption)
				}
			}

			navMsg := whatsapp.BuildQuickReplyButtons(
				"Navigasi Infografik",
				"Selesai menampilkan infografik dari portal BPS Bangka Barat. Silakan pilih menu berikutnya:",
				"PST BPS Kabupaten Bangka Barat",
				[]whatsapp.QuickReplyButton{
					{ID: "99", DisplayText: "Menu Sebelumnya (99)"},
					{ID: "0", DisplayText: "Menu Utama (0)"},
				},
			)
			return r.sender.SendInteractiveMessage(ctx, p.JID, navMsg)
		}
	}

	// 3. Fallback ke rilis grafis visual dari database scraper
	if r.scraper != nil {
		items, fresh, err := r.scraper.LatestPressReleases(ctx, "id", 2)
		if err == nil && len(items) > 0 {
			_ = r.sender.SendTextMessage(ctx, p.JID, fmt.Sprintf("Menampilkan %d infografik rilis data terbaru BPS Bangka Barat ... ", len(items)))
			for _, item := range items {
				if item.ImageURL != "" {
					caption := fmt.Sprintf("*%s*\n%s\nhttps://bangkabaratkab.bps.go.id/id/infographic", item.Title, formatFreshnessFooter(fresh))
					_ = r.sender.SendImageFromURL(ctx, p.JID, item.ImageURL, caption)
				}
			}

			navMsg := whatsapp.BuildQuickReplyButtons(
				"Navigasi Infografik",
				"Selesai menampilkan infografik visual dari portal BPS Bangka Barat. Silakan pilih menu berikutnya:",
				"PST BPS Kabupaten Bangka Barat",
				[]whatsapp.QuickReplyButton{
					{ID: "99", DisplayText: "Menu Sebelumnya (99)"},
					{ID: "0", DisplayText: "Menu Utama (0)"},
				},
			)
			return r.sender.SendInteractiveMessage(ctx, p.JID, navMsg)
		}
	}

	return r.sender.SendTextMessage(ctx, p.JID, "Belum ada infografik terbaru untuk kategori ini.\nCek infografik lengkap di: https://bangkabaratkab.bps.go.id/id/infographic\n\nKetik *99* untuk kembali atau *0* ke menu utama.")
}

// HandleMenu11_TentangBPS handles Tentang BPS (Resolves TD-01 & TD-06)
func (r *BotRouter) HandleMenu11_TentangBPS(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu11_About)

	videoPath := "storage/videos/menu11_TentangBPS.mp4"
	if _, err := os.Stat(videoPath); err == nil {
		_ = r.sender.SendTextMessage(ctx, p.JID, "*Menyiapkan video profil resmi BPS Kabupaten Bangka Barat...* \nMohon tunggu beberapa detik, video akan terkirim dan bisa langsung diputar di WhatsApp.")
		caption := "*Mengenal BPS Lebih Dekat*\n\nProfil Resmi Badan Pusat Statistik Kabupaten Bangka Barat.\n\nKomplek Perkantoran Pemda Terpadu Bangka Barat"
		if err := r.sender.SendVideoFromFile(ctx, p.JID, videoPath, caption); err != nil {
			log.Printf("[VIDEO] Gagal kirim video profil: %v", err)
		}
	}

	text := fmt.Sprintf("*Mengenal BPS Lebih Dekat*\n\nSetelah menjelajahi banyak data di menu sebelumnya, apakah Sahabat Data sudah mengenal lebih dekat Badan Pusat Statistik (BPS)? \n\nSeperti kata pepatah _'Tak Kenal Maka Tak Sayang'_. Yuk kita mengenal BPS kita tercinta lebih dalam.\n\n*Tonton Video Profil BPS*:\n%s\n\n*Alamat Kantor*:\nKomplek Perkantoran Pemerintah Daerah Terpadu\nKabupaten Bangka Barat, Kepulauan Bangka Belitung", r.cfg.TentangBPSVideoURL)

	msg := whatsapp.BuildQuickReplyButtons(
		"Profil BPS Bangka Barat",
		text,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "0", DisplayText: "Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

func formatFreshnessFooter(f scraper.Freshness) string {
	msg := fmt.Sprintf("ℹSumber: bangkabaratkab.bps.go.id · diperbarui %s WIB", f.UpdatedAt.Format("02 Jan 2006 15:04"))
	if f.Stale {
		msg += " *(data mungkin belum terbaru, cek langsung di portal)*"
	}
	return msg
}

