// Package target provides domain entities for external monitoring targets.
// External targets are services, containers, or hosts that supervizio
// monitors but does not manage (no lifecycle control).
package target

import "context"

// Discoverer is a port interface for discovering external targets.
// Infrastructure adapters implement this interface to provide
// target discovery for specific platforms (systemd, Docker, etc.).
type Discoverer interface {
	// Discover finds all available targets of this type.
	// This is typically called at startup and periodically to refresh.
	//
	// Params:
	//   - ctx: context for cancellation and timeout.
	//
	// Returns:
	//   - []ExternalTarget: the discovered targets.
	//   - error: any error that occurred during discovery.
	Discover(ctx context.Context) ([]ExternalTarget, error)

	// Type returns the target type this discoverer handles.
	//
	// Returns:
	//   - Type: the target type (e.g., TypeSystemd, TypeDocker).
	Type() Type
}

// DiscoveryConfig contains configuration for a specific discoverer.
// It defines filters, patterns, and connection details for discovering
// targets from external platforms like systemd, Docker, or Kubernetes.
type DiscoveryConfig struct {
	// Enabled indicates if this discoverer should run.
	Enabled bool

	// Patterns are filter patterns for target selection.
	// For systemd: unit name patterns (e.g., "nginx.service", "*.target").
	// For Docker: label selectors.
	Patterns []string

	// Labels are required labels for target selection.
	// Only targets with matching labels will be discovered.
	Labels map[string]string

	// Namespaces limits discovery to specific namespaces.
	// For Kubernetes: namespace names.
	// For Nomad: namespace names.
	Namespaces []string

	// LabelSelector is a label selector expression.
	// For Kubernetes: "app=nginx,version=v1".
	LabelSelector string

	// SocketPath overrides the default socket path for container runtimes.
	// For Docker: "/var/run/docker.sock".
	// For Podman: "/run/podman/podman.sock".
	SocketPath string

	// Address is the API address for remote orchestrators.
	// For Nomad: "http://localhost:4646".
	// For Kubernetes: API server address (usually from kubeconfig).
	Address string

	// KubeconfigPath is the path to kubeconfig file.
	// For Kubernetes only.
	KubeconfigPath string
}

// NewDiscoveryConfig creates a new discovery configuration with defaults.
//
// Returns:
//   - DiscoveryConfig: a default configuration with discovery disabled.
func NewDiscoveryConfig() DiscoveryConfig {
	// Return default config with discovery disabled and empty collections.
	return DiscoveryConfig{
		Enabled:    false,
		Patterns:   nil,
		Labels:     make(map[string]string, defaultMapCapacity),
		Namespaces: nil,
	}
}

// WithEnabled enables or disables this discoverer.
//
// Params:
//   - enabled: whether discovery is enabled.
//
// Returns:
//   - DiscoveryConfig: the config for method chaining.
func (c DiscoveryConfig) WithEnabled(enabled bool) DiscoveryConfig {
	c.Enabled = enabled
	// Return modified config to enable fluent API pattern.
	return c
}

// WithPatterns sets the filter patterns.
//
// Params:
//   - patterns: the filter patterns.
//
// Returns:
//   - DiscoveryConfig: the config for method chaining.
func (c DiscoveryConfig) WithPatterns(patterns ...string) DiscoveryConfig {
	c.Patterns = patterns
	// Return modified config to enable fluent API pattern.
	return c
}

// WithNamespaces sets the namespace filter.
//
// Params:
//   - namespaces: the namespace names.
//
// Returns:
//   - DiscoveryConfig: the config for method chaining.
func (c DiscoveryConfig) WithNamespaces(namespaces ...string) DiscoveryConfig {
	c.Namespaces = namespaces
	// Return modified config to enable fluent API pattern.
	return c
}

// WithSocketPath sets the socket path for container runtimes.
//
// Params:
//   - path: the socket path.
//
// Returns:
//   - DiscoveryConfig: the config for method chaining.
func (c DiscoveryConfig) WithSocketPath(path string) DiscoveryConfig {
	c.SocketPath = path
	// Return modified config to enable fluent API pattern.
	return c
}
