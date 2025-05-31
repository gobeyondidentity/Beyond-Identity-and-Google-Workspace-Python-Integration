package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		configYAML  string
		expectError bool
		validate    func(*Config) error
	}{
		{
			name: "valid config",
			configYAML: `
app:
  log_level: "info"
  test_mode: true
google_workspace:
  domain: "test.com"
  super_admin_email: "admin@test.com"
  service_account_key_path: "/tmp/test.json"
beyond_identity:
  api_token: "test-token"
  scim_base_url: "https://api.test.com/scim/v2"
  native_api_url: "https://api.test.com/v2"
  group_prefix: "Test_"
sync:
  groups:
    - "group1@test.com"
  retry_attempts: 3
  retry_delay_seconds: 30
server:
  port: 8080
  schedule_enabled: false
  schedule: "0 */6 * * *"
`,
			expectError: false,
			validate: func(c *Config) error {
				if c.App.LogLevel != "info" {
					t.Errorf("Expected log_level 'info', got '%s'", c.App.LogLevel)
				}
				if c.GoogleWorkspace.Domain != "test.com" {
					t.Errorf("Expected domain 'test.com', got '%s'", c.GoogleWorkspace.Domain)
				}
				if c.BeyondIdentity.APIToken != "test-token" {
					t.Errorf("Expected api_token 'test-token', got '%s'", c.BeyondIdentity.APIToken)
				}
				if len(c.Sync.Groups) != 1 || c.Sync.Groups[0] != "group1@test.com" {
					t.Errorf("Expected groups ['group1@test.com'], got %v", c.Sync.Groups)
				}
				return nil
			},
		},
		{
			name:        "invalid yaml",
			configYAML:  "invalid: yaml: content: [",
			expectError: true,
		},
		{
			name:        "empty config",
			configYAML:  "",
			expectError: false, // Should load with defaults
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			err := os.WriteFile(configPath, []byte(tt.configYAML), 0644)
			if err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			// Test Load
			config, err := Load(configPath)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if config == nil {
				t.Errorf("Expected config, got nil")
				return
			}

			// Run validation if provided
			if tt.validate != nil {
				if err := tt.validate(config); err != nil {
					t.Errorf("Validation failed: %v", err)
				}
			}
		})
	}
}

func TestFindConfigFile(t *testing.T) {
	tests := []struct {
		name        string
		setupFiles  map[string]string // filename -> content
		expectFound bool
		expectFile  string
	}{
		{
			name: "finds config.yaml in current dir",
			setupFiles: map[string]string{
				"config.yaml": "test: content",
			},
			expectFound: true,
			expectFile:  "config.yaml",
		},
		{
			name: "finds config.yml when yaml doesn't exist",
			setupFiles: map[string]string{
				"config.yml": "test: content",
			},
			expectFound: true,
			expectFile:  "config.yml",
		},
		{
			name:        "no config files found",
			setupFiles:  map[string]string{},
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory and change to it
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

			// Test FindConfigFile
			found, err := FindConfigFile()

			if tt.expectFound {
				if err != nil {
					t.Errorf("Expected to find config file, got error: %v", err)
				}
				if filepath.Base(found) != tt.expectFile {
					t.Errorf("Expected to find %s, got %s", tt.expectFile, filepath.Base(found))
				}
			} else {
				if err == nil {
					t.Errorf("Expected error when no config found, got file: %s", found)
				}
			}
		})
	}
}

func TestSetDefaults(t *testing.T) {
	config := &Config{}
	config.SetDefaults()

	tests := []struct {
		name     string
		expected interface{}
		actual   interface{}
	}{
		{"default log level", "info", config.App.LogLevel},
		{"default SCIM base URL", "https://api.byndid.com/scim/v2", config.BeyondIdentity.SCIMBaseURL},
		{"default native API URL", "https://api.byndid.com/v2", config.BeyondIdentity.NativeAPIURL},
		{"default group prefix", "GoogleSCIM_", config.BeyondIdentity.GroupPrefix},
		{"default retry attempts", 3, config.Sync.RetryAttempts},
		{"default retry delay", 30, config.Sync.RetryDelaySeconds},
		{"default server port", 8080, config.Server.Port},
		{"default schedule", "0 */6 * * *", config.Server.Schedule},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actual != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, tt.actual)
			}
		})
	}
}

func TestSetDefaults_PreservesExistingValues(t *testing.T) {
	config := &Config{
		App: AppConfig{
			LogLevel: "debug",
		},
		BeyondIdentity: BeyondIdentityConfig{
			SCIMBaseURL: "https://custom.api.com/scim",
		},
		Server: ServerConfig{
			Port: 9090,
		},
	}

	config.SetDefaults()

	// Should preserve existing values
	if config.App.LogLevel != "debug" {
		t.Errorf("Expected existing log level to be preserved, got %s", config.App.LogLevel)
	}
	if config.BeyondIdentity.SCIMBaseURL != "https://custom.api.com/scim" {
		t.Errorf("Expected existing SCIM URL to be preserved, got %s", config.BeyondIdentity.SCIMBaseURL)
	}
	if config.Server.Port != 9090 {
		t.Errorf("Expected existing port to be preserved, got %d", config.Server.Port)
	}

	// Should set defaults for unset values
	if config.BeyondIdentity.NativeAPIURL != "https://api.byndid.com/v2" {
		t.Errorf("Expected default native API URL, got %s", config.BeyondIdentity.NativeAPIURL)
	}
}
