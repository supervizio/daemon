//go:build linux

// Package discovery provides infrastructure adapters for target discovery.
package discovery

import "github.com/kodflow/daemon/internal/domain/target"

// defaultDockerSocketPath is the default path to the Docker socket on Linux.
const defaultDockerSocketPath string = "/var/run/docker.sock"

// defaultPodmanSocketPath is the default path to the Podman socket on Linux.
const defaultPodmanSocketPath string = "/run/podman/podman.sock"

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

// createOpenRCDiscoverer creates an OpenRC discoverer.
// On Linux, OpenRC is not typically available (except Alpine/Gentoo).
// TODO: Implement OpenRC discoverer for Alpine/Gentoo.
//
// Returns:
//   - target.Discoverer: the OpenRC discoverer or nil.
func (f *Factory) createOpenRCDiscoverer() target.Discoverer {
	// OpenRC discoverer not yet implemented on Linux.
	// This will be added for Alpine/Gentoo support.
	return nil
}

// createBSDRCDiscoverer creates a BSD rc.d discoverer.
// On Linux, BSD rc.d is not available.
//
// Returns:
//   - target.Discoverer: always nil on Linux.
func (f *Factory) createBSDRCDiscoverer() target.Discoverer {
	// BSD rc.d is not available on Linux.
	// This stub exists for Linux where BSD rc.d is not used.
	return nil
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

// createPodmanDiscoverer creates a Podman discoverer.
// It reads socket path and labels from configuration with sensible defaults.
//
// Returns:
//   - target.Discoverer: the Podman discoverer or nil if creation fails.
func (f *Factory) createPodmanDiscoverer() target.Discoverer {
	// Return nil when Podman config is missing.
	if f.config.Podman == nil {
		// Return nil discoverer for missing configuration.
		return nil
	}

	socketPath := f.config.Podman.SocketPath
	// Apply default socket path when not configured.
	if socketPath == "" {
		socketPath = defaultPodmanSocketPath
	}

	labels := f.config.Podman.Labels
	// Allocate empty map when no labels configured.
	if labels == nil {
		labels = make(map[string]string, 0)
	}

	// Create Podman discoverer with socket path and label filter.
	return NewPodmanDiscoverer(socketPath, labels)
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
// It reads address, namespace, and job filter from configuration.
//
// Returns:
//   - target.Discoverer: the Nomad discoverer or nil.
func (f *Factory) createNomadDiscoverer() target.Discoverer {
	// Return nil when Nomad config is missing.
	if f.config.Nomad == nil {
		// Return nil discoverer for missing configuration.
		return nil
	}

	// Create Nomad discoverer with configuration.
	return NewNomadDiscoverer(f.config.Nomad)
}

// createPortScanDiscoverer creates a port scan discoverer on Linux.
// Port scan discovery reads /proc/net/tcp to find listening ports.
//
// Returns:
//   - target.Discoverer: the port scan discoverer or nil if creation fails.
func (f *Factory) createPortScanDiscoverer() target.Discoverer {
	// Return nil when PortScan config is missing.
	if f.config.PortScan == nil {
		// Return nil discoverer for missing configuration.
		return nil
	}

	// Create port scan discoverer with configuration.
	return NewPortScanDiscoverer(f.config.PortScan)
}
