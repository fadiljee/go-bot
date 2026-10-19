package gsheet

// DefaultTopicsCatalog provides built-in BPS Bangka Barat subject and indicator topics
// for instant, reliable keyword search even when external sheets are syncing or restricted.
func DefaultTopicsCatalog() []SearchResultItem {
	return []SearchResultItem{
		// 1. Sosial & Kependudukan
		{
			Category:    "Layanan Data - Sosial Kependudukan",
			Title:       "Statistik Kependudukan dan Migrasi",
			Description: "Demografi, jumlah penduduk, piramida penduduk, migrasi, proyeksi penduduk Bangka Barat.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=519",
		},
		{
			Category:    "Layanan Data - Sosial Kependudukan",
			Title:       "Statistik Kemiskinan dan Permasalahan Sosial",
			Description: "Pengukuran kemiskinan (P0, P1, P2), garis kemiskinan, perlindungan sosial, dan kondisi kesejahteraan.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=563",
		},
		{
			Category:    "Layanan Data - Sosial Kependudukan",
			Title:       "Statistik Tenaga Kerja & Ketenagakerjaan",
			Description: "Angkatan kerja, TPAK, TPT (Tingkat Pengangguran Terbuka), status pekerjaan, dan lapangan usaha.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=520",
		},
		{
			Category:    "Layanan Data - Sosial Kependudukan",
			Title:       "Statistik Pendidikan",
			Description: "Partisipasi sekolah (APS, APM, APK), angka melek huruf, fasilitas pendidikan, dan jenjang sekolah.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=521",
		},
		{
			Category:    "Layanan Data - Sosial Kependudukan",
			Title:       "Statistik Kesehatan",
			Description: "Angka harapan hidup (AHH), fasilitas kesehatan (Puskesmas, RS), imunisasi, dan gizi balita.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=522",
		},
		{
			Category:    "Layanan Data - Sosial Kependudukan",
			Title:       "Statistik Konsumsi dan Pengeluaran",
			Description: "Pengeluaran per kapita, pola konsumsi makanan dan non-makanan, kecukupan kalori dan protein.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=523",
		},
		{
			Category:    "Layanan Data - Sosial Kependudukan",
			Title:       "Statistik Perumahan dan Pemukiman",
			Description: "Akses sanitasi layak, sumber air minum bersih, status kepemilikan rumah, dan lantai terluas.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=525",
		},
		{
			Category:    "Layanan Data - Sosial Kependudukan",
			Title:       "Statistik Kebudayaan",
			Description: "Fasilitas kebudayaan, cagar budaya, museum, perpustakaan, dan kesenian tradisional Bangka Barat.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=527",
		},
		{
			Category:    "Layanan Data - Sosial Kependudukan",
			Title:       "Statistik Politik dan Komunitas",
			Description: "Partisipasi pemilu, keterwakilan legislatif DPRD, dan partisipasi sosial kemasyarakatan.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=528",
		},
		{
			Category:    "Layanan Data - Sosial Kependudukan",
			Title:       "Statistik Regional dan Area Kecil",
			Description: "Profil kecamatan di Bangka Barat (Muntok, Simpang Teritip, Jebus, Kelapa, Tempilang, Parittiga).",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=540",
		},
		{
			Category:    "Layanan Data - Sosial Kependudukan",
			Title:       "Statistik Gender dan Anak",
			Description: "Indeks Pembangunan Gender (IPG), Indeks Ketimpangan Gender (IKG), dan perlindungan anak.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=564",
		},
		{
			Category:    "Layanan Data - Sosial Kependudukan",
			Title:       "Statistik Masyarakat Informasi (TIK)",
			Description: "Penetrasi internet, kepemilikan ponsel/komputer, dan literasi teknologi informasi.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=565",
		},

		// 2. Ekonomi
		{
			Category:    "Layanan Data - Ekonomi",
			Title:       "Statistik Neraca Ekonomi & PDRB",
			Description: "Produk Domestik Regional Bruto (PDRB) ADHB & ADHK, struktur ekonomi, dan pertumbuhan ekonomi daerah.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=531",
		},
		{
			Category:    "Layanan Data - Ekonomi",
			Title:       "Statistik Bisnis dan Industri",
			Description: "Industri pengolahan mikro, kecil, menengah, dan korporasi industri manufaktur.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=532",
		},
		{
			Category:    "Layanan Data - Ekonomi",
			Title:       "Statistik Keuangan Daerah & Fiskal",
			Description: "Realisasi APBD, pendapatan asli daerah (PAD), dana transfer, dan belanja daerah Bangka Barat.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=534",
		},
		{
			Category:    "Layanan Data - Ekonomi",
			Title:       "Statistik Harga-Harga & Inflasi",
			Description: "Indeks Harga Konsumen (IHK), laju inflasi tahunan/bulanan, dan pemantauan harga bahan pokok.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=536",
		},
		{
			Category:    "Layanan Data - Ekonomi",
			Title:       "Statistik Biaya Tenaga Kerja & Upah",
			Description: "Upah Minimum Kabupaten (UMK) Bangka Barat, upah riil, dan struktur biaya ketenagakerjaan.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=537",
		},
		{
			Category:    "Layanan Data - Ekonomi",
			Title:       "Statistik Energi & Kelistrikan",
			Description: "Produksi dan distribusi listrik PLN, pemakaian daya, dan bauran energi daerah.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=558",
		},
		{
			Category:    "Layanan Data - Ekonomi",
			Title:       "Statistik Pertambangan, Manufaktur & Konstruksi",
			Description: "Komoditas pertambangan timah, galian, dan output sektor konstruksi di Bangka Barat.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=559",
		},
		{
			Category:    "Layanan Data - Ekonomi",
			Title:       "Statistik Transportasi & Pelabuhan Tanjung Kalian",
			Description: "Arus kapal, kendaraan logistik, penumpang feri Pelabuhan Tanjung Kalian Muntok, dan armada jalan.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=560",
		},
		{
			Category:    "Layanan Data - Ekonomi",
			Title:       "Statistik Pariwisata & Perhotelan",
			Description: "Tingkat Penghunian Kamar (TPK) hotel, jumlah kunjungan wisata Menumbing dan pesisir pantai.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=561",
		},
		{
			Category:    "Layanan Data - Ekonomi",
			Title:       "Statistik Perbankan & Jasa Keuangan",
			Description: "Penghimpunan DPK, penyaluran kredit UMKM/modal kerja, dan perkembangan lembaga keuangan.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=562",
		},

		// 3. Pertanian
		{
			Category:    "Layanan Data - Pertanian",
			Title:       "Statistik Pertanian, Perkebunan & Perikanan",
			Description: "Produksi kelapa sawit, lada putih (Muntok White Pepper), karet, padi sawah, dan hasil perikanan tangkap.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=557",
		},
		{
			Category:    "Layanan Data - Pertanian",
			Title:       "Statistik Lingkungan Hidup & Kehutanan",
			Description: "Tutupan hutan, curah hujan, iklim, dan pengelolaan lingkungan di Bangka Barat.",
			Link:        "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=539",
		},

		// 4. Indikator Strategis Ringkasan
		{
			Category:    "Indikator Strategis",
			Title:       "Indikator Kependudukan Bangka Barat",
			Description: "Jumlah penduduk, laju pertumbuhan penduduk, rasio jenis kelamin, dan piramida penduduk.",
			Link:        "http://s.bps.go.id/IndikatorStrategisBabar",
		},
		{
			Category:    "Indikator Strategis",
			Title:       "Indikator Kemiskinan Bangka Barat",
			Description: "Persentase penduduk miskin, garis kemiskinan, indeks kedalaman dan keparahan kemiskinan.",
			Link:        "http://s.bps.go.id/IndikatorStrategisBabar",
		},
		{
			Category:    "Indikator Strategis",
			Title:       "Indikator Ketenagakerjaan (TPT & TPAK)",
			Description: "Tingkat Pengangguran Terbuka (TPT) dan Tingkat Partisipasi Angkatan Kerja (TPAK).",
			Link:        "http://s.bps.go.id/IndikatorStrategisBabar",
		},
		{
			Category:    "Indikator Strategis",
			Title:       "Indikator Pertumbuhan Ekonomi (PDRB)",
			Description: "Laju pertumbuhan Produk Domestik Regional Bruto (PDRB) Kabupaten Bangka Barat.",
			Link:        "http://s.bps.go.id/IndikatorStrategisBabar",
		},
		{
			Category:    "Indikator Strategis",
			Title:       "Indikator Inflasi & Indeks Harga Konsumen",
			Description: "Tingkat inflasi tahunan (y-on-y) dan bulanan (m-to-m) pemantauan komoditas harga pasar.",
			Link:        "http://s.bps.go.id/IndikatorStrategisBabar",
		},
		{
			Category:    "Indikator Strategis",
			Title:       "Indikator Kesejahteraan & Indeks Pembangunan Manusia (IPM)",
			Description: "Capaian IPM, Angka Harapan Hidup, Harapan Lama Sekolah, dan Pengeluaran Riil per Kapita.",
			Link:        "http://s.bps.go.id/IndikatorStrategisBabar",
		},
		{
			Category:    "Indikator Strategis",
			Title:       "Indikator Sektor Pertanian & Komoditas Unggulan",
			Description: "Produktivitas komoditas perkebunan sawit dan lada putih Muntok White Pepper.",
			Link:        "http://s.bps.go.id/IndikatorStrategisBabar",
		},

		// 5. Produk Resmi BPS
		{
			Category:    "Produk Publikasi",
			Title:       "Publikasi Kabupaten Bangka Barat Dalam Angka",
			Description: "Kompilasi komprehensif data statistik tahunan seluruh sektor Kabupaten Bangka Barat.",
			Link:        "https://bangkabaratkab.bps.go.id/id/publication",
		},
		{
			Category:    "Produk Rilis",
			Title:       "Berita Resmi Statistik (BRS) Bangka Barat",
			Description: "Rilis resmi berkala inflasi, ketenagakerjaan, kemiskinan, PDRB, dan ekspor-impor.",
			Link:        "https://bangkabaratkab.bps.go.id/id/pressrelease",
		},
		{
			Category:    "Layanan Data Mikro",
			Title:       "Data Mikro BPS (Silastik)",
			Description: "Portal resmi unduh data mikro survei dan sensus BPS (Susenas, Sakernas, Sensus).",
			Link:        "https://silastik.bps.go.id/",
		},
	}
}

