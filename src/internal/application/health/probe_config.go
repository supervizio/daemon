// Package health provides health monitoring for services.
package health

import (
	"time"
)

const (
	// defaultProbeInterval is the default time between probe executions.
	defaultProbeInterval time.Duration = 10 * time.Second
	// defaultProbeTimeout is the default timeout for each probe execution.
	defaultProbeTimeout time.Duration = 5 * time.Second
	// defaultProbeSuccessThreshold is the default number of consecutive successes to mark healthy.
	defaultProbeSuccessThreshold int = 1
	// defaultProbeFailureThreshold is the default number of consecutive failures to mark unhealthy.
	defaultProbeFailureThreshold int = 3
)

// ProbeConfig defines the timing and thresholds for a probe.
// It controls how frequently probes run and when to transition between healthy/unhealthy states.
type ProbeConfig struct {
	// Interval between probe executions.
	Interval time.Duration
	// Timeout for each probe execution.
	Timeout time.Duration
	// SuccessThreshold is the number of consecutive successes to mark healthy.
	SuccessThreshold int
	// FailureThreshold is the number of consecutive failures to mark unhealthy.
	FailureThreshold int
}

// DefaultProbeConfig returns a ProbeConfig with sensible defaults.
//
// Returns:
//   - ProbeConfig: configuration with default interval (10s), timeout (5s), success threshold (1), and failure threshold (3).
func DefaultProbeConfig() ProbeConfig {
	// construct config with default values
	return ProbeConfig{
		Interval:         defaultProbeInterval,
		Timeout:          defaultProbeTimeout,
		SuccessThreshold: defaultProbeSuccessThreshold,
		FailureThreshold: defaultProbeFailureThreshold,
	}
}
