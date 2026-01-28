// Package config provides domain value objects for service configuration.
package config

import "github.com/kodflow/daemon/internal/domain/shared"

// ProbeType constants define the supported probe protocols.
// These values are used to specify how listeners should be probed.
const (
	// ProbeTypeTCP performs TCP connection checks.
	ProbeTypeTCP string = "tcp"
	// ProbeTypeUDP performs UDP checks.
	ProbeTypeUDP string = "udp"
	// ProbeTypeHTTP performs HTTP endpoint checks.
	ProbeTypeHTTP string = "http"
	// ProbeTypeGRPC performs gRPC health checks.
	ProbeTypeGRPC string = "grpc"
	// ProbeTypeExec executes a command.
	ProbeTypeExec string = "exec"
	// ProbeTypeICMP performs ICMP ping checks.
	ProbeTypeICMP string = "icmp"
)

// Default HTTP method for probe requests.
const defaultHTTPMethod string = "GET"

// ProbeConfig defines the configuration for probing a listener.
// It specifies timing, thresholds, and protocol-specific settings for health probes.
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

// NewProbeConfig creates a new probe configuration with the specified type.
//
// Params:
//   - probeType: the type of probe (tcp, udp, http, grpc, exec, icmp).
//
// Returns:
//   - ProbeConfig: a new probe configuration with defaults applied.
func NewProbeConfig(probeType string) ProbeConfig {
	return ProbeConfig{
		Type:             probeType,
		Interval:         shared.Seconds(defaultProbeInterval),
		Timeout:          shared.Seconds(defaultProbeTimeout),
		SuccessThreshold: defaultProbeSuccessThreshold,
		FailureThreshold: defaultProbeFailureThreshold,
		Method:           defaultHTTPMethod,
		StatusCode:       defaultHTTPStatusOK,
	}
}

// DefaultProbeConfig returns a ProbeConfig with sensible defaults.
//
// Params:
//   - probeType: the type of probe.
//
// Returns:
//   - ProbeConfig: a configuration with default values.
func DefaultProbeConfig(probeType string) ProbeConfig {
	return NewProbeConfig(probeType)
}
