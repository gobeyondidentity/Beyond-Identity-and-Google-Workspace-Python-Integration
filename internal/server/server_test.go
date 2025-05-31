package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/gobeyondidentity/google-workspace-provisioner/internal/config"
	"github.com/gobeyondidentity/google-workspace-provisioner/internal/sync"
)

// Mock sync engine for testing
type mockSyncEngine struct {
	shouldError bool
	result      *sync.SyncResult
}

func (m *mockSyncEngine) Sync() (*sync.SyncResult, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock sync error")
	}
	return m.result, nil
}

// Helper to create a test server without external dependencies
func createTestServer(t *testing.T) *Server {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
		},
		BeyondIdentity: config.BeyondIdentityConfig{
			SCIMBaseURL: "https://test.com/scim",
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Reduce log noise during tests

	server := &Server{
		config:  cfg,
		logger:  logger,
		metrics: NewMetrics(),
		syncEngine: &mockSyncEngine{
			result: &sync.SyncResult{
				GroupsProcessed:    2,
				UsersCreated:       5,
				UsersUpdated:       1,
				GroupsCreated:      1,
				MembershipsAdded:   6,
				MembershipsRemoved: 0,
				Errors:             nil,
			},
		},
	}

	return server
}

func TestHandleHealth(t *testing.T) {
	server := createTestServer(t)
	router := mux.NewRouter()
	server.registerRoutes(router)

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", status)
	}

	var response HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response.Status)
	}

	if response.Version == "" {
		t.Error("Expected version to be set")
	}

	if response.Services["google_workspace"] != "ok" {
		t.Error("Expected google_workspace service to be ok")
	}

	if response.Services["beyond_identity"] != "ok" {
		t.Error("Expected beyond_identity service to be ok")
	}

	if response.SyncEnabled {
		t.Error("Expected sync to be disabled when no scheduler is configured")
	}
}

func TestHandleSync_Success(t *testing.T) {
	server := createTestServer(t)
	router := mux.NewRouter()
	server.registerRoutes(router)

	req, err := http.NewRequest("POST", "/sync", bytes.NewBuffer([]byte{}))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", status)
	}

	var response SyncResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", response.Status)
	}

	if response.Message != "Sync operation completed" {
		t.Errorf("Expected message 'Sync operation completed', got '%s'", response.Message)
	}

	if response.Result == nil {
		t.Error("Expected result to be present")
	} else {
		if response.Result.GroupsProcessed != 2 {
			t.Errorf("Expected 2 groups processed, got %d", response.Result.GroupsProcessed)
		}
		if response.Result.UsersCreated != 5 {
			t.Errorf("Expected 5 users created, got %d", response.Result.UsersCreated)
		}
	}

	if response.Error != "" {
		t.Errorf("Expected no error, got '%s'", response.Error)
	}
}

func TestHandleSync_Error(t *testing.T) {
	server := createTestServer(t)
	// Make the sync engine return an error
	server.syncEngine = &mockSyncEngine{
		shouldError: true,
	}

	router := mux.NewRouter()
	server.registerRoutes(router)

	req, err := http.NewRequest("POST", "/sync", bytes.NewBuffer([]byte{}))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Expected status code 500, got %d", status)
	}

	var response SyncResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response.Status != "error" {
		t.Errorf("Expected status 'error', got '%s'", response.Status)
	}

	if response.Error == "" {
		t.Error("Expected error message to be present")
	}

	if response.Result != nil {
		t.Error("Expected result to be nil on error")
	}
}

func TestHandleMetrics(t *testing.T) {
	server := createTestServer(t)

	// Record some test metrics
	testResult := &sync.SyncResult{
		GroupsProcessed:    3,
		UsersCreated:       7,
		UsersUpdated:       2,
		GroupsCreated:      1,
		MembershipsAdded:   9,
		MembershipsRemoved: 1,
		Errors:             nil,
	}
	server.metrics.RecordSync(testResult, 150*time.Millisecond)

	router := mux.NewRouter()
	server.registerRoutes(router)

	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", status)
	}

	var response MetricsStats
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response.TotalSyncs != 1 {
		t.Errorf("Expected 1 total sync, got %d", response.TotalSyncs)
	}

	if response.SuccessfulSyncs != 1 {
		t.Errorf("Expected 1 successful sync, got %d", response.SuccessfulSyncs)
	}

	if response.TotalUsersCreated != 7 {
		t.Errorf("Expected 7 users created, got %d", response.TotalUsersCreated)
	}
}

