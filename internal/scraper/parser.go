package scraper

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

// Parser separates fetching from HTML parsing (§6.3).
type Parser interface {
	Kind() Kind
	ListURL(base *url.URL, lang string, page int) string
	ParseList(body []byte, pageURL *url.URL) ([]Item, Pagination, error)
	ParseDetail(body []byte, pageURL *url.URL, it *Item) error
}


// GenericKindParser implements Parser driven by KindSelector from YAML (§8.2, §8.3).
type GenericKindParser struct {
	kind      Kind
	selectors KindSelector
}

// NewGenericKindParser creates a new Parser for a specific Kind.
func NewGenericKindParser(kind Kind, selectors KindSelector) *GenericKindParser {
	return &GenericKindParser{
		kind:      kind,
		selectors: selectors,
	}
}

func (p *GenericKindParser) Kind() Kind {
	return p.kind
}

func (p *GenericKindParser) ListURL(base *url.URL, lang string, page int) string {
	if lang == "" {
		lang = "id"
	}
	tpl := p.selectors.ListURL
	tpl = strings.ReplaceAll(tpl, "{lang}", lang)
	tpl = strings.ReplaceAll(tpl, "{page}", strconv.Itoa(page))

	rel, err := url.Parse(tpl)
	if err != nil {
		return base.String()
	}
	return base.ResolveReference(rel).String()
}

func (p *GenericKindParser) ParseList(body []byte, pageURL *url.URL) (items []Item, pagination Pagination, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("parser %s panic: %v", p.kind, r)
		}
	}()

	doc, err := makeDoc(body)
	if err != nil {
		return nil, pagination, err
	}

	pagination.Page = 1
	if q := pageURL.Query().Get("page"); q != "" {
		if pNum, err := strconv.Atoi(q); err == nil && pNum > 0 {
			pagination.Page = pNum
		}
	}

	// Check next page indicator
	if p.selectors.List.NextPage != "" {
		nextSel := doc.Find(p.selectors.List.NextPage)
		if nextSel.Length() > 0 {
			pagination.HasNext = true
		}
	}

	// Find items
	doc.Find(p.selectors.List.Item).Each(func(i int, s *goquery.Selection) {
		title := ""
		if p.selectors.List.Title != "" {
			title = CleanText(s.Find(p.selectors.List.Title).First().Text())
		}
		if title == "" {
			// Fallback: s.Text() if no specific title selector or empty
			title = CleanText(s.Find("a").First().Text())
		}

		rawHref := ""
		if p.selectors.List.Link != "" {
			rawHref, _ = s.Find(p.selectors.List.Link).First().Attr("href")
		}
		if rawHref == "" {
			rawHref, _ = s.Find("a").First().Attr("href")
		}

		if rawHref == "" {
			return
		}

		itemURL := resolveURL(pageURL, rawHref)
		sourceKey := extractSourceKey(itemURL)

		dateText := ""
		if p.selectors.List.Date != "" {
			dateText = CleanText(s.Find(p.selectors.List.Date).First().Text())
		}
		releaseDate, _ := ParseDate(dateText, itemURL)

		it := Item{
			Kind:        p.kind,
			Lang:        detectLang(pageURL),
			SourceKey:   sourceKey,
			Title:       title,
			URL:         itemURL,
			ReleaseDate: releaseDate,
			IsActive:    true,
		}
		it.ContentHash = CalculateContentHash(&it)
		items = append(items, it)
	})

	return items, pagination, nil
}

