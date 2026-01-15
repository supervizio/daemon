// Package config provides domain value objects for service configuration.
package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/config"
)

// TestConfig_FindService tests the FindService method of Config.
//
// Params:
//   - t: testing context
func TestConfig_FindService(t *testing.T) {
	cfg := &config.Config{
		Services: []config.ServiceConfig{
			{Name: "web", Command: "/bin/web"},
			{Name: "api", Command: "/bin/api"},
			{Name: "worker", Command: "/bin/worker"},
		},
	}

	// testCase defines a test case for FindService
	type testCase struct {
		name        string
		serviceName string
		wantNil     bool
		wantName    string
		wantCommand string
	}

	// tests defines all test cases for FindService
	tests := []testCase{
		{
			name:        "finds existing service",
			serviceName: "api",
			wantNil:     false,
			wantName:    "api",
			wantCommand: "/bin/api",
		},
		{
			name:        "returns nil for non-existing service",
			serviceName: "unknown",
			wantNil:     true,
			wantName:    "",
			wantCommand: "",
		},
	}

	// Iterate over all test cases
	for _, tc := range tests {
		// Run each test case as a subtest
		t.Run(tc.name, func(t *testing.T) {
			svc := cfg.FindService(tc.serviceName)
			// Check if the result matches expectations
			if tc.wantNil {
				assert.Nil(t, svc)
			} else {
				assert.NotNil(t, svc)
				assert.Equal(t, tc.wantName, svc.Name)
				assert.Equal(t, tc.wantCommand, svc.Command)
			}
		})
	}
}

// TestConfig_Validate tests the Validate method of Config.
//
// Params:
//   - t: testing context
func TestConfig_Validate(t *testing.T) {
	// testCase defines a test case for Validate
	type testCase struct {
		name      string
		cfg       *config.Config
		wantError bool
	}

	// tests defines all test cases for Validate
	tests := []testCase{
		{
			name: "valid config with at least one service",
			cfg: &config.Config{
				Services: []config.ServiceConfig{
					{Name: "app", Command: "/bin/app"},
				},
			},
			wantError: false,
		},
		{
			name: "invalid config with no services",
			cfg: &config.Config{
				Services: nil,
			},
			wantError: true,
		},
	}

	// Iterate over all test cases
	for _, tc := range tests {
		// Run each test case as a subtest
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			// Check if the error matches expectations
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestConfig_GetServiceLogPath tests the GetServiceLogPath method of Config.
//
// Params:
//   - t: testing context
func TestConfig_GetServiceLogPath(t *testing.T) {
	// testCase defines a test case for GetServiceLogPath
	type testCase struct {
		name        string
		baseDir     string
		serviceName string
		filename    string
		want        string
	}

	// tests defines all test cases for GetServiceLogPath
	tests := []testCase{
		{
			name:        "constructs correct path with service name and filename",
			baseDir:     "/var/log/daemon",
			serviceName: "myservice",
			filename:    "stdout.log",
			want:        "/var/log/daemon/myservice/stdout.log",
		},
	}

	// Iterate over all test cases
	for _, tc := range tests {
		// Run each test case as a subtest
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				Logging: config.LoggingConfig{
					BaseDir: tc.baseDir,
				},
			}
			path := cfg.GetServiceLogPath(tc.serviceName, tc.filename)
			assert.Equal(t, tc.want, path)
		})
	}
}

// TestDefaultConfig tests the DefaultConfig function returns correct defaults.
//
// Params:
//   - t: testing context
func TestDefaultConfig(t *testing.T) {
	// testCase defines a test case for DefaultConfig
	type testCase struct {
		name  string
		check func(t *testing.T, cfg *config.Config)
	}

	// tests defines all test cases for DefaultConfig
	tests := []testCase{
		{
			name: "returns correct default values",
			check: func(t *testing.T, cfg *config.Config) {
				assert.Equal(t, "1", cfg.Version)
				assert.Equal(t, "/var/log/daemon", cfg.Logging.BaseDir)
				assert.Equal(t, "iso8601", cfg.Logging.Defaults.TimestampFormat)
				assert.Equal(t, "100MB", cfg.Logging.Defaults.Rotation.MaxSize)
				assert.Equal(t, 10, cfg.Logging.Defaults.Rotation.MaxFiles)
			},
		},
	}

	// Iterate over all test cases
	for _, tc := range tests {
		// Run each test case as a subtest
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			tc.check(t, cfg)
		})
	}
}

// TestNewConfig tests the NewConfig constructor function.
//
// Params:
//   - t: testing context
func TestNewConfig(t *testing.T) {
	// testCase defines a test case for NewConfig
	type testCase struct {
		name         string
		services     []config.ServiceConfig
		wantVersion  string
		wantCount    int
		wantFirstSvc string
	}

	// tests defines all test cases for NewConfig
	tests := []testCase{
		{
			name:         "creates config with empty services",
			services:     nil,
			wantVersion:  "1",
			wantCount:    0,
			wantFirstSvc: "",
		},
		{
			name: "creates config with single service",
			services: []config.ServiceConfig{
				{Name: "app1", Command: "/bin/app1"},
			},
			wantVersion:  "1",
			wantCount:    1,
			wantFirstSvc: "app1",
		},
		{
			name: "creates config with multiple services",
			services: []config.ServiceConfig{
				{Name: "web", Command: "/bin/web"},
				{Name: "api", Command: "/bin/api"},
				{Name: "worker", Command: "/bin/worker"},
			},
			wantVersion:  "1",
			wantCount:    3,
			wantFirstSvc: "web",
		},
	}

	// Iterate over all test cases
	for _, tc := range tests {
		// Run each test case as a subtest
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.NewConfig(tc.services)
			assert.NotNil(t, cfg)
			assert.Equal(t, tc.wantVersion, cfg.Version)
			assert.Len(t, cfg.Services, tc.wantCount)
			// Verify logging defaults are set
			assert.NotEmpty(t, cfg.Logging.BaseDir)
			// Verify first service name if services exist
			if tc.wantCount > 0 {
				assert.Equal(t, tc.wantFirstSvc, cfg.Services[0].Name)
			}
		})
	}
}

// TestNewServiceConfig tests the NewServiceConfig constructor function.
//
// Params:
//   - t: testing context
func TestNewServiceConfig(t *testing.T) {
	// testCase defines a test case for NewServiceConfig
	type testCase struct {
		name        string
		serviceName string
		command     string
		wantName    string
		wantCommand string
		wantPolicy  config.RestartPolicy
		wantRetries int
	}

	// tests defines all test cases for NewServiceConfig
	tests := []testCase{
		{
			name:        "creates service config with correct defaults",
			serviceName: "myapp",
			command:     "/usr/bin/myapp",
			wantName:    "myapp",
			wantCommand: "/usr/bin/myapp",
			wantPolicy:  config.RestartOnFailure,
			wantRetries: 3,
		},
	}

	// Iterate over all test cases
	for _, tc := range tests {
		// Run each test case as a subtest
		t.Run(tc.name, func(t *testing.T) {
			svc := config.NewServiceConfig(tc.serviceName, tc.command)
			assert.Equal(t, tc.wantName, svc.Name)
			assert.Equal(t, tc.wantCommand, svc.Command)
			assert.Equal(t, tc.wantPolicy, svc.Restart.Policy)
			assert.Equal(t, tc.wantRetries, svc.Restart.MaxRetries)
		})
	}
}
