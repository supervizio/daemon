// Package yaml provides YAML configuration loading infrastructure.
package yaml

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/service"
)

// Test_applyDefaults tests the applyDefaults function.
//
// Params:
//   - t: testing context for assertions and error reporting
func Test_applyDefaults(t *testing.T) {
	// Define test cases for table-driven testing.
	tests := []struct {
		name                    string
		cfg                     ConfigDTO
		expectedVersion         string
		expectedBaseDir         string
		expectedTimestampFormat string
		expectedMaxSize         string
		expectedMaxFiles        int
	}{
		{
			name:                    "empty_config_gets_all_defaults",
			cfg:                     ConfigDTO{},
			expectedVersion:         defaultVersion,
			expectedBaseDir:         defaultBaseDir,
			expectedTimestampFormat: defaultTimestampFormat,
			expectedMaxSize:         defaultMaxSize,
			expectedMaxFiles:        defaultMaxFiles,
		},
		{
			name: "partial_config_preserves_set_values",
			cfg: ConfigDTO{
				Version: "2",
				Logging: LoggingConfigDTO{
					BaseDir: "/custom/log/path",
				},
			},
			expectedVersion:         "2",
			expectedBaseDir:         "/custom/log/path",
			expectedTimestampFormat: defaultTimestampFormat,
			expectedMaxSize:         defaultMaxSize,
			expectedMaxFiles:        defaultMaxFiles,
		},
		{
			name: "custom_timestamp_format_preserved",
			cfg: ConfigDTO{
				Logging: LoggingConfigDTO{
					Defaults: LogDefaultsDTO{
						TimestampFormat: "rfc3339",
					},
				},
			},
			expectedVersion:         defaultVersion,
			expectedBaseDir:         defaultBaseDir,
			expectedTimestampFormat: "rfc3339",
			expectedMaxSize:         defaultMaxSize,
			expectedMaxFiles:        defaultMaxFiles,
		},
		{
			name: "custom_rotation_max_size_preserved",
			cfg: ConfigDTO{
				Logging: LoggingConfigDTO{
					Defaults: LogDefaultsDTO{
						Rotation: RotationConfigDTO{
							MaxSize: "200MB",
						},
					},
				},
			},
			expectedVersion:         defaultVersion,
			expectedBaseDir:         defaultBaseDir,
			expectedTimestampFormat: defaultTimestampFormat,
			expectedMaxSize:         "200MB",
			expectedMaxFiles:        defaultMaxFiles,
		},
		{
			name: "custom_rotation_max_files_preserved",
			cfg: ConfigDTO{
				Logging: LoggingConfigDTO{
					Defaults: LogDefaultsDTO{
						Rotation: RotationConfigDTO{
							MaxFiles: 20,
						},
					},
				},
			},
			expectedVersion:         defaultVersion,
			expectedBaseDir:         defaultBaseDir,
			expectedTimestampFormat: defaultTimestampFormat,
			expectedMaxSize:         defaultMaxSize,
			expectedMaxFiles:        20,
		},
		{
			name: "config_with_services_applies_service_defaults",
			cfg: ConfigDTO{
				Services: []ServiceConfigDTO{
					{Name: "test-svc"},
				},
			},
			expectedVersion:         defaultVersion,
			expectedBaseDir:         defaultBaseDir,
			expectedTimestampFormat: defaultTimestampFormat,
			expectedMaxSize:         defaultMaxSize,
			expectedMaxFiles:        defaultMaxFiles,
		},
		{
			name: "all_custom_values_preserved",
			cfg: ConfigDTO{
				Version: "3",
				Logging: LoggingConfigDTO{
					BaseDir: "/opt/logs",
					Defaults: LogDefaultsDTO{
						TimestampFormat: "unix",
						Rotation: RotationConfigDTO{
							MaxSize:  "500MB",
							MaxFiles: 50,
						},
					},
				},
			},
			expectedVersion:         "3",
			expectedBaseDir:         "/opt/logs",
			expectedTimestampFormat: "unix",
			expectedMaxSize:         "500MB",
			expectedMaxFiles:        50,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply defaults to the configuration.
			applyDefaults(&tt.cfg)

			// Assert expected values.
			assert.Equal(t, tt.expectedVersion, tt.cfg.Version)
			assert.Equal(t, tt.expectedBaseDir, tt.cfg.Logging.BaseDir)
			assert.Equal(t, tt.expectedTimestampFormat, tt.cfg.Logging.Defaults.TimestampFormat)
			assert.Equal(t, tt.expectedMaxSize, tt.cfg.Logging.Defaults.Rotation.MaxSize)
			assert.Equal(t, tt.expectedMaxFiles, tt.cfg.Logging.Defaults.Rotation.MaxFiles)
		})
	}
}

