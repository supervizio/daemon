// Package yaml provides internal white-box tests for YAML types.
// It tests internal implementation details and edge cases.
package yaml

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// marshalTextToString is a helper that converts MarshalText result to string.
//
// Params:
//   - d: the duration to marshal
//
// Returns:
//   - string: the marshaled text as string
//   - error: any marshaling error
func marshalTextToString(d *Duration) (string, error) {
	bytes, err := d.MarshalText()

	// Return empty string on error
	if err != nil {
		return "", err
	}

	// Return the string representation
	return string(bytes), nil
}

// TestDuration_EdgeCases tests Duration with various edge case values.
// It verifies that zero, negative, and large Duration values marshal correctly.
//
// Params:
//   - t: testing context
func TestDuration_EdgeCases(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name     string
		duration Duration
		expected string
	}{
		{
			name:     "zero value",
			duration: Duration(0),
			expected: "0s",
		},
		{
			name:     "negative value",
			duration: Duration(-5 * time.Second),
			expected: "-5s",
		},
		{
			name:     "large value (24 hours)",
			duration: Duration(24 * time.Hour),
			expected: "24h0m0s",
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			resultStr, err := marshalTextToString(&testCase.duration)

			// Verify no error occurs.
			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, resultStr)
		})
	}
}

// TestConfigDTO_EdgeCases tests ConfigDTO with edge case configurations.
// It verifies that empty service lists are handled correctly.
//
// Params:
//   - t: testing context
func TestConfigDTO_EdgeCases(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name            string
		dto             *ConfigDTO
		configPath      string
		expectedVersion string
		expectEmpty     bool
	}{
		{
			name: "empty services list",
			dto: &ConfigDTO{
				Version:  "1.0",
				Services: nil,
			},
			configPath:      "/config.yaml",
			expectedVersion: "1.0",
			expectEmpty:     true,
		},
		{
			name: "empty services slice",
			dto: &ConfigDTO{
				Version:  "2.0",
				Services: nil,
			},
			configPath:      "/etc/app.yaml",
			expectedVersion: "2.0",
			expectEmpty:     true,
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.dto.ToDomain(testCase.configPath)

			// Verify services list state.
			if testCase.expectEmpty {
				assert.Empty(t, result.Services)
			}
			assert.Equal(t, testCase.expectedVersion, result.Version)
		})
	}
}

// TestServiceConfigDTO_EdgeCases tests ServiceConfigDTO with edge case configurations.
// It verifies that empty health checks and nil environment are handled correctly.
//
// Params:
//   - t: testing context
func TestServiceConfigDTO_EdgeCases(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name                    string
		dto                     *ServiceConfigDTO
		expectEmptyHealthChecks bool
		expectNilEnvironment    bool
	}{
		{
			name: "empty health checks",
			dto: &ServiceConfigDTO{
				Name:         "test",
				Command:      "/bin/test",
				HealthChecks: nil,
			},
			expectEmptyHealthChecks: true,
			expectNilEnvironment:    true,
		},
		{
			name: "nil environment",
			dto: &ServiceConfigDTO{
				Name:        "test",
				Command:     "/bin/test",
				Environment: nil,
			},
			expectEmptyHealthChecks: true,
			expectNilEnvironment:    true,
		},
		{
			name: "empty health checks slice",
			dto: &ServiceConfigDTO{
				Name:         "test",
				Command:      "/bin/test",
				HealthChecks: nil,
			},
			expectEmptyHealthChecks: true,
			expectNilEnvironment:    true,
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.dto.ToDomain()

			// Verify health checks list state.
			if testCase.expectEmptyHealthChecks {
				assert.Empty(t, result.HealthChecks)
			}
			// Verify environment map state.
			if testCase.expectNilEnvironment {
				assert.Nil(t, result.Environment)
			}
		})
	}
}

// TestRestartConfigDTO_EdgeCases tests RestartConfigDTO with edge case configurations.
// It verifies that zero delay values are handled correctly.
//
// Params:
//   - t: testing context
func TestRestartConfigDTO_EdgeCases(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name             string
		dto              *RestartConfigDTO
		expectedDelay    time.Duration
		expectedDelayMax time.Duration
	}{
		{
			name: "zero durations",
			dto: &RestartConfigDTO{
				Policy:     "always",
				MaxRetries: 0,
				Delay:      Duration(0),
				DelayMax:   Duration(0),
			},
			expectedDelay:    time.Duration(0),
			expectedDelayMax: time.Duration(0),
		},
		{
			name: "zero with different policy",
			dto: &RestartConfigDTO{
				Policy:     "on-failure",
				MaxRetries: 3,
				Delay:      Duration(0),
				DelayMax:   Duration(0),
			},
			expectedDelay:    time.Duration(0),
			expectedDelayMax: time.Duration(0),
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.dto.ToDomain()

			// Verify zero durations are preserved.
			assert.Equal(t, testCase.expectedDelay, result.Delay.Duration())
			assert.Equal(t, testCase.expectedDelayMax, result.DelayMax.Duration())
		})
	}
}

// TestHealthCheckDTO_EdgeCases tests HealthCheckDTO with minimal configuration.
// It verifies that minimal health check config is handled correctly.
//
// Params:
//   - t: testing context
func TestHealthCheckDTO_EdgeCases(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name                string
		dto                 *HealthCheckDTO
		expectedType        string
		expectEmptyName     bool
		expectEmptyEndpoint bool
	}{
		{
			name: "minimal command configuration",
			dto: &HealthCheckDTO{
				Type: "command",
			},
			expectedType:        "command",
			expectEmptyName:     true,
			expectEmptyEndpoint: true,
		},
		{
			name: "minimal tcp configuration",
			dto: &HealthCheckDTO{
				Type: "tcp",
			},
			expectedType:        "tcp",
			expectEmptyName:     true,
			expectEmptyEndpoint: true,
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.dto.ToDomain()

			// Use String() method instead of type conversion.
			typeStr := result.Type.String()

			// Verify minimal configuration is preserved.
			assert.Equal(t, testCase.expectedType, typeStr)
			if testCase.expectEmptyName {
				assert.Empty(t, result.Name)
			}
			if testCase.expectEmptyEndpoint {
				assert.Empty(t, result.Endpoint)
			}
		})
	}
}

