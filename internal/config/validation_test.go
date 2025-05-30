package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorFields []string // Expected field names in error
	}{
		{
			name: "valid config",
			config: &Config{
				App: AppConfig{
					LogLevel: "info",
				},
				GoogleWorkspace: GoogleWorkspaceConfig{
					Domain:                 "test.com",
					SuperAdminEmail:        "admin@test.com",
					ServiceAccountKeyPath:  "/tmp/test.json",
				},
				BeyondIdentity: BeyondIdentityConfig{
					APIToken: "test-token",
				},
				Sync: SyncConfig{
					Groups: []string{"group1@test.com"},
				},
				Server: ServerConfig{
					Port: 8080,
				},
			},
			expectError: false,
		},
		{
			name: "missing required fields",
			config: &Config{
				App: AppConfig{
					LogLevel: "info",
				},
			},
			expectError: true,
			errorFields: []string{
				"google_workspace.domain",
				"google_workspace.super_admin_email",
				"google_workspace.service_account_key_path",
				"beyond_identity.api_token",
				"sync.groups",
			},
		},
		{
			name: "invalid log level",
			config: &Config{
				App: AppConfig{
					LogLevel: "invalid",
				},
				GoogleWorkspace: GoogleWorkspaceConfig{
					Domain:                 "test.com",
					SuperAdminEmail:        "admin@test.com",
					ServiceAccountKeyPath:  "/tmp/test.json",
				},
				BeyondIdentity: BeyondIdentityConfig{
					APIToken: "test-token",
				},
				Sync: SyncConfig{
					Groups: []string{"group1@test.com"},
				},
			},
			expectError: true,
			errorFields: []string{"app.log_level"},
		},
		{
			name: "invalid email format in groups",
			config: &Config{
				App: AppConfig{
					LogLevel: "info",
				},
				GoogleWorkspace: GoogleWorkspaceConfig{
					Domain:                 "test.com",
					SuperAdminEmail:        "admin@test.com",
					ServiceAccountKeyPath:  "/tmp/test.json",
				},
				BeyondIdentity: BeyondIdentityConfig{
					APIToken: "test-token",
				},
				Sync: SyncConfig{
					Groups: []string{"invalid-email", "valid@test.com"},
				},
			},
			expectError: true,
			errorFields: []string{"sync.groups[0]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp service account file if path is specified
			if tt.config.GoogleWorkspace.ServiceAccountKeyPath != "" && !tt.expectError {
				tmpDir := t.TempDir()
				serviceAccountPath := filepath.Join(tmpDir, "test.json")
				err := os.WriteFile(serviceAccountPath, []byte(`{"type": "service_account"}`), 0644)
				if err != nil {
					t.Fatalf("Failed to create test service account file: %v", err)
				}
				tt.config.GoogleWorkspace.ServiceAccountKeyPath = serviceAccountPath
			}

			err := tt.config.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error, got nil")
					return
				}

				// Check that expected fields are mentioned in error
				errorStr := err.Error()
				for _, field := range tt.errorFields {
					if !containsField(errorStr, field) {
						t.Errorf("Expected error to mention field '%s', but error was: %s", field, errorStr)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestValidateWithOptions(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		options     ValidateOptions
		expectError bool
	}{
		{
			name: "skip API token validation",
			config: &Config{
				App: AppConfig{
					LogLevel: "info",
				},
				GoogleWorkspace: GoogleWorkspaceConfig{
					Domain:                 "test.com",
					SuperAdminEmail:        "admin@test.com",
					ServiceAccountKeyPath:  "/tmp/test.json",
				},
				BeyondIdentity: BeyondIdentityConfig{
					APIToken: "", // Empty token
				},
				Sync: SyncConfig{
					Groups: []string{"group1@test.com"},
				},
				Server: ServerConfig{
					Port: 8080,
				},
			},
			options: ValidateOptions{
				SkipAPIToken: true,
			},
			expectError: false,
		},
		{
			name: "do not skip API token validation",
			config: &Config{
				App: AppConfig{
					LogLevel: "info",
				},
				GoogleWorkspace: GoogleWorkspaceConfig{
					Domain:                 "test.com",
					SuperAdminEmail:        "admin@test.com",
					ServiceAccountKeyPath:  "/tmp/test.json",
				},
				BeyondIdentity: BeyondIdentityConfig{
					APIToken: "", // Empty token
				},
				Sync: SyncConfig{
					Groups: []string{"group1@test.com"},
				},
				Server: ServerConfig{
					Port: 8080,
				},
			},
			options: ValidateOptions{
				SkipAPIToken: false,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp service account file
			tmpDir := t.TempDir()
			serviceAccountPath := filepath.Join(tmpDir, "test.json")
			err := os.WriteFile(serviceAccountPath, []byte(`{"type": "service_account"}`), 0644)
			if err != nil {
				t.Fatalf("Failed to create test service account file: %v", err)
			}
			tt.config.GoogleWorkspace.ServiceAccountKeyPath = serviceAccountPath

			err = tt.config.ValidateWithOptions(tt.options)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := ValidationError{
		Field:   "test.field",
		Message: "test message",
	}

	expected := "validation error for field 'test.field': test message"
	if err.Error() != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, err.Error())
	}
}

func TestValidationErrors(t *testing.T) {
	errors := ValidationErrors{
		ValidationError{Field: "field1", Message: "message1"},
		ValidationError{Field: "field2", Message: "message2"},
	}

	expected := "validation error for field 'field1': message1; validation error for field 'field2': message2"
	if errors.Error() != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, errors.Error())
	}
}

// Helper function to check if an error string contains a field name
func containsField(errorStr string, field string) bool {
	return strings.Contains(errorStr, field)
}