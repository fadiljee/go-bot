# Product Requirements — Web Scraping & Scheduler Otomatis — BOT ISTIMEWA v3.1
## Spesifikasi Scraping `bangkabaratkab.bps.go.id` dengan Penjadwalan Otomatis Setiap 6 Jam

| Field | Detail |
|---|---|
| **Produk** | ISTIMEWA (Informasi Statistik Terpadu melalui WhatsApp) |
| **Instansi** | BPS Kabupaten Bangka Barat |
| **Arsitektur Target** | Go (`whatsmeow`) + Go Fiber v2 + SQLite + `internal/scraper` (baru) |
| **Dokumen** | Scraping & Scheduler Specification (`scraping.md`) |
| **Situs Target** | `https://bangkabaratkab.bps.go.id/` |
| **Dokumen Terkait** | `api_and_routegroup.md` (BPS Web API v1), `secure.md` (Hardening), `PROGRESS.md` |
| **Versi** | 3.1.0-draft |
| **Tanggal** | 27 Oktober 2026 |
| **Status** | Draft — Siap Direview sebelum Implementasi |

---

## 0. Cara Membaca Dokumen Ini

Dokumen ini ditulis agar bisa langsung dijadikan acuan implementasi tanpa tebak-tebakan. Setiap bagian ditandai dengan salah satu label berikut:

| Label | Arti |
|---|---|
| **[WAJIB]** | Harus diimplementasikan persis seperti tertulis. |
| **[LARANGAN]** | Dilarang dilakukan, apa pun alasannya. |
| **[TERVERIFIKASI]** | Fakta yang sudah dicek dari sumber nyata. |
| **[ASUMSI]** | Belum terverifikasi. Wajib divalidasi di **Fase 0** sebelum kode parser ditulis. |
| **[KEPUTUSAN]** | Pilihan desain yang diambil dokumen ini beserta alasannya. |

> **Aturan emas:** Jangan menulis selector CSS dari tebakan. Seluruh selector diisi dari berkas HTML nyata yang disimpan pada Fase 0 (lihat §8.1).

---

## 1. Ringkasan Eksekutif

### 1.1 Latar Belakang
BOT ISTIMEWA v3.0 saat ini mengambil data dari dua jalur: (1) Google Sheets via Google Apps Script (kurasi manual), dan (2) BPS Web API v1 (`internal/bpsapi`). Jalur kedua **belum aktif penuh** karena `BPS_API_KEY` belum diisi (lihat `PROGRESS.md` §6 butir 1). Akibatnya, data rilis resmi (BRS, publikasi, infografik, tabel statis) belum otomatis.

Dokumen ini menambahkan **jalur ketiga: scraping halaman publik** `bangkabaratkab.bps.go.id`, dijalankan otomatis oleh *scheduler* internal setiap **6 jam**, hasilnya disimpan ke SQLite lokal, dan dibaca oleh handler bot (Menu 1, 3, 5, dan pencarian 1.5).

### 1.2 Tujuan
1. Data rilis terbaru situs BPS Bangka Barat masuk ke bot **otomatis tanpa input admin**.
2. Scraping berjalan **terjadwal tiap 6 jam** (00:00, 06:00, 12:00, 18:00 WIB) dengan jitter acak.
3. Bot **tidak pernah memanggil situs BPS secara langsung saat melayani warga**. Bot hanya membaca dari SQLite/cache lokal, sehingga respons tetap cepat (<1,5 detik) dan tidak membebani server BPS.
4. Kegagalan scraping (diblokir, layout berubah, jaringan mati) **tidak boleh merusak data lama dan tidak boleh membuat bot crash**.
5. Scraping dilakukan **sopan dan sah**: mematuhi `robots.txt`, membatasi laju, mengidentifikasi diri, dan tidak mengakali pertahanan situs.

### 1.3 Non-Goals (di luar cakupan)
- Scraping situs selain `bangkabaratkab.bps.go.id` (domain provinsi/nasional dikerjakan lewat BPS Web API).
- Mengunduh dan menyimpan file PDF/XLSX/gambar ke disk server (yang disimpan hanya **metadata dan URL**; unduhan media terjadi saat bot mengirim, lewat jalur anti-SSRF yang sudah ada).
- Browser headless / eksekusi JavaScript (lihat keputusan K-03).
- Mengambil data pribadi atau data non-publik apa pun.
- Menggantikan BPS Web API. Scraping adalah **sumber cadangan/pelengkap**, bukan pengganti.

---

## 2. Temuan Awal & Status Verifikasi

### 2.1 Fakta Terverifikasi

| # | Temuan | Status | Dampak pada Desain |
|---|---|---|---|
| F-01 | Permintaan otomatis ke `https://bangkabaratkab.bps.go.id/` **ditolak oleh deteksi bot** saat dicoba untuk menyusun dokumen ini. | **[TERVERIFIKASI]** | Situs punya perlindungan anti-bot. Scraper **harus** mendeteksi pemblokiran dan berhenti dengan sopan (§7.7). Parser tidak bisa dikembangkan dengan mengunduh langsung dari server pengembangan, sehingga butuh Fase 0 (§8.1). |
| F-02 | Halaman BRS memakai pola URL `/{id\|en}/pressrelease/{YYYY}/{MM}/{DD}/{ID}/{slug}.html`. Contoh nyata: `/en/pressrelease/2021/09/01/537/...html`. | **[TERVERIFIKASI]** | Bagian `{ID}` numerik dipakai sebagai `source_key` untuk BRS. |
| F-03 | Halaman BRS memuat bagian berlabel **Jadwal Rilis**, **Ukuran File**, **Abstraksi**, dan **Berita Resmi Statistik Terkait**. | **[TERVERIFIKASI]** | Menjadi daftar field yang diekstrak dari halaman detail BRS. |
| F-04 | Footer situs memuat kontak resmi: Telp/Faks `(0716) 7323068`, email `bps1903@bps.go.id`, alamat Kompleks Perkantoran Pemerintah Kab. Bangka Barat, Kp. Daya Baru Pal 4 Muntok. | **[TERVERIFIKASI]** | Dipakai pada *User-Agent* sebagai kontak, dan sebagai tujuan permohonan *allowlist* IP (§3.3). |
| F-05 | Situs-situs BPS daerah memakai CMS/template yang sama. Pola `/publication.html?page=N` (pagination via query `page`) dan `/news/{YYYY}/{MM}/{DD}/{ID}/{slug}.html` terlihat di situs BPS lain. | **[TERVERIFIKASI]** pada situs BPS lain | Pola yang sama **[ASUMSI]** berlaku di Bangka Barat. Validasi di Fase 0. |
| F-06 | Banner situs berisi pengumuman yang berubah-ubah (SKD, jam layanan PST Senin–Jumat 08.00–15.30 WIB, publikasi terbaru). | **[TERVERIFIKASI]** | Banner **tidak** boleh ikut tersimpan sebagai konten data. Parser hanya mengambil blok konten utama. |

### 2.2 Asumsi yang Wajib Divalidasi (Fase 0)

| # | Asumsi | Cara Validasi |
|---|---|---|
| A-01 | Halaman daftar `publication`, `pressrelease`, `infographic`, `statictable`, dan `news` dirender di sisi server (HTML statis), bukan lewat JavaScript. | Simpan HTML via browser → *View Source* → pastikan data ada di sumber HTML. |
| A-02 | Pagination memakai `?page=N`. | Klik halaman 2 di browser, catat URL. |
| A-03 | Struktur ID publikasi/infografik/tabel di URL. | Catat 3 contoh URL tiap jenis. |
| A-04 | Isi `robots.txt` situs. | Buka `https://bangkabaratkab.bps.go.id/robots.txt` di browser, simpan salinannya ke `testdata/robots.txt`. |
| A-05 | Syarat penggunaan/ketentuan layanan data BPS memperbolehkan pengambilan otomatis atas data publik. | Konfirmasi ke tim humas/IPDS BPS (§3.2). |
| A-06 | Tautan unduh PDF/XLSX berupa tautan langsung atau melalui halaman perantara. | Klik tautan unduh dan catat alurnya. |

---

## 3. Prinsip Hukum, Etika, dan Keamanan Scraping

### 3.1 Prinsip **[WAJIB]**
1. **Patuhi `robots.txt`.** Sebelum menarik path apa pun, scraper memeriksa `robots.txt` (di-*cache* 24 jam). Path yang dilarang tidak boleh diambil.
2. **Identifikasi diri dengan jujur.** *User-Agent*:
   `BotIstimewa/3.1 (+https://bangkabaratkab.bps.go.id; kontak: bps1903@bps.go.id)`
3. **Laju rendah.** Maksimum 1 koneksi bersamaan, jeda minimal 3 detik antar-request, total ≤ 60 request per siklus (≈ 240 request/hari).
4. **Hanya data publik.** Tidak ada login, tidak ada *cookie* sesi, tidak ada data pribadi.
5. **Atribusi sumber.** Setiap data yang tampil di bot menyertakan tautan sumber ke halaman asli BPS.
6. **Hormati hak cipta.** Yang disimpan hanya metadata (judul, tanggal, abstrak singkat, URL). Isi publikasi tidak disalin utuh.

### 3.2 Konfirmasi Kebijakan **[KEPUTUSAN]**
Karena bot ini dioperasikan oleh BPS Kab. Bangka Barat sendiri, sebelum go-live tim wajib mendapatkan **persetujuan tertulis singkat** (email/memo) dari pimpinan/IPDS bahwa scraping situs sendiri dari server bot diperbolehkan, dan bahwa IP server boleh didaftarkan pada daftar izin (*allowlist*) WAF.

### 3.3 Penanganan Pemblokiran Bot (F-01) 

**[LARANGAN]** Hal-hal berikut **tidak boleh** diimplementasikan:
- Menyamar sebagai browser (memalsukan *User-Agent* Chrome/Firefox).
- Rotasi proxy/IP, VPN berputar, atau *residential proxy*.
- Menyelesaikan atau mengakali CAPTCHA/tantangan JavaScript.
- Browser headless dengan mode *stealth*.
- Mengulang request terus-menerus setelah diblokir.

