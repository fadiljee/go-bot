package scraper

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type indicatorCacheEntry struct {
	indicators []StrategicIndicator
	freshness  Freshness
	expiresAt  time.Time
}

var (
	indCacheMu sync.RWMutex
	indCache   = make(map[string]indicatorCacheEntry)
)

// InvalidateIndicatorCache clears the memory cache for indicators.
func InvalidateIndicatorCache() {
	indCacheMu.Lock()
	defer indCacheMu.Unlock()
	indCache = make(map[string]indicatorCacheEntry)
}

// GetAllStrategicIndicators returns all 7 indicators enriched with scraped documents.
func (p *Provider) GetAllStrategicIndicators(ctx context.Context) ([]StrategicIndicator, Freshness, error) {
	indCacheMu.RLock()
	cached, ok := indCache["all"]
	indCacheMu.RUnlock()
	if ok && time.Now().Before(cached.expiresAt) {
		return cached.indicators, cached.freshness, nil
	}

	res := make([]StrategicIndicator, 7)
	for i := 1; i <= 7; i++ {
		ind, _, err := p.GetStrategicIndicator(ctx, strconv.Itoa(i))
		if err != nil {
			return nil, Freshness{}, err
		}
		res[i-1] = *ind
	}

	freshness := p.calculateFreshness(ctx, nil)

	indCacheMu.Lock()
	indCache["all"] = indicatorCacheEntry{
		indicators: res,
		freshness:  freshness,
		expiresAt:  time.Now().Add(5 * time.Minute),
	}
	indCacheMu.Unlock()

	return res, freshness, nil
}

// GetStrategicIndicator returns a specific strategic indicator by ID (1-7) or slug.
func (p *Provider) GetStrategicIndicator(ctx context.Context, idOrSlug string) (*StrategicIndicator, Freshness, error) {
	key := strings.ToLower(strings.TrimSpace(idOrSlug))

	indCacheMu.RLock()
	cached, ok := indCache["ind:"+key]
	indCacheMu.RUnlock()
	if ok && time.Now().Before(cached.expiresAt) && len(cached.indicators) > 0 {
		return &cached.indicators[0], cached.freshness, nil
	}

	var ind *StrategicIndicator
	var err error

	switch key {
	case "1", "kependudukan", "penduduk":
		ind, err = p.buildKependudukan(ctx)
	case "2", "pertanian", "perkebunan":
		ind, err = p.buildPertanian(ctx)
	case "3", "kemiskinan", "miskin":
		ind, err = p.buildKemiskinan(ctx)
	case "4", "ketenagakerjaan", "tenagakerja", "pekerjaan":
		ind, err = p.buildKetenagakerjaan(ctx)
	case "5", "ekonomi", "pdrb", "pertumbuhan-ekonomi":
		ind, err = p.buildEkonomi(ctx)
	case "6", "inflasi", "harga", "ihk":
		ind, err = p.buildInflasi(ctx)
	case "7", "ipm", "kesejahteraan", "pembangunan-manusia":
		ind, err = p.buildIPM(ctx)
	default:
		return nil, Freshness{}, fmt.Errorf("indikator strategis tidak ditemukan untuk id/slug: %q (tersedia 1-7)", idOrSlug)
	}

	if err != nil {
		return nil, Freshness{}, err
	}

	freshness := p.calculateFreshness(ctx, ind.Items)

	indCacheMu.Lock()
	indCache["ind:"+key] = indicatorCacheEntry{
		indicators: []StrategicIndicator{*ind},
		freshness:  freshness,
		expiresAt:  time.Now().Add(5 * time.Minute),
	}
	indCacheMu.Unlock()

	return ind, freshness, nil
}

