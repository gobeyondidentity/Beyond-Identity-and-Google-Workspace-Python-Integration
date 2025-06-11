package main

import (
	"fmt"
	"os"

	"github.com/gobeyondidentity/google-workspace-provisioner/internal/bi"
	"github.com/gobeyondidentity/google-workspace-provisioner/internal/config"
	"github.com/gobeyondidentity/google-workspace-provisioner/internal/gws"
	"github.com/gobeyondidentity/google-workspace-provisioner/internal/logger"
	"github.com/gobeyondidentity/google-workspace-provisioner/internal/server"
	"github.com/gobeyondidentity/google-workspace-provisioner/internal/setup"
	"github.com/gobeyondidentity/google-workspace-provisioner/internal/sync"
	"github.com/gobeyondidentity/google-workspace-provisioner/internal/wizard"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfg     *config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "scim-sync",
	Short: "Google Workspace to Beyond Identity SCIM synchronization tool",
	Long: `A tool for synchronizing users and groups from Google Workspace
to Beyond Identity using SCIM protocol.

This application supports two modes:
- One-shot mode: Run synchronization once and exit
- Server mode: Run continuously with scheduled synchronization and HTTP API`,
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run SCIM synchronization once",
	Long: `Run a single synchronization operation from Google Workspace to Beyond Identity.
This will sync all configured groups and their members.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSync()
	},
}

// validateConfigCmd represents the validate-config command
var validateConfigCmd = &cobra.Command{
	Use:   "validate-config",
	Short: "Validate configuration file",
	Long:  `Validate the configuration file for syntax and required fields.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return validateConfig()
	},
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run in server mode with HTTP API and optional scheduling",
	Long: `Run the application in server mode. This provides an HTTP API for manual sync operations,
health checks, and metrics. If scheduling is enabled in configuration, automatic sync operations
will run according to the specified cron schedule.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServer()
	},
}

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup and configuration wizard",
	Long:  `Interactive setup wizard to help configure Go SCIM sync for first-time use.`,
}

// setupWizardCmd represents the setup wizard subcommand
var setupWizardCmd = &cobra.Command{
	Use:   "wizard",
	Short: "Run interactive configuration wizard",
	Long:  `Run an interactive wizard to create configuration file with guided prompts.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSetupWizard()
	},
}

// setupValidateCmd represents the setup validate subcommand
var setupValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate current setup and connectivity",
	Long:  `Validate configuration file, environment variables, and test connectivity to external services.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSetupValidation()
	},
}

// setupDocsCmd represents the setup docs subcommand
var setupDocsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate setup and API documentation",
	Long:  `Generate comprehensive documentation including setup guide, API reference, and troubleshooting.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDocsGeneration()
	},
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print version information for scim-sync.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("scim-sync version 0.1.0 (MVP)")
		fmt.Println("Built with Go")
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")

	// Add setup subcommands
	setupCmd.AddCommand(setupWizardCmd)
	setupCmd.AddCommand(setupValidateCmd)
	setupCmd.AddCommand(setupDocsCmd)

	// Add commands
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(validateConfigCmd)
	rootCmd.AddCommand(versionCmd)
}

