// Package healthcheck provides domain abstractions for service probing.
package healthcheck

import "time"

// DefaultTimeout is the default probe timeout.
const DefaultTimeout time.Duration = 5 * time.Second

// DefaultInterval is the default probe interval.
const DefaultInterval time.Duration = 10 * time.Second

// DefaultSuccessThreshold is the default number of successes needed.
const DefaultSuccessThreshold int = 1

// DefaultFailureThreshold is the default number of failures needed.
const DefaultFailureThreshold int = 3

// Config contains probe configuration parameters.
// It defines timing and threshold settings for probe execution.
type Config struct {
	// Timeout is the maximum duration for a single probe execution.
	// If the probe doesn't complete within this time, it fails.
	Timeout time.Duration

	// Interval is the time between probe executions.
	// The probe scheduler uses this to determine when to run the next probe.
	Interval time.Duration

	// SuccessThreshold is the number of consecutive successes needed
	// to transition from unhealthy to healthy state.
	SuccessThreshold int

	// FailureThreshold is the number of consecutive failures needed
	// to transition from healthy to unhealthy state.
	FailureThreshold int
}

// NewConfig creates a new probe configuration with default values.
//
// Returns:
//   - Config: a configuration with default timeout, interval, and thresholds.
func NewConfig() Config {
	// Return configuration with defaults.
	return Config{
		Timeout:          DefaultTimeout,
		Interval:         DefaultInterval,
		SuccessThreshold: DefaultSuccessThreshold,
		FailureThreshold: DefaultFailureThreshold,
	}
}

// WithTimeout returns a copy with the specified timeout.
//
// Params:
//   - timeout: the new timeout value.
//
// Returns:
//   - Config: a copy of the config with updated timeout.
func (c Config) WithTimeout(timeout time.Duration) Config {
	// Set the new timeout value.
	c.Timeout = timeout

	// Return copy with new timeout.
	return c
}

// WithInterval returns a copy with the specified interval.
//
// Params:
//   - interval: the new interval value.
//
// Returns:
//   - Config: a copy of the config with updated interval.
func (c Config) WithInterval(interval time.Duration) Config {
	// Set the new interval value.
	c.Interval = interval

	// Return copy with new interval.
	return c
}

// WithSuccessThreshold returns a copy with the specified success threshold.
//
// Params:
//   - threshold: the new success threshold value.
//
// Returns:
//   - Config: a copy of the config with updated success threshold.
func (c Config) WithSuccessThreshold(threshold int) Config {
	// Set the new success threshold value.
	c.SuccessThreshold = threshold

	// Return copy with new threshold.
	return c
}

// WithFailureThreshold returns a copy with the specified failure threshold.
//
// Params:
//   - threshold: the new failure threshold value.
//
// Returns:
//   - Config: a copy of the config with updated failure threshold.
func (c Config) WithFailureThreshold(threshold int) Config {
	// Set the new failure threshold value.
	c.FailureThreshold = threshold

	// Return copy with new threshold.
	return c
}

// Validate validates the configuration.
//
// Returns:
//   - error: nil if valid, otherwise an error describing the problem.
func (c Config) Validate() error {
	// Check timeout is positive.
	if c.Timeout <= 0 {
		// Return error for invalid timeout.
		return ErrInvalidTimeout
	}

	// Check interval is positive.
	if c.Interval <= 0 {
		// Return error for invalid interval.
		return ErrInvalidInterval
	}

	// Check success threshold is positive.
	if c.SuccessThreshold <= 0 {
		// Return error for invalid threshold.
		return ErrInvalidSuccessThreshold
	}

	// Check failure threshold is positive.
	if c.FailureThreshold <= 0 {
		// Return error for invalid threshold.
		return ErrInvalidFailureThreshold
	}

	// Return nil if all validations pass.
	return nil
}
