# Product Requirements & Route Group Architecture — BOT ISTIMEWA v3.0
## Integrasi BPS Web API v1 (`https://webapi.bps.go.id/v1/api/list`) & Arsitektur Route Group Go Fiber

| Field | Detail |
|---|---|
| **Produk** | ISTIMEWA (Informasi Statistik Terpadu melalui WhatsApp) |
| **Instansi** | BPS Kabupaten Bangka Barat |
| **Arsitektur Target** | Go (`whatsmeow`) + Go Fiber v2 + BPS Web API v1 + In-Memory Cache |
| **Dokumen** | API & Route Group Architecture Specification (`api_and_routegroup.md`) |
| **Base API** | `https://webapi.bps.go.id/v1/api/list` |
| **Versi** | 3.0.0 |
| **Tanggal** | 20 Oktober 2026 |
| **Status** | Implemented & Verified |

---

## 1. Ringkasan Eksekutif & Latar Belakang

### 1.1 Latar Belakang
Pada arsitektur awal BOT ISTIMEWA, data statistik publikasi, indikator, dan infografis ditarik secara tidak langsung melalui Google Apps Script (GAS) yang terhubung ke Google Sheets. Meskipun handal untuk kurasi internal, pendekatan ini memiliki keterbatasan dalam menyajikan data rilis statistik terkini (seperti Berita Resmi Statistik/BRS, katalog publikasi lengkap, daftar subjek, dan variabel dinamis) yang dikelola langsung di server pusat Badan Pusat Statistik (BPS RI).

Untuk meningkatkan kualitas layanan statistik di Kabupaten Bangka Barat, BOT ISTIMEWA v3.0 mengintegrasikan **BPS Web API v1** (`https://webapi.bps.go.id/v1/api/list`) secara *native* ke dalam ekosistem Go, dibungkus dalam arsitektur **Route Group Go Fiber** yang aman, cepat, dan dilengkapi dengan *in-memory caching*.

### 1.2 Tujuan Arsitektur
1. **Integrasi Native BPS Web API**: Menyediakan *client package* Go (`internal/bpsapi`) untuk berkomunikasi langsung dengan `https://webapi.bps.go.id/v1/api/list`.
2. **Modular Fiber Route Group**: Menyediakan grup rute terstruktur `/api/v1/bps/*` di Go Fiber yang memetakan model-model data BPS.
3. **Resilience & Caching**: Menerapkan *in-memory caching* (TTL 15 menit) dengan mekanisme *stale-if-error fallback* untuk mengantisipasi downtime atau pemblokiran WAF pada server BPS pusat.
4. **Security & Anti-SSRF**: Memvalidasi URL tujuan, mengamankan endpoint API dengan middleware `X-API-Key`, dan mencegah paparan kredensial `BPS_API_KEY` ke publik.

---

## 2. Spesifikasi Endpoint BPS Web API v1 (`list`)

### 2.1 Pola URL Hierarkis
Endpoint `list` BPS Web API menggunakan format URL berbasis segmen path:
```
https://webapi.bps.go.id/v1/api/list/model/{model}/lang/{lang}/domain/{domain}/key/{key}/[optional_params...]
```

| Parameter Segmen | Tipe | Keterangan | Nilai Default |
|---|---|---|---|
| `model` | String | Model data statistik yang diminta | `subject`, `publication`, dll. |
| `lang` | String | Bahasa output (`ind` untuk Indonesia, `eng` untuk Inggris) | `ind` |
| `domain` | String | Kode wilayah BPS (4 digit kab/kota, 2 digit provinsi, `0000` nasional) | `1903` (Bangka Barat) |
| `key` | String | Kunci API resmi dari portal developer BPS | Diambil dari `BPS_API_KEY` (.env) |

### 2.2 Model Data yang Didukung
1. **`subject`**: Mengambil daftar subjek/topik statistik (Sosial, Ekonomi, Pertanian).
   - *Optional:* `/subcat/{id}`, `/page/{page}`
2. **`subcat`**: Mengambil daftar kategori subjek statistik.
3. **`var`**: Mengambil daftar variabel statistik dalam suatu subjek.
   - *Optional:* `/subject/{subject_id}`, `/page/{page}`
4. **`statictable`**: Mengambil daftar tabel statistik statis.
   - *Optional:* `/subject/{id}`, `/keyword/{keyword}`, `/month/{m}`, `/year/{y}`, `/page/{page}`
5. **`publication`**: Mengambil katalog buku publikasi statistik BPS.
   - *Optional:* `/keyword/{keyword}`, `/month/{m}`, `/year/{y}`, `/page/{page}`
6. **`pressrelease`**: Mengambil Berita Resmi Statistik (BRS) beserta tautan PDF dan slide.
   - *Optional:* `/keyword/{keyword}`, `/month/{m}`, `/year/{y}`, `/page/{page}`
7. **`infographic`**: Mengambil katalog gambar infografis data statistik.
   - *Optional:* `/keyword/{keyword}`, `/page/{page}`
8. **`data`**: Mengambil data dinamis nilai variabel statistik.
   - *Optional:* `/var/{var_id}`, `/turvar/{id}`, `/vervar/{id}`, `/th/{year_id}`, `/turth/{id}`

