// Package yaml_test provides black-box tests for the yaml package.
// It tests the public API of YAML configuration types and their domain conversions.
package yaml_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yamlv3 "gopkg.in/yaml.v3"

	"github.com/kodflow/daemon/internal/infrastructure/config/yaml"
)

// MarshalTexter defines the interface for types that can marshal to text.
type MarshalTexter interface {
	MarshalText() ([]byte, error)
}

// marshalDurationText is a helper that converts MarshalText result to string.
//
// Params:
//   - d: the text marshaler to marshal
//
// Returns:
//   - string: the marshaled text as string
//   - error: any marshaling error
func marshalDurationText(d MarshalTexter) (string, error) {
	bytes, err := d.MarshalText()

	// Return empty string on error
	if err != nil {
		return "", err
	}

	// Return the string representation
	return string(bytes), nil
}

// TestDuration_UnmarshalYAML tests Duration unmarshaling from YAML.
// It verifies that duration strings are correctly parsed.
//
// Params:
//   - t: testing context
func TestDuration_UnmarshalYAML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{
			name:     "valid seconds",
			input:    "30s",
			expected: 30 * time.Second,
			wantErr:  false,
		},
		{
			name:     "valid minutes",
			input:    "5m",
			expected: 5 * time.Minute,
			wantErr:  false,
		},
		{
			name:     "valid hours",
			input:    "2h",
			expected: 2 * time.Hour,
			wantErr:  false,
		},
		{
			name:     "valid complex duration",
			input:    "1h30m",
			expected: 90 * time.Minute,
			wantErr:  false,
		},
		{
			name:     "invalid duration",
			input:    "invalid",
			expected: 0,
			wantErr:  true,
		},
	}

	// Iterate through test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var d yaml.Duration
			yamlInput := tt.input

			err := yamlv3.Unmarshal([]byte(yamlInput), &d)

			// Check error expectation
			if tt.wantErr {
				// Expect an error for invalid input
				assert.Error(t, err)

				// Return early on expected error
				return
			}

			// Expect no error for valid input
			require.NoError(t, err)
			assert.Equal(t, tt.expected, time.Duration(d))
		})
	}
}

// TestDuration_UnmarshalYAML_NonStringValue tests Duration unmarshaling when
// the YAML value is not a string (e.g., array, map, or number).
// This tests the error path when unmarshal(&s) fails.
//
// Params:
//   - t: testing context
func TestDuration_UnmarshalYAML_NonStringValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "array value fails unmarshal",
			input: "[1, 2, 3]",
		},
		{
			name:  "map value fails unmarshal",
			input: "key: value",
		},
		{
			name:  "nested object fails unmarshal",
			input: "outer:\n  inner: value",
		},
	}

	// Iterate through test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var d yaml.Duration

			err := yamlv3.Unmarshal([]byte(tt.input), &d)

			// Expect an error when value is not a string
			assert.Error(t, err)
		})
	}
}

// TestDuration_MarshalText tests Duration marshaling via TextMarshaler.
// It verifies that Duration values are correctly serialized to strings.
//
// Params:
//   - t: testing context
func TestDuration_MarshalText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration yaml.Duration
		expected string
	}{
		{
			name:     "zero duration",
			duration: yaml.Duration(0),
			expected: "0s",
		},
		{
			name:     "seconds",
			duration: yaml.Duration(30 * time.Second),
			expected: "30s",
		},
		{
			name:     "minutes",
			duration: yaml.Duration(5 * time.Minute),
			expected: "5m0s",
		},
		{
			name:     "hours",
			duration: yaml.Duration(2 * time.Hour),
			expected: "2h0m0s",
		},
	}

	// Iterate through test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resultStr, err := marshalDurationText(&tt.duration)

			// Expect no error for marshaling
			require.NoError(t, err)
			assert.Equal(t, tt.expected, resultStr)
		})
	}
}

