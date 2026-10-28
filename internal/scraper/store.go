package scraper

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// Store provides thread-safe access to the SQLite scraper database (§9).
type Store struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewStore initializes the SQLite database with WAL mode and schema migrations (§9.1, §9.2).
func NewStore(dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
		return nil, fmt.Errorf("store: failed to create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("store: failed to open sqlite db: %w", err)
	}

	// Restrict db file permissions (§9.1)
	_ = os.Chmod(dbPath, 0600)

	s := &Store{db: db}
	if err := s.initPragmas(); err != nil {
		db.Close()
		return nil, err
	}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}

	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) initPragmas() error {
	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA busy_timeout=5000;",
		"PRAGMA foreign_keys=ON;",
		"PRAGMA synchronous=NORMAL;",
	}
	for _, p := range pragmas {
		if _, err := s.db.Exec(p); err != nil {
			return fmt.Errorf("store: pragma %s error: %w", p, err)
		}
	}
	return nil
}

func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS scraped_items (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		kind            TEXT    NOT NULL CHECK (kind IN ('pressrelease','publication','infographic','statictable','news')),
		lang            TEXT    NOT NULL DEFAULT 'id' CHECK (lang IN ('id','en')),
		source_key      TEXT    NOT NULL,
		title           TEXT    NOT NULL,
		url             TEXT    NOT NULL,
		release_date    TEXT,
		abstract        TEXT,
		file_url        TEXT,
		file_type       TEXT,
		file_size_text  TEXT,
		thumbnail_url   TEXT,
		image_url       TEXT,
		subject         TEXT,
		content_hash    TEXT    NOT NULL,
		is_active       INTEGER NOT NULL DEFAULT 1,
		missed_runs     INTEGER NOT NULL DEFAULT 0,
		first_seen_at   TEXT    NOT NULL,
		last_seen_at    TEXT    NOT NULL,
		last_changed_at TEXT    NOT NULL,
		UNIQUE (kind, lang, source_key)
	);
	CREATE INDEX IF NOT EXISTS idx_items_kind_date ON scraped_items(kind, lang, is_active, release_date DESC);
	CREATE INDEX IF NOT EXISTS idx_items_title     ON scraped_items(kind, title COLLATE NOCASE);

	CREATE TABLE IF NOT EXISTS scrape_runs (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		trigger       TEXT NOT NULL CHECK (trigger IN ('cron','manual','startup')),
		kind          TEXT NOT NULL,
		started_at    TEXT NOT NULL,
		finished_at   TEXT,
		status        TEXT NOT NULL,
		requests      INTEGER NOT NULL DEFAULT 0,
		items_found   INTEGER NOT NULL DEFAULT 0,
		items_new     INTEGER NOT NULL DEFAULT 0,
		items_updated INTEGER NOT NULL DEFAULT 0,
		items_skipped INTEGER NOT NULL DEFAULT 0,
		error_message TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_runs_started ON scrape_runs(started_at DESC);

	CREATE TABLE IF NOT EXISTS page_cache (
		url           TEXT PRIMARY KEY,
		etag          TEXT,
		last_modified TEXT,
		body_sha256   TEXT,
		fetched_at    TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS breaker_state (
		id               INTEGER PRIMARY KEY CHECK (id = 1),
		consecutive_fail INTEGER NOT NULL DEFAULT 0,
		open_until       TEXT,
		last_reason      TEXT,
		updated_at       TEXT NOT NULL
	);
	INSERT OR IGNORE INTO breaker_state (id, consecutive_fail, updated_at) VALUES (1, 0, datetime('now'));
	`
	_, err := s.db.Exec(schema)
	return err
}

// ApplyItems executes a transaction to upsert items and update missed_runs (§9.3).
func (s *Store) ApplyItems(ctx context.Context, kind Kind, lang string, items []Item, isFullSuccess bool) (newCount, updatedCount int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	nowStr := time.Now().UTC().Format(time.RFC3339)
	seenKeys := make(map[string]bool)

	for _, it := range items {
		seenKeys[it.SourceKey] = true
		dateStr := ""
		if it.ReleaseDate != nil {
			dateStr = it.ReleaseDate.Format("2006-01-02")
		}

		var existingID int64
		var existingHash string
		row := tx.QueryRowContext(ctx,
			"SELECT id, content_hash FROM scraped_items WHERE kind = ? AND lang = ? AND source_key = ?",
			string(it.Kind), it.Lang, it.SourceKey,
		)
		err := row.Scan(&existingID, &existingHash)

		if err == sql.ErrNoRows {
			// INSERT
			_, err = tx.ExecContext(ctx, `
				INSERT INTO scraped_items (
					kind, lang, source_key, title, url, release_date, abstract,
					file_url, file_type, file_size_text, thumbnail_url, image_url,
					subject, content_hash, is_active, missed_runs,
					first_seen_at, last_seen_at, last_changed_at
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, 0, ?, ?, ?)
			`, string(it.Kind), it.Lang, it.SourceKey, it.Title, it.URL, dateStr, it.Abstract,
				it.FileURL, it.FileType, it.FileSizeText, it.ThumbnailURL, it.ImageURL,
				it.Subject, it.ContentHash, nowStr, nowStr, nowStr)
			if err != nil {
				return 0, 0, fmt.Errorf("store insert failed: %w", err)
			}
			newCount++
		} else if err == nil {
			// Exists
			if existingHash == it.ContentHash {
				// No change, just refresh last_seen_at
				_, err = tx.ExecContext(ctx, `
					UPDATE scraped_items SET last_seen_at = ?, missed_runs = 0, is_active = 1
					WHERE id = ?
				`, nowStr, existingID)
			} else {
				// Changed
				_, err = tx.ExecContext(ctx, `
					UPDATE scraped_items SET
						title = ?, url = ?, release_date = ?, abstract = ?,
						file_url = ?, file_type = ?, file_size_text = ?,
						thumbnail_url = ?, image_url = ?, subject = ?,
						content_hash = ?, is_active = 1, missed_runs = 0,
						last_seen_at = ?, last_changed_at = ?
					WHERE id = ?
				`, it.Title, it.URL, dateStr, it.Abstract,
					it.FileURL, it.FileType, it.FileSizeText,
					it.ThumbnailURL, it.ImageURL, it.Subject,
					it.ContentHash, nowStr, nowStr, existingID)
				updatedCount++
			}
			if err != nil {
				return 0, 0, fmt.Errorf("store update failed: %w", err)
			}
		} else {
			return 0, 0, err
		}
	}

	// Update missed_runs only on full success (§9.3, K-07)
	if isFullSuccess && len(seenKeys) > 0 {
		// Increment missed_runs for active items of this kind/lang that weren't seen
		rows, err := tx.QueryContext(ctx,
			"SELECT id, source_key, missed_runs FROM scraped_items WHERE kind = ? AND lang = ? AND is_active = 1",
			string(kind), lang,
		)
		if err == nil {
			type unobserved struct {
				id     int64
				missed int
			}
			var unobservedList []unobserved
			for rows.Next() {
				var id int64
				var key string
				var missed int
				if err := rows.Scan(&id, &key, &missed); err == nil {
					if !seenKeys[key] {
						unobservedList = append(unobservedList, unobserved{id: id, missed: missed + 1})
					}
				}
			}
			rows.Close()

			for _, u := range unobservedList {
				isActive := 1
				if u.missed >= 3 {
					isActive = 0 // deactivate after 3 consecutive missed runs (§9.3)
				}
				_, _ = tx.ExecContext(ctx,
					"UPDATE scraped_items SET missed_runs = ?, is_active = ? WHERE id = ?",
					u.missed, isActive, u.id,
				)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}
	return newCount, updatedCount, nil
}

// StartRun records the start of a scrape run in SQLite.
func (s *Store) StartRun(ctx context.Context, trigger, kind string) (*RunResult, error) {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx,
		"INSERT INTO scrape_runs (trigger, kind, started_at, status) VALUES (?, ?, ?, 'running')",
		trigger, kind, nowStr,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &RunResult{
		ID:        id,
		Trigger:   trigger,
		Kind:      kind,
		StartedAt: time.Now(),
		Status:    "running",
	}, nil
}

// FinishRun marks a scrape run as completed.
func (s *Store) FinishRun(ctx context.Context, r *RunResult) error {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `
		UPDATE scrape_runs SET
			finished_at = ?, status = ?, requests = ?, items_found = ?,
			items_new = ?, items_updated = ?, items_skipped = ?, error_message = ?
		WHERE id = ?
	`, nowStr, r.Status, r.Requests, r.ItemsFound,
		r.ItemsNew, r.ItemsUpdated, r.ItemsSkipped, r.ErrorMessage, r.ID)
	return err
}

// LogSkipped logs an overlap or breaker skip.
func (s *Store) LogSkipped(ctx context.Context, trigger, status string) {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	_, _ = s.db.ExecContext(ctx, `
		INSERT INTO scrape_runs (trigger, kind, started_at, finished_at, status)
		VALUES (?, 'all', ?, ?, ?)
	`, trigger, nowStr, nowStr, status)
}

// GetLatestItems retrieves active items for a given kind and language (§6.3, §11.2).
func (s *Store) GetLatestItems(ctx context.Context, kind Kind, lang string, limit int) ([]Item, error) {
	if limit <= 0 {
		limit = 10
	}
	if lang == "" {
		lang = "id"
	}

	query := `
		SELECT id, kind, lang, source_key, title, url, release_date, abstract,
		       file_url, file_type, file_size_text, thumbnail_url, image_url,
		       subject, content_hash, is_active, missed_runs,
		       first_seen_at, last_seen_at, last_changed_at
		FROM scraped_items
		WHERE kind = ? AND lang = ? AND is_active = 1
		ORDER BY release_date DESC, id DESC
		LIMIT ?
	`
	rows, err := s.db.QueryContext(ctx, query, string(kind), lang, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanItems(rows)
}

// SearchItems searches active items using parameterized LIKE query (§11.2, §11.3).
func (s *Store) SearchItems(ctx context.Context, kind Kind, keyword string, limit int) ([]Item, error) {
	if limit <= 0 {
		limit = 10
	}
	cleaned := CleanText(keyword)
	if len([]rune(cleaned)) > 100 {
		cleaned = string([]rune(cleaned)[:100])
	}
	param := "%" + cleaned + "%"

	query := `
		SELECT id, kind, lang, source_key, title, url, release_date, abstract,
		       file_url, file_type, file_size_text, thumbnail_url, image_url,
		       subject, content_hash, is_active, missed_runs,
		       first_seen_at, last_seen_at, last_changed_at
		FROM scraped_items
		WHERE kind = ? AND is_active = 1 AND (title LIKE ? OR subject LIKE ?)
		ORDER BY release_date DESC, id DESC
		LIMIT ?
	`
	rows, err := s.db.QueryContext(ctx, query, string(kind), param, param, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanItems(rows)
}

// SearchAllItems searches active items across all kinds using parameterized LIKE query.
func (s *Store) SearchAllItems(ctx context.Context, keyword string, limit int) ([]Item, error) {
	if limit <= 0 {
		limit = 10
	}
	cleaned := CleanText(keyword)
	if len([]rune(cleaned)) > 100 {
		cleaned = string([]rune(cleaned)[:100])
	}
	param := "%" + cleaned + "%"

	query := `
		SELECT id, kind, lang, source_key, title, url, release_date, abstract,
		       file_url, file_type, file_size_text, thumbnail_url, image_url,
		       subject, content_hash, is_active, missed_runs,
		       first_seen_at, last_seen_at, last_changed_at
		FROM scraped_items
		WHERE is_active = 1 AND (title LIKE ? OR subject LIKE ? OR abstract LIKE ?)
		ORDER BY release_date DESC, id DESC
		LIMIT ?
	`
	rows, err := s.db.QueryContext(ctx, query, param, param, param, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanItems(rows)
}

func scanItems(rows *sql.Rows) ([]Item, error) {
	var items []Item
	for rows.Next() {
		var it Item
		var kindStr, relDate, fSeen, lSeen, lChanged string
		var isActive int
		err := rows.Scan(
			&it.ID, &kindStr, &it.Lang, &it.SourceKey, &it.Title, &it.URL,
			&relDate, &it.Abstract, &it.FileURL, &it.FileType, &it.FileSizeText,
			&it.ThumbnailURL, &it.ImageURL, &it.Subject, &it.ContentHash,
			&isActive, &it.MissedRuns, &fSeen, &lSeen, &lChanged,
		)
		if err != nil {
			return nil, err
		}
		it.Kind = Kind(kindStr)
		it.IsActive = (isActive == 1)

		if relDate != "" {
			if t, err := time.Parse("2006-01-02", relDate); err == nil {
				wibDate := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, GetWIBLocation())
				it.ReleaseDate = &wibDate
			}
		}
		if t, err := time.Parse(time.RFC3339, fSeen); err == nil {
			it.FirstSeenAt = t
		}
		if t, err := time.Parse(time.RFC3339, lSeen); err == nil {
			it.LastSeenAt = t
		}
		if t, err := time.Parse(time.RFC3339, lChanged); err == nil {
			it.LastChangedAt = t
		}

		items = append(items, it)
	}
	return items, nil
}

// GetPageCache retrieves conditional GET data (§7.4, §9.1).
func (s *Store) GetPageCache(ctx context.Context, pageURL string) (etag, lastModified string, err error) {
	row := s.db.QueryRowContext(ctx, "SELECT etag, last_modified FROM page_cache WHERE url = ?", pageURL)
	var e, lm sql.NullString
	if err := row.Scan(&e, &lm); err != nil {
		return "", "", err
	}
	return e.String, lm.String, nil
}

// SetPageCache updates conditional GET headers for a URL.
func (s *Store) SetPageCache(ctx context.Context, pageURL, etag, lastModified, bodySha256 string) error {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO page_cache (url, etag, last_modified, body_sha256, fetched_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(url) DO UPDATE SET
			etag = excluded.etag,
			last_modified = excluded.last_modified,
			body_sha256 = excluded.body_sha256,
			fetched_at = excluded.fetched_at
	`, pageURL, etag, lastModified, bodySha256, nowStr)
	return err
}

