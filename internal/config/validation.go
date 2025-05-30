package config

import (
	"fmt"
	"os"
	"strings"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// ValidateOptions provides options for validation
type ValidateOptions struct {
	SkipAPIToken bool // Skip API token validation (useful during setup)
}

// Validate validates the configuration and returns any errors
func (c *Config) Validate() error {
	return c.ValidateWithOptions(ValidateOptions{})
}

// ValidateWithOptions validates the configuration with custom options
func (c *Config) ValidateWithOptions(opts ValidateOptions) error {
	var errors ValidationErrors

	// Validate App config
	if c.App.LogLevel != "" {
		validLevels := []string{"debug", "info", "warn", "error"}
		if !contains(validLevels, c.App.LogLevel) {
			errors = append(errors, ValidationError{
				Field:   "app.log_level",
				Message: fmt.Sprintf("must be one of: %v", validLevels),
			})
		}
	}

	// Validate Google Workspace config
	if c.GoogleWorkspace.Domain == "" {
		errors = append(errors, ValidationError{
			Field:   "google_workspace.domain",
			Message: "domain is required",
		})
	}

	if c.GoogleWorkspace.SuperAdminEmail == "" {
		errors = append(errors, ValidationError{
			Field:   "google_workspace.super_admin_email",
			Message: "super admin email is required",
		})
	}

	if c.GoogleWorkspace.ServiceAccountKeyPath == "" {
		errors = append(errors, ValidationError{
			Field:   "google_workspace.service_account_key_path",
			Message: "service account key path is required",
		})
	} else {
		// Check if service account key file exists
		if _, err := os.Stat(c.GoogleWorkspace.ServiceAccountKeyPath); os.IsNotExist(err) {
			errors = append(errors, ValidationError{
				Field:   "google_workspace.service_account_key_path",
				Message: fmt.Sprintf("service account key file not found: %s", c.GoogleWorkspace.ServiceAccountKeyPath),
			})
		}
	}

	// Validate Beyond Identity config
	if !opts.SkipAPIToken && c.BeyondIdentity.APIToken == "" {
		errors = append(errors, ValidationError{
			Field:   "beyond_identity.api_token",
			Message: "API token is required",
		})
	}

	// Validate Sync config
	if len(c.Sync.Groups) == 0 {
		errors = append(errors, ValidationError{
			Field:   "sync.groups",
			Message: "at least one group must be specified",
		})
	}

	// Validate email formats
	for i, group := range c.Sync.Groups {
		if !strings.Contains(group, "@") {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("sync.groups[%d]", i),
				Message: fmt.Sprintf("invalid email format: %s", group),
			})
		}
	}

	if c.Sync.RetryAttempts < 0 {
		errors = append(errors, ValidationError{
			Field:   "sync.retry_attempts",
			Message: "retry attempts must be non-negative",
		})
	}

	if c.Sync.RetryDelaySeconds < 0 {
		errors = append(errors, ValidationError{
			Field:   "sync.retry_delay_seconds",
			Message: "retry delay seconds must be non-negative",
		})
	}

	// Validate server configuration
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		errors = append(errors, ValidationError{
			Field:   "server.port",
			Message: "port must be between 1 and 65535",
		})
	}

	// Validate cron schedule if scheduling is enabled
	if c.Server.ScheduleEnabled && c.Server.Schedule == "" {
		errors = append(errors, ValidationError{
			Field:   "server.schedule",
			Message: "schedule must be provided when schedule_enabled is true",
		})
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}