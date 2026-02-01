// Package yaml_test provides black-box tests for the YAML configuration loader.
package yaml_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/infrastructure/persistence/config/yaml"
)

// Test configuration constants for YAML loader tests.
const (
	testValidMinimalConfig string = `
version: "1"
services:
  - name: test-service
    command: /bin/echo
    args: ["hello"]
`

	testValidConfigHealthChecks string = `
version: "1"
services:
  - name: web-service
    command: /bin/server
    health_checks:
      - name: http-check
        type: http
        endpoint: http://localhost:8080/health
        interval: 10s
        timeout: 5s
        retries: 3
`

	testValidBasicConfig string = `
version: "1"
services:
  - name: my-service
    command: /bin/true
`

	testConfigForReload string = `
version: "1"
services:
  - name: test-service
    command: /bin/echo
`

	testMinimalConfigForDefaults string = `
version: "1"
services:
  - name: minimal-service
    command: /bin/true
`

	testConfigForPathCheck string = `
version: "1"
services:
  - name: test-service
    command: /bin/echo
`
)

// TestNewLoader tests the NewLoader constructor function.
//
// Params:
//   - t: testing context for assertions and error reporting
func TestNewLoader(t *testing.T) {
	// Define test cases for table-driven testing.
	tests := []struct {
		name string
	}{
		{
			name: "creates_non_nil_loader",
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new loader instance.
			loader := yaml.NewLoader()

			// Assert the loader is not nil.
			assert.NotNil(t, loader)
		})
	}
}

// TestLoader_Load tests the Load method with valid configuration.
//
// Params:
//   - t: testing context for assertions and error reporting
func TestLoader_Load(t *testing.T) {
	// Define test cases for table-driven testing.
	tests := []struct {
		name     string
		content  string
		wantErr  bool
		validate func(t *testing.T, cfg any)
	}{
		{
			name:    "valid_minimal_config",
			content: testValidMinimalConfig,
			wantErr: false,
			validate: func(t *testing.T, cfg any) {
				// Assert configuration is not nil.
				assert.NotNil(t, cfg)
			},
		},
		{
			name:    "valid_config_with_health_checks",
			content: testValidConfigHealthChecks,
			wantErr: false,
			validate: func(t *testing.T, cfg any) {
				// Assert configuration is not nil.
				assert.NotNil(t, cfg)
			},
		},
		{
			name:    "invalid_yaml_syntax",
			content: "invalid: yaml: content:",
			wantErr: true,
			validate: func(t *testing.T, cfg any) {
				// Assert configuration is nil for invalid input.
				assert.Nil(t, cfg)
			},
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test files.
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			// Write test configuration to temporary file.
			err := os.WriteFile(configPath, []byte(tt.content), 0o644)
			require.NoError(t, err)

			// Create a new loader and load the configuration.
			loader := yaml.NewLoader()
			cfg, err := loader.Load(configPath)

			// Check error expectation.
			if tt.wantErr {
				// Assert error occurred when expected.
				assert.Error(t, err)
			} else {
				// Assert no error when success expected.
				assert.NoError(t, err)
			}

			// Run custom validation.
			tt.validate(t, cfg)
		})
	}
}

// TestLoader_Load_FileNotFound tests Load with non-existent file.
//
// Params:
//   - t: testing context for assertions and error reporting
func TestLoader_Load_FileNotFound(t *testing.T) {
	// Define test cases for table-driven testing.
	tests := []struct {
		name        string
		path        string
		errContains string
	}{
		{
			name:        "non_existent_path",
			path:        "/non/existent/path/config.yaml",
			errContains: "reading config file",
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new loader instance.
			loader := yaml.NewLoader()

			// Attempt to load a non-existent file.
			cfg, err := loader.Load(tt.path)

			// Assert error occurred.
			assert.Error(t, err)
			// Assert configuration is nil.
			assert.Nil(t, cfg)
			// Assert error message contains expected text.
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

// TestLoader_Parse tests the Parse method with YAML bytes.
//
// Params:
//   - t: testing context for assertions and error reporting
func TestLoader_Parse(t *testing.T) {
	// Define test cases for table-driven testing.
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "valid_config",
			data:    []byte(testValidBasicConfig),
			wantErr: false,
		},
		{
			name:    "empty_config",
			data:    []byte(""),
			wantErr: true,
		},
		{
			name:    "invalid_yaml",
			data:    []byte("not: valid: yaml:"),
			wantErr: true,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new loader instance.
			loader := yaml.NewLoader()

			// Parse the YAML data.
			cfg, err := loader.Parse(tt.data)

			// Check error expectation.
			if tt.wantErr {
				// Assert error occurred when expected.
				assert.Error(t, err)
				// Assert configuration is nil on error.
				assert.Nil(t, cfg)
			} else {
				// Assert no error when success expected.
				assert.NoError(t, err)
				// Assert configuration is not nil on success.
				assert.NotNil(t, cfg)
			}
		})
	}
}

