package wizard

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gobeyondidentity/google-workspace-provisioner/internal/config"
)

// ANSI color codes
const (
	colorReset = "\033[0m"
	colorTeal  = "\033[36m"
	colorRed   = "\033[31m"
)

// Wizard handles interactive configuration setup
type Wizard struct {
	reader *bufio.Reader
	config *config.Config
}

// NewWizard creates a new configuration wizard
func NewWizard() *Wizard {
	// Create reader with larger buffer to handle long API tokens
	reader := bufio.NewReaderSize(os.Stdin, 8192)
	return &Wizard{
		reader: reader,
		config: &config.Config{},
	}
}

// Run starts the interactive configuration wizard
func (w *Wizard) Run() error {
	fmt.Println("Welcome to the Go SCIM Sync Configuration Wizard!")
	fmt.Println("This wizard will help you set up your configuration for syncing users from Google Workspace to Beyond Identity.")
	fmt.Println()

	// Application settings
	if err := w.configureApp(); err != nil {
		return fmt.Errorf("failed to configure app settings: %w", err)
	}

	// Google Workspace settings
	if err := w.configureGoogleWorkspace(); err != nil {
		return fmt.Errorf("failed to configure Google Workspace: %w", err)
	}

	// Beyond Identity settings
	if err := w.configureBeyondIdentity(); err != nil {
		return fmt.Errorf("failed to configure Beyond Identity: %w", err)
	}

	// Sync settings
	if err := w.configureSync(); err != nil {
		return fmt.Errorf("failed to configure sync settings: %w", err)
	}

	// Server settings
	if err := w.configureServer(); err != nil {
		return fmt.Errorf("failed to configure server settings: %w", err)
	}

	// Set defaults and validate (skip API token validation if not set)
	w.config.SetDefaults()
	skipAPIToken := w.config.BeyondIdentity.APIToken == ""
	if err := w.config.ValidateWithOptions(config.ValidateOptions{SkipAPIToken: skipAPIToken}); err != nil {
		fmt.Printf("%sConfiguration validation failed: %v%s\n", colorRed, err, colorReset)
		fmt.Println("Please review your settings and try again.")
		fmt.Println()
		fmt.Println("You can:")
		fmt.Println("1. Create the missing files and run the wizard again")
		fmt.Println("2. Edit the generated config.yaml manually")
		fmt.Println("3. Use './scim-sync validate-config' to check your configuration")
		return nil // Exit gracefully without showing CLI help
	}

	// Save configuration
	return w.saveConfiguration()
}

// configureApp configures application-level settings
func (w *Wizard) configureApp() error {
	fmt.Printf("%sApplication Settings%s\n", colorTeal, colorReset)
	fmt.Println("═══════════════════════")

	// Log level
	logLevel := w.promptWithDefault("Log level (debug, info, warn, error)", "info")
	w.config.App.LogLevel = logLevel

	// Test mode
	testMode := w.promptYesNo("Enable test mode? (recommended for first run)", true)
	w.config.App.TestMode = testMode

	if testMode {
		fmt.Println("Test mode enabled - no actual changes will be made during sync operations")
	}

	fmt.Println()
	return nil
}

// configureGoogleWorkspace configures Google Workspace settings
func (w *Wizard) configureGoogleWorkspace() error {
	fmt.Printf("%sGoogle Workspace Configuration%s\n", colorTeal, colorReset)
	fmt.Println("═════════════════════════════════")

	// Domain
	domain := w.promptRequired("Google Workspace domain (e.g., company.com)")
	w.config.GoogleWorkspace.Domain = domain

	// Super admin email
	defaultAdmin := fmt.Sprintf("admin@%s", domain)
	adminEmail := w.promptWithDefault("Super admin email", defaultAdmin)
	w.config.GoogleWorkspace.SuperAdminEmail = adminEmail

	// Service account key path
	fmt.Println("\nService Account Setup:")
	fmt.Println("You need a Google Cloud service account with domain-wide delegation.")
	fmt.Println("See: https://developers.google.com/admin-sdk/directory/v1/guides/delegation")

	keyPath := w.promptRequired("Path to service account JSON file")

	// Expand relative paths
	if !filepath.IsAbs(keyPath) {
		cwd, _ := os.Getwd()
		keyPath = filepath.Join(cwd, keyPath)
	}

	// Check if file exists
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		fmt.Printf("%sWarning: File does not exist at %s%s\n", colorRed, keyPath, colorReset)
		fmt.Println("Make sure to place your service account file there before running sync.")
	} else {
		fmt.Println("Service account file found")
	}

	w.config.GoogleWorkspace.ServiceAccountKeyPath = keyPath

	fmt.Println()
	return nil
}

