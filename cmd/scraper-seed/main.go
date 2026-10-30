package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"bot-istimewa/internal/config"
	"bot-istimewa/internal/scraper"
)

const browserUA = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"

type BRSItem struct {
	ID        int    `json:"brs_id"`
	Title     string `json:"title"`
	Abstract  string `json:"abstract"`
	Date      string `json:"rl_date"`
	PDF       string `json:"pdf"`
	Size      string `json:"size"`
	Thumbnail string `json:"thumbnail"`
}

type PubItem struct {
	ID        string `json:"pub_id"`
	Title     string `json:"title"`
	Abstract  string `json:"abstract"`
	Date      string `json:"rl_date"`
	Cover     string `json:"cover"`
	PDF       string `json:"pdf"`
	Size      string `json:"size"`
}

type NewsItem struct {
	ID        int    `json:"news_id"`
	Title     string `json:"title"`
	News      string `json:"news"`
	Date      string `json:"rl_date"`
	Picture   string `json:"picture"`
}

func main() {
	cfg := config.LoadConfig()
	store, err := scraper.NewStore(cfg.ScraperDBPath)
	if err != nil {
		log.Fatalf("Error opening store: %v", err)
	}
	defer store.Close()

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}

	ctx := context.Background()

	fmt.Println("[SEED SCRAPER] Mengambil data langsung dari Next.js Server Action BPS Bangka Barat...")

	// 1. Scrape Press Releases
	scrapeBRS(ctx, client, store)

	// 2. Scrape Publications
	scrapePubs(ctx, client, store)

	// 3. Scrape News
	scrapeNews(ctx, client, store)

	// Print summary
	counts, _ := store.GetCountsByKind(ctx)
	fmt.Println("\n[DATABASE HASIL SCRAPING LIVE BPS]:")
	for k, count := range counts {
		fmt.Printf(" - %-15s: %d items aktif\n", k, count)
	}

	// Print some sample items from DB
	fmt.Println("\n[CONTOH DATA TERSIMPAN DI STORAGE/SCRAPER.DB]:")
	for _, kind := range []scraper.Kind{scraper.KindPressRelease, scraper.KindPublication, scraper.KindNews} {
		items, _ := store.GetLatestItems(ctx, kind, "id", 2)
		for _, it := range items {
			fmt.Printf(" [%s] %s (%s)\n  -> File: %s\n", it.Kind, it.Title, it.ReleaseDate.Format("2006-01-02"), it.FileURL)
		}
	}
}

func extractJSONBlock(resp string, prefix string) []byte {
	start := strings.Index(resp, prefix)
	if start == -1 {
		return nil
	}
	// Try backward trimming until valid JSON is parsed
	for end := len(resp); end > start+10; end-- {
		sub := resp[start:end]
		if sub[len(sub)-1] == ']' {
			var check []interface{}
			if err := json.Unmarshal([]byte(sub), &check); err == nil {
				return []byte(sub)
			}
		}
	}
	return nil
}

func scrapeBRS(ctx context.Context, client *http.Client, store *scraper.Store) {
	fmt.Println("\n1. Mengambil Berita Resmi Statistik (BRS)...")
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://bangkabaratkab.bps.go.id/id/pressrelease", nil)
	req.Header.Set("User-Agent", browserUA)
	if resp, err := client.Do(req); err == nil && resp != nil {
		resp.Body.Close()
	}

	postReq, _ := http.NewRequestWithContext(ctx, "POST", "https://bangkabaratkab.bps.go.id/id/pressrelease", strings.NewReader(`["id",{"page":1}]`))
	postReq.Header.Set("User-Agent", browserUA)
	postReq.Header.Set("Next-Action", "60faa5fc36a0363681ae70a99e757ca04fc85e382f")
	postReq.Header.Set("Content-Type", "text/plain;charset=UTF-8")

	pResp, err := client.Do(postReq)
	if err != nil {
		log.Printf("POST pressrelease error: %v", err)
		return
	}
	defer pResp.Body.Close()

	body, _ := io.ReadAll(pResp.Body)
	jsonBlock := extractJSONBlock(string(body), `[{"brs_id":`)
	if jsonBlock == nil {
		log.Printf("BRS array block not found")
		return
	}

	var items []BRSItem
	if err := json.Unmarshal(jsonBlock, &items); err != nil {
		log.Printf("Unmarshal BRS error: %v", err)
		return
	}

	fmt.Printf("   Ditemukan %d rilis BRS resmi.\n", len(items))
	var scrapedItems []scraper.Item
	for _, it := range items {
		t, _ := time.Parse("2006-01-02", it.Date)
		itemURL := fmt.Sprintf("https://bangkabaratkab.bps.go.id/id/pressrelease/%s/%d/%s.html",
			strings.ReplaceAll(it.Date, "-", "/"), it.ID, url.PathEscape(it.Title))

		scraped := scraper.Item{
			Kind:         scraper.KindPressRelease,
			Lang:         "id",
			SourceKey:    fmt.Sprintf("%d", it.ID),
			URL:          itemURL,
			Title:        it.Title,
			Abstract:     cleanHTML(it.Abstract),
			ReleaseDate:  &t,
			FileURL:      it.PDF,
			FileSizeText: it.Size,
			ImageURL:     it.Thumbnail,
			ThumbnailURL: it.Thumbnail,
			IsActive:     true,
		}
		scraped.ContentHash = scraper.CalculateContentHash(&scraped)
		scrapedItems = append(scrapedItems, scraped)
	}

	n, u, err := store.ApplyItems(ctx, scraper.KindPressRelease, "id", scrapedItems, true)
	fmt.Printf("   Berhasil disimpan: %d baru, %d diperbarui (err=%v)\n", n, u, err)
}

