package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"bot-istimewa/internal/scraper"
	"bot-istimewa/internal/whatsapp"
	"go.mau.fi/whatsmeow/types"
)

// HandleMenu3_Produk renders Produk Statistik menu
func (r *BotRouter) HandleMenu3_Produk(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu3)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "Berita Resmi Statistik (BRS)", Description: "Rilis data terbaru BPS Bangka Barat"},
		{ID: "2", Title: "Publikasi Daerah", Description: "Buku & laporan statistik resmi BPS"},
		{ID: "3", Title: "Data Mikro (SILASTIK)", Description: "Akses data unit level via SILASTIK"},
	}

	introText := "*BRS & PUBLIKASI RESMI*\nBPS Kabupaten Bangka Barat\n\nSilakan pilih jenis produk statistik yang Sahabat butuhkan:\n━━━━━━━━━━━━━━━━━━"
	footerText := "━━━━━━━━━━━━━━━━━━\nBalas dengan *angka menu*\n• Ketik *0* untuk Menu Utama\n```Statistik Bermakna, Statistik Berdampak```"
	body := whatsapp.BuildBodyText(introText, items, footerText)

	msg := whatsapp.BuildInteractiveList(
		"Produk Statistik",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Produk Statistik",
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
		sb.WriteString("*BERITA RESMI STATISTIK (BRS)*\n")
		sb.WriteString("BPS Kabupaten Bangka Barat\n\n")
		sb.WriteString("Rilis resmi indikator strategis dan sosial-ekonomi terkini:\n")
		sb.WriteString("━━━━━━━━━━━━━━━━━━\n")

		// Ambil dari database Scraper lokal SQLite
		if r.scraper != nil {
			scrapedItems, fresh, err := r.scraper.LatestPressReleases(ctx, "id", 4)
			if err == nil && len(scrapedItems) > 0 {
				for i, item := range scrapedItems {
					cleanTitle := scraper.CleanHTML(item.Title)
					cleanTitle = strings.TrimSuffix(cleanTitle, ".")
					sb.WriteString(fmt.Sprintf("%d. *%s*\n", i+1, cleanTitle))
					if item.ReleaseDate != nil {
						sb.WriteString(fmt.Sprintf("   • Tanggal Rilis: %s\n", item.ReleaseDate.Format("02 Jan 2006")))
					}
					if item.Abstract != "" {
						summary := scraper.SummarizeText(item.Abstract, 240)
						if summary != "" {
							sb.WriteString(fmt.Sprintf("   • Ringkasan: %s\n", summary))
						}
					}
					sb.WriteString("\n")
				}
				if fresh.UpdatedAt.Year() > 2000 {
					sb.WriteString(fmt.Sprintf("Pembaruan: %s WIB\n\n", fresh.UpdatedAt.Format("02 Jan 2006 15:04")))
				}

				// Kirim gambar rilis teratas jika tersedia
				if len(scrapedItems) > 0 && scrapedItems[0].ImageURL != "" {
					_ = r.sender.SendTextMessage(ctx, p.JID, "Menyiapkan infografis rilis terbaru...")
					_ = r.sender.SendImageFromURL(ctx, p.JID, scrapedItems[0].ImageURL, fmt.Sprintf("*%s*", scrapedItems[0].Title))
				}
			}
		}

		sb.WriteString("━━━━━━━━━━━━━━━━━━\n")
		sb.WriteString("Untuk salinan digital lengkap PDF Berita Resmi Statistik, Sahabat dapat mengajukan di Menu *6* (Layanan Konsultasi PST).")
		content = sb.String()

	case 2:
		title = "Publikasi BPS"
		var sb strings.Builder
		sb.WriteString("*PUBLIKASI RESMI DAERAH*\n")
		sb.WriteString("BPS Kabupaten Bangka Barat\n\n")
		sb.WriteString("Koleksi buku dan laporan statistik komprehensif daerah:\n")
		sb.WriteString("━━━━━━━━━━━━━━━━━━\n")

		// Ambil dari database Scraper lokal SQLite
		if r.scraper != nil {
			scrapedPubs, fresh, err := r.scraper.LatestPublications(ctx, "id", 4)
			if err == nil && len(scrapedPubs) > 0 {
				for i, item := range scrapedPubs {
					cleanTitle := scraper.CleanHTML(item.Title)
					cleanTitle = strings.TrimSuffix(cleanTitle, ".")
					sb.WriteString(fmt.Sprintf("%d. *%s*\n", i+1, cleanTitle))
					if item.ReleaseDate != nil {
						sb.WriteString(fmt.Sprintf("   • Tanggal Terbit: %s\n", item.ReleaseDate.Format("02 Jan 2006")))
					}
					if item.Abstract != "" {
						summary := scraper.SummarizeText(item.Abstract, 240)
						if summary != "" {
							sb.WriteString(fmt.Sprintf("   • Cakupan Isi: %s\n", summary))
						}
					}
					sb.WriteString("\n")
				}
				if fresh.UpdatedAt.Year() > 2000 {
					sb.WriteString(fmt.Sprintf("Pembaruan: %s WIB\n\n", fresh.UpdatedAt.Format("02 Jan 2006 15:04")))
				}

				// Kirim gambar cover publikasi teratas jika tersedia
				if len(scrapedPubs) > 0 && (scrapedPubs[0].ImageURL != "" || scrapedPubs[0].ThumbnailURL != "") {
					coverURL := scrapedPubs[0].ImageURL
					if coverURL == "" {
						coverURL = scrapedPubs[0].ThumbnailURL
					}
					_ = r.sender.SendTextMessage(ctx, p.JID, "Menyiapkan sampul publikasi terbaru...")
					_ = r.sender.SendImageFromURL(ctx, p.JID, coverURL, fmt.Sprintf("*%s*", scrapedPubs[0].Title))
				}
			}
		}

		sb.WriteString("━━━━━━━━━━━━━━━━━━\n")
		sb.WriteString("Untuk salinan digital buku publikasi lengkap (PDF), Sahabat dapat menghubungi petugas di Menu *6* (Layanan Konsultasi PST).")
		content = sb.String()

	case 3:
		title = "Layanan Data Mikro (SILASTIK)"
		content = "*LAYANAN DATA MIKRO (SILASTIK)*\n" +
			"BPS Kabupaten Bangka Barat\n\n" +
			"Data mikro adalah data mentah tingkat individu atau rumah tangga hasil sensus dan survei yang telah di-anonymize untuk keperluan riset ilmiah dan kajian kebijakan.\n" +
			"━━━━━━━━━━━━━━━━━━\n" +
			"*Koleksi Data Mikro BPS*:\n" +
			"1. *Susenas*: Sosial ekonomi, pengeluaran & konsumsi\n" +
			"2. *Sakernas*: Ketenagakerjaan & pengangguran\n" +
			"3. *Sensus Pertanian (ST)*: Profil petani & struktur usaha tani\n" +
			"4. *Sensus Penduduk (SP)*: Parameter demografi & mobilitas\n\n" +
			"*Persyaratan Permohonan*:\n" +
			"1. Surat Pengantar Resmi Perguruan Tinggi / Lembaga\n" +
			"2. Proposal Riset / Penelitian\n" +
			"3. Identitas Pemohon (KTP / KTM)\n" +
			"4. Komitmen Kerahasiaan Data (UU No. 16/1997)\n\n" +
			"━━━━━━━━━━━━━━━━━━\n" +
			"Sahabat dapat berkonsultasi permohonan data mikro dengan petugas di Menu *6* (Layanan Konsultasi PST)."
	}

	msg := whatsapp.BuildQuickReplyButtons(
		title,
		content,
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "Menu Sebelumnya"},
			{ID: "0", DisplayText: "Menu Utama"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleMenu4_Video renders Video Literasi menu
func (r *BotRouter) HandleMenu4_Video(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu4)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "Statistik Sosial Kependudukan", Description: "Video edukasi data sosial & kependudukan"},
		{ID: "2", Title: "Statistik Pertanian", Description: "Video edukasi data sektor pertanian"},
		{ID: "3", Title: "Statistik Ekonomi", Description: "Video edukasi PDRB, inflasi & ekonomi"},
		{ID: "4", Title: "Sensus BPS", Description: "Video Sensus Penduduk & Sensus Pertanian"},
		{ID: "5", Title: "Statistik Sektoral", Description: "Video data lintas sektor OPD & Romantik"},
		{ID: "6", Title: "Statistik Khusus", Description: "Video topik statistik khusus & tematik"},
		{ID: "7", Title: "Metodologi Statistik", Description: "Video konsep & metodologi pengumpulan data"},
		{ID: "8", Title: "Tips & Trik Statistik", Description: "Video cara baca & manfaatkan data"},
	}

	introText := "*VIDEO EDUKASI LITERASI*\nBPS Kabupaten Bangka Barat\n\nYuk tonton video animasi edukasi BPS untuk memahami data statistik lebih mudah:\n━━━━━━━━━━━━━━━━━━"
	footerText := "━━━━━━━━━━━━━━━━━━\nBalas dengan *angka menu*\n• Ketik *0* untuk Menu Utama\n```Statistik Bermakna, Statistik Berdampak```"
	body := whatsapp.BuildBodyText(introText, items, footerText)

	msg := whatsapp.BuildInteractiveList(
		"Video Literasi Statistik",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Topik Video",
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
		text := "*VIDEO LITERASI: PERTANIAN*\nBPS Kabupaten Bangka Barat\n━━━━━━━━━━━━━━━━━━\n\nVideo edukasi sektor pertanian sedang dalam tahap produksi oleh tim Diseminasi BPS Bangka Barat.\n\n*Kanal Resmi YouTube BPS Bangka Barat*:\nhttps://www.youtube.com/channel/UCGAZVY1p74KlQbs1L1KhGtg\n━━━━━━━━━━━━━━━━━━"
		return r.sendVideoPlaceholder(ctx, p.JID, "Statistik Pertanian", text)
	case 3:
		// Sub-menu for Ekonomi (5 videos)
		return r.HandleVideoSub_Ekonomi(ctx, s, p)
	case 4:
		// Sensus - Belum ada video lokal
		text := "*VIDEO LITERASI: SENSUS BPS*\nBPS Kabupaten Bangka Barat\n━━━━━━━━━━━━━━━━━━\n\nVideo edukasi Sensus Penduduk dan Sensus Pertanian sedang dipersiapkan.\n\n*Kanal Resmi YouTube BPS Bangka Barat*:\nhttps://www.youtube.com/channel/UCGAZVY1p74KlQbs1L1KhGtg\n━━━━━━━━━━━━━━━━━━"
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
		text := "*VIDEO LITERASI: METODOLOGI*\nBPS Kabupaten Bangka Barat\n━━━━━━━━━━━━━━━━━━\n\nVideo konsep & metodologi pengumpulan data sedang dipersiapkan.\n\n*Kanal Resmi YouTube BPS Bangka Barat*:\nhttps://www.youtube.com/channel/UCGAZVY1p74KlQbs1L1KhGtg\n━━━━━━━━━━━━━━━━━━"
		return r.sendVideoPlaceholder(ctx, p.JID, "Metodologi Statistik", text)
	case 8:
		// Tips & Trik - Belum ada video lokal
		text := "*VIDEO LITERASI: TIPS & TRIK*\nBPS Kabupaten Bangka Barat\n━━━━━━━━━━━━━━━━━━\n\nVideo panduan tips & trik membaca data statistik segera hadir.\n\n*Kanal Resmi YouTube BPS Bangka Barat*:\nhttps://www.youtube.com/channel/UCGAZVY1p74KlQbs1L1KhGtg\n━━━━━━━━━━━━━━━━━━"
		return r.sendVideoPlaceholder(ctx, p.JID, "Tips & Trik Statistik", text)
	default:
		return r.HandleMenu4_Video(ctx, s, p)
	}
}

