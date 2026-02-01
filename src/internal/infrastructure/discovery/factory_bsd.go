//go:build freebsd || openbsd || netbsd

// Package discovery provides infrastructure adapters for target discovery.
package discovery

import "github.com/kodflow/daemon/internal/domain/target"

// createSystemdDiscoverer creates a systemd discoverer.
// On BSD systems, systemd is not available.
//
// Returns:
//   - target.Discoverer: always nil on BSD.
func (f *Factory) createSystemdDiscoverer() target.Discoverer {
	// Systemd is not available on BSD systems.
	// This stub exists for BSD systems where systemd is not used.
	return nil
}

// createOpenRCDiscoverer creates an OpenRC discoverer.
// On BSD systems, OpenRC is not available.
//
// Returns:
//   - target.Discoverer: always nil on BSD.
func (f *Factory) createOpenRCDiscoverer() target.Discoverer {
	// OpenRC is not available on BSD systems.
	// This stub exists for BSD systems where OpenRC is not used.
	return nil
}

// createBSDRCDiscoverer creates a BSD rc.d discoverer.
// It reads patterns from configuration and creates a discoverer for BSD rc.d services.
//
// Returns:
//   - target.Discoverer: the BSD rc.d discoverer or nil if creation fails.
func (f *Factory) createBSDRCDiscoverer() target.Discoverer {
	// Return nil when BSD rc.d config is missing.
	if f.config.BSDRC == nil {
		// Return nil discoverer for missing configuration.
		return nil
	}

	// Return nil when BSD rc.d discovery is disabled.
	if !f.config.BSDRC.Enabled {
		// Return nil discoverer when disabled.
		return nil
	}

	patterns := f.config.BSDRC.Patterns
	// Use nil slice for better memory efficiency when no patterns.
	if patterns == nil {
		patterns = nil
	}

	// Create BSD rc.d discoverer with configured patterns.
	return NewBSDRCDiscoverer(patterns)
}

// createDockerDiscoverer creates a Docker discoverer.
// Docker is available on BSD systems.
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
		socketPath = "/var/run/docker.sock"
	}

	labels := f.config.Docker.Labels
	// Allocate empty map when no labels configured.
	if labels == nil {
		labels = make(map[string]string, 0)
	}

	// Create Docker discoverer with socket path and label filter.
	return NewDockerDiscoverer(socketPath, labels)
}

// createPodmanDiscoverer creates a Podman discoverer.
// TODO: Implement Podman support for BSD.
//
// Returns:
//   - target.Discoverer: always nil (not yet implemented on BSD).
func (f *Factory) createPodmanDiscoverer() target.Discoverer {
	// Podman support on BSD is not yet implemented.
	return nil
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

// createPortScanDiscoverer creates a port scan discoverer.
// TODO: Implement port scanning discovery.
//
// Returns:
//   - target.Discoverer: always nil (not yet implemented).
func (f *Factory) createPortScanDiscoverer() target.Discoverer {
	// Port scan discoverer not yet implemented.
	return nil
}