// TestConfigDTO_ToDomain tests ConfigDTO to domain conversion.
// It verifies that all fields are correctly mapped to the domain model.
//
// Params:
//   - t: testing context
func TestConfigDTO_ToDomain(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name                string
		dto                 *yaml.ConfigDTO
		configPath          string
		expectedVersion     string
		expectedBaseDir     string
		expectedServiceLen  int
		expectedServiceName string
	}{
		{
			name: "full configuration converts correctly",
			dto: &yaml.ConfigDTO{
				Version: "1.0",
				Logging: yaml.LoggingConfigDTO{
					BaseDir: "/var/log",
					Defaults: yaml.LogDefaultsDTO{
						TimestampFormat: "2006-01-02",
						Rotation: yaml.RotationConfigDTO{
							MaxSize:  "100MB",
							MaxAge:   "7d",
							MaxFiles: 5,
							Compress: true,
						},
					},
				},
				Services: []yaml.ServiceConfigDTO{
					{
						Name:    "test-service",
						Command: "/bin/test",
					},
				},
			},
			configPath:          "/path/to/config.yaml",
			expectedVersion:     "1.0",
			expectedBaseDir:     "/var/log",
			expectedServiceLen:  1,
			expectedServiceName: "test-service",
		},
		{
			name: "different version and path",
			dto: &yaml.ConfigDTO{
				Version: "2.0",
				Logging: yaml.LoggingConfigDTO{
					BaseDir: "/opt/logs",
				},
				Services: []yaml.ServiceConfigDTO{
					{
						Name:    "another-service",
						Command: "/usr/bin/app",
					},
				},
			},
			configPath:          "/etc/daemon/config.yaml",
			expectedVersion:     "2.0",
			expectedBaseDir:     "/opt/logs",
			expectedServiceLen:  1,
			expectedServiceName: "another-service",
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.dto.ToDomain(testCase.configPath)

			// Verify version is correctly set.
			assert.Equal(t, testCase.expectedVersion, result.Version)
			assert.Equal(t, testCase.configPath, result.ConfigPath)
			assert.Equal(t, testCase.expectedBaseDir, result.Logging.BaseDir)
			assert.Len(t, result.Services, testCase.expectedServiceLen)
			if testCase.expectedServiceLen > 0 {
				assert.Equal(t, testCase.expectedServiceName, result.Services[0].Name)
			}
		})
	}
}