func scrapePubs(ctx context.Context, client *http.Client, store *scraper.Store) {
	fmt.Println("\n2. Mengambil Publikasi Daerah...")
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://bangkabaratkab.bps.go.id/id/publication", nil)
	req.Header.Set("User-Agent", browserUA)
	if resp, err := client.Do(req); err == nil && resp != nil {
		resp.Body.Close()
	}

	postReq, _ := http.NewRequestWithContext(ctx, "POST", "https://bangkabaratkab.bps.go.id/id/publication", strings.NewReader(`["id",{"page":1}]`))
	postReq.Header.Set("User-Agent", browserUA)
	postReq.Header.Set("Next-Action", "60624e3a7c825a0b98392f2120a4870bd26b8f90f2")
	postReq.Header.Set("Content-Type", "text/plain;charset=UTF-8")

	pResp, err := client.Do(postReq)
	if err != nil {
		log.Printf("POST publication error: %v", err)
		return
	}
	defer pResp.Body.Close()

	body, _ := io.ReadAll(pResp.Body)
	jsonBlock := extractJSONBlock(string(body), `[{"pub_id":`)
	if jsonBlock == nil {
		log.Printf("Publication array block not found")
		return
	}

	var items []PubItem
	if err := json.Unmarshal(jsonBlock, &items); err != nil {
		log.Printf("Unmarshal Pubs error: %v", err)
		return
	}

	fmt.Printf("   Ditemukan %d publikasi resmi.\n", len(items))
	var scrapedItems []scraper.Item
	for _, it := range items {
		t, _ := time.Parse("2006-01-02", it.Date)
		itemURL := fmt.Sprintf("https://bangkabaratkab.bps.go.id/id/publication/%s/%s.html",
			strings.ReplaceAll(it.Date, "-", "/"), it.ID)

		scraped := scraper.Item{
			Kind:         scraper.KindPublication,
			Lang:         "id",
			SourceKey:    it.ID,
			URL:          itemURL,
			Title:        it.Title,
			Abstract:     cleanHTML(it.Abstract),
			ReleaseDate:  &t,
			FileURL:      it.PDF,
			FileSizeText: it.Size,
			ImageURL:     it.Cover,
			ThumbnailURL: it.Cover,
			IsActive:     true,
		}
		scraped.ContentHash = scraper.CalculateContentHash(&scraped)
		scrapedItems = append(scrapedItems, scraped)
	}

	n, u, err := store.ApplyItems(ctx, scraper.KindPublication, "id", scrapedItems, true)
	fmt.Printf("   Berhasil disimpan: %d baru, %d diperbarui (err=%v)\n", n, u, err)
}

func scrapeNews(ctx context.Context, client *http.Client, store *scraper.Store) {
	fmt.Println("\n3. Mengambil Berita Kegiatan BPS...")
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://bangkabaratkab.bps.go.id/id/news", nil)
	req.Header.Set("User-Agent", browserUA)
	if resp, err := client.Do(req); err == nil && resp != nil {
		resp.Body.Close()
	}

	postReq, _ := http.NewRequestWithContext(ctx, "POST", "https://bangkabaratkab.bps.go.id/id/news", strings.NewReader(`["id",{"page":1}]`))
	postReq.Header.Set("User-Agent", browserUA)
	postReq.Header.Set("Next-Action", "608404231d97eefc336168319bc8c592791e348afc")
	postReq.Header.Set("Content-Type", "text/plain;charset=UTF-8")

	pResp, err := client.Do(postReq)
	if err != nil {
		log.Printf("POST news error: %v", err)
		return
	}
	defer pResp.Body.Close()

	body, _ := io.ReadAll(pResp.Body)
	jsonBlock := extractJSONBlock(string(body), `[{"news_id":`)
	if jsonBlock == nil {
		log.Printf("News array block not found")
		return
	}

	var items []NewsItem
	if err := json.Unmarshal(jsonBlock, &items); err != nil {
		log.Printf("Unmarshal News error: %v", err)
		return
	}

	fmt.Printf("   Ditemukan %d berita resmi.\n", len(items))
	var scrapedItems []scraper.Item
	for _, it := range items {
		t, _ := time.Parse("2006-01-02", it.Date)
		itemURL := fmt.Sprintf("https://bangkabaratkab.bps.go.id/id/news/%s/%d/%s.html",
			strings.ReplaceAll(it.Date, "-", "/"), it.ID, url.PathEscape(it.Title))

		scraped := scraper.Item{
			Kind:         scraper.KindNews,
			Lang:         "id",
			SourceKey:    fmt.Sprintf("%d", it.ID),
			URL:          itemURL,
			Title:        it.Title,
			Abstract:     cleanHTML(it.News),
			ReleaseDate:  &t,
			ImageURL:     it.Picture,
			ThumbnailURL: it.Picture,
			IsActive:     true,
		}
		scraped.ContentHash = scraper.CalculateContentHash(&scraped)
		scrapedItems = append(scrapedItems, scraped)
	}

	n, u, err := store.ApplyItems(ctx, scraper.KindNews, "id", scrapedItems, true)
	fmt.Printf("   Berhasil disimpan: %d baru, %d diperbarui (err=%v)\n", n, u, err)
}

func cleanHTML(s string) string {
	return scraper.CleanHTML(s)
}