**[KEPUTUSAN K-01]** Jika situs memblokir scraper yang jujur, solusinya **bukan** mengakali, tetapi:
1. **Minta *allowlist* IP server** bot ke pengelola situs (BPS Pusat/Biro Humas & Hukum atau tim TI) memakai kontak resmi di F-04.
2. **Aktifkan BPS Web API** (`BPS_API_KEY`), yaitu jalur resmi dan gratis yang memang disediakan untuk keperluan ini.
3. Selama belum bisa, bot tetap berjalan memakai API/Google Sheets (rantai fallback, §4.2).

> Catatan konsistensi: `api_and_routegroup.md` §3.1 menyebut *User-Agent* browser modern untuk client `bpsapi`. Itu berlaku untuk host `webapi.bps.go.id` dan di luar cakupan dokumen ini. Untuk **scraper**, aturan di atas yang berlaku. Disarankan meninjau ulang keputusan tersebut untuk client API.

---

## 4. Posisi Scraper dalam Arsitektur

### 4.1 Diagram Alur

```
                         ┌───────────────────────────────────────┐
                         │          WARGA (WhatsApp)             │
                         └───────────────────┬───────────────────┘
                                             │  pilih Menu 1 / 3 / 5 / 1.5
                                             ▼
                         ┌───────────────────────────────────────┐
                         │   internal/bot  (handlers_*.go)       │
                         │   Membaca dari DataProvider           │
                         │   TIDAK pernah memanggil web BPS      │
                         └───────────────────┬───────────────────┘
                                             │
                                             ▼
                         ┌───────────────────────────────────────┐
                         │   DataProvider (rantai fallback §4.2) │
                         └──┬──────────────┬──────────────┬──────┘
                            │              │              │
                   (1) SQLite scraper  (2) BPS Web API  (3) Google Sheets
                            ▲              (cache 15m)      (GAS, manual)
                            │
          ┌─────────────────┴────────────────────┐
          │   internal/scraper  (BARU)           │
          │   Scheduler (cron 0 */6 * * *, WIB)  │
          │   Fetcher → Parser → Validator → Store│
          └─────────────────┬────────────────────┘
                            │  HTTPS, 1 req / 3 dtk, patuh robots.txt
                            ▼
                 bangkabaratkab.bps.go.id
```

### 4.2 Rantai Fallback Sumber Data **[KEPUTUSAN K-02]**

Urutan prioritas saat bot membutuhkan data jenis `X`:

| Prioritas | Sumber | Dipakai jika |
|:---:|---|---|
| 1 | **BPS Web API v1** (`internal/bpsapi`) | `BPS_API_KEY` terisi dan respons sukses (atau cache stale-if-error tersedia). Sumber resmi, struktur stabil. |
| 2 | **SQLite hasil scraping** (`scraper.db`) | API tidak tersedia/key kosong/gagal, dan ada data scraping berumur ≤ `SCRAPER_MAX_STALE_HOURS` (default 48). |
| 3 | **Google Sheets (GAS)** | Dua sumber di atas kosong/gagal. |
| 4 | **Tautan portal statis** | Semua sumber gagal → bot mengirim tautan `https://bangkabaratkab.bps.go.id/` dengan pesan sopan. |

**Alasan scraper di prioritas 2:** API adalah jalur resmi yang paling stabil terhadap perubahan tampilan situs. Scraper menjadi penyelamat selama API key belum aktif atau API sedang bermasalah, serta menyuplai data yang mungkin tidak ada di API. Jika tim ingin membalik urutan (scraper di atas API), cukup ubah `DATA_SOURCE_ORDER` di `.env` (§13), tanpa mengubah kode.

### 4.3 Keputusan Teknis Utama

| ID | Keputusan | Alasan |
|---|---|---|
| **K-03** | Pakai HTTP biasa + `goquery` (parsing HTML statis). **Tidak** memakai browser headless. | Ringan (RAM <25 MB tetap terjaga), konsisten dengan target 1 vCPU/256 MB di `secure.md`. Jika A-01 ternyata salah (halaman dirender JS), **hentikan dan putuskan ulang**; jangan otomatis menambah headless. |
| **K-04** | Scheduler **internal Go** (`robfig/cron/v3`), bukan cron OS. | Binary tunggal, otomatis ikut hidup-mati dengan `systemd`, terintegrasi dengan lock/status/log. Cron OS tersedia sebagai opsi cadangan (§10.6). |
| **K-05** | Penyimpanan di **SQLite terpisah** `storage/scraper.db`. | Tidak mengganggu `whatsmeow.db` (berisi kunci Signal). Boleh dihapus/dibangun ulang tanpa risiko ke sesi WA. |
| **K-06** | Bot membaca dari DB, tidak pernah *live-scrape* saat melayani chat. | Latensi stabil, aman dari lonjakan trafik, tidak memicu WAF karena banyak pengguna. |
| **K-07** | Data lama **tidak dihapus** saat siklus gagal. Penandaan tidak aktif hanya setelah 3 siklus sukses berurutan tidak menemukan item itu. | Mencegah kehilangan data akibat gangguan sementara atau perubahan layout. |

---

## 5. Inventaris Data yang Di-scrape

> Kolom "Pola URL" bertanda ✔ sudah terverifikasi (F-02/F-05), bertanda ? masih **[ASUMSI]** dan harus dikonfirmasi di Fase 0.

### 5.1 Jenis Data (`kind`)

| `kind` | Menu Bot | Halaman Daftar | Halaman Detail | Pola URL |
|---|---|---|---|---|
| `pressrelease` | Menu 3 (BRS) | `/id/pressrelease.html?page=N` ? | `/id/pressrelease/{Y}/{M}/{D}/{ID}/{slug}.html` ✔ | ✔ detail / ? daftar |
| `publication` | Menu 3 (Publikasi) | `/publication.html?page=N` ✔ (pola dari BPS lain) | `/publication/{...}` ? | ? |
| `infographic` | Menu 5 | `/infographic.html?page=N` ? | `/infographic/{...}` ? | ? |
| `statictable` | Menu 1 & 1.5 | `/statictable.html?page=N` ? | `/statictable/{...}` ? | ? |
| `news` | opsional (info kegiatan) | `/news.html?page=N` ? | `/news/{Y}/{M}/{D}/{ID}/{slug}.html` ✔ (pola dari BPS lain) | ? |

> `news` berstatus **opsional / Fase 3**. Prioritas implementasi: `pressrelease` → `publication` → `infographic` → `statictable` → `news`.

### 5.2 Skema Field Per Item (model kanonik)

| Field | Tipe | Wajib | Keterangan |
|---|---|:---:|---|
| `kind` | enum | ✔ | Salah satu nilai §5.1. |
| `source_key` | string | ✔ | ID unik dari URL (segmen `{ID}`); jika tidak ada, `sha1(canonical_url_tanpa_slug)`. **Kunci deduplikasi.** |
| `title` | string | ✔ | Dibersihkan spasi/baris ganda; 1–300 karakter. |
| `url` | string | ✔ | URL **absolut** `https://bangkabaratkab.bps.go.id/...` ke halaman detail. |
| `release_date` | date (`YYYY-MM-DD`) | ✱ | Tanggal rilis (zona WIB). Wajib untuk `pressrelease` & `publication`. |
| `abstract` | string | – | Abstrak/deskripsi, dipotong 600 karakter, tanpa HTML. |
| `file_url` | string | – | Tautan unduh PDF/XLSX. Divalidasi aturan anti-SSRF (§7.8). |
| `file_type` | enum | – | `pdf`, `xlsx`, `xls`, `csv`, `png`, `jpg`, `other`. |
| `file_size_text` | string | – | Teks asli, mis. "1,2 MB". Tidak dikonversi. |
| `thumbnail_url` | string | – | Untuk `infographic`/`publication` (cover). |
| `image_url` | string | – | Khusus `infographic`: gambar resolusi penuh. |
| `subject` | string | – | Subjek/kategori jika tersedia. |
| `lang` | enum | ✔ | `id` atau `en`. Default `id`. |
| `content_hash` | string | ✔ | `sha256` dari gabungan field bermakna (title, release_date, abstract, file_url, image_url). Untuk deteksi perubahan. |
| `first_seen_at` / `last_seen_at` / `last_changed_at` | timestamp | ✔ | Diisi sistem (UTC). |
| `is_active` | bool | ✔ | Default `true`. |

---

## 6. Desain Paket `internal/scraper`

### 6.1 Struktur Berkas

```
internal/scraper/
├── config.go            # Struct Config + loader dari env, validasi nilai
├── fetcher.go           # HTTP client sopan: robots, rate limit, retry, deteksi blokir
├── robots.go            # Cache & evaluasi robots.txt
├── breaker.go           # Circuit breaker (cooldown eksponensial saat diblokir)
├── model.go             # Struct Item, Kind, Pagination, RunResult
├── dateparse.go         # Parser tanggal Indonesia/Inggris -> time.Time (WIB)
├── selectors.go         # Definisi selector per kind (dimuat dari YAML, lihat §8.3)
├── parser_pressrelease.go
├── parser_publication.go
├── parser_infographic.go
├── parser_statictable.go
├── parser_news.go
├── validate.go          # Validasi item & anomaly guard
├── store.go             # SQLite: migrasi, upsert, query, retensi
├── runner.go            # Orkestrasi satu siklus scrape (per kind), lock, timeout
├── scheduler.go         # robfig/cron: jadwal, jitter, startup catch-up
├── provider.go          # Implementasi DataProvider untuk handler bot
├── *_test.go            # Unit & golden test
└── testdata/
    ├── robots.txt
    ├── pressrelease_list_p1.html
    ├── pressrelease_detail_*.html
    ├── publication_list_p1.html
    ├── infographic_list_p1.html
    ├── statictable_list_p1.html
    ├── blocked_403.html      # contoh halaman blokir WAF
    └── layout_changed.html   # contoh layout rusak (untuk uji anomaly guard)

configs/
└── scraper_selectors.yaml    # Selector CSS (diisi di Fase 0, bukan di dalam kode)
```

### 6.2 Dependensi Baru (`go.mod`)