// TestServiceConfigDTO_ToDomain tests ServiceConfigDTO to domain conversion.
// It verifies that service configuration fields are correctly mapped.
//
// Params:
//   - t: testing context
func TestServiceConfigDTO_ToDomain(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name                     string
		dto                      *yaml.ServiceConfigDTO
		expectedName             string
		expectedCommand          string
		expectedArgs             []string
		expectedUser             string
		expectedGroup            string
		expectedWorkingDirectory string
		expectedEnvironment      map[string]string
		expectedDependsOn        []string
		expectedOneshot          bool
		expectedHealthCheckLen   int
	}{
		{
			name: "full service configuration converts correctly",
			dto: &yaml.ServiceConfigDTO{
				Name:             "my-service",
				Command:          "/usr/bin/app",
				Args:             []string{"--config", "/etc/app.conf"},
				User:             "appuser",
				Group:            "appgroup",
				WorkingDirectory: "/var/app",
				Environment:      map[string]string{"ENV": "production"},
				DependsOn:        []string{"database"},
				Oneshot:          true,
				Restart: yaml.RestartConfigDTO{
					Policy:     "always",
					MaxRetries: 3,
				},
				HealthChecks: []yaml.HealthCheckDTO{
					{
						Name: "http-check",
						Type: "http",
					},
				},
			},
			expectedName:             "my-service",
			expectedCommand:          "/usr/bin/app",
			expectedArgs:             []string{"--config", "/etc/app.conf"},
			expectedUser:             "appuser",
			expectedGroup:            "appgroup",
			expectedWorkingDirectory: "/var/app",
			expectedEnvironment:      map[string]string{"ENV": "production"},
			expectedDependsOn:        []string{"database"},
			expectedOneshot:          true,
			expectedHealthCheckLen:   1,
		},
		{
			name: "minimal service configuration",
			dto: &yaml.ServiceConfigDTO{
				Name:    "simple-service",
				Command: "/bin/simple",
			},
			expectedName:             "simple-service",
			expectedCommand:          "/bin/simple",
			expectedArgs:             nil,
			expectedUser:             "",
			expectedGroup:            "",
			expectedWorkingDirectory: "",
			expectedEnvironment:      nil,
			expectedDependsOn:        nil,
			expectedOneshot:          false,
			expectedHealthCheckLen:   0,
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.dto.ToDomain()

			// Verify all service fields are correctly mapped.
			assert.Equal(t, testCase.expectedName, result.Name)
			assert.Equal(t, testCase.expectedCommand, result.Command)
			assert.Equal(t, testCase.expectedArgs, result.Args)
			assert.Equal(t, testCase.expectedUser, result.User)
			assert.Equal(t, testCase.expectedGroup, result.Group)
			assert.Equal(t, testCase.expectedWorkingDirectory, result.WorkingDirectory)
			assert.Equal(t, testCase.expectedEnvironment, result.Environment)
			assert.Equal(t, testCase.expectedDependsOn, result.DependsOn)
			assert.Equal(t, testCase.expectedOneshot, result.Oneshot)
			assert.Len(t, result.HealthChecks, testCase.expectedHealthCheckLen)
		})
	}
}

// TestRestartConfigDTO_ToDomain tests RestartConfigDTO to domain conversion.
// It verifies that restart policy settings are correctly mapped.
//
// Params:
//   - t: testing context
func TestRestartConfigDTO_ToDomain(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name               string
		dto                *yaml.RestartConfigDTO
		expectedPolicy     string
		expectedMaxRetries int
		expectedDelay      time.Duration
		expectedDelayMax   time.Duration
	}{
		{
			name: "on-failure policy with delays",
			dto: &yaml.RestartConfigDTO{
				Policy:     "on-failure",
				MaxRetries: 5,
				Delay:      yaml.Duration(10 * time.Second),
				DelayMax:   yaml.Duration(60 * time.Second),
			},
			expectedPolicy:     "on-failure",
			expectedMaxRetries: 5,
			expectedDelay:      10 * time.Second,
			expectedDelayMax:   60 * time.Second,
		},
		{
			name: "always policy with different values",
			dto: &yaml.RestartConfigDTO{
				Policy:     "always",
				MaxRetries: 0,
				Delay:      yaml.Duration(5 * time.Second),
				DelayMax:   yaml.Duration(30 * time.Second),
			},
			expectedPolicy:     "always",
			expectedMaxRetries: 0,
			expectedDelay:      5 * time.Second,
			expectedDelayMax:   30 * time.Second,
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.dto.ToDomain()

			// Use String() method instead of type conversion.
			policyStr := result.Policy.String()

			// Verify restart configuration fields are correctly mapped.
			assert.Equal(t, testCase.expectedPolicy, policyStr)
			assert.Equal(t, testCase.expectedMaxRetries, result.MaxRetries)
			assert.Equal(t, testCase.expectedDelay, result.Delay.Duration())
			assert.Equal(t, testCase.expectedDelayMax, result.DelayMax.Duration())
		})
	}
}