func (p *GenericKindParser) ParseDetail(body []byte, pageURL *url.URL, it *Item) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("parser %s detail panic: %v", p.kind, r)
		}
	}()

	doc, err := makeDoc(body)
	if err != nil {
		return err
	}

	// 1. Abstract
	if p.selectors.Detail.Abstract != "" {
		abs := CleanText(doc.Find(p.selectors.Detail.Abstract).First().Text())
		if len([]rune(abs)) > 600 {
			abs = string([]rune(abs)[:600]) + "…"
		}
		if abs != "" {
			it.Abstract = abs
		}
	}

	// 2. Release Date (if missing or more accurate in detail)
	if p.selectors.Detail.ReleaseDate != "" {
		dtText := CleanText(doc.Find(p.selectors.Detail.ReleaseDate).First().Text())
		if parsed, err := ParseDate(dtText, it.URL); err == nil && parsed != nil {
			it.ReleaseDate = parsed
		}
	}

	// 3. File Size
	if p.selectors.Detail.FileSize != "" {
		sizeText := CleanText(doc.Find(p.selectors.Detail.FileSize).First().Text())
		if sizeText != "" {
			it.FileSizeText = sizeText
		}
	}

	// 4. File Link
	if p.selectors.Detail.FileLink != "" {
		if href, ok := doc.Find(p.selectors.Detail.FileLink).First().Attr("href"); ok && href != "" {
			fURL := resolveURL(pageURL, href)
			it.FileURL = fURL
			it.FileType = detectFileType(fURL)
		}
	}

	// 5. Image Link
	if p.selectors.Detail.ImageLink != "" {
		imgSel := doc.Find(p.selectors.Detail.ImageLink).First()
		src, _ := imgSel.Attr("src")
		if src == "" {
			src, _ = imgSel.Attr("data-src")
		}
		if src != "" {
			it.ImageURL = resolveURL(pageURL, src)
		}
	}

	it.ContentHash = CalculateContentHash(it)
	return nil
}

func makeDoc(body []byte) (*goquery.Document, error) {
	r, err := charset.NewReader(bytes.NewReader(body), "text/html")
	if err != nil {
		r = bytes.NewReader(body)
	}
	return goquery.NewDocumentFromReader(r)
}

func resolveURL(base *url.URL, ref string) string {
	ref = strings.TrimSpace(ref)
	// Remove fragment
	if idx := strings.Index(ref, "#"); idx != -1 {
		ref = ref[:idx]
	}
	u, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return base.ResolveReference(u).String()
}

var dateIDInURLRe = regexp.MustCompile(`/\d{4}/\d{1,2}/\d{1,2}/(\d+)(?:/|$)`)
var genericIDInURLRe = regexp.MustCompile(`/(\d+)(?:/[^/]*\.html|/|$)`)

func extractSourceKey(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err == nil {
		// Pattern F-02: /{YYYY}/{MM}/{DD}/{ID}/...
		if m := dateIDInURLRe.FindStringSubmatch(u.Path); len(m) == 2 {
			return m[1]
		}
		// Generic segment ID
		matches := genericIDInURLRe.FindAllStringSubmatch(u.Path, -1)
		if len(matches) > 0 {
			// Take the last numeric segment
			last := matches[len(matches)-1]
			if len(last) == 2 {
				return last[1]
			}
		}
	}

	// Fallback (§8.5): sha1 of host + path
	h := sha1.New()
	if u != nil {
		h.Write([]byte(strings.ToLower(u.Host + u.Path)))
	} else {
		h.Write([]byte(strings.ToLower(rawURL)))
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func detectLang(u *url.URL) string {
	if strings.Contains(u.Path, "/en/") || strings.HasPrefix(u.Path, "/en") {
		return "en"
	}
	return "id"
}

func detectFileType(fileURL string) string {
	lower := strings.ToLower(fileURL)
	switch {
	case strings.HasSuffix(lower, ".pdf"):
		return "pdf"
	case strings.HasSuffix(lower, ".xlsx"):
		return "xlsx"
	case strings.HasSuffix(lower, ".xls"):
		return "xls"
	case strings.HasSuffix(lower, ".csv"):
		return "csv"
	case strings.HasSuffix(lower, ".png"):
		return "png"
	case strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg"):
		return "jpg"
	default:
		return "other"
	}
}