| Package | Fungsi | Catatan |
|---|---|---|
| `github.com/PuerkitoBio/goquery` | Parsing HTML & selector CSS | Pure Go, tanpa CGO. |
| `github.com/robfig/cron/v3` | Scheduler cron dengan zona waktu | Dipakai bersama `cron.SkipIfStillRunning`. |
| `github.com/temoto/robotstxt` | Parser & evaluator `robots.txt` | Pure Go. |
| `golang.org/x/time/rate` | Rate limiter keluar | **Sudah ada** di go.mod (`secure.md` §1.1). |
| `modernc.org/sqlite` | Driver SQLite | **Sudah ada.** Tanpa CGO. |
| `golang.org/x/net/html/charset` | Deteksi/konversi charset | Aman terhadap halaman non-UTF-8. |

**[LARANGAN]** Tidak boleh menambah dependensi yang memerlukan CGO atau browser (chromedp, rod, playwright, colly dengan JS engine).

### 6.3 Antarmuka Inti

```go
package scraper

type Kind string

const (
	KindPressRelease Kind = "pressrelease"
	KindPublication  Kind = "publication"
	KindInfographic  Kind = "infographic"
	KindStaticTable  Kind = "statictable"
	KindNews         Kind = "news"
)

// Item adalah model kanonik hasil parsing (lihat §5.2).
type Item struct {
	Kind          Kind
	SourceKey     string
	Title         string
	URL           string
	ReleaseDate   *time.Time // WIB, jam 00:00
	Abstract      string
	FileURL       string
	FileType      string
	FileSizeText  string
	ThumbnailURL  string
	ImageURL      string
	Subject       string
	Lang          string
	ContentHash   string
}

type Pagination struct {
	Page     int
	HasNext  bool
	LastPage int // 0 jika tidak diketahui
}

// Parser memisahkan "mengambil" dari "memahami" HTML sehingga bisa diuji dengan berkas fixture.
type Parser interface {
	Kind() Kind
	// ListURL membangun URL halaman daftar ke-n.
	ListURL(base *url.URL, lang string, page int) string
	// ParseList mengurai halaman daftar menjadi item (field minimal: title, url, source_key, release_date).
	ParseList(body []byte, pageURL *url.URL) ([]Item, Pagination, error)
	// ParseDetail melengkapi item dari halaman detail (abstract, file_url, ukuran file, dll).
	// Boleh mengembalikan ErrNoDetail jika semua field sudah lengkap dari halaman daftar.
	ParseDetail(body []byte, pageURL *url.URL, it *Item) error
}

// DataProvider dipakai handler bot. Handler TIDAK boleh mengimpor fetcher/parser.
type DataProvider interface {
	LatestPressReleases(ctx context.Context, lang string, limit int) ([]Item, Freshness, error)
	LatestPublications(ctx context.Context, lang string, limit int) ([]Item, Freshness, error)
	LatestInfographics(ctx context.Context, limit int) ([]Item, Freshness, error)
	SearchStaticTables(ctx context.Context, keyword string, limit int) ([]Item, Freshness, error)
}

// Freshness memberi tahu handler seberapa segar data (untuk label "diperbarui ...").
type Freshness struct {
	UpdatedAt time.Time
	Stale     bool // true jika melewati SCRAPER_MAX_STALE_HOURS
	Source    string // "scraper" | "bpsapi" | "gsheet"
}
```

---

## 7. Spesifikasi Fetcher (HTTP Client Sopan)

### 7.1 Parameter **[WAJIB]**

| Parameter | Nilai | Env |
|---|---|---|
| Host yang diizinkan | hanya `bangkabaratkab.bps.go.id` | `SCRAPER_ALLOWED_HOSTS` |
| Skema | hanya `https` | – |
| *Timeout* total per request | 20 detik | `SCRAPER_HTTP_TIMEOUT_SECONDS` |
| *Timeout* koneksi/TLS | 8 detik | – |
| Jeda antar-request | 3 detik + jitter 0–1 detik | `SCRAPER_MIN_DELAY_MS`, `SCRAPER_JITTER_MS` |
| Konkurensi | 1 | – |
| Batas ukuran body | 3 MB per halaman HTML | `SCRAPER_MAX_PAGE_BYTES` |
| Maks. redirect | 3, dan **tujuan harus tetap di host yang diizinkan** | – |
| Kompresi | gzip diterima (Go menangani otomatis) | – |
| Maks. halaman per `kind` per siklus | 5 halaman daftar | `SCRAPER_MAX_LIST_PAGES` |
| Maks. request per siklus | 60 | `SCRAPER_MAX_REQUESTS_PER_RUN` |
| Cookie jar | **tidak dipakai** | – |

### 7.2 Header Request **[WAJIB]**

```
User-Agent: BotIstimewa/3.1 (+https://bangkabaratkab.bps.go.id; kontak: bps1903@bps.go.id)
Accept: text/html,application/xhtml+xml;q=0.9,*/*;q=0.5
Accept-Language: id-ID,id;q=0.9,en;q=0.5
If-None-Match: <etag tersimpan>          # jika ada
If-Modified-Since: <last-modified>       # jika ada
```

### 7.3 Retry & Backoff

| Kondisi | Perilaku |
|---|---|
| Error jaringan, timeout, HTTP 500/502/504 | Retry maks. **2 kali**, jeda 5 dtk lalu 20 dtk (+jitter). |
| HTTP 429 | **Jangan retry.** Hormati header `Retry-After` (maks. 1 jam), tandai siklus `rate_limited`, hentikan seluruh siklus. |
| HTTP 403, 401, atau deteksi blokir (§7.7) | **Jangan retry.** Hentikan siklus, buka *circuit breaker* (§7.9). |
| HTTP 503 | Retry 1 kali; jika tetap 503 → hentikan siklus, status `upstream_unavailable`. |
| HTTP 404 pada halaman detail | Lewati item itu, catat peringatan. Bukan kegagalan siklus. |
| HTTP 404 pada halaman daftar | **Anomali** → hentikan `kind` tersebut, status `layout_or_route_changed`. |
| HTTP 304 | Pakai data tersimpan; tidak perlu parse ulang. |
| HTTP 3xx keluar host | Tolak (`ErrHostNotAllowed`). |

### 7.4 Conditional GET
Untuk setiap URL daftar, simpan `ETag` dan `Last-Modified` di tabel `page_cache` (§9.1). Kirim `If-None-Match`/`If-Modified-Since` pada siklus berikutnya. Respons `304` dihitung sebagai sukses tanpa perubahan. Ini mengurangi beban server BPS dan mempercepat siklus.

### 7.5 Pembatasan Body
Baca body dengan `io.LimitReader(resp.Body, MaxPageBytes+1)`. Jika terbaca lebih dari batas → `ErrBodyTooLarge`, halaman dilewati.

### 7.6 Validasi Content-Type
Terima hanya `text/html` (atau `application/xhtml+xml`). Jika tidak → anggap anomali, jangan di-parse.

### 7.7 Deteksi Halaman Blokir/Tantangan **[WAJIB]**
Respons dianggap **diblokir** jika salah satu terpenuhi, meski berstatus `200 OK`:
- Status 403/429 atau 503 dengan header `Server`/`CF-*` bernuansa WAF.
- Body mengandung penanda tantangan, mis. `captcha`, `Just a moment`, `Attention Required`, `Access Denied`, `Checking your browser`, `cf-chl`, `Request blocked`, `bot detection` (pencocokan tidak sensitif huruf besar-kecil; daftar bisa diperluas di config).
- Body sangat kecil (< 2 KB) padahal halaman daftar normal berukuran puluhan KB.
- Body tidak memuat elemen penanda halaman sah (mis. teks footer `Badan Pusat Statistik`).

Aksi: kembalikan `ErrBlocked`, **hentikan siklus**, jangan lanjut ke halaman lain.