// TestHealthCheckDTO_ToDomain tests HealthCheckDTO to domain conversion.
// It verifies that health check parameters are correctly mapped.
//
// Params:
//   - t: testing context
func TestHealthCheckDTO_ToDomain(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name               string
		dto                *yaml.HealthCheckDTO
		expectedName       string
		expectedType       string
		expectedInterval   time.Duration
		expectedTimeout    time.Duration
		expectedRetries    int
		expectedEndpoint   string
		expectedMethod     string
		expectedStatusCode int
	}{
		{
			name: "http health check converts correctly",
			dto: &yaml.HealthCheckDTO{
				Name:       "api-health",
				Type:       "http",
				Interval:   yaml.Duration(30 * time.Second),
				Timeout:    yaml.Duration(5 * time.Second),
				Retries:    3,
				Endpoint:   "http://localhost:8080/health",
				Method:     "GET",
				StatusCode: 200,
				Host:       "localhost",
				Port:       8080,
				Command:    "",
			},
			expectedName:       "api-health",
			expectedType:       "http",
			expectedInterval:   30 * time.Second,
			expectedTimeout:    5 * time.Second,
			expectedRetries:    3,
			expectedEndpoint:   "http://localhost:8080/health",
			expectedMethod:     "GET",
			expectedStatusCode: 200,
		},
		{
			name: "tcp health check converts correctly",
			dto: &yaml.HealthCheckDTO{
				Name:     "tcp-health",
				Type:     "tcp",
				Interval: yaml.Duration(15 * time.Second),
				Timeout:  yaml.Duration(3 * time.Second),
				Host:     "127.0.0.1",
				Port:     3306,
			},
			expectedName:       "tcp-health",
			expectedType:       "tcp",
			expectedInterval:   15 * time.Second,
			expectedTimeout:    3 * time.Second,
			expectedRetries:    0,
			expectedEndpoint:   "",
			expectedMethod:     "",
			expectedStatusCode: 0,
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.dto.ToDomain()

			// Use String() method instead of type conversion.
			typeStr := result.Type.String()

			// Verify health check fields are correctly mapped.
			assert.Equal(t, testCase.expectedName, result.Name)
			assert.Equal(t, testCase.expectedType, typeStr)
			assert.Equal(t, testCase.expectedInterval, result.Interval.Duration())
			assert.Equal(t, testCase.expectedTimeout, result.Timeout.Duration())
			assert.Equal(t, testCase.expectedRetries, result.Retries)
			assert.Equal(t, testCase.expectedEndpoint, result.Endpoint)
			assert.Equal(t, testCase.expectedMethod, result.Method)
			assert.Equal(t, testCase.expectedStatusCode, result.StatusCode)
		})
	}
}

// TestLoggingConfigDTO_ToDomain tests LoggingConfigDTO to domain conversion.
// It verifies that logging configuration fields are correctly mapped.
//
// Params:
//   - t: testing context
func TestLoggingConfigDTO_ToDomain(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name                    string
		dto                     *yaml.LoggingConfigDTO
		expectedBaseDir         string
		expectedTimestampFormat string
		expectedMaxSize         string
		expectedMaxFiles        int
		expectedCompress        bool
	}{
		{
			name: "full logging configuration converts correctly",
			dto: &yaml.LoggingConfigDTO{
				BaseDir: "/var/log/app",
				Defaults: yaml.LogDefaultsDTO{
					TimestampFormat: "2006-01-02T15:04:05",
					Rotation: yaml.RotationConfigDTO{
						MaxSize:  "50MB",
						MaxAge:   "30d",
						MaxFiles: 10,
						Compress: true,
					},
				},
			},
			expectedBaseDir:         "/var/log/app",
			expectedTimestampFormat: "2006-01-02T15:04:05",
			expectedMaxSize:         "50MB",
			expectedMaxFiles:        10,
			expectedCompress:        true,
		},
		{
			name: "different logging configuration",
			dto: &yaml.LoggingConfigDTO{
				BaseDir: "/opt/logs",
				Defaults: yaml.LogDefaultsDTO{
					TimestampFormat: "RFC3339",
					Rotation: yaml.RotationConfigDTO{
						MaxSize:  "100MB",
						MaxAge:   "7d",
						MaxFiles: 5,
						Compress: false,
					},
				},
			},
			expectedBaseDir:         "/opt/logs",
			expectedTimestampFormat: "RFC3339",
			expectedMaxSize:         "100MB",
			expectedMaxFiles:        5,
			expectedCompress:        false,
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.dto.ToDomain()

			// Verify logging configuration fields are correctly mapped.
			assert.Equal(t, testCase.expectedBaseDir, result.BaseDir)
			assert.Equal(t, testCase.expectedTimestampFormat, result.Defaults.TimestampFormat)
			assert.Equal(t, testCase.expectedMaxSize, result.Defaults.Rotation.MaxSize)
			assert.Equal(t, testCase.expectedMaxFiles, result.Defaults.Rotation.MaxFiles)
			assert.Equal(t, testCase.expectedCompress, result.Defaults.Rotation.Compress)
		})
	}
}

