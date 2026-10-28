# BOT ISTIMEWA

Dokumentasi Resmi Chatbot WhatsApp Badan Pusat Statistik (BPS) Kabupaten Bangka Barat.

Aplikasi ini dibangun menggunakan arsitektur native Go (Golang) untuk menyediakan layanan informasi statistik terpadu, rilis data berkala, reservasi konsultasi Pelayanan Statistik Terpadu (PST), dan edukasi literasi statistik langsung melalui WhatsApp.

---

## Tech Stack

Proyek ini dibangun menggunakan teknologi modern yang efisien, berkinerja tinggi, dan hemat memori:

- **Bahasa Pemrograman**: Go (Golang) versi 1.22 ke atas (Pure Go, tanpa ketergantungan CGO).
- **WhatsApp Core Engine**: `go.mau.fi/whatsmeow` (Protokol WhatsApp Multi-Device resmi terenkripsi Signal E2EE).
- **Web & API Framework**: `github.com/gofiber/fiber/v2` (Server HTTP ringan untuk antarmuka pemindai QR Code dan endpoint internal).
- **Database Engine**: SQLite melalui driver `modernc.org/sqlite` (Penyimpanan lokal untuk sesi whatsmeow, data hasil scraping, dan tiket reservasi PST).
- **In-Memory Cache & Rate Limiting**: `github.com/patrickmn/go-cache` dan `golang.org/x/time/rate` (Token-bucket rate limiter per nomor pengguna).
- **Integrasi Data**:
  - Mesin Scraper internal untuk ekstraksi otomatis portal resmi BPS Bangka Barat.
  - Client BPS Web API v1 (`internal/bpsapi`).
  - Google Apps Script (GAS) API client dengan mekanisme cache dan fallback data offline.
- **Media Engine**: FFmpeg untuk optimasi dan kompresi video literasi statistik (MP4 H.264/AAC FastStart) agar dapat diputar langsung di dalam aplikasi WhatsApp.

---

## Struktur Direktori

```text
go-bot/
├── cmd/
│   ├── server/
│   │   └── main.go                 # Entrypoint server bot WhatsApp dan server web QR
│   ├── scraper-cli/
│   │   └── main.go                 # CLI tool untuk menjalankan scraping manual
│   └── scraper-seed/
│       └── main.go                 # Seeder inisialisasi data awal database scraper
├── internal/
│   ├── api/
│   │   ├── server.go               # REST API Fiber v2 dan Web QR scanner (/qr)
│   │   └── handlers_scraper.go     # Endpoint API manajemen scraper
│   ├── bot/
│   │   ├── router.go               # Table-driven state machine dispatcher navigasi bot
│   │   ├── session.go              # Manajemen sesi pengguna, history stack, dan idle timer
│   │   ├── handlers_main.go        # Handler menu utama, reset, dan autentikasi admin
│   │   ├── handlers_layanan.go     # Menu 1: Layanan Data dan pencarian kata kunci
│   │   ├── handlers_indikator.go   # Menu 2: Indikator Strategis daerah
│   │   ├── handlers_media.go       # Menu 3 (BRS/Publikasi), Menu 4 (Video), Menu 5 (Infografik)
│   │   └── handlers_support.go     # Menu 6 (Konsultasi/PST), Menu 7-10 (Langganan, Pengaduan, FAQ)
│   ├── scraper/
│   │   ├── runner.go               # Eksekutor scraper Next.js Server Actions
│   │   ├── fetcher.go              # HTTP client dengan cookie jar dan rotasi User-Agent
│   │   ├── store.go                # Abstraksi database SQLite untuk data rilis dan tiket PST
│   │   ├── provider.go             # Penyedia data terpadu untuk router bot
│   │   └── scheduler.go            # Penjadwal otomatis scraping berbasis cron
│   ├── config/
│   │   └── config.go               # Loader konfigurasi dari environment variable
│   ├── locale/
│   │   └── locale.go               # Sistem lokalisasi bahasa (id_ID)
│   └── whatsapp/
│       ├── client.go               # Wrapper koneksi whatsmeow dan penanganan QR code
│       ├── interactive.go          # Builder pesan tombol interaktif dan list dropdown
│       ├── parser.go               # Parser pesan masuk WhatsApp
│       └── sender.go               # Pengirim pesan teks, gambar, video, dan stiker
├── configs/
│   └── scraper_selectors.yaml      # Konfigurasi selector HTML dan rute scraping
├── locales/
│   └── id_ID.json                  # Kamus teks dan pesan layanan resmi
├── storage/
│   ├── whatsmeow.db                # Basis data sesi koneksi WhatsApp (diabaikan git)
│   ├── scraper.db                  # Basis data artikel, publikasi, dan tiket PST (diabaikan git)
│   └── videos/                     # Koleksi file video literasi kompresi MP4 (diabaikan git)
├── .env.example                    # Template konfigurasi variabel environment
├── .env                            # File konfigurasi aktif lokal (diabaikan git)
├── bot-istimewa                    # Binary executable server bot
├── go.mod
└── go.sum
```

