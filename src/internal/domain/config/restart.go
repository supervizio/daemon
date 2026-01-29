// Package config provides domain value objects for service configuration.
package config

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
	// StabilityWindow specifies the duration of stable running required
	// before the restart counter resets. If not set, defaults to 5 minutes.
	StabilityWindow shared.Duration
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
	// convert policy to string
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
	// determine restart based on policy
	switch r.Policy {
	// always restart if under retry limit
	case RestartAlways:
		// restart if under retry limit
		return attempts < r.MaxRetries
	// restart only on failure
	case RestartOnFailure:
		// skip restart on clean exit
		if exitCode == 0 {
			// no restart on success
			return false
		}
		// restart on failure if under limit
		return attempts < r.MaxRetries
	// never restart the service
	case RestartNever:
		// never restart
		return false
	// restart unless explicitly stopped
	case RestartUnless:
		// restart unless explicitly stopped
		return true
	// unknown policy
	default:
		// unknown policy defaults to no restart
		return false
	}
}

// DefaultRestartConfig returns a RestartConfig with sensible defaults.
//
// Returns:
//   - RestartConfig: a configuration with on-failure policy and standard retry settings.
func DefaultRestartConfig() RestartConfig {
	// create config with on-failure policy
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
	// create config with specified policy
	return RestartConfig{
		Policy:     policy,
		MaxRetries: defaultRestartMaxRetries,
		Delay:      shared.Seconds(defaultRestartDelaySecs),
	}
}
