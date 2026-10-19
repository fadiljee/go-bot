# Security, Dependencies & Operational Hardening — BOT ISTIMEWA v3.0
## Spesifikasi Keamanan, Kebutuhan Sistem, & Anti-Banned WhatsApp

| Field | Detail |
|---|---|
| **Produk** | ISTIMEWA (Informasi Statistik Terpadu melalui WhatsApp)[cite: 1, 2] |
| **Instansi** | BPS Kabupaten Bangka Barat[cite: 1, 2] |
| **Arsitektur Target** | Go (`whatsmeow`) + Go Fiber + SQLite |
| **Dokumen** | Security & Infrastructure Specification (`secure.md`) |
| **Versi** | 3.0.0 |
| **Tanggal** | 19 Oktober 2026 |
| **Status** | Ready for Implementation |

---

## 1. Daftar Kebutuhan Sistem & Dependencies (*Bill of Materials*)

### 1.1 Go Modules & Libraries (`go.mod`)

| Kategori | Package / Library | Fungsi & Alasan Penggunaan |
|---|---|---|
| **WA Core Engine** | `go.mau.fi/whatsmeow` | Koneksi langsung WebSocket WhatsApp Multi-Device, enkripsi E2EE Signal, manajemen sesi, dan pengiriman *Interactive Message*. |
| **Protobuf** | `google.golang.org/protobuf` | Merakit struktur payload `waE2E.InteractiveMessage` (`single_select` & `quick_reply`)[cite: 2]. |
| **SQLite Driver** | `modernc.org/sqlite` | Driver SQLite *pure Go* (tanpa CGO) untuk menyimpan database sesi WA (`session.db`) dan log antrean *blast*. Mudah di-*cross-compile* ke Linux. |
| **HTTP & API Server** | `github.com/gofiber/fiber/v2` | Web server ringan untuk *health check*, *QR viewer*, *webhook cache clear*, dan trigger *blast* berlangganan. |
| **In-Memory Cache** | `github.com/patrickmn/go-cache` | Menyimpan cache JSON dari 6 endpoint Google Apps Script (TTL 15 menit)[cite: 1, 2] serta menyimpan state navigasi pengguna. |
| **Rate Limiter** | `golang.org/x/time/rate` | *Token bucket rate limiter* per nomor pengguna (mencegah spam/flooding chat masuk). |
| **Environment** | `github.com/joho/godotenv` | Memuat variabel sensitif dari file `.env` agar tidak ada kredensial *hardcoded* di kode sumber[cite: 1, 2]. |
| **QR Terminal/Web** | `github.com/skip2/go-qrcode` | Merender QR code login WhatsApp sebagai gambar PNG (via endpoint `/qr` terproteksi) atau di terminal. |
| **Logger** | `log/slog` *(Built-in Go 1.21+)* | Pencatatan log sistem dan *audit trail* keamanan dalam format JSON terstruktur[cite: 1]. |

### 1.2 Kebutuhan Infrastruktur Server (Deployment)

| Spesifikasi | Minimum | Rekomendasi Produksi |
|---|---|---|
| **CPU** | 1 vCPU | 1 - 2 vCPU |
| **RAM** | 256 MB | 512 MB - 1 GB |
| **Disk / Storage** | 1 GB Persistent Disk | 5 GB SSD (Wajib *persistent volume* untuk `session.db`) |
| **OS** | Linux (Debian/Ubuntu/Alpine) | Ubuntu 22.04 / 24.04 LTS |
| **Reverse Proxy** | Cloudflare Tunnel (`cloudflared`) | Nginx / Cloudflare Tunnel (Zero-Open-Port) |

---

## 2. Arsitektur Message Queue & Anti-Banned (Operational Security)

Karena menggunakan *Unofficial WhatsApp API* (`whatsmeow`), pengiriman pesan keluar **dilarang keras** dilakukan secara langsung tanpa antrean ketika trafik tinggi. Sistem wajib menggunakan **Dual-Priority Message Queue**.

```
Pesan Keluar
    │
    ├──► [1] InteractiveQueue (Prioritas Tinggi - Balasan Menu Warga)
    │         └── Jeda Acak: 600ms – 1200ms + Indikator "Sedang mengetik..."
    │
    └──► [2] BlastQueue (Prioritas Rendah - Broadcast Langganan & Outro SKD)
              └── Jeda Acak: 5000ms – 10000ms (Hanya jalan jika InteractiveQueue kosong)
```

