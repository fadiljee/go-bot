package gsheet

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"bot-istimewa/internal/config"
)

type CacheEntry struct {
	Data      any
	ExpiresAt time.Time
}

type Client struct {
	cfg        *config.Config
	httpClient *http.Client
	cacheMu    sync.RWMutex
	cache      map[string]*CacheEntry
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 12 * time.Second,
		},
		cache: make(map[string]*CacheEntry),
	}
}

// Prewarm warms up the cache for all 6 sheets in background
func (c *Client) Prewarm() {
	go func() {
		log.Println("[CACHE] Pre-warming Google Sheets cache in background...")
		_ = c.RefreshAll()
		log.Println("[CACHE] Google Sheets pre-warming complete.")
	}()
}

// RefreshAll clears and re-fetches all sheets
func (c *Client) RefreshAll() error {
	var errs []string
	if _, err := c.FetchLayananDataRaw(); err != nil {
		errs = append(errs, fmt.Sprintf("layanan_data: %v", err))
	}
	if _, err := c.FetchIndikatorRaw(); err != nil {
		errs = append(errs, fmt.Sprintf("indikator: %v", err))
	}
	if _, err := c.FetchVideoRaw(); err != nil {
		errs = append(errs, fmt.Sprintf("video: %v", err))
	}
	if _, err := c.FetchInfografikRaw(); err != nil {
		errs = append(errs, fmt.Sprintf("infografik: %v", err))
	}
	if _, err := c.FetchPakarRaw(); err != nil {
		errs = append(errs, fmt.Sprintf("pakar: %v", err))
	}
	if _, err := c.FetchFAQRaw(); err != nil {
		errs = append(errs, fmt.Sprintf("faq: %v", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("some sheets failed to refresh: %s", strings.Join(errs, "; "))
	}
	return nil
}

func (c *Client) getCached(key string) (any, bool) {
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()
	entry, ok := c.cache[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.ExpiresAt) {
		return entry.Data, false // Stale, but returned for fallback
	}
	return entry.Data, true
}

func (c *Client) setCached(key string, data any) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	c.cache[key] = &CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(time.Duration(c.cfg.GASCacheTTLMinutes) * time.Minute),
	}
}

type standardGASResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func fetchGAS(client *http.Client, url string, target any) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %s", resp.Status)
	}

	const maxBodyBytes = 16 * 1024 * 1024 // 16 MB
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return err
	}

	trimmedBody := strings.TrimSpace(string(body))
	if strings.HasPrefix(trimmedBody, "<") {
		return fmt.Errorf("Google Apps Script mengembalikan halaman HTML/Login. Pastikan deployment GAS disetel ke 'Who has access: Anyone (Siapa saja)'")
	}

	var res standardGASResponse
	if err := json.Unmarshal(body, &res); err != nil {
		// Try unmarshaling directly into target if not standard wrapper
		return json.Unmarshal(body, target)
	}

	return json.Unmarshal(res.Data, target)
}

// -------------------------------------------------------------
// 1. Layanan Data
// -------------------------------------------------------------

func (c *Client) FetchLayananDataRaw() ([]LayananDataItem, error) {
	const key = "layanan_data"
	if data, valid := c.getCached(key); valid {
		return data.([]LayananDataItem), nil
	}

	var items []LayananDataItem
	err := fetchGAS(c.httpClient, c.cfg.GASUrlLayananData, &items)
	if err != nil {
		if stale, ok := c.getCached(key); ok && stale != nil {
			log.Printf("[CACHE] GAS error on %s, serving stale cache: %v", key, err)
			return stale.([]LayananDataItem), nil
		}
		log.Printf("[FALLBACK] GAS error on %s, serving fallback dataset: %v", key, err)
		return DefaultLayananData(), nil
	}

	c.setCached(key, items)
	return items, nil
}