// TestRotationConfigDTO_ToDomain tests RotationConfigDTO to domain conversion.
// It verifies that rotation parameters are correctly mapped.
//
// Params:
//   - t: testing context
func TestRotationConfigDTO_ToDomain(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name             string
		dto              *yaml.RotationConfigDTO
		expectedMaxSize  string
		expectedMaxAge   string
		expectedMaxFiles int
		expectedCompress bool
	}{
		{
			name: "rotation with compression enabled",
			dto: &yaml.RotationConfigDTO{
				MaxSize:  "100MB",
				MaxAge:   "7d",
				MaxFiles: 5,
				Compress: true,
			},
			expectedMaxSize:  "100MB",
			expectedMaxAge:   "7d",
			expectedMaxFiles: 5,
			expectedCompress: true,
		},
		{
			name: "rotation with compression disabled",
			dto: &yaml.RotationConfigDTO{
				MaxSize:  "50MB",
				MaxAge:   "30d",
				MaxFiles: 10,
				Compress: false,
			},
			expectedMaxSize:  "50MB",
			expectedMaxAge:   "30d",
			expectedMaxFiles: 10,
			expectedCompress: false,
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.dto.ToDomain()

			// Verify rotation configuration fields are correctly mapped.
			assert.Equal(t, testCase.expectedMaxSize, result.MaxSize)
			assert.Equal(t, testCase.expectedMaxAge, result.MaxAge)
			assert.Equal(t, testCase.expectedMaxFiles, result.MaxFiles)
			assert.Equal(t, testCase.expectedCompress, result.Compress)
		})
	}
}

// TestServiceLoggingDTO_ToDomain tests ServiceLoggingDTO to domain conversion.
// It verifies that service logging streams are correctly mapped.
//
// Params:
//   - t: testing context
func TestServiceLoggingDTO_ToDomain(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name               string
		dto                *yaml.ServiceLoggingDTO
		expectedStdoutFile string
		expectedStderrFile string
	}{
		{
			name: "both streams configured",
			dto: &yaml.ServiceLoggingDTO{
				Stdout: yaml.LogStreamConfigDTO{
					File:            "stdout.log",
					TimestampFormat: "2006-01-02",
				},
				Stderr: yaml.LogStreamConfigDTO{
					File:            "stderr.log",
					TimestampFormat: "2006-01-02",
				},
			},
			expectedStdoutFile: "stdout.log",
			expectedStderrFile: "stderr.log",
		},
		{
			name: "different file names",
			dto: &yaml.ServiceLoggingDTO{
				Stdout: yaml.LogStreamConfigDTO{
					File:            "app-out.log",
					TimestampFormat: "RFC3339",
				},
				Stderr: yaml.LogStreamConfigDTO{
					File:            "app-err.log",
					TimestampFormat: "RFC3339",
				},
			},
			expectedStdoutFile: "app-out.log",
			expectedStderrFile: "app-err.log",
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.dto.ToDomain()

			// Verify service logging streams are correctly mapped.
			assert.Equal(t, testCase.expectedStdoutFile, result.Stdout.FilePath)
			assert.Equal(t, testCase.expectedStderrFile, result.Stderr.FilePath)
		})
	}
}

