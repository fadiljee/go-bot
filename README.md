# BOT ISTIMEWA

> Chatbot WhatsApp & REST API Layanan Statistik BPS Kabupaten Bangka Barat berbasis Go (`whatsmeow` & `Fiber`).

---

## Fitur Utama

- **Layanan Data Statistik (Menu 1)**: Katalog topik Sosial & Kependudukan, Ekonomi, Pertanian, dan Sektoral langsung dari database lokal.
- **Indikator Strategis (Menu 2)**: Data inflasi, kemiskinan, ketenagakerjaan, PDRB, IPM, dan indikator penting lainnya.
- **Konsultasi & Reservasi PST (Menu 6)**: Pengajuan jadwal konsultasi data dengan nomor tiket otomatis dan notifikasi ke grup internal kantor.
- **Internal Web Scraper**: Penarikan berkala data tabel, publikasi, dan Berita Resmi Statistik (BRS) dari portal BPS Bangka Barat.
- **REST API**: Endpoint terintegrasi untuk kebutuhan dashboard atau aplikasi pihak ketiga.

---

## Cara Menjalankan

### 1. Prasyarat
- Go 1.21 ke atas
- Git

### 2. Konfigurasi
Salin template konfigurasi `.env.example`:
```bash
cp .env.example .env
```
Sesuaikan konfigurasi di `.env` (port, secret key, nomor admin, dsb).

### 3. Jalankan Aplikasi
```bash
# Langsung jalankan
go run ./cmd/server

# Atau kompilasi menjadi binary
go build -o bot-istimewa ./cmd/server
./bot-istimewa
```

*(Opsional) CLI Scraper manual:*
```bash
go run ./cmd/scraper-cli -kind all
```

### 4. Pindai QR WhatsApp
1. Pindai QR ASCII yang tampil di terminal, **atau**
2. Buka browser di `http://localhost:3000/qr`.
3. Buka WhatsApp di HP -> **Perangkat Tertaut** -> **Tautkan Perangkat** -> Pindai QR.
4. Sesi login otomatis tersimpan di [`storage/whatsmeow.db`](file:///home/fadil/fadil/go-bot/storage/whatsmeow.db).

---

## Ringkasan REST API

Default port: `3000`. Endpoint `/api/*` dilindungi API Key via header `X-API-Key: <SECRET>` atau `Authorization: Bearer <SECRET>`.

| Endpoint | Method | Keterangan |
| :--- | :---: | :--- |
| `/health` | `GET` | Health check & status server |
| `/qr` | `GET` | Halaman scan QR code di browser |
| `/api/v1/indicators` | `GET` | 7 Indikator Strategis (`?format=text` atau `json`) |
| `/api/v1/indicators/:id` | `GET` | Detail indikator berdasarkan ID / slug (contoh: `inflasi`) |
| `/api/v1/layanan-data` | `GET` | Katalog kategori dan topik data statistik |
| `/api/v1/layanan-data/:cat/:topic` | `GET` | Daftar tabel statistik & publikasi per topik |
| `/api/v1/pakar` | `GET` | Daftar narahubung pakar statistik |
| `/api/v1/scraper/status` | `GET` | Status & statistik data mesin scraper |
| `/api/v1/scraper/run` | `POST` | Memicu proses scraping manual |
| `/api/v1/blast` | `POST` | Kirim broadcast WhatsApp massal |

---

## Testing

Jalankan test suite:
```bash
go test -v ./...
```

---