func (c *Client) GetLayananData(no string) (string, error) {
	items, err := c.FetchLayananDataRaw()
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, item := range items {
		if item.No.String() == no {
			sb.WriteString(fmt.Sprintf("📊 *Data* : \n%s\n\n🌐 *Link* : \n%s\n◪◩◪◩◪◩◪◩◪◩◪◩◪◩◪◩◪◩◪◩◪◩\n\n\n", item.Judul, item.Link))
		}
	}
	return sb.String(), nil
}

// -------------------------------------------------------------
// 2. Indikator Strategis
// -------------------------------------------------------------

func (c *Client) FetchIndikatorRaw() ([]IndikatorItem, error) {
	const key = "indikator"
	if data, valid := c.getCached(key); valid {
		return data.([]IndikatorItem), nil
	}

	var items []IndikatorItem
	err := fetchGAS(c.httpClient, c.cfg.GASUrlIndikator, &items)
	if err != nil {
		if stale, ok := c.getCached(key); ok && stale != nil {
			log.Printf("[CACHE] GAS error on %s, serving stale cache: %v", key, err)
			return stale.([]IndikatorItem), nil
		}
		log.Printf("[FALLBACK] GAS error on %s, serving fallback dataset: %v", key, err)
		return DefaultIndikators(), nil
	}

	c.setCached(key, items)
	return items, nil
}

func (c *Client) GetIndikator(no, preferredYear string) (string, string, error) {
	items, err := c.FetchIndikatorRaw()
	if err != nil {
		return "", "", err
	}

	var sb strings.Builder
	var imgUrl string
	for _, item := range items {
		if item.No.String() == no {
			nilai := item.GetNilai(preferredYear)
			penjelasan := item.GetPenjelasan(preferredYear)
			sb.WriteString(fmt.Sprintf("📝 *%s*\n\n_*%s*_\n\n%s\n◉◎◉◎◉◎◉◎◉◎◉◎◉◎◉◎◉◎◉◎◉◎◉◎◉◎◉\n\n", item.Indikator, nilai, penjelasan))
			if imgUrl == "" && item.Infografis != "" {
				imgUrl = item.Infografis
			}
		}
	}
	return sb.String(), imgUrl, nil
}

// -------------------------------------------------------------
// 3. Video Literasi Statistik
// -------------------------------------------------------------

func (c *Client) FetchVideoRaw() ([]VideoItem, error) {
	const key = "video"
	if data, valid := c.getCached(key); valid {
		return data.([]VideoItem), nil
	}

	var items []VideoItem
	err := fetchGAS(c.httpClient, c.cfg.GASUrlVideo, &items)
	if err != nil {
		if stale, ok := c.getCached(key); ok && stale != nil {
			log.Printf("[CACHE] GAS error on %s, serving stale cache: %v", key, err)
			return stale.([]VideoItem), nil
		}
		log.Printf("[FALLBACK] GAS error on %s, serving fallback dataset: %v", key, err)
		return DefaultVideos(), nil
	}

	c.setCached(key, items)
	return items, nil
}

func (c *Client) GetVideoStat(no string) (string, error) {
	items, err := c.FetchVideoRaw()
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, item := range items {
		if item.No.String() == no {
			sb.WriteString(fmt.Sprintf("🎥 *%s*\n\n🌐 *Link* : \n%s\n\n%s\n\n🔘🔘🔘🔘🔘⚪🔘🔘🔘🔘🔘\n\n", item.Judul, item.Link, item.Penjelasan))
		}
	}
	return sb.String(), nil
}

// -------------------------------------------------------------
// 4. Infografik
// -------------------------------------------------------------

func (c *Client) FetchInfografikRaw() ([]InfografikItem, error) {
	const key = "infografik"
	if data, valid := c.getCached(key); valid {
		return data.([]InfografikItem), nil
	}

	var items []InfografikItem
	err := fetchGAS(c.httpClient, c.cfg.GASUrlInfografik, &items)
	if err != nil {
		if stale, ok := c.getCached(key); ok && stale != nil {
			log.Printf("[CACHE] GAS error on %s, serving stale cache: %v", key, err)
			return stale.([]InfografikItem), nil
		}
		log.Printf("[FALLBACK] GAS error on %s, serving fallback dataset: %v", key, err)
		return DefaultInfografiks(), nil
	}

	c.setCached(key, items)
	return items, nil
}

