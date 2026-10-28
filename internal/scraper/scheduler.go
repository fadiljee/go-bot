package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/robfig/cron/v3"
)

type cronLogger struct{}

func (cronLogger) Info(msg string, keysAndValues ...interface{}) {
	slog.Info(msg, keysAndValues...)
}

func (cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	slog.Error(msg, append(keysAndValues, "err", err)...)
}

// Scheduler handles periodic cron execution with timezone, jitter, and startup catch-up (§10).
type Scheduler struct {
	cfg     *Config
	runner  *Runner
	store   *Store
	cron    *cron.Cron
	ctx     context.Context
	cancel  context.CancelFunc
	entryID cron.EntryID
}

// NewScheduler creates a new scraper scheduler.
func NewScheduler(cfg *Config, runner *Runner, store *Store) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		cfg:    cfg,
		runner: runner,
		store:  store,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start registers the cron schedule and starts background scheduling (§10.3).
func (s *Scheduler) Start() error {
	if !s.cfg.Enabled {
		slog.Info("scraper scheduler disabled via config (SCRAPER_ENABLED=false)")
		return nil
	}

	loc, err := time.LoadLocation(s.cfg.Timezone)
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}

	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	s.cron = cron.New(
		cron.WithLocation(loc),
		cron.WithParser(parser),
		cron.WithChain(
			cron.Recover(cronLogger{}),
			cron.SkipIfStillRunning(cronLogger{}),
		),
	)

	entryID, err := s.cron.AddFunc(s.cfg.CronSpec, func() {
		// Jitter delay: 0-300s in production (§10.1)
		jitterMax := 300
		if s.cfg.AppEnv == "test" {
			jitterMax = 1
		}
		jitter := time.Duration(rand.Intn(jitterMax)) * time.Second

		select {
		case <-time.After(jitter):
		case <-s.ctx.Done():
			return
		}

		_, _ = s.runner.RunAll(s.ctx, TriggerCron)
	})

	if err != nil {
		return fmt.Errorf("scheduler: cron spec %q invalid: %w", s.cfg.CronSpec, err)
	}
	s.entryID = entryID

	s.cron.Start()
	slog.Info("scraper scheduler aktif", "spec", s.cfg.CronSpec, "tz", loc.String(), "next_run", s.cron.Entry(entryID).Next)

	// Startup catch-up in background goroutine (§10.2 butir 5)
	go s.catchUpOnStartup()

	return nil
}

// NextRun returns the scheduled next run time.
func (s *Scheduler) NextRun() time.Time {
	if s.cron != nil && s.entryID > 0 {
		return s.cron.Entry(s.entryID).Next
	}
	return time.Time{}
}

// Stop terminates the scheduler gracefully (§10.3).
func (s *Scheduler) Stop(ctx context.Context) {
	s.cancel()
	if s.cron != nil {
		stopCtx := s.cron.Stop()
		select {
		case <-stopCtx.Done():
		case <-ctx.Done():
		}
	}
}

func (s *Scheduler) catchUpOnStartup() {
	// Delay before catch-up to prevent resource contention during startup (§10.2 butir 5)
	startupDelay := 60 * time.Second
	if s.cfg.AppEnv == "test" {
		startupDelay = 10 * time.Millisecond
	}

	select {
	case <-time.After(startupDelay):
	case <-s.ctx.Done():
		return
	}

	// Check if data is stale or DB empty
	ageHours, err := s.store.GetDataAgeHours(s.ctx)
	if err != nil || ageHours >= float64(s.cfg.StartupCatchupHours) || ageHours == 0 {
		slog.Info("scraper running startup catch-up cycle", "last_age_hours", ageHours)
		_, _ = s.runner.RunAll(s.ctx, TriggerStartup)
	}
}
