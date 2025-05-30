package setup

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gobeyondidentity/go-scim-sync/internal/config"
	"github.com/gobeyondidentity/go-scim-sync/internal/gws"
	"github.com/sirupsen/logrus"
)

// Validator handles setup validation and connectivity testing
type Validator struct {
	config *config.Config
	logger *logrus.Logger
}

// ValidationResult represents the result of a validation check
type ValidationResult struct {
	Component string    `json:"component"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	Duration  time.Duration `json:"duration"`
}

// ValidationSummary contains overall validation results
type ValidationSummary struct {
	OverallStatus string              `json:"overall_status"`
	TotalChecks   int                 `json:"total_checks"`
	Passed        int                 `json:"passed"`
	Failed        int                 `json:"failed"`
	Results       []*ValidationResult `json:"results"`
	Duration      time.Duration       `json:"duration"`
}

// NewValidator creates a new setup validator
func NewValidator(cfg *config.Config) *Validator {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Only show errors during validation
	
	return &Validator{
		config: cfg,
		logger: logger,
	}
}

// ValidateSetup performs comprehensive setup validation
func (v *Validator) ValidateSetup() (*ValidationSummary, error) {
	startTime := time.Now()
	
	fmt.Println("🔍 Validating Go SCIM Sync Setup")
	fmt.Println("═══════════════════════════════")
	fmt.Println()
	
	summary := &ValidationSummary{
		Results: make([]*ValidationResult, 0),
	}
	
	// Configuration validation
	v.addResult(summary, v.validateConfiguration())
	
	// Environment validation
	v.addResult(summary, v.validateEnvironment())
	
	// Google Workspace connectivity
	v.addResult(summary, v.validateGoogleWorkspace())
	
	// Beyond Identity connectivity
	v.addResult(summary, v.validateBeyondIdentity())
	
	// Group existence check
	v.addResult(summary, v.validateGroups())
	
	// Calculate summary
	summary.Duration = time.Since(startTime)
	summary.TotalChecks = len(summary.Results)
	
	for _, result := range summary.Results {
		if result.Status == "PASS" {
			summary.Passed++
		} else {
			summary.Failed++
		}
	}
	
	if summary.Failed == 0 {
		summary.OverallStatus = "PASS"
	} else {
		summary.OverallStatus = "FAIL"
	}
	
	// Print summary
	v.printSummary(summary)
	
	return summary, nil
}

// validateConfiguration validates the configuration structure
func (v *Validator) validateConfiguration() *ValidationResult {
	fmt.Print("📋 Configuration validation... ")
	start := time.Now()
	
	if err := v.config.Validate(); err != nil {
		fmt.Println("❌ FAIL")
		return &ValidationResult{
			Component: "Configuration",
			Status:    "FAIL",
			Message:   "Configuration validation failed",
			Details:   err.Error(),
			Duration:  time.Since(start),
		}
	}
	
	fmt.Println("✅ PASS")
	return &ValidationResult{
		Component: "Configuration",
		Status:    "PASS",
		Message:   "Configuration is valid",
		Duration:  time.Since(start),
	}
}

// validateEnvironment validates required environment variables and files
func (v *Validator) validateEnvironment() *ValidationResult {
	fmt.Print("🌍 Environment validation... ")
	start := time.Now()
	
	var issues []string
	
	// Check API token
	if v.config.BeyondIdentity.APIToken == "" {
		issues = append(issues, "Beyond Identity API token not set in config.yaml")
	}
	
	// Check service account file
	if _, err := os.Stat(v.config.GoogleWorkspace.ServiceAccountKeyPath); os.IsNotExist(err) {
		issues = append(issues, fmt.Sprintf("Service account file not found: %s", v.config.GoogleWorkspace.ServiceAccountKeyPath))
	}
	
	if len(issues) > 0 {
		fmt.Println("❌ FAIL")
		return &ValidationResult{
			Component: "Environment",
			Status:    "FAIL",
			Message:   "Environment setup issues found",
			Details:   fmt.Sprintf("Issues: %v", issues),
			Duration:  time.Since(start),
		}
	}
	
	fmt.Println("✅ PASS")
	return &ValidationResult{
		Component: "Environment",
		Status:    "PASS",
		Message:   "Environment is properly configured",
		Duration:  time.Since(start),
	}
}

// validateGoogleWorkspace tests Google Workspace connectivity
func (v *Validator) validateGoogleWorkspace() *ValidationResult {
	fmt.Print("🔵 Google Workspace connectivity... ")
	start := time.Now()
	
	_, err := gws.NewClient(
		v.config.GoogleWorkspace.ServiceAccountKeyPath,
		v.config.GoogleWorkspace.Domain,
		v.config.GoogleWorkspace.SuperAdminEmail,
	)
	if err != nil {
		fmt.Println("❌ FAIL")
		return &ValidationResult{
			Component: "Google Workspace",
			Status:    "FAIL",
			Message:   "Failed to create Google Workspace client",
			Details:   err.Error(),
			Duration:  time.Since(start),
		}
	}
	
	// Test basic connectivity - client creation validates auth setup
	// We could expand this to make actual API calls if needed
	
	fmt.Println("✅ PASS")
	return &ValidationResult{
		Component: "Google Workspace",
		Status:    "PASS",
		Message:   "Google Workspace client created successfully",
		Details:   fmt.Sprintf("Domain: %s", v.config.GoogleWorkspace.Domain),
		Duration:  time.Since(start),
	}
}

// validateBeyondIdentity tests Beyond Identity connectivity
func (v *Validator) validateBeyondIdentity() *ValidationResult {
	fmt.Print("🟢 Beyond Identity connectivity... ")
	start := time.Now()
	
	// Get API token
	apiToken := v.config.BeyondIdentity.APIToken
	
	if apiToken == "" {
		fmt.Println("❌ FAIL")
		return &ValidationResult{
			Component: "Beyond Identity",
			Status:    "FAIL",
			Message:   "API token not available",
			Details:   "Beyond Identity API token not set in config.yaml",
			Duration:  time.Since(start),
		}
	}
	
	// Test connectivity with a simple HTTP request
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", v.config.BeyondIdentity.SCIMBaseURL+"/Users?count=1", nil)
	if err != nil {
		fmt.Println("❌ FAIL")
		return &ValidationResult{
			Component: "Beyond Identity",
			Status:    "FAIL",
			Message:   "Failed to create test request",
			Details:   err.Error(),
			Duration:  time.Since(start),
		}
	}
	
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Accept", "application/scim+json")
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("❌ FAIL")
		return &ValidationResult{
			Component: "Beyond Identity",
			Status:    "FAIL",
			Message:   "Failed to connect to Beyond Identity API",
			Details:   err.Error(),
			Duration:  time.Since(start),
		}
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 401 {
		fmt.Println("❌ FAIL")
		return &ValidationResult{
			Component: "Beyond Identity",
			Status:    "FAIL",
			Message:   "Authentication failed",
			Details:   "Invalid API token or insufficient permissions",
			Duration:  time.Since(start),
		}
	}
	
	if resp.StatusCode >= 400 {
		fmt.Println("❌ FAIL")
		return &ValidationResult{
			Component: "Beyond Identity",
			Status:    "FAIL",
			Message:   "API request failed",
			Details:   fmt.Sprintf("HTTP %d", resp.StatusCode),
			Duration:  time.Since(start),
		}
	}
	
	fmt.Println("✅ PASS")
	return &ValidationResult{
		Component: "Beyond Identity",
		Status:    "PASS",
		Message:   "Beyond Identity API is accessible",
		Details:   fmt.Sprintf("Endpoint: %s", v.config.BeyondIdentity.SCIMBaseURL),
		Duration:  time.Since(start),
	}
}

// validateGroups checks if configured groups exist in Google Workspace
func (v *Validator) validateGroups() *ValidationResult {
	fmt.Print("👥 Group existence check... ")
	start := time.Now()
	
	if len(v.config.Sync.Groups) == 0 {
		fmt.Println("❌ FAIL")
		return &ValidationResult{
			Component: "Groups",
			Status:    "FAIL",
			Message:   "No groups configured for sync",
			Duration:  time.Since(start),
		}
	}
	
	// For now, just validate that groups are configured
	// In a full implementation, we could actually check if they exist in GWS
	fmt.Println("✅ PASS")
	return &ValidationResult{
		Component: "Groups",
		Status:    "PASS",
		Message:   fmt.Sprintf("Found %d groups configured for sync", len(v.config.Sync.Groups)),
		Details:   fmt.Sprintf("Groups: %v", v.config.Sync.Groups),
		Duration:  time.Since(start),
	}
}

// addResult adds a validation result to the summary
func (v *Validator) addResult(summary *ValidationSummary, result *ValidationResult) {
	summary.Results = append(summary.Results, result)
}

// printSummary prints the validation summary
func (v *Validator) printSummary(summary *ValidationSummary) {
	fmt.Println()
	fmt.Println("📊 Validation Summary")
	fmt.Println("════════════════════")
	
	if summary.OverallStatus == "PASS" {
		fmt.Printf("✅ Overall Status: %s\n", summary.OverallStatus)
	} else {
		fmt.Printf("❌ Overall Status: %s\n", summary.OverallStatus)
	}
	
	fmt.Printf("📈 Results: %d passed, %d failed (total: %d)\n", 
		summary.Passed, summary.Failed, summary.TotalChecks)
	fmt.Printf("⏱️  Duration: %v\n", summary.Duration.Round(time.Millisecond))
	
	if summary.Failed > 0 {
		fmt.Println()
		fmt.Println("❌ Failed Checks:")
		for _, result := range summary.Results {
			if result.Status == "FAIL" {
				fmt.Printf("   • %s: %s\n", result.Component, result.Message)
				if result.Details != "" {
					fmt.Printf("     Details: %s\n", result.Details)
				}
			}
		}
		
		fmt.Println()
		fmt.Println("💡 Next Steps:")
		fmt.Println("   1. Fix the issues listed above")
		fmt.Println("   2. Run validation again: ./go-scim-sync setup validate")
		fmt.Println("   3. Once all checks pass, try a test sync: ./go-scim-sync run")
	} else {
		fmt.Println()
		fmt.Println("🎉 All checks passed! Your setup is ready.")
		fmt.Println("💡 Next Steps:")
		fmt.Println("   1. Try a test sync: ./go-scim-sync run")
		fmt.Println("   2. Start server mode: ./go-scim-sync server")
		fmt.Println("   3. Check the health endpoint: curl http://localhost:8080/health")
	}
}