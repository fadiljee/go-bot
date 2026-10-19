package gsheet

import (
	"encoding/json"
	"testing"
)

func TestFlexStringUnmarshaling(t *testing.T) {
	testJSON := `{
		"no": 12,
		"judul": "Jumlah Penduduk",
		"link": "https://bangwabaratkab.bps.go.id"
	}`

	var item LayananDataItem
	if err := json.Unmarshal([]byte(testJSON), &item); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if item.No.String() != "12" {
		t.Errorf("expected 12, got %s", item.No.String())
	}

	if item.Judul != "Jumlah Penduduk" {
		t.Errorf("expected 'Jumlah Penduduk', got %s", item.Judul)
	}
}

func TestIndikatorDynamicYear(t *testing.T) {
	item := IndikatorItem{
		Indikator:  "Kemiskinan",
		Nilai2025:  "4.41",
		Nilai2024:  "4.56",
		Nilai:      "4.50",
	}

	if val := item.GetNilai("2025"); val != "4.41" {
		t.Errorf("expected 2025 value 4.41, got %s", val)
	}

	if val := item.GetNilai("2024"); val != "4.50" {
		t.Errorf("expected fallback 4.50, got %s", val)
	}
}
