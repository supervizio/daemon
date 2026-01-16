// Package config provides domain value objects for service configuration.
package config

import "github.com/kodflow/daemon/internal/domain/shared"

const (
	// defaultMaxRetries is the default number of restart attempts on failure.
	defaultMaxRetries int = 3
	// defaultRestartDelaySeconds is the default delay in seconds between restart attempts.
	defaultRestartDelaySeconds int = 5
)

// ServiceConfig defines a single service configuration.
// It specifies the command, environment, restart policy, and health checks.
type ServiceConfig struct {
	// Name is the unique identifier for this service.
	Name string
	// Command is the executable path or command to run.
	Command string
	// Args contains command-line arguments passed to the command.
	Args []string
	// User specifies the username under which the service runs.
	User string
	// Group specifies the group under which the service runs.
	Group string
	// WorkingDirectory specifies the working directory for the service process.
	WorkingDirectory string
	// Environment contains key-value pairs of environment variables.
	Environment map[string]string
	// Restart defines the restart behavior when the service exits.
	Restart RestartConfig
	// HealthChecks defines the health check configurations for this service.
	//
	// Deprecated: Use Listeners with Probe configuration instead.
	HealthChecks []HealthCheckConfig
	// Listeners defines the network listeners with probe configurations.
	// Each listener specifies a port and optional health probe.
	Listeners []ListenerConfig
	// Logging defines per-service logging configuration.
	Logging ServiceLogging
	// DependsOn lists service names that must start before this service.
	DependsOn []string
	// Oneshot indicates the service runs once and exits without restart.
	Oneshot bool
}

// NewServiceConfig creates a new ServiceConfig with the given name and command.
//
// Params:
//   - name: unique identifier for the service
//   - command: executable path or command to run
//
// Returns:
//   - ServiceConfig: service configuration with default restart policy
func NewServiceConfig(name, command string) ServiceConfig {
	// Return service config with default restart policy settings
	return ServiceConfig{
		Name:    name,
		Command: command,
		Restart: RestartConfig{
			Policy:     RestartOnFailure,
			MaxRetries: defaultMaxRetries,
			Delay:      shared.Seconds(defaultRestartDelaySeconds),
		},
	}
}
