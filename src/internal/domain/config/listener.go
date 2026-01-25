// Package config provides domain value objects for service configuration.
package config

// Default listener probe configuration values.
const (
	// defaultProbeInterval is the default interval between probes (10 seconds).
	defaultProbeInterval int = 10
	// defaultProbeTimeout is the default timeout for probe responses (5 seconds).
	defaultProbeTimeout int = 5
	// defaultProbeSuccessThreshold is the default number of successes to mark ready.
	defaultProbeSuccessThreshold int = 1
	// defaultProbeFailureThreshold is the default number of failures to mark not ready.
	defaultProbeFailureThreshold int = 3
)

// ListenerConfig defines a network listener with optional health probing.
// It specifies the port, protocol, and probe configuration for a listener.
type ListenerConfig struct {
	// Name is the unique identifier for this listener.
	// Examples: "http", "admin", "grpc".
	Name string

	// Port is the port number the service listens on.
	Port int

	// Protocol is the network protocol.
	// Supported values: "tcp" (default), "udp".
	Protocol string

	// Address is the optional bind address.
	// Empty means bind to all interfaces (0.0.0.0).
	Address string

	// Exposed indicates whether this port should be publicly accessible.
	// Used for status display:
	//   - Green: port state matches exposed setting
	//   - Yellow: mismatch (exposed but unreachable, or not exposed but reachable)
	//   - Red: expected port but nothing listening
	Exposed bool

	// Probe contains the probe configuration for this listener.
	// If nil, no probing is performed (only port listening is checked).
	Probe *ProbeConfig
}

// NewListenerConfig creates a new listener configuration.
//
// Params:
//   - name: unique identifier for the listener.
//   - port: port number the service listens on.
//
// Returns:
//   - ListenerConfig: listener configuration with TCP protocol.
func NewListenerConfig(name string, port int) ListenerConfig {
	// Return listener config with TCP protocol as the default.
	return ListenerConfig{
		Name:     name,
		Port:     port,
		Protocol: "tcp",
	}
}

// WithProbe adds probe configuration to the listener.
//
// Params:
//   - probe: the probe configuration pointer to avoid large struct copy.
//
// Returns:
//   - ListenerConfig: listener with probe configuration.
func (l ListenerConfig) WithProbe(probe *ProbeConfig) ListenerConfig {
	// Attach probe to listener for health checking.
	l.Probe = probe
	// Return the modified listener with probe attached.
	return l
}

// WithTCPProbe adds a TCP probe configuration.
//
// Returns:
//   - ListenerConfig: listener with TCP probe.
func (l ListenerConfig) WithTCPProbe() ListenerConfig {
	// Create TCP probe and attach to listener.
	probe := DefaultProbeConfig("tcp")
	// Return listener with TCP probe for connection checks.
	return l.WithProbe(&probe)
}

// WithHTTPProbe adds an HTTP probe configuration.
//
// Params:
//   - path: the HTTP endpoint path.
//
// Returns:
//   - ListenerConfig: listener with HTTP probe.
func (l ListenerConfig) WithHTTPProbe(path string) ListenerConfig {
	// Create HTTP probe config with the specified path.
	probe := DefaultProbeConfig("http")
	probe.Path = path
	// Return listener with HTTP probe for endpoint checks.
	return l.WithProbe(&probe)
}

// WithGRPCProbe adds a gRPC probe configuration.
//
// Params:
//   - service: the gRPC service name to check.
//
// Returns:
//   - ListenerConfig: listener with gRPC probe.
func (l ListenerConfig) WithGRPCProbe(service string) ListenerConfig {
	// Create gRPC probe config with the specified service name.
	probe := DefaultProbeConfig("grpc")
	probe.Service = service
	// Return listener with gRPC probe for health checks.
	return l.WithProbe(&probe)
}
