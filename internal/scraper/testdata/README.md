# Testdata Fixtures — BPS Kabupaten Bangka Barat

Berkas-berkas di direktori ini adalah fixture HTML statis untuk pengujian golden test parser dan deteksi anomali pada `internal/scraper` tanpa bergantung pada koneksi internet langsung ke `bangkabaratkab.bps.go.id`.

## Status Asumsi Fase 0 (§2.2):
- **A-01**: Halaman publikasi, BRS, infografik, dan tabel statis dirender di sisi server (Server-Side Rendering / SSR) -> **Terkonfirmasi**.
- **A-02**: Pagination menggunakan pola `?page=N` -> **Terkonfirmasi**.
- **A-03**: Struktur ID pada URL BRS dan publikasi berada di segmen URL `/YYYY/MM/DD/{ID}/...` -> **Terkonfirmasi**.
- **A-04**: File `robots.txt` melarang direktori privat seperti `/admin/` dan memperbolehkan akses halaman publik -> **Terkonfirmasi**.
- **A-05**: Pengambilan data metadata publik dengan laju rendah dan atribusi sumber disetujui -> **Terkonfirmasi**.
- **A-06**: Tautan unduhan file langsung mengarah ke file `.pdf` / `.xlsx` resmi BPS -> **Terkonfirmasi**.
