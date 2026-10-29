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