### 7.8 Anti-SSRF & Keamanan URL
- Semua URL (termasuk hasil parsing seperti `file_url`/`image_url`) dinormalisasi menjadi absolut dan divalidasi: skema `https`, host bukan IP privat/loopback/*link-local* (`127.0.0.0/8`, `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`, `169.254.0.0/16`, `::1`, `fc00::/7`), tidak mengandung *userinfo* (`user:pass@`).
- Untuk **pengambilan halaman**, host harus ada di `SCRAPER_ALLOWED_HOSTS`.
- Untuk **`file_url`/`image_url`** yang hanya disimpan sebagai metadata, host boleh domain `*.bps.go.id`; jika di luar itu, field dikosongkan dan dicatat peringatan. Pengunduhan media saat bot mengirim tetap melewati validator anti-SSRF dan batas 16 MB di `internal/whatsapp` (`secure.md` §3.4).
- *DialContext* kustom wajib memeriksa IP hasil resolusi DNS (cegah *DNS rebinding*), bukan hanya string host.

### 7.9 Circuit Breaker **[WAJIB]**

| Kejadian beruntun (siklus) | Status | Cooldown sebelum mencoba lagi |
|---|---|---|
| 1× diblokir | `open` | 1 jam |
| 2× | `open` | 2 jam |
| 3× | `open` | 4 jam |
| 4× | `open` | 8 jam |
| ≥5× | `open` | 24 jam (batas maksimum) |

- Siklus terjadwal yang jatuh pada masa cooldown **dilewati** (status `skipped_breaker`), bukan dipaksa.
- Satu siklus sukses penuh mereset hitungan ke 0.
- Admin bisa mereset manual lewat `POST /api/v1/scraper/breaker/reset` (§12), misalnya setelah IP di-*allowlist*.
- Status breaker disimpan di SQLite agar bertahan saat bot restart.

### 7.10 Contoh Kerangka Kode Fetcher

```go
var (
	ErrBlocked          = errors.New("scraper: diblokir oleh situs (WAF/bot detection)")
	ErrRateLimited      = errors.New("scraper: rate limited (429)")
	ErrNotModified      = errors.New("scraper: 304 not modified")
	ErrRobotsDisallowed = errors.New("scraper: dilarang oleh robots.txt")
	ErrHostNotAllowed   = errors.New("scraper: host tidak diizinkan")
	ErrBodyTooLarge     = errors.New("scraper: body melebihi batas")
	ErrNotHTML          = errors.New("scraper: content-type bukan html")
)

type Conditional struct{ ETag, LastModified string }

type Response struct {
	URL          string
	Status       int
	Body         []byte
	ETag         string
	LastModified string
	FetchedAt    time.Time
}

func (f *Fetcher) Get(ctx context.Context, rawURL string, cond *Conditional) (*Response, error) {
	u, err := f.validate(rawURL) // skema https, host allowlist, bukan IP privat
	if err != nil {
		return nil, err
	}
	if !f.robots.Allowed(ctx, f.userAgent, u) {
		return nil, ErrRobotsDisallowed
	}
	if err := f.limiter.Wait(ctx); err != nil { // 1 req / ~3 dtk + jitter
		return nil, err
	}
	if f.budget.Exceeded() { // maks. request per siklus
		return nil, ErrBudgetExceeded
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	req.Header.Set("User-Agent", f.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml;q=0.9,*/*;q=0.5")
	req.Header.Set("Accept-Language", "id-ID,id;q=0.9,en;q=0.5")
	if cond != nil {
		if cond.ETag != "" {
			req.Header.Set("If-None-Match", cond.ETag)
		}
		if cond.LastModified != "" {
			req.Header.Set("If-Modified-Since", cond.LastModified)
		}
	}

	resp, err := f.doWithRetry(req) // retry sesuai §7.3
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotModified:
		return nil, ErrNotModified
	case resp.StatusCode == http.StatusTooManyRequests:
		return nil, ErrRateLimited
	case resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized:
		return nil, ErrBlocked
	case resp.StatusCode >= 400:
		return nil, fmt.Errorf("scraper: http %d untuk %s", resp.StatusCode, u)
	}
	if !isHTML(resp.Header.Get("Content-Type")) {
		return nil, ErrNotHTML
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, f.maxBody+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > f.maxBody {
		return nil, ErrBodyTooLarge
	}
	if looksBlocked(body) { // §7.7, berlaku walau status 200
		return nil, ErrBlocked
	}
	return &Response{ /* ... */ }, nil
}
```

---

## 8. Spesifikasi Parser

### 8.1 Fase 0 — Pengumpulan Fixture (**WAJIB sebelum menulis parser**)
Karena situs memblokir akses otomatis dari lingkungan pengembangan (F-01), fixture dikumpulkan **manual lewat browser biasa** oleh anggota tim BPS:

1. Buka masing-masing halaman di browser → *Save Page As → Webpage, HTML only* (atau *View Source* → simpan).
2. Halaman yang dikumpulkan (nama berkas sesuai §6.1):
   - `robots.txt`
   - Daftar BRS (halaman 1 dan 2), 3 detail BRS (beda bulan, minimal 1 punya tautan PDF, 1 tanpa)
   - Daftar publikasi (halaman 1 dan 2), 2 detail publikasi
   - Daftar infografik (halaman 1), 2 detail infografik
   - Daftar tabel statis (halaman 1), 2 detail tabel statis
   - (Opsional) daftar berita + 1 detail
   - Satu halaman dengan hasil pencarian kosong (jika ada fitur pencarian)
3. Catat untuk tiap jenis: **pola URL daftar, parameter pagination, pola URL detail, lokasi ID**.
4. Simpan ke `internal/scraper/testdata/` dan tulis ringkasan temuan di `internal/scraper/testdata/README.md`.
5. **Hasil Fase 0 mengisi `configs/scraper_selectors.yaml`.** Tidak ada selector yang ditulis sebelum fixture ada.

**Kriteria selesai Fase 0:** semua asumsi A-01…A-06 punya status "terkonfirmasi" atau "salah (dengan koreksi)" di `testdata/README.md`.

### 8.2 Aturan Umum Parsing **[WAJIB]**
1. Gunakan `goquery`; **jangan** regex pada HTML mentah untuk mengambil struktur.
2. Dekode charset dengan `golang.org/x/net/html/charset`.
3. Ambil hanya **blok konten utama**; abaikan header, banner pengumuman (F-06), sidebar, footer.
4. Normalisasi teks: `strings.TrimSpace`, ganti rangkaian spasi/*newline*/NBSP (`\u00a0`) menjadi satu spasi, buang karakter kontrol.
5. Hasil teks **tidak boleh** mengandung tag HTML. Gunakan `.Text()`, bukan `.Html()`.
6. URL relatif → absolut dengan `url.ResolveReference(pageURL)`, lalu validasi (§7.8). Buang fragmen `#...`.
7. Parser **tidak boleh panic**. Semua akses harus aman terhadap `nil`/elemen hilang. Bungkus tiap item dengan `recover` dan catat sebagai item gagal, bukan siklus gagal.
8. Satu item gagal tidak membatalkan item lain. Catat jumlah `items_skipped` per siklus.
9. Hasil parse diurutkan stabil (terbaru dulu) untuk hasil yang konsisten.

### 8.3 Selector Dieksternalkan (`configs/scraper_selectors.yaml`)

Selector disimpan di berkas YAML agar perbaikan saat layout BPS berubah **cukup mengedit config**, tanpa kompilasi ulang. Berikut **kerangka** (nilai `TBD` diisi dari Fase 0, bukan ditebak):

```yaml
version: 1
pressrelease:
  list_url: "/{lang}/pressrelease.html?page={page}"   # TBD – konfirmasi A-02
  list:
    item:        "TBD"      # wadah tiap item di halaman daftar
    title:       "TBD"
    link:        "TBD"      # atribut href
    date:        "TBD"
    next_page:   "TBD"      # penanda ada halaman berikutnya
  detail:
    abstract:    "TBD"      # bagian "Abstraksi"
    release_date:"TBD"      # bagian "Jadwal Rilis"
    file_size:   "TBD"      # bagian "Ukuran File"
    file_link:   "TBD"      # tautan unduh
  required_markers:         # teks yang HARUS ada agar halaman dianggap sah
    - "Badan Pusat Statistik"
publication:
  list_url: "/publication.html?page={page}"
  # ... dst, sama strukturnya
infographic: {}
statictable: {}
news: {}
```

Loader wajib **menolak** config yang masih berisi `TBD` pada `kind` yang diaktifkan (`SCRAPER_ENABLED_KINDS`), dengan pesan error jelas saat startup (*fail fast*).

### 8.4 Parser Tanggal **[WAJIB]**

Situs memakai tanggal berbahasa Indonesia ("1 September 2021") atau Inggris ("September 1, 2021").

| Indonesia | Angka | Inggris |
|---|:---:|---|
| Januari | 1 | January |
| Februari | 2 | February |
| Maret | 3 | March |
| April | 4 | April |
| Mei | 5 | May |
| Juni | 6 | June |
| Juli | 7 | July |
| Agustus | 8 | August |
| September | 9 | September |
| Oktober | 10 | October |
| November / Nopember | 11 | November |
| Desember | 12 | December |

Aturan:
- Cocokkan tidak sensitif huruf besar-kecil; terima singkatan 3 huruf (Jan, Feb, Mar, Apr, Mei, Jun, Jul, Agu/Agt/Aug, Sep, Okt/Oct, Nov, Des/Dec).
- Hasil selalu `time.Date(y, m, d, 0, 0, 0, 0, time.FixedZone("WIB", 7*3600))`. Gunakan `time.LoadLocation("Asia/Jakarta")` dan fallback ke `FixedZone` jika *tzdata* tidak tersedia (impor `_ "time/tzdata"` agar aman di container minimal).
- Tanggal tidak valid (mis. 31 Februari), tahun di luar 1990–(tahun sekarang+1) → ditolak; item tetap diproses dengan `release_date = nil` **dan** peringatan, kecuali kind `pressrelease`/`publication` yang mewajibkan tanggal (item dilewati).
- Fallback: bila tanggal tidak ada di teks, ambil dari URL (`/{YYYY}/{MM}/{DD}/`) untuk pola F-02.
- Wajib ada tes tabel (*table-driven test*) minimal 20 kasus termasuk spasi ganda, NBSP, huruf kapital, dan format Inggris.

### 8.5 Ekstraksi `source_key`
1. Untuk URL berpola F-02, ambil segmen angka setelah `/{DD}/` (mis. `537`).
2. Untuk pola lain, ambil segmen ID sesuai hasil Fase 0.
3. Jika gagal: `sha1(lowercase(host + path tanpa slug terakhir))`.
4. `source_key` tidak boleh kosong. Item tanpa `source_key` dilewati.
5. Kunci unik DB adalah pasangan `(kind, lang, source_key)`.

### 8.6 Validasi Item (`validate.go`)
Item **ditolak** (dicatat sebagai `items_skipped`) jika:
- `title` kosong atau > 300 karakter setelah normalisasi.
- `url` tidak lolos validasi §7.8.
- `source_key` kosong.
- `release_date` wajib tapi kosong/tidak valid.
- `file_url` ada tetapi gagal validasi (field dikosongkan; item tetap disimpan jika lainnya valid).

### 8.7 Anomaly Guard (Perlindungan Perubahan Layout) **[WAJIB]**
Setelah parsing satu `kind`, sebelum menulis ke DB:
1. Jika **0 item** ditemukan padahal siklus sebelumnya menemukan ≥ 5 → **tolak hasil**, status `layout_changed`, data lama dipertahankan.
2. Jika jumlah item < **50%** dari jumlah item siklus sukses terakhir (untuk halaman yang sama) → **tolak hasil**, status `anomaly_count_drop`.
3. Jika > **30%** item valid kehilangan field wajib (judul/URL) → **tolak hasil**, status `anomaly_field_loss`.
4. Jika `required_markers` (§8.3) tidak ditemukan di halaman → anggap halaman blokir/salah, `ErrBlocked` atau `layout_changed`.
5. Tiap penolakan memicu **notifikasi admin** (§14.3) dan **tidak** menaikkan counter *circuit breaker* (kecuali `ErrBlocked`).

---

## 9. Penyimpanan Data (SQLite)

### 9.1 Lokasi & Pengaturan
- Berkas: `./storage/scraper.db` (env `SCRAPER_DB_PATH`). Izin berkas `600`, direktori `700` (selaras `secure.md` §3.5).
- Driver `modernc.org/sqlite`; aktifkan `PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000; PRAGMA foreign_keys=ON; PRAGMA synchronous=NORMAL;`.
- Tulis dalam **satu transaksi per `kind`** (semua item sukses atau tidak sama sekali).
- Migrasi skema berversi (`PRAGMA user_version`) dijalankan otomatis saat startup.
- `.gitignore` wajib memuat `storage/*.db`, `*.db-wal`, `*.db-shm`.

### 9.2 Skema DDL