// configureBeyondIdentity configures Beyond Identity settings
func (w *Wizard) configureBeyondIdentity() error {
	fmt.Printf("%sBeyond Identity Configuration%s\n", colorTeal, colorReset)
	fmt.Println("═══════════════════════════════")

	// API token
	fmt.Println("API Token Setup:")
	fmt.Println("You need a Beyond Identity API token with SCIM permissions.")
	fmt.Println()

	token := w.promptAPIToken("Beyond Identity API token")
	w.config.BeyondIdentity.APIToken = token

	if token == "" {
		fmt.Printf("%sAPI token not set - you'll need to add it to config.yaml manually%s\n", colorRed, colorReset)
	} else {
		fmt.Println("API token configured")
	}

	// SCIM base URL
	scimURL := w.promptWithDefault("SCIM API base URL", "https://api.byndid.com/scim/v2")
	w.config.BeyondIdentity.SCIMBaseURL = scimURL

	// Native API URL
	nativeURL := w.promptWithDefault("Native API base URL", "https://api.byndid.com/v2")
	w.config.BeyondIdentity.NativeAPIURL = nativeURL

	// Group prefix
	groupPrefix := w.promptWithDefault("Group name prefix", "GoogleSCIM_")
	w.config.BeyondIdentity.GroupPrefix = groupPrefix

	fmt.Println()
	return nil
}

// configureSync configures synchronization settings
func (w *Wizard) configureSync() error {
	fmt.Printf("%sSynchronization Settings%s\n", colorTeal, colorReset)
	fmt.Println("═══════════════════════════")

	// Groups to sync
	fmt.Println("Groups to Sync:")
	fmt.Println("Enter Google Workspace group email addresses to sync.")
	fmt.Println("Press Enter on an empty line when done.")

	var groups []string
	for {
		group := w.prompt(fmt.Sprintf("Group %d email (or press Enter to finish)", len(groups)+1))
		if group == "" {
			break
		}

		// Basic email validation
		if !strings.Contains(group, "@") {
			fmt.Printf("%sPlease enter a valid email address%s\n", colorRed, colorReset)
			continue
		}

		groups = append(groups, group)
		fmt.Printf("Added: %s\n", group)
	}

	if len(groups) == 0 {
		// Add at least one group
		group := w.promptRequired("At least one group is required")
		groups = append(groups, group)
	}

	w.config.Sync.Groups = groups

	// Retry settings
	retryAttempts := w.promptIntWithDefault("Retry attempts for failed operations", 3)
	w.config.Sync.RetryAttempts = retryAttempts

	retryDelay := w.promptIntWithDefault("Retry delay (seconds)", 30)
	w.config.Sync.RetryDelaySeconds = retryDelay

	fmt.Println()
	return nil
}

// configureServer configures server mode settings
func (w *Wizard) configureServer() error {
	fmt.Printf("%sServer Mode Configuration%s\n", colorTeal, colorReset)
	fmt.Println("════════════════════════════")

	// Port
	port := w.promptIntWithDefault("HTTP server port", 8080)
	w.config.Server.Port = port

	// Scheduling
	enableScheduling := w.promptYesNo("Enable automatic sync scheduling?", false)
	w.config.Server.ScheduleEnabled = enableScheduling

	if enableScheduling {
		fmt.Println("\nSchedule Configuration:")
		fmt.Println("Enter a cron schedule expression.")
		fmt.Println("Examples:")
		fmt.Println("  '0 */6 * * *'   - Every 6 hours")
		fmt.Println("  '0 0 * * *'     - Daily at midnight")
		fmt.Println("  '0 9 * * 1-5'   - Weekdays at 9 AM")

		schedule := w.promptWithDefault("Cron schedule", "0 */6 * * *")
		w.config.Server.Schedule = schedule

		fmt.Printf("Scheduled sync: %s\n", schedule)
	} else {
		w.config.Server.Schedule = "0 */6 * * *" // Default, but disabled
		fmt.Println("Manual sync only - use HTTP API to trigger syncs")
	}

	fmt.Println()
	return nil
}

