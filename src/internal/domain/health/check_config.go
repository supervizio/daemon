// Package health provides domain abstractions for service probing.
package health

import "time"

// DefaultTimeout is the default probe timeout.
const DefaultTimeout time.Duration = 5 * time.Second

// DefaultInterval is the default probe interval.
const DefaultInterval time.Duration = 10 * time.Second

// DefaultSuccessThreshold is the default number of successes needed.
const DefaultSuccessThreshold int = 1

// DefaultFailureThreshold is the default number of failures needed.
const DefaultFailureThreshold int = 3

// CheckConfig contains probe configuration parameters.
// It defines timing and threshold settings for probe execution.
type CheckConfig struct {
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

// NewCheckConfig creates a new probe configuration with default values.
//
// Returns:
//   - CheckConfig: a configuration with default timeout, interval, and thresholds.
func NewCheckConfig() CheckConfig {
	// return config with default values
	return CheckConfig{
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
//   - CheckConfig: a copy of the config with updated timeout.
func (c CheckConfig) WithTimeout(timeout time.Duration) CheckConfig {
	// update timeout and return copy
	c.Timeout = timeout
	// return updated config
	return c
}

// WithInterval returns a copy with the specified interval.
//
// Params:
//   - interval: the new interval value.
//
// Returns:
//   - CheckConfig: a copy of the config with updated interval.
func (c CheckConfig) WithInterval(interval time.Duration) CheckConfig {
	// update interval and return copy
	c.Interval = interval
	// return updated config
	return c
}

// WithSuccessThreshold returns a copy with the specified success threshold.
//
// Params:
//   - threshold: the new success threshold value.
//
// Returns:
//   - CheckConfig: a copy of the config with updated success threshold.
func (c CheckConfig) WithSuccessThreshold(threshold int) CheckConfig {
	// update success threshold and return copy
	c.SuccessThreshold = threshold
	// return updated config
	return c
}

// WithFailureThreshold returns a copy with the specified failure threshold.
//
// Params:
//   - threshold: the new failure threshold value.
//
// Returns:
//   - CheckConfig: a copy of the config with updated failure threshold.
func (c CheckConfig) WithFailureThreshold(threshold int) CheckConfig {
	// update failure threshold and return copy
	c.FailureThreshold = threshold
	// return updated config
	return c
}

// Validate validates the configuration.
//
// Returns:
//   - error: nil if valid, otherwise an error describing the problem.
func (c CheckConfig) Validate() error {
	// timeout must be positive
	if c.Timeout <= 0 {
		// invalid timeout
		return ErrInvalidTimeout
	}

	// interval must be positive
	if c.Interval <= 0 {
		// invalid interval
		return ErrInvalidInterval
	}

	// success threshold must be positive
	if c.SuccessThreshold <= 0 {
		// invalid success threshold
		return ErrInvalidSuccessThreshold
	}

	// failure threshold must be positive
	if c.FailureThreshold <= 0 {
		// invalid failure threshold
		return ErrInvalidFailureThreshold
	}

	// all validations passed
	return nil
}
