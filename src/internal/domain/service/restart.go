// Package service provides domain value objects for service configuration.
package service

import "github.com/kodflow/daemon/internal/domain/shared"

const (
	// defaultRestartMaxRetries is the default number of restart attempts.
	defaultRestartMaxRetries int = 3
	// defaultRestartDelaySecs is the default delay in seconds between restart attempts.
	defaultRestartDelaySecs int = 5
)

// RestartConfig defines service restart behavior.
// It controls restart policy, retry limits, and exponential backoff delays.
type RestartConfig struct {
	// Policy specifies when the service should be restarted.
	Policy RestartPolicy
	// MaxRetries specifies the maximum number of restart attempts.
	MaxRetries int
	// Delay specifies the initial delay between restart attempts.
	Delay shared.Duration
	// DelayMax specifies the maximum delay for exponential backoff.
	DelayMax shared.Duration
}

// RestartPolicy defines when to restart a service.
type RestartPolicy string

// Restart policy constants.
const (
	// RestartAlways restarts the service regardless of exit status.
	RestartAlways RestartPolicy = "always"
	// RestartOnFailure restarts only on non-zero exit code.
	RestartOnFailure RestartPolicy = "on-failure"
	// RestartNever never restarts the service after exit.
	RestartNever RestartPolicy = "never"
	// RestartUnless restarts unless the service was explicitly stopped.
	RestartUnless RestartPolicy = "unless-stopped"
)

// String returns the string representation of the restart policy.
//
// Returns:
//   - string: the policy value as a string.
func (p RestartPolicy) String() string {
	// Convert policy to its underlying string type.
	return string(p)
}

// ShouldRestartOnExit determines if the service should restart based on exit code.
//
// Params:
//   - exitCode: the exit code returned by the process.
//   - attempts: the number of restart attempts already made.
//
// Returns:
//   - bool: true if the service should be restarted, false otherwise.
func (r *RestartConfig) ShouldRestartOnExit(exitCode, attempts int) bool {
	// Evaluate restart policy to determine behavior.
	switch r.Policy {
	// Handle always restart policy.
	case RestartAlways:
		// Restart if attempts are below the maximum retry limit.
		return attempts < r.MaxRetries
	// Handle restart on failure policy.
	case RestartOnFailure:
		// Check if process exited successfully.
		if exitCode == 0 {
			// Do not restart on successful exit.
			return false
		}
		// Restart failed process if attempts are below limit.
		return attempts < r.MaxRetries
	// Handle never restart policy.
	case RestartNever:
		// Never restart regardless of exit status.
		return false
	// Handle unless-stopped policy.
	case RestartUnless:
		// Always restart unless explicitly stopped.
		return true
	// Handle unknown policies.
	default:
		// Default to no restart for unrecognized policies.
		return false
	}
}

// DefaultRestartConfig returns a RestartConfig with sensible defaults.
//
// Returns:
//   - RestartConfig: a configuration with on-failure policy and standard retry settings.
func DefaultRestartConfig() RestartConfig {
	// Build default configuration with reasonable values.
	return RestartConfig{
		Policy:     RestartOnFailure,
		MaxRetries: defaultRestartMaxRetries,
		Delay:      shared.Seconds(defaultRestartDelaySecs),
	}
}

// NewRestartConfig creates a new RestartConfig with the given policy.
//
// Params:
//   - policy: the restart policy to use.
//
// Returns:
//   - RestartConfig: a restart configuration with the given policy and default settings.
func NewRestartConfig(policy RestartPolicy) RestartConfig {
	// Return a new RestartConfig with the specified policy and default values.
	return RestartConfig{
		Policy:     policy,
		MaxRetries: defaultRestartMaxRetries,
		Delay:      shared.Seconds(defaultRestartDelaySecs),
	}
}
