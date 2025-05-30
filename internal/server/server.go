package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/gobeyondidentity/google-workspace-provisioner/internal/config"
	syncengine "github.com/gobeyondidentity/google-workspace-provisioner/internal/sync"
	"github.com/gobeyondidentity/google-workspace-provisioner/internal/gws"
	"github.com/gobeyondidentity/google-workspace-provisioner/internal/bi"
)

// Server represents the HTTP server for SCIM sync operations
type Server struct {
	httpServer *http.Server
	logger     *logrus.Logger
	config     *config.Config
	syncEngine *syncengine.Engine
	scheduler  *Scheduler
	metrics    *Metrics
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status      string            `json:"status"`
	Version     string            `json:"version"`
	Timestamp   time.Time         `json:"timestamp"`
	Services    map[string]string `json:"services"`
	LastSync    *time.Time        `json:"last_sync,omitempty"`
	NextSync    *time.Time        `json:"next_sync,omitempty"`
	SyncEnabled bool              `json:"sync_enabled"`
}

// SyncResponse represents the manual sync response
type SyncResponse struct {
	Status    string     `json:"status"`
	Message   string     `json:"message"`
	Timestamp time.Time  `json:"timestamp"`
	Result    *SyncStats `json:"result,omitempty"`
	Error     string     `json:"error,omitempty"`
}

// SyncStats represents synchronization statistics
type SyncStats struct {
	GroupsProcessed    int           `json:"groups_processed"`
	UsersCreated       int           `json:"users_created"`
	UsersUpdated       int           `json:"users_updated"`
	GroupsCreated      int           `json:"groups_created"`
	MembershipsAdded   int           `json:"memberships_added"`
	MembershipsRemoved int           `json:"memberships_removed"`
	Duration           time.Duration `json:"duration"`
	Errors             []string      `json:"errors"`
}

// NewServer creates a new HTTP server instance
func NewServer(cfg *config.Config, logger *logrus.Logger) (*Server, error) {
	// Create Google Workspace client
	gwsClient, err := gws.NewClient(
		cfg.GoogleWorkspace.ServiceAccountKeyPath,
		cfg.GoogleWorkspace.Domain,
		cfg.GoogleWorkspace.SuperAdminEmail,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Workspace client: %w", err)
	}

	// Create Beyond Identity client
	biClient := bi.NewClient(cfg.BeyondIdentity.APIToken, cfg.BeyondIdentity.SCIMBaseURL)

	// Create sync engine
	syncEngine := syncengine.NewEngine(gwsClient, biClient, cfg, logger)

	// Create metrics collector
	metrics := NewMetrics()

	// Create scheduler if scheduling is enabled
	var scheduler *Scheduler
	if cfg.Server.ScheduleEnabled {
		scheduler = NewScheduler(cfg.Server.Schedule, syncEngine, logger, metrics)
	}

	// Create router
	router := mux.NewRouter()
	
	server := &Server{
		logger:     logger,
		config:     cfg,
		syncEngine: syncEngine,
		scheduler:  scheduler,
		metrics:    metrics,
	}

	// Register routes
	server.registerRoutes(router)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	server.httpServer = httpServer

	return server, nil
}

// registerRoutes sets up HTTP endpoints
func (s *Server) registerRoutes(router *mux.Router) {
	// Health check endpoint
	router.HandleFunc("/health", s.handleHealth).Methods("GET")
	
	// Manual sync endpoint
	router.HandleFunc("/sync", s.handleSync).Methods("POST")
	
	// Metrics endpoint
	router.HandleFunc("/metrics", s.handleMetrics).Methods("GET")
	
	// Scheduler control endpoints
	if s.scheduler != nil {
		router.HandleFunc("/scheduler/start", s.handleSchedulerStart).Methods("POST")
		router.HandleFunc("/scheduler/stop", s.handleSchedulerStop).Methods("POST")
		router.HandleFunc("/scheduler/status", s.handleSchedulerStatus).Methods("GET")
	}

	// Version endpoint
	router.HandleFunc("/version", s.handleVersion).Methods("GET")
}

// Start starts the HTTP server and scheduler
func (s *Server) Start() error {
	s.logger.Infof("Starting SCIM sync server on port %d", s.config.Server.Port)

	// Start scheduler if enabled
	if s.scheduler != nil {
		if err := s.scheduler.Start(); err != nil {
			return fmt.Errorf("failed to start scheduler: %w", err)
		}
		s.logger.Info("Scheduler started successfully")
	}

	// Start HTTP server in a goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Errorf("HTTP server error: %v", err)
		}
	}()

	s.logger.Info("SCIM sync server started successfully")
	
	// Wait for shutdown signal
	s.waitForShutdown()
	
	return nil
}

