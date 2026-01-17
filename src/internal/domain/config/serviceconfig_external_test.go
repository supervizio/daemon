// Package config_test provides black-box tests for ServiceConfig.
package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/shared"
)

// TestServiceConfig_NewServiceConfig verifies the NewServiceConfig constructor.
//
// Params:
//   - t: testing context for assertions.
func TestServiceConfig_NewServiceConfig(t *testing.T) {
	// defaultMaxRetries is the expected default number of restart attempts.
	const defaultMaxRetries int = 3
	// defaultRestartDelaySeconds is the expected default delay between restarts.
	const defaultRestartDelaySeconds int = 5

	tests := []struct {
		name            string
		serviceName     string
		command         string
		expectedName    string
		expectedCommand string
		expectedPolicy  config.RestartPolicy
		expectedRetries int
		expectedDelay   shared.Duration
	}{
		{
			name:            "basic service",
			serviceName:     "web",
			command:         "/bin/web",
			expectedName:    "web",
			expectedCommand: "/bin/web",
			expectedPolicy:  config.RestartOnFailure,
			expectedRetries: defaultMaxRetries,
			expectedDelay:   shared.Seconds(defaultRestartDelaySeconds),
		},
		{
			name:            "empty name",
			serviceName:     "",
			command:         "/bin/app",
			expectedName:    "",
			expectedCommand: "/bin/app",
			expectedPolicy:  config.RestartOnFailure,
			expectedRetries: defaultMaxRetries,
			expectedDelay:   shared.Seconds(defaultRestartDelaySeconds),
		},
		{
			name:            "empty command",
			serviceName:     "worker",
			command:         "",
			expectedName:    "worker",
			expectedCommand: "",
			expectedPolicy:  config.RestartOnFailure,
			expectedRetries: defaultMaxRetries,
			expectedDelay:   shared.Seconds(defaultRestartDelaySeconds),
		},
		{
			name:            "command with path",
			serviceName:     "api",
			command:         "/usr/local/bin/api-server",
			expectedName:    "api",
			expectedCommand: "/usr/local/bin/api-server",
			expectedPolicy:  config.RestartOnFailure,
			expectedRetries: defaultMaxRetries,
			expectedDelay:   shared.Seconds(defaultRestartDelaySeconds),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NewServiceConfig(tt.serviceName, tt.command)
			assert.Equal(t, tt.expectedName, cfg.Name)
			assert.Equal(t, tt.expectedCommand, cfg.Command)
			assert.Equal(t, tt.expectedPolicy, cfg.Restart.Policy)
			assert.Equal(t, tt.expectedRetries, cfg.Restart.MaxRetries)
			assert.Equal(t, tt.expectedDelay, cfg.Restart.Delay)
		})
	}
}

// TestServiceConfig_Fields verifies ServiceConfig field access.
//
// Params:
//   - t: testing context for assertions.
func TestServiceConfig_Fields(t *testing.T) {
	tests := []struct {
		name         string
		cfg          config.ServiceConfig
		expectedArgs []string
		expectedUser string
		expectedDir  string
	}{
		{
			name: "with args",
			cfg: config.ServiceConfig{
				Name:    "test",
				Command: "/bin/test",
				Args:    []string{"-v", "--port=8080"},
			},
			expectedArgs: []string{"-v", "--port=8080"},
			expectedUser: "",
			expectedDir:  "",
		},
		{
			name: "with user and directory",
			cfg: config.ServiceConfig{
				Name:             "app",
				Command:          "/bin/app",
				User:             "daemon",
				WorkingDirectory: "/opt/app",
			},
			expectedArgs: nil,
			expectedUser: "daemon",
			expectedDir:  "/opt/app",
		},
		{
			name: "oneshot service",
			cfg: config.ServiceConfig{
				Name:    "setup",
				Command: "/bin/setup.sh",
				Oneshot: true,
			},
			expectedArgs: nil,
			expectedUser: "",
			expectedDir:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedArgs, tt.cfg.Args)
			assert.Equal(t, tt.expectedUser, tt.cfg.User)
			assert.Equal(t, tt.expectedDir, tt.cfg.WorkingDirectory)
		})
	}
}
