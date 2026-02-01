// Package config provides domain value objects for service configuration.
package config

import (
	"github.com/kodflow/daemon/internal/domain/shared"
)

// TargetConfig defines a statically configured external target.
// Static targets are manually configured for monitoring external services.
type TargetConfig struct {
	// Name is the unique target name.
	Name string

	// Type specifies the target type.
	// Supported: "remote", "docker", "kubernetes", "nomad", "custom".
	Type string

	// Address is the target address for remote targets.
	// Format: "host:port" for TCP/UDP, "host" for ICMP.
	Address string

	// Container is the container ID or name for docker/podman targets.
	Container string

	// Namespace is the namespace for kubernetes targets.
	Namespace string

	// Service is the service name for kubernetes targets.
	Service string

	// Probe configures the health probe for this target.
	Probe ProbeConfig

	// Interval overrides the default probe interval.
	Interval shared.Duration

	// Timeout overrides the default probe timeout.
	Timeout shared.Duration

	// Labels are metadata labels for filtering and grouping.
	Labels map[string]string
}