// Test_applyServiceDefaults tests the applyServiceDefaults function.
//
// Params:
//   - t: testing context for assertions and error reporting
func Test_applyServiceDefaults(t *testing.T) {
	// Define test cases for table-driven testing.
	tests := []struct {
		name               string
		svc                ServiceConfigDTO
		logging            LoggingConfigDTO
		expectedStdoutFile string
		expectedStderrFile string
		expectedMaxRetries int
	}{
		{
			name: "service_gets_default_log_files",
			svc: ServiceConfigDTO{
				Name: "test-service",
			},
			logging: LoggingConfigDTO{
				Defaults: LogDefaultsDTO{
					TimestampFormat: "iso8601",
					Rotation: RotationConfigDTO{
						MaxSize:  "50MB",
						MaxFiles: 5,
					},
				},
			},
			expectedStdoutFile: "test-service.out.log",
			expectedStderrFile: "test-service.err.log",
			expectedMaxRetries: defaultMaxRetries,
		},
		{
			name: "service_preserves_custom_log_files",
			svc: ServiceConfigDTO{
				Name: "my-app",
				Logging: ServiceLoggingDTO{
					Stdout: LogStreamConfigDTO{
						File: "custom-stdout.log",
					},
					Stderr: LogStreamConfigDTO{
						File: "custom-stderr.log",
					},
				},
			},
			logging: LoggingConfigDTO{
				Defaults: LogDefaultsDTO{
					TimestampFormat: "iso8601",
				},
			},
			expectedStdoutFile: "custom-stdout.log",
			expectedStderrFile: "custom-stderr.log",
			expectedMaxRetries: defaultMaxRetries,
		},
		{
			name: "service_inherits_timestamp_format",
			svc: ServiceConfigDTO{
				Name: "inherit-svc",
			},
			logging: LoggingConfigDTO{
				Defaults: LogDefaultsDTO{
					TimestampFormat: "rfc3339",
					Rotation: RotationConfigDTO{
						MaxSize:  "100MB",
						MaxFiles: 10,
					},
				},
			},
			expectedStdoutFile: "inherit-svc.out.log",
			expectedStderrFile: "inherit-svc.err.log",
			expectedMaxRetries: defaultMaxRetries,
		},
		{
			name: "service_preserves_custom_stdout_only",
			svc: ServiceConfigDTO{
				Name: "partial-svc",
				Logging: ServiceLoggingDTO{
					Stdout: LogStreamConfigDTO{
						File: "my-stdout.log",
					},
				},
			},
			logging: LoggingConfigDTO{
				Defaults: LogDefaultsDTO{
					TimestampFormat: "iso8601",
				},
			},
			expectedStdoutFile: "my-stdout.log",
			expectedStderrFile: "partial-svc.err.log",
			expectedMaxRetries: defaultMaxRetries,
		},
		{
			name: "service_preserves_custom_stderr_only",
			svc: ServiceConfigDTO{
				Name: "partial-svc-2",
				Logging: ServiceLoggingDTO{
					Stderr: LogStreamConfigDTO{
						File: "my-stderr.log",
					},
				},
			},
			logging: LoggingConfigDTO{
				Defaults: LogDefaultsDTO{
					TimestampFormat: "iso8601",
				},
			},
			expectedStdoutFile: "partial-svc-2.out.log",
			expectedStderrFile: "my-stderr.log",
			expectedMaxRetries: defaultMaxRetries,
		},
		{
			name: "service_with_health_checks",
			svc: ServiceConfigDTO{
				Name: "health-svc",
				HealthChecks: []HealthCheckDTO{
					{Type: string(service.HealthCheckHTTP)},
				},
			},
			logging: LoggingConfigDTO{
				Defaults: LogDefaultsDTO{
					TimestampFormat: "iso8601",
					Rotation: RotationConfigDTO{
						MaxSize:  "100MB",
						MaxFiles: 10,
					},
				},
			},
			expectedStdoutFile: "health-svc.out.log",
			expectedStderrFile: "health-svc.err.log",
			expectedMaxRetries: defaultMaxRetries,
		},
		{
			name: "service_inherits_rotation_from_defaults",
			svc: ServiceConfigDTO{
				Name: "rotate-svc",
			},
			logging: LoggingConfigDTO{
				Defaults: LogDefaultsDTO{
					TimestampFormat: "iso8601",
					Rotation: RotationConfigDTO{
						MaxSize:  "250MB",
						MaxFiles: 25,
					},
				},
			},
			expectedStdoutFile: "rotate-svc.out.log",
			expectedStderrFile: "rotate-svc.err.log",
			expectedMaxRetries: defaultMaxRetries,
		},
		{
			name: "service_custom_stdout_timestamp_preserved",
			svc: ServiceConfigDTO{
				Name: "ts-svc",
				Logging: ServiceLoggingDTO{
					Stdout: LogStreamConfigDTO{
						TimestampFormat: "unix",
					},
				},
			},
			logging: LoggingConfigDTO{
				Defaults: LogDefaultsDTO{
					TimestampFormat: "iso8601",
				},
			},
			expectedStdoutFile: "ts-svc.out.log",
			expectedStderrFile: "ts-svc.err.log",
			expectedMaxRetries: defaultMaxRetries,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply service defaults.
			applyServiceDefaults(&tt.svc, &tt.logging)

			// Assert expected values.
			assert.Equal(t, tt.expectedStdoutFile, tt.svc.Logging.Stdout.File)
			assert.Equal(t, tt.expectedStderrFile, tt.svc.Logging.Stderr.File)
			assert.Equal(t, tt.expectedMaxRetries, tt.svc.Restart.MaxRetries)
		})
	}
}

