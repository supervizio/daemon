// Package config provides domain value objects for service configuration.
package config

import (
	"github.com/kodflow/daemon/internal/domain/shared"
)

// MonitoringDefaults specifies default values for external targets.
// These values apply to all targets unless overridden per-target.
type MonitoringDefaults struct {
	// Interval is the default time between probes.
	Interval shared.Duration

	// Timeout is the default probe timeout.
	Timeout shared.Duration

	// SuccessThreshold is the default consecutive successes to mark healthy.
	SuccessThreshold int

	// FailureThreshold is the default consecutive failures to mark unhealthy.
	FailureThreshold int
}
