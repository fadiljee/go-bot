package scraper

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// Config holds the configuration for the BPS scraper (§13).
type Config struct {
	Enabled                bool          `json:"enabled"`
	BaseURL                string        `json:"base_url"`
	AllowedHosts           []string      `json:"allowed_hosts"`
	UserAgent              string        `json:"user_agent"`
	DBPath                 string        `json:"db_path"`
	SelectorsPath          string        `json:"selectors_path"`
	CronSpec               string        `json:"cron_spec"`
	Timezone               string        `json:"timezone"`
	RunTimeout             time.Duration `json:"run_timeout"`
	StartupCatchupHours    int           `json:"startup_catchup_hours"`
	ManualCooldownMinutes  int           `json:"manual_cooldown_minutes"`
	EnabledKinds           []Kind        `json:"enabled_kinds"`
	MinDelayMs             int           `json:"min_delay_ms"`
	JitterMs               int           `json:"jitter_ms"`
	HTTPTimeoutSeconds     int           `json:"http_timeout_seconds"`
	MaxPageBytes           int64         `json:"max_page_bytes"`
	MaxListPages           int           `json:"max_list_pages"`
	MaxRequestsPerRun      int           `json:"max_requests_per_run"`
	RespectRobots          bool          `json:"respect_robots"`
	MaxStaleHours          int           `json:"max_stale_hours"`
	DataSourceOrder        []string      `json:"data_source_order"`
	NotifyAdmin            bool          `json:"notify_admin"`
	NotifyAfterFailures    int           `json:"notify_after_failures"`
	AppEnv                 string        `json:"app_env"`
}

// DefaultConfig returns the production-safe default config (§13).
func DefaultConfig() *Config {
	return &Config{
		Enabled:               true,
		BaseURL:               "https://bangkabaratkab.bps.go.id",
		AllowedHosts:          []string{"bangkabaratkab.bps.go.id"},
		UserAgent:             "BotIstimewa/3.1 (+https://bangkabaratkab.bps.go.id; kontak: bps1903@bps.go.id)",
		DBPath:                "./storage/scraper.db",
		SelectorsPath:         "./configs/scraper_selectors.yaml",
		CronSpec:              "0 */6 * * *",
		Timezone:              "Asia/Jakarta",
		RunTimeout:            10 * time.Minute,
		StartupCatchupHours:   6,
		ManualCooldownMinutes: 5,
		EnabledKinds: []Kind{
			KindPressRelease,
			KindPublication,
			KindInfographic,
			KindStaticTable,
		},
		MinDelayMs:          3000,
		JitterMs:            1000,
		HTTPTimeoutSeconds:  20,
		MaxPageBytes:        3145728, // 3 MB
		MaxListPages:        5,
		MaxRequestsPerRun:   60,
		RespectRobots:       true,
		MaxStaleHours:       48,
		DataSourceOrder:     []string{"bpsapi", "scraper", "gsheet"},
		NotifyAdmin:         true,
		NotifyAfterFailures: 2,
		AppEnv:              "production",
	}
}

// Validate checks configuration against required safety rules (§13).
func (c *Config) Validate() error {
	if !c.Enabled {
		return nil
	}

	// 1. Cron spec validation
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	if _, err := parser.Parse(c.CronSpec); err != nil {
		return fmt.Errorf("config: invalid SCRAPER_CRON_SPEC %q: %w", c.CronSpec, err)
	}

	// 2. Allowed hosts validation
	if len(c.AllowedHosts) == 0 {
		return errors.New("config: SCRAPER_ALLOWED_HOSTS cannot be empty")
	}
	for _, host := range c.AllowedHosts {
		h := strings.TrimSpace(strings.ToLower(host))
		if h == "" {
			return errors.New("config: SCRAPER_ALLOWED_HOSTS contains empty host")
		}
		if h == "localhost" || h == "127.0.0.1" || h == "::1" {
			if c.AppEnv != "test" {
				return fmt.Errorf("config: SCRAPER_ALLOWED_HOSTS cannot contain localhost or loopback in non-test mode: %s", h)
			}
		}
		if ip := net.ParseIP(h); ip != nil {
			if c.AppEnv != "test" {
				return fmt.Errorf("config: SCRAPER_ALLOWED_HOSTS cannot be direct IP address in non-test mode: %s", h)
			}
		}
	}

	// 3. BaseURL validation
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return fmt.Errorf("config: invalid SCRAPER_BASE_URL: %w", err)
	}
	if u.Scheme != "https" && (c.AppEnv != "test" || u.Scheme != "http") {
		return fmt.Errorf("config: SCRAPER_BASE_URL must use https scheme: %s", c.BaseURL)
	}

	// 4. Rate and delay limits
	if c.MinDelayMs < 1000 {
		return fmt.Errorf("config: SCRAPER_MIN_DELAY_MS cannot be less than 1000ms: %d", c.MinDelayMs)
	}
	if c.MaxRequestsPerRun > 200 {
		return fmt.Errorf("config: SCRAPER_MAX_REQUESTS_PER_RUN cannot exceed 200: %d", c.MaxRequestsPerRun)
	}

	// 5. Robots.txt in production
	if !c.RespectRobots && c.AppEnv == "production" {
		return errors.New("config: SCRAPER_RESPECT_ROBOTS cannot be false in production")
	}

	return nil
}
