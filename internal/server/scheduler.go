package server

import (
	"fmt"
	"sync"
	"time"

	syncengine "github.com/gobeyondidentity/google-workspace-provisioner/internal/sync"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// Scheduler handles scheduled sync operations
type Scheduler struct {
	cron       *cron.Cron
	schedule   string
	syncEngine *syncengine.Engine
	logger     *logrus.Logger
	metrics    *Metrics
	mu         sync.RWMutex
	running    bool
	lastSync   *time.Time
	nextSync   *time.Time
}

// NewScheduler creates a new scheduler
func NewScheduler(schedule string, syncEngine *syncengine.Engine, logger *logrus.Logger, metrics *Metrics) *Scheduler {
	// Create cron with logging
	c := cron.New(cron.WithLogger(cron.VerbosePrintfLogger(logger)))

	return &Scheduler{
		cron:       c,
		schedule:   schedule,
		syncEngine: syncEngine,
		logger:     logger,
		metrics:    metrics,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler is already running")
	}

	// Add the sync job
	entryID, err := s.cron.AddFunc(s.schedule, s.runSync)
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	// Start the cron scheduler
	s.cron.Start()
	s.running = true

	// Calculate next sync time
	entries := s.cron.Entries()
	if len(entries) > 0 {
		nextTime := entries[0].Next
		s.nextSync = &nextTime
	}

	s.logger.Infof("Scheduler started with schedule '%s' (entry ID: %d)", s.schedule, entryID)
	if s.nextSync != nil {
		s.logger.Infof("Next sync scheduled for: %s", s.nextSync.Format(time.RFC3339))
	}

	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	// Stop the cron scheduler and wait for running jobs to complete
	ctx := s.cron.Stop()
	<-ctx.Done()

	s.running = false
	s.nextSync = nil

	s.logger.Info("Scheduler stopped")
}

// IsRunning returns whether the scheduler is currently running
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetLastSync returns the time of the last sync operation
func (s *Scheduler) GetLastSync() *time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastSync
}

// GetNextSync returns the time of the next scheduled sync
func (s *Scheduler) GetNextSync() *time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running {
		return nil
	}

	// Get the latest next time from cron entries
	entries := s.cron.Entries()
	if len(entries) > 0 {
		nextTime := entries[0].Next
		return &nextTime
	}

	return s.nextSync
}

// runSync executes a sync operation (called by cron)
func (s *Scheduler) runSync() {
	s.logger.Info("Starting scheduled sync operation")

	startTime := time.Now()
	result, err := s.syncEngine.Sync()
	duration := time.Since(startTime)

	// Update last sync time
	s.mu.Lock()
	s.lastSync = &startTime

	// Update next sync time
	entries := s.cron.Entries()
	if len(entries) > 0 {
		nextTime := entries[0].Next
		s.nextSync = &nextTime
	}
	s.mu.Unlock()

	if err != nil {
		s.logger.Errorf("Scheduled sync failed: %v", err)
		s.metrics.RecordFailedSync(err, duration)
	} else {
		s.logger.Infof("Scheduled sync completed successfully in %v", duration)
		s.metrics.RecordSync(result, duration)

		// Log summary
		if len(result.Errors) > 0 {
			s.logger.Warnf("Scheduled sync completed with %d errors", len(result.Errors))
		}
	}
}