### 2.1 Aturan Mitigasi Banned Meta
1. **Simulasi Kehadiran Manusia (*Chat Presence*):**
   Sebelum mengirim balasan menu interaktif, worker wajib mengirim sinyal `ChatPresenceComposing` (*"sedang mengetik..."*), menunggu jeda acak (*jitter*), lalu mengirim `ChatPresencePaused` sebelum melepas pesan.
2. **Pemisahan Jalur Prioritas:**
   Saat admin melakukan *Blast Berlangganan* (FR-07)[cite: 1] ke ratusan nomor, warga yang sedang mengakses menu data harus tetap dilayani dalam waktu < 1.5 detik melalui `InteractiveQueue`.
3. **Tombol *Opt-Out* / Berhenti Berlangganan:**
   Setiap pesan *blast* wajib menyertakan tombol `Quick Reply`: **`[ 🔕 Berhenti Berlangganan ]`**. Hal ini mencegah pengguna menekan tombol **"Report Spam / Laporkan"** bawaan WhatsApp yang menjadi pemicu utama pemblokiran nomor oleh Meta.

### 2.2 Implementasi `internal/whatsapp/queue.go`

```go
package whatsapp

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

type OutboundJob struct {
	JID          types.JID
	Message      *waE2E.Message
	SimulateType bool
}

type MessageQueue struct {
	client           *whatsmeow.Client
	interactiveQueue chan OutboundJob
	blastQueue       chan OutboundJob
}

func NewMessageQueue(cli *whatsmeow.Client) *MessageQueue {
	mq := &MessageQueue{
		client:           cli,
		interactiveQueue: make(chan OutboundJob, 500),
		blastQueue:       make(chan OutboundJob, 2000),
	}
	go mq.worker()
	return mq
}

func (mq *MessageQueue) EnqueueInteractive(jid types.JID, msg *waE2E.Message) {
	mq.interactiveQueue <- OutboundJob{JID: jid, Message: msg, SimulateType: true}
}

func (mq *MessageQueue) EnqueueBlast(jid types.JID, msg *waE2E.Message) {
	mq.blastQueue <- OutboundJob{JID: jid, Message: msg, SimulateType: false}
}

func (mq *MessageQueue) worker() {
	for {
		select {
		case job := <-mq.interactiveQueue:
			mq.sendWithJitter(job, 600, 1200)
		default:
			select {
			case job := <-mq.interactiveQueue:
				mq.sendWithJitter(job, 600, 1200)
			case job := <-mq.blastQueue:
				mq.sendWithJitter(job, 5000, 10000)
			case <-time.After(100 * time.Millisecond):
			}
		}
	}
}

func (mq *MessageQueue) sendWithJitter(job OutboundJob, minMs, maxMs int) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if job.SimulateType {
		_ = mq.client.SendChatPresence(job.JID, types.ChatPresenceComposing, types.ChatPresenceMediaText)
	}

	delay := time.Duration(minMs+rand.Intn(maxMs-minMs)) * time.Millisecond
	time.Sleep(delay)

	if job.SimulateType {
		_ = mq.client.SendChatPresence(job.JID, types.ChatPresencePaused, types.ChatPresenceMediaText)
	}

	if _, err := mq.client.SendMessage(ctx, job.JID, job.Message); err != nil {
		slog.Error("Gagal mengirim pesan WA", "jid", job.JID.String(), "error", err)
	}
}
```

---

## 3. Spesifikasi Keamanan Berlapis (*Security Hardening*)

### 3.1 Lapis 1: Pengamanan Mode Admin via Chat (Resolusi TD-02)
Pada versi lama, kata kunci `login admin - 1903` dan sandi `supertim1903` tertulis langsung di `router.js` dan bisa dicoba oleh nomor mana saja[cite: 1, 2]. Di versi v3.0, berlaku aturan berikut:
1. **Environment Config:** Kredensial dimuat dari `.env` (`ADMIN_TRIGGER_KEYWORD` dan `ADMIN_PASSWORD`).
2. **Phone Number Whitelist (`ADMIN_WHITELIST_NUMBERS`):**
   Hanya nomor WhatsApp yang terdaftar di `.env` yang diizinkan memicu state login admin. Jika nomor asing mengetik keyword admin, bot akan memperlakukannya sebagai teks tidak valid biasa (*silent reject*).