func (p *Provider) buildKependudukan(ctx context.Context) (*StrategicIndicator, error) {
	ind := &StrategicIndicator{
		ID:          1,
		Slug:        "kependudukan",
		Title:       "Indikator Kependudukan",
		Category:    "Kependudukan & Migrasi",
		Description: "Data statistik demografi penduduk Kabupaten Bangka Barat.",
		Source:      "BPS Kabupaten Bangka Barat",
		SourceURL:   "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=519",
		UpdatedAt:   "2026-03-02",
		Metrics: []IndicatorMetric{
			{Key: "total_penduduk", Label: "Jumlah Penduduk", Value: "± 214.450 Jiwa", Note: "Hasil Proyeksi BPS"},
			{Key: "laju_pertumbuhan", Label: "Laju Pertumbuhan Penduduk", Value: "1,48% per tahun"},
			{Key: "kepadatan_penduduk", Label: "Kepadatan Penduduk", Value: "76 jiwa/km²"},
			{Key: "sex_ratio", Label: "Rasio Jenis Kelamin (Sex Ratio)", Value: "106,8", Note: "107 laki-laki per 100 perempuan"},
		},
	}

	items, _, _ := p.SearchItems(ctx, "Indikator Strategis", "id", 2)
	moreItems, _, _ := p.SearchItems(ctx, "Dalam Angka", "id", 1)
	ind.Items = append(items, moreItems...)
	if len(ind.Items) > 0 && ind.Items[0].ReleaseDate != nil {
		ind.UpdatedAt = ind.Items[0].ReleaseDate.Format("2006-01-02")
		if ind.Items[0].FileURL != "" {
			ind.SourceURL = ind.Items[0].FileURL
		}
	}
	return ind, nil
}

func (p *Provider) buildPertanian(ctx context.Context) (*StrategicIndicator, error) {
	ind := &StrategicIndicator{
		ID:          2,
		Slug:        "pertanian",
		Title:       "Indikator Pertanian & Perkebunan",
		Category:    "Pertanian, Perkebunan & Perikanan",
		Description: "Data statistik komoditas perkebunan, pangan, dan perikanan Kabupaten Bangka Barat.",
		Source:      "BPS Kabupaten Bangka Barat (Sensus Pertanian & Statistik Perkebunan)",
		SourceURL:   "https://bangkabaratkab.bps.go.id",
		UpdatedAt:   "2026-10-12",
		Metrics: []IndicatorMetric{
			{Key: "kelapa_sawit", Label: "Kelapa Sawit", Value: "Luas areal > 60.000 Ha", Note: "Komoditas perkebunan utama"},
			{Key: "lada", Label: "Lada (Muntok White Pepper)", Value: "Komoditas Unggulan", Note: "Khas Bangka Barat dengan pasar ekspor"},
			{Key: "karet", Label: "Karet", Value: "Komoditas Penyokong", Note: "Perkebunan rakyat"},
			{Key: "perikanan_laut", Label: "Perikanan Tangkap Laut", Value: "Sentra Muntok, Jebus, Tempilang"},
		},
	}

	newsItems, _, _ := p.SearchItems(ctx, "SITASI", "id", 1)
	pubItems, _, _ := p.SearchItems(ctx, "Kabupaten Bangka Barat Dalam Angka", "id", 1)
	ind.Items = append(newsItems, pubItems...)
	return ind, nil
}

func (p *Provider) buildKemiskinan(ctx context.Context) (*StrategicIndicator, error) {
	ind := &StrategicIndicator{
		ID:          3,
		Slug:        "kemiskinan",
		Title:       "Indikator Kemiskinan",
		Category:    "Sosial & Kemiskinan",
		Description: "Tingkat dan indeks kemiskinan makro Kabupaten Bangka Barat bersumber dari data Susenas.",
		Source:      "BPS Kabupaten Bangka Barat (Data Susenas)",
		SourceURL:   "https://bangkabaratkab.bps.go.id",
		UpdatedAt:   "2025-12-30",
		Metrics: []IndicatorMetric{
			{Key: "p0", Label: "Persentase Penduduk Miskin (P0)", Value: "2,75%", Note: "Salah satu terendah di Kep. Bangka Belitung"},
			{Key: "garis_kemiskinan", Label: "Garis Kemiskinan", Value: "Rp 580.000,- /kapita/bulan"},
			{Key: "p1", Label: "Indeks Kedalaman Kemiskinan (P1)", Value: "0,32"},
			{Key: "p2", Label: "Indeks Keparahan Kemiskinan (P2)", Value: "0,06"},
		},
	}

	items, _, _ := p.SearchItems(ctx, "Kesejahteraan Rakyat", "id", 2)
	ind.Items = items
	if len(ind.Items) > 0 && ind.Items[0].ReleaseDate != nil {
		ind.UpdatedAt = ind.Items[0].ReleaseDate.Format("2006-01-02")
		if ind.Items[0].FileURL != "" {
			ind.SourceURL = ind.Items[0].FileURL
		}
	}
	return ind, nil
}

