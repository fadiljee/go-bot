package gsheet

import (
	"encoding/json"
	"fmt"
)

// FlexString handles fields that could be int, float, or string in JSON
type FlexString string

func (f *FlexString) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		*f = FlexString(s)
		return nil
	}
	var num float64
	if err := json.Unmarshal(b, &num); err == nil {
		*f = FlexString(fmt.Sprintf("%v", int64(num)))
		return nil
	}
	*f = ""
	return nil
}

func (f FlexString) String() string {
	return string(f)
}

// LayananDataItem represents an item in Layanan Data
type LayananDataItem struct {
	No    FlexString `json:"no"`
	Judul string     `json:"judul"`
	Link  string     `json:"link"`
}

// IndikatorItem represents an item in Indikator Strategis
// Handles nilai/nilai_2024/nilai_2025 dynamically (Fix TD-05)
type IndikatorItem struct {
	No             FlexString `json:"no"`
	Indikator      string     `json:"indikator"`
	Nilai          string     `json:"nilai"`
	Nilai2024      string     `json:"nilai_2024"`
	Nilai2025      string     `json:"nilai_2025"`
	Penjelasan     string     `json:"penjelasan"`
	Penjelasan2024 string     `json:"penjelasan_2024"`
	Penjelasan2025 string     `json:"penjelasan_2025"`
	Infografis     string     `json:"infografis"`
}

func (i *IndikatorItem) GetNilai(preferredYear string) string {
	if preferredYear == "2025" && i.Nilai2025 != "" {
		return i.Nilai2025
	}
	if i.Nilai != "" {
		return i.Nilai
	}
	if i.Nilai2024 != "" {
		return i.Nilai2024
	}
	return i.Nilai2025
}

func (i *IndikatorItem) GetPenjelasan(preferredYear string) string {
	if preferredYear == "2025" && i.Penjelasan2025 != "" {
		return i.Penjelasan2025
	}
	if i.Penjelasan != "" {
		return i.Penjelasan
	}
	if i.Penjelasan2024 != "" {
		return i.Penjelasan2024
	}
	return i.Penjelasan2025
}

// VideoItem represents an item in Video Literasi Statistik
type VideoItem struct {
	No         FlexString `json:"no"`
	Judul      string     `json:"judul"`
	Link       string     `json:"link"`
	Penjelasan string     `json:"penjelasan"`
}

// InfografikItem represents an item in Infografik Statistik
type InfografikItem struct {
	No    FlexString `json:"no"`
	Judul string     `json:"judul"`
	Link  string     `json:"link"`
}

// PakarItem represents consultation experts and piket officers
type PakarItem struct {
	No           FlexString `json:"no"`
	Nama         string     `json:"nama"`
	Penjelasan   string     `json:"penjelasan"`
	Link         string     `json:"link"`
	Foto         string     `json:"foto"`
	PetugasPiket string     `json:"petugas_piket"`
	LinkPetugas  string     `json:"link_petugas"`
	FotoPetugas  string     `json:"foto_petugas"`
}

// FAQItem represents an FAQ question and answer
type FAQItem struct {
	No         FlexString `json:"no"`
	DisplayNo  FlexString `json:"No"`
	Pertanyaan string     `json:"pertanyaan"`
	Jawaban    string     `json:"jawaban"`
}

// SearchResultItem represents cross-sheet search result item (Menu 1.5)
type SearchResultItem struct {
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
}