// TestLoader_Reload tests the Reload method.
//
// Params:
//   - t: testing context for assertions and error reporting
func TestLoader_Reload(t *testing.T) {
	// Define test cases for table-driven testing.
	tests := []struct {
		name        string
		loadFirst   bool
		wantErr     bool
		errContains string
	}{
		{
			name:        "reload_without_previous_load",
			loadFirst:   false,
			wantErr:     true,
			errContains: "no configuration loaded",
		},
		{
			name:      "reload_after_load",
			loadFirst: true,
			wantErr:   false,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new loader instance.
			loader := yaml.NewLoader()

			// Load a configuration first if required.
			if tt.loadFirst {
				// Create temporary directory for test files.
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")

				// Write valid configuration to temporary file.
				err := os.WriteFile(configPath, []byte(testConfigForReload), 0o644)
				require.NoError(t, err)

				// Load the configuration.
				_, err = loader.Load(configPath)
				require.NoError(t, err)
			}

			// Attempt to reload the configuration.
			cfg, err := loader.Reload()

			// Check error expectation.
			if tt.wantErr {
				// Assert error occurred when expected.
				assert.Error(t, err)
				// Assert configuration is nil on error.
				assert.Nil(t, cfg)
				// Assert error message contains expected text.
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				// Assert no error when success expected.
				assert.NoError(t, err)
				// Assert configuration is not nil on success.
				assert.NotNil(t, cfg)
			}
		})
	}
}

// TestLoader_DefaultsApplied tests that default values are applied.
//
// Params:
//   - t: testing context for assertions and error reporting
func TestLoader_DefaultsApplied(t *testing.T) {
	// Define test cases for table-driven testing.
	tests := []struct {
		name                    string
		content                 string
		expectedBaseDir         string
		expectedTimestampFormat string
		expectedVersion         string
	}{
		{
			name:                    "minimal_config_gets_defaults",
			content:                 testMinimalConfigForDefaults,
			expectedBaseDir:         "/var/log/daemon",
			expectedTimestampFormat: "iso8601",
			expectedVersion:         "1",
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test files.
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			// Write test configuration to temporary file.
			err := os.WriteFile(configPath, []byte(tt.content), 0o644)
			require.NoError(t, err)

			// Create a new loader and load the configuration.
			loader := yaml.NewLoader()
			cfg, err := loader.Load(configPath)

			// Assert no error occurred.
			require.NoError(t, err)
			// Assert configuration is not nil.
			require.NotNil(t, cfg)

			// Assert default logging base directory was applied.
			assert.Equal(t, tt.expectedBaseDir, cfg.Logging.BaseDir)
			// Assert default timestamp format was applied.
			assert.Equal(t, tt.expectedTimestampFormat, cfg.Logging.Defaults.TimestampFormat)
			// Assert version is set.
			assert.Equal(t, tt.expectedVersion, cfg.Version)
		})
	}
}

// TestLoader_ConfigPath tests that ConfigPath is set after Load.
//
// Params:
//   - t: testing context for assertions and error reporting
func TestLoader_ConfigPath(t *testing.T) {
	// Define test cases for table-driven testing.
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "config_path_set_after_load",
			content: testConfigForPathCheck,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test files.
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			// Write test configuration to temporary file.
			err := os.WriteFile(configPath, []byte(tt.content), 0o644)
			require.NoError(t, err)

			// Create a new loader and load the configuration.
			loader := yaml.NewLoader()
			cfg, err := loader.Load(configPath)

			// Assert no error occurred.
			require.NoError(t, err)
			// Assert ConfigPath matches the loaded path.
			assert.Equal(t, configPath, cfg.ConfigPath)
		})
	}
}