// TestLogStreamConfigDTO_ToDomain tests LogStreamConfigDTO to domain conversion.
// It verifies that log stream configuration fields are correctly mapped.
//
// Params:
//   - t: testing context
func TestLogStreamConfigDTO_ToDomain(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name                     string
		dto                      *yaml.LogStreamConfigDTO
		expectedFilePath         string
		expectedFormat           string
		expectedRotationMaxSize  string
		expectedRotationMaxFiles int
		expectedRotationCompress bool
	}{
		{
			name: "full stream configuration without compression",
			dto: &yaml.LogStreamConfigDTO{
				File:            "app.log",
				TimestampFormat: "2006-01-02T15:04:05",
				Rotation: yaml.RotationConfigDTO{
					MaxSize:  "25MB",
					MaxAge:   "14d",
					MaxFiles: 7,
					Compress: false,
				},
			},
			expectedFilePath:         "app.log",
			expectedFormat:           "2006-01-02T15:04:05",
			expectedRotationMaxSize:  "25MB",
			expectedRotationMaxFiles: 7,
			expectedRotationCompress: false,
		},
		{
			name: "stream configuration with compression",
			dto: &yaml.LogStreamConfigDTO{
				File:            "service.log",
				TimestampFormat: "RFC3339",
				Rotation: yaml.RotationConfigDTO{
					MaxSize:  "50MB",
					MaxAge:   "7d",
					MaxFiles: 10,
					Compress: true,
				},
			},
			expectedFilePath:         "service.log",
			expectedFormat:           "RFC3339",
			expectedRotationMaxSize:  "50MB",
			expectedRotationMaxFiles: 10,
			expectedRotationCompress: true,
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.dto.ToDomain()

			// Verify log stream configuration fields are correctly mapped.
			assert.Equal(t, testCase.expectedFilePath, result.FilePath)
			assert.Equal(t, testCase.expectedFormat, result.Format)
			assert.Equal(t, testCase.expectedRotationMaxSize, result.RotationConfig.MaxSize)
			assert.Equal(t, testCase.expectedRotationMaxFiles, result.RotationConfig.MaxFiles)
			assert.Equal(t, testCase.expectedRotationCompress, result.RotationConfig.Compress)
		})
	}
}

// TestLogDefaultsDTO_ToDomain tests LogDefaultsDTO to domain conversion.
// It verifies that log defaults are correctly mapped.
//
// Params:
//   - t: testing context
func TestLogDefaultsDTO_ToDomain(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name                     string
		dto                      *yaml.LogDefaultsDTO
		expectedTimestampFormat  string
		expectedRotationMaxSize  string
		expectedRotationMaxAge   string
		expectedRotationFiles    int
		expectedRotationCompress bool
	}{
		{
			name: "defaults with compression enabled",
			dto: &yaml.LogDefaultsDTO{
				TimestampFormat: "RFC3339",
				Rotation: yaml.RotationConfigDTO{
					MaxSize:  "10MB",
					MaxAge:   "3d",
					MaxFiles: 3,
					Compress: true,
				},
			},
			expectedTimestampFormat:  "RFC3339",
			expectedRotationMaxSize:  "10MB",
			expectedRotationMaxAge:   "3d",
			expectedRotationFiles:    3,
			expectedRotationCompress: true,
		},
		{
			name: "defaults with different format",
			dto: &yaml.LogDefaultsDTO{
				TimestampFormat: "2006-01-02",
				Rotation: yaml.RotationConfigDTO{
					MaxSize:  "50MB",
					MaxAge:   "7d",
					MaxFiles: 5,
					Compress: false,
				},
			},
			expectedTimestampFormat:  "2006-01-02",
			expectedRotationMaxSize:  "50MB",
			expectedRotationMaxAge:   "7d",
			expectedRotationFiles:    5,
			expectedRotationCompress: false,
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.dto.ToDomain()

			// Verify log defaults are correctly mapped.
			assert.Equal(t, testCase.expectedTimestampFormat, result.TimestampFormat)
			assert.Equal(t, testCase.expectedRotationMaxSize, result.Rotation.MaxSize)
			assert.Equal(t, testCase.expectedRotationMaxAge, result.Rotation.MaxAge)
			assert.Equal(t, testCase.expectedRotationFiles, result.Rotation.MaxFiles)
			assert.Equal(t, testCase.expectedRotationCompress, result.Rotation.Compress)
		})
	}
}

