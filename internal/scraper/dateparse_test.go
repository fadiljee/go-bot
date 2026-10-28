package scraper

import (
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		fallbackURL string
		wantYear    int
		wantMonth   time.Month
		wantDay     int
		wantErr     bool
	}{
		// Indonesian full month
		{"ID full Jan", "15 Januari 2024", "", 2024, time.January, 15, false},
		{"ID full Feb", "28 Februari 2023", "", 2023, time.February, 28, false},
		{"ID full Mar", "01 Maret 2022", "", 2022, time.March, 1, false},
		{"ID full Apr", "10 April 2021", "", 2021, time.April, 10, false},
		{"ID full Mei", "2 Mei 2020", "", 2020, time.May, 2, false},
		{"ID full Jun", "30 Juni 2024", "", 2024, time.June, 30, false},
		{"ID full Jul", "17 Juli 2022", "", 2022, time.July, 17, false},
		{"ID full Agu", "17 Agustus 2023", "", 2023, time.August, 17, false},
		{"ID full Sep", "01 September 2021", "", 2021, time.September, 1, false},
		{"ID full Okt", "28 Oktober 2024", "", 2024, time.October, 28, false},
		{"ID full Nov", "10 November 2023", "", 2023, time.November, 10, false},
		{"ID full Nopember", "10 Nopember 2023", "", 2023, time.November, 10, false},
		{"ID full Des", "25 Desember 2022", "", 2022, time.December, 25, false},

		// Indonesian short month & variations
		{"ID short agu", "17 Agu 2024", "", 2024, time.August, 17, false},
		{"ID short agt", "17 Agt 2024", "", 2024, time.August, 17, false},
		{"ID short okt uppercase", "05 OKT 2023", "", 2023, time.October, 5, false},
		{"ID with NBSP and extra spaces", " 1 \u00a0 September   2021 ", "", 2021, time.September, 1, false},
		{"ID with day name prefix", "Senin, 01 September 2021", "", 2021, time.September, 1, false},

		// English formats
		{"EN Month Day Year", "September 1, 2021", "", 2021, time.September, 1, false},
		{"EN with ordinal suffix", "October 24th, 2023", "", 2023, time.October, 24, false},
		{"EN short month", "Aug 15, 2022", "", 2022, time.August, 15, false},

		// ISO format
		{"ISO format", "2024-05-12", "", 2024, time.May, 12, false},

		// Fallback from URL
		{"Empty text with URL fallback", "", "https://bangkabaratkab.bps.go.id/id/pressrelease/2021/09/01/537/slug.html", 2021, time.September, 1, false},
		{"Gibberish text with URL fallback", "jadwal rilis tbd", "https://bangkabaratkab.bps.go.id/en/pressrelease/2022/11/15/888/slug.html", 2022, time.November, 15, false},

		// Invalid dates
		{"Feb 31 invalid", "31 Februari 2023", "", 0, 0, 0, true},
		{"Year too old", "15 Januari 1980", "", 0, 0, 0, true},
		{"Future year too far", "15 Januari 2099", "", 0, 0, 0, true},
		{"Completely invalid string without URL", "tidak ada tanggal", "", 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := ParseDate(tt.text, tt.fallbackURL)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil result: %v", res)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res == nil {
				t.Fatalf("expected date, got nil")
			}
			if res.Year() != tt.wantYear || res.Month() != tt.wantMonth || res.Day() != tt.wantDay {
				t.Errorf("got %04d-%02d-%02d, want %04d-%02d-%02d",
					res.Year(), res.Month(), res.Day(), tt.wantYear, tt.wantMonth, tt.wantDay)
			}
			if res.Location().String() != "Asia/Jakarta" && res.Location().String() != "WIB" {
				t.Errorf("expected WIB or Asia/Jakarta timezone, got: %s", res.Location().String())
			}
		})
	}
}
