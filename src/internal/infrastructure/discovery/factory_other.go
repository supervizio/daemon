//go:build !unix

// Package discovery provides infrastructure adapters for target discovery.
package discovery

import "github.com/kodflow/daemon/internal/domain/target"

// createSystemdDiscoverer is not available on non-Unix platforms.
//
// Returns:
//   - target.Discoverer: always nil.
func (f *Factory) createSystemdDiscoverer() target.Discoverer {
	// return nil on non-unix
	return nil
}

// createDockerDiscoverer is not available on non-Unix platforms.
//
// Returns:
//   - target.Discoverer: always nil.
func (f *Factory) createDockerDiscoverer() target.Discoverer {
	// return nil on non-unix
	return nil
}

// createKubernetesDiscoverer creates a Kubernetes discoverer.
//
// Returns:
//   - target.Discoverer: always nil (not yet implemented).
func (f *Factory) createKubernetesDiscoverer() target.Discoverer {
	// return nil (not yet implemented)
	return nil
}

// createNomadDiscoverer creates a Nomad discoverer.
//
// Returns:
//   - target.Discoverer: always nil (not yet implemented).
func (f *Factory) createNomadDiscoverer() target.Discoverer {
	// return nil (not yet implemented)
	return nil
}