func TestHandleSchedulerStart_NoScheduler(t *testing.T) {
	server := createTestServer(t)
	// Explicitly set scheduler to nil
	server.scheduler = nil

	router := mux.NewRouter()
	server.registerRoutes(router)

	req, err := http.NewRequest("POST", "/scheduler/start", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// When scheduler is nil, the route doesn't get registered, so we get 404
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Expected status code 404, got %d", status)
	}
}

func TestHandleSchedulerStop_NoScheduler(t *testing.T) {
	server := createTestServer(t)
	// Explicitly set scheduler to nil
	server.scheduler = nil

	router := mux.NewRouter()
	server.registerRoutes(router)

	req, err := http.NewRequest("POST", "/scheduler/stop", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// When scheduler is nil, the route doesn't get registered, so we get 404
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Expected status code 404, got %d", status)
	}
}

func TestHandleSchedulerStatus_NoScheduler(t *testing.T) {
	server := createTestServer(t)
	// Explicitly set scheduler to nil
	server.scheduler = nil

	router := mux.NewRouter()
	server.registerRoutes(router)

	req, err := http.NewRequest("GET", "/scheduler/status", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// When scheduler is nil, the route doesn't get registered, so we get 404
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Expected status code 404, got %d", status)
	}
}

func TestHandleVersion(t *testing.T) {
	server := createTestServer(t)
	router := mux.NewRouter()
	server.registerRoutes(router)

	req, err := http.NewRequest("GET", "/version", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", status)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response["version"] == "" {
		t.Error("Expected version to be set")
	}

	if response["mode"] != "server" {
		t.Errorf("Expected mode 'server', got '%s'", response["mode"])
	}

	if response["build_time"] == "" {
		t.Error("Expected build_time to be set")
	}
}

func TestErrorStrings(t *testing.T) {
	tests := []struct {
		name     string
		errors   []error
		expected []string
	}{
		{
			name:     "nil errors",
			errors:   nil,
			expected: nil,
		},
		{
			name:     "empty errors",
			errors:   []error{},
			expected: nil,
		},
		{
			name: "single error",
			errors: []error{
				fmt.Errorf("test error"),
			},
			expected: []string{"test error"},
		},
		{
			name: "multiple errors",
			errors: []error{
				fmt.Errorf("first error"),
				fmt.Errorf("second error"),
			},
			expected: []string{"first error", "second error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errorStrings(tt.errors)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d error strings, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Expected error string '%s', got '%s'", expected, result[i])
				}
			}
		})
	}
}

func TestSyncResponse(t *testing.T) {
	response := SyncResponse{
		Status:    "success",
		Message:   "Test message",
		Timestamp: time.Now(),
		Result: &SyncStats{
			GroupsProcessed:    5,
			UsersCreated:       10,
			UsersUpdated:       3,
			GroupsCreated:      2,
			MembershipsAdded:   15,
			MembershipsRemoved: 1,
			Duration:           250 * time.Millisecond,
			Errors:             []string{"test error"},
		},
		Error: "",
	}

	if response.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", response.Status)
	}

	if response.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", response.Message)
	}

	if response.Result == nil {
		t.Error("Expected result to be present")
	} else {
		if response.Result.GroupsProcessed != 5 {
			t.Errorf("Expected 5 groups processed, got %d", response.Result.GroupsProcessed)
		}
		if response.Result.UsersCreated != 10 {
			t.Errorf("Expected 10 users created, got %d", response.Result.UsersCreated)
		}
		if response.Result.Duration != 250*time.Millisecond {
			t.Errorf("Expected duration 250ms, got %v", response.Result.Duration)
		}
	}
}

func TestHealthResponse(t *testing.T) {
	lastSync := time.Now().Add(-1 * time.Hour)
	nextSync := time.Now().Add(1 * time.Hour)

	response := HealthResponse{
		Status:      "healthy",
		Version:     "1.0.0",
		Timestamp:   time.Now(),
		Services:    map[string]string{"test": "ok"},
		LastSync:    &lastSync,
		NextSync:    &nextSync,
		SyncEnabled: true,
	}

	if response.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response.Status)
	}

	if response.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", response.Version)
	}

	if response.Services["test"] != "ok" {
		t.Errorf("Expected service status 'ok', got '%s'", response.Services["test"])
	}

	if !response.SyncEnabled {
		t.Error("Expected sync to be enabled")
	}

	if response.LastSync == nil {
		t.Error("Expected last sync to be set")
	}

	if response.NextSync == nil {
		t.Error("Expected next sync to be set")
	}
}