---

## Panduan Instalasi & Persiapan

### 1. Prasyarat Sistem
- **Go**: Versi 1.21 ke atas (direkomendasikan versi 1.22+).
- **Sistem Operasi**: Linux, macOS, atau Windows.
- **Koneksi Internet**: Diperlukan untuk sinkronisasi protokol WhatsApp dan scraping web portal.

### 2. Konfigurasi Environment (.env)
Salin file template `.env.example` menjadi `.env`:
```bash
cp .env.example .env
```

Sesuaikan parameter konfigurasi pada file `.env`:
```env
# Konfigurasi Server Web & Port
APP_ENV=production
APP_HOST=127.0.0.1
APP_PORT=3000
API_SECRET_KEY=ganti_dengan_random_string_minimal_32_karakter

# Keamanan Admin Chat WhatsApp
ADMIN_TRIGGER_KEYWORD="login admin - 1903"
ADMIN_PASSWORD=ganti_dengan_password_admin_yang_kuat
ADMIN_WHITELIST_NUMBERS=6281234567890
ADMIN_MAX_RETRY=3
ADMIN_LOCKOUT_MINUTES=60
DEVELOPER_PHONE=6281234567890

# Notifikasi Grup Internal Kantor (PST)
INTERNAL_PST_GROUP_JID=120363xxxxxx@g.us

# Konfigurasi Scraper Portal BPS Bangka Barat
SCRAPER_ENABLED=true
SCRAPER_BASE_URL=https://bangkabaratkab.bps.go.id
SCRAPER_CRON_SPEC="0 */6 * * *"
SCRAPER_DB_PATH=./storage/scraper.db
```

---

## Cara Menjalankan Bot

### 1. Kompilasi Binary
Kompilasi source code menjadi binary server:
```bash
go build -o bot-istimewa ./cmd/server
```

*(Opsional) Kompilasi alat bantu CLI scraper manual:*
```bash
go build -o scraper-cli ./cmd/scraper-cli
```

### 2. Menjalankan Bot
Jalankan binary bot:
```bash
./bot-istimewa
```

Untuk menjalankan di latar belakang (*background service*) pada server Linux:
```bash
nohup ./bot-istimewa > bot.log 2>&1 &
```

### 3. Pindai QR Code WhatsApp (Pairing)
Saat bot dijalankan pertama kali:
1. **Melalui Terminal**: QR Code ASCII akan langsung ditampilkan di layar konsol.
2. **Melalui Browser**: Buka alamat web lokal pada peramban:
   ```text
   http://localhost:3000/qr
   ```
3. Buka aplikasi WhatsApp di ponsel pengelola bot -> **Perangkat Tertaut** -> **Tautkan Perangkat** -> Pindai QR Code yang tampil.
4. Sesi koneksi akan tersimpan secara terenkripsi di `storage/whatsmeow.db`. Bot tidak perlu memindai ulang QR Code saat dilakukan restart aplikasi.

---

## Panduan Penggunaan Layanan Bot

Bot menyediakan 11 menu layanan utama yang dapat diakses oleh masyarakat melalui chat WhatsApp:

### Struktur Menu Utama
- **Menu 1: Layanan Data**  
  Menyediakan katalog tabel data statistik sektoral dan umum (Sosial Kependudukan, Ekonomi, Pertanian, Sektoral), serta fitur **Pencarian Kata Kunci** langsung ke basis data BPS.
- **Menu 2: Indikator Strategis**  
  Menampilkan angka-angka strategis resmi terkini seperti Kemiskinan, Inflasi, Indeks Pembangunan Manusia (IPM), Pertumbuhan Ekonomi (PDRB), Ketenagakerjaan, dan Pertanian.
- **Menu 3: Produk Statistik**  
  Akses langsung ke ringkasan Berita Resmi Statistik (BRS), katalog buku publikasi resmi (PDF), dan panduan akses data mikro (SILASTIK).
- **Menu 4: Video Literasi Statistik**  
  Video edukasi resmi berdurasi ringkas yang dikirim langsung oleh bot dalam format MP4 dan dapat diputar seketika di WhatsApp (Membahas Kemiskinan, IPM, Ketenagakerjaan, PDRB, Inflasi, Romantik Statistik Sektoral, dan Statistik Khusus).
