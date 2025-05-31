package setup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gobeyondidentity/google-workspace-provisioner/internal/config"
)

func TestNewValidator(t *testing.T) {
	cfg := &config.Config{}
	validator := NewValidator(cfg)

	if validator == nil {
		t.Error("Expected validator to be created, got nil")
		return
	}

	if validator.config != cfg {
		t.Error("Expected validator config to match input config")
	}

	if validator.logger == nil {
		t.Error("Expected logger to be initialized")
	}
}

func TestValidateConfiguration(t *testing.T) {
	tests := []struct {
		name         string
		config       *config.Config
		expectStatus string
		expectError  bool
	}{
		{
			name: "valid configuration",
			config: &config.Config{
				App: config.AppConfig{
					LogLevel: "info",
				},
				GoogleWorkspace: config.GoogleWorkspaceConfig{
					Domain:                "test.com",
					SuperAdminEmail:       "admin@test.com",
					ServiceAccountKeyPath: "/tmp/test.json",
				},
				BeyondIdentity: config.BeyondIdentityConfig{
					APIToken: "test-token",
				},
				Sync: config.SyncConfig{
					Groups: []string{"group@test.com"},
				},
				Server: config.ServerConfig{
					Port: 8080,
				},
			},
			expectStatus: "PASS",
			expectError:  false,
		},
		{
			name: "invalid configuration",
			config: &config.Config{
				App: config.AppConfig{
					LogLevel: "invalid",
				},
			},
			expectStatus: "FAIL",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp service account file for valid config
			if !tt.expectError && tt.config.GoogleWorkspace.ServiceAccountKeyPath != "" {
				tmpDir := t.TempDir()
				serviceAccountPath := filepath.Join(tmpDir, "test.json")
				err := os.WriteFile(serviceAccountPath, []byte(`{"type": "service_account"}`), 0644)
				if err != nil {
					t.Fatalf("Failed to create test service account file: %v", err)
				}
				tt.config.GoogleWorkspace.ServiceAccountKeyPath = serviceAccountPath
			}

			validator := NewValidator(tt.config)
			result := validator.validateConfiguration()

			if result.Status != tt.expectStatus {
				t.Errorf("Expected status %s, got %s", tt.expectStatus, result.Status)
			}

			if result.Component != "Configuration" {
				t.Errorf("Expected component 'Configuration', got %s", result.Component)
			}

			if result.Duration == 0 {
				t.Error("Expected duration to be measured")
			}
		})
	}
}

func TestValidateEnvironment(t *testing.T) {
	tests := []struct {
		name         string
		config       *config.Config
		setupFiles   map[string]string
		expectStatus string
	}{
		{
			name: "valid environment",
			config: &config.Config{
				BeyondIdentity: config.BeyondIdentityConfig{
					APIToken: "test-token",
				},
				GoogleWorkspace: config.GoogleWorkspaceConfig{
					ServiceAccountKeyPath: "test-service-account.json",
				},
			},
			setupFiles: map[string]string{
				"test-service-account.json": `{"type": "service_account"}`,
			},
			expectStatus: "PASS",
		},
		{
			name: "missing API token",
			config: &config.Config{
				BeyondIdentity: config.BeyondIdentityConfig{
					APIToken: "",
				},
				GoogleWorkspace: config.GoogleWorkspaceConfig{
					ServiceAccountKeyPath: "test-service-account.json",
				},
			},
			setupFiles: map[string]string{
				"test-service-account.json": `{"type": "service_account"}`,
			},
			expectStatus: "FAIL",
		},
		{
			name: "missing service account file",
			config: &config.Config{
				BeyondIdentity: config.BeyondIdentityConfig{
					APIToken: "test-token",
				},
				GoogleWorkspace: config.GoogleWorkspaceConfig{
					ServiceAccountKeyPath: "nonexistent.json",
				},
			},
			setupFiles:   map[string]string{},
			expectStatus: "FAIL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for test files
			tmpDir := t.TempDir()
			oldWd, _ := os.Getwd()
			defer func() { _ = os.Chdir(oldWd) }()
			_ = os.Chdir(tmpDir)

			// Create test files
			for filename, content := range tt.setupFiles {
				err := os.WriteFile(filename, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file %s: %v", filename, err)
				}
			}

			validator := NewValidator(tt.config)
			result := validator.validateEnvironment()

			if result.Status != tt.expectStatus {
				t.Errorf("Expected status %s, got %s", tt.expectStatus, result.Status)
			}

			if result.Component != "Environment" {
				t.Errorf("Expected component 'Environment', got %s", result.Component)
			}
		})
	}
}

