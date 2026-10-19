package gsheet

import (
	"fmt"
	"strings"
)

// SearchAll scans cached data and built-in topics across Layanan Data, Indikator, Video, Infografik, FAQ, and BPS Catalog
func (c *Client) SearchAll(query string) ([]SearchResultItem, error) {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil, nil
	}

	words := strings.Fields(q)
	var results []SearchResultItem
	seen := make(map[string]bool)

	addResult := func(item SearchResultItem) {
		key := strings.ToLower(item.Title + "|" + item.Link)
		if !seen[key] {
			seen[key] = true
			results = append(results, item)
		}
	}

	matches := func(texts ...string) bool {
		combined := strings.ToLower(strings.Join(texts, " "))
		if strings.Contains(combined, q) {
			return true
		}
		for _, w := range words {
			if len(w) >= 3 && strings.Contains(combined, w) {
				return true
			}
		}
		return false
	}

	// 1. Search in Indikator Strategis (High priority)
	if indikators, err := c.FetchIndikatorRaw(); err == nil {
		for _, item := range indikators {
			if matches(item.Indikator, item.Penjelasan, item.Penjelasan2024, item.Penjelasan2025, item.Nilai, item.Nilai2024, item.Nilai2025) {
				desc := item.GetPenjelasan(c.cfg.ActiveIndicatorYear)
				if val := item.GetNilai(c.cfg.ActiveIndicatorYear); val != "" {
					desc = fmt.Sprintf("Nilai: %s. %s", val, desc)
				}
				addResult(SearchResultItem{
					Category:    "Indikator Strategis",
					Title:       item.Indikator,
					Description: desc,
					Link:        item.Infografis,
				})
				if len(results) >= 10 {
					return results, nil
				}
			}
		}
	}

	// 2. Search in Topics Catalog (Official BPS Bangka Barat Subjects)
	for _, topic := range DefaultTopicsCatalog() {
		if matches(topic.Title, topic.Description, topic.Category) {
			addResult(topic)
			if len(results) >= 10 {
				return results, nil
			}
		}
	}

	// 3. Search in Layanan Data
	if layanans, err := c.FetchLayananDataRaw(); err == nil {
		for _, item := range layanans {
			if matches(item.Judul) {
				addResult(SearchResultItem{
					Category:    "Layanan Data",
					Title:       item.Judul,
					Description: "Tabel Data & Publikasi BPS Bangka Barat",
					Link:        item.Link,
				})
				if len(results) >= 10 {
					return results, nil
				}
			}
		}
	}

	// 4. Search in Video Literasi
	if videos, err := c.FetchVideoRaw(); err == nil {
		for _, item := range videos {
			if matches(item.Judul, item.Penjelasan) {
				addResult(SearchResultItem{
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

	// 5. Search in Infografik
	if infografiks, err := c.FetchInfografikRaw(); err == nil {
		for _, item := range infografiks {
			if matches(item.Judul) {
				addResult(SearchResultItem{
					Category:    "Infografik",
					Title:       item.Judul,
					Description: "Visualisasi Data & Galeri BPS",
					Link:        item.Link,
				})
				if len(results) >= 10 {
					return results, nil
				}
			}
		}
	}

	// 6. Search in FAQs
	if faqs, err := c.FetchFAQRaw(); err == nil {
		for _, item := range faqs {
			if matches(item.Pertanyaan, item.Jawaban) {
				addResult(SearchResultItem{
					Category:    "Tanya Jawab (FAQ)",
					Title:       item.Pertanyaan,
					Description: item.Jawaban,
					Link:        "https://bangkabaratkab.bps.go.id",
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
		return fmt.Sprintf("🔍 Hasil pencarian untuk kata kunci *\"%s\"*:\n\n_Mohon maaf sobat data, data dengan kata kunci tersebut tidak ditemukan._\n\nSilakan coba dengan kata kunci lain (contoh: *kemiskinan*, *inflasi*, *penduduk*, *pdrb*).", query)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔍 Hasil pencarian untuk kata kunci *\"%s\"* (%d hasil ditemukan):\n\n", query, len(results)))

	for i, r := range results {
		sb.WriteString(fmt.Sprintf("*%d. [%s] %s*\n", i+1, r.Category, r.Title))
		if r.Description != "" {
			desc := r.Description
			if len(desc) > 140 {
				desc = desc[:137] + "..."
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