// Test_applyRestartDefaults tests the applyRestartDefaults function.
//
// Params:
//   - t: testing context for assertions and error reporting
func Test_applyRestartDefaults(t *testing.T) {
	// Define test cases for table-driven testing.
	tests := []struct {
		name               string
		restart            RestartConfigDTO
		expectedPolicy     string
		expectedMaxRetries int
	}{
		{
			name:               "empty_restart_gets_defaults",
			restart:            RestartConfigDTO{},
			expectedPolicy:     string(service.RestartOnFailure),
			expectedMaxRetries: defaultMaxRetries,
		},
		{
			name: "custom_policy_preserved",
			restart: RestartConfigDTO{
				Policy: "always",
			},
			expectedPolicy:     "always",
			expectedMaxRetries: defaultMaxRetries,
		},
		{
			name: "custom_max_retries_preserved",
			restart: RestartConfigDTO{
				MaxRetries: 10,
			},
			expectedPolicy:     string(service.RestartOnFailure),
			expectedMaxRetries: 10,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply restart defaults.
			applyRestartDefaults(&tt.restart)

			// Assert expected values.
			assert.Equal(t, tt.expectedPolicy, tt.restart.Policy)
			assert.Equal(t, tt.expectedMaxRetries, tt.restart.MaxRetries)
		})
	}
}

// Test_applyHealthCheckDefaults tests the applyHealthCheckDefaults function.
//
// Params:
//   - t: testing context for assertions and error reporting
func Test_applyHealthCheckDefaults(t *testing.T) {
	// Define test cases for table-driven testing.
	tests := []struct {
		name               string
		hc                 HealthCheckDTO
		expectedRetries    int
		expectedMethod     string
		expectedStatusCode int
	}{
		{
			name:               "empty_health_check_gets_default_retries",
			hc:                 HealthCheckDTO{},
			expectedRetries:    defaultHealthRetries,
			expectedMethod:     "",
			expectedStatusCode: 0,
		},
		{
			name: "http_health_check_gets_http_defaults",
			hc: HealthCheckDTO{
				Type: string(service.HealthCheckHTTP),
			},
			expectedRetries:    defaultHealthRetries,
			expectedMethod:     defaultHTTPMethod,
			expectedStatusCode: defaultHTTPStatus,
		},
		{
			name: "http_with_custom_method_preserved",
			hc: HealthCheckDTO{
				Type:   string(service.HealthCheckHTTP),
				Method: "POST",
			},
			expectedRetries:    defaultHealthRetries,
			expectedMethod:     "POST",
			expectedStatusCode: defaultHTTPStatus,
		},
		{
			name: "tcp_health_check_no_http_defaults",
			hc: HealthCheckDTO{
				Type: string(service.HealthCheckTCP),
			},
			expectedRetries:    defaultHealthRetries,
			expectedMethod:     "",
			expectedStatusCode: 0,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply health check defaults.
			applyHealthCheckDefaults(&tt.hc)

			// Assert expected values.
			assert.Equal(t, tt.expectedRetries, tt.hc.Retries)
			assert.Equal(t, tt.expectedMethod, tt.hc.Method)
			assert.Equal(t, tt.expectedStatusCode, tt.hc.StatusCode)
		})
	}
}

// Test_parseDuration tests the parseDuration function.
//
// Params:
//   - t: testing context for assertions and error reporting
func Test_parseDuration(t *testing.T) {
	// Define test cases for table-driven testing.
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid_seconds",
			input:   "5s",
			wantErr: false,
		},
		{
			name:    "valid_minutes",
			input:   "1m",
			wantErr: false,
		},
		{
			name:    "valid_complex_duration",
			input:   "1m30s",
			wantErr: false,
		},
		{
			name:    "invalid_duration",
			input:   "invalid",
			wantErr: true,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the duration.
			_, err := parseDuration(tt.input)

			// Check error expectation.
			if tt.wantErr {
				// Assert error occurred when expected.
				assert.Error(t, err)
			} else {
				// Assert no error when success expected.
				assert.NoError(t, err)
			}
		})
	}
}
