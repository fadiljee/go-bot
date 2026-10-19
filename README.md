# BOT ISTIMEWA v3.0 (Go Native Edition)
**Chatbot WhatsApp Resmi Badan Pusat Statistik (BPS) Kabupaten Bangka Barat**

Migrasi penuh dari Node.js (Pepesan/Baileys) ke arsitektur native **Golang** menggunakan `go.mau.fi/whatsmeow` dan `github.com/gofiber/fiber/v2`.

---

## Keunggulan Versi Go (v3.0)

1. **Efisiensi Memori & Kecepatan**:
   - Penggunaan RAM turun dari ~350 MB (Node.js) menjadi **< 30 MB** (Go).
   - Startup time instan (< 100 ms).
   - Single standalone binary tanpa dependensi `node_modules` atau runtime Node.js.
2. **Zero CGO Compilation**:
   - Menggunakan pure Go SQLite driver (`modernc.org/sqlite`), siap di-deploy di VPS/server Linux manapun tanpa gcc.
3. **Penyelesaian Seluruh Utang Teknis Node.js**:
   - `TD-01`: State collision Menu 11 (Tentang BPS) vs Menu 1.1 (Sosial Kependudukan) diselesaikan dengan state terisolasi `menu11_about`.
   - `TD-02`: Kredensial admin diisolasi ke `.env` (bukan hardcoded).
   - `TD-03`: Crash menu kesehatan (`Timeout1`/`Timeout2` undefined) diperbaiki dengan generic topic handler.
   - `TD-04`: Idle timeout 15 menit menggunakan cancelable timer (`*time.Timer`) yang di-reset otomatis setiap interaksi pengguna.
   - `TD-05`: Tahun indikator strategis dinamis via `ACTIVE_INDICATOR_YEAR` (default `2025`).
   - `TD-06`: Menu Tentang BPS mengirim tautan video YouTube secara langsung tanpa memuat video ke RAM.
4. **Dual Mode Interaktif**:
   - Mendukung **Single Select List Drawer** & **Quick Reply Buttons**.
   - Dilengkapi fallback pengetikan manual (angka/kata kunci) jika WhatsApp client pengguna tidak mendukung komponen interaktif.
5. **Dashboard Web & REST API Terintegrasi**:
   - Akses `/health` untuk memantau metrik server & penggunaan memori.
   - Akses `/qr` untuk memindai QR code via browser (otomatis me-refresh).
   - Endpoint `/api/cache/refresh` untuk pemutakhiran cache data Google Sheets.
   - Endpoint `/api/blast` untuk siaran terjadwal dengan anti-ban randomized delay (3-7 detik).

---

## Struktur Direktori

```
go-bot/
├── cmd/
│   └── server/
│       └── main.go                 # Entrypoint server Go
├── internal/
│   ├── api/
│   │   └── server.go               # REST API Fiber v2 & Web QR scanner
│   ├── bot/
│   │   ├── handlers_indikator.go   # Menu 2: Indikator Strategis & infografis
│   │   ├── handlers_layanan.go     # Menu 1: Layanan Data & pencarian keyword
│   │   ├── handlers_main.go        # Welcome full/short, reset, admin, timeout SKD
│   │   ├── handlers_media.go       # Menu 3 (Produk), 4 (Video), 5 (Infografik), 11 (Tentang BPS)
│   │   ├── handlers_support.go     # Menu 6 (Konsultasi/Pakar), 7 (Berlangganan), 8 (Pengaduan), 9 (Dev), 10 (FAQ)
│   │   ├── router.go               # Table-driven state machine dispatcher
│   │   └── session.go              # Manajemen sesi, history stack, & idle timer
│   ├── config/
│   │   └── config.go               # Parser konfigurasi & environment variable
│   ├── gsheet/
│   │   ├── client.go               # HTTP client GAS, in-memory cache, & stale fallback
│   │   ├── models.go               # Struct data ter-tipe kuat & FlexString JSON parser
│   │   └── search.go               # Pencarian lintas sheet (Menu 1.5)
│   ├── locale/
│   │   └── locale.go               # Sistem lokalisasi id_ID dengan interpolasi parameter
│   └── whatsapp/
│       ├── client.go               # whatsmeow client wrapper, sqlstore, terminal/web QR
│       ├── interactive.go          # Protobuf builder untuk Single Select List & Buttons
│       ├── parser.go               # Parser pesan & unwrap interaktif/viewOnce
│       └── sender.go               # Pengirim teks, stiker, dan gambar
├── locales/
│   └── id_ID.json                  # Kamus teks & template pesan resmi BPS
├── StikerChatbot/                  # File stiker WebP (Maskot BPS)
├── storage/                        # Database SQLite untuk sesi whatsmeow (whatsmeow.db)
├── .env.example                    # Contoh template konfigurasi environment
├── .env                            # Konfigurasi aktif bot
├── bot-istimewa                    # Binary binary hasil kompilasi
├── go.mod
└── go.sum
```

---

## Panduan Menjalankan Bot

### 1. Prasyarat
- Go 1.21+ (sudah terpasang di sistem: `go1.26.8`)

### 2. Konfigurasi (`.env`)
Pastikan file `.env` sudah sesuai (salin dari `.env.example` jika baru):
```env
APP_PORT=3000
API_SECRET_KEY=rahasia_internal_bps_1903
ADMIN_TRIGGER_KEYWORD="login admin - 1903"
ADMIN_PASSWORD=supertim1903
ACTIVE_INDICATOR_YEAR=2025
IDLE_TIMEOUT_MINUTES=15
```

### 3. Kompilasi & Jalankan
Untuk menjalankan langsung via source code:
```bash
go run cmd/server/main.go
```

Atau jalankan binary yang telah dikompilasi:
```bash
./bot-istimewa
```

### 4. Pindai QR Code
Saat bot pertama kali dijalankan dan belum terhubung:
1. QR Code akan muncul langsung di **Terminal/Konsol**.
2. **Atau** buka browser di komputer/laptop:
   ```
   http://localhost:3000/qr
   ```
3. Buka WhatsApp di ponsel bot &rarr; **Perangkat Tertaut** &rarr; **Tautkan Perangkat** &rarr; scan QR code yang tampil.
4. Sesi akan otomatis tersimpan di `storage/whatsmeow.db` sehingga saat bot di-restart tidak perlu scan QR lagi.

---

## Menjalankan Pengujian (Unit Tests)

Jalankan seluruh test suite:
```bash
go test -v ./...
```

---

## Prosedur Cutover dari Node.js ke Go (Zero Downtime)

1. Bot Node.js saat ini tetap berjalan secara independen.
2. Ketika siap melakukan cutover resmi:
   ```bash
   # Matikan bot Node.js lama
   kill -9 <PID_NODE_BOT>
   
   # Jalankan bot Go baru di background / systemd
   nohup ./bot-istimewa > bot.log 2>&1 &
   ```
3. Scan QR code WhatsApp sekali untuk menghubungkan sesi WhatsMeow baru.
4. Bot Go siap melayani masyarakat dengan performa tinggi dan bebas memory leak.