// Implementation of BreakerStore interface (§7.9)

func (s *Store) GetBreakerState(ctx context.Context) (int, *time.Time, string, error) {
	row := s.db.QueryRowContext(ctx, "SELECT consecutive_fail, open_until, last_reason FROM breaker_state WHERE id = 1")
	var fails int
	var untilStr, reason sql.NullString
	if err := row.Scan(&fails, &untilStr, &reason); err != nil {
		return 0, nil, "", err
	}
	var until *time.Time
	if untilStr.Valid && untilStr.String != "" {
		if t, err := time.Parse(time.RFC3339, untilStr.String); err == nil {
			until = &t
		}
	}
	return fails, until, reason.String, nil
}

func (s *Store) SaveBreakerState(ctx context.Context, consecutiveFail int, openUntil *time.Time, lastReason string) error {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	var untilStr sql.NullString
	if openUntil != nil {
		untilStr = sql.NullString{String: openUntil.UTC().Format(time.RFC3339), Valid: true}
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE breaker_state SET
			consecutive_fail = ?, open_until = ?, last_reason = ?, updated_at = ?
		WHERE id = 1
	`, consecutiveFail, untilStr, lastReason, nowStr)
	return err
}

// Diagnostic helper methods for API (§12)

func (s *Store) GetLastRun(ctx context.Context) (*RunResult, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, trigger, kind, started_at, finished_at, status,
		       requests, items_found, items_new, items_updated, items_skipped, error_message
		FROM scrape_runs
		ORDER BY id DESC LIMIT 1
	`)
	var r RunResult
	var sAt, fAt, errMsg sql.NullString
	err := row.Scan(
		&r.ID, &r.Trigger, &r.Kind, &sAt, &fAt, &r.Status,
		&r.Requests, &r.ItemsFound, &r.ItemsNew, &r.ItemsUpdated, &r.ItemsSkipped, &errMsg,
	)
	if err != nil {
		return nil, err
	}
	if sAt.Valid {
		r.StartedAt, _ = time.Parse(time.RFC3339, sAt.String)
	}
	if fAt.Valid {
		t, _ := time.Parse(time.RFC3339, fAt.String)
		r.FinishedAt = &t
	}
	r.ErrorMessage = errMsg.String
	return &r, nil
}

