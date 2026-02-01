// Package config provides domain value objects for service configuration.
package config

// DockerDiscoveryConfig configures Docker container discovery.
// Docker discovery monitors containers running on the local Docker daemon.
type DockerDiscoveryConfig struct {
	// Enabled activates Docker discovery.
	Enabled bool

	// SocketPath is the Docker socket path.
	// Default: "/var/run/docker.sock".
	SocketPath string

	// Labels filters containers by label.
	// Example: {"supervizio.monitor": "true"}.
	Labels map[string]string
}
