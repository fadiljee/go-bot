package gsheet

import (
	"testing"

	"bot-istimewa/internal/config"
)

func TestSearchAllKeywords(t *testing.T) {
	cfg := &config.Config{
		ActiveIndicatorYear: "2025",
		GASCacheTTLMinutes:  15,
	}
	client := NewClient(cfg)

	tests := []struct {
		keyword  string
		minCount int
	}{
		{"kemiskinan", 1},
		{"inflasi", 1},
		{"penduduk", 1},
		{"pdrb", 1},
		{"sawit", 1},
		{"tenaga kerja", 1},
		{"kesehatan", 1},
	}

	for _, tt := range tests {
		t.Run(tt.keyword, func(t *testing.T) {
			results, err := client.SearchAll(tt.keyword)
			if err != nil {
				t.Fatalf("SearchAll(%q) returned error: %v", tt.keyword, err)
			}
			if len(results) < tt.minCount {
				t.Errorf("SearchAll(%q) expected at least %d results, got %d", tt.keyword, tt.minCount, len(results))
			}
			formatted := FormatSearchResults(tt.keyword, results)
			if formatted == "" {
				t.Errorf("FormatSearchResults(%q) returned empty string", tt.keyword)
			}
		})
	}
}