// initConfig reads in config file and ENV variables
func initConfig() {
	var err error

	if cfgFile != "" {
		// Use config file from the flag
		cfg, err = config.Load(cfgFile)
	} else {
		// Find config file in standard locations
		cfgFile, err = config.FindConfigFile()
		if err != nil {
			// Only exit on run command, not on other commands
			return
		}
		cfg, err = config.Load(cfgFile)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Set defaults
	cfg.SetDefaults()
}

// runSync executes the main synchronization logic
func runSync() error {
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Setup logger
	log := logger.Setup(cfg.App.LogLevel, cfg.App.TestMode)

	// Log process start info
	logger.LogProcessStart(log, cfg.Sync.Groups, cfg.App.LogLevel)
	log.Info("Starting main sync process")

	// Create Google Workspace client
	gwsClient, err := gws.NewClient(
		cfg.GoogleWorkspace.ServiceAccountKeyPath,
		cfg.GoogleWorkspace.Domain,
		cfg.GoogleWorkspace.SuperAdminEmail,
	)
	if err != nil {
		log.Errorf("Failed to create Google Workspace client: %v", err)
		return fmt.Errorf("failed to create Google Workspace client: %w", err)
	}

	// Create Beyond Identity client
	biClient := bi.NewClient(cfg.BeyondIdentity.APIToken, cfg.BeyondIdentity.SCIMBaseURL, cfg.BeyondIdentity.NativeAPIURL)

	// Create sync engine
	engine := sync.NewEngine(gwsClient, biClient, cfg, log)

	// Run synchronization
	result, err := engine.Sync()
	if err != nil {
		log.Errorf("Sync process failed: %v", err)
		return err
	}

	// Log final results
	if len(result.Errors) > 0 {
		log.Warnf("Sync completed with %d errors", len(result.Errors))
		for _, syncErr := range result.Errors {
			log.Errorf("Sync error: %v", syncErr)
		}
	} else {
		log.Info("Sync process completed successfully")
	}

	return nil
}

// validateConfig validates the configuration file
func validateConfig() error {
	// Load config if not already loaded
	if cfg == nil {
		var err error
		if cfgFile != "" {
			cfg, err = config.Load(cfgFile)
		} else {
			cfgFile, err = config.FindConfigFile()
			if err != nil {
				return fmt.Errorf("no config file found: %w", err)
			}
			cfg, err = config.Load(cfgFile)
		}

		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		cfg.SetDefaults()
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Configuration validation failed:\n%v\n", err)
		return err
	}

	fmt.Printf("✅ Configuration file '%s' is valid\n", cfgFile)
	fmt.Printf("   - Google Workspace domain: %s\n", cfg.GoogleWorkspace.Domain)
	fmt.Printf("   - Groups to sync: %d\n", len(cfg.Sync.Groups))
	fmt.Printf("   - Test mode: %t\n", cfg.App.TestMode)
	fmt.Printf("   - Log level: %s\n", cfg.App.LogLevel)

	return nil
}

// runServer executes server mode
func runServer() error {
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Setup logger
	log := logger.Setup(cfg.App.LogLevel, cfg.App.TestMode)

	// Log server start info
	log.Infof("Starting SCIM sync server on port %d", cfg.Server.Port)
	if cfg.Server.ScheduleEnabled {
		log.Infof("Scheduling enabled with cron: %s", cfg.Server.Schedule)
	} else {
		log.Info("Scheduling disabled - manual sync only")
	}

	// Create and start server
	srv, err := server.NewServer(cfg, log)
	if err != nil {
		log.Errorf("Failed to create server: %v", err)
		return fmt.Errorf("failed to create server: %w", err)
	}

	return srv.Start()
}

// runSetupWizard executes the interactive configuration wizard
func runSetupWizard() error {
	w := wizard.NewWizard()
	return w.Run()
}

// runSetupValidation executes setup validation
func runSetupValidation() error {
	// Load existing configuration if available
	if cfg == nil {
		var err error
		if cfgFile != "" {
			cfg, err = config.Load(cfgFile)
		} else {
			cfgFile, err = config.FindConfigFile()
			if err != nil {
				return fmt.Errorf("no config file found - run 'setup wizard' first: %w", err)
			}
			cfg, err = config.Load(cfgFile)
		}

		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		cfg.SetDefaults()
	}

	validator := setup.NewValidator(cfg)
	summary, err := validator.ValidateSetup()
	if err != nil {
		return err
	}

	// Exit with error code if validation failed
	if summary.OverallStatus != "PASS" {
		os.Exit(1)
	}

	return nil
}

// runDocsGeneration generates documentation
func runDocsGeneration() error {
	outputDir := "./docs"
	if len(os.Args) > 3 {
		outputDir = os.Args[3]
	}

	fmt.Printf("Generating documentation in %s...\n", outputDir)
	return setup.GenerateDocumentation(outputDir)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
