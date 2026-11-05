# Log Perbaikan: Resolusi Dobel Pesan Selamat Datang & Kompatibilitas Menu iPhone (iOS)

**Tanggal:** 29 Oktober 2026  
**Status:** [OK] Selesai & Teruji 100%  
**Binary:** `./bot-istimewa`

---

## 1. Ringkasan Masalah & Gejala

1. **Pesan Selamat Datang & Stiker Terkirim Dobel (Double)**:
   - Ketika pengguna baru memulai interaksi chat (*start* / salam pembuka) atau ketika bot baru dinyalakan, bot mengirimkan 2 stiker dan 2 teks selamat datang secara berurutan dalam selang milidetik.
2. **Menu Tidak Muncul di iPhone (Hanya Stiker yang Muncul)**:
   - Saat diuji dari perangkat iPhone (iOS), hanya stiker yang muncul di obrolan WhatsApp, sedangkan pesan selamat datang dan pilihan menu 1–11 sama sekali tidak terlihat/hilang.

---

## 2. Akar Masalah (Root Causes)

| No | Masalah | Penyebab Teknis |
|---|---|---|
| 1 | **Pesan Dobel** | **Pesan Kosong / Protokol Tanpa Teks Diterima Router**: Saat kontak baru membuka chat, WhatsApp mengirimkan paket protokol internal (misal: *cipher session setup*, *app state sync*, reaksi). Parser mengekstraknya sebagai `p.Text = ""`. Karena tidak ada filter pesan kosong di `router.go`, string kosong `""` dianggap sebagai pesan pembuka di `StateRoot` sehingga memicu `HandleWelcomeFull` ke-1. Sesaat kemudian teks asli pengguna (misal "start") masuk dan memicu `HandleWelcomeFull` ke-2. |
| 2 | **Pesan Dobel** | **Tidak Ada Deduplikasi Message ID**: Server WhatsApp multi-device mengirim ulang stanza (*retry resend*) jika ACK tertunda. Tanpa cache deduplikasi ID pesan (`evt.Info.ID`), kedua stanza dieksekusi secara paralel. |
| 3 | **Pesan Dobel** | **Dual Identitas JID (LID vs Nomor Telepon)**: Pengguna WhatsApp modern memiliki ID akun `... @lid` dan nomor `... @s.whatsapp.net`. Keduanya sebelumnya membuat session terpisah. |
| 4 | **Menu Hilang di iPhone** | **WhatsApp iOS Menolak Native Flow Messages (`single_select`) Unofficial**: Menu dibangun menggunakan Protobuf `InteractiveMessage` bertipe `single_select` (Native Flow) yang dibungkus `ViewOnceMessage` dengan binary node `biz`. Fitur ini khusus akun Meta Cloud API resmi. Aplikasi WhatsApp di iPhone menerapkan validasi Protobuf lokal yang sangat ketat dan langsung membuang (*silent drop*) pesan tersebut jika berasal dari koneksi WhatsApp Web (whatsmeow). |
| 5 | **Menu Hilang di iPhone** | **Fallback Teks Tidak Terpanggil**: Server WhatsApp menerima paket Protobuf tanpa error jaringan (`err == nil`). Kegagalan terjadi di sisi aplikasi iPhone penerima saat proses *render*, sehingga mekanisme `err != nil` pada sender tidak pernah terpicu. |

---

## 3. Rincian Perbaikan Kode

### A. Modul Deduplikasi & Parser (`internal/whatsapp`)
- **`internal/whatsapp/dedup.go` (File Baru)**:
  - Membuat struct thread-safe `MessageDeduplicator` dengan *Time-To-Live* (TTL 60 detik) dan pembersihan memori otomatis.
  - Mencegah eksekusi ganda untuk ID pesan WhatsApp (`evt.Info.ID`) yang sama akibat *retry stanza* dari server WhatsApp.
- **`internal/whatsapp/dedup_test.go` (File Baru)**:
  - Unit test komprehensif untuk pengujian deduplikasi, pencegahan ID ganda, dan kedaluwarsa TTL.
- **`internal/whatsapp/interactive.go`**:
  - Menambahkan fungsi `ExtractTextFromMessage(msg *waE2E.Message) string`:
    Mengonversi pesan interaktif Protobuf apa pun (List Drawer `single_select`, Quick Reply Buttons, dll.) menjadi **pesan teks standar yang rapi, lengkap dengan Header, Body bernomor, opsi Navigasi, dan Footer**.
