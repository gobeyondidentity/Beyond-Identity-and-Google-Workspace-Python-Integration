package wizard

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gobeyondidentity/google-workspace-provisioner/internal/config"
)

func TestNewWizard(t *testing.T) {
	wizard := NewWizard()

	if wizard == nil {
		t.Error("Expected wizard to be created, got nil")
		return
	}

	if wizard.reader == nil {
		t.Error("Expected reader to be initialized")
	}

	if wizard.config == nil {
		t.Error("Expected config to be initialized")
	}
}

func TestPrompt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		question string
		expected string
	}{
		{
			name:     "simple input",
			input:    "test answer\n",
			question: "Test question",
			expected: "test answer",
		},
		{
			name:     "input with whitespace",
			input:    "  spaced answer  \n",
			question: "Test question",
			expected: "spaced answer",
		},
		{
			name:     "empty input",
			input:    "\n",
			question: "Test question",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create wizard with test input
			reader := bufio.NewReaderSize(strings.NewReader(tt.input), 8192)
			wizard := &Wizard{
				reader: reader,
				config: &config.Config{},
			}

			result := wizard.prompt(tt.question)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestPromptRequired(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid input on first try",
			input:    "valid answer\n",
			expected: "valid answer",
		},
		{
			name:     "empty then valid input",
			input:    "\nvalid answer\n",
			expected: "valid answer",
		},
		{
			name:     "multiple empty then valid",
			input:    "\n\n\nvalid answer\n",
			expected: "valid answer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReaderSize(strings.NewReader(tt.input), 8192)
			wizard := &Wizard{
				reader: reader,
				config: &config.Config{},
			}

			result := wizard.promptRequired("Test question")
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestPromptWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue string
		expected     string
	}{
		{
			name:         "use provided input",
			input:        "custom value\n",
			defaultValue: "default",
			expected:     "custom value",
		},
		{
			name:         "use default on empty input",
			input:        "\n",
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReaderSize(strings.NewReader(tt.input), 8192)
			wizard := &Wizard{
				reader: reader,
				config: &config.Config{},
			}

			result := wizard.promptWithDefault("Test question", tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestPromptYesNo(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue bool
		expected     bool
	}{
		{
			name:         "yes input",
			input:        "y\n",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "yes full word",
			input:        "yes\n",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "no input",
			input:        "n\n",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "no full word",
			input:        "no\n",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "use default on empty",
			input:        "\n",
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "invalid then valid input",
			input:        "invalid\ny\n",
			defaultValue: false,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReaderSize(strings.NewReader(tt.input), 8192)
			wizard := &Wizard{
				reader: reader,
				config: &config.Config{},
			}

			result := wizard.promptYesNo("Test question", tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPromptIntWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue int
		expected     int
	}{
		{
			name:         "valid integer input",
			input:        "42\n",
			defaultValue: 10,
			expected:     42,
		},
		{
			name:         "use default on empty",
			input:        "\n",
			defaultValue: 10,
			expected:     10,
		},
		{
			name:         "invalid then valid input",
			input:        "invalid\n123\n",
			defaultValue: 10,
			expected:     123,
		},
		{
			name:         "negative number",
			input:        "-5\n",
			defaultValue: 10,
			expected:     -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReaderSize(strings.NewReader(tt.input), 8192)
			wizard := &Wizard{
				reader: reader,
				config: &config.Config{},
			}

			result := wizard.promptIntWithDefault("Test question", tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	wizard := &Wizard{
		config: &config.Config{},
	}

	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		{
			name:     "valid JWT token",
			token:    "test-header-part-12345678901234567890.test-payload-part-12345678901234567890.test-signature-part-12345678901234567890",
			expected: true,
		},
		{
			name:     "token with only 2 parts",
			token:    "header.payload",
			expected: false,
		},
		{
			name:     "token with 4 parts",
			token:    "part1.part2.part3.part4",
			expected: false,
		},
		{
			name:     "very short token",
			token:    "a.b.c",
			expected: false,
		},
		{
			name:     "empty token",
			token:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wizard.validateToken(tt.token)
			if result != tt.expected {
				t.Errorf("Expected %v for token validation, got %v", tt.expected, result)
			}
		})
	}
}

func TestExtractTokenFromPythonConfig(t *testing.T) {
	tests := []struct {
		name          string
		fileContent   string
		expectToken   bool
		expectedToken string
	}{
		{
			name: "valid python config with token",
			fileContent: `# Python config
BI_TENANT_API_TOKEN = "test-header-part-12345678901234567890.test-payload-part-12345678901234567890.test-signature-part-12345678901234567890"
OTHER_CONFIG = "value"`,
			expectToken:   true,
			expectedToken: "test-header-part-12345678901234567890.test-payload-part-12345678901234567890.test-signature-part-12345678901234567890",
		},
		{
			name: "python config without token",
			fileContent: `# Python config
OTHER_CONFIG = "value"`,
			expectToken: false,
		},
		{
			name: "python config with invalid token",
			fileContent: `# Python config
BI_TENANT_API_TOKEN = "invalid.token"`,
			expectToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.py")

			err := os.WriteFile(configPath, []byte(tt.fileContent), 0644)
			if err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			wizard := &Wizard{
				config: &config.Config{},
			}

			result := wizard.extractTokenFromPythonConfig(configPath)

			if tt.expectToken {
				if result == "" {
					t.Errorf("Expected token to be extracted, got empty string")
				}
				if result != tt.expectedToken {
					t.Errorf("Expected token '%s', got '%s'", tt.expectedToken, result)
				}
			} else {
				if result != "" {
					t.Errorf("Expected no token to be extracted, got '%s'", result)
				}
			}
		})
	}
}
