# BOT ISTIMEWA 
**Chatbot WhatsApp & REST API Resmi Badan Pusat Statistik (BPS) Kabupaten Bangka Barat**

Migrasi penuh ke arsitektur native **Golang** menggunakan `go.mau.fi/whatsmeow`, `github.com/gofiber/fiber/v2`, dan mesin internal **Web Scraper BPS Bangka Barat** berbasis SQLite. Bot ini beroperasi **100% mandiri tanpa ketergantungan pada Google Apps Script (GAS / Google Sheets)** serta menyediakan layanan interaktif WhatsApp, **REST API Indikator Strategis (Menu 2)**, dan **REST API Layanan Data Statistik (Menu 1)**.

---

## Fitur Utama & Keunggulan

1. **Mandiri Tanpa Ketergantungan Eksternal (Zero-GAS Dependency)**:
   - Seluruh data publikasi, Berita Resmi Statistik (BRS), berita kegiatan, tabel statistik statis, dan indikator strategis dilayani langsung oleh **mesin scraper internal** dan database SQLite lokal ([`storage/scraper.db`](file:///home/fadil/fadil/go-bot/storage/scraper.db)).
   - Tidak lagi bergantung pada kuota atau latensi Google Apps Script / Google Sheets.

2. **REST API Layanan Data Statistik (Menu 1 Terintegrasi)**:
   - Menyediakan katalog lengkap topik data BPS: **Sosial & Kependudukan (12 topik)**, **Ekonomi & Perdagangan (10 topik)**, **Pertanian & Pertambangan (2 topik)**, dan **Statistik Sektoral & Khusus (1 topik)**.
   - Endpoint: `/api/v1/layanan-data`, `/api/v1/layanan-data/:category`, dan `/api/v1/layanan-data/:category/:topic` (serta alias bahasa Inggris `/api/v1/services/data`).
   - Secara dinamis menyajikan daftar tabel statistik (`statictable`) dan publikasi terkait hasil scraping portal BPS Kabupaten Bangka Barat lengkap dengan tautan unduh/portal resmi.

3. **REST API Indikator Strategis Daerah (Menu 2 Terintegrasi)**:
   - Menyediakan endpoint REST API terstruktur untuk 7 Indikator Strategis resmi BPS Bangka Barat: Kependudukan, Pertanian, Kemiskinan, Ketenagakerjaan, Pertumbuhan Ekonomi (PDRB), Harga & Inflasi terkini, serta Kesejahteraan & IPM.
   - Dilengkapi parameter format dinamis (`?format=json` atau `?format=text` siap broadcast) dan filter kategori.
   - Terintegrasi dua arah: data di-scrape berkala, di-cache di memori, dan disajikan secara real-time ke WhatsApp Chatbot dan aplikasi pihak ketiga / dashboard.

4. **Mesin Scraper Terintegrasi (Next.js Server Actions Extractor)**:
   - Ditulis murni dalam Go (tanpa headless browser / tanpa Chrome / tanpa Puppeteer).
   - Menarik data langsung dari portal resmi [bangkabaratkab.bps.go.id](https://bangkabaratkab.bps.go.id/):
     - **100+ Tabel Statistik Statis** (`statictable`) via Next.js Server Actions.
     - **34+ Berita Resmi Statistik (BRS)** beserta thumbnail dan berkas PDF.
     - **50+ Publikasi Statistik** beserta cover buku resolusi tinggi dan berkas PDF.
     - **10+ Berita Kegiatan** beserta foto dokumentasi lapangan.
   - Penjadwalan otomatis via Cron (`0 */4 * * *`) setiap 4 jam dengan automatic cache invalidation, *circuit breaker*, *rate limiting*, dan *jitter delay*.

5. **Penyajian Media Interaktif di WhatsApp**:
   - **Kirim Cover Gambar Otomatis**: Bot langsung mengirimkan foto cover publikasi atau thumbnail rilis BRS ke chat WhatsApp pengguna (`SendImageFromURL`) saat memilih rincian data.
   - **Video In-App Player**: Video literasi statistik menggunakan tautan YouTube resmi BPS yang dapat diputar langsung di dalam aplikasi WhatsApp (*Picture-in-Picture*) tanpa membebani storage server (0 MB local video storage).

6. **Reservasi Konsultasi PST & Notifikasi Grup Internal Kantor**:
   - **Alur Pengajuan Jadwal Interaktif (Menu 6)**: Pemohon (masyarakat/OPD) dapat mengajukan jadwal konsultasi statistik dan kebutuhan data secara langsung melalui bot.
   - **Penerbitan Nomor Tiket Resmi**: Setiap pengajuan diterbitkan nomor tiket resmi (contoh: `#PST-261028-561`) dan disimpan ke tabel `consultation_bookings` di SQLite.
   - **Auto-Forward ke Grup WA Kantor**: Bot secara otomatis meneruskan rincian pemohon ke grup WhatsApp internal kantor (`INTERNAL_PST_GROUP_JID`) agar petugas piket PST dan pakar terkait dapat segera menindaklanjuti.
   - **Auto-Discovery Group JID**: Bot secara otomatis mendeteksi dan menampilkan daftar Group JID yang diikuti di terminal saat terhubung, memudahkan konfigurasi pada lingkungan lokal maupun produksi.

7. **Efisiensi Memori & Keamanan Maksimal**:
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
│   │   ├── server.go               # REST API Fiber v2, Web QR scanner (/qr) & endpoints
│   │   ├── indicators_test.go      # Unit test REST API indikator strategis
│   │   ├── layanan_data_test.go    # Unit test REST API layanan data statistik
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
│   │   ├── layanan_data.go         # Mapper & query Layanan Data (Sosduk, Ekonomi, Pertanian, Sektoral)
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
API_SECRET_KEY=ganti_dengan_token_rahasia_anda

# Keamanan Admin
ADMIN_TRIGGER_KEYWORD="login admin"
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

---

## Dokumentasi REST API Lengkap (Master API Reference)

Server REST API dibangun menggunakan framework performa tinggi **Fiber v2** dan berjalan di port `APP_PORT` (default: `3000`). Seluruh endpoint di bawah `/api/*` dilindungi oleh `API_SECRET_KEY`.

### 1. Metode Autentikasi API
Kirimkan API key pada setiap request melalui salah satu metode berikut:
- **Header**: `X-API-Key: <API_SECRET_KEY>`
- **Bearer Token**: `Authorization: Bearer <API_SECRET_KEY>`
- **Query Parameter**: `?key=<API_SECRET_KEY>`

---

### 2. Daftar Lengkap Endpoint

#### A. Public & Sistem (Tanpa / Dengan Auth)
| Method | Endpoint | Auth | Deskripsi |
| :--- | :--- | :---: | :--- |
| `GET` | `/health` | Publik | Health check server, penggunaan memori RAM, status database scraper, dan circuit breaker. |
| `GET` | `/qr` | Publik | Halaman web pemindai QR code WhatsApp di browser (otomatis terkunci setelah login). |
| `GET` | `/api/v1/qr` | Perlu | Mengambil payload / data QR code WhatsApp dalam format JSON Base64. |

#### B. Indikator Strategis Daerah (Menu 2)
| Method | Endpoint | Auth | Deskripsi |
| :--- | :--- | :---: | :--- |
| `GET` | `/api/v1/indicators` | Perlu | Mengambil seluruh 7 indikator strategis beserta metrik dan link unduh dokumen resmi. |
| `GET` | `/api/v1/indicators/:id` | Perlu | Mengambil 1 indikator berdasarkan ID (`1`-`7`) atau slug (`inflasi`, `kemiskinan`, `kependudukan`, dll). |
| `GET` | `/api/v1/indikator` | Perlu | Alias berbahasa Indonesia untuk `/api/v1/indicators`. |
| `GET` | `/api/v1/indikator/:id` | Perlu | Alias berbahasa Indonesia untuk `/api/v1/indicators/:id`. |

*Parameter Query Khusus Indikator:*
* `format=text`: Mengembalikan format teks pesan terstruktur yang siap kirim via WhatsApp atau terminal.
* `category=<slug>`: Memfilter daftar indikator berdasarkan nama kategori atau slug.

#### C. Layanan Data Statistik (Menu 1)
| Method | Endpoint | Auth | Deskripsi |
| :--- | :--- | :---: | :--- |
| `GET` | `/api/v1/layanan-data` | Perlu | Mengambil seluruh 4 kategori layanan data (Sosial Kependudukan, Ekonomi, Pertanian, Sektoral) beserta ringkasan topik. |
| `GET` | `/api/v1/layanan-data/:category` | Perlu | Mengambil rincian 1 kategori (`sosduk`, `ekonomi`, `pertanian`, `sektoral`) beserta daftar topik di dalamnya. |
| `GET` | `/api/v1/layanan-data/:category/:topic` | Perlu | Mengambil daftar tabel statistik (`statictable`) dan publikasi terkait suatu topik (misal: `kemiskinan`, `penduduk`, `inflasi`, `sawit`, dll). |
| `GET` | `/api/v1/services/data` | Perlu | Alias bahasa Inggris untuk `/api/v1/layanan-data`. |
| `GET` | `/api/v1/services/data/:category` | Perlu | Alias bahasa Inggris untuk `/api/v1/layanan-data/:category`. |
| `GET` | `/api/v1/services/data/:category/:topic` | Perlu | Alias bahasa Inggris untuk `/api/v1/layanan-data/:category/:topic`. |

*Parameter Query Khusus Layanan Data:*
* `format=text`: Mengembalikan format teks pesan WhatsApp berisi ringkasan angka statistik dan tabel terkait siap kirim.
* `limit=<n>`: Batas jumlah tabel/publikasi per topik (default: 5).

#### D. Narahubung Para Pakar (Menu 6 & Mode Admin)
| Method | Endpoint | Auth | Deskripsi |
| :--- | :--- | :---: | :--- |
| `GET` | `/api/v1/pakar` | Perlu | Mengambil daftar lengkap narahubung 8 bidang pakar statistik dari database SQLite. |
| `GET` | `/api/v1/pakar/:id` | Perlu | Mengambil detail 1 narahubung pakar berdasarkan ID bidang (`1`-`8`). |
| `POST` | `/api/v1/pakar/:id` | Perlu | Memperbarui nama dan nomor WhatsApp pakar bidang tertentu (`{"name": "...", "phone": "..."}`). |
| `GET` | `/api/v1/experts` | Perlu | Alias bahasa Inggris untuk `/api/v1/pakar`. |
| `GET` | `/api/v1/experts/:id` | Perlu | Alias bahasa Inggris untuk `/api/v1/pakar/:id`. |

#### E. Pengelolaan & Kontrol Web Scraper (`/api/v1/scraper`)
| Method | Endpoint | Auth | Deskripsi |
| :--- | :--- | :---: | :--- |
| `GET` | `/api/v1/scraper/status` | Perlu | Status scraper, umur data (hours), jumlah item per kind, status circuit breaker, dan jadwal run berikutnya. |
| `POST` | `/api/v1/scraper/run` | Perlu | Menjalankan proses scraping manual on-demand (query `?kind=all`, `pressrelease`, `publication`, `statictable`, dll). |
| `GET` | `/api/v1/scraper/runs` | Perlu | Riwayat log audit siklus scraping (`?limit=10`). |
| `GET` | `/api/v1/scraper/items` | Perlu | Mengambil daftar item hasil scraping di database SQLite (`?kind=...&lang=id&limit=20`). |
| `POST` | `/api/v1/scraper/breaker/reset` | Perlu | Mereset status circuit breaker scraper secara manual jika sedang terbuka (*open*). |

#### F. Broadcast & Manajemen Cache (`/api/v1`)
| Method | Endpoint | Auth | Deskripsi |
| :--- | :--- | :---: | :--- |
| `POST` | `/api/v1/blast` | Perlu | Mengirimkan pesan siaran massal (*broadcast blast*) asynchronous dengan delay acak anti-banned. |
| `POST` | `/api/v1/cache/clear` | Perlu | Membersihkan in-memory cache data di server. |

#### G. BPS Web API Gateway / Proxy (`/api/v1/bps`)
Gateway proxy terintegrasi ke BPS Web API resmi (`domain: 1903` Bangka Barat):
| Method | Endpoint | Auth | Deskripsi |
| :--- | :--- | :---: | :--- |
| `GET` | `/api/v1/bps/list` | Perlu | Query list dinamis BPS API (`?model=...&page=...`). |
| `GET` | `/api/v1/bps/subjects` | Perlu | Daftar subjek statistik BPS Bangka Barat. |
| `GET` | `/api/v1/bps/subcats` | Perlu | Daftar subkategori statistik BPS. |
| `GET` | `/api/v1/bps/variables` | Perlu | Daftar variabel statistik BPS. |
| `GET` | `/api/v1/bps/statictables` | Perlu | Tabel statistik statis dari portal resmi BPS. |
| `GET` | `/api/v1/bps/publications` | Perlu | Katalog buku publikasi BPS Bangka Barat. |
| `GET` | `/api/v1/bps/pressreleases` | Perlu | Daftar Berita Resmi Statistik (BRS) terkini. |
| `GET` | `/api/v1/bps/infographics` | Perlu | Katalog infografis visual BPS. |
| `GET` | `/api/v1/bps/data` | Perlu | Data agregat statistik dinamis BPS. |
| `POST` | `/api/v1/bps/cache/clear` | Perlu | Membersihkan in-memory cache gateway Web API BPS. |

---

### 3. Contoh Penggunaan (`curl`)

**A. Cek Status Scraper & Kesehatan Sistem:**
```bash
curl http://localhost:3000/health
```

**B. Mengambil Seluruh Indikator Strategis (JSON):**
```bash
curl -H "X-API-Key: <YOUR_API_SECRET_KEY>" \
  http://localhost:3000/api/v1/indicators
```

**C. Mengambil Indikator Inflasi Terkini (Format Plaintext Siap Broadcast):**
```bash
curl -H "X-API-Key: <YOUR_API_SECRET_KEY>" \
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
```

**D. Mengambil Layanan Data Statistik (Menu 1) Topik Kemiskinan:**
```bash
# Format JSON (katalog tabel statis & publikasi)
curl -H "X-API-Key: <YOUR_API_SECRET_KEY>" \
  http://localhost:3000/api/v1/layanan-data/sosduk/kemiskinan

# Format Plaintext siap kirim WhatsApp
curl -H "X-API-Key: <YOUR_API_SECRET_KEY>" \
  "http://localhost:3000/api/v1/layanan-data/sosduk/kemiskinan?format=text"
```

**E. Memicu Scraping Manual via API:**
```bash
curl -X POST -H "X-API-Key: <YOUR_API_SECRET_KEY>" \
  "http://localhost:3000/api/v1/scraper/run?kind=all"
```

**F. Mengirim Broadcast WhatsApp (Blast):**
```bash
curl -X POST -H "X-API-Key: <YOUR_API_SECRET_KEY>" \
  -H "Content-Type: application/json" \
  -d '{
    "recipients": ["628123456789", "628987654321"],
    "message": "Halo Sahabat Data! BRS Inflasi Bangka Barat September 2026 telah rilis."
  }' \
  http://localhost:3000/api/v1/blast
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

Uji spesifik REST API Indikator Strategis (Menu 2):
```bash
go test -v -run TestAPIIndicators ./internal/api
```

Uji spesifik REST API Layanan Data Statistik (Menu 1):
```bash
go test -v -run TestAPILayananData ./internal/api
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

## Mode Admin WhatsApp & Pengelolaan Narahubung Pakar

Bot dilengkapi dashboard khusus administrator yang dapat diakses langsung melalui chat WhatsApp untuk keperluan pemeliharaan dan pembaruan data:

### 1. Cara Masuk ke Mode Admin
1. Kirim pesan ke bot WhatsApp dengan kata kunci: `login admin` (atau variasi `login admin - 1903`, `login admin 1903`).
2. Masukkan password administrator sesuai `ADMIN_PASSWORD` di `.env` (dilindungi sistem *lockout* otomatis jika 3 kali salah).

### 2. Fitur Dashboard Admin
* **Menu 1: Kelola Narahubung Para Pakar**:
  - Menampilkan daftar 8 bidang pakar statistik (Sosial Kependudukan, Pertanian, Ekonomi, Distribusi/Inflasi, Sektoral, Khusus, Metodologi, dan Diseminasi).
  - Admin dapat memperbarui nomor WhatsApp dan nama pakar langsung dari chat:
    ```text
    <nomor_bidang> <nomor_wa> <nama_pakar>
    Contoh: 1 08123456789 Bpk. Rahmat, S.Tr.Stat.
    ```
  - Perubahan otomatis disimpan ke database SQLite (`storage/scraper.db`) dan langsung ter-update di REST API (`/api/v1/pakar`).
* **Menu 2: Status Mesin Scraper & Database Lokal**:
  - Menampilkan jumlah total item data tersimpan di SQLite per kategori (Tabel Statis, Publikasi, BRS, Berita Kegiatan).
  - Menampilkan ringkasan eksekusi scraping terakhir (status, pemicu, durasi, jumlah data baru/update).
* **Menu 3: Manual Scraping On-Demand**:
  - Memicu mesin scraper secara instan di background untuk menyedot rilis data terbaru dari portal BPS Bangka Barat.
  - Setelah tuntas (~1 menit), bot otomatis mengirimkan pesan konfirmasi laporan rekap ke WhatsApp admin.
* **Menu 4: Tautan Google Sheets (Legacy)**:
  - Menyajikan tautan cadangan spreadsheet pengelolaan data historis.
* **Menu 0: Keluar dari Mode Admin**.

---

## Keamanan Repositori Publik & Privasi Data

Repositori ini bersifat **Publik (Open Source)** untuk transparansi dan kolaborasi. Seluruh data rahasia dan konfigurasi internal wajib dijaga dengan standar berikut:

1. **JANGAN PERNAH Commit File Kredensial & Database**:
   - `.env` : Berisi kunci API, password admin, dan token WhatsApp.
   - `storage/whatsmeow.db` : Berisi kunci sesi enkripsi WhatsApp (*Signal Protocol keys*).
   - `storage/scraper.db` : Database operasional lokal dan riwayat tiket permohonan PST.
   - Seluruh file di atas **sudah terdaftar di `.gitignore`** dan tidak boleh di-push ke GitHub.
2. **Gunakan Template `.env.example`**:
   - Selalu gunakan `.env.example` untuk mendokumentasikan variabel konfigurasi dengan nilai placeholder (*dummy*).
3. **Penyembunyian Nomor & Identitas**:
   - Jangan menyertakan nomor telepon riil, sandi produksi, atau URL webhook pribadi di dalam kode program atau dokumentasi markdown.
4. **Perlindungan Token & API Secret**:
   - Selalu ganti `API_SECRET_KEY` dan `ADMIN_PASSWORD` pada lingkungan produksi dengan kombinasi yang kuat dan unik.

---

## Standar Keamanan & Perlindungan
- **Anti-Flooding**: Dilengkapi *Token Bucket Rate Limiter* (maksimal 5 pesan per 10 detik per pengguna).
- **Anti-Brute Force Admin**: Maksimal 3 kali percobaan salah sebelum nomor dikunci (*lockout*) selama 60 menit.
- **Anti-SSRF**: Validasi skema HTTPS dan pemblokiran akses ke alamat loopback/privat internal.
- **Data Privacy**: Kerahasiaan nomor dan data pengguna dilindungi sesuai amanat UU No. 16 Tahun 1997 tentang Statistik.

---