// saveConfiguration saves the configuration to a file
func (w *Wizard) saveConfiguration() error {
	fmt.Printf("%sSave Configuration%s\n", colorTeal, colorReset)
	fmt.Println("════════════════════")

	// Default config path
	defaultPath := "./config.yaml"
	configPath := w.promptWithDefault("Configuration file path", defaultPath)

	// Create directory if needed
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(configPath); err == nil {
		overwrite := w.promptYesNo(fmt.Sprintf("File %s already exists. Overwrite?", configPath), false)
		if !overwrite {
			fmt.Printf("%sConfiguration not saved%s\n", colorRed, colorReset)
			return fmt.Errorf("user chose not to overwrite existing file")
		}
	}

	// Save configuration
	if err := config.Save(w.config, configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Configuration saved to: %s\n", configPath)
	fmt.Println()

	// Show next steps
	w.showNextSteps(configPath)

	return nil
}

// showNextSteps displays next steps for the user
func (w *Wizard) showNextSteps(configPath string) {
	fmt.Printf("%sSetup Complete!%s\n", colorTeal, colorReset)
	fmt.Println("═════════════════")
	fmt.Println()

	// Check if API token was set
	if w.config.BeyondIdentity.APIToken == "" {
		fmt.Printf("%sImportant: Your API token is not set!%s\n", colorRed, colorReset)
		fmt.Printf("   Edit %s and add your Beyond Identity API token to:\n", configPath)
		fmt.Println("   beyond_identity.api_token: \"your-actual-token-here\"")
		fmt.Println()
	}

	fmt.Println("Next steps:")
	fmt.Println("1. Validate config:   ./scim-sync validate-config")
	fmt.Println("2. Test sync:         ./scim-sync run")
	fmt.Println("3. Start server:      ./scim-sync server")
	fmt.Println()
	fmt.Println("Documentation:")
	fmt.Println("   - Run './scim-sync --help' for command options")
	fmt.Println("   - Server API will be available at http://localhost:8080")
	fmt.Println("   - Health check: curl http://localhost:8080/health")
	fmt.Println()

	if w.config.App.TestMode {
		fmt.Println("Test mode is enabled - no actual changes will be made")
		fmt.Println("   Set 'test_mode: false' in config when ready for production")
	}
}

// Helper methods for prompting user input

func (w *Wizard) prompt(question string) string {
	fmt.Printf("%s: ", question)
	input, err := w.reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		return ""
	}
	return strings.TrimSpace(input)
}

func (w *Wizard) promptRequired(question string) string {
	for {
		value := w.prompt(question)
		if value != "" {
			return value
		}
		fmt.Printf("%sThis field is required%s\n", colorRed, colorReset)
	}
}

func (w *Wizard) promptWithDefault(question, defaultValue string) string {
	value := w.prompt(fmt.Sprintf("%s [%s]", question, defaultValue))
	if value == "" {
		return defaultValue
	}
	return value
}

func (w *Wizard) promptYesNo(question string, defaultValue bool) bool {
	defaultStr := "y/N"
	if defaultValue {
		defaultStr = "Y/n"
	}

	for {
		response := w.prompt(fmt.Sprintf("%s [%s]", question, defaultStr))
		if response == "" {
			return defaultValue
		}

		response = strings.ToLower(response)
		if response == "y" || response == "yes" {
			return true
		}
		if response == "n" || response == "no" {
			return false
		}

		fmt.Printf("%sPlease enter 'y' or 'n'%s\n", colorRed, colorReset)
	}
}

func (w *Wizard) promptIntWithDefault(question string, defaultValue int) int {
	for {
		response := w.prompt(fmt.Sprintf("%s [%d]", question, defaultValue))
		if response == "" {
			return defaultValue
		}

		if value, err := strconv.Atoi(response); err == nil {
			return value
		}

		fmt.Printf("%sPlease enter a valid number%s\n", colorRed, colorReset)
	}
}

func (w *Wizard) promptAPIToken(question string) string {
	fmt.Printf("%s\n", question)
	fmt.Println("Read from file path:")

	return w.promptTokenFromFile()
}


func (w *Wizard) promptTokenFromFile() string {
	filePath := w.promptRequired("Path to file containing API token")

	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("%sError reading file: %v%s\n", colorRed, err, colorReset)
		return ""
	}

	token := strings.TrimSpace(string(content))
	if token == "" {
		fmt.Printf("%sFile is empty%s\n", colorRed, colorReset)
		return ""
	}

	if w.validateToken(token) {
		fmt.Println("Token loaded from file")
		return token
	}

	return ""
}

func (w *Wizard) extractTokenFromPythonConfig(pythonConfigPath string) string {
	content, err := os.ReadFile(pythonConfigPath)
	if err != nil {
		fmt.Printf("%sError reading %s: %v%s\n", colorRed, pythonConfigPath, err, colorReset)
		return ""
	}

	// Look for BI_TENANT_API_TOKEN = "..." pattern
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "BI_TENANT_API_TOKEN = \"") {
			// Extract token between quotes
			start := strings.Index(line, "\"") + 1
			end := strings.LastIndex(line, "\"")
			if start > 0 && end > start {
				token := line[start:end]
				if w.validateToken(token) {
					fmt.Println("Token extracted from Python config")
					return token
				}
			}
		}
	}

	fmt.Printf("%sCould not find valid token in %s%s\n", colorRed, pythonConfigPath, colorReset)
	return ""
}

func (w *Wizard) validateToken(token string) bool {
	// Basic JWT validation - should have 3 parts separated by dots
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		fmt.Printf("%sToken should be a JWT (3 parts separated by dots), got %d parts%s\n", colorRed, len(parts), colorReset)
		fmt.Println("   Make sure you copied the complete token")
		return false
	}

	// Check minimum length (JWTs are typically quite long)
	if len(token) < 100 {
		fmt.Printf("%sToken seems short (%d chars). Make sure you copied the complete token%s\n", colorRed, len(token), colorReset)
		return false
	}

	return true
}