func (p *Provider) buildKetenagakerjaan(ctx context.Context) (*StrategicIndicator, error) {
	ind := &StrategicIndicator{
		ID:          4,
		Slug:        "ketenagakerjaan",
		Title:       "Indikator Ketenagakerjaan",
		Category:    "Tenaga Kerja",
		Description: "Data tingkat pengangguran dan partisipasi angkatan kerja bersumber dari Sakernas.",
		Source:      "BPS Kabupaten Bangka Barat (Data Sakernas)",
		SourceURL:   "https://bangkabaratkab.bps.go.id",
		UpdatedAt:   "2026-03-02",
		Metrics: []IndicatorMetric{
			{Key: "tpt", Label: "Tingkat Pengangguran Terbuka (TPT)", Value: "3,92%"},
			{Key: "tpak", Label: "Tingkat Partisipasi Angkatan Kerja (TPAK)", Value: "68,45%"},
			{Key: "lapangan_usaha", Label: "Sektor Usaha Utama", Value: "Pertanian/Perkebunan, Perdagangan, dan Industri Pengolahan"},
		},
	}

	items, _, _ := p.SearchItems(ctx, "Ketenagakerjaan", "id", 2)
	ind.Items = items
	return ind, nil
}

func (p *Provider) buildEkonomi(ctx context.Context) (*StrategicIndicator, error) {
	ind := &StrategicIndicator{
		ID:          5,
		Slug:        "ekonomi",
		Title:       "Indikator Pertumbuhan Ekonomi & PDRB",
		Category:    "Perekonomian Makro",
		Description: "Pertumbuhan ekonomi dan nilai Produk Domestik Regional Bruto Kabupaten Bangka Barat.",
		Source:      "BPS Kabupaten Bangka Barat (Publikasi PDRB)",
		SourceURL:   "https://bangkabaratkab.bps.go.id",
		UpdatedAt:   "2026-10-01",
		Metrics: []IndicatorMetric{
			{Key: "lpe", Label: "Laju Pertumbuhan Ekonomi (LPE)", Value: "3,35%", Note: "Triwulan II-2026 (y-on-y)"},
			{Key: "pdrb_adhb", Label: "PDRB Atas Dasar Harga Berlaku (ADHB)", Value: "± Rp 14,8 Triliun"},
			{Key: "pdrb_adhk", Label: "PDRB Atas Dasar Harga Konstan (ADHK)", Value: "± Rp 9,7 Triliun"},
			{Key: "struktur_ekonomi", Label: "Struktur Ekonomi Utama", Value: "Industri Pengolahan, Pertanian/Perkebunan, dan Pertambangan"},
		},
	}

	items, _, _ := p.SearchItems(ctx, "PDRB", "id", 2)
	news, _, _ := p.SearchItems(ctx, "Tumbuh 3,35", "id", 1)
	ind.Items = append(news, items...)
	return ind, nil
}

