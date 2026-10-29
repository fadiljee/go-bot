# BOT ISTIMEWA v3.3 (Go Native, Scraper & Strategic Indicators API Edition)
**Chatbot WhatsApp & REST API Resmi Badan Pusat Statistik (BPS) Kabupaten Bangka Barat**

Migrasi penuh ke arsitektur native **Golang** menggunakan `go.mau.fi/whatsmeow`, `github.com/gofiber/fiber/v2`, dan mesin internal **Web Scraper BPS Bangka Barat** berbasis SQLite. Bot ini beroperasi **100% mandiri tanpa ketergantungan pada Google Apps Script (GAS / Google Sheets)** serta menyediakan layanan interaktif WhatsApp dan **REST API Indikator Strategis**.

---

## Fitur Utama & Keunggulan

1. **Mandiri Tanpa Ketergantungan Eksternal (Zero-GAS Dependency)**:
   - Seluruh data publikasi, Berita Resmi Statistik (BRS), berita kegiatan, dan indikator strategis dilayani langsung oleh **mesin scraper internal** dan database SQLite lokal ([`storage/scraper.db`](storage/scraper.db)).
   - Tidak lagi bergantung pada kuota atau latensi Google Apps Script / Google Sheets.

2. **REST API Indikator Strategis Daerah (Menu 2 Terintegrasi)**:
   - Menyediakan endpoint REST API terstruktur untuk 7 Indikator Strategis resmi BPS Bangka Barat: Kependudukan, Pertanian, Kemiskinan, Ketenagakerjaan, Pertumbuhan Ekonomi (PDRB), Harga & Inflasi terkini, serta Kesejahteraan & IPM.
   - Dilengkapi parameter format dinamis (`?format=json` atau `?format=text` siap broadcast) dan filter kategori.
   - Terintegrasi dua arah: data di-scrape berkala, di-cache di memori, dan disajikan secara real-time ke WhatsApp Chatbot dan aplikasi pihak ketiga / dashboard.

