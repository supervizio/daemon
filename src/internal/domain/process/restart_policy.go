// Package process provides domain entities and value objects for process lifecycle management.
package process

import (
	"time"

	"github.com/kodflow/daemon/internal/domain/service"
)

// Restart tracker constants.
const (
	// DefaultStabilityWindow is the duration of stable running required
	// before the restart counter is reset.
	DefaultStabilityWindow time.Duration = 5 * time.Minute

	// DefaultMaxDelayMultiplier is the default multiplier for max delay
	// when no explicit max delay is configured.
	DefaultMaxDelayMultiplier int = 10

	// MaxBackoffAttempts is the maximum number of attempts to use in
	// backoff calculation to prevent integer overflow.
	MaxBackoffAttempts int = 30
)

// RestartTracker tracks restart attempts for a service and implements
// exponential backoff logic to prevent rapid restart cycles.
type RestartTracker struct {
	// config holds the restart configuration from the service definition.
	config *service.RestartConfig

	// attempts tracks the current number of restart attempts since the last reset.
	attempts int

	// lastAttempt records the timestamp of the most recent restart attempt.
	lastAttempt time.Time

	// window defines the duration of stable running required before
	// the restart counter is reset.
	window time.Duration
}

// NewRestartTracker creates a new restart tracker with the given configuration.
//
// Params:
//   - cfg: the restart configuration containing policy, max retries, and delays
//
// Returns:
//   - *RestartTracker: a new restart tracker instance
func NewRestartTracker(cfg *service.RestartConfig) *RestartTracker {
	// Initialize and return a new tracker with default stability window.
	return &RestartTracker{
		config: cfg,
		window: DefaultStabilityWindow,
	}
}

// ShouldRestart determines if a restart should be attempted based on the
// configured restart policy and the current number of attempts.
//
// Params:
//   - exitCode: the exit code returned by the process when it terminated
//
// Returns:
//   - bool: true if a restart should be attempted
func (rt *RestartTracker) ShouldRestart(exitCode int) bool {
	// Evaluate restart decision based on the configured policy.
	switch rt.config.Policy {
	// RestartAlways: restart up to MaxRetries regardless of exit code.
	case service.RestartAlways:
		// Return true if attempts remain within the configured limit.
		return rt.attempts < rt.config.MaxRetries
	// RestartOnFailure: restart only if the process exited with an error.
	case service.RestartOnFailure:
		// Check if the process exited successfully (exit code 0).
		if exitCode == 0 {
			// Do not restart on successful exit.
			return false
		}
		// Restart if attempts remain within the configured limit.
		return rt.attempts < rt.config.MaxRetries
	// RestartNever: never restart the process.
	case service.RestartNever:
		// Always return false for never-restart policy.
		return false
	// RestartUnless: always restart unless explicitly stopped.
	case service.RestartUnless:
		// Always return true for unless-stopped policy.
		return true
	// Default case: unknown policy, do not restart.
	default:
		// Return false for safety on unknown policies.
		return false
	}
}

// RecordAttempt records a restart attempt by incrementing the attempt counter
// and updating the last attempt timestamp.
//
// Returns:
//   - void: this method modifies the tracker state
func (rt *RestartTracker) RecordAttempt() {
	rt.attempts++
	rt.lastAttempt = time.Now()
}

// Reset resets the restart counter to zero.
//
// Returns:
//   - void: this method modifies the tracker state
func (rt *RestartTracker) Reset() {
	rt.attempts = 0
}

// MaybeReset resets the counter if the process has been running stably
// for at least the configured stability window duration.
//
// Params:
//   - uptime: the duration the process has been running since the last restart
//
// Returns:
//   - void: this method may modify the tracker state
func (rt *RestartTracker) MaybeReset(uptime time.Duration) {
	// Check if the uptime exceeds the stability window.
	if uptime >= rt.window {
		rt.Reset()
	}
}

// Attempts returns the current number of restart attempts.
//
// Returns:
//   - int: the current restart attempt count
func (rt *RestartTracker) Attempts() int {
	// Return the current attempt counter value.
	return rt.attempts
}

// NextDelay calculates the next restart delay using exponential backoff.
// The delay doubles with each attempt: delay = baseDelay * 2^attempts,
// capped at the configured maximum delay.
//
// Returns:
//   - time.Duration: the calculated delay before the next restart attempt
func (rt *RestartTracker) NextDelay() time.Duration {
	baseDelay := rt.config.Delay.Duration()
	maxDelay := rt.config.DelayMax.Duration()

	// Apply default max delay if not explicitly configured.
	if maxDelay == 0 {
		maxDelay = baseDelay * time.Duration(DefaultMaxDelayMultiplier)
	}

	// Exponential backoff: delay * 2^attempts
	// Cap attempts to prevent overflow
	attempts := min(rt.attempts, MaxBackoffAttempts)
	// #nosec G115 - attempts is capped to MaxBackoffAttempts (30), safe for uint conversion
	delay := baseDelay * time.Duration(1<<uint(attempts))

	// Return the smaller of the calculated delay and the maximum delay.
	return min(delay, maxDelay)
}

// IsExhausted returns true if all restart attempts have been used.
//
// Returns:
//   - bool: true if the restart attempt limit has been reached
func (rt *RestartTracker) IsExhausted() bool {
	// Compare current attempts against the configured maximum.
	return rt.attempts >= rt.config.MaxRetries
}

// SetWindow sets the stability window duration.
//
// Params:
//   - window: the new stability window duration
//
// Returns:
//   - void: this method modifies the tracker state
func (rt *RestartTracker) SetWindow(window time.Duration) {
	rt.window = window
}
