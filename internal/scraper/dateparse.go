package scraper

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata" // ensure embedded tzdata is available
)

var (
	wibLocation *time.Location
	cleanSpaceRe = regexp.MustCompile(`[\s\x{00a0}]+`)
	urlDateRe    = regexp.MustCompile(`/(\d{4})/(\d{1,2})/(\d{1,2})/`)
)

func init() {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	wibLocation = loc
}

// GetWIBLocation returns the Asia/Jakarta (WIB) time location.
func GetWIBLocation() *time.Location {
	return wibLocation
}

var monthMap = map[string]time.Month{
	// Indonesian
	"januari":   time.January,
	"februari":  time.February,
	"maret":     time.March,
	"april":     time.April,
	"mei":       time.May,
	"juni":      time.June,
	"juli":      time.July,
	"agustus":   time.August,
	"september": time.September,
	"oktober":   time.October,
	"november":  time.November,
	"nopember":  time.November,
	"desember":  time.December,

	// Short Indonesian
	"jan": time.January,
	"feb": time.February,
	"mar": time.March,
	"apr": time.April,
	"jun": time.June,
	"jul": time.July,
	"agu": time.August,
	"agt": time.August,
	"sep": time.September,
	"okt": time.October,
	"nov": time.November,
	"des": time.December,

	// English
	"january":  time.January,
	"february": time.February,
	"march":    time.March,
	"may":      time.May,
	"june":     time.June,
	"july":     time.July,
	"august":   time.August,
	"october":  time.October,
	"december": time.December,

	// Short English
	"aug": time.August,
	"oct": time.October,
	"dec": time.December,
}

// ParseDate parses Indonesian or English date strings or falls back to date segments in a URL.
// Returns normalized time at 00:00:00 WIB.
func ParseDate(text string, fallbackURL string) (*time.Time, error) {
	cleaned := cleanSpaceRe.ReplaceAllString(strings.TrimSpace(text), " ")

	if cleaned != "" {
		if t, err := parseDateString(cleaned); err == nil {
			return t, nil
		}
	}

	if fallbackURL != "" {
		if t, err := parseDateFromURL(fallbackURL); err == nil {
			return t, nil
		}
	}

	return nil, errors.New("dateparse: unable to parse date")
}

var (
	// e.g. "1 September 2021", "01 Sep 2021", "1-September-2021", "1/09/2021"
	dayMonthYearRe = regexp.MustCompile(`(?i)^(\d{1,2})[\s\-\./]+([a-zA-Z]+|\d{1,2})[\s\-\./]+(\d{4})$`)
	// e.g. "September 1, 2021" or "Sep 1 2021"
	monthDayYearRe = regexp.MustCompile(`(?i)^([a-zA-Z]+)[\s\-\./]+(\d{1,2})(?:st|nd|rd|th)?,?[\s\-\./]+(\d{4})$`)
	// e.g. "2021-09-01" or "2021/09/01"
	isoRe = regexp.MustCompile(`^(\d{4})[\-\./](\d{1,2})[\-\./](\d{1,2})$`)
)

func parseDateString(text string) (*time.Time, error) {
	// Strip day name if present, e.g. "Rabu, 01 September 2021" or "Wednesday, Sep 1, 2021"
	if idx := strings.Index(text, ","); idx != -1 {
		// Could be "Wednesday, September 1, 2021" or "01 Sep 2021, 14:00"
		parts := strings.Split(text, ",")
		for _, part := range parts {
			partClean := strings.TrimSpace(part)
			if t, err := tryMatchDatePatterns(partClean); err == nil {
				return t, nil
			}
		}
	}

	return tryMatchDatePatterns(text)
}

func tryMatchDatePatterns(text string) (*time.Time, error) {
	text = strings.TrimSpace(text)

	// 1. ISO format: YYYY-MM-DD
	if m := isoRe.FindStringSubmatch(text); len(m) == 4 {
		y, _ := strconv.Atoi(m[1])
		mo, _ := strconv.Atoi(m[2])
		d, _ := strconv.Atoi(m[3])
		return buildDate(y, time.Month(mo), d)
	}

	// 2. Day Month Year: e.g. "1 September 2021"
	if m := dayMonthYearRe.FindStringSubmatch(text); len(m) == 4 {
		d, _ := strconv.Atoi(m[1])
		y, _ := strconv.Atoi(m[3])
		moStr := strings.ToLower(m[2])

		var month time.Month
		if moNum, err := strconv.Atoi(moStr); err == nil {
			month = time.Month(moNum)
		} else if mVal, ok := monthMap[moStr]; ok {
			month = mVal
		} else {
			return nil, fmt.Errorf("unknown month: %s", moStr)
		}
		return buildDate(y, month, d)
	}

	// 3. Month Day Year: e.g. "September 1, 2021"
	if m := monthDayYearRe.FindStringSubmatch(text); len(m) == 4 {
		moStr := strings.ToLower(m[1])
		d, _ := strconv.Atoi(m[2])
		y, _ := strconv.Atoi(m[3])

		month, ok := monthMap[moStr]
		if !ok {
			return nil, fmt.Errorf("unknown month: %s", moStr)
		}
		return buildDate(y, month, d)
	}

	return nil, errors.New("no pattern matched")
}

func parseDateFromURL(rawURL string) (*time.Time, error) {
	m := urlDateRe.FindStringSubmatch(rawURL)
	if len(m) == 4 {
		y, _ := strconv.Atoi(m[1])
		mo, _ := strconv.Atoi(m[2])
		d, _ := strconv.Atoi(m[3])
		return buildDate(y, time.Month(mo), d)
	}
	return nil, errors.New("no date found in url")
}

func buildDate(year int, month time.Month, day int) (*time.Time, error) {
	currentYear := time.Now().In(wibLocation).Year()
	if year < 1990 || year > currentYear+1 {
		return nil, fmt.Errorf("invalid year: %d", year)
	}
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("invalid month: %d", month)
	}
	if day < 1 || day > 31 {
		return nil, fmt.Errorf("invalid day: %d", day)
	}

	t := time.Date(year, month, day, 0, 0, 0, 0, wibLocation)
	// Validate day didn't overflow (e.g. Feb 31 -> March)
	if t.Day() != day || t.Month() != month || t.Year() != year {
		return nil, fmt.Errorf("invalid calendar date: %04d-%02d-%02d", year, month, day)
	}

	return &t, nil
}