// HandleVideoSub_Sosduk renders sub-list for Sosial Kependudukan videos
func (r *BotRouter) HandleVideoSub_Sosduk(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu4_Sosduk)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "Garis Kemiskinan", Description: "Konsep penentuan batas garis kemiskinan"},
		{ID: "2", Title: "Indeks Pembangunan Manusia", Description: "Memahami indikator IPM BPS"},
		{ID: "3", Title: "Ketenagakerjaan", Description: "Definisi angkatan kerja & konsep kerja"},
		{ID: "4", Title: "Pengangguran Kota vs Desa", Description: "Mengapa pengangguran kota lebih tinggi"},
		{ID: "5", Title: "Hitung Angka Kemiskinan", Description: "Metode penghitungan angka kemiskinan"},
	}

	introText := "*Menu Utama* > *[4] Video Edukasi* > *Sosial Kependudukan*\n\nPilih video edukasi yang ingin Sahabat tonton langsung di WhatsApp:\n━━━━━━━━━━━━━━━━━━"
	footerText := "━━━━━━━━━━━━━━━━━━\nBalas dengan *angka menu*\n• Ketik *99* untuk Kembali\n• Ketik *0* untuk Menu Utama"
	body := whatsapp.BuildBodyText(introText, items, footerText)

	msg := whatsapp.BuildInteractiveList(
		"Video Sosial Kependudukan",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Judul Video",
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
		{ID: "1", Title: "Membaca PDRB Lapangan Usaha", Description: "Struktur ekonomi & sektor unggulan daerah"},
		{ID: "2", Title: "Memahami PDB per Kapita", Description: "Pendapatan rata-rata dan kesejahteraan"},
		{ID: "3", Title: "Memaknai Data Inflasi", Description: "Indeks Harga Konsumen & laju inflasi"},
		{ID: "4", Title: "Kualitas Hidup dengan IPM", Description: "Kaitan ekonomi dengan pembangunan manusia"},
		{ID: "5", Title: "Pertumbuhan Ekonomi", Description: "Konsep laju pertumbuhan ekonomi daerah"},
	}

	introText := "*Menu Utama* > *[4] Video Edukasi* > *Statistik Ekonomi*\n\nPilih video edukasi yang ingin Sahabat tonton langsung di WhatsApp:\n━━━━━━━━━━━━━━━━━━"
	footerText := "━━━━━━━━━━━━━━━━━━\nBalas dengan *angka menu*\n• Ketik *99* untuk Kembali\n• Ketik *0* untuk Menu Utama"
	body := whatsapp.BuildBodyText(introText, items, footerText)

	msg := whatsapp.BuildInteractiveList(
		"Video Statistik Ekonomi",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Judul Video",
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
	caption := fmt.Sprintf("*Video Edukasi: %s*\nBPS Kabupaten Bangka Barat\n\n%s\n\n━━━━━━━━━━━━━━━━━━\n```Statistik Bermakna, Statistik Berdampak```", title, desc)

	// Check if local compressed video exists
	if _, err := os.Stat(filePath); err == nil {
		_ = r.sender.SendTextMessage(ctx, jid, fmt.Sprintf("*Menyiapkan video '%s'...*\nMohon tunggu beberapa detik, video akan otomatis terkirim dan dapat diputar langsung di WhatsApp.", title))

		if err := r.sender.SendVideoFromFile(ctx, jid, filePath, caption); err != nil {
			log.Printf("[VIDEO] Gagal kirim file video: %v, fallback ke teks info", err)
			_ = r.sender.SendTextMessage(ctx, jid, caption)
		}
	} else {
		log.Printf("[VIDEO] File video %s tidak ditemukan di disk, kirim teks deskripsi", filePath)
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
		{ID: "1", Title: "Statistik Demografi dan Sosial", Description: "Kependudukan, kemiskinan, pendidikan & sosial"},
		{ID: "2", Title: "Statistik Ekonomi", Description: "PDRB, inflasi, perdagangan & keuangan"},
		{ID: "3", Title: "Statistik Lingkungan Hidup dan Multidomain", Description: "Pertanian, perikanan, energi, iklim & lingkungan"},
	}

	introText := "*GALERI INFOGRAFIS DAERAH*\nBPS Kabupaten Bangka Barat\n\nVisualisasi ringkas data statistik Bangka Barat berdasarkan subjek:\n━━━━━━━━━━━━━━━━━━"
	footerText := "━━━━━━━━━━━━━━━━━━\nBalas dengan *angka menu (1 - 3)*\n• Ketik *0* untuk Menu Utama\n```Statistik Bermakna, Statistik Berdampak```"
	body := whatsapp.BuildBodyText(introText, items, footerText)

	msg := whatsapp.BuildInteractiveList(
		"Infografik Statistik",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Subjek Infografik",
		[]whatsapp.ListSection{
			{Title: "Subjek Infografik", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

func (r *BotRouter) HandleInfografikDetail(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage, idx int) error {
	var subjectKeyword string
	var categoryTitle string
	switch idx {
	case 1:
		subjectKeyword = "Sosial"
		categoryTitle = "Statistik Demografi dan Sosial"
	case 2:
		subjectKeyword = "Ekonomi"
		categoryTitle = "Statistik Ekonomi"
	case 3:
		subjectKeyword = "Lingkungan"
		categoryTitle = "Statistik Lingkungan Hidup dan Multidomain"
	default:
		subjectKeyword = ""
		categoryTitle = "Semua Kategori"
	}

	// 1. Tarik infografik sesuai subjek dari DataProvider Scraper SQLite
	if r.scraper != nil {
		infoItems, fresh, err := r.scraper.LatestInfographicsBySubject(ctx, subjectKeyword, 2)
		if err == nil && len(infoItems) > 0 {
			_ = r.sender.SendTextMessage(ctx, p.JID, fmt.Sprintf("Menyiapkan infografik %s terbaru dari portal BPS Bangka Barat...", categoryTitle))
			for _, item := range infoItems {
				caption := fmt.Sprintf("*%s*\nSubjek: %s\nhttps://bangkabaratkab.bps.go.id/id/infographic\n%s", item.Title, item.Subject, formatFreshnessFooter(fresh))
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

	return r.sender.SendTextMessage(ctx, p.JID, fmt.Sprintf("Belum ada infografik terbaru untuk kategori %s.\nCek infografik lengkap di: https://bangkabaratkab.bps.go.id/id/infographic\n\nKetik *99* untuk kembali atau *0* ke menu utama.", categoryTitle))
}

// HandleMenu11_TentangBPS handles Tentang BPS (Resolves TD-01 & TD-06)
func (r *BotRouter) HandleMenu11_TentangBPS(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu11_About)

	videoPath := "storage/videos/menu11_TentangBPS.mp4"
	if _, err := os.Stat(videoPath); err == nil {
		_ = r.sender.SendTextMessage(ctx, p.JID, "*Menyiapkan video profil resmi BPS Bangka Barat...*\nMohon tunggu beberapa detik, video akan terkirim dan bisa langsung diputar di WhatsApp.")
		caption := "*Mengenal BPS Lebih Dekat*\nBadan Pusat Statistik Kabupaten Bangka Barat\n\nKomplek Perkantoran Pemerintah Daerah Terpadu Bangka Barat"
		if err := r.sender.SendVideoFromFile(ctx, p.JID, videoPath, caption); err != nil {
			log.Printf("[VIDEO] Gagal kirim video profil: %v", err)
		}
	}

	text := fmt.Sprintf("*TENTANG BPS BANGKA BARAT*\nBadan Pusat Statistik Kabupaten Bangka Barat\n\n"+
		"Kenal lebih dekat dengan Badan Pusat Statistik (BPS) Kabupaten Bangka Barat! Kami hadir sebagai penyedia data statistik berkualitas untuk kemajuan daerah.\n\n"+
		"━━━━━━━━━━━━━━━━━━\n"+
		"*Alamat Kantor*:\n"+
		"Komplek Perkantoran Pemerintah Daerah Terpadu\n"+
		"Kabupaten Bangka Barat, Kepulauan Bangka Belitung\n\n"+
		"*Jam Layanan PST*:\n"+
		"• Senin – Kamis: 08.00 – 15.30 WIB\n"+
		"• Jumat: 08.00 – 15.00 WIB\n\n"+
		"*Portal Resmi*:\n"+
		"https://bangkabaratkab.bps.go.id\n\n"+
		"*Video Profil*:\n"+
		"%s\n"+
		"━━━━━━━━━━━━━━━━━━\n\n"+
		"```Statistik Bermakna, Statistik Berdampak```", r.cfg.TentangBPSVideoURL)

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
	msg := fmt.Sprintf("Sumber: bangkabaratkab.bps.go.id · diperbarui %s WIB", f.UpdatedAt.Format("02 Jan 2006 15:04"))
	if f.Stale {
		msg += " *(data mungkin belum terbaru, cek langsung di portal)*"
	}
	return msg
}

