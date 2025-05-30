package server

import (
	"sync"
	"time"

	syncengine "github.com/gobeyondidentity/google-workspace-provisioner/internal/sync"
)

// Metrics collects and tracks synchronization metrics
type Metrics struct {
	mu                    sync.RWMutex
	totalSyncs            int
	successfulSyncs       int
	failedSyncs           int
	totalUsersCreated     int
	totalUsersUpdated     int
	totalGroupsCreated    int
	totalGroupsProcessed  int
	totalMembershipsAdded int
	totalMembershipsRemoved int
	lastSyncDuration      time.Duration
	averageSyncDuration   time.Duration
	lastSyncTime          *time.Time
	lastError             error
	uptime                time.Time
}

// MetricsStats represents the current metrics statistics
type MetricsStats struct {
	TotalSyncs             int           `json:"total_syncs"`
	SuccessfulSyncs        int           `json:"successful_syncs"`
	FailedSyncs            int           `json:"failed_syncs"`
	SuccessRate            float64       `json:"success_rate"`
	TotalUsersCreated      int           `json:"total_users_created"`
	TotalUsersUpdated      int           `json:"total_users_updated"`
	TotalGroupsCreated     int           `json:"total_groups_created"`
	TotalGroupsProcessed   int           `json:"total_groups_processed"`
	TotalMembershipsAdded  int           `json:"total_memberships_added"`
	TotalMembershipsRemoved int           `json:"total_memberships_removed"`
	LastSyncDuration       time.Duration `json:"last_sync_duration"`
	AverageSyncDuration    time.Duration `json:"average_sync_duration"`
	LastSyncTime           *time.Time    `json:"last_sync_time"`
	LastError              string        `json:"last_error,omitempty"`
	Uptime                 time.Duration `json:"uptime"`
}

// NewMetrics creates a new metrics collector
func NewMetrics() *Metrics {
	return &Metrics{
		uptime: time.Now(),
	}
}

// RecordSync records a successful sync operation
func (m *Metrics) RecordSync(result *syncengine.SyncResult, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.totalSyncs++
	if len(result.Errors) == 0 {
		m.successfulSyncs++
	} else {
		m.failedSyncs++
	}
	
	m.totalUsersCreated += result.UsersCreated
	m.totalUsersUpdated += result.UsersUpdated
	m.totalGroupsCreated += result.GroupsCreated
	m.totalGroupsProcessed += result.GroupsProcessed
	m.totalMembershipsAdded += result.MembershipsAdded
	m.totalMembershipsRemoved += result.MembershipsRemoved
	
	m.lastSyncDuration = duration
	
	// Calculate average duration
	if m.totalSyncs > 0 {
		totalDuration := time.Duration(int64(m.averageSyncDuration) * int64(m.totalSyncs-1))
		m.averageSyncDuration = (totalDuration + duration) / time.Duration(m.totalSyncs)
	} else {
		m.averageSyncDuration = duration
	}
	
	now := time.Now()
	m.lastSyncTime = &now
	
	// Clear last error on successful sync
	if len(result.Errors) == 0 {
		m.lastError = nil
	} else if len(result.Errors) > 0 {
		m.lastError = result.Errors[0] // Store first error
	}
}

// RecordFailedSync records a failed sync operation
func (m *Metrics) RecordFailedSync(err error, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.totalSyncs++
	m.failedSyncs++
	m.lastSyncDuration = duration
	m.lastError = err
	
	// Calculate average duration
	if m.totalSyncs > 0 {
		totalDuration := time.Duration(int64(m.averageSyncDuration) * int64(m.totalSyncs-1))
		m.averageSyncDuration = (totalDuration + duration) / time.Duration(m.totalSyncs)
	} else {
		m.averageSyncDuration = duration
	}
	
	now := time.Now()
	m.lastSyncTime = &now
}

// GetStats returns the current metrics statistics
func (m *Metrics) GetStats() *MetricsStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var successRate float64
	if m.totalSyncs > 0 {
		successRate = float64(m.successfulSyncs) / float64(m.totalSyncs) * 100
	}
	
	var lastErrorStr string
	if m.lastError != nil {
		lastErrorStr = m.lastError.Error()
	}
	
	return &MetricsStats{
		TotalSyncs:              m.totalSyncs,
		SuccessfulSyncs:         m.successfulSyncs,
		FailedSyncs:             m.failedSyncs,
		SuccessRate:             successRate,
		TotalUsersCreated:       m.totalUsersCreated,
		TotalUsersUpdated:       m.totalUsersUpdated,
		TotalGroupsCreated:      m.totalGroupsCreated,
		TotalGroupsProcessed:    m.totalGroupsProcessed,
		TotalMembershipsAdded:   m.totalMembershipsAdded,
		TotalMembershipsRemoved: m.totalMembershipsRemoved,
		LastSyncDuration:        m.lastSyncDuration,
		AverageSyncDuration:     m.averageSyncDuration,
		LastSyncTime:            m.lastSyncTime,
		LastError:               lastErrorStr,
		Uptime:                  time.Since(m.uptime),
	}
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.totalSyncs = 0
	m.successfulSyncs = 0
	m.failedSyncs = 0
	m.totalUsersCreated = 0
	m.totalUsersUpdated = 0
	m.totalGroupsCreated = 0
	m.totalGroupsProcessed = 0
	m.totalMembershipsAdded = 0
	m.totalMembershipsRemoved = 0
	m.lastSyncDuration = 0
	m.averageSyncDuration = 0
	m.lastSyncTime = nil
	m.lastError = nil
	m.uptime = time.Now()
}