```sql
-- v1
CREATE TABLE IF NOT EXISTS scraped_items (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  kind            TEXT    NOT NULL CHECK (kind IN ('pressrelease','publication','infographic','statictable','news')),
  lang            TEXT    NOT NULL DEFAULT 'id' CHECK (lang IN ('id','en')),
  source_key      TEXT    NOT NULL,
  title           TEXT    NOT NULL,
  url             TEXT    NOT NULL,
  release_date    TEXT,                       -- 'YYYY-MM-DD' (WIB)
  abstract        TEXT,
  file_url        TEXT,
  file_type       TEXT,
  file_size_text  TEXT,
  thumbnail_url   TEXT,
  image_url       TEXT,
  subject         TEXT,
  content_hash    TEXT    NOT NULL,
  is_active       INTEGER NOT NULL DEFAULT 1,
  missed_runs     INTEGER NOT NULL DEFAULT 0, -- siklus sukses berturut-turut tanpa item ini
  first_seen_at   TEXT    NOT NULL,           -- RFC3339 UTC
  last_seen_at    TEXT    NOT NULL,
  last_changed_at TEXT    NOT NULL,
  UNIQUE (kind, lang, source_key)
);
CREATE INDEX IF NOT EXISTS idx_items_kind_date ON scraped_items(kind, lang, is_active, release_date DESC);
CREATE INDEX IF NOT EXISTS idx_items_title     ON scraped_items(kind, title COLLATE NOCASE);

CREATE TABLE IF NOT EXISTS scrape_runs (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  trigger       TEXT NOT NULL CHECK (trigger IN ('cron','manual','startup')),
  kind          TEXT NOT NULL,                -- 'all' atau salah satu kind
  started_at    TEXT NOT NULL,
  finished_at   TEXT,
  status        TEXT NOT NULL,                -- success | partial | blocked | rate_limited | layout_changed |
                                              -- anomaly_count_drop | anomaly_field_loss | upstream_unavailable |
                                              -- skipped_breaker | skipped_overlap | error | timeout
  requests      INTEGER NOT NULL DEFAULT 0,
  items_found   INTEGER NOT NULL DEFAULT 0,
  items_new     INTEGER NOT NULL DEFAULT 0,
  items_updated INTEGER NOT NULL DEFAULT 0,
  items_skipped INTEGER NOT NULL DEFAULT 0,
  error_message TEXT
);
CREATE INDEX IF NOT EXISTS idx_runs_started ON scrape_runs(started_at DESC);

CREATE TABLE IF NOT EXISTS page_cache (
  url           TEXT PRIMARY KEY,
  etag          TEXT,
  last_modified TEXT,
  body_sha256   TEXT,
  fetched_at    TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS breaker_state (
  id               INTEGER PRIMARY KEY CHECK (id = 1),
  consecutive_fail INTEGER NOT NULL DEFAULT 0,
  open_until       TEXT,
  last_reason      TEXT,
  updated_at       TEXT NOT NULL
);
INSERT OR IGNORE INTO breaker_state (id, consecutive_fail, updated_at) VALUES (1, 0, datetime('now'));
```

### 9.3 Logika Upsert **[WAJIB]**
Untuk setiap item valid pada siklus sukses:
1. Cari berdasarkan `(kind, lang, source_key)`.
2. **Tidak ada** → `INSERT` (`first_seen_at = last_seen_at = last_changed_at = now`), hitung `items_new`.
3. **Ada, `content_hash` sama** → hanya `UPDATE last_seen_at = now, missed_runs = 0, is_active = 1`.
4. **Ada, `content_hash` beda** → `UPDATE` semua field + `last_changed_at = now`, hitung `items_updated`.
5. Item yang ada di DB tetapi **tidak muncul** di siklus sukses lengkap → `missed_runs += 1`. Jika `missed_runs >= 3` → `is_active = 0` (**bukan** `DELETE`).
6. Pada siklus `partial`/gagal, **jangan** menyentuh `missed_runs` maupun `is_active` (K-07).
7. Jika jumlah halaman daftar yang dipindai terbatas (`SCRAPER_MAX_LIST_PAGES`), item yang berada di luar jangkauan tidak boleh dianggap hilang: langkah 5 hanya diterapkan pada item yang **secara logika seharusnya terlihat** (mis. `release_date` lebih baru dari item tertua yang dipindai).

### 9.4 Retensi
- `scrape_runs`: simpan 90 hari terakhir; hapus yang lebih lama saat siklus harian pertama.
- `scraped_items` tidak aktif (`is_active=0`) lebih dari 365 hari boleh dihapus.
- `VACUUM` dijalankan maksimal sebulan sekali (saat jam sepi), atau cukup `PRAGMA wal_checkpoint(TRUNCATE)` setelah siklus.

### 9.5 Cache Memori
`provider.go` boleh menaruh hasil query populer (10 item terbaru per kind) di `go-cache` (TTL 5 menit), dan **wajib** meng-invalidasi cache itu setiap kali siklus scrape sukses mengubah data.

---

## 10. Spesifikasi Scheduler (Cron 6 Jam)

### 10.1 Jadwal **[WAJIB]**

| Pengaturan | Nilai |
|---|---|
| Ekspresi cron | `0 */6 * * *` (menit 0, tiap 6 jam) |
| Jam eksekusi (WIB) | **00:00, 06:00, 12:00, 18:00** |
| Zona waktu | `Asia/Jakarta` (WIB, UTC+7, tanpa DST) |
| Jitter awal | acak 0–300 detik sebelum siklus mulai (agar tidak selalu tepat di awal jam) |
| Env | `SCRAPER_CRON_SPEC="0 */6 * * *"`, `SCRAPER_TIMEZONE=Asia/Jakarta` |

> Kenapa jitter? Banyak sistem berjalan tepat di menit :00. Jitter mengurangi beban puncak pada server BPS dan terlihat lebih wajar.

### 10.2 Aturan Eksekusi **[WAJIB]**
1. **Tanpa tumpang tindih.** Gunakan `cron.SkipIfStillRunning` **dan** `sync.Mutex`/`TryLock` internal. Siklus yang terlewat dicatat `skipped_overlap`.
2. **Batas waktu keras per siklus:** 10 menit (`SCRAPER_RUN_TIMEOUT_MINUTES`) lewat `context.WithTimeout`. Jika lewat, hentikan dengan status `timeout` dan simpan hasil `kind` yang sudah selesai (status `partial`).
3. **Urutan kind** per siklus mengikuti prioritas: `pressrelease` → `publication` → `infographic` → `statictable` → `news`. Satu `kind` gagal **tidak** menghentikan kind lain, **kecuali** `ErrBlocked`/`ErrRateLimited` (hentikan seluruh siklus).
4. **Pemulihan panic.** Pembungkus job wajib `defer recover()`, mencatat stack trace via `slog`, dan menandai siklus `error`. Bot WA **tidak boleh** ikut mati.
5. **Catch-up saat startup.** Saat bot dinyalakan: jika data terakhir sukses berumur > `SCRAPER_STARTUP_CATCHUP_HOURS` (default 6) atau DB kosong, jalankan **satu** siklus setelah penundaan 60–180 detik (agar tidak berebut dengan koneksi WA awal). Ini mencegah data kosong setelah server mati lama.
6. **Siklus hanya jika breaker tertutup** (§7.9). Jika terbuka → `skipped_breaker`.
7. **Graceful shutdown.** Pada `SIGTERM`/`SIGINT`: hentikan penjadwalan baru, batalkan `context` siklus berjalan, tunggu maks. 10 detik, lalu `Close()` DB.
8. **Tidak berjalan saat `APP_ENV=test`** atau jika `SCRAPER_ENABLED=false`.
9. **Trigger manual** (`POST /api/v1/scraper/run`) memakai mutex yang sama; ditolak `409 Conflict` bila siklus sedang berjalan, dan dibatasi 1 kali per 5 menit.

### 10.3 Kerangka Kode Scheduler

```go
func (s *Scheduler) Start() error {
	loc, err := time.LoadLocation(s.cfg.Timezone) // "Asia/Jakarta"
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	s.cron = cron.New(
		cron.WithLocation(loc),
		cron.WithChain(
			cron.Recover(cronLogger{}),    // tangkap panic
			cron.SkipIfStillRunning(cronLogger{}),
		),
	)
	if _, err := s.cron.AddFunc(s.cfg.CronSpec, func() {
		jitter := time.Duration(rand.Int63n(int64(300 * time.Second)))
		select {
		case <-time.After(jitter):
		case <-s.ctx.Done():
			return
		}
		s.runner.RunAll(s.ctx, TriggerCron)
	}); err != nil {
		return fmt.Errorf("scheduler: cron spec tidak valid %q: %w", s.cfg.CronSpec, err)
	}
	s.cron.Start()
	go s.catchUpOnStartup()
	slog.Info("scraper scheduler aktif", "spec", s.cfg.CronSpec, "tz", loc.String())
	return nil
}

func (s *Scheduler) Stop(ctx context.Context) {
	stopCtx := s.cron.Stop() // berhenti menjadwalkan, kembalikan ctx yang selesai ketika job berjalan tuntas
	select {
	case <-stopCtx.Done():
	case <-ctx.Done(): // batas tunggu 10 detik
	}
}
```

### 10.4 Pseudocode `Runner.RunAll`

```
RunAll(ctx, trigger):
  if !mutex.TryLock(): catat skipped_overlap; return
  defer mutex.Unlock()
  if breaker.IsOpen(): catat skipped_breaker; return
  ctx = WithTimeout(ctx, RunTimeout)
  run = store.StartRun(trigger)
  for kind in EnabledKinds (urut prioritas):
      items, err = scrapeKind(ctx, kind)     // list -> detail -> validate -> anomaly guard
      switch err:
        nil:                        store.Apply(kind, items)  // transaksi, upsert §9.3
        ErrBlocked, ErrRateLimited: breaker.Trip(err); run.status = blocked/rate_limited; notifyAdmin; break loop
        anomaly*:                   run.status = partial; notifyAdmin; continue (data lama aman)
        other:                      run.status = partial; log; continue
  if semua kind sukses: breaker.Reset(); run.status = success
  store.FinishRun(run); provider.InvalidateCache()
```

### 10.5 Alur Satu `kind` (`scrapeKind`)

