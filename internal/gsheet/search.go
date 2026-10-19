package gsheet

import (
	"fmt"
	"strings"
)

// SearchAll scans cached data across Layanan Data, Indikator, Video, and Infografik
func (c *Client) SearchAll(query string) ([]SearchResultItem, error) {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil, nil
	}

	var results []SearchResultItem

	// 1. Search in Layanan Data
	if layanans, err := c.FetchLayananDataRaw(); err == nil {
		for _, item := range layanans {
			if strings.Contains(strings.ToLower(item.Judul), q) {
				results = append(results, SearchResultItem{
					Category:    "Layanan Data",
					Title:       item.Judul,
					Description: "Tabel / Publikasi BPS",
					Link:        item.Link,
				})
				if len(results) >= 10 {
					return results, nil
				}
			}
		}
	}

	// 2. Search in Indikator Strategis
	if indikators, err := c.FetchIndikatorRaw(); err == nil {
		for _, item := range indikators {
			if strings.Contains(strings.ToLower(item.Indikator), q) ||
				strings.Contains(strings.ToLower(item.Penjelasan), q) ||
				strings.Contains(strings.ToLower(item.Penjelasan2024), q) ||
				strings.Contains(strings.ToLower(item.Penjelasan2025), q) {
				results = append(results, SearchResultItem{
					Category:    "Indikator Strategis",
					Title:       item.Indikator,
					Description: item.GetPenjelasan(c.cfg.ActiveIndicatorYear),
					Link:        item.Infografis,
				})
				if len(results) >= 10 {
					return results, nil
				}
			}
		}
	}

	// 3. Search in Video
	if videos, err := c.FetchVideoRaw(); err == nil {
		for _, item := range videos {
			if strings.Contains(strings.ToLower(item.Judul), q) || strings.Contains(strings.ToLower(item.Penjelasan), q) {
				results = append(results, SearchResultItem{
					Category:    "Video Literasi",
					Title:       item.Judul,
					Description: item.Penjelasan,
					Link:        item.Link,
				})
				if len(results) >= 10 {
					return results, nil
				}
			}
		}
	}

	// 4. Search in Infografik
	if infografiks, err := c.FetchInfografikRaw(); err == nil {
		for _, item := range infografiks {
			if strings.Contains(strings.ToLower(item.Judul), q) {
				results = append(results, SearchResultItem{
					Category:    "Infografik",
					Title:       item.Judul,
					Description: "Visualisasi Data",
					Link:        item.Link,
				})
				if len(results) >= 10 {
					return results, nil
				}
			}
		}
	}

	return results, nil
}

// FormatSearchResults formats slice of SearchResultItem into WhatsApp readable text
func FormatSearchResults(query string, results []SearchResultItem) string {
	if len(results) == 0 {
		return fmt.Sprintf("🔍 Hasil pencarian untuk kata kunci *\"%s\"*:\n\n_Mohon maaf sobat data, data dengan kata kunci tersebut tidak ditemukan._\n\nSilakan coba dengan kata kunci lain (contoh: *kemiskinan*, *inflasi*, *penduduk*).", query)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔍 Hasil pencarian untuk kata kunci *\"%s\"* (%d hasil):\n\n", query, len(results)))

	for i, r := range results {
		sb.WriteString(fmt.Sprintf("*%d. [%s] %s*\n", i+1, r.Category, r.Title))
		if r.Description != "" {
			desc := r.Description
			if len(desc) > 120 {
				desc = desc[:117] + "..."
			}
			sb.WriteString(fmt.Sprintf("   _%s_\n", desc))
		}
		if r.Link != "" {
			sb.WriteString(fmt.Sprintf("   🌐 %s\n", r.Link))
		}
		sb.WriteString("◪◩◪◩◪◩◪◩◪◩◪◩◪◩◪◩\n\n")
	}

	return sb.String()
}
