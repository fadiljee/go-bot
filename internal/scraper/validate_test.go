package scraper

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestValidateItem(t *testing.T) {
	now := time.Now()
	allowedHosts := []string{"bangkabaratkab.bps.go.id"}

	// Valid PressRelease
	validItem := &Item{
		Kind:        KindPressRelease,
		SourceKey:   "123",
		Title:       "Inflasi Bangka Barat September 2024",
		URL:         "https://bangkabaratkab.bps.go.id/id/pressrelease/2024/09/01/123/inflasi.html",
		ReleaseDate: &now,
		FileURL:     "https://bangkabaratkab.bps.go.id/download/123.pdf",
	}

	if err := ValidateItem(validItem, allowedHosts, "test"); err != nil {
		t.Fatalf("expected valid item, got error: %v", err)
	}
	if validItem.ContentHash == "" {
		t.Errorf("expected content hash to be generated")
	}

	// Missing Title
	badItem := *validItem
	badItem.Title = "   "
	if err := ValidateItem(&badItem, allowedHosts, "test"); !errors.Is(err, ErrInvalidItem) {
		t.Errorf("expected ErrInvalidItem for empty title, got: %v", err)
	}

	// Title > 300 characters
	badItem = *validItem
	badItem.Title = strings.Repeat("A", 305)
	if err := ValidateItem(&badItem, allowedHosts, "test"); !errors.Is(err, ErrInvalidItem) {
		t.Errorf("expected ErrInvalidItem for long title, got: %v", err)
	}

	// Missing ReleaseDate for pressrelease
	badItem = *validItem
	badItem.ReleaseDate = nil
	if err := ValidateItem(&badItem, allowedHosts, "test"); !errors.Is(err, ErrInvalidItem) {
		t.Errorf("expected ErrInvalidItem for missing release date, got: %v", err)
	}

	// Non-BPS FileURL gets cleared but item passes
	itemWithBadFile := *validItem
	itemWithBadFile.FileURL = "https://external-hacker.com/malware.exe"
	if err := ValidateItem(&itemWithBadFile, allowedHosts, "production"); err != nil {
		t.Fatalf("expected item to pass validation even with cleared file_url, got %v", err)
	}
	if itemWithBadFile.FileURL != "" {
		t.Errorf("expected non-bps file_url to be cleared, got %s", itemWithBadFile.FileURL)
	}
}

func TestAnomalyGuard(t *testing.T) {
	// 1. Zero items when previously >= 5
	err := AnomalyGuard(KindPressRelease, 0, 10, []byte("ok"), nil)
	if !errors.Is(err, ErrAnomalyLayoutChanged) {
		t.Errorf("expected ErrAnomalyLayoutChanged, got: %v", err)
	}

	// 2. Count drop > 50%
	err = AnomalyGuard(KindPressRelease, 3, 10, []byte("ok"), nil)
	if !errors.Is(err, ErrAnomalyCountDrop) {
		t.Errorf("expected ErrAnomalyCountDrop, got: %v", err)
	}

	// 3. Normal count (no drop)
	err = AnomalyGuard(KindPressRelease, 8, 10, []byte("ok"), nil)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// 4. Missing required marker
	err = AnomalyGuard(KindPressRelease, 10, 10, []byte("<html><body>Hello</body></html>"), []string{"Badan Pusat Statistik"})
	if !errors.Is(err, ErrAnomalyMarkerMissing) {
		t.Errorf("expected ErrAnomalyMarkerMissing, got: %v", err)
	}
}

func TestCleanHTML(t *testing.T) {
	input := "Pada Bulan Agustus Tahun 2026 terjadi inflasi <i>year on year (y-on-y)</i> Kabupaten Bangka Barat sebesar 4,43 persen dengan Indeks Harga Konsumen (IHK) sebesar 107,44. <i/>\nInflasi <i>y-on-y</i> terjadi karena kenaikan indeks &amp; harga &nbsp; makanan."
	expected := "Pada Bulan Agustus Tahun 2026 terjadi inflasi year on year (y-on-y) Kabupaten Bangka Barat sebesar 4,43 persen dengan Indeks Harga Konsumen (IHK) sebesar 107,44. Inflasi y-on-y terjadi karena kenaikan indeks & harga makanan."

	res := CleanHTML(input)
	if res != expected {
		t.Errorf("CleanHTML mismatch.\nGot:      %q\nExpected: %q", res, expected)
	}
}

func TestSummarizeText(t *testing.T) {
	input := "Pada Bulan Agustus Tahun 2026 terjadi inflasi <i>year on year (y-on-y)</i> Kabupaten Bangka Barat sebesar 4,43 persen dengan Indeks Harga Konsumen (IHK) sebesar 107,44. \nInflasi y-on-y terjadi karena kenaikan indeks di seluruh kelompok pengeluaran, yaitu makanan 6,80 persen."
	res := SummarizeText(input, 220)

	expected := "Pada Bulan Agustus Tahun 2026 terjadi inflasi year on year (y-on-y) Kabupaten Bangka Barat sebesar 4,43 persen dengan Indeks Harga Konsumen (IHK) sebesar 107,44."
	if res != expected {
		t.Errorf("SummarizeText mismatch.\nGot:      %q\nExpected: %q", res, expected)
	}
}
