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
	apiSecret := getEnv("API_SECRET_KEY", "rahasia_internal_bps_1903")
	adminTrigger := getEnv("ADMIN_TRIGGER_KEYWORD", "login admin - 1903")
	adminPass := getEnv("ADMIN_PASSWORD", "supertim1903")
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

		GASUrlLayananData:  getEnv("GAS_URL_LAYANAN_DATA", "https://script.google.com/macros/s/AKfycbwn-d5Qfv905faN0yvDg4Y00Kp7TdYKwesRu-wrfshyOJS8E_w414LmgobZqVobUvnW/exec"),
		GASUrlIndikator:    getEnv("GAS_URL_INDIKATOR", "https://script.google.com/macros/s/AKfycbwps593uXfWT44gGB4S1kiQANcvWKHyPJj2TZClPCCbcRbtrl9BsS4x0RGrqFbkkik/exec"),
		GASUrlVideo:        getEnv("GAS_URL_VIDEO", "https://script.google.com/macros/s/AKfycbxuu7Lz_9HrC7naCC2nwdF7p4ZofAPE6ocQLQ08zsVVITTQk5OqoYRxKPhTetkMh-0O7w/exec"),
		GASUrlInfografik:   getEnv("GAS_URL_INFOGRAFIK", "https://script.google.com/macros/s/AKfycby7TtnxNaIPSNSulPeJSwR--J1oTEgSiYg9vmCbUT6dWEc0b-e21_YR1IOmlrkd9K0Xdw/exec"),
		GASUrlPakar:        getEnv("GAS_URL_PAKAR", "https://script.google.com/macros/s/AKfycbxpa_z60hhF5zuc5BPwOFNHOYEtdqWlqn7rQTGbu6JoHvvNENZpKZsB39Wk6h0WDQKm/exec"),
		GASUrlFAQ:          getEnv("GAS_URL_FAQ", "https://script.google.com/macros/s/AKfycbzi27w5oe8pd_97iDvsOu3M0j4YcNkqB_AXFxyeG24gCkob7IIHKWOYxNtJCSzzBoyPCQ/exec"),
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
