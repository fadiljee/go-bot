package scraper

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type layananCacheEntry struct {
	categories []DataCategory
	freshness  Freshness
	expiresAt  time.Time
}

var (
	layananCacheMu sync.RWMutex
	layananCache   = make(map[string]layananCacheEntry)
)

// InvalidateLayananDataCache clears the in-memory cache for Layanan Data.
func InvalidateLayananDataCache() {
	layananCacheMu.Lock()
	defer layananCacheMu.Unlock()
	layananCache = make(map[string]layananCacheEntry)
}

type topicDef struct {
	id         int
	slug       string
	title      string
	keyword    string
	desc       string
	subjectURL string
}

var sosdukTopicDefs = []topicDef{
	{1, "kependudukan", "Statistik Kependudukan dan Migrasi", "penduduk", "Statistik kependudukan, demografi, struktur dan pertumbuhan populasi, migrasi, dan keluarga.", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=519"},
	{2, "kemiskinan", "Statistik Kondisi Tempat Tinggal, Kemiskinan & Sosial Lintas Sektor", "kemiskinan", "Statistik garis kemiskinan, jumlah penduduk miskin (P0), indeks kedalaman (P1), dan keparahan (P2).", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=563"},
	{3, "tenaga-kerja", "Statistik Tenaga Kerja", "tenaga kerja", "Tingkat Pengangguran Terbuka (TPT), Tingkat Partisipasi Angkatan Kerja (TPAK), dan lapangan usaha utama.", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=520"},
	{4, "pendidikan", "Statistik Pendidikan", "pendidikan", "Angka Partisipasi Sekolah (APS), Angka Partisipasi Murni (APM/APK), dan Angka Melek Huruf.", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=521"},
	{5, "kesehatan", "Statistik Kesehatan", "kesehatan", "Fasilitas kesehatan, angka kesakitan (morbiditas), jaminan kesehatan, dan tenaga medis.", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=522"},
	{6, "konsumsi", "Statistik Konsumsi dan Pendapatan", "konsumsi", "Rata-rata pengeluaran dan konsumsi kalori/protein per kapita penduduk.", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=523"},
	{7, "perumahan", "Statistik Pemukiman dan Perumahan", "tempat tinggal", "Kondisi perumahan, sumber air minum layak, sanitasi, dan fasilitas tempat tinggal.", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=524"},
	{8, "budaya", "Statistik Budaya", "budaya", "Kegiatan kebudayaan, cagar budaya, dan sarana rekreasi/sosial.", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=525"},
	{9, "politik", "Statistik Aktivitas Politik dan Komunitas", "politik", "Partisipasi pemilu, organisasi kemasyarakatan, dan keterwakilan politik.", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=526"},
	{10, "regional", "Statistik Regional dan Statistik Area Kecil", "regional", "Statistik kewilayahan, luas wilayah kecamatan/desa, dan data spasial statistik.", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=527"},
	{11, "gender", "Statistik Gender dan Kelompok Populasi Khusus", "gender", "Indeks Pembangunan Gender (IPG), Indeks Ketimpangan Gender (IKG), anak, lansia, dan disabilitas.", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=528"},
	{12, "informasi", "Statistik Masyarakat Informasi", "informasi", "Akses teknologi informasi, kepemilikan gawai, dan pemanfaatan internet.", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=529"},
}

var ekonomiTopicDefs = []topicDef{
	{1, "pdrb", "Neraca Ekonomi & PDRB", "pdrb", "Produk Domestik Regional Bruto (PDRB) ADHB & ADHK menurut lapangan usaha dan pengeluaran.", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=530"},
	{2, "industri", "Statistik Bisnis & Industri", "industri", "Industri mikro, kecil, sedang, dan besar di Kabupaten Bangka Barat.", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=531"},
	{3, "keuangan", "Keuangan Pemerintah & Sektor Publik", "keuangan", "Realisasi APBD, pendapatan asli daerah (PAD), dan belanja pemerintah daerah.", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=532"},
	{4, "harga", "Harga-Harga & Indeks Konsumen", "harga", "Indeks Harga Konsumen (IHK), inflasi bulanan/tahunan, dan pemantauan harga bahan pokok.", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=533"},
	{5, "upah", "Biaya Tenaga Kerja & Upah", "upah", "Upah minimum kabupaten (UMK) dan rata-rata upah pekerja.", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=534"},
	{6, "energi", "Statistik Energi & Kelistrikan", "listrik", "Daya terpasang, produksi listrik, dan konsumsi energi daerah.", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=535"},
	{7, "pertambangan", "Pertambangan, Manufaktur & Konstruksi", "tambang", "Produksi timah, pasir kuarsa, industri pengolahan, dan Indeks Kemahalan Konstruksi (IKK).", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=536"},
	{8, "transportasi", "Transportasi & Pergudangan", "transportasi", "Arus lalu lintas pelabuhan Tanjung Kalian Muntok, angkutan jalan, dan kendaraan bermotor.", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=537"},
	{9, "pariwisata", "Pariwisata & Akomodasi", "hotel", "Kunjungan wisatawan, jumlah hotel/penginapan, dan Tingkat Penghunian Kamar (TPK).", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=538"},
	{10, "keuangan-finansial", "Perbankan, Asuransi & Finansial", "bank", "Lembaga keuangan bank, koperasi, dan penyaluran kredit perbankan.", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=539"},
}

var pertanianTopicDefs = []topicDef{
	{1, "tanaman-perkebunan", "Pertanian, Perkebunan & Peternakan", "pertanian", "Komoditas kelapa sawit, lada (Muntok White Pepper), karet, padi, dan populasi ternak.", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=540"},
	{2, "perikanan-lingkungan", "Perikanan, Kehutanan & Lingkungan", "lingkungan", "Perikanan tangkap laut pesisir, budidaya tambak, hutan lindung, dan lingkungan hidup.", "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=541"},
}

var sektoralTopicDefs = []topicDef{
	{1, "sektoral-opd", "Statistik Sektoral OPD Bangka Barat", "sektoral", "Data statistik sektoral yang dikelola oleh Organisasi Perangkat Daerah (OPD) Pemkab Bangka Barat bersama BPS sebagai Pembina Data.", "https://data.bangkabaratkab.go.id/"},
}

// GetLayananDataCategories returns all 4 categories of Layanan Data (Menu 1).
func (p *Provider) GetLayananDataCategories(ctx context.Context) ([]DataCategory, Freshness, error) {
	layananCacheMu.RLock()
	cached, ok := layananCache["all"]
	layananCacheMu.RUnlock()
	if ok && time.Now().Before(cached.expiresAt) {
		return cached.categories, cached.freshness, nil
	}

	slugs := []string{"sosduk", "ekonomi", "pertanian", "sektoral"}
	cats := make([]DataCategory, len(slugs))
	for i, s := range slugs {
		c, _, err := p.GetLayananDataCategory(ctx, s)
		if err != nil {
			return nil, Freshness{}, err
		}
		cats[i] = *c
	}

	freshness := p.calculateFreshness(ctx, nil)

	layananCacheMu.Lock()
	layananCache["all"] = layananCacheEntry{
		categories: cats,
		freshness:  freshness,
		expiresAt:  time.Now().Add(5 * time.Minute),
	}
	layananCacheMu.Unlock()

	return cats, freshness, nil
}

// GetLayananDataCategory returns 1 category with all its topics enriched with scraped data.
func (p *Provider) GetLayananDataCategory(ctx context.Context, categorySlug string) (*DataCategory, Freshness, error) {
	slug := strings.ToLower(strings.TrimSpace(categorySlug))

	layananCacheMu.RLock()
	cached, ok := layananCache["cat:"+slug]
	layananCacheMu.RUnlock()
	if ok && time.Now().Before(cached.expiresAt) && len(cached.categories) > 0 {
		return &cached.categories[0], cached.freshness, nil
	}

	var cat *DataCategory
	var defs []topicDef

	switch slug {
	case "sosduk", "sosial", "sosial-kependudukan", "1":
		cat = &DataCategory{
			ID:          "1",
			Slug:        "sosduk",
			Title:       "Statistik Sosial Kependudukan",
			Description: "Koleksi data demografi, ketenagakerjaan, pendidikan, kemiskinan, dan kesehatan masyarakat Kabupaten Bangka Barat.",
		}
		defs = sosdukTopicDefs
	case "ekonomi", "perekonomian", "2":
		cat = &DataCategory{
			ID:          "2",
			Slug:        "ekonomi",
			Title:       "Statistik Ekonomi",
			Description: "Data makroekonomi daerah, PDRB, inflasi, industri, perdagangan, dan pariwisata Kabupaten Bangka Barat.",
		}
		defs = ekonomiTopicDefs
	case "pertanian", "agrikultur", "3":
		cat = &DataCategory{
			ID:          "3",
			Slug:        "pertanian",
			Title:       "Statistik Pertanian",
			Description: "Data komoditas unggulan perkebunan kelapa sawit, lada Muntok White Pepper, karet, dan perikanan tangkap laut.",
		}
		defs = pertanianTopicDefs
	case "sektoral", "opd", "4":
		cat = &DataCategory{
			ID:          "4",
			Slug:        "sektoral",
			Title:       "Statistik Sektoral",
			Description: "Data sektoral yang diproduksi oleh Organisasi Perangkat Daerah (OPD) Pemkab Bangka Barat terhubung ke portal Satu Data.",
		}
		defs = sektoralTopicDefs
	default:
		return nil, Freshness{}, fmt.Errorf("kategori layanan data tidak ditemukan: %q (pilihan: sosduk, ekonomi, pertanian, sektoral)", categorySlug)
	}

	topics := make([]DataTopic, len(defs))
	for i, d := range defs {
		topic := p.enrichTopic(ctx, d)
		topics[i] = *topic
	}
	cat.Topics = topics

	freshness := p.calculateFreshness(ctx, nil)

	layananCacheMu.Lock()
	layananCache["cat:"+slug] = layananCacheEntry{
		categories: []DataCategory{*cat},
		freshness:  freshness,
		expiresAt:  time.Now().Add(5 * time.Minute),
	}
	layananCacheMu.Unlock()

	return cat, freshness, nil
}

// GetLayananDataTopic returns detail of 1 specific topic under a category.
func (p *Provider) GetLayananDataTopic(ctx context.Context, categorySlug, topicIdOrSlug string) (*DataTopic, Freshness, error) {
	cat, freshness, err := p.GetLayananDataCategory(ctx, categorySlug)
	if err != nil {
		return nil, Freshness{}, err
	}

	query := strings.ToLower(strings.TrimSpace(topicIdOrSlug))
	for _, t := range cat.Topics {
		if strconv.Itoa(t.ID) == query || t.Slug == query || strings.Contains(strings.ToLower(t.Title), query) {
			return &t, freshness, nil
		}
	}

	return nil, Freshness{}, fmt.Errorf("topik %q tidak ditemukan pada kategori %q", topicIdOrSlug, categorySlug)
}

func (p *Provider) enrichTopic(ctx context.Context, d topicDef) *DataTopic {
	t := &DataTopic{
		ID:          d.id,
		Slug:        d.slug,
		Title:       d.title,
		Description: d.desc,
		SubjectURL:  d.subjectURL,
	}

	// 1. Ambil tabel statistik hasil scraping SQLite
	if tables, err := p.store.SearchItems(ctx, KindStaticTable, d.keyword, 3); err == nil && len(tables) > 0 {
		t.Tables = tables
	}

	// 2. Ambil publikasi & rilis terkait hasil scraping SQLite
	var relItems []Item
	if pubs, err := p.store.SearchItems(ctx, KindPublication, d.keyword, 2); err == nil {
		relItems = append(relItems, pubs...)
	}
	if prs, err := p.store.SearchItems(ctx, KindPressRelease, d.keyword, 1); err == nil {
		relItems = append(relItems, prs...)
	}
	t.Items = relItems

	return t
}