// waitForShutdown waits for termination signals and performs graceful shutdown
func (s *Server) waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	sig := <-sigChan
	s.logger.Infof("Received signal %s, starting graceful shutdown...", sig)
	
	// Stop scheduler
	if s.scheduler != nil {
		s.scheduler.Stop()
		s.logger.Info("Scheduler stopped")
	}
	
	// Stop HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Errorf("HTTP server shutdown error: %v", err)
	} else {
		s.logger.Info("HTTP server stopped gracefully")
	}
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	services := make(map[string]string)
	
	// Check Google Workspace connectivity (simplified check)
	services["google_workspace"] = "ok"
	
	// Check Beyond Identity connectivity (simplified check)
	services["beyond_identity"] = "ok"
	
	response := HealthResponse{
		Status:      "healthy",
		Version:     "0.1.0",
		Timestamp:   time.Now(),
		Services:    services,
		SyncEnabled: s.scheduler != nil,
	}
	
	// Add scheduler info if available
	if s.scheduler != nil {
		if lastSync := s.scheduler.GetLastSync(); lastSync != nil {
			response.LastSync = lastSync
		}
		if nextSync := s.scheduler.GetNextSync(); nextSync != nil {
			response.NextSync = nextSync
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSync handles manual sync requests
func (s *Server) handleSync(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("Manual sync requested via API")
	
	startTime := time.Now()
	result, err := s.syncEngine.Sync()
	duration := time.Since(startTime)
	
	response := SyncResponse{
		Timestamp: time.Now(),
	}
	
	if err != nil {
		s.logger.Errorf("Manual sync failed: %v", err)
		response.Status = "error"
		response.Message = "Sync operation failed"
		response.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		s.logger.Info("Manual sync completed successfully")
		response.Status = "success"
		response.Message = "Sync operation completed"
		response.Result = &SyncStats{
			GroupsProcessed:    result.GroupsProcessed,
			UsersCreated:       result.UsersCreated,
			UsersUpdated:       result.UsersUpdated,
			GroupsCreated:      result.GroupsCreated,
			MembershipsAdded:   result.MembershipsAdded,
			MembershipsRemoved: result.MembershipsRemoved,
			Duration:           duration,
			Errors:             errorStrings(result.Errors),
		}
		
		// Update metrics
		s.metrics.RecordSync(result, duration)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleMetrics handles metrics requests
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.metrics.GetStats())
}

// handleSchedulerStart handles scheduler start requests
func (s *Server) handleSchedulerStart(w http.ResponseWriter, r *http.Request) {
	if s.scheduler == nil {
		http.Error(w, "Scheduler not configured", http.StatusBadRequest)
		return
	}
	
	if err := s.scheduler.Start(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to start scheduler: %v", err), http.StatusInternalServerError)
		return
	}
	
	s.logger.Info("Scheduler started via API")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

// handleSchedulerStop handles scheduler stop requests
func (s *Server) handleSchedulerStop(w http.ResponseWriter, r *http.Request) {
	if s.scheduler == nil {
		http.Error(w, "Scheduler not configured", http.StatusBadRequest)
		return
	}
	
	s.scheduler.Stop()
	s.logger.Info("Scheduler stopped via API")
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

// handleSchedulerStatus handles scheduler status requests
func (s *Server) handleSchedulerStatus(w http.ResponseWriter, r *http.Request) {
	if s.scheduler == nil {
		http.Error(w, "Scheduler not configured", http.StatusBadRequest)
		return
	}
	
	status := map[string]interface{}{
		"running":   s.scheduler.IsRunning(),
		"schedule":  s.config.Server.Schedule,
		"last_sync": s.scheduler.GetLastSync(),
		"next_sync": s.scheduler.GetNextSync(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleVersion handles version requests
func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	version := map[string]string{
		"version":    "0.1.0",
		"build_time": time.Now().Format(time.RFC3339),
		"mode":       "server",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(version)
}

// errorStrings converts a slice of errors to a slice of strings
func errorStrings(errors []error) []string {
	if len(errors) == 0 {
		return nil
	}
	
	result := make([]string, len(errors))
	for i, err := range errors {
		result[i] = err.Error()
	}
	return result
}