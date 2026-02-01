// Package target provides domain entities for external monitoring targets.
// External targets are services, containers, or hosts that supervizio
// monitors but does not manage (no lifecycle control).
package target

import "time"

// Shared constants used across the target package.
const (
	// defaultMapCapacity is the default capacity hint for label maps.
	defaultMapCapacity int = 4

	// defaultInterval is the default time between consecutive probes.
	defaultInterval time.Duration = 30 * time.Second

	// defaultTimeout is the default maximum time to wait for a probe response.
	defaultTimeout time.Duration = 5 * time.Second

	// defaultSuccessThreshold is the default consecutive successes to mark healthy.
	defaultSuccessThreshold int = 1

	// defaultFailureThreshold is the default consecutive failures to mark unhealthy.
	defaultFailureThreshold int = 3
)