// DefaultIndikators returns fallback strategic indicators data
func DefaultIndikators() []IndikatorItem {
	return []IndikatorItem{
		{
			No:             "1",
			Indikator:      "Jumlah Penduduk Kabupaten Bangka Barat",
			Nilai:          "214.453 Jiwa",
			Nilai2024:      "214.453 Jiwa",
			Nilai2025:      "217.200 Jiwa",
			Penjelasan:     "Jumlah penduduk Kabupaten Bangka Barat berdasarkan hasil proyeksi interim BPS. Tersebar di 6 kecamatan dengan konsentrasi tertinggi di Kecamatan Mentok.",
			Penjelasan2024: "Jumlah penduduk Kabupaten Bangka Barat tahun 2024 tercatat sebanyak 214.453 jiwa.",
			Penjelasan2025: "Jumlah penduduk Kabupaten Bangka Barat tahun 2025 diproyeksikan bertambah menjadi sekitar 217.200 jiwa seiring pertumbuhan alami.",
			Infografis:     "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=519",
		},
		{
			No:             "2",
			Indikator:      "Sektor Pertanian & Perkebunan",
			Nilai:          "Komoditas Kelapa Sawit & Lada Putih",
			Nilai2024:      "Perkebunan menjadi penopang utama PDRB",
			Nilai2025:      "Muntok White Pepper terus didorong sebagai komoditas ekspor bernilai tinggi",
			Penjelasan:     "Sektor pertanian, kehutanan, dan perikanan merupakan salah satu sektor penggerak utama perekonomian masyarakat Bangka Barat.",
			Penjelasan2024: "Produksi perkebunan kelapa sawit dan lada putih tetap stabil menjadi sumber mata pencaharian utama warga.",
			Penjelasan2025: "Program hilirisasi dan sertifikasi indikasi geografis Muntok White Pepper terus ditingkatkan.",
			Infografis:     "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=557",
		},
		{
			No:             "3",
			Indikator:      "Tingkat Kemiskinan",
			Nilai:          "4,41%",
			Nilai2024:      "4,56%",
			Nilai2025:      "4,41%",
			Penjelasan:     "Persentase penduduk miskin Kabupaten Bangka Barat terus menunjukkan tren penurunan dan menjadi salah satu yang terendah di Provinsi Kepulauan Bangka Belitung.",
			Penjelasan2024: "Pada tahun 2024, tingkat kemiskinan tercatat sebesar 4,56% dengan Garis Kemiskinan Rp 612.430 per kapita/bulan.",
			Penjelasan2025: "Pada tahun 2025, tingkat kemiskinan diperkirakan turun ke angka 4,41% didukung efektivitas program perlindungan sosial.",
			Infografis:     "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=563",
		},
		{
			No:             "4",
			Indikator:      "Tingkat Pengangguran Terbuka (TPT)",
			Nilai:          "4,12%",
			Nilai2024:      "4,30%",
			Nilai2025:      "4,12%",
			Penjelasan:     "Tingkat Pengangguran Terbuka (TPT) Kabupaten Bangka Barat mengalami perbaikan didorong penyerapan tenaga kerja di sektor pertanian dan perdagangan.",
			Penjelasan2024: "TPT Kabupaten Bangka Barat tahun 2024 berada pada level 4,30% dengan TPAK mencapai 68,5%.",
			Penjelasan2025: "TPT tahun 2025 membaik menjadi 4,12% seiring bergairahnya kegiatan ekonomi lokal.",
			Infografis:     "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=520",
		},
		{
			No:             "5",
			Indikator:      "Pertumbuhan Ekonomi (PDRB)",
			Nilai:          "4,82%",
			Nilai2024:      "4,55%",
			Nilai2025:      "4,82%",
			Penjelasan:     "Laju Pertumbuhan Ekonomi (LPE) Kabupaten Bangka Barat dihitung berdasarkan Produk Domestik Regional Bruto (PDRB) atas dasar harga konstan.",
			Penjelasan2024: "Pertumbuhan ekonomi tahun 2024 tumbuh sebesar 4,55% didukung sektor pertanian dan industri pengolahan.",
			Penjelasan2025: "Pertumbuhan ekonomi tahun 2025 diproyeksikan menguat ke 4,82% seiring peningkatan aktivitas hilirisasi dan investasi.",
			Infografis:     "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=531",
		},
		{
			No:             "6",
			Indikator:      "Inflasi & Harga Konsumen",
			Nilai:          "2,24%",
			Nilai2024:      "2,45%",
			Nilai2025:      "2,24%",
			Penjelasan:     "Laju inflasi tahunan (y-on-y) di Kabupaten Bangka Barat terjaga stabil dalam sasaran target nasional berkat sinergi TPID.",
			Penjelasan2024: "Inflasi tahun 2024 berada pada tingkat yang stabil di 2,45%.",
			Penjelasan2025: "Inflasi tahun 2025 terkendali pada kisaran 2,24% dengan pasokan pangan strategis yang terjaga.",
			Infografis:     "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=536",
		},
		{
			No:             "7",
			Indikator:      "Indeks Pembangunan Manusia (IPM)",
			Nilai:          "72,15",
			Nilai2024:      "71,85",
			Nilai2025:      "72,15",
			Penjelasan:     "IPM Kabupaten Bangka Barat berstatus 'Tinggi' dengan pertumbuhan positif pada dimensi umur panjang dan hidup sehat, pengetahuan, dan standar hidup layak.",
			Penjelasan2024: "Capaian IPM tahun 2024 sebesar 71,85 dengan Umur Harapan Hidup (AHH) mencapai 73,12 tahun.",
			Penjelasan2025: "IPM tahun 2025 meningkat menjadi 72,15 didukung peningkatan pengeluaran riil per kapita dan harapan lama sekolah.",
			Infografis:     "http://s.bps.go.id/IndikatorStrategisBabar",
		},
	}
}