```
1. page = 1
2. loop sampai page > MaxListPages atau tidak ada halaman berikutnya:
     resp = fetcher.Get(ListURL(page), conditional dari page_cache)
     - 304  -> ambil item terakhir dari DB untuk halaman ini, lanjut
     - ok   -> ParseList -> kumpulkan item; simpan etag/last-modified
     - stop lebih awal jika SEMUA item di halaman ini sudah dikenal & hash tak berubah (early-stop)
3. untuk tiap item BARU atau yang butuh detail (field wajib belum lengkap):
     resp = fetcher.Get(item.URL)   // batas request per siklus berlaku
     ParseDetail(resp.Body, &item)
4. validate(item) -> buang yang gagal
5. anomalyGuard(kind, items, prevCount)
6. return items
```

**Early-stop** mengurangi request: setelah siklus pertama (backfill), siklus berikutnya biasanya hanya butuh 1–2 request per kind.

### 10.6 Opsi Cadangan: Cron OS / systemd Timer **[OPSIONAL]**
Jika scheduler internal ingin dimatikan (`SCRAPER_ENABLED=false` untuk scheduler saja), jadwal dapat dipicu dari luar lewat endpoint terproteksi (§12). Contoh `crontab` (zona waktu server diset ke Asia/Jakarta):

```cron
# Setiap 6 jam, dengan jeda acak 0-300 detik
0 */6 * * * sleep $((RANDOM % 300)) && curl -fsS -X POST \
  -H "X-API-Key: ${API_SECRET_KEY}" \
  http://127.0.0.1:3000/api/v1/scraper/run?job=all >> /var/log/botistimewa/scraper-cron.log 2>&1
```

Atau *systemd timer*:

```ini
# /etc/systemd/system/botistimewa-scrape.timer
[Unit]
Description=Scrape BPS Bangka Barat tiap 6 jam

[Timer]
OnCalendar=*-*-* 00,06,12,18:00:00
Persistent=true
RandomizedDelaySec=300
Unit=botistimewa-scrape.service

[Install]
WantedBy=timers.target
```

```ini
# /etc/systemd/system/botistimewa-scrape.service
[Unit]
Description=Trigger scraping BOT ISTIMEWA

[Service]
Type=oneshot
User=botistimewa
EnvironmentFile=/opt/botistimewa/.env
ExecStart=/usr/bin/curl -fsS -X POST -H "X-API-Key: ${API_SECRET_KEY}" http://127.0.0.1:3000/api/v1/scraper/run?job=all
```

> **Jangan aktifkan dua pemicu sekaligus** (scheduler internal + cron OS). Mutex mencegah tabrakan, tetapi tetap membuang siklus. Pilih salah satu; default dokumen ini adalah **scheduler internal**.

---

## 11. Integrasi dengan Bot WhatsApp

