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

// ICMPMode defines how ICMP probes should operate.
// It controls whether to use native ICMP packets or TCP fallback.
type ICMPMode string

// ICMP mode constants.
const (
	// ICMPModeNative uses real ICMP echo requests.
	// Requires CAP_NET_RAW capability on Linux or root privileges.
	ICMPModeNative ICMPMode = "native"

	// ICMPModeFallback uses TCP connection probes instead of ICMP.
	// Works without special privileges, suitable for containers.
	ICMPModeFallback ICMPMode = "fallback"

	// ICMPModeAuto automatically detects capability and uses native if available.
	// Falls back to TCP if ICMP socket creation fails.
	ICMPModeAuto ICMPMode = "auto"
)

// Default HTTP method for probe requests.
const defaultHTTPMethod string = "GET"

// ProbeConfig defines the configuration for probing a listener.
// It specifies timing, thresholds, and protocol-specific settings for health probes.
type ProbeConfig struct {
	// Type specifies the probe type.
	// Supported values: "tcp", "udp", "http", "grpc", "exec", "icmp".
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

	// ICMPMode specifies how ICMP probes should operate.
	// Valid values: "native", "fallback", "auto".
	// Default is "auto" for automatic capability detection.
	ICMPMode ICMPMode
}

// NewProbeConfig creates a new probe configuration with the specified type.
//
// Params:
//   - probeType: the type of probe (tcp, udp, http, grpc, exec, icmp).
//
// Returns:
//   - ProbeConfig: a new probe configuration with defaults applied.
func NewProbeConfig(probeType string) ProbeConfig {
	// create probe with default intervals and thresholds
	return ProbeConfig{
		Type:             probeType,
		Interval:         shared.Seconds(defaultProbeInterval),
		Timeout:          shared.Seconds(defaultProbeTimeout),
		SuccessThreshold: defaultProbeSuccessThreshold,
		FailureThreshold: defaultProbeFailureThreshold,
		Method:           defaultHTTPMethod,
		StatusCode:       defaultHTTPStatusOK,
		ICMPMode:         ICMPModeAuto, // auto-detect capability
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
	// delegate to constructor
	return NewProbeConfig(probeType)
}
