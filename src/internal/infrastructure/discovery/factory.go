// Package discovery provides infrastructure adapters for target discovery.
// It implements the domain/target.Discoverer interface for various platforms.
package discovery

import (
	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/target"
)

// maxDiscovererTypes is the maximum number of discoverer types.
// Used for pre-allocating the discoverers slice to avoid reallocations.
const maxDiscovererTypes int = 4

// Factory creates discoverers based on configuration.
// It provides a unified way to instantiate platform-specific discoverers.
type Factory struct {
	// config contains the discovery configuration.
	config *config.DiscoveryConfig
}

// NewFactory creates a new discoverer factory.
//
// Params:
//   - cfg: the discovery configuration.
//
// Returns:
//   - *Factory: a new factory instance.
func NewFactory(cfg *config.DiscoveryConfig) *Factory {
	// Construct factory with discovery configuration.
	return &Factory{
		config: cfg,
	}
}

// CreateDiscoverers creates all enabled discoverers from configuration.
// It iterates through discoverer types and adds enabled instances.
//
// Returns:
//   - []target.Discoverer: slice of enabled discoverers.
func (f *Factory) CreateDiscoverers() []target.Discoverer {
	// Check if configuration is provided.
	if f.config == nil {
		// Return nil for missing configuration.
		return nil
	}

	// Pre-allocate for common case of 1-2 discoverers.
	discoverers := make([]target.Discoverer, 0, maxDiscovererTypes)

	// Add each discoverer type if enabled in configuration.
	discoverers = f.addSystemdDiscoverer(discoverers)
	discoverers = f.addDockerDiscoverer(discoverers)
	discoverers = f.addKubernetesDiscoverer(discoverers)
	discoverers = f.addNomadDiscoverer(discoverers)

	// Return all enabled discoverers.
	return discoverers
}

// addSystemdDiscoverer adds systemd discoverer if enabled.
//
// Params:
//   - discoverers: existing discoverer list.
//
// Returns:
//   - []target.Discoverer: updated discoverer list.
func (f *Factory) addSystemdDiscoverer(discoverers []target.Discoverer) []target.Discoverer {
	// Check if systemd discovery is configured and enabled.
	if f.config.Systemd != nil && f.config.Systemd.Enabled {
		// Create systemd discoverer instance.
		if discoverer := f.createSystemdDiscoverer(); discoverer != nil {
			discoverers = append(discoverers, discoverer)
		}
	}
	// Return updated list.
	return discoverers
}

// addDockerDiscoverer adds Docker discoverer if enabled.
//
// Params:
//   - discoverers: existing discoverer list.
//
// Returns:
//   - []target.Discoverer: updated discoverer list.
func (f *Factory) addDockerDiscoverer(discoverers []target.Discoverer) []target.Discoverer {
	// Check if Docker discovery is configured and enabled.
	if f.config.Docker != nil && f.config.Docker.Enabled {
		// Create Docker discoverer instance.
		if discoverer := f.createDockerDiscoverer(); discoverer != nil {
			discoverers = append(discoverers, discoverer)
		}
	}
	// Return updated list.
	return discoverers
}

// addKubernetesDiscoverer adds Kubernetes discoverer if enabled.
//
// Params:
//   - discoverers: existing discoverer list.
//
// Returns:
//   - []target.Discoverer: updated discoverer list.
func (f *Factory) addKubernetesDiscoverer(discoverers []target.Discoverer) []target.Discoverer {
	// Check if Kubernetes discovery is configured and enabled.
	if f.config.Kubernetes != nil && f.config.Kubernetes.Enabled {
		// Create Kubernetes discoverer instance.
		if discoverer := f.createKubernetesDiscoverer(); discoverer != nil {
			discoverers = append(discoverers, discoverer)
		}
	}
	// Return updated list.
	return discoverers
}

// addNomadDiscoverer adds Nomad discoverer if enabled.
//
// Params:
//   - discoverers: existing discoverer list.
//
// Returns:
//   - []target.Discoverer: updated discoverer list.
func (f *Factory) addNomadDiscoverer(discoverers []target.Discoverer) []target.Discoverer {
	// Check if Nomad discovery is configured and enabled.
	if f.config.Nomad != nil && f.config.Nomad.Enabled {
		// Create Nomad discoverer instance.
		if discoverer := f.createNomadDiscoverer(); discoverer != nil {
			discoverers = append(discoverers, discoverer)
		}
	}
	// Return updated list.
	return discoverers
}
