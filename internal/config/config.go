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
	DeveloperPhone      string
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
	InternalPSTGroupJID string

	// WebAPI BPS (https://webapi.bps.go.id/v1/api/list)
	BPSAPIBaseURL      string
	BPSAPIKey          string
	BPSDefaultDomain   string
	BPSDefaultLang     string
	BPSCacheTTLMinutes int

	// Scraper BPS (bangkabaratkab.bps.go.id)
	AppEnv                     string
	ScraperEnabled             bool
	ScraperBaseURL             string
	ScraperAllowedHosts        []string
	ScraperUserAgent           string
	ScraperDBPath              string
	ScraperSelectorsPath       string
	ScraperCronSpec            string
	ScraperTimezone            string
	ScraperRunTimeoutMinutes   int
	ScraperStartupCatchupHours int
	ScraperManualCooldownMinutes int
	ScraperEnabledKinds        []string
	ScraperMinDelayMs          int
	ScraperJitterMs            int
	ScraperHTTPTimeoutSeconds  int
	ScraperMaxPageBytes        int64
	ScraperMaxListPages        int
	ScraperMaxRequestsPerRun   int
	ScraperRespectRobots       bool
	ScraperMaxStaleHours       int
	DataSourceOrder            []string
	ScraperNotifyAdmin         bool
	ScraperNotifyAfterFailures int
}

func LoadConfig() *Config {
	// Load .env if present
	_ = godotenv.Load()

	appHost := getEnv("APP_HOST", "127.0.0.1")
	port := getEnv("APP_PORT", "3000")
	apiSecret := getEnv("API_SECRET_KEY", "change_this_secret_in_production_32chars")
	adminTrigger := getEnv("ADMIN_TRIGGER_KEYWORD", "login admin - 1903")
	adminPass := getEnv("ADMIN_PASSWORD", "ganti_dengan_password_admin_yang_kuat")
	devPhone := strings.TrimSpace(getEnv("DEVELOPER_PHONE", ""))
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
		DeveloperPhone:             devPhone,
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
		InternalPSTGroupJID: strings.TrimSpace(getEnv("INTERNAL_PST_GROUP_JID", "")),

		BPSAPIBaseURL:      getEnv("BPS_API_BASE_URL", "https://webapi.bps.go.id/v1/api/list"),
		BPSAPIKey:          getEnv("BPS_API_KEY", ""),
		BPSDefaultDomain:   getEnv("BPS_DEFAULT_DOMAIN", "1903"),
		BPSDefaultLang:     getEnv("BPS_DEFAULT_LANG", "ind"),
		BPSCacheTTLMinutes: getEnvInt("BPS_CACHE_TTL_MINUTES", 15),

		AppEnv:                     getEnv("APP_ENV", "production"),
		ScraperEnabled:             getEnvBool("SCRAPER_ENABLED", true),
		ScraperBaseURL:             getEnv("SCRAPER_BASE_URL", "https://bangkabaratkab.bps.go.id"),
		ScraperAllowedHosts:        getEnvSlice("SCRAPER_ALLOWED_HOSTS", []string{"bangkabaratkab.bps.go.id"}),
		ScraperUserAgent:           getEnv("SCRAPER_USER_AGENT", "BotIstimewa/3.1 (+https://bangkabaratkab.bps.go.id; kontak: bps1903@bps.go.id)"),
		ScraperDBPath:              getEnv("SCRAPER_DB_PATH", "./storage/scraper.db"),
		ScraperSelectorsPath:       getEnv("SCRAPER_SELECTORS_PATH", "./configs/scraper_selectors.yaml"),
		ScraperCronSpec:            getEnv("SCRAPER_CRON_SPEC", "0 */6 * * *"),
		ScraperTimezone:            getEnv("SCRAPER_TIMEZONE", "Asia/Jakarta"),
		ScraperRunTimeoutMinutes:   getEnvInt("SCRAPER_RUN_TIMEOUT_MINUTES", 10),
		ScraperStartupCatchupHours: getEnvInt("SCRAPER_STARTUP_CATCHUP_HOURS", 6),
		ScraperManualCooldownMinutes: getEnvInt("SCRAPER_MANUAL_COOLDOWN_MINUTES", 5),
		ScraperEnabledKinds:        getEnvSlice("SCRAPER_ENABLED_KINDS", []string{"pressrelease", "publication", "infographic", "statictable"}),
		ScraperMinDelayMs:          getEnvInt("SCRAPER_MIN_DELAY_MS", 3000),
		ScraperJitterMs:            getEnvInt("SCRAPER_JITTER_MS", 1000),
		ScraperHTTPTimeoutSeconds:  getEnvInt("SCRAPER_HTTP_TIMEOUT_SECONDS", 20),
		ScraperMaxPageBytes:        int64(getEnvInt("SCRAPER_MAX_PAGE_BYTES", 3145728)),
		ScraperMaxListPages:        getEnvInt("SCRAPER_MAX_LIST_PAGES", 5),
		ScraperMaxRequestsPerRun:   getEnvInt("SCRAPER_MAX_REQUESTS_PER_RUN", 60),
		ScraperRespectRobots:       getEnvBool("SCRAPER_RESPECT_ROBOTS", true),
		ScraperMaxStaleHours:       getEnvInt("SCRAPER_MAX_STALE_HOURS", 48),
		DataSourceOrder:            getEnvSlice("DATA_SOURCE_ORDER", []string{"bpsapi", "scraper", "gsheet"}),
		ScraperNotifyAdmin:         getEnvBool("SCRAPER_NOTIFY_ADMIN", true),
		ScraperNotifyAfterFailures: getEnvInt("SCRAPER_NOTIFY_AFTER_FAILURES", 2),
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

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		lower := strings.ToLower(strings.TrimSpace(val))
		return lower == "true" || lower == "1" || lower == "yes"
	}
	return defaultVal
}

func getEnvSlice(key string, defaultVal []string) []string {
	if val := os.Getenv(key); val != "" {
		var res []string
		for _, part := range strings.Split(val, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				res = append(res, trimmed)
			}
		}
		if len(res) > 0 {
			return res
		}
	}
	return defaultVal
}