// DefaultLayananData returns default fallback tables
func DefaultLayananData() []LayananDataItem {
	return []LayananDataItem{
		{No: "111", Judul: "Statistik Kependudukan dan Migrasi", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=519"},
		{No: "112", Judul: "Statistik Kemiskinan dan Perlindungan Sosial", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=563"},
		{No: "113", Judul: "Statistik Tenaga Kerja & Angkatan Kerja", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=520"},
		{No: "114", Judul: "Statistik Pendidikan & Melek Huruf", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=521"},
		{No: "115", Judul: "Statistik Kesehatan & Angka Harapan Hidup", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=522"},
		{No: "116", Judul: "Statistik Pengeluaran dan Konsumsi Rumah Tangga", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=523"},
		{No: "117", Judul: "Statistik Perumahan dan Fasilitas Hunian", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=525"},
		{No: "118", Judul: "Statistik Budaya dan Pariwisata", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=527"},
		{No: "119", Judul: "Statistik Politik dan Komunitas Sosial", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=528"},
		{No: "1110", Judul: "Statistik Regional Kecamatan Bangka Barat", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=540"},
		{No: "1111", Judul: "Statistik Gender dan Perlindungan Kelompok Khusus", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=564"},
		{No: "1112", Judul: "Statistik TIK dan Akses Internet", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=565"},
		{No: "121", Judul: "Statistik PDRB & Neraca Wilayah", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=531"},
		{No: "122", Judul: "Statistik Perusahaan dan Industri", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=532"},
		{No: "123", Judul: "Statistik Keuangan Pemda dan Pajak Daerah", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=534"},
		{No: "124", Judul: "Statistik Indeks Harga Konsumen (IHK) dan Inflasi", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=536"},
		{No: "125", Judul: "Statistik Upah dan Biaya Tenaga Kerja", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=537"},
		{No: "126", Judul: "Statistik Pasokan dan Konsumsi Listrik/Energi", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=558"},
		{No: "127", Judul: "Statistik Pertambangan Timah dan Industri Pengolahan", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=559"},
		{No: "128", Judul: "Statistik Transportasi & Pelabuhan Tanjung Kalian Muntok", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=560"},
		{No: "129", Judul: "Statistik Pariwisata dan Perhotelan Bangka Barat", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=561"},
		{No: "1210", Judul: "Statistik Perbankan dan Kredit Daerah", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=562"},
		{No: "131", Judul: "Statistik Pertanian, Perkebunan Sawit, Karet, Lada & Perikanan", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=557"},
		{No: "132", Judul: "Statistik Lingkungan Hidup, Iklim dan Kehutanan", Link: "https://bangkabaratkab.bps.go.id/id/statistics-table?subject=539"},
	}
}

// DefaultVideos returns fallback video items
func DefaultVideos() []VideoItem {
	return []VideoItem{
		{
			No:         "1",
			Judul:      "Literasi Statistik: Mengenal Sensus dan Survei BPS",
			Link:       "https://www.youtube.com/@bpskabupatenbangkabarat",
			Penjelasan: "Video edukasi mengenai perbedaan sensus dan survei serta pemanfaatan data BPS untuk perencanaan daerah.",
		},
		{
			No:         "2",
			Judul:      "Literasi Statistik: Memahami Inflasi dan Indeks Harga Konsumen",
			Link:       "https://www.youtube.com/@bpskabupatenbangkabarat",
			Penjelasan: "Video penjelasan bagaimana BPS menghitung angka inflasi dan pengaruhnya terhadap daya beli masyarakat.",
		},
	}
}

// DefaultInfografiks returns fallback infographics
func DefaultInfografiks() []InfografikItem {
	return []InfografikItem{
		{
			No:    "1",
			Judul: "Infografik Indikator Strategis Bangka Barat",
			Link:  "https://bangkabaratkab.bps.go.id/id/gallery",
		},
		{
			No:    "2",
			Judul: "Infografik Sensus Pertanian Bangka Barat",
			Link:  "https://bangkabaratkab.bps.go.id/id/gallery",
		},
	}
}

// DefaultFAQs returns fallback FAQ items
func DefaultFAQs() []FAQItem {
	return []FAQItem{
		{
			No:         "1",
			Pertanyaan: "Bagaimana cara memperoleh data statistik BPS Kabupaten Bangka Barat?",
			Jawaban:    "Sahabat Data dapat mengunduh data secara gratis melalui website https://bangkabaratkab.bps.go.id atau mengunjungi Pelayanan Statistik Terpadu (PST) BPS Kabupaten Bangka Barat.",
		},
		{
			No:         "2",
			Pertanyaan: "Apakah layanan data di BPS Bangka Barat berbayar?",
			Jawaban:    "Sebagian besar publikasi, tabel statistik, dan indikator strategis dapat diakses secara GRATIS (Rp 0). Untuk data mikro dan peta digital tertentu mengacu pada tarif PNBP PP No. 7 Tahun 2015.",
		},
		{
			No:         "3",
			Pertanyaan: "Apa itu Pelayanan Statistik Terpadu (PST)?",
			Jawaban:    "PST adalah unit layanan tatap muka maupun daring BPS yang menyediakan konsultasi data, rekomendasi kegiatan statistik, dan penjualan publikasi.",
		},
	}
}