// TestListenerDTO_ToDomain tests ListenerDTO to domain conversion.
// It verifies that listener configuration fields are correctly mapped.
//
// Params:
//   - t: testing context
func TestListenerDTO_ToDomain(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name              string
		dto               *yaml.ListenerDTO
		expectedName      string
		expectedPort      int
		expectedProtocol  string
		expectedAddress   string
		hasProbe          bool
		expectedProbeType string
	}{
		{
			name: "listener with tcp protocol and probe",
			dto: &yaml.ListenerDTO{
				Name:     "http-listener",
				Port:     8080,
				Protocol: "tcp",
				Address:  "0.0.0.0",
				Probe: yaml.ProbeDTO{
					Type:             "http",
					Interval:         yaml.Duration(30 * time.Second),
					Timeout:          yaml.Duration(5 * time.Second),
					SuccessThreshold: 1,
					FailureThreshold: 3,
					Path:             "/health",
					Method:           "GET",
					StatusCode:       200,
				},
			},
			expectedName:      "http-listener",
			expectedPort:      8080,
			expectedProtocol:  "tcp",
			expectedAddress:   "0.0.0.0",
			hasProbe:          true,
			expectedProbeType: "http",
		},
		{
			name: "listener with default protocol",
			dto: &yaml.ListenerDTO{
				Name:    "grpc-listener",
				Port:    9090,
				Address: "127.0.0.1",
			},
			expectedName:     "grpc-listener",
			expectedPort:     9090,
			expectedProtocol: "tcp",
			expectedAddress:  "127.0.0.1",
			hasProbe:         false,
		},
		{
			name: "listener with udp protocol",
			dto: &yaml.ListenerDTO{
				Name:     "dns-listener",
				Port:     53,
				Protocol: "udp",
				Address:  "0.0.0.0",
			},
			expectedName:     "dns-listener",
			expectedPort:     53,
			expectedProtocol: "udp",
			expectedAddress:  "0.0.0.0",
			hasProbe:         false,
		},
		{
			name: "listener without probe",
			dto: &yaml.ListenerDTO{
				Name:     "admin-listener",
				Port:     9000,
				Protocol: "tcp",
			},
			expectedName:     "admin-listener",
			expectedPort:     9000,
			expectedProtocol: "tcp",
			expectedAddress:  "",
			hasProbe:         false,
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.dto.ToDomain()

			// Verify listener fields are correctly mapped.
			assert.Equal(t, testCase.expectedName, result.Name)
			assert.Equal(t, testCase.expectedPort, result.Port)
			assert.Equal(t, testCase.expectedProtocol, result.Protocol)
			assert.Equal(t, testCase.expectedAddress, result.Address)

			// Check probe presence.
			if testCase.hasProbe {
				assert.NotNil(t, result.Probe)
				assert.Equal(t, testCase.expectedProbeType, result.Probe.Type)
			} else {
				assert.Nil(t, result.Probe)
			}
		})
	}
}