3. **Brute-Force Protection:**
   Maksimal **3 kali percobaan salah password** per nomor. Jika gagal 3 kali berturut-turut, nomor tersebut dikunci (*lockout*) dari percobaan login selama **60 menit** dan dicatat di log keamanan.
4. **Auto-Expire Admin Session:**
   Sesi admin otomatis kembali ke state pengguna biasa setelah menu link Google Sheet dikirimkan.

### 3.2 Lapis 2: Inbound Filtering & Anti-Flooding (DDoS Protection)
1. **Strict Source Filtering:**
   Bot wajib mengabaikan pesan yang berasal dari:
   - Grup WhatsApp (`evt.Info.IsGroup == true`)
   - Status / Story Broadcast (`evt.Info.Chat.Server == "broadcast"`)
   - Pesan dari nomor bot sendiri (`evt.Info.IsFromMe == true`)
   - Pesan lama sebelum waktu *startup* bot (mencegah *spam* balasan massal saat bot baru menyala setelah *offline*).
2. **Per-User Token Bucket Rate Limit:**
   Setiap nomor pengguna dibatasi maksimal **5 pesan per 10 detik**. Jika pengguna mengirim pesan secara membabi buta (misal 20 pesan dalam 3 detik), bot mengaktifkan status *cooldown* selama **1 menit** untuk nomor tersebut dan mengabaikan seluruh pesannya tanpa membalas (agar tidak memicu *infinite loop* antar-bot).
3. **Sanitasi & Batas Karakter Input:**
   Input teks pengguna dipangkas maksimal **100 karakter** (`strings.TrimSpace`) sebelum diproses oleh router atau fitur Pencarian Kata Kunci (Menu 1.5)[cite: 1, 2].

### 3.3 Lapis 3: Keamanan HTTP Server (Go Fiber API)
1. **Localhost Binding:**
   Secara *default*, Go Fiber berjalan di `127.0.0.1:3000` agar tidak terekspos langsung ke jaringan publik tanpa melewati *Reverse Proxy* atau Cloudflare Tunnel.
2. **API Key Authentication Middleware:**
   Endpoint internal wajib menyertakan header `X-API-Key: <API_SECRET_KEY>` atau `Authorization: Bearer <API_SECRET_KEY>`:
   - `POST /api/v1/cache/clear` (Dipanggil oleh Google Apps Script saat admin mengubah Sheet).
   - `POST /api/v1/blast` (Memicu pengiriman pesan langganan).
   - `GET /api/v1/qr` (Menampilkan QR Code jika sesi terputus).
3. **Auto-Lock Endpoint QR:**
   Endpoint `/api/v1/qr` otomatis mengembalikan `403 Forbidden` apabila status `whatsappClient.IsLoggedIn() == true`.

### 3.4 Lapis 4: Keamanan Pengambilan Media & URL (Anti-SSRF & Anti-OOM)
Karena bot menarik gambar infografik (Menu 5), foto pakar (Menu 6), dan video (Menu 11) dari URL di Google Sheets[cite: 1]:
1. **Validasi Skema & Domain (Anti-SSRF):**
   Hanya izinkan URL dengan protokol `https://` dan blokir akses ke IP lokal/privat (`127.0.0.1`, `localhost`, `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`, `169.254.169.254`).
2. **Batas Ukuran Unduhan Media (`io.LimitReader`):**
   Batasi pembacaan *stream* file media di memori maksimal **16 MB**[cite: 1]:
   ```go
   const MaxMediaBytes = 16 * 1024 * 1024 // 16 MB
   limitedReader := io.LimitReader(resp.Body, MaxMediaBytes)
   mediaBytes, err := io.ReadAll(limitedReader)
   ```
   Ini mencegah server kehabisan RAM (*Out-Of-Memory Crash*) jika link di Google Sheets tidak sengaja mengarah ke file berukuran ratusan MB.

### 3.5 Lapis 5: Keamanan File Sesi & Sistem Operasi
File `session.db` menyimpan kunci privat enkripsi Signal WhatsApp. Siapa pun yang memiliki salinan file ini dapat mengambil alih sesi nomor WhatsApp BPS tanpa perlu *scan* QR.
1. **File Permissions (Linux):**
   ```bash
   chmod 600 .env
   chmod 600 ./data/session.db
   chmod 700 ./data
   ```
