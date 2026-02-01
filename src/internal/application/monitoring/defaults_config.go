// Package monitoring provides the application service for external target monitoring.
package monitoring

import "time"

// DefaultsConfig contains default timing and threshold values.
// These values are used when targets do not specify their own configuration.
type DefaultsConfig struct {
	// Interval is the default probe interval.
	Interval time.Duration

	// Timeout is the default probe timeout.
	Timeout time.Duration

	// SuccessThreshold is the default consecutive successes for healthy.
	SuccessThreshold int

	// FailureThreshold is the default consecutive failures for unhealthy.
	FailureThreshold int
}

// NewDefaultsConfig creates a new DefaultsConfig with default values.
//
// Returns:
//   - DefaultsConfig: a new instance with package defaults applied.
func NewDefaultsConfig() DefaultsConfig {
	// construct config with default values
	return DefaultsConfig{
		Interval:         DefaultInterval,
		Timeout:          DefaultTimeout,
		SuccessThreshold: DefaultSuccessThreshold,
		FailureThreshold: DefaultFailureThreshold,
	}
}
