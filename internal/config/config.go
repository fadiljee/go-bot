package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppHost             string
	AppPort             string
	APISecretKey        string
	AdminTriggerKeyword string
	AdminPassword       string
	AdminWhitelistNumbers []string
	AdminMaxRetry       int
	AdminLockoutMinutes int

	UserRateLimitPer10s int
	UserCooldownSeconds int

	QueueInteractiveMinDelayMs int
	QueueInteractiveMaxDelayMs int
	QueueBlastMinDelayMs       int
	QueueBlastMaxDelayMs       int

	MaxMediaSizeMB           int
	HTTPClientTimeoutSeconds int

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

	appHost := getEnv("APP_HOST", "127.0.0.1")
	port := getEnv("APP_PORT", "3000")
	apiSecret := getEnv("API_SECRET_KEY", "rahasia_internal_bps_1903")
	adminTrigger := getEnv("ADMIN_TRIGGER_KEYWORD", "login admin - 1903")
	adminPass := getEnv("ADMIN_PASSWORD", "supertim1903")
	browserName := getEnv("WA_BROWSER_NAME", "BPS KABUPATEN BANGKA BARAT")
	indicatorYear := getEnv("ACTIVE_INDICATOR_YEAR", "2025")

	adminMaxRetry := getEnvInt("ADMIN_MAX_RETRY", 3)
	adminLockoutMin := getEnvInt("ADMIN_LOCKOUT_MINUTES", 60)
	userRateLimit := getEnvInt("USER_RATE_LIMIT_PER_10S", 5)
	userCooldownSec := getEnvInt("USER_COOLDOWN_SECONDS", 60)

	queueInterMin := getEnvInt("QUEUE_INTERACTIVE_MIN_DELAY_MS", 600)
	queueInterMax := getEnvInt("QUEUE_INTERACTIVE_MAX_DELAY_MS", 1200)
	queueBlastMin := getEnvInt("QUEUE_BLAST_MIN_DELAY_MS", 5000)
	queueBlastMax := getEnvInt("QUEUE_BLAST_MAX_DELAY_MS", 10000)

	maxMediaMB := getEnvInt("MAX_MEDIA_SIZE_MB", 16)
	httpTimeoutSec := getEnvInt("HTTP_CLIENT_TIMEOUT_SECONDS", 10)

	idleTimeout := getEnvInt("IDLE_TIMEOUT_MINUTES", 15)
	cacheTTL := getEnvInt("GAS_CACHE_TTL_MINUTES", 15)

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

	var adminWhitelist []string
	rawAdminWhitelist := os.Getenv("ADMIN_WHITELIST_NUMBERS")
	if rawAdminWhitelist != "" {
		for _, num := range strings.Split(rawAdminWhitelist, ",") {
			trimmed := strings.TrimSpace(num)
			if trimmed != "" {
				adminWhitelist = append(adminWhitelist, trimmed)
			}
		}
	}

	return &Config{
		AppHost:                    appHost,
		AppPort:                    port,
		APISecretKey:               apiSecret,
		AdminTriggerKeyword:        adminTrigger,
		AdminPassword:              adminPass,
		AdminWhitelistNumbers:      adminWhitelist,
		AdminMaxRetry:              adminMaxRetry,
		AdminLockoutMinutes:        adminLockoutMin,
		UserRateLimitPer10s:        userRateLimit,
		UserCooldownSeconds:        userCooldownSec,
		QueueInteractiveMinDelayMs: queueInterMin,
		QueueInteractiveMaxDelayMs: queueInterMax,
		QueueBlastMinDelayMs:       queueBlastMin,
		QueueBlastMaxDelayMs:       queueBlastMax,
		MaxMediaSizeMB:             maxMediaMB,
		HTTPClientTimeoutSeconds:   httpTimeoutSec,
		WABrowserName:              browserName,
		AllowedNumbers:             allowedNumbers,
		IdleTimeoutMinutes:         idleTimeout,
		ActiveIndicatorYear:        indicatorYear,

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

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil && i > 0 {
			return i
		}
	}
	return defaultVal
}
