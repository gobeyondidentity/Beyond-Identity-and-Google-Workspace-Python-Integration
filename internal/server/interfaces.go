package server

import "github.com/gobeyondidentity/google-workspace-provisioner/internal/sync"

// SyncEngine interface for sync operations
type SyncEngine interface {
	Sync() (*sync.SyncResult, error)
}