func (s *Store) GetRuns(ctx context.Context, limit int) ([]RunResult, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, trigger, kind, started_at, finished_at, status,
		       requests, items_found, items_new, items_updated, items_skipped, error_message
		FROM scrape_runs
		ORDER BY id DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []RunResult
	for rows.Next() {
		var r RunResult
		var sAt, fAt, errMsg sql.NullString
		if err := rows.Scan(
			&r.ID, &r.Trigger, &r.Kind, &sAt, &fAt, &r.Status,
			&r.Requests, &r.ItemsFound, &r.ItemsNew, &r.ItemsUpdated, &r.ItemsSkipped, &errMsg,
		); err == nil {
			if sAt.Valid {
				r.StartedAt, _ = time.Parse(time.RFC3339, sAt.String)
			}
			if fAt.Valid {
				t, _ := time.Parse(time.RFC3339, fAt.String)
				r.FinishedAt = &t
			}
			r.ErrorMessage = errMsg.String
			runs = append(runs, r)
		}
	}
	return runs, nil
}

func (s *Store) GetCountsByKind(ctx context.Context) (map[string]int, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT kind, COUNT(*) FROM scraped_items WHERE is_active = 1 GROUP BY kind
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var k string
		var count int
		if err := rows.Scan(&k, &count); err == nil {
			counts[k] = count
		}
	}
	return counts, nil
}

func (s *Store) GetDataAgeHours(ctx context.Context) (float64, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT finished_at FROM scrape_runs WHERE status = 'success' ORDER BY id DESC LIMIT 1
	`)
	var fAt sql.NullString
	if err := row.Scan(&fAt); err != nil {
		return 0, err
	}
	if !fAt.Valid || fAt.String == "" {
		return 0, nil
	}
	t, err := time.Parse(time.RFC3339, fAt.String)
	if err != nil {
		return 0, err
	}
	return time.Since(t).Hours(), nil
}
