package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort             string
	APISecretKey        string
	AdminTriggerKeyword string
	AdminPassword       string
	WABrowserName       string
	AllowedNumbers      []string
	IdleTimeoutMinutes  int
	ActiveIndicatorYear string

	GASUrlLayananData  string
	GASUrlIndikator    string
	GASUrlVideo        string
	GASUrlInfografik   string
	GASUrlPakar        string
	GASUrlFAQ          string
	GASCacheTTLMinutes int

	TentangBPSVideoURL string
}

func LoadConfig() *Config {
	// Load .env if present
	_ = godotenv.Load()

	port := getEnv("APP_PORT", "3000")
	apiSecret := getEnv("API_SECRET_KEY", "")
	adminTrigger := getEnv("ADMIN_TRIGGER_KEYWORD", "login admin - 1903")
	adminPass := getEnv("ADMIN_PASSWORD", "")
	browserName := getEnv("WA_BROWSER_NAME", "BPS KABUPATEN BANGKA BARAT")
	indicatorYear := getEnv("ACTIVE_INDICATOR_YEAR", "2025")

	idleTimeout, err := strconv.Atoi(getEnv("IDLE_TIMEOUT_MINUTES", "15"))
	if err != nil || idleTimeout <= 0 {
		idleTimeout = 15
	}

	cacheTTL, err := strconv.Atoi(getEnv("GAS_CACHE_TTL_MINUTES", "15"))
	if err != nil || cacheTTL <= 0 {
		cacheTTL = 15
	}

	var allowedNumbers []string
	rawAllowed := os.Getenv("ALLOWED_NUMBERS")
	if rawAllowed != "" {
		for _, num := range strings.Split(rawAllowed, ",") {
			trimmed := strings.TrimSpace(num)
			if trimmed != "" {
				allowedNumbers = append(allowedNumbers, trimmed)
			}
		}
	}

	return &Config{
		AppPort:             port,
		APISecretKey:        apiSecret,
		AdminTriggerKeyword: adminTrigger,
		AdminPassword:       adminPass,
		WABrowserName:       browserName,
		AllowedNumbers:      allowedNumbers,
		IdleTimeoutMinutes:  idleTimeout,
		ActiveIndicatorYear: indicatorYear,

		GASUrlLayananData:  getEnv("GAS_URL_LAYANAN_DATA", ""),
		GASUrlIndikator:    getEnv("GAS_URL_INDIKATOR", ""),
		GASUrlVideo:        getEnv("GAS_URL_VIDEO", ""),
		GASUrlInfografik:   getEnv("GAS_URL_INFOGRAFIK", ""),
		GASUrlPakar:        getEnv("GAS_URL_PAKAR", ""),
		GASUrlFAQ:          getEnv("GAS_URL_FAQ", ""),
		GASCacheTTLMinutes: cacheTTL,

		TentangBPSVideoURL: getEnv("TENTANG_BPS_VIDEO_URL", "https://www.youtube.com/watch?v=BFj6yi8HLAg"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