// TestRotationConfigDTO_EdgeCases tests RotationConfigDTO with default values.
// It verifies that default rotation values are handled correctly.
//
// Params:
//   - t: testing context
func TestRotationConfigDTO_EdgeCases(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name             string
		dto              *RotationConfigDTO
		expectEmptySize  bool
		expectEmptyAge   bool
		expectedMaxFiles int
		expectedCompress bool
	}{
		{
			name:             "default values (empty struct)",
			dto:              &RotationConfigDTO{},
			expectEmptySize:  true,
			expectEmptyAge:   true,
			expectedMaxFiles: 0,
			expectedCompress: false,
		},
		{
			name: "zero files with compression false",
			dto: &RotationConfigDTO{
				MaxSize:  "",
				MaxAge:   "",
				MaxFiles: 0,
				Compress: false,
			},
			expectEmptySize:  true,
			expectEmptyAge:   true,
			expectedMaxFiles: 0,
			expectedCompress: false,
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.dto.ToDomain()

			// Verify default values are preserved.
			if testCase.expectEmptySize {
				assert.Empty(t, result.MaxSize)
			}
			if testCase.expectEmptyAge {
				assert.Empty(t, result.MaxAge)
			}
			assert.Equal(t, testCase.expectedMaxFiles, result.MaxFiles)
			assert.Equal(t, testCase.expectedCompress, result.Compress)
		})
	}
}

// TestLogStreamConfigDTO_EdgeCases tests LogStreamConfigDTO with empty file.
// It verifies that empty file path is handled correctly.
//
// Params:
//   - t: testing context
func TestLogStreamConfigDTO_EdgeCases(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name              string
		dto               *LogStreamConfigDTO
		expectEmptyPath   bool
		expectEmptyFormat bool
	}{
		{
			name: "empty file path",
			dto: &LogStreamConfigDTO{
				File: "",
			},
			expectEmptyPath:   true,
			expectEmptyFormat: true,
		},
		{
			name:              "empty struct",
			dto:               &LogStreamConfigDTO{},
			expectEmptyPath:   true,
			expectEmptyFormat: true,
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.dto.ToDomain()

			// Verify empty file path is preserved.
			if testCase.expectEmptyPath {
				assert.Empty(t, result.FilePath)
			}
			if testCase.expectEmptyFormat {
				assert.Empty(t, result.Format)
			}
		})
	}
}

// TestServiceLoggingDTO_EdgeCases tests ServiceLoggingDTO with empty streams.
// It verifies that empty log stream configs are handled correctly.
//
// Params:
//   - t: testing context
func TestServiceLoggingDTO_EdgeCases(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name              string
		dto               *ServiceLoggingDTO
		expectEmptyStdout bool
		expectEmptyStderr bool
	}{
		{
			name:              "empty struct",
			dto:               &ServiceLoggingDTO{},
			expectEmptyStdout: true,
			expectEmptyStderr: true,
		},
		{
			name: "empty stream configs",
			dto: &ServiceLoggingDTO{
				Stdout: LogStreamConfigDTO{},
				Stderr: LogStreamConfigDTO{},
			},
			expectEmptyStdout: true,
			expectEmptyStderr: true,
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.dto.ToDomain()

			// Verify empty streams are preserved.
			if testCase.expectEmptyStdout {
				assert.Empty(t, result.Stdout.FilePath)
			}
			if testCase.expectEmptyStderr {
				assert.Empty(t, result.Stderr.FilePath)
			}
		})
	}
}

// TestLoggingConfigDTO_EdgeCases tests LoggingConfigDTO with empty base dir.
// It verifies that empty base directory is handled correctly.
//
// Params:
//   - t: testing context
func TestLoggingConfigDTO_EdgeCases(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name            string
		dto             *LoggingConfigDTO
		expectEmptyBase bool
	}{
		{
			name: "empty base directory",
			dto: &LoggingConfigDTO{
				BaseDir: "",
			},
			expectEmptyBase: true,
		},
		{
			name:            "empty struct",
			dto:             &LoggingConfigDTO{},
			expectEmptyBase: true,
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.dto.ToDomain()

			// Verify empty base directory is preserved.
			if testCase.expectEmptyBase {
				assert.Empty(t, result.BaseDir)
			}
		})
	}
}

// TestLogDefaultsDTO_EdgeCases tests LogDefaultsDTO with empty format.
// It verifies that empty timestamp format is handled correctly.
//
// Params:
//   - t: testing context
func TestLogDefaultsDTO_EdgeCases(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name              string
		dto               *LogDefaultsDTO
		expectEmptyFormat bool
	}{
		{
			name: "empty timestamp format",
			dto: &LogDefaultsDTO{
				TimestampFormat: "",
			},
			expectEmptyFormat: true,
		},
		{
			name:              "empty struct",
			dto:               &LogDefaultsDTO{},
			expectEmptyFormat: true,
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.dto.ToDomain()

			// Verify empty format is preserved.
			if testCase.expectEmptyFormat {
				assert.Empty(t, result.TimestampFormat)
			}
		})
	}
}