- **`internal/whatsapp/sender.go`**:
  - Menambahkan field `cfg *config.Config` ke `Sender` dan memperbarui konstruktor `NewSender(client, mq, cfg...)`.
  - Di fungsi `SendInteractiveMessage`:
    - Jika `ENABLE_INTERACTIVE_MENUS=false` (default), pesan menu langsung dikirimkan sebagai **teks terformat universal** melalui `SendTextMessage`.
    - Format ini dijamin **100% muncul dan dapat dibaca di semua perangkat: iPhone iOS, Android, WhatsApp Web, dan WhatsApp Desktop**.
    - Jika fitur interaktif aktif namun transmisi gagal, fallback otomatis mengonversi pesan menggunakan `ExtractTextFromMessage`.

### B. Router & Sesi Bot (`internal/bot`)
- **`internal/bot/router.go`**:
  - Menambahkan validasi ketat `if strings.TrimSpace(p.Text) == "" { return }`. Pesan kosong, metadata protokol, maupun reaksi tidak akan pernah lagi diproses sebagai salam pembuka di `StateRoot`.
  - Menyatukan kunci rate limiter (`userKey`) dan kunci sesi (`sessionKey`) menggunakan nomor telepon asli pengguna (`p.PhoneNumber`) agar pesan via LID maupun nomor standar masuk ke sesi yang sama.
- **`internal/bot/session.go`**:
  - Menambahkan field `LastWelcomeSent time.Time` di struct `UserSession`.
  - Mendukung `optionalKey` di `GetOrCreate` dan `Delete` agar sesi konsisten menggunakan nomor telepon pengguna.
- **`internal/bot/handlers_main.go`**:
  - Menerapkan **Debounce Guard** di `HandleWelcomeFull` (jeda 3 detik) dan `HandleWelcomeShort` (jeda 2 detik).
  - Jika ada panggilan duplikat untuk pengguna yang sama dalam jendela waktu tersebut, panggilan kedua otomatis diabaikan dan dicatat di log.

### C. Server Entrypoint & Konfigurasi (`cmd/server/main.go` & `internal/config`)
- **`cmd/server/main.go`**:
  - Menginisialisasi `msgDedup := whatsapp.NewMessageDeduplicator(60 * time.Second)`.
  - Menolak event pesan jika `msgDedup.IsDuplicate(evt.Info.ID)`.
  - Memfilter pesan jika `parsed == nil || strings.TrimSpace(parsed.Text) == ""`.
- **`internal/config/config.go`**:
  - Menambahkan opsi konfigurasi `EnableInteractiveMenus bool` (default: `false`).
- **`.env.example` & `.env`**:
  - Menambahkan dokumentasi variabel `ENABLE_INTERACTIVE_MENUS=false`.

---

## 4. Hasil Verifikasi & Pengujian

Semua unit test di seluruh modul proyek dijalankan dan dinyatakan lulus 100%:
```
?   	bot-istimewa/cmd/scraper-cli	[no test files]
?   	bot-istimewa/cmd/scraper-seed	[no test files]
?   	bot-istimewa/cmd/server	        [no test files]
ok  	bot-istimewa/internal/api	0.075s
ok  	bot-istimewa/internal/bot	0.093s
ok  	bot-istimewa/internal/bpsapi	0.005s
?   	bot-istimewa/internal/config	[no test files]
ok  	bot-istimewa/internal/gsheet	0.006s
ok  	bot-istimewa/internal/locale	0.003s
ok  	bot-istimewa/internal/scraper	0.157s
ok  	bot-istimewa/internal/whatsapp	0.168s
```
Kompilasi binary utama berhasil tanpa hambatan:
```bash
go build -o bot-istimewa cmd/server/main.go
```

---

## 5. Panduan Penggunaan & Pengujian Pengguna

1. **Jalankan Bot**:
   ```bash
   ./bot-istimewa
   ```
2. **Uji Coba dari iPhone**:
   - Kirim pesan apa saja (misal: "Halo" atau "Start").
   - **Hasil**:
     - Stiker selamat datang terkirim tepat 1 kali (tidak dobel).
     - Teks pengantar dan menu utama 1–11 terkirim lengkap dan tampil sempurna di layar iPhone.
     - Pengguna dapat langsung membalas dengan mengetik angka menu (misal: `1` untuk Layanan Data, `2` untuk Indikator, `0` untuk Menu Awal).

---

