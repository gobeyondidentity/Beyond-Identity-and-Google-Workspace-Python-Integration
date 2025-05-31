package logger

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

// PythonCompatibleFormatter formats log messages to match Python output
type PythonCompatibleFormatter struct{}

// Format formats a logrus entry to match Python logging format
// Python format: 2025-05-30 12:21:53,426 - INFO - Starting sync process
func (f *PythonCompatibleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05,000")
	level := entry.Level.String()
	message := entry.Message

	// Convert level to uppercase to match Python
	switch entry.Level {
	case logrus.DebugLevel:
		level = "DEBUG"
	case logrus.InfoLevel:
		level = "INFO"
	case logrus.WarnLevel:
		level = "WARNING"
	case logrus.ErrorLevel:
		level = "ERROR"
	case logrus.FatalLevel:
		level = "CRITICAL"
	case logrus.PanicLevel:
		level = "CRITICAL"
	}

	formatted := fmt.Sprintf("%s - %s - %s\n", timestamp, level, message)
	return []byte(formatted), nil
}

// Setup configures the logger with Python-compatible formatting
func Setup(logLevel string, testMode bool) *logrus.Logger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Use custom formatter for Python compatibility
	logger.SetFormatter(&PythonCompatibleFormatter{})

	// Output to stdout (matching Python behavior)
	logger.SetOutput(os.Stdout)

	// Log startup information
	logger.Info("Starting Google Workspace to Beyond Identity sync process")

	if testMode {
		logger.Info("TEST MODE ENABLED - No actual changes will be made")
	}

	return logger
}

// LogProcessStart logs the start of processing with group information
func LogProcessStart(logger *logrus.Logger, groups []string, logLevel string) {
	if len(groups) == 1 {
		logger.Infof("Configured to sync the following group: %s", groups[0])
	} else {
		logger.Infof("Configured to sync the following groups: %s", joinGroups(groups))
	}

	logger.Infof("Logging enabled at %s level", logLevel)
}

// joinGroups joins group names with proper formatting
func joinGroups(groups []string) string {
	if len(groups) == 0 {
		return ""
	}
	if len(groups) == 1 {
		return groups[0]
	}
	if len(groups) == 2 {
		return groups[0] + ", " + groups[1]
	}

	result := ""
	for i, group := range groups {
		if i == len(groups)-1 {
			result += ", " + group
		} else if i == 0 {
			result = group
		} else {
			result += ", " + group
		}
	}
	return result
}