2. **Non-Root Execution:**
   Jalankan aplikasi menggunakan *user* sistem khusus tanpa akses *shell* (`botistimewa`), bukan sebagai `root`.
3. **Git Protection (`.gitignore`):**
   Pastikan `.env`, `*.db`, `*.db-shm`, `*.db-wal`, dan `session/` masuk ke dalam `.gitignore`[cite: 1].

---

## 4. Template Konfigurasi `.env.example`

Salin template ini menjadi `.env` pada server produksi:

```env
# ==========================================
# 1. KONFIGURASI SERVER & GO FIBER
# ==========================================
APP_ENV=production
APP_HOST=127.0.0.1
APP_PORT=3000
API_SECRET_KEY=ganti_dengan_random_string_minimal_32_karakter

# ==========================================
# 2. IDENTITAS & SESI WHATSAPP
# ==========================================
WA_DEVICE_NAME="BPS KABUPATEN BANGKA BARAT"
WA_DB_PATH=./data/session.db
# Kosongkan ALLOWED_NUMBERS untuk mode publik. Isi (pisahkan koma) untuk mode testing.
ALLOWED_NUMBERS=

# ==========================================
# 3. KEAMANAN ADMIN CHAT
# ==========================================
ADMIN_TRIGGER_KEYWORD="login admin - 1903"
ADMIN_PASSWORD=ganti_password_admin_yang_kuat_1903
# Hanya nomor di bawah ini yang bisa login admin di chat WA (format 628xxx)
ADMIN_WHITELIST_NUMBERS=6285860125305
ADMIN_MAX_RETRY=3
ADMIN_LOCKOUT_MINUTES=60

# ==========================================
# 4. RATE LIMITER, IDLE TIMER & QUEUE
# ==========================================
USER_RATE_LIMIT_PER_10S=5
USER_COOLDOWN_SECONDS=60
IDLE_TIMEOUT_MINUTES=15
QUEUE_INTERACTIVE_MIN_DELAY_MS=600
QUEUE_INTERACTIVE_MAX_DELAY_MS=1200
QUEUE_BLAST_MIN_DELAY_MS=5000
QUEUE_BLAST_MAX_DELAY_MS=10000

# ==========================================
# 5. ENDPOINT GOOGLE APPS SCRIPT & CACHE
# ==========================================
CACHE_TTL_MINUTES=15
HTTP_CLIENT_TIMEOUT_SECONDS=10
MAX_MEDIA_SIZE_MB=16

GAS_URL_LAYANAN_DATA=[https://script.google.com/macros/s/xxx/exec](https://script.google.com/macros/s/xxx/exec)
GAS_URL_INDIKATOR=[https://script.google.com/macros/s/xxx/exec](https://script.google.com/macros/s/xxx/exec)
GAS_URL_VIDEO=[https://script.google.com/macros/s/xxx/exec](https://script.google.com/macros/s/xxx/exec)
GAS_URL_INFOGRAFIK=[https://script.google.com/macros/s/xxx/exec](https://script.google.com/macros/s/xxx/exec)
GAS_URL_PAKAR=[https://script.google.com/macros/s/xxx/exec](https://script.google.com/macros/s/xxx/exec)
GAS_URL_FAQS=[https://script.google.com/macros/s/xxx/exec](https://script.google.com/macros/s/xxx/exec)
```

---

## 5. Checklist Audit Sebelum Go-Live (Pre-Deployment Checklist)

- [ ] File `.env` sudah diisi dengan `API_SECRET_KEY` dan `ADMIN_PASSWORD` baru (bukan bawaan contoh).
- [ ] Nomor HP admin BPS sudah didaftarkan di `ADMIN_WHITELIST_NUMBERS`.
- [ ] Permission file `.env` dan `data/session.db` sudah diset ke `600`.
- [ ] Bot hanya merespons pesan yang masuk *setelah* waktu bot dijalankan (*ignore historical messages on connect*).
- [ ] Uji kirim 10 pesan beruntun dalam 2 detik dari nomor tester → pastikan *rate limiter* menahan spam dan tidak membalas semua pesan sekaligus.
- [ ] Uji login admin dari nomor yang **tidak** ada di *whitelist* → pastikan ditolak meskipun password benar.
- [ ] Uji matikan koneksi internet/blokir URL Google Apps Script sementara → pastikan fitur *Stale-If-Error Cache* tetap menyajikan data terakhir tanpa membuat bot *crash*.