func (c *Client) GetInfografik(no string) ([]string, []string, error) {
	items, err := c.FetchInfografikRaw()
	if err != nil {
		return nil, nil, err
	}

	var titles []string
	var urls []string
	for _, item := range items {
		if item.No.String() == no {
			titles = append(titles, item.Judul)
			urls = append(urls, item.Link)
		}
	}
	return titles, urls, nil
}

// -------------------------------------------------------------
// 5. Pakar & Konsultasi
// -------------------------------------------------------------

func (c *Client) FetchPakarRaw() ([]PakarItem, error) {
	const key = "pakar"
	if data, valid := c.getCached(key); valid {
		return data.([]PakarItem), nil
	}

	var items []PakarItem
	err := fetchGAS(c.httpClient, c.cfg.GASUrlPakar, &items)
	if err != nil {
		if stale, ok := c.getCached(key); ok && stale != nil {
			log.Printf("[CACHE] GAS error on %s, serving stale cache: %v", key, err)
			return stale.([]PakarItem), nil
		}
		return nil, err
	}

	c.setCached(key, items)
	return items, nil
}

func (c *Client) GetPakar(no string) (pakarText, pakarImg, piketText, piketImg string, err error) {
	items, err := c.FetchPakarRaw()
	if err != nil {
		return "", "", "", "", err
	}

	for _, element := range items {
		if element.No.String() == no {
			pakarText = fmt.Sprintf("🤝 Perkenalkan Sahabat Data, ini adalah %s\n\n%s\n\nBeliau akan membantu menjawab pertanyaan Sahabat Data.\n\nMohon diingat untuk menghubungi beliau hanya pada waktu kerja dan hari kerja yah sobat data 🙆‍♀️\n\n Silakan klik link berikut.\n\n📞 %s", element.Nama, element.Penjelasan, element.Link)
			pakarImg = element.Foto

			piketText = fmt.Sprintf("🤝 Perkenalkan Sahabat Data, ini adalah %s\n\nDia yang akan membantu menjawab pertanyaan Sahabat terkait Data dan Statistik.\n\nDimohon untuk menghubungi petugas hanya pada waktu kerja dan hari kerja yah sobat data 🙆‍♀️\n\n Silakan klik link berikut.\n\n📞 %s", element.PetugasPiket, element.LinkPetugas)
			piketImg = element.FotoPetugas
			break
		}
	}
	return
}

// -------------------------------------------------------------
// 6. FAQs
// -------------------------------------------------------------

func (c *Client) FetchFAQRaw() ([]FAQItem, error) {
	const key = "faq"
	if data, valid := c.getCached(key); valid {
		return data.([]FAQItem), nil
	}

	var items []FAQItem
	err := fetchGAS(c.httpClient, c.cfg.GASUrlFAQ, &items)
	if err != nil {
		if stale, ok := c.getCached(key); ok && stale != nil {
			log.Printf("[CACHE] GAS error on %s, serving stale cache: %v", key, err)
			return stale.([]FAQItem), nil
		}
		log.Printf("[FALLBACK] GAS error on %s, serving fallback dataset: %v", key, err)
		return DefaultFAQs(), nil
	}

	c.setCached(key, items)
	return items, nil
}

func (c *Client) GetFAQs(no string) (string, error) {
	items, err := c.FetchFAQRaw()
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, element := range items {
		if element.No.String() == no || element.DisplayNo.String() == no {
			disp := element.DisplayNo.String()
			if disp == "" {
				disp = element.No.String()
			}
			sb.WriteString(fmt.Sprintf("*%s* 🤔 *: %s*\n\n 👩‍🏫 : %s\n◪◩◪◩◪◩◪◩◪◩◪◩◪◩◪◩◪◩◪◩◪◩\n\n\n", disp, element.Pertanyaan, element.Jawaban))
		}
	}
	return sb.String(), nil
}
