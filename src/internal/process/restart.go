package process

import (
	"time"

	"github.com/kodflow/daemon/internal/config"
)

// RestartTracker tracks restart attempts for a service.
type RestartTracker struct {
	config     *config.RestartConfig
	attempts   int
	lastAttempt time.Time
	window     time.Duration
}

// NewRestartTracker creates a new restart tracker.
func NewRestartTracker(cfg *config.RestartConfig) *RestartTracker {
	return &RestartTracker{
		config: cfg,
		window: 5 * time.Minute, // Reset counter after 5 minutes of stable running
	}
}

// ShouldRestart determines if a restart should be attempted.
func (rt *RestartTracker) ShouldRestart(exitCode int) bool {
	switch rt.config.Policy {
	case config.RestartAlways:
		return rt.attempts < rt.config.MaxRetries
	case config.RestartOnFailure:
		if exitCode == 0 {
			return false
		}
		return rt.attempts < rt.config.MaxRetries
	case config.RestartNever:
		return false
	case config.RestartUnless:
		// Always restart unless manually stopped
		return true
	default:
		return false
	}
}

// RecordAttempt records a restart attempt.
func (rt *RestartTracker) RecordAttempt() {
	rt.attempts++
	rt.lastAttempt = time.Now()
}

// Reset resets the restart counter.
func (rt *RestartTracker) Reset() {
	rt.attempts = 0
}

// MaybeReset resets the counter if the process has been stable.
func (rt *RestartTracker) MaybeReset(uptime time.Duration) {
	if uptime >= rt.window {
		rt.Reset()
	}
}

// Attempts returns the current number of restart attempts.
func (rt *RestartTracker) Attempts() int {
	return rt.attempts
}

// NextDelay calculates the next restart delay with exponential backoff.
func (rt *RestartTracker) NextDelay() time.Duration {
	baseDelay := rt.config.Delay.Duration()
	maxDelay := rt.config.DelayMax.Duration()

	if maxDelay == 0 {
		maxDelay = baseDelay * 10
	}

	// Exponential backoff: delay * 2^attempts
	delay := baseDelay * time.Duration(1<<uint(rt.attempts))

	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

// IsExhausted returns true if all restart attempts have been used.
func (rt *RestartTracker) IsExhausted() bool {
	return rt.attempts >= rt.config.MaxRetries
}
