//go:build unix && !linux && !freebsd && !openbsd && !netbsd

// Package discovery provides infrastructure adapters for target discovery.
package discovery

import "github.com/kodflow/daemon/internal/domain/target"

// defaultDockerSocketPath is the default path to the Docker socket on Unix.
const defaultDockerSocketPath string = "/var/run/docker.sock"

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

// createOpenRCDiscoverer creates an OpenRC discoverer on Unix systems.
// On non-Linux Unix systems, returns nil (OpenRC is Linux-only).
//
// Returns:
//   - target.Discoverer: the OpenRC discoverer or nil if not available.
func (f *Factory) createOpenRCDiscoverer() target.Discoverer {
	// OpenRC is only available on Linux.
	// This stub exists for non-Linux Unix systems (BSD, macOS).
	return nil
}

// createBSDRCDiscoverer creates a BSD rc.d discoverer on Unix systems.
// This is handled in factory_bsd.go for BSD systems, returns nil on macOS.
//
// Returns:
//   - target.Discoverer: the BSD rc.d discoverer or nil.
func (f *Factory) createBSDRCDiscoverer() target.Discoverer {
	// BSD rc.d is handled by factory_bsd.go for FreeBSD/OpenBSD/NetBSD.
	// On macOS and other Unix systems, return nil.
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
		socketPath = defaultDockerSocketPath
	}

	labels := f.config.Docker.Labels
	// check if labels need initialization
	if labels == nil {
		labels = make(map[string]string)
	}

	// return docker discoverer
	return NewDockerDiscoverer(socketPath, labels)
}

// createPodmanDiscoverer creates a Podman discoverer on Unix systems.
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

// createPortScanDiscoverer creates a port scan discoverer on Unix systems.
// Port scanning requires /proc filesystem which is Linux-specific.
//
// Returns:
//   - target.Discoverer: always nil on non-Linux Unix systems.
func (f *Factory) createPortScanDiscoverer() target.Discoverer {
	// Port scanning requires /proc filesystem for network socket information.
	// This is Linux-specific and not available on other Unix systems.
	return nil
}

// createOpenRCDiscoverer creates an OpenRC discoverer.
// On non-Linux Unix systems, OpenRC is not typically available.
//
// Returns:
//   - target.Discoverer: always nil.
func (f *Factory) createOpenRCDiscoverer() target.Discoverer {
	// OpenRC is not typically available on non-Linux Unix systems.
	// This stub exists for macOS and other Unix systems.
	return nil
}

// createBSDRCDiscoverer creates a BSD rc.d discoverer.
// On non-BSD Unix systems (like macOS), BSD rc.d is not available.
//
// Returns:
//   - target.Discoverer: always nil.
func (f *Factory) createBSDRCDiscoverer() target.Discoverer {
	// BSD rc.d is not available on non-BSD Unix systems.
	// This stub exists for macOS and other non-BSD Unix systems.
	return nil
}

// createPortScanDiscoverer creates a port scan discoverer.
// On Unix systems, port scanning is not yet implemented.
//
// Returns:
//   - target.Discoverer: always nil (not yet implemented).
func (f *Factory) createPortScanDiscoverer() target.Discoverer {
	// Port scan discoverer not yet implemented.
	// This stub exists for Unix systems.
	return nil
}