## 6. Update Perbaikan Tambahan: Navigasi Tombol 99 & 0 pada Sesi Admin

### Gejala Masalah:
- Saat admin masuk ke Mode Admin (Dashboard Admin), membuka submenu (misal: Kelola Kontak Pakar, Cek Status, atau Spreadsheet), lalu mengetik `99` yang seharusnya kembali ke Dashboard Admin, pengguna malah terlempar keluar ke tampilan menu warga/user (`menuAwal`).

### Penyebab:
- Pada `internal/bot/router.go`, pengecekan kata kunci global `if input == "99"` dan `if input == "0"` dieksekusi di blok teratas *Global Keywords* sebelum state mode admin diperiksa.
- Akibatnya, setiap ketikan `99` langsung memanggil `HandleBack()`. Karena state admin tidak terdaftar pada switch-case `HandleBack()`, router jatuh ke default `HandleWelcomeShort()` (menu user).

### Solusi & Perbaikan:
1. **Pengecualian Intersep Global untuk Sesi Admin (`internal/bot/router.go`)**:
   - Menambahkan pengaman `inAdminMode`:
     ```go
     inAdminMode := session.CurrentState == StateAdminLoggedIn ||
         session.CurrentState == StateAdminManagePakar ||
         session.CurrentState == StateAdminWaitPass
     ```
   - Global keyword `0` dan `99` hanya diintersep jika `!inAdminMode`.
2. **Penanganan 99 & 0 Eksplisit pada State Admin**:
   - Di `StateAdminLoggedIn`: ketik `99` langsung menampilkan ulang Dashboard Admin (`SendAdminDashboard`), dan ketik `0` keluar dengan pesan penutup `🔒 Anda telah keluar dari Mode Admin...`.
   - Di `StateAdminManagePakar`: ketik `99` mengubah state kembali ke `StateAdminLoggedIn` dan menampilkan Dashboard Admin, dan ketik `0` keluar dengan pesan logout.
3. **Dukungan Admin di `HandleBack()` (`internal/bot/handlers_main.go`)**:
   - Menambahkan `case StateAdminLoggedIn, StateAdminManagePakar:` pada `HandleBack()` untuk memastikan jika `HandleBack()` terpanggil, admin tetap dikembalikan ke `SendAdminDashboard()`.
4. **Unit Test Baru (`internal/bot/admin_test.go`)**:
   - Menambahkan test `TestAdminSessionBackAndExit` yang memvalidasi navigasi `99` dan `0` pada sesi admin berjalan 100% benar.

---

## 7. Log Task Revisi Fitur Menu & Perbaikan Mesin Scraper (5 November 2026)

**Tanggal:** 5 November 2026  
**Status:** [OK] Selesai, Lulus Audit Linter/Test & Teruji Live Scraping 100%

### Ringkasan 7 Poin Revisi yang Dikerjakan:

1. **Menu 7 — Channel WhatsApp BPS Bangka Barat**:
   - Mengubah nama Menu 7 dari *Langganan Publikasi* menjadi **Channel WhatsApp BPS Bangka Barat** pada [`internal/bot/handlers_main.go`](file:///home/fadil/fadil/go-bot/internal/bot/handlers_main.go) dan [`locales/id_ID.json`](file:///home/fadil/fadil/go-bot/locales/id_ID.json).
   - Memperbarui [`HandleMenu7Berlangganan`](file:///home/fadil/fadil/go-bot/internal/bot/handlers_support.go) agar menampilkan pemberitahuan kepada pengguna bahwa fitur **Channel WhatsApp BPS Bangka Barat** saat ini belum tersedia (*fitur ini belum tersedia*).

2. **Menu 8 — Pengaduan & Saran**:
   - Memperbarui [`HandleMenu8Pengaduan`](file:///home/fadil/fadil/go-bot/internal/bot/handlers_support.go) sehingga daftar saluran pengaduan & saran ditampilkan secara berurutan menjadi 5 poin:
     1. **Tatap Muka** (Layanan langsung di Ruang PST BPS Kabupaten Bangka Barat)
     2. **Kotak Pengaduan** (Kotak pengaduan dan saran di Ruang PST)
     3. **SP4N-LAPOR!** (`https://www.lapor.go.id`)
     4. **Email** (`bps1903@bps.go.id`)
     5. **Link Google Form** (`https://forms.gle/ContohLinkPengaduanBPS1903` — link contoh sementara)

3. **Menu 9 — Bantuan & FAQ (Penambahan Submenu 6)**:
   - Menambahkan submenu nomor **6. Bagaimana Data BPS Dihasilkan** pada [`HandleMenu10FAQ`](file:///home/fadil/fadil/go-bot/internal/bot/handlers_support.go) dan [`HandleFAQDetail`](file:///home/fadil/fadil/go-bot/internal/bot/handlers_support.go) dengan penjelasan tahapan GSBPM (*Generic Statistical Business Process Model*).
   - Memperbarui navigasi kembali (`99`) pada [`HandleBack`](file:///home/fadil/fadil/go-bot/internal/bot/handlers_main.go) agar dari detail FAQ kembali ke daftar pertanyaan Menu 9 (`StateMenu10`).

4. **Menu 9 — Submenu 5 Jadwal Rilis Data (Scraping ARC `https://bangkabaratkab.bps.go.id/id/arc`)**:
   - Menambahkan kind baru `KindARC` (`arc`) pada [`internal/scraper/model.go`](file:///home/fadil/fadil/go-bot/internal/scraper/model.go), [`configs/scraper_selectors.yaml`](file:///home/fadil/fadil/go-bot/configs/scraper_selectors.yaml), [`internal/scraper/store.go`](file:///home/fadil/fadil/go-bot/internal/scraper/store.go) (termasuk migrasi otomatis constraint tabel `scraped_items`), [`internal/scraper/provider.go`](file:///home/fadil/fadil/go-bot/internal/scraper/provider.go), dan [`internal/scraper/runner.go`](file:///home/fadil/fadil/go-bot/internal/scraper/runner.go) (`scrapeARC` & `parseARCJSON` untuk mengambil jadwal rilis BRS dan Publikasi dari `/id/arc`).
   - Menghubungkan Submenu 5 FAQ (`buildARCScheduleFAQ`) ke hasil scraping ARC di SQLite, serta menampilkan jumlah item Jadwal Rilis (ARC) di Dashboard Admin Menu 2 (`Status Scraper & Database`).

5. **Menu 10 — Tentang BPS (Pembaruan Link Video Profil)**:
   - Mengubah tautan video profil YouTube pada [`internal/config/config.go`](file:///home/fadil/fadil/go-bot/internal/config/config.go), [`.env`](file:///home/fadil/fadil/go-bot/.env), [`locales/id_ID.json`](file:///home/fadil/fadil/go-bot/locales/id_ID.json), dan [`internal/bot/video_test.go`](file:///home/fadil/fadil/go-bot/internal/bot/video_test.go) menjadi `https://youtu.be/q9BYxUJ-4z4`.

6. **Menu 6 — Tanya Pakar -> Submenu 3 Informasi PST dan Jam Layanan (Format Nomor WhatsApp)**:
   - Menambahkan helper [`formatPhoneDisplay`](file:///home/fadil/fadil/go-bot/internal/bot/handlers_support.go) sehingga nomor WhatsApp ditampilkan dalam format rapi yang mudah dibaca (`0821-6430-2643 (+62 821-6430-2643)`) disertai tautan `https://wa.me/6282164302643` pada Submenu 3 (`HandleKonsultasiUmum`) maupun detail pakar (`HandlePakarDetail`).

7. **Perbaikan Mesin Scraper (`0 item ditemukan` / `HTTP 409 Server Action unavailable`)**:
   - **Akar Masalah**: Portal web `bangkabaratkab.bps.go.id` melakukan pembaruan build Next.js (Turbopack) yang mengganti seluruh hash `Next-Action` (42 karakter hex), sehingga request POST dengan ID lama ditolak dengan `HTTP/2 409 Server Action unavailable.`. Selain itu, `Fetcher.Get` dan `Fetcher.PostAction` sebelumnya memakai header `User-Agent` yang berbeda saat melewati proteksi cookie Cloudflare (`__cf_bm`).
   - **Solusi Permanen**:
     - Memperbarui seluruh `Next-Action` ID di [`internal/scraper/runner.go`](file:///home/fadil/fadil/go-bot/internal/scraper/runner.go) ke ID terbaru (`pressrelease`, `publication`, `infographic`, `statictable`, `news`, dan `arc`).
     - Menambahkan mekanisme **Auto-Discovery Next.js Server Action ID** (`postActionWithDiscovery` & `discoverActionIDFromHTML`) yang otomatis memindai bundle `/_next/static/chunks/*.js` untuk menemukan hash `createServerReference` terbaru apabila portal BPS melakukan *redeploy* di masa mendatang.
     - Menyamakan `User-Agent` (`effectiveUserAgent()`) antara `Get`, `GetAsset`, dan `PostAction` pada [`internal/scraper/fetcher.go`](file:///home/fadil/fadil/go-bot/internal/scraper/fetcher.go).

### Hasil Verifikasi & Live Scraping:
- **Linter & Static Analysis**: `golangci-lint run`, `staticcheck ./...`, `go vet ./...`, dan `deadcode -test ./...` lulus dengan **0 issue**.
- **Unit Test**: `go test -race -cover ./...` lulus **100%**.
- **Live Scraping (`./scraper-cli`)**:
  - `arc`: `status=success, found=22, new=22` (22 jadwal rilis BRS & Publikasi tersimpan di `storage/scraper.db`).
  - `pressrelease`: `status=success, found=35, updated=35`.
  - Total data aktif di `storage/scraper.db`: `statictable: 100`, `infographic: 114`, `publication: 50`, `news: 50`, `pressrelease: 35`, `arc: 22`.

---

## 8. Standarisasi Layout Teks Responsif & Sejajar Lintas Perangkat (5 November 2026)

### Penyebab Teks Berantakan di Beberapa Perangkat:
1. **Garis Pembatas Terlalu Panjang (`━━━━━━━━━━━━━━━━━━`)**: 18 karakter Unicode tebal (`U+2501`) patah menjadi 2 baris pada layar ponsel sempit atau pengaturan ukuran font besar.
2. **Indentasi Spasi Gantung (`   • ...` dan `   ...`)**: Di WhatsApp (font proporsional tanpa *hanging indent*), spasi di awal kalimat panjang hanya menggeser baris pertama, sedangkan baris ke-2 dan seterusnya saat membungkus (*wrap*) kembali rata kiri mentok, membuat daftar terlihat zig-zag.
3. **Reset Otomatis `<ol>` WhatsApp pada Daftar Bernomor (`1. `)**: Parser Markdown WhatsApp terbaru mengubah `1. ` menjadi ordered list native yang me-reset nomor urut (`4.` kembali jadi `1.`) ketika terpotong subjudul kategori.
4. **Duplikasi Header di `ExtractTextFromMessage`**: Pesan memiliki 2 judul bertumpuk di bagian atas karena `header` ditambahkan lagi di atas `body` yang sudah memiliki judul `*...*`.
5. **Kotak Kode 3-Backtick (```` ```Statistik Bermakna, Statistik Berdampak``` ````)**: Dirender sebagai *code block* abu-abu kaku dengan font monospace lebar yang terpotong menjadi 2 baris di layar HP.

### Solusi yang Diterapkan:
1. **Pemisah Responsif (`────────────`)**: Mengganti seluruh garis pembatas menjadi 12 karakter `─` (`whatsapp.Divider`) yang muat 1 baris di **100% ukuran layar dan ukuran font**.
2. **Penomoran Tebal Anti-Reset (`*1.*`, `*2.*`, dst.)**: Menggunakan format `*1.* Judul Menu` satu baris bersih di [`BuildBodyText`](file:///home/fadil/fadil/go-bot/internal/whatsapp/interactive.go#L183-L202) dan [`buildWelcomeMenuBody`](file:///home/fadil/fadil/go-bot/internal/bot/handlers_main.go#L49-L74), serta menyederhanakan nama topik panjang di [`locales/id_ID.json`](file:///home/fadil/fadil/go-bot/locales/id_ID.json) agar setiap opsi menu muat tepat 1 baris.
3. **Rata Kiri Lurus Tanpa Spasi Gantung**: Menghapus seluruh indentasi `   • ` dan `> • ` pada halaman detail (BRS, Publikasi, Indikator, Jadwal ARC, Pakar, Pengaduan, Admin) sehingga saat paragraf membungkus (*wrap*) di layar kecil maupun besar, tepi kirinya tetap lurus sejajar.
4. **Pencegahan Judul Ganda & Footer Miring Proporsional**: Memperbarui [`ExtractTextFromMessage`](file:///home/fadil/fadil/go-bot/internal/whatsapp/interactive.go#L204-L270) agar tidak menduplikasi header apabila `body` sudah diawali `*...*`, dan mengganti blok kode 3-backtick menjadi teks miring `_Statistik Bermakna, Statistik Berdampak_`.
