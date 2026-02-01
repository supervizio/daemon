//go:build linux

// Package discovery provides infrastructure adapters for target discovery.
package discovery

import "github.com/kodflow/daemon/internal/domain/target"

// defaultDockerSocketPath is the default path to the Docker socket on Linux.
const defaultDockerSocketPath string = "/var/run/docker.sock"

// createSystemdDiscoverer creates a systemd discoverer on Linux.
// It reads patterns from configuration and creates a discoverer for systemd services.
//
// Returns:
//   - target.Discoverer: the systemd discoverer or nil if creation fails.
func (f *Factory) createSystemdDiscoverer() target.Discoverer {
	// Return nil when systemd config is missing.
	if f.config.Systemd == nil {
		// Return nil discoverer for missing configuration.
		return nil
	}

	patterns := f.config.Systemd.Patterns
	// Use nil slice for better memory efficiency when no patterns.
	if patterns == nil {
		patterns = nil
	}

	// Create systemd discoverer with configured patterns.
	return NewSystemdDiscoverer(patterns)
}

// createDockerDiscoverer creates a Docker discoverer.
// It reads socket path and labels from configuration with sensible defaults.
//
// Returns:
//   - target.Discoverer: the Docker discoverer or nil if creation fails.
func (f *Factory) createDockerDiscoverer() target.Discoverer {
	// Return nil when Docker config is missing.
	if f.config.Docker == nil {
		// Return nil discoverer for missing configuration.
		return nil
	}

	socketPath := f.config.Docker.SocketPath
	// Apply default socket path when not configured.
	if socketPath == "" {
		socketPath = defaultDockerSocketPath
	}

	labels := f.config.Docker.Labels
	// Allocate empty map when no labels configured.
	if labels == nil {
		labels = make(map[string]string, 0)
	}

	// Create Docker discoverer with socket path and label filter.
	return NewDockerDiscoverer(socketPath, labels)
}

// createKubernetesDiscoverer creates a Kubernetes discoverer.
// TODO: Implement using k8s client-go.
//
// Returns:
//   - target.Discoverer: the Kubernetes discoverer or nil.
func (f *Factory) createKubernetesDiscoverer() target.Discoverer {
	// Return nil when Kubernetes config is missing.
	if f.config.Kubernetes == nil {
		// Return nil discoverer for missing configuration.
		return nil
	}

	// Kubernetes discoverer requires k8s client-go.
	// For now, return nil. Implementation will be added later.
	return nil
}

// createNomadDiscoverer creates a Nomad discoverer.
// TODO: Implement using Nomad API client.
//
// Returns:
//   - target.Discoverer: the Nomad discoverer or nil.
func (f *Factory) createNomadDiscoverer() target.Discoverer {
	// Return nil when Nomad config is missing.
	if f.config.Nomad == nil {
		// Return nil discoverer for missing configuration.
		return nil
	}

	// Nomad discoverer requires nomad API client.
	// For now, return nil. Implementation will be added later.
	return nil
}
