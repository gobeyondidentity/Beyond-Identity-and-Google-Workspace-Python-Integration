package server

import (
	"fmt"
	"testing"
	"time"

	"github.com/gobeyondidentity/google-workspace-provisioner/internal/sync"
)

func TestNewMetrics(t *testing.T) {
	metrics := NewMetrics()

	if metrics == nil {
		t.Error("Expected metrics to be created, got nil")
		return
	}

	if metrics.uptime.IsZero() {
		t.Error("Expected uptime to be set")
	}

	// Check that uptime is recent (within last second)
	if time.Since(metrics.uptime) > time.Second {
		t.Error("Expected uptime to be recent")
	}
}

func TestRecordSync(t *testing.T) {
	metrics := NewMetrics()

	// Record a successful sync
	result := &sync.SyncResult{
		GroupsProcessed:    2,
		UsersCreated:       5,
		UsersUpdated:       3,
		GroupsCreated:      1,
		MembershipsAdded:   8,
		MembershipsRemoved: 2,
		Errors:             nil,
	}

	metrics.RecordSync(result, 100*time.Millisecond)

	if metrics.totalSyncs != 1 {
		t.Errorf("Expected totalSyncs to be 1, got %d", metrics.totalSyncs)
	}

	if metrics.successfulSyncs != 1 {
		t.Errorf("Expected successfulSyncs to be 1, got %d", metrics.successfulSyncs)
	}

	if metrics.failedSyncs != 0 {
		t.Errorf("Expected failedSyncs to be 0, got %d", metrics.failedSyncs)
	}

	if metrics.totalUsersCreated != 5 {
		t.Errorf("Expected totalUsersCreated to be 5, got %d", metrics.totalUsersCreated)
	}

	if metrics.totalUsersUpdated != 3 {
		t.Errorf("Expected totalUsersUpdated to be 3, got %d", metrics.totalUsersUpdated)
	}

	if metrics.totalGroupsCreated != 1 {
		t.Errorf("Expected totalGroupsCreated to be 1, got %d", metrics.totalGroupsCreated)
	}

	if metrics.totalGroupsProcessed != 2 {
		t.Errorf("Expected totalGroupsProcessed to be 2, got %d", metrics.totalGroupsProcessed)
	}

	if metrics.totalMembershipsAdded != 8 {
		t.Errorf("Expected totalMembershipsAdded to be 8, got %d", metrics.totalMembershipsAdded)
	}

	if metrics.totalMembershipsRemoved != 2 {
		t.Errorf("Expected totalMembershipsRemoved to be 2, got %d", metrics.totalMembershipsRemoved)
	}

	if metrics.lastSyncDuration != 100*time.Millisecond {
		t.Errorf("Expected lastSyncDuration to be 100ms, got %v", metrics.lastSyncDuration)
	}

	if metrics.lastSyncTime == nil {
		t.Error("Expected lastSyncTime to be set")
	}
}

func TestRecordSyncWithErrors(t *testing.T) {
	metrics := NewMetrics()

	// Record a failed sync
	result := &sync.SyncResult{
		GroupsProcessed:    1,
		UsersCreated:       0,
		UsersUpdated:       0,
		GroupsCreated:      0,
		MembershipsAdded:   0,
		MembershipsRemoved: 0,
		Errors:             []error{fmt.Errorf("test error")},
	}

	metrics.RecordSync(result, 50*time.Millisecond)

	if metrics.totalSyncs != 1 {
		t.Errorf("Expected totalSyncs to be 1, got %d", metrics.totalSyncs)
	}

	if metrics.successfulSyncs != 0 {
		t.Errorf("Expected successfulSyncs to be 0, got %d", metrics.successfulSyncs)
	}

	if metrics.failedSyncs != 1 {
		t.Errorf("Expected failedSyncs to be 1, got %d", metrics.failedSyncs)
	}
}

func TestCalculateSuccessRate(t *testing.T) {
	metrics := NewMetrics()

	// No syncs yet
	stats := metrics.GetStats()
	if stats.SuccessRate != 0.0 {
		t.Errorf("Expected success rate 0.0 for no syncs, got %f", stats.SuccessRate)
	}

	// Record successful syncs
	for i := 0; i < 8; i++ {
		result := &sync.SyncResult{
			Errors: nil,
		}
		metrics.RecordSync(result, 100*time.Millisecond)
	}

	// Record failed syncs
	for i := 0; i < 2; i++ {
		result := &sync.SyncResult{
			Errors: []error{fmt.Errorf("test error")},
		}
		metrics.RecordSync(result, 100*time.Millisecond)
	}

	stats = metrics.GetStats()
	expected := 80.0 // 8 successful out of 10 total
	if stats.SuccessRate != expected {
		t.Errorf("Expected success rate %f, got %f", expected, stats.SuccessRate)
	}
}

func TestCalculateAverageDuration(t *testing.T) {
	metrics := NewMetrics()

	// No syncs yet
	stats := metrics.GetStats()
	if stats.AverageSyncDuration != 0 {
		t.Errorf("Expected average duration 0 for no syncs, got %v", stats.AverageSyncDuration)
	}

	// Record syncs with different durations
	durations := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		300 * time.Millisecond,
	}

	for _, duration := range durations {
		result := &sync.SyncResult{
			Errors: nil,
		}
		metrics.RecordSync(result, duration)
	}

	stats = metrics.GetStats()
	expected := 200 * time.Millisecond // (100 + 200 + 300) / 3
	if stats.AverageSyncDuration != expected {
		t.Errorf("Expected average duration %v, got %v", expected, stats.AverageSyncDuration)
	}
}

func TestCalculateUptime(t *testing.T) {
	metrics := NewMetrics()

	// Wait a small amount of time
	time.Sleep(10 * time.Millisecond)

	stats := metrics.GetStats()

	if stats.Uptime < 10*time.Millisecond {
		t.Errorf("Expected uptime to be at least 10ms, got %v", stats.Uptime)
	}

	if stats.Uptime > 1*time.Second {
		t.Errorf("Expected uptime to be less than 1s, got %v", stats.Uptime)
	}
}

func TestSyncResult(t *testing.T) {
	result := &sync.SyncResult{
		GroupsProcessed:    5,
		UsersCreated:       10,
		UsersUpdated:       3,
		GroupsCreated:      2,
		MembershipsAdded:   15,
		MembershipsRemoved: 1,
		Errors:             []error{fmt.Errorf("test error")},
	}

	if result.GroupsProcessed != 5 {
		t.Errorf("Expected GroupsProcessed 5, got %d", result.GroupsProcessed)
	}

	if result.UsersCreated != 10 {
		t.Errorf("Expected UsersCreated 10, got %d", result.UsersCreated)
	}

	if result.UsersUpdated != 3 {
		t.Errorf("Expected UsersUpdated 3, got %d", result.UsersUpdated)
	}

	if result.GroupsCreated != 2 {
		t.Errorf("Expected GroupsCreated 2, got %d", result.GroupsCreated)
	}

	if result.MembershipsAdded != 15 {
		t.Errorf("Expected MembershipsAdded 15, got %d", result.MembershipsAdded)
	}

	if result.MembershipsRemoved != 1 {
		t.Errorf("Expected MembershipsRemoved 1, got %d", result.MembershipsRemoved)
	}

	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}
}