// TestProbeDTO_ToDomain tests ProbeDTO to domain conversion.
// It verifies that probe configuration fields are correctly mapped with defaults.
//
// Params:
//   - t: testing context
func TestProbeDTO_ToDomain(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name                     string
		dto                      *yaml.ProbeDTO
		expectedType             string
		expectedInterval         time.Duration
		expectedTimeout          time.Duration
		expectedSuccessThreshold int
		expectedFailureThreshold int
		expectedPath             string
		expectedMethod           string
		expectedStatusCode       int
		expectedService          string
		expectedCommand          string
	}{
		{
			name: "http probe with all fields",
			dto: &yaml.ProbeDTO{
				Type:             "http",
				Interval:         yaml.Duration(30 * time.Second),
				Timeout:          yaml.Duration(5 * time.Second),
				SuccessThreshold: 2,
				FailureThreshold: 5,
				Path:             "/healthz",
				Method:           "POST",
				StatusCode:       201,
			},
			expectedType:             "http",
			expectedInterval:         30 * time.Second,
			expectedTimeout:          5 * time.Second,
			expectedSuccessThreshold: 2,
			expectedFailureThreshold: 5,
			expectedPath:             "/healthz",
			expectedMethod:           "POST",
			expectedStatusCode:       201,
		},
		{
			name: "http probe with defaults",
			dto: &yaml.ProbeDTO{
				Type:     "http",
				Interval: yaml.Duration(10 * time.Second),
				Timeout:  yaml.Duration(2 * time.Second),
				Path:     "/health",
			},
			expectedType:             "http",
			expectedInterval:         10 * time.Second,
			expectedTimeout:          2 * time.Second,
			expectedSuccessThreshold: 1,
			expectedFailureThreshold: 3,
			expectedPath:             "/health",
			expectedMethod:           "GET",
			expectedStatusCode:       200,
		},
		{
			name: "tcp probe",
			dto: &yaml.ProbeDTO{
				Type:     "tcp",
				Interval: yaml.Duration(15 * time.Second),
				Timeout:  yaml.Duration(3 * time.Second),
			},
			expectedType:             "tcp",
			expectedInterval:         15 * time.Second,
			expectedTimeout:          3 * time.Second,
			expectedSuccessThreshold: 1,
			expectedFailureThreshold: 3,
			expectedMethod:           "GET",
			expectedStatusCode:       200,
		},
		{
			name: "grpc probe with service",
			dto: &yaml.ProbeDTO{
				Type:     "grpc",
				Interval: yaml.Duration(20 * time.Second),
				Timeout:  yaml.Duration(4 * time.Second),
				Service:  "my.grpc.Service",
			},
			expectedType:             "grpc",
			expectedInterval:         20 * time.Second,
			expectedTimeout:          4 * time.Second,
			expectedSuccessThreshold: 1,
			expectedFailureThreshold: 3,
			expectedService:          "my.grpc.Service",
			expectedMethod:           "GET",
			expectedStatusCode:       200,
		},
		{
			name: "exec probe with command",
			dto: &yaml.ProbeDTO{
				Type:    "exec",
				Command: "/bin/check-health",
				Args:    []string{"--verbose"},
			},
			expectedType:             "exec",
			expectedInterval:         10 * time.Second,
			expectedTimeout:          5 * time.Second,
			expectedSuccessThreshold: 1,
			expectedFailureThreshold: 3,
			expectedCommand:          "/bin/check-health",
			expectedMethod:           "GET",
			expectedStatusCode:       200,
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.dto.ToDomain()

			// Verify probe configuration fields are correctly mapped.
			assert.Equal(t, testCase.expectedType, result.Type)
			assert.Equal(t, testCase.expectedInterval, result.Interval.Duration())
			assert.Equal(t, testCase.expectedTimeout, result.Timeout.Duration())
			assert.Equal(t, testCase.expectedSuccessThreshold, result.SuccessThreshold)
			assert.Equal(t, testCase.expectedFailureThreshold, result.FailureThreshold)
			assert.Equal(t, testCase.expectedPath, result.Path)
			assert.Equal(t, testCase.expectedMethod, result.Method)
			assert.Equal(t, testCase.expectedStatusCode, result.StatusCode)
			assert.Equal(t, testCase.expectedService, result.Service)
			assert.Equal(t, testCase.expectedCommand, result.Command)
		})
	}
}