func TestValidateGroups(t *testing.T) {
	tests := []struct {
		name         string
		config       *config.Config
		expectStatus string
	}{
		{
			name: "groups configured",
			config: &config.Config{
				Sync: config.SyncConfig{
					Groups: []string{"group1@test.com", "group2@test.com"},
				},
			},
			expectStatus: "PASS",
		},
		{
			name: "no groups configured",
			config: &config.Config{
				Sync: config.SyncConfig{
					Groups: []string{},
				},
			},
			expectStatus: "FAIL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewValidator(tt.config)
			result := validator.validateGroups()

			if result.Status != tt.expectStatus {
				t.Errorf("Expected status %s, got %s", tt.expectStatus, result.Status)
			}

			if result.Component != "Groups" {
				t.Errorf("Expected component 'Groups', got %s", result.Component)
			}
		})
	}
}

func TestValidationResult(t *testing.T) {
	result := &ValidationResult{
		Component: "Test",
		Status:    "PASS",
		Message:   "Test passed",
		Details:   "Additional details",
		Duration:  100 * time.Millisecond,
	}

	if result.Component != "Test" {
		t.Errorf("Expected component 'Test', got %s", result.Component)
	}

	if result.Status != "PASS" {
		t.Errorf("Expected status 'PASS', got %s", result.Status)
	}

	if result.Message != "Test passed" {
		t.Errorf("Expected message 'Test passed', got %s", result.Message)
	}

	if result.Details != "Additional details" {
		t.Errorf("Expected details 'Additional details', got %s", result.Details)
	}

	if result.Duration != 100*time.Millisecond {
		t.Errorf("Expected duration 100ms, got %v", result.Duration)
	}
}

func TestValidationSummary(t *testing.T) {
	summary := &ValidationSummary{
		OverallStatus: "PASS",
		TotalChecks:   3,
		Passed:        2,
		Failed:        1,
		Results: []*ValidationResult{
			{Component: "Test1", Status: "PASS"},
			{Component: "Test2", Status: "PASS"},
			{Component: "Test3", Status: "FAIL"},
		},
		Duration: 500 * time.Millisecond,
	}

	if summary.OverallStatus != "PASS" {
		t.Errorf("Expected overall status 'PASS', got %s", summary.OverallStatus)
	}

	if summary.TotalChecks != 3 {
		t.Errorf("Expected 3 total checks, got %d", summary.TotalChecks)
	}

	if summary.Passed != 2 {
		t.Errorf("Expected 2 passed checks, got %d", summary.Passed)
	}

	if summary.Failed != 1 {
		t.Errorf("Expected 1 failed check, got %d", summary.Failed)
	}

	if len(summary.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(summary.Results))
	}

	if summary.Duration != 500*time.Millisecond {
		t.Errorf("Expected duration 500ms, got %v", summary.Duration)
	}
}

func TestAddResult(t *testing.T) {
	validator := NewValidator(&config.Config{})
	summary := &ValidationSummary{
		Results: make([]*ValidationResult, 0),
	}

	result := &ValidationResult{
		Component: "Test",
		Status:    "PASS",
	}

	validator.addResult(summary, result)

	if len(summary.Results) != 1 {
		t.Errorf("Expected 1 result after adding, got %d", len(summary.Results))
	}

	if summary.Results[0] != result {
		t.Error("Expected added result to match input result")
	}
}
