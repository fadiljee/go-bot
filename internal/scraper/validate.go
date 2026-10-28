package scraper

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

var (
	ErrInvalidItem           = errors.New("validate: item tidak valid")
	ErrAnomalyLayoutChanged  = errors.New("validate: layout berubah (0 item ditemukan)")
	ErrAnomalyCountDrop      = errors.New("validate: penurunan jumlah item drastis (>50%)")
	ErrAnomalyFieldLoss      = errors.New("validate: kehilangan field wajib (>30%)")
	ErrAnomalyMarkerMissing  = errors.New("validate: required marker halaman BPS tidak ditemukan")
)

var spaceCleanRe = regexp.MustCompile(`[\s\x{00a0}]+`)

// CleanText strips control characters, collapses multiple spaces and NBSP, and trims string (§8.2, Lampiran C).
func CleanText(s string) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, s)
	return strings.TrimSpace(spaceCleanRe.ReplaceAllString(s, " "))
}

// CalculateContentHash computes sha256 hash across canonical fields (§5.2).
func CalculateContentHash(it *Item) string {
	h := sha256.New()
	h.Write([]byte(it.Title))
	if it.ReleaseDate != nil {
		h.Write([]byte(it.ReleaseDate.Format("2006-01-02")))
	}
	h.Write([]byte(it.Abstract))
	h.Write([]byte(it.FileURL))
	h.Write([]byte(it.ImageURL))
	return hex.EncodeToString(h.Sum(nil))
}

// ValidateItem validates fields according to §8.6.
func ValidateItem(it *Item, allowedHosts []string, appEnv string) error {
	if it == nil {
		return errors.New("item is nil")
	}

	it.Title = CleanText(it.Title)
	if it.Title == "" {
		return fmt.Errorf("%w: title is empty", ErrInvalidItem)
	}
	if len([]rune(it.Title)) > 300 {
		return fmt.Errorf("%w: title exceeds 300 characters", ErrInvalidItem)
	}

	if strings.TrimSpace(it.SourceKey) == "" {
		return fmt.Errorf("%w: source_key is empty", ErrInvalidItem)
	}

	// Validate main URL
	if it.URL == "" {
		return fmt.Errorf("%w: url is empty", ErrInvalidItem)
	}
	u, err := url.Parse(it.URL)
	if err != nil {
		return fmt.Errorf("%w: invalid url: %v", ErrInvalidItem, err)
	}
	if appEnv != "test" && u.Scheme != "https" {
		return fmt.Errorf("%w: url scheme must be https", ErrInvalidItem)
	}
	if appEnv != "test" {
		if err := checkIPAntiSSRF(u.Hostname()); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidItem, err)
		}
	}

	// Release date required for pressrelease & publication (§5.2, §8.6)
	if it.Kind == KindPressRelease || it.Kind == KindPublication {
		if it.ReleaseDate == nil {
			return fmt.Errorf("%w: release_date is required for %s", ErrInvalidItem, it.Kind)
		}
	}

	// Validate FileURL if present
	if it.FileURL != "" {
		fu, err := url.Parse(it.FileURL)
		if err != nil || (appEnv != "test" && fu.Scheme != "https") || (appEnv != "test" && !isBpsDomain(fu.Hostname())) {
			// Clear invalid file_url but retain item (§7.8, §8.6)
			it.FileURL = ""
		}
	}

	// Validate ImageURL if present
	if it.ImageURL != "" {
		iu, err := url.Parse(it.ImageURL)
		if err != nil || (appEnv != "test" && iu.Scheme != "https") || (appEnv != "test" && !isBpsDomain(iu.Hostname())) {
			it.ImageURL = ""
		}
	}

	// Ensure content_hash is calculated
	if it.ContentHash == "" {
		it.ContentHash = CalculateContentHash(it)
	}

	return nil
}

// AnomalyGuard checks for drastic drops or layout breakage (§8.7).
func AnomalyGuard(kind Kind, currentCount int, previousCount int, htmlBody []byte, requiredMarkers []string) error {
	// 1. Check required markers if provided
	if len(requiredMarkers) > 0 && len(htmlBody) > 0 {
		lowerBody := strings.ToLower(string(htmlBody))
		for _, marker := range requiredMarkers {
			if !strings.Contains(lowerBody, strings.ToLower(marker)) {
				return fmt.Errorf("%w: marker %q not found", ErrAnomalyMarkerMissing, marker)
			}
		}
	}

	// 2. 0 items found when previously had >= 5
	if currentCount == 0 && previousCount >= 5 {
		return fmt.Errorf("%w: 0 items found for %s (previously %d)", ErrAnomalyLayoutChanged, kind, previousCount)
	}

	// 3. Count drop > 50%
	if previousCount >= 4 && currentCount < (previousCount/2) {
		return fmt.Errorf("%w: count dropped from %d to %d for %s", ErrAnomalyCountDrop, previousCount, currentCount, kind)
	}

	return nil
}

func isBpsDomain(hostname string) bool {
	h := strings.ToLower(hostname)
	return h == "bps.go.id" || strings.HasSuffix(h, ".bps.go.id")
}
