package scraper

import (
	"context"
	"time"
)

type Kind string

const (
	KindPressRelease Kind = "pressrelease"
	KindPublication  Kind = "publication"
	KindInfographic  Kind = "infographic"
	KindStaticTable  Kind = "statictable"
	KindNews         Kind = "news"
)

// Item represents the canonical model for scraped data (§5.2).
type Item struct {
	ID            int64      `json:"id,omitempty"`
	Kind          Kind       `json:"kind"`
	Lang          string     `json:"lang"`
	SourceKey     string     `json:"source_key"`
	Title         string     `json:"title"`
	URL           string     `json:"url"`
	ReleaseDate   *time.Time `json:"release_date,omitempty"`
	Abstract      string     `json:"abstract,omitempty"`
	FileURL       string     `json:"file_url,omitempty"`
	FileType      string     `json:"file_type,omitempty"`
	FileSizeText  string     `json:"file_size_text,omitempty"`
	ThumbnailURL  string     `json:"thumbnail_url,omitempty"`
	ImageURL      string     `json:"image_url,omitempty"`
	Subject       string     `json:"subject,omitempty"`
	ContentHash   string     `json:"content_hash"`
	IsActive      bool       `json:"is_active"`
	MissedRuns    int        `json:"missed_runs"`
	FirstSeenAt   time.Time  `json:"first_seen_at"`
	LastSeenAt    time.Time  `json:"last_seen_at"`
	LastChangedAt time.Time  `json:"last_changed_at"`
}

// Pagination contains list pagination info (§6.3).
type Pagination struct {
	Page     int  `json:"page"`
	HasNext  bool `json:"has_next"`
	LastPage int  `json:"last_page"` // 0 if unknown
}

// Freshness tells handlers how fresh the data is (§6.3).
type Freshness struct {
	UpdatedAt time.Time `json:"updated_at"`
	Stale     bool      `json:"stale"`  // true if older than SCRAPER_MAX_STALE_HOURS
	Source    string    `json:"source"` // "scraper" | "bpsapi" | "gsheet"
}

// IndicatorMetric represents an individual metric within a strategic indicator.
type IndicatorMetric struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value string `json:"value"`
	Note  string `json:"note,omitempty"`
}

// StrategicIndicator represents one of the 7 strategic statistical indicators (Menu 2).
type StrategicIndicator struct {
	ID          int               `json:"id"`
	Slug        string            `json:"slug"`
	Title       string            `json:"title"`
	Category    string            `json:"category"`
	Description string            `json:"description"`
	Metrics     []IndicatorMetric `json:"metrics"`
	Source      string            `json:"source"`
	SourceURL   string            `json:"source_url,omitempty"`
	ImageURL    string            `json:"image_url,omitempty"`
	UpdatedAt   string            `json:"updated_at"`
	Items       []Item            `json:"items,omitempty"`
}

// DataTopic represents an individual statistical topic within Layanan Data (Menu 1).
type DataTopic struct {
	ID          int    `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	SubjectURL  string `json:"subject_url,omitempty"`
	Tables      []Item `json:"tables,omitempty"`
	Items       []Item `json:"items,omitempty"`
}

// DataCategory represents a major category under Layanan Data (Menu 1: Sosduk, Ekonomi, Pertanian, Sektoral).
type DataCategory struct {
	ID          string      `json:"id"`
	Slug        string      `json:"slug"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Topics      []DataTopic `json:"topics,omitempty"`
}

// DataProvider is consumed by bot handlers and API (§6.3).
type DataProvider interface {
	LatestPressReleases(ctx context.Context, lang string, limit int) ([]Item, Freshness, error)
	LatestPublications(ctx context.Context, lang string, limit int) ([]Item, Freshness, error)
	LatestInfographics(ctx context.Context, limit int) ([]Item, Freshness, error)
	SearchStaticTables(ctx context.Context, keyword string, limit int) ([]Item, Freshness, error)
	SearchItems(ctx context.Context, keyword string, lang string, limit int) ([]Item, Freshness, error)
	CreateConsultationBooking(ctx context.Context, b *ConsultationBooking) error
	GetStrategicIndicator(ctx context.Context, idOrSlug string) (*StrategicIndicator, Freshness, error)
	GetAllStrategicIndicators(ctx context.Context) ([]StrategicIndicator, Freshness, error)
	GetLayananDataCategories(ctx context.Context) ([]DataCategory, Freshness, error)
	GetLayananDataCategory(ctx context.Context, categorySlug string) (*DataCategory, Freshness, error)
	GetLayananDataTopic(ctx context.Context, categorySlug, topicIdOrSlug string) (*DataTopic, Freshness, error)
	GetExpertContacts(ctx context.Context) ([]ExpertContact, error)
	GetExpertContact(ctx context.Context, bidangID int) (*ExpertContact, error)
	UpdateExpertContact(ctx context.Context, bidangID int, name, phone string) error
}

// ExpertContact represents a statistical subject-matter expert contact.
type ExpertContact struct {
	ID          int       `json:"id"`
	BidangID    int       `json:"bidang_id"`
	BidangTitle string    `json:"bidang_title"`
	PakarName   string    `json:"pakar_name"`
	Phone       string    `json:"phone"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ConsultationBooking represents a consultation schedule request from WhatsApp bot users.
type ConsultationBooking struct {
	ID            int64     `json:"id"`
	TicketNo      string    `json:"ticket_no"`
	UserJID       string    `json:"user_jid"`
	UserName      string    `json:"user_name"`
	Institution   string    `json:"institution"`
	Topic         string    `json:"topic"`
	PreferredTime string    `json:"preferred_time"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
}

// RunResult tracks the result of a scrape run in SQLite (§9.2).
type RunResult struct {
	ID           int64      `json:"id"`
	Trigger      string     `json:"trigger"` // "cron" | "manual" | "startup"
	Kind         string     `json:"kind"`    // "all" or specific kind
	StartedAt    time.Time  `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	Status       string     `json:"status"`
	Requests     int        `json:"requests"`
	ItemsFound   int        `json:"items_found"`
	ItemsNew     int        `json:"items_new"`
	ItemsUpdated int        `json:"items_updated"`
	ItemsSkipped int        `json:"items_skipped"`
	ErrorMessage string     `json:"error_message,omitempty"`
}