### 2.3 Standar Format Respons JSON BPS
BPS Web API v1 mengembalikan format khusus di mana array `data` berisi 2 elemen:
- **Elemen 0**: Metadata pagination (`page`, `pages`, `per_page`, `total`, `count`).
- **Elemen 1**: Array objek data sesuai model yang diminta.

Contoh payload respons sukses:
```json
{
  "status": "OK",
  "data-availability": "available",
  "data": [
    {
      "page": 1,
      "pages": 4,
      "per_page": 10,
      "total": 35
    },
    [
      {
        "sub_id": 1,
        "title": "Kependudukan dan Migrasi",
        "subcat_id": 1,
        "subcat": "Sosial dan Kependudukan",
        "ntabel": 24
      }
    ]
  ]
}
```

---

## 3. Desain Paket Go: `internal/bpsapi`

Struktur paket modular untuk komunikasi dengan BPS Web API:

```
internal/bpsapi/
├── models.go       # Struct Go typed (Subject, Publication, PressRelease, dll.)
├── client.go       # HTTP Client, URL Builder, Cache In-Memory, Stale Fallback
└── client_test.go  # Unit test URL builder, parser, cache, & error handling
```

### 3.1 Fitur Utama Client BPS
1. **Safe URL Builder**: Merakit segmen URL dengan aman dan melakukan *path encoding* terhadap karakter khusus (seperti spasi pada *keyword*).
2. **Bypass WAF & Resilient HTTP**: Menggunakan `User-Agent` browser modern dan `Timeout` 12 detik agar tidak terblokir oleh *Perimeter WAF*.
3. **Dual-Layer Cache (TTL 15 Menit)**:
   - Cache hit: Langsung mengembalikan data dalam hitungan mikrodetik tanpa request HTTP ke server BPS.
   - Stale-If-Error: Jika server BPS mengalami timeout/error 5xx, client menyajikan data cache terakhir (*graceful degradation*).
4. **Generic JSON Unmarshaler**: Mengurai struktur ganda `[Pagination, []Items]` secara otomatis dan aman dari *type assertion panic*.

---

## 4. Arsitektur Go Fiber Route Group (`/api/v1/bps`)

Semua endpoint BPS Web API diekspos melalui route group `/api/v1/bps` di dalam server Go Fiber (`internal/api/server.go`), dilindungi oleh `authMiddleware` menggunakan token `X-API-Key` atau `Authorization: Bearer`.

```
/api
  └── /v1
        └── /bps (Route Group)
              ├── GET  /list            # Generic proxy builder (model, page, dll.)
              ├── GET  /subjects        # Daftar subjek statistik
              ├── GET  /subcats         # Kategori subjek
              ├── GET  /variables       # Variabel statistik (filter ?subject=...)
              ├── GET  /statictables    # Tabel statis (filter ?subject=..., ?keyword=...)
              ├── GET  /publications    # Publikasi BPS (filter ?year=..., ?keyword=...)
              ├── GET  /pressreleases   # Berita Resmi Statistik (BRS)
              ├── GET  /infographics    # Infografis BPS
              ├── GET  /data            # Nilai data variabel dinamis
              └── POST /cache/clear     # Invalidasi in-memory cache BPS
```

### 4.1 Parameter Query yang Didukung
Setiap endpoint menerima parameter opsional via query string:
- `domain`: Menimpa domain default (misal: `1900` untuk Provinsi Babel, `0000` untuk Nasional).
- `lang`: Bahasa (`ind` / `eng`).
- `page`: Nomor halaman (default: `1`).
- `keyword`: Kata kunci pencarian judul.
- `year`: Filter tahun.
- `subject`: ID subjek untuk filter variabel atau tabel statis.

### 4.2 Format Respons Standar Route Group
Semua rute di bawah `/api/v1/bps` mengembalikan respons JSON standar:
```json
{
  "success": true,
  "model": "publication",
  "domain": "1903",
  "lang": "ind",
  "pagination": {
    "page": 1,
    "pages": 5,
    "per_page": 10,
    "total": 42
  },
  "data": [ ... ]
}
```

---

## 5. Konfigurasi Environment (`.env`)

Variabel baru yang ditambahkan ke `.env` dan `.env.example`:

```env
# ==========================================
# 7. KONFIGURASI WEBAPI BPS v1
# ==========================================
BPS_API_BASE_URL=https://webapi.bps.go.id/v1/api/list
BPS_API_KEY=ganti_dengan_api_key_resmi_bps
BPS_DEFAULT_DOMAIN=1903
BPS_DEFAULT_LANG=ind
BPS_CACHE_TTL_MINUTES=15
```

---

## 6. Checklist Verifikasi & Pengujian

- [x] Struct models untuk seluruh objek BPS (`Subject`, `Publication`, `PressRelease`, dll.).
- [x] URL builder menangani seluruh kombinasi parameter query.
- [x] Parser berhasil mengurai struktur 2-elemen `data: [Pagination, []Items]`.
- [x] In-memory cache berjalan dengan TTL 15 menit dan thread-safe (`sync.RWMutex`).
- [x] Stale fallback aktif saat request HTTP gagal.
- [x] Route group Go Fiber `/api/v1/bps/*` terintegrasi dan terlindungi `X-API-Key`.
- [x] Endpoint cache invalidation `/api/v1/bps/cache/clear` berfungsi.
- [x] Seluruh unit test (`go test ./...`) lulus 100%.