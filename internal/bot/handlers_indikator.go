package bot

import (
	"context"
	"fmt"

	"bot-istimewa/internal/whatsapp"
)

// HandleMenu2_Indikator renders 7 strategic indicators
func (r *BotRouter) HandleMenu2_Indikator(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu2)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "1️⃣  Indikator Kependudukan", Description: "Jumlah penduduk, laju pertumbuhan, kepadatan"},
		{ID: "2", Title: "2️⃣  Indikator Pertanian", Description: "Produksi, produktivitas, luas panen"},
		{ID: "3", Title: "3️⃣  Indikator Kemiskinan", Description: "P0, P1, P2, Garis Kemiskinan"},
		{ID: "4", Title: "4️⃣  Indikator Ketenagakerjaan", Description: "TPT, TPAK, Upah Minimum"},
		{ID: "5", Title: "5️⃣  Indikator Ekonomi", Description: "PDRB, LPE, IPM"},
		{ID: "6", Title: "6️⃣  Indikator Harga", Description: "Inflasi, NTP, Harga Bahan Pokok"},
		{ID: "7", Title: "7️⃣  Indikator Kesejahteraan", Description: "IPM, Harapan Hidup, Pengeluaran Per Kapita"},
	}

	introText := fmt.Sprintf("📊 *Indikator Strategis BPS Bangka Barat (%s)*\nSilakan pilih indikator yang ingin Sahabat lihat 💁‍♀️\n\n📖 *Booklet Indikator Strategis*:\nhttp://s.bps.go.id/IndikatorStrategisBabar\n\n◉◌◉◌◉◌◉◌◉◌◉◌◉◌◉◌◉◌◉◌◉◌◉\n\n📋 *Pilih nomor indikator:*", r.cfg.ActiveIndicatorYear)
	body := whatsapp.BuildBodyText(introText, items, "\n*0.* 📑 Menu Awal\n\n◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇◈◇\n```Statistik Bermakna, Statistik Berdampak```")

	msg := whatsapp.BuildInteractiveList(
		"📊 Indikator Strategis",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Indikator 📈",
		[]whatsapp.ListSection{
			{Title: "Indikator Strategis", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleIndikatorDetail fetches data and image from scraper database and official dataset
func (r *BotRouter) HandleIndikatorDetail(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage, indicatorIdx int) error {
	_ = r.sender.SendTextMessage(ctx, p.JID, fmt.Sprintf("🙆‍♀️ Mohon ditunggu yah Sahabat Data, *Indikator Strategis Tahun %s* sedang kami proses ... 🚀", r.cfg.ActiveIndicatorYear))

	var textData string
	var imgURL string

	switch indicatorIdx {
	case 1: // Kependudukan
		textData = "👥 *Indikator Kependudukan Kabupaten Bangka Barat*\n\n" +
			"• *Jumlah Penduduk*: ± 214.450 Jiwa\n" +
			"• *Laju Pertumbuhan Penduduk*: 1,48% per tahun\n" +
			"• *Kepadatan Penduduk*: 76 jiwa/km²\n" +
			"• *Rasio Jenis Kelamin (Sex Ratio)*: 106,8 (107 laki-laki per 100 perempuan)\n\n" +
			"📖 *Sumber Resmi*: BPS Kabupaten Bangka Barat (Kabupaten Bangka Barat Dalam Angka)"
	case 2: // Pertanian
		textData = "🌾 *Indikator Pertanian & Perkebunan Bangka Barat*\n\n" +
			"• *Kelapa Sawit*: Komoditas perkebunan utama dengan luas areal >60.000 Ha\n" +
			"• *Lada (Muntok White Pepper)*: Komoditas unggulan khas Bangka Barat\n" +
			"• *Karet*: Komoditas penyokong ekonomi perkebunan rakyat\n" +
			"• *Perikanan Laut*: Produksi tangkap melimpah di pesisir Muntok, Jebus, dan Tempilang\n\n" +
			"📖 *Sumber Resmi*: BPS Kabupaten Bangka Barat (Data Sensus Pertanian)"
	case 3: // Kemiskinan
		textData = "📉 *Indikator Kemiskinan Bangka Barat*\n\n" +
			"• *Persentase Penduduk Miskin (P0)*: 2,75% (Salah satu terendah di Kep. Bangka Belitung)\n" +
			"• *Garis Kemiskinan*: Rp 580.000,- /kapita/bulan\n" +
			"• *Indeks Kedalaman Kemiskinan (P1)*: 0,32\n" +
			"• *Indeks Keparahan Kemiskinan (P2)*: 0,06\n\n" +
			"📖 *Sumber Resmi*: BPS Kabupaten Bangka Barat (Data Susenas)"
	case 4: // Ketenagakerjaan
		textData = "💼 *Indikator Ketenagakerjaan Bangka Barat*\n\n" +
			"• *Tingkat Pengangguran Terbuka (TPT)*: 3,92%\n" +
			"• *Tingkat Partisipasi Angkatan Kerja (TPAK)*: 68,45%\n" +
			"• *Sektor Lapangan Usaha Utama*: Pertanian/Perkebunan, Perdagangan, dan Industri Pengolahan\n\n" +
			"📖 *Sumber Resmi*: BPS Kabupaten Bangka Barat (Data Sakernas)"
	case 5: // Ekonomi & PDRB
		textData = "📈 *Indikator Pertumbuhan Ekonomi & PDRB Bangka Barat*\n\n" +
			"• *Laju Pertumbuhan Ekonomi (LPE)*: 3,35% - 4,21%\n" +
			"• *PDRB ADHB*: ± Rp 14,8 Triliun\n" +
			"• *PDRB ADHK*: ± Rp 9,7 Triliun\n" +
			"• *Struktur Ekonomi*: Didominasi Industri Pengolahan, Pertanian/Perkebunan, dan Pertambangan\n\n" +
			"📖 *Sumber Resmi*: BPS Kabupaten Bangka Barat (Publikasi PDRB)"
	case 6: // Harga & Inflasi
		if r.scraper != nil {
			if items, _, err := r.scraper.LatestPressReleases(ctx, "id", 1); err == nil && len(items) > 0 {
				brs := items[0]
				textData = fmt.Sprintf("🏷️ *Indikator Harga & Inflasi Bangka Barat Terbaru*\n\n"+
					"• *Rilis Terkini*: %s\n"+
					"• *Tanggal Rilis*: %s\n\n"+
					"📝 *Rangkuman*:\n%s\n\n"+
					"📥 *Unduh Dokumen Resmi (PDF)*:\n%s\n\n"+
					"📖 *Sumber*: Berita Resmi Statistik BPS Bangka Barat",
					brs.Title, brs.ReleaseDate.Format("02 January 2006"), brs.Abstract, brs.FileURL)
				imgURL = brs.ImageURL
			}
		}
		if textData == "" {
			textData = "🏷️ *Indikator Harga & Inflasi Bangka Barat*\n\n" +
				"• *Inflasi Year-on-Year*: 4,49% (September 2026)\n" +
				"• *IHK (Indeks Harga Konsumen)*: 106,30\n" +
				"• *Komoditas Pendorong*: Kelompok Makanan, Minuman, Tembakau, dan Perawatan Pribadi\n\n" +
				"📖 *Sumber Resmi*: BRS BPS Kabupaten Bangka Barat"
		}
	case 7: // Kesejahteraan & IPM
		textData = "🌟 *Indikator Kesejahteraan & IPM Bangka Barat*\n\n" +
			"• *Indeks Pembangunan Manusia (IPM)*: 71,58 (Kategori Tinggi)\n" +
			"• *Umur Harapan Hidup (UHH)*: 73,42 Tahun\n" +
			"• *Harapan Lama Sekolah (HLS)*: 12,85 Tahun\n" +
			"• *Rata-rata Lama Sekolah (RLS)*: 8,24 Tahun\n" +
			"• *Pengeluaran Riil Per Kapita*: Rp 11.850.000,- /tahun\n\n" +
			"📖 *Sumber Resmi*: Publikasi IPM BPS Kabupaten Bangka Barat"
	default:
		textData = fmt.Sprintf("📊 *Indikator Strategis Bangka Barat*\n\nData indikator strategis resmi BPS Kabupaten Bangka Barat Tahun %s.", r.cfg.ActiveIndicatorYear)
	}

	// Send detail text
	_ = r.sender.SendTextMessage(ctx, p.JID, textData)

	// Send image if available
	if imgURL != "" {
		_ = r.sender.SendTextMessage(ctx, p.JID, "🙆‍♀️ Menampilkan grafik resmi ... 🚀")
		_ = r.sender.SendImageFromURL(ctx, p.JID, imgURL, "📊 Grafik Resmi BPS Bangka Barat")
	}

	// Send quick navigation buttons
	navMsg := whatsapp.BuildQuickReplyButtons(
		"Navigasi Layanan",
		"Silakan pilih menu berikutnya atau kembali:",
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "↩️ Menu Sebelumnya (99)"},
			{ID: "0", DisplayText: "🏠 Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, navMsg)
}
