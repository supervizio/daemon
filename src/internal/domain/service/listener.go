// Package service provides domain value objects for service configuration.
package service

import "github.com/kodflow/daemon/internal/domain/shared"

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

	// Probe contains the probe configuration for this listener.
	// If nil, no probing is performed (only port listening is checked).
	Probe *ProbeConfig
}

// ProbeConfig defines the configuration for probing a listener.
type ProbeConfig struct {
	// Type specifies the probe type.
	// Supported values: "tcp", "udp", "http", "grpc", "exec".
	Type string

	// Interval specifies the time between consecutive probes.
	Interval shared.Duration

	// Timeout specifies the maximum time to wait for a probe response.
	Timeout shared.Duration

	// SuccessThreshold specifies consecutive successes to mark ready.
	SuccessThreshold int

	// FailureThreshold specifies consecutive failures to mark not ready.
	FailureThreshold int

	// Path specifies the HTTP endpoint path for HTTP probes.
	// Example: "/health", "/ready".
	Path string

	// Method specifies the HTTP method for HTTP probes.
	// Example: "GET", "HEAD".
	Method string

	// StatusCode specifies the expected HTTP status code.
	// Default is 200 if not specified.
	StatusCode int

	// Service specifies the gRPC service name for gRPC probes.
	// Empty string means check server overall health.
	Service string

	// Command specifies the command for exec probes.
	Command string

	// Args specifies the command arguments for exec probes.
	Args []string
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
	// Return listener config with TCP protocol.
	return ListenerConfig{
		Name:     name,
		Port:     port,
		Protocol: "tcp",
	}
}

// WithProbe adds probe configuration to the listener.
//
// Params:
//   - probe: the probe configuration.
//
// Returns:
//   - ListenerConfig: listener with probe configuration.
func (l ListenerConfig) WithProbe(probe ProbeConfig) ListenerConfig {
	// Add probe to listener.
	l.Probe = &probe
	return l
}

// WithTCPProbe adds a TCP probe configuration.
//
// Returns:
//   - ListenerConfig: listener with TCP probe.
func (l ListenerConfig) WithTCPProbe() ListenerConfig {
	// Return listener with TCP probe.
	return l.WithProbe(DefaultProbeConfig("tcp"))
}

// WithHTTPProbe adds an HTTP probe configuration.
//
// Params:
//   - path: the HTTP endpoint path.
//
// Returns:
//   - ListenerConfig: listener with HTTP probe.
func (l ListenerConfig) WithHTTPProbe(path string) ListenerConfig {
	// Create HTTP probe config.
	probe := DefaultProbeConfig("http")
	probe.Path = path
	// Return listener with HTTP probe.
	return l.WithProbe(probe)
}

// WithGRPCProbe adds a gRPC probe configuration.
//
// Params:
//   - service: the gRPC service name to check.
//
// Returns:
//   - ListenerConfig: listener with gRPC probe.
func (l ListenerConfig) WithGRPCProbe(service string) ListenerConfig {
	// Create gRPC probe config.
	probe := DefaultProbeConfig("grpc")
	probe.Service = service
	// Return listener with gRPC probe.
	return l.WithProbe(probe)
}

// DefaultProbeConfig returns a ProbeConfig with sensible defaults.
//
// Params:
//   - probeType: the type of probe.
//
// Returns:
//   - ProbeConfig: a configuration with default values.
func DefaultProbeConfig(probeType string) ProbeConfig {
	// Return probe config with defaults.
	return ProbeConfig{
		Type:             probeType,
		Interval:         shared.Seconds(defaultProbeInterval),
		Timeout:          shared.Seconds(defaultProbeTimeout),
		SuccessThreshold: defaultProbeSuccessThreshold,
		FailureThreshold: defaultProbeFailureThreshold,
		Method:           "GET",
		StatusCode:       200,
	}
}

// ProbeType constants.
const (
	// ProbeTypeTCP performs TCP connection checks.
	ProbeTypeTCP = "tcp"
	// ProbeTypeUDP performs UDP checks.
	ProbeTypeUDP = "udp"
	// ProbeTypeHTTP performs HTTP endpoint checks.
	ProbeTypeHTTP = "http"
	// ProbeTypeGRPC performs gRPC health checks.
	ProbeTypeGRPC = "grpc"
	// ProbeTypeExec executes a command.
	ProbeTypeExec = "exec"
	// ProbeTypeICMP performs ICMP ping checks.
	ProbeTypeICMP = "icmp"
)