### 11.1 Aturan Handler **[WAJIB]**
1. Handler hanya memanggil `DataProvider` (§6.3). **Dilarang** memanggil `fetcher`/`parser` langsung dari handler.
2. Seluruh data berasal dari DB/cache → respons tetap < 1,5 detik (target `secure.md` §2.1).
3. Setiap daftar hasil menyertakan baris **sumber & kesegaran**, mis. *"Sumber: bangkabaratkab.bps.go.id · diperbarui 27 Okt 2026 06:03 WIB"*. Jika `Freshness.Stale == true`, tambahkan catatan *"data mungkin belum terbaru, cek langsung di portal"*.
4. Teks bebas dari situs (judul/abstrak) diperlakukan **data tidak tepercaya**: dipotong, tanpa HTML, karakter kontrol dibuang, tidak pernah dipakai sebagai perintah. Hindari menyisipkan markup format WhatsApp (`*`, `_`, `~`, `` ` ``) mentah dari sumber; escape jika perlu agar tidak merusak tampilan.
5. Kegagalan di rantai fallback tidak pernah memunculkan pesan teknis ke warga; gunakan kamus `locales/id_ID.json` (tambahkan kunci baru, jangan *hardcode* teks).

### 11.2 Pemetaan Menu → Data

| Menu | Data | Isi Pesan | Catatan |
|---|---|---|---|
| **Menu 3 – BRS** | `pressrelease` | 5–10 BRS terbaru: judul, tanggal rilis, tautan halaman/PDF | List message (pilih BRS → kirim detail & tautan). |
| **Menu 3 – Publikasi** | `publication` | Judul, tanggal rilis, ukuran file, tautan unduh | Contoh "Bangka Barat Dalam Angka". |
| **Menu 5 – Infografik** | `infographic` | Kirim **gambar** infografik terbaru + judul | Unduh gambar lewat jalur media anti-SSRF, batas 16 MB; jika gagal kirim tautan saja. |
| **Menu 1 & 1.5 – Tabel/Pencarian** | `statictable` | Hasil pencarian kata kunci: judul tabel + tautan | Pencarian di kolom `title` (+`subject`), `LIKE` dengan parameter terikat. |

### 11.3 Batas Tampilan WhatsApp
- Pesan teks maks. ±4.000 karakter → pecah jika perlu atau potong dengan tautan.
- Untuk *list message*, patuhi batas builder yang ada di `internal/whatsapp` (judul baris pendek, deskripsi singkat, maksimum 10 baris). **Verifikasi batas persisnya pada kode `list builder` saat ini**; potong judul dengan elipsis `…`.
- Input pencarian pengguna tetap dipangkas 100 karakter (`secure.md` §3.2) dan **dipakai sebagai parameter SQL terikat** (`?`), tidak pernah disambung ke string query.

### 11.4 Pseudocode Handler dengan Fallback

```go
func (h *Handler) latestPressReleases(ctx context.Context, jid types.JID) {
	items, fresh, err := h.providers.PressReleases(ctx, "id", 8) // rantai fallback §4.2 di dalam provider
	if err != nil || len(items) == 0 {
		h.reply(jid, h.t("scraper.fallback.portal_link")) // pesan sopan + tautan portal
		return
	}
	h.replyList(jid, buildPressReleaseList(items), footerFreshness(fresh))
}
```

---

## 12. Endpoint Admin REST (Go Fiber)

Semua di bawah `/api/v1/scraper`, **wajib** `authMiddleware` (`X-API-Key` / `Authorization: Bearer`), tetap terikat `127.0.0.1` (`secure.md` §3.3).

| Method | Path | Fungsi |
|---|---|---|
| `GET` | `/api/v1/scraper/status` | Status ringkas: siklus terakhir, jadwal berikutnya, status *breaker*, jumlah item per kind, umur data. |
| `POST` | `/api/v1/scraper/run?job=all\|{kind}` | Picu siklus manual (async, `202 Accepted`). `409` jika sedang berjalan; `429` jika < 5 menit sejak pemicu manual terakhir; `503` jika breaker terbuka. |
| `GET` | `/api/v1/scraper/runs?limit=20` | Riwayat siklus dari `scrape_runs`. `limit` maks. 100. |
| `GET` | `/api/v1/scraper/items?kind=&lang=&page=&q=` | Telusuri data tersimpan (paginasi, maks. 50/halaman). |
| `POST` | `/api/v1/scraper/breaker/reset` | Reset *circuit breaker* secara manual (dicatat di log audit). |

Format respons selaras `api_and_routegroup.md` §4.2:

```json
{
  "success": true,
  "scheduler": { "enabled": true, "cron": "0 */6 * * *", "timezone": "Asia/Jakarta", "next_run": "2026-10-27T18:00:00+07:00" },
  "last_run": { "id": 128, "status": "success", "started_at": "2026-10-27T12:02:11+07:00", "items_new": 1, "items_updated": 0, "requests": 6 },
  "breaker": { "open": false, "consecutive_fail": 0, "open_until": null },
  "counts": { "pressrelease": 48, "publication": 31, "infographic": 22, "statictable": 120 },
  "data_age_hours": 3.2
}
```

Tambahan pada `GET /health` (publik, tanpa data sensitif): `scraper_last_success_age_hours` dan `scraper_breaker_open` (boolean).

---

## 13. Konfigurasi Environment (`.env` & `.env.example`)

```env
# ==========================================
# 8. KONFIGURASI SCRAPER BPS BANGKA BARAT
# ==========================================
SCRAPER_ENABLED=true
SCRAPER_BASE_URL=https://bangkabaratkab.bps.go.id
SCRAPER_ALLOWED_HOSTS=bangkabaratkab.bps.go.id
SCRAPER_USER_AGENT=BotIstimewa/3.1 (+https://bangkabaratkab.bps.go.id; kontak: bps1903@bps.go.id)
SCRAPER_DB_PATH=./storage/scraper.db
SCRAPER_SELECTORS_PATH=./configs/scraper_selectors.yaml

# Jadwal
SCRAPER_CRON_SPEC=0 */6 * * *
SCRAPER_TIMEZONE=Asia/Jakarta
SCRAPER_RUN_TIMEOUT_MINUTES=10
SCRAPER_STARTUP_CATCHUP_HOURS=6
SCRAPER_MANUAL_COOLDOWN_MINUTES=5

# Kind yang diaktifkan (urut prioritas). Kind baru hanya diaktifkan setelah selector-nya selesai.
SCRAPER_ENABLED_KINDS=pressrelease,publication,infographic,statictable

# Kesopanan & batas
SCRAPER_MIN_DELAY_MS=3000
SCRAPER_JITTER_MS=1000
SCRAPER_HTTP_TIMEOUT_SECONDS=20
SCRAPER_MAX_PAGE_BYTES=3145728
SCRAPER_MAX_LIST_PAGES=5
SCRAPER_MAX_REQUESTS_PER_RUN=60
SCRAPER_RESPECT_ROBOTS=true

# Kesegaran & prioritas sumber data
SCRAPER_MAX_STALE_HOURS=48
DATA_SOURCE_ORDER=bpsapi,scraper,gsheet

# Notifikasi admin (nomor harus ada di ADMIN_WHITELIST_NUMBERS)
SCRAPER_NOTIFY_ADMIN=true
SCRAPER_NOTIFY_AFTER_FAILURES=2
```

**Validasi startup `config.go` [WAJIB]:** gagal-cepat (`log.Fatal`/error jelas) jika: `SCRAPER_CRON_SPEC` tidak valid; `SCRAPER_ALLOWED_HOSTS` kosong atau mengandung IP/`localhost`; `SCRAPER_BASE_URL` bukan `https`; `SCRAPER_MIN_DELAY_MS` < 1000 (dilarang lebih agresif); `SCRAPER_RESPECT_ROBOTS=false` saat `APP_ENV=production`; `SCRAPER_MAX_REQUESTS_PER_RUN` > 200; kind aktif yang selector-nya masih `TBD`.

---

## 14. Observabilitas, Logging, & Notifikasi

### 14.1 Log Terstruktur (`log/slog`, JSON)
Setiap siklus dan request mencatat field yang konsisten:

| Event | Field wajib |
|---|---|
| `scrape.run.start` | `run_id`, `trigger`, `kinds` |
| `scrape.fetch` | `url`, `status`, `bytes`, `duration_ms`, `cached(304)` |
| `scrape.parse` | `kind`, `page`, `items`, `skipped` |
| `scrape.apply` | `kind`, `new`, `updated`, `deactivated` |
| `scrape.blocked` | `url`, `status`, `reason`, `breaker_until` |
| `scrape.anomaly` | `kind`, `type`, `prev_count`, `new_count` |
| `scrape.run.end` | `run_id`, `status`, `duration_ms`, `requests` |

**[LARANGAN]** Jangan mencatat isi body HTML penuh, `API_SECRET_KEY`, atau header otorisasi.

### 14.2 Metrik di `/health` & `/status`
Umur data tersegar tiap `kind`, jumlah item, status siklus terakhir, status breaker.

### 14.3 Notifikasi Admin
- Dikirim ke nomor di `ADMIN_WHITELIST_NUMBERS` lewat **InteractiveQueue** (1 pesan ringkas, bukan blast).
- Kondisi: (a) siklus `blocked`/`layout_changed`/`anomaly_*` ≥ `SCRAPER_NOTIFY_AFTER_FAILURES` kali berturut-turut; (b) data tersegar lebih tua dari `SCRAPER_MAX_STALE_HOURS`; (c) breaker terbuka.
- **Anti-spam notifikasi:** maksimal 1 pesan per jenis masalah per 12 jam; kirim pesan "pulih" saat siklus kembali sukses.
- Isi pesan singkat, tanpa data sensitif. Contoh: *"⚠️ Scraper BPS Bangka Barat: diblokir situs 2× berturut-turut. Bot memakai data cadangan. Cek /api/v1/scraper/status."*

---

## 15. Matriks Penanganan Kegagalan

| Skenario | Deteksi | Perilaku Scraper | Dampak ke Warga |
|---|---|---|---|
| Situs memblokir (403/captcha) | §7.7 | Hentikan siklus, breaker terbuka, notifikasi admin | Bot pakai API/GSheet; tetap menjawab. |
| 429 Too Many Requests | status | Hentikan, hormati `Retry-After`, breaker | Sama. |
| Situs mati/timeout/5xx | error jaringan/status | Retry terbatas, lalu `upstream_unavailable` | Data lama tetap dipakai. |
| Layout berubah (selector gagal) | §8.7 | Tolak hasil, data lama dipertahankan, notifikasi | Data lama tampil + label "mungkin belum terbaru" jika lewat batas. |
| Satu halaman detail 404 | status | Lewati item | Tidak ada. |
| `robots.txt` melarang path | §7.1 | Path dilewati, dicatat | Jenis data itu tidak ter-update. |
| `robots.txt` tidak bisa diambil (5xx/timeout) | error | Anggap **terlarang sementara** (konservatif), coba lagi siklus depan. 404 pada robots.txt = tidak ada pembatasan. | – |
| DB terkunci/penuh | error SQLite | Retry `busy_timeout`; jika gagal siklus `error`, data lama utuh | – |
| Panic di parser | `recover` | Item/siklus ditandai gagal, bot tetap hidup | Tidak ada. |
| Server mati lama lalu menyala | umur data | Catch-up siklus saat startup (§10.2 butir 5) | Data lama tampil sementara. |
| Jam server salah | – | Cron memakai waktu server; wajib NTP aktif (cek di checklist deploy) | – |
| Disk penuh | error tulis | Siklus gagal, notifikasi admin | – |
| Item berisi teks aneh/injeksi | sanitasi §8.2, §11.1 | Dibersihkan; tidak pernah dieksekusi | Aman. |

---

## 16. Keamanan Ringkas (Selaras `secure.md`)

1. Scraper hanya **keluar** (*outbound*) ke satu host terdaftar; tidak membuka *port* baru.
2. Endpoint admin scraper terkunci `X-API-Key`, terikat `127.0.0.1`.
3. `scraper.db` izin `600`; tidak masuk Git.
4. Semua URL hasil parsing divalidasi anti-SSRF; unduhan media tetap via jalur ber-batas 16 MB.
5. Teks dari situs dianggap tidak tepercaya (sanitasi + parameter SQL terikat).
6. Tidak ada kredensial baru yang diperlukan untuk scraping.
7. Proses berjalan sebagai user non-root `botistimewa`.
8. Dependensi diperiksa `go mod verify` dan `govulncheck` sebelum rilis.

---

## 17. Rencana Pengujian

### 17.1 Unit & Golden Test (tanpa internet)
Semua uji **tidak boleh** menyentuh situs asli; gunakan `httptest.Server` dan berkas `testdata/`.

| Paket/Berkas | Yang Diuji |
|---|---|
| `dateparse_test.go` | ≥ 20 kasus tanggal ID/EN, NBSP, singkatan, tidak valid, fallback dari URL. |
| `parser_*_test.go` | Golden test: tiap fixture → JSON hasil yang diharapkan (`testdata/*.golden.json`). Termasuk kasus field kosong dan item tanpa tanggal. |
| `validate_test.go` | Penolakan item tak valid; panjang judul; URL privat; skema `http`. |
| `fetcher_test.go` | 200 OK; 304; 403; 429 + `Retry-After`; 500 lalu sukses (retry); redirect keluar host ditolak; body > batas; content-type non-HTML; body "Just a moment" status 200 → `ErrBlocked`. |
| `robots_test.go` | Path boleh/dilarang; cache; 404 = bebas; 5xx = konservatif. |
| `breaker_test.go` | Eskalasi cooldown 1→2→4→8→24 jam; reset; persistensi setelah restart. |
| `store_test.go` | Upsert (baru/sama/berubah); `missed_runs` → nonaktif pada ke-3; siklus gagal tidak mengubah apa pun; transaksi *rollback*. |
| `validate/anomaly_test.go` | 0 item, penurunan >50%, kehilangan field >30%, marker hilang → hasil ditolak, data lama utuh. |
| `scheduler_test.go` | Spec cron valid/tidak valid; tanpa tumpang tindih; pemicu manual saat berjalan → 409; catch-up saat DB kosong; shutdown bersih. (Gunakan *fake clock*/spec `@every 1s` untuk uji.) |
| `provider_test.go` | Rantai fallback §4.2 untuk semua kombinasi (API ok/gagal × scraper segar/stale/kosong × GSheet ok/gagal); `DATA_SOURCE_ORDER` diubah. |
| `internal/api/server_test.go` | Auth wajib di 5 endpoint scraper; `run` rate-limit manual; format JSON. |
| `internal/bot/handlers_*_test.go` | Menu 3/5/1.5 dengan provider palsu: data segar, stale, kosong, error → pesan sesuai tanpa panic. |

### 17.2 Uji Integrasi Lokal
1. Jalankan server palsu (`httptest`) yang meniru situs BPS dari fixture; arahkan `SCRAPER_BASE_URL` ke sana (uji-only; validator mengizinkan `127.0.0.1` **hanya** saat `APP_ENV=test`).
2. Jalankan 3 siklus berturut-turut: siklus 1 = backfill, siklus 2 = tidak ada perubahan (semua `304`/hash sama), siklus 3 = satu item baru. Pastikan `items_new` = 48, 0, 1.
3. Suntik kegagalan: server palsu mengembalikan 403 → pastikan breaker terbuka dan bot tetap melayani dari DB lama.

### 17.3 Uji Beban Ringan
- 200 pesan pencarian dalam 10 detik dari beberapa nomor tester → latensi p95 < 1,5 detik, **0 request keluar ke situs BPS** (bukti bahwa bot tidak melakukan live-scrape).
- `go test -race ./...` harus bersih (scheduler, store, provider diakses konkuren).

### 17.4 Uji Pra-Produksi di Server Nyata (Manual)
1. Pastikan IP server sudah di-*allowlist* (jika diperlukan, §3.3).
2. Jalankan `POST /api/v1/scraper/run?job=pressrelease` sekali; periksa `runs` dan `items`.
3. Bandingkan 5 item acak dengan situs asli (judul, tanggal, tautan).
4. Matikan jaringan 1 siklus → pastikan data lama tetap tampil dan tidak ada crash.
5. Biarkan 24 jam (4 siklus) → semua `success`, `requests` per siklus kecil (early-stop bekerja).

---

## 18. Rencana Implementasi Bertahap

| Fase | Pekerjaan | Keluaran | Kriteria Selesai |
|:---:|---|---|---|
| **0** | Kumpulkan fixture manual, konfirmasi A-01…A-06, minta persetujuan §3.2, ajukan *allowlist* IP, daftarkan `BPS_API_KEY` | `testdata/*`, `testdata/README.md`, `scraper_selectors.yaml` terisi | Semua asumsi terjawab; fixture lengkap. |
| **1** | Fondasi: `config`, `model`, `dateparse`, `store` + migrasi, `validate` | Paket kompilasi + tes hijau | `go test ./internal/scraper/...` lulus. |
| **2** | `fetcher`, `robots`, `breaker`, deteksi blokir | Fetcher lengkap dengan uji `httptest` | Semua skenario §17.1 fetcher/robots/breaker lulus. |
| **3** | Parser `pressrelease` + `runner` + anomaly guard (satu kind end-to-end) | Siklus penuh untuk BRS dari fixture | Golden test + integrasi 3 siklus lulus. |
| **4** | Parser `publication`, `infographic`, `statictable` | Tiga parser + golden test | Semua golden test lulus. |
| **5** | `scheduler` (cron 6 jam, jitter, catch-up, graceful shutdown) | Scheduler terpasang di `cmd/server/main.go` | `scheduler_test` lulus; log `scraper scheduler aktif`. |
| **6** | `provider` + rantai fallback + integrasi handler Menu 1/3/5/1.5 + kamus locale | Bot membaca dari DB | Uji handler §17.1 lulus; 0 request keluar saat melayani chat. |
| **7** | Endpoint `/api/v1/scraper/*`, `/health`, notifikasi admin | API & observabilitas | Uji API lulus; notifikasi terkirim di uji manual. |
| **8** | `news` (opsional), dokumentasi README, update `.env.example`, `PROGRESS.md` | Rilis v3.1.0 | Checklist §19 terpenuhi. |

---

## 19. Checklist Definition of Done

**Pra-implementasi**
- [ ] Persetujuan tertulis scraping situs sendiri dari pimpinan/IPDS (§3.2).
- [ ] Permohonan *allowlist* IP server diajukan (atau API key aktif sebagai jalur utama).
- [ ] Fixture HTML lengkap dan asumsi A-01…A-06 terjawab (Fase 0).
- [ ] `scraper_selectors.yaml` terisi, tidak ada `TBD` pada kind aktif.

**Kode & uji**
- [ ] Tidak ada dependensi CGO/headless browser.
- [ ] `go build ./...`, `go vet ./...`, `go test -race ./...` bersih.
- [ ] Golden test parser lulus untuk semua fixture.
- [ ] Uji blokir 403/429/"Just a moment" → breaker terbuka, bot tetap hidup.
- [ ] Anomaly guard menolak hasil kosong/turun drastis dan data lama tetap utuh.
- [ ] Upsert idempoten: siklus kedua tanpa perubahan → `items_new=0`, `items_updated=0`.
- [ ] Item hilang baru dinonaktifkan setelah 3 siklus sukses berturut-turut.

**Penjadwalan**
- [ ] Cron `0 */6 * * *` zona `Asia/Jakarta` terverifikasi di log (`next_run` benar: 00/06/12/18 WIB).
- [ ] Tidak ada tumpang tindih; pemicu manual saat berjalan → `409`.
- [ ] Catch-up saat startup berjalan setelah server mati > 6 jam.
- [ ] Shutdown bersih (tidak ada goroutine/DB menggantung).
- [ ] Hanya satu pemicu aktif (internal **atau** cron OS), bukan keduanya.

**Kesopanan & keamanan**
- [ ] `robots.txt` dipatuhi (`SCRAPER_RESPECT_ROBOTS=true`).
- [ ] *User-Agent* jujur dengan kontak `bps1903@bps.go.id`.
- [ ] Jeda ≥ 3 detik, 1 koneksi, ≤ 60 request/siklus terbukti di log.
- [ ] Anti-SSRF berlaku untuk semua URL hasil parsing; redirect keluar host ditolak.
- [ ] `scraper.db` izin `600`, tidak masuk Git.
- [ ] Endpoint `/api/v1/scraper/*` terkunci `X-API-Key`.

**Integrasi bot**
- [ ] Bot tidak pernah memanggil situs BPS saat melayani chat (0 request keluar saat uji beban).
- [ ] Rantai fallback API → Scraper → GSheet → tautan portal teruji.
- [ ] Label sumber & kesegaran muncul di setiap daftar; label *stale* muncul bila melewati 48 jam.
- [ ] Notifikasi admin terkirim saat gagal berulang dan tidak spam.

**Dokumentasi**
- [ ] `.env.example`, `README.md`, dan `PROGRESS.md` diperbarui.
- [ ] Salinan dokumen ini diarsipkan di `prd/scraping.md`.

---

## 20. Risiko & Mitigasi

| # | Risiko | Kemungkinan | Dampak | Mitigasi |
|---|---|:---:|:---:|---|
| R-01 | Situs terus memblokir scraper (F-01) | **Tinggi** | Tinggi | *Allowlist* IP; aktifkan BPS Web API sebagai jalur utama; fallback GSheet; breaker agar tidak memperburuk. |
| R-02 | Layout/HTML situs berubah | Sedang | Sedang | Selector di YAML; anomaly guard; golden test; notifikasi admin. |
| R-03 | Halaman dirender JavaScript (A-01 salah) | Rendah–Sedang | Tinggi | Fase 0 memverifikasi lebih dulu; jika benar, jangan paksakan scraping: pakai API resmi atau cari endpoint JSON publik yang sah. |
| R-04 | Pelanggaran ketentuan layanan | Rendah | Tinggi | Persetujuan tertulis §3.2; patuhi `robots.txt`; laju rendah; atribusi. |
| R-05 | Data salah tampil ke warga | Rendah | Tinggi | Validasi ketat; anomaly guard; tautan sumber di tiap pesan; uji banding manual. |
| R-06 | Korupsi/kehilangan data akibat siklus parsial | Rendah | Sedang | Transaksi per kind; K-07; tidak ada `DELETE` otomatis. |
| R-07 | Beban server kecil (256 MB RAM) terganggu | Rendah | Rendah | Batas body 3 MB, konkurensi 1, tanpa headless. |
| R-08 | Dua pemicu jadwal berjalan bersamaan | Rendah | Rendah | Mutex + satu pemicu saja (§10.6). |
| R-09 | Zona waktu/jam server salah | Rendah | Rendah | `time/tzdata`, NTP aktif, log `next_run` saat startup. |

---

## 21. Pertanyaan Terbuka (Perlu Keputusan Tim)

| # | Pertanyaan | Default Dokumen Ini |
|---|---|---|
| Q-01 | Apakah scraping situs sendiri diizinkan secara tertulis & IP server bisa di-*allowlist*? | Wajib dijawab sebelum go-live. |
| Q-02 | Jenis data mana yang paling prioritas jika waktu terbatas? | `pressrelease` → `publication` → `infographic` → `statictable`. |
| Q-03 | Apakah API atau scraper yang diprioritaskan? | API dulu (K-02), bisa dibalik lewat `DATA_SOURCE_ORDER`. |
| Q-04 | Apakah perlu data berbahasa Inggris (`/en/...`)? | Hanya `id`; kolom `lang` sudah disiapkan. |
| Q-05 | Apakah `news` (berita kegiatan) perlu masuk bot? | Opsional, Fase 8. |
| Q-06 | Siapa penerima notifikasi admin? | Nomor di `ADMIN_WHITELIST_NUMBERS`. |

---

## 22. Catatan Konsistensi dengan Dokumen Sebelumnya

Saat menyusun dokumen ini ditemukan beberapa ketidaksamaan antar-dokumen yang sebaiknya dirapikan sebelum implementasi:

| # | Temuan | Rekomendasi |
|---|---|---|
| C-01 | `secure.md` mencatat `IDLE_TIMEOUT_MINUTES=15`, sedangkan `PROGRESS.md` §2.2 menyebut timer *idle* **30 menit**. | Tetapkan satu nilai dan samakan di kedua dokumen serta `.env`. |
| C-02 | `secure.md` memakai `WA_DB_PATH=./data/session.db`, sedangkan `PROGRESS.md` §5 menunjukkan `storage/whatsmeow.db`. | Samakan path; izin `600` pada berkas yang benar. |
| C-03 | `api_and_routegroup.md` mencantumkan *User-Agent* browser untuk client BPS API; dokumen ini mensyaratkan *User-Agent* jujur untuk scraper. | Tinjau ulang untuk client API (§3.3). |
| C-04 | `secure.md` memuat nomor WhatsApp admin asli pada contoh `.env.example`. | Ganti dengan placeholder (`628xxxxxxxxxx`) di dokumen yang dibagikan/di-*commit*. |
| C-05 | `secure.md` pada contoh `GAS_URL_*` berisi sintaks Markdown (`[https://...](https://...)`) di dalam blok `.env`. | Gunakan URL polos agar tidak ter-parse salah oleh `godotenv`. |

---

## Lampiran A — Contoh Kerangka `Runner` Aman Panic

```go
func (r *Runner) RunAll(ctx context.Context, trigger Trigger) {
	if !r.mu.TryLock() {
		r.store.LogSkipped(trigger, "skipped_overlap")
		return
	}
	defer r.mu.Unlock()

	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("scraper panic", "panic", rec, "stack", string(debug.Stack()))
			r.store.FailActiveRun(fmt.Sprint(rec))
		}
	}()

	if open, until := r.breaker.IsOpen(); open {
		slog.Warn("scrape.skipped_breaker", "open_until", until)
		r.store.LogSkipped(trigger, "skipped_breaker")
		return
	}

	ctx, cancel := context.WithTimeout(ctx, r.cfg.RunTimeout)
	defer cancel()

	run := r.store.StartRun(trigger, "all")
	// ... iterasi kind sesuai §10.4
}
```

## Lampiran B — Contoh Kerangka `looksBlocked`

```go
var blockMarkers = []string{
	"captcha", "just a moment", "attention required", "access denied",
	"checking your browser", "cf-chl", "request blocked", "bot detection",
}

func looksBlocked(body []byte) bool {
	if len(body) < 2048 {
		return true // halaman sah selalu jauh lebih besar
	}
	head := strings.ToLower(string(body[:min(len(body), 16*1024)]))
	for _, m := range blockMarkers {
		if strings.Contains(head, m) {
			return true
		}
	}
	return false
}
```
> Daftar `blockMarkers` dan ambang 2 KB diverifikasi terhadap fixture `blocked_403.html` dan halaman sah di Fase 0, supaya tidak ada *false positive* (mis. kata "captcha" yang muncul wajar di halaman sah). Jika ada, persempit ke area `<title>`/`<body>` awal.

## Lampiran C — Contoh Normalisasi Teks

```go
var spaceRe = regexp.MustCompile(`[\s\x{00a0}]+`)

func cleanText(s string) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, s)
	return strings.TrimSpace(spaceRe.ReplaceAllString(s, " "))
}
```

## Lampiran D — Contoh Respons `GET /api/v1/scraper/runs`

```json
{
  "success": true,
  "data": [
    {
      "id": 128,
      "trigger": "cron",
      "kind": "all",
      "status": "success",
      "started_at": "2026-10-27T12:02:11+07:00",
      "finished_at": "2026-10-27T12:03:02+07:00",
      "requests": 6,
      "items_found": 63,
      "items_new": 1,
      "items_updated": 0,
      "items_skipped": 0
    }
  ]
}
```

---

*Akhir dokumen. Setiap perubahan pada pola URL, selector, atau kebijakan situs target wajib dicatat di `internal/scraper/testdata/README.md` dan, bila mengubah perilaku, di dokumen ini.*