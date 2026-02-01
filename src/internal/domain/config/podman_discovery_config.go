// Package config provides domain value objects for service configuration.
package config

// PodmanDiscoveryConfig configures Podman container discovery.
// Podman discovery monitors containers running on the local Podman daemon.
type PodmanDiscoveryConfig struct {
	// Enabled activates Podman discovery.
	Enabled bool

	// SocketPath is the Podman socket path.
	// Default: "/run/podman/podman.sock".
	SocketPath string

	// Labels filters containers by label.
	Labels map[string]string
}
