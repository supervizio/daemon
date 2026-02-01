//go:build unix && !linux

// Package discovery provides infrastructure adapters for target discovery.
package discovery

import "github.com/kodflow/daemon/internal/domain/target"

// createSystemdDiscoverer creates a systemd discoverer on Unix systems.
// On non-Linux Unix systems, returns nil (systemd is Linux-only).
//
// Returns:
//   - target.Discoverer: the systemd discoverer or nil if not available.
func (f *Factory) createSystemdDiscoverer() target.Discoverer {
	// Systemd is only available on Linux.
	// This stub exists for non-Linux Unix systems (BSD, macOS).
	return nil
}

// createDockerDiscoverer creates a Docker discoverer.
//
// Returns:
//   - target.Discoverer: the Docker discoverer or nil if creation fails.
func (f *Factory) createDockerDiscoverer() target.Discoverer {
	// check if docker config exists
	if f.config.Docker == nil {
		// return nil for missing config
		return nil
	}

	socketPath := f.config.Docker.SocketPath
	// check if socket path needs default
	if socketPath == "" {
		socketPath = "/var/run/docker.sock"
	}

	labels := f.config.Docker.Labels
	// check if labels need initialization
	if labels == nil {
		labels = make(map[string]string)
	}

	// return docker discoverer
	return NewDockerDiscoverer(socketPath, labels)
}

// createKubernetesDiscoverer creates a Kubernetes discoverer.
// TODO: Implement using k8s client-go.
//
// Returns:
//   - target.Discoverer: the Kubernetes discoverer or nil.
func (f *Factory) createKubernetesDiscoverer() target.Discoverer {
	// check if kubernetes config exists
	if f.config.Kubernetes == nil {
		// return nil for missing config
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
	// check if nomad config exists
	if f.config.Nomad == nil {
		// return nil for missing config
		return nil
	}

	// Nomad discoverer requires nomad API client.
	// For now, return nil. Implementation will be added later.
	return nil
}
