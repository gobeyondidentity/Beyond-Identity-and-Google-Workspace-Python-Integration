package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	App             AppConfig             `yaml:"app"`
	GoogleWorkspace GoogleWorkspaceConfig `yaml:"google_workspace"`
	BeyondIdentity  BeyondIdentityConfig  `yaml:"beyond_identity"`
	Sync            SyncConfig            `yaml:"sync"`
	Server          ServerConfig          `yaml:"server"`
}

// AppConfig contains application-level settings
type AppConfig struct {
	LogLevel string `yaml:"log_level"`
	TestMode bool   `yaml:"test_mode"`
}

// GoogleWorkspaceConfig contains Google Workspace API settings
type GoogleWorkspaceConfig struct {
	Domain                string `yaml:"domain"`
	SuperAdminEmail       string `yaml:"super_admin_email"`
	ServiceAccountKeyPath string `yaml:"service_account_key_path"`
}

// BeyondIdentityConfig contains Beyond Identity API settings
type BeyondIdentityConfig struct {
	APIToken     string `yaml:"api_token"`
	SCIMBaseURL  string `yaml:"scim_base_url"`
	NativeAPIURL string `yaml:"native_api_url"`
	GroupPrefix  string `yaml:"group_prefix"`
}

// SyncConfig contains synchronization settings
type SyncConfig struct {
	Groups               []string `yaml:"groups"`
	EnrollmentGroupEmail string   `yaml:"enrollment_group_email"`
	EnrollmentGroupName  string   `yaml:"enrollment_group_name"`
	RetryAttempts        int      `yaml:"retry_attempts"`
	RetryDelaySeconds    int      `yaml:"retry_delay_seconds"`
}

// ServerConfig contains server mode settings
type ServerConfig struct {
	Port            int    `yaml:"port"`
	ScheduleEnabled bool   `yaml:"schedule_enabled"`
	Schedule        string `yaml:"schedule"`
}

// Load loads configuration from a YAML file
func Load(configPath string) (*Config, error) {
	// Read the configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Substitute environment variables
	configData := os.ExpandEnv(string(data))

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal([]byte(configData), &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	return &config, nil
}

// FindConfigFile searches for configuration file in common locations
func FindConfigFile() (string, error) {
	locations := []string{
		"./config.yaml",
		"./config.yml",
		"~/.config/scim-sync/config.yaml",
		"~/.config/scim-sync/config.yml",
	}

	for _, location := range locations {
		// Expand home directory
		if strings.HasPrefix(location, "~/") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				continue
			}
			location = strings.Replace(location, "~", homeDir, 1)
		}

		if _, err := os.Stat(location); err == nil {
			return location, nil
		}
	}

	return "", fmt.Errorf("no configuration file found in any of these locations: %v", locations)
}

// SetDefaults sets default values for configuration
func (c *Config) SetDefaults() {
	if c.App.LogLevel == "" {
		c.App.LogLevel = "info"
	}

	if c.BeyondIdentity.SCIMBaseURL == "" {
		c.BeyondIdentity.SCIMBaseURL = "https://api.byndid.com/scim/v2"
	}

	if c.BeyondIdentity.NativeAPIURL == "" {
		c.BeyondIdentity.NativeAPIURL = "https://api.byndid.com/v2"
	}

	if c.BeyondIdentity.GroupPrefix == "" {
		c.BeyondIdentity.GroupPrefix = "GoogleSCIM_"
	}

	if c.Sync.RetryAttempts == 0 {
		c.Sync.RetryAttempts = 3
	}

	if c.Sync.RetryDelaySeconds == 0 {
		c.Sync.RetryDelaySeconds = 30
	}

	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}

	if c.Server.Schedule == "" {
		c.Server.Schedule = "0 */6 * * *" // Every 6 hours by default
	}

	if c.Sync.RetryDelaySeconds == 0 {
		c.Sync.RetryDelaySeconds = 30
	}

	if c.Sync.EnrollmentGroupEmail == "" {
		c.Sync.EnrollmentGroupEmail = "byid-enrolled@" + c.GoogleWorkspace.Domain
	}

	if c.Sync.EnrollmentGroupName == "" {
		c.Sync.EnrollmentGroupName = "BYID Enrolled"
	}
}
