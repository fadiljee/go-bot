package bot

import (
	"context"
	"fmt"
	"strconv"

	"bot-istimewa/internal/locale"
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

// HandleIndikatorDetail fetches data and image from Google Sheet and displays
func (r *BotRouter) HandleIndikatorDetail(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage, indicatorIdx int) error {
	noStr := strconv.Itoa(indicatorIdx)
	_ = r.sender.SendTextMessage(ctx, p.JID, fmt.Sprintf("🙆‍♀️ Mohon ditunggu yah Sahabat Data, *Indikator Strategis Tahun %s* sedang kami proses ... 🚀", r.cfg.ActiveIndicatorYear))

	textData, imgURL, err := r.gsheet.GetIndikator(noStr, r.cfg.ActiveIndicatorYear)
	if err != nil || textData == "" {
		fallbackText := fmt.Sprintf("📝 *Indikator Strategis: %s*\n\nData sedang dalam pemutakhiran oleh BPS Bangka Barat.", locale.T(fmt.Sprintf("menu2%d", indicatorIdx)))
		textData = fallbackText
	}

	// Send detail text
	_ = r.sender.SendTextMessage(ctx, p.JID, textData)

	// Send infografik image if available
	if imgURL != "" {
		_ = r.sender.SendTextMessage(ctx, p.JID, "🙆‍♀️ Infografik sedang disiapkan ... 🚀")
		_ = r.sender.SendImageFromURL(ctx, p.JID, imgURL, "📊 Infografik Indikator Strategis")
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