- **Menu 5: Infografik**  
  Visualisasi grafik ringkas keadaan sosial, ekonomi, dan pertanian daerah.
- **Menu 6: Konsultasi Statistik & Reservasi PST**  
  - Pengajuan jadwal konsultasi data bagi masyarakat dan OPD.
  - Penerbitan nomor tiket antrean resmi.
  - Penerusan otomatis (*auto-forward*) informasi pemohon ke grup WhatsApp internal petugas kantor.
  - Kontak tanya pakar bidang statistik.
- **Menu 7: Berlangganan Data**  
  Formulir pendaftaran nomor WhatsApp untuk menerima siaran berkala produk rilis statistik terbaru.
- **Menu 8: Pengaduan & Aspirasi Layanan**  
  Saluran penyampaian kritik, saran, dan pengaduan layanan (SP4N-LAPOR!, email, dan kotak saran).
- **Menu 9: Kontak Pengembang**  
  Saluran komunikasi teknis dengan tim IT pengembang bot.
- **Menu 10: Frequently Asked Questions (FAQ)**  
  Daftar jawaban resmi atas pertanyaan yang paling sering diajukan mengenai data dan layanan BPS.
- **Menu 11: Profil BPS Bangka Barat**  
  Video profil kelembagaan, tugas pokok dan fungsi, serta alamat kantor BPS Kabupaten Bangka Barat.

### Navigasi Global
Pengguna dapat mengetik perintah berikut kapan saja:
- `0` : Kembali ke Menu Utama
- `99` : Kembali ke Menu Sebelumnya (*Back Stack*)
- `reset` : Menghapus riwayat sesi percakapan
- `berhenti_langganan` : Berhenti menerima siaran berita berkala

---

## Konfigurasi Notifikasi Grup Kantor (PST)

Untuk mengaktifkan penerusan otomatis jadwal konsultasi baru ke grup WhatsApp petugas kantor:

1. Tambahkan nomor WhatsApp bot ke dalam grup kantor yang dituju.
2. Jalankan bot (`./bot-istimewa`). Saat terhubung, bot akan mendeteksi dan mencetak seluruh nama grup beserta JID-nya di log terminal:
   ```text
   [DAFTAR GRUP WA BOT]:
      1. Nama: 'PST Internal BPS Babar' -> JID: 120363xxxxxx@g.us
   ```
3. Salin nilai JID yang berakhiran `@g.us` tersebut ke variabel `INTERNAL_PST_GROUP_JID=` pada file `.env`.
4. Restart bot. Setiap ada pemohon yang mengajukan jadwal konsultasi, bot akan otomatis meneruskan detail permohonan ke grup tersebut.

---

## Pengujian Sistem (Unit Testing)

Jalankan seluruh rangkaian pengujian otomatis untuk memverifikasi fungsionalitas sistem:
```bash
GOCACHE=$(pwd)/.gocache go test -v ./...
```

Uji spesifik komponen per modul:
```bash
# Pengujian alur reservasi konsultasi
GOCACHE=$(pwd)/.gocache go test -v -run TestConsultationBookingWorkflow ./internal/bot

# Pengujian alur video literasi
GOCACHE=$(pwd)/.gocache go test -v -run TestVideoLiterasiNavigationAndDelivery ./internal/bot

# Pengujian modul WhatsApp interactive list & button
GOCACHE=$(pwd)/.gocache go test -v ./internal/whatsapp
```

---

## Standar Keamanan & Perlindungan

- **Perlindungan Kredensial**: File `.env`, folder `storage/`, database SQLite, dan file konfigurasi privat diabaikan dari riwayat Git melalui `.gitignore`.
- **Anti-Flooding**: Dilengkapi *Token Bucket Rate Limiter* (maksimal 5 pesan per 10 detik per pengguna) untuk mencegah spam dan proteksi dari pemblokiran WhatsApp.
- **Anti-Brute Force Admin**: Akses admin dilindungi whitelist nomor telepon dan sistem *lockout* 60 menit setelah 3 kali kegagalan kata sandi.
- **Anti-SSRF**: Pengecekan ketat pada pengunduhan media eksternal (hanya mengizinkan skema HTTPS resmi dan memblokir alamat IP lokal/loopback).
- **Kerahasiaan Data**: Kerahasiaan identitas dan data pemohon dilindungi sesuai amanat Undang-Undang No. 16 Tahun 1997 tentang Statistik.

---

**Badan Pusat Statistik Kabupaten Bangka Barat**  
Komplek Perkantoran Pemerintah Daerah Terpadu, Kabupaten Bangka Barat, Kepulauan Bangka Belitung.