3. **Mesin Scraper Terintegrasi (Next.js Server Actions Extractor)**:
   - Ditulis murni dalam Go (tanpa headless browser / tanpa Chrome / tanpa Puppeteer).
   - Menarik data langsung dari portal resmi [bangkabaratkab.bps.go.id](https://bangkabaratkab.bps.go.id/):
     - **34+ Berita Resmi Statistik (BRS)** beserta thumbnail dan berkas PDF.
     - **50+ Publikasi Statistik** beserta cover buku resolusi tinggi dan berkas PDF.
     - **10+ Berita Kegiatan** beserta foto dokumentasi lapangan.
     - Tabel statistik statis dan data indikator strategis daerah.
   - Penjadwalan otomatis via Cron (`0 */4 * * *`) setiap 4 jam dengan automatic cache invalidation, *circuit breaker*, *rate limiting*, dan *jitter delay*.

4. **Penyajian Media Interaktif di WhatsApp**:
   - **Kirim Cover Gambar Otomatis**: Bot langsung mengirimkan foto cover publikasi atau thumbnail rilis BRS ke chat WhatsApp pengguna (`SendImageFromURL`) saat memilih rincian data.
   - **Video In-App Player**: Video literasi statistik menggunakan tautan YouTube resmi BPS yang dapat diputar langsung di dalam aplikasi WhatsApp (*Picture-in-Picture*) tanpa membebani storage server (0 MB local video storage).

5. **Reservasi Konsultasi PST & Notifikasi Grup Internal Kantor**:
   - **Alur Pengajuan Jadwal Interaktif (Menu 6)**: Pemohon (masyarakat/OPD) dapat mengajukan jadwal konsultasi statistik dan kebutuhan data secara langsung melalui bot.
   - **Penerbitan Nomor Tiket Resmi**: Setiap pengajuan diterbitkan nomor tiket resmi (contoh: `#PST-261028-561`) dan disimpan ke tabel `consultation_bookings` di SQLite.
   - **Auto-Forward ke Grup WA Kantor**: Bot secara otomatis meneruskan rincian pemohon ke grup WhatsApp internal kantor (`INTERNAL_PST_GROUP_JID`) agar petugas piket PST dan pakar terkait dapat segera menindaklanjuti.
   - **Auto-Discovery Group JID**: Bot secara otomatis mendeteksi dan menampilkan daftar Group JID yang diikuti di terminal saat terhubung, memudahkan konfigurasi pada lingkungan lokal maupun produksi.

6. **Efisiensi Memori & Keamanan Maksimal**:
   - Penggunaan RAM sangat rendah (**< 35 MB** saat aktif).
   - Proteksi banjir pesan (*Token-Bucket Rate Limiting* per nomor).
   - Sanitasi input anti-ReDoS (maksimal 100 karakter).
   - Autentikasi Admin & API terisolasi dengan proteksi token dan *brute-force lockout*.

---

## Struktur Direktori

```text
go-bot/
├── cmd/
│   ├── server/
│   │   └── main.go                 # Entrypoint server bot WhatsApp & REST API
│   ├── scraper-cli/
│   │   └── main.go                 # Alat CLI untuk scraping manual & cek data
│   └── scraper-seed/
│       └── main.go                 # Seeder database scraper
├── internal/
│   ├── api/
│   │   ├── server.go               # REST API Fiber v2, Web QR scanner (/qr) & endpoint indikator
│   │   ├── indicators_test.go      # Unit test REST API indikator strategis
│   │   └── handlers_scraper.go     # Endpoint API status data scraper (/api/v1/scraper)
│   ├── bot/
│   │   ├── router.go               # Table-driven state machine dispatcher
│   │   ├── session.go              # Manajemen sesi, history stack, & idle timer
│   │   ├── handlers_main.go        # Welcome full/short, reset, admin, timeout SKD
│   │   ├── handlers_layanan.go     # Menu 1: Layanan Data & pencarian SQLite
│   │   ├── handlers_indikator.go   # Menu 2: Indikator Strategis dinamis dari database scraper
│   │   ├── handlers_media.go       # Menu 3 (BRS/Publikasi), 4 (Video), 5 (Infografik)
│   │   └── handlers_support.go     # Menu 6 (Konsultasi/Booking), 7 (Langganan), 8, 9, 10 (FAQ)
│   ├── scraper/
│   │   ├── indicators.go           # Engine kurasi & metadata 7 Indikator Strategis
│   │   ├── indicators_test.go      # Unit test provider indikator strategis
│   │   ├── runner.go               # Engine scraper Next.js Server Actions
│   │   ├── fetcher.go              # HTTP client dengan cookie jar & anti-bot bypass
│   │   ├── store.go                # Penyimpanan SQLite (scraped_items, bookings, runs)
│   │   ├── provider.go             # Data provider & in-memory caching untuk router bot
│   │   ├── scheduler.go            # Penjadwal otomatis scraping (Cron)
│   │   └── breaker.go              # Proteksi Circuit Breaker kegagalan scraper
│   ├── config/
│   │   └── config.go               # Parser konfigurasi & environment variable
│   ├── locale/
│   │   └── locale.go               # Sistem lokalisasi bahasa Indonesia (id_ID)
│   └── whatsapp/
│       ├── client.go               # whatsmeow client wrapper, sqlstore, session manager
│       ├── interactive.go          # Protobuf builder untuk Single Select List & Buttons
│       ├── parser.go               # Parser pesan masuk, unwrap list & native flow
│       └── sender.go               # Pengirim pesan teks, gambar via URL, stiker, buttons
├── configs/
│   └── scraper_selectors.yaml      # Konfigurasi selector & URL target scraper
├── storage/
│   ├── whatsmeow.db                # Database SQLite sesi koneksi WhatsApp
│   └── scraper.db                  # Database SQLite data hasil scraping & permohonan PST
├── .env.example                    # Template konfigurasi environment
├── .env                            # Konfigurasi aktif bot
├── bot-istimewa                    # Binary server utama hasil kompilasi
├── scraper-cli                     # Binary CLI scraper hasil kompilasi
├── go.mod
└── go.sum
```

---

## Panduan Instalasi & Menjalankan Bot

### 1. Prasyarat Sistem
- **Golang**: Versi 1.21 ke atas (direkomendasikan Go 1.22+).
- **Sistem Operasi**: Linux / macOS / Windows.

### 2. Konfigurasi Environment (`.env`)
Salin file template `.env.example` menjadi `.env`:
```bash
cp .env.example .env
```

Buka dan sesuaikan variabel di dalam `.env`:
```env
# Server & Port
APP_PORT=3000
API_SECRET_KEY=rahasia_internal_bps_1903

# Keamanan Admin
ADMIN_TRIGGER_KEYWORD="login admin - 1903"
ADMIN_PASSWORD=ganti_password_admin_yang_kuat
ADMIN_WHITELIST_NUMBERS=

# Notifikasi Grup Kantor (PST)
# Masukkan ID grup berakhiran @g.us
INTERNAL_PST_GROUP_JID=120xxxxxxxxxxxxxxxx@g.us

# Konfigurasi Scraper BPS Bangka Barat
SCRAPER_ENABLED=true
SCRAPER_BASE_URL=https://bangkabaratkab.bps.go.id
SCRAPER_CRON_SPEC="0 */4 * * *"
SCRAPER_ENABLED_KINDS=pressrelease,publication,infographic,statictable,news
SCRAPER_DB_PATH=./storage/scraper.db
```

### 3. Kompilasi & Jalankan Binary
Kompilasi server utama:
```bash
go build -o bot-istimewa ./cmd/server
```

Jalankan bot:
```bash
./bot-istimewa
```

*(Opsional) Kompilasi CLI Scraper untuk eksekusi manual:*
```bash
go build -o scraper-cli ./cmd/scraper-cli
./scraper-cli -kind all
```

### 4. Pindai QR Code WhatsApp
Saat bot pertama kali dijalankan:
1. **Lewat Terminal**: QR code ASCII akan langsung muncul di terminal.
2. **Lewat Browser**: Buka alamat web lokal:
   ```
   http://localhost:3000/qr
   ```
3. Buka WhatsApp di ponsel bot &rarr; **Perangkat Tertaut** &rarr; **Tautkan Perangkat** &rarr; Pindai QR code yang tampil.
4. Sesi akan otomatis tersimpan di [`storage/whatsmeow.db`](storage/whatsmeow.db), sehingga bot tidak perlu pindai ulang saat di-restart.

---

## Dokumentasi REST API Indikator Strategis (Menu 2)

Bot Istimewa menyediakan endpoint REST API untuk integrasi aplikasi eksternal, dashboard pimpinan daerah, atau portal web OPD. Seluruh endpoint dilindungi oleh `API_SECRET_KEY`.

### 1. Daftar Endpoint

| Method | Endpoint | Deskripsi |
| :--- | :--- | :--- |
| `GET` | `/api/v1/indicators` | Mengambil seluruh 7 indikator strategis beserta metrik dan link unduh dokumen resmi. |
| `GET` | `/api/v1/indicators/:id` | Mengambil 1 indikator berdasarkan ID (`1`-`7`) atau slug (`inflasi`, `kemiskinan`, `kependudukan`, dll). |
| `GET` | `/api/v1/indikator` | Alias berbahasa Indonesia untuk endpoint indikator. |
| `GET` | `/api/v1/indikator/:id` | Alias berbahasa Indonesia per ID/slug. |

### 2. Autentikasi
Gunakan salah satu metode autentikasi berikut:
- **Header**: `X-API-Key: rahasia_internal_bps_1903`
- **Bearer**: `Authorization: Bearer rahasia_internal_bps_1903`
- **Query Parameter**: `?key=rahasia_internal_bps_1903`

### 3. Parameter Tambahan (Query Parameters)
- `format=text`: Mengembalikan format teks pesan terstruktur yang siap dikirimkan via chat/WhatsApp atau dicetak di terminal.
- `category=<kata_kunci>`: Memfilter daftar indikator berdasarkan nama kategori atau slug.

### 4. Contoh Penggunaan (`curl`)

**A. Mengambil Seluruh Indikator (Format JSON):**
```bash
curl -H "X-API-Key: rahasia_internal_bps_1903" \
  http://localhost:3000/api/v1/indicators
```

**B. Mengambil Indikator Inflasi Terkini (Format Teks WhatsApp):**
```bash
curl -H "X-API-Key: rahasia_internal_bps_1903" \
  "http://localhost:3000/api/v1/indicators/inflasi?format=text"
```

*Contoh Output Plaintext:*
```text
*Indikator Harga & Inflasi*

• *Inflasi Year-on-Year (y-on-y)*: 4,49% (01 October 2026)
• *Indeks Harga Konsumen (IHK)*: 106,30
• *Komoditas Pendorong Inflasi*: Kelompok Makanan, Minuman, Tembakau, dan Perawatan Pribadi

*Sumber*: Berita Resmi Statistik (BRS) BPS Kabupaten Bangka Barat
*Dokumen Resmi*: September 2026, Inflasi Year on Year (y-on-y) Kabupaten Bangka Barat sebesar 4,49 persen.
*Unduh PDF*: https://webapi.bps.go.id/download.php?f=...
```

---

## Cara Mengetahui Group JID WhatsApp

Untuk mengisi `INTERNAL_PST_GROUP_JID`:
1. Masukkan nomor WhatsApp bot ke dalam grup kantor yang diinginkan.
2. Jalankan bot (`./bot-istimewa`). Saat terhubung, bot akan **otomatis mendeteksi dan mencetak seluruh nama grup beserta JID-nya** di terminal:
   ```text
    [DAFTAR GRUP WA BOT (2 grup)]:
      1. Nama: 'PST Internal BPS Babar'  JID: 120xxxxxxxxxxxxxxx@g.us
   ```
3. Atau, ketik pesan apa saja di grup tersebut, terminal akan mencetak:
   ```text
    [INFO GRUP WA] Pesan masuk di Grup JID: 120xxxxxxxxxxxxxxxx@g.us
   ```
4. Salin string JID tersebut ke baris `INTERNAL_PST_GROUP_JID=` di file `.env`, lalu restart bot.

---

## Pengujian (Unit Tests)

Jalankan seluruh test suite otomatis:
```bash
go test -v ./...
```

Uji spesifik REST API Indikator Strategis:
```bash
go test -v -run TestAPIIndicators ./internal/api
```

Uji spesifik provider & parser metrik indikator:
```bash
go test -v -run TestProvider_Indicators ./internal/scraper
```

Uji spesifik alur reservasi konsultasi & notifikasi grup:
```bash
go test -v -run TestConsultationBookingWorkflow ./internal/bot
```

---

## Standar Keamanan & Perlindungan
- **Anti-Flooding**: Dilengkapi *Token Bucket Rate Limiter* (maksimal 5 pesan per 10 detik per pengguna).
- **Anti-Brute Force Admin**: Maksimal 3 kali percobaan salah sebelum nomor dikunci (*lockout*) selama 60 menit.
- **Anti-SSRF**: Validasi skema HTTPS dan pemblokiran akses ke alamat loopback/privat internal.
- **Data Privacy**: Kerahasiaan nomor dan data pengguna dilindungi sesuai amanat UU No. 16 Tahun 1997 tentang Statistik.

---
**BPS Kabupaten Bangka Barat**  
*Statistik Berkualitas untuk Indonesia Maju*