func (p *Provider) buildInflasi(ctx context.Context) (*StrategicIndicator, error) {
	ind := &StrategicIndicator{
		ID:          6,
		Slug:        "inflasi",
		Title:       "Indikator Harga & Inflasi",
		Category:    "Harga & Konsumsi",
		Description: "Perkembangan Indeks Harga Konsumen (IHK) dan tingkat inflasi terkini Kabupaten Bangka Barat.",
		Source:      "Berita Resmi Statistik (BRS) BPS Kabupaten Bangka Barat",
		SourceURL:   "https://bangkabaratkab.bps.go.id/id/pressrelease",
		UpdatedAt:   "2026-10-01",
		Metrics: []IndicatorMetric{
			{Key: "inflasi_yoy", Label: "Inflasi Year-on-Year (y-on-y)", Value: "4,49%", Note: "Periode Rilis Terkini"},
			{Key: "ihk", Label: "Indeks Harga Konsumen (IHK)", Value: "106,30"},
			{Key: "komoditas_pendorong", Label: "Komoditas Pendorong Inflasi", Value: "Kelompok Makanan, Minuman, Tembakau, dan Perawatan Pribadi"},
		},
	}

	// Dynamic check from latest press releases in scraper store
	items, _, err := p.LatestPressReleases(ctx, "id", 3)
	if err == nil && len(items) > 0 {
		ind.Items = items
		latestBRS := items[0]
		if latestBRS.ReleaseDate != nil {
			ind.UpdatedAt = latestBRS.ReleaseDate.Format("2006-01-02")
		}
		if latestBRS.FileURL != "" {
			ind.SourceURL = latestBRS.FileURL
		} else if latestBRS.URL != "" {
			ind.SourceURL = latestBRS.URL
		}
		if latestBRS.ImageURL != "" {
			ind.ImageURL = latestBRS.ImageURL
		}

		// Try dynamic extraction of rate from title if available, e.g. "sebesar 4,49 persen"
		reRate := regexp.MustCompile(`([0-9]+[,\.][0-9]+)\s*persen`)
		m := reRate.FindStringSubmatch(latestBRS.Title)
		if len(m) > 1 {
			rateVal := m[1] + "%"
			for i := range ind.Metrics {
				if ind.Metrics[i].Key == "inflasi_yoy" {
					ind.Metrics[i].Value = rateVal
					if latestBRS.ReleaseDate != nil {
						ind.Metrics[i].Note = latestBRS.ReleaseDate.Format("02 January 2006")
					}
				}
			}
		}
	}

	return ind, nil
}

func (p *Provider) buildIPM(ctx context.Context) (*StrategicIndicator, error) {
	ind := &StrategicIndicator{
		ID:          7,
		Slug:        "ipm",
		Title:       "Indikator Kesejahteraan & IPM",
		Category:    "Kesejahteraan & Pembangunan Manusia",
		Description: "Indeks Pembangunan Manusia (IPM) dan indikator umur harapan hidup serta pendidikan.",
		Source:      "Publikasi IPM BPS Kabupaten Bangka Barat",
		SourceURL:   "https://bangkabaratkab.bps.go.id",
		UpdatedAt:   "2026-06-30",
		Metrics: []IndicatorMetric{
			{Key: "ipm", Label: "Indeks Pembangunan Manusia (IPM)", Value: "71,58", Note: "Kategori Tinggi"},
			{Key: "uhh", Label: "Umur Harapan Hidup (UHH)", Value: "73,42 Tahun"},
			{Key: "hls", Label: "Harapan Lama Sekolah (HLS)", Value: "12,85 Tahun"},
			{Key: "rls", Label: "Rata-rata Lama Sekolah (RLS)", Value: "8,24 Tahun"},
			{Key: "pengeluaran_perkapita", Label: "Pengeluaran Riil Per Kapita", Value: "Rp 11.850.000,- /tahun"},
		},
	}

	items, _, _ := p.SearchItems(ctx, "Pembangunan Manusia", "id", 2)
	ind.Items = items
	if len(ind.Items) > 0 && ind.Items[0].ReleaseDate != nil {
		ind.UpdatedAt = ind.Items[0].ReleaseDate.Format("2006-01-02")
		if ind.Items[0].FileURL != "" {
			ind.SourceURL = ind.Items[0].FileURL
		}
	}
	return ind, nil
}
