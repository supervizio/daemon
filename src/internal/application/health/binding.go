// Package health provides health monitoring for services.
package health

import (
	"time"
)

// ProbeType identifies the type of health probe.
type ProbeType string

const (
	// ProbeTCP is a TCP connection probe.
	ProbeTCP ProbeType = "tcp"
	// ProbeUDP is a UDP probe.
	ProbeUDP ProbeType = "udp"
	// ProbeHTTP is an HTTP request probe.
	ProbeHTTP ProbeType = "http"
	// ProbeGRPC is a gRPC health check probe.
	ProbeGRPC ProbeType = "grpc"
	// ProbeExec is a command execution probe.
	ProbeExec ProbeType = "exec"
	// ProbeICMP is an ICMP ping probe.
	ProbeICMP ProbeType = "icmp"
)

// ProbeTarget defines the target for a health probe.
type ProbeTarget struct {
	// Address is the target address (host:port).
	Address string
	// Path is the HTTP path (for HTTP probes).
	Path string
	// Service is the gRPC service name (for gRPC probes).
	Service string
	// Method is the HTTP method (for HTTP probes).
	Method string
	// StatusCode is the expected HTTP status code.
	StatusCode int
	// Command is the command to execute (for exec probes).
	Command string
	// Args are the command arguments (for exec probes).
	Args []string
}

// ProbeConfig defines the timing and thresholds for a probe.
type ProbeConfig struct {
	// Interval between probe executions.
	Interval time.Duration
	// Timeout for each probe execution.
	Timeout time.Duration
	// SuccessThreshold is the number of consecutive successes to mark healthy.
	SuccessThreshold int
	// FailureThreshold is the number of consecutive failures to mark unhealthy.
	FailureThreshold int
}

// DefaultProbeConfig returns a ProbeConfig with sensible defaults.
func DefaultProbeConfig() ProbeConfig {
	return ProbeConfig{
		Interval:         10 * time.Second,
		Timeout:          5 * time.Second,
		SuccessThreshold: 1,
		FailureThreshold: 3,
	}
}

// ProbeBinding associates a listener with its probe configuration.
// This is the application-level wiring between listeners and health checks.
type ProbeBinding struct {
	// ListenerName is the name of the listener to probe.
	ListenerName string
	// Type is the probe type.
	Type ProbeType
	// Target is the probe target configuration.
	Target ProbeTarget
	// Config is the probe timing configuration.
	Config ProbeConfig
}

// NewProbeBinding creates a new probe binding.
func NewProbeBinding(listenerName string, probeType ProbeType, target ProbeTarget) *ProbeBinding {
	return &ProbeBinding{
		ListenerName: listenerName,
		Type:         probeType,
		Target:       target,
		Config:       DefaultProbeConfig(),
	}
}

// WithConfig sets a custom probe configuration.
func (b *ProbeBinding) WithConfig(config ProbeConfig) *ProbeBinding {
	b.Config = config
	return b
}
