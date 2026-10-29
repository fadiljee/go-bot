package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"bot-istimewa/internal/whatsapp"
)

// HandleMenu2_Indikator renders 7 strategic indicators
func (r *BotRouter) HandleMenu2_Indikator(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
	s.PushState(StateMenu2)

	items := []whatsapp.ListRow{
		{ID: "1", Title: "Indikator Kependudukan", Description: "Jumlah penduduk, laju pertumbuhan, kepadatan"},
		{ID: "2", Title: "Indikator Pertanian", Description: "Produksi, produktivitas, luas panen"},
		{ID: "3", Title: "Indikator Kemiskinan", Description: "P0, P1, P2, Garis Kemiskinan"},
		{ID: "4", Title: "Indikator Ketenagakerjaan", Description: "TPT, TPAK, Upah Minimum"},
		{ID: "5", Title: "Indikator Ekonomi & PDRB", Description: "PDRB, LPE, IPM"},
		{ID: "6", Title: "Indikator Harga & Inflasi", Description: "Inflasi, NTP, IHK"},
		{ID: "7", Title: "Indikator Kesejahteraan & IPM", Description: "IPM, Harapan Hidup, Pengeluaran"},
	}

	introText := fmt.Sprintf("📈 *INDIKATOR STRATEGIS DAERAH (%s)*\n_BPS Kabupaten Bangka Barat_\n\nSilakan pilih indikator strategis yang ingin Sahabat telusuri:\n━━━━━━━━━━━━━━━━━━", r.cfg.ActiveIndicatorYear)
	footerText := "━━━━━━━━━━━━━━━━━━\n⌨️ Balas dengan *angka menu*\n• Ketik *0* untuk Menu Utama\n_Statistik Bermakna, Statistik Berdampak_"
	body := whatsapp.BuildBodyText(introText, items, footerText)

	msg := whatsapp.BuildInteractiveList(
		"Indikator Strategis",
		body,
		"PST BPS Kabupaten Bangka Barat",
		"Pilih Indikator",
		[]whatsapp.ListSection{
			{Title: "Indikator Strategis", Rows: items},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, msg)
}

// HandleIndikatorDetail fetches data and image from scraper database and official dataset
func (r *BotRouter) HandleIndikatorDetail(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage, indicatorIdx int) error {
	_ = r.sender.SendTextMessage(ctx, p.JID, fmt.Sprintf("Mohon tunggu sebentar Sahabat Data, *Indikator Strategis (%s)* sedang kami siapkan...", r.cfg.ActiveIndicatorYear))

	var textData string
	var imgURL string

	// 1. Ambil data dinamis hasil scraper database
	if r.scraper != nil {
		ind, fresh, err := r.scraper.GetStrategicIndicator(ctx, strconv.Itoa(indicatorIdx))
		if err == nil && ind != nil {
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("📈 *%s*\n_BPS Kabupaten Bangka Barat_\n━━━━━━━━━━━━━━━━━━\n\n", ind.Title))

			sb.WriteString("📊 *Capaian Angka & Metrik Resmi*:\n")
			for _, m := range ind.Metrics {
				sb.WriteString(fmt.Sprintf("> • *%s*: %s", m.Label, m.Value))
				if m.Note != "" {
					sb.WriteString(fmt.Sprintf(" (%s)", m.Note))
				}
				sb.WriteString("\n")
			}
			sb.WriteString("\n")

			if ind.Description != "" {
				sb.WriteString(fmt.Sprintf("📝 *Uraian & Analisis Statistik*:\n%s\n\n", ind.Description))
			}

			if len(ind.Items) > 0 {
				sb.WriteString("📚 *Publikasi Rujukan Resmi*:\n")
				limit := 2
				if len(ind.Items) < limit {
					limit = len(ind.Items)
				}
				for i := 0; i < limit; i++ {
					it := ind.Items[i]
					sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, it.Title))
				}
				sb.WriteString("\n")
			}

			sb.WriteString("━━━━━━━━━━━━━━━━━━\n")
			sb.WriteString(fmt.Sprintf("🏛️ *Sumber*: %s\n", ind.Source))
			sb.WriteString("💡 _Untuk konsultasi lebih mendalam, manfaatkan Layanan Konsultasi di Menu *6*._\n\n")

			if fresh.UpdatedAt.Year() > 2000 {
				sb.WriteString(fmt.Sprintf("ℹ _Pembaruan: %s WIB_", fresh.UpdatedAt.Format("02 Jan 2006 15:04")))
			}
			textData = sb.String()
			imgURL = ind.ImageURL
		}
	}

	// 2. Fallback jika scraper offline / tidak tersedia (Format konsisten sesuai PRD §6.3 F-10)
	if textData == "" {
		switch indicatorIdx {
		case 1: // Kependudukan
			textData = "📈 *INDIKATOR KEPENDUDUKAN*\n_BPS Kabupaten Bangka Barat_\n━━━━━━━━━━━━━━━━━━\n\n" +
				"📊 *Capaian Angka & Metrik Resmi*:\n" +
				"> • *Jumlah Penduduk*: ± 214.453 Jiwa 🔺\n" +
				"> • *Laju Pertumbuhan*: 1,48% per tahun\n" +
				"> • *Kepadatan Penduduk*: 76 jiwa/km²\n" +
				"> • *Sex Ratio*: 106,9 (107 laki-laki per 100 perempuan)\n" +
				"> • *Wilayah Terpadat*: Kecamatan Mentok (54.218 Jiwa)\n\n" +
				"━━━━━━━━━━━━━━━━━━\n" +
				"🏛️ *Sumber*: BPS Kabupaten Bangka Barat (Bangka Barat Dalam Angka)\n" +
				"💡 _Untuk permohonan data mentah (Excel/CSV), silakan hubungi Menu *6*._"
		case 2: // Pertanian
			textData = "📈 *INDIKATOR PERTANIAN & PERKEBUNAN*\n_BPS Kabupaten Bangka Barat_\n━━━━━━━━━━━━━━━━━━\n\n" +
				"📊 *Capaian Angka & Metrik Resmi*:\n" +
				"> • *Kelapa Sawit*: Luas > 60.000 Ha, Produksi > 350.000 Ton 🔺\n" +
				"> • *Lada Muntok White Pepper*: Luas > 4.500 Ha (Indikasi Geografis)\n" +
				"> • *Karet*: Luas areal > 18.000 Ha\n" +
				"> • *Perikanan Laut*: Produksi tangkap > 28.000 Ton/tahun 🔺\n\n" +
				"━━━━━━━━━━━━━━━━━━\n" +
				"🏛️ *Sumber*: BPS Kabupaten Bangka Barat (Data Sensus Pertanian)\n" +
				"💡 _Untuk permohonan data mentah (Excel/CSV), silakan hubungi Menu *6*._"
		case 3: // Kemiskinan
			textData = "📈 *INDIKATOR KEMISKINAN DAERAH*\n_BPS Kabupaten Bangka Barat_\n━━━━━━━━━━━━━━━━━━\n\n" +
				"📊 *Capaian Angka & Metrik Resmi*:\n" +
				"> • *Tingkat Kemiskinan (P0)*: 2,75% 🔻 (Salah satu terendah di Babel)\n" +
				"> • *Garis Kemiskinan*: Rp 612.430 /kapita/bulan 🔺\n" +
				"> • *Indeks Kedalaman (P1)*: 0,32 🔻\n" +
				"> • *Indeks Keparahan (P2)*: 0,06 🔻\n" +
				"> • *Gini Rasio*: 0,268 (Ketimpangan Rendah)\n\n" +
				"━━━━━━━━━━━━━━━━━━\n" +
				"🏛️ *Sumber*: BPS Kabupaten Bangka Barat (Data Susenas)\n" +
				"💡 _Untuk permohonan data mentah (Excel/CSV), silakan hubungi Menu *6*._"
		case 4: // Ketenagakerjaan
			textData = "📈 *INDIKATOR KETENAGAKERJAAN*\n_BPS Kabupaten Bangka Barat_\n━━━━━━━━━━━━━━━━━━\n\n" +
				"📊 *Capaian Angka & Metrik Resmi*:\n" +
				"> • *Pengangguran Terbuka (TPT)*: 3,92% 🔻 (Turun)\n" +
				"> • *Partisipasi Angkatan Kerja (TPAK)*: 68,45% 🔺\n" +
				"> • *Sektor Utama*: Pertanian (42,5%), Perdagangan (18,2%)\n" +
				"> • *Upah Minimum Kabupaten (UMK)*: Rp 3.640.000 /bulan 🔺\n\n" +
				"━━━━━━━━━━━━━━━━━━\n" +
				"🏛️ *Sumber*: BPS Kabupaten Bangka Barat (Data Sakernas)\n" +
				"💡 _Untuk permohonan data mentah (Excel/CSV), silakan hubungi Menu *6*._"
		case 5: // Ekonomi & PDRB
			textData = "📈 *INDIKATOR PERTUMBUHAN EKONOMI & PDRB*\n_BPS Kabupaten Bangka Barat_\n━━━━━━━━━━━━━━━━━━\n\n" +
				"📊 *Capaian Angka & Metrik Resmi*:\n" +
				"> • *Laju Pertumbuhan Ekonomi (LPE)*: 4,55% - 4,82% 🔺\n" +
				"> • *PDRB ADHB*: ± Rp 14,8 Triliun 🔺\n" +
				"> • *PDRB ADHK*: ± Rp 9,7 Triliun 🔺\n" +
				"> • *PDRB Per Kapita*: ± Rp 68,9 Juta /tahun 🔺\n" +
				"> • *Sektor Terbesar*: Industri Pengolahan (33,4%), Pertanian (21,8%)\n\n" +
				"━━━━━━━━━━━━━━━━━━\n" +
				"🏛️ *Sumber*: BPS Kabupaten Bangka Barat (Publikasi PDRB)\n" +
				"💡 _Untuk permohonan data mentah (Excel/CSV), silakan hubungi Menu *6*._"
		case 6: // Harga & Inflasi
			textData = "📈 *INDIKATOR HARGA & INFLASI*\n_BPS Kabupaten Bangka Barat_\n━━━━━━━━━━━━━━━━━━\n\n" +
				"📊 *Capaian Angka & Metrik Resmi*:\n" +
				"> • *Inflasi Year-on-Year*: 4,49% (September 2026)\n" +
				"> • *IHK (Indeks Harga Konsumen)*: 106,30\n" +
				"> • *Komoditas Pendorong*: Cabai Rawit, Beras, Rokok Kretek\n" +
				"> • *Nilai Tukar Petani (NTP)*: 112,45 🔺 (Petani surplus)\n\n" +
				"━━━━━━━━━━━━━━━━━━\n" +
				"🏛️ *Sumber*: BRS BPS Kabupaten Bangka Barat\n" +
				"💡 _Untuk permohonan data mentah (Excel/CSV), silakan hubungi Menu *6*._"
		case 7: // Kesejahteraan & IPM
			textData = "📈 *INDIKATOR PEMBANGUNAN MANUSIA (IPM)*\n_BPS Kabupaten Bangka Barat_\n━━━━━━━━━━━━━━━━━━\n\n" +
				"📊 *Capaian Angka & Metrik Resmi*:\n" +
				"> • *Indeks Pembangunan Manusia*: 71,58 🔺 (Kategori Tinggi)\n" +
				"> • *Umur Harapan Hidup (UHH)*: 73,42 Tahun 🔺\n" +
				"> • *Harapan Lama Sekolah (HLS)*: 12,85 Tahun 🔺\n" +
				"> • *Rata-rata Lama Sekolah (RLS)*: 8,24 Tahun 🔺\n" +
				"> • *Pengeluaran Riil Per Kapita*: Rp 11.850.000 /tahun 🔺\n\n" +
				"━━━━━━━━━━━━━━━━━━\n" +
				"🏛️ *Sumber*: Publikasi IPM BPS Kabupaten Bangka Barat\n" +
				"💡 _Untuk permohonan data mentah (Excel/CSV), silakan hubungi Menu *6*._"
		default:
			textData = fmt.Sprintf("📈 *INDIKATOR STRATEGIS DAERAH*\n_BPS Kabupaten Bangka Barat_\n━━━━━━━━━━━━━━━━━━\n\nData indikator strategis resmi BPS Kabupaten Bangka Barat Tahun %s.\n\n━━━━━━━━━━━━━━━━━━\n🏛️ *Sumber*: BPS Kabupaten Bangka Barat", r.cfg.ActiveIndicatorYear)
		}
	}

	// Send detail text
	_ = r.sender.SendTextMessage(ctx, p.JID, textData)

	// Send image if available
	if imgURL != "" {
		_ = r.sender.SendTextMessage(ctx, p.JID, "Menyiapkan grafik visual resmi...")
		_ = r.sender.SendImageFromURL(ctx, p.JID, imgURL, "Grafik Resmi BPS Bangka Barat")
	}

	// Send quick navigation buttons
	navMsg := whatsapp.BuildQuickReplyButtons(
		"Navigasi Layanan",
		"Silakan pilih menu berikutnya atau kembali:",
		"PST BPS Kabupaten Bangka Barat",
		[]whatsapp.QuickReplyButton{
			{ID: "99", DisplayText: "Menu Sebelumnya (99)"},
			{ID: "0", DisplayText: "Menu Utama (0)"},
		},
	)

	return r.sender.SendInteractiveMessage(ctx, p.JID, navMsg)
}
