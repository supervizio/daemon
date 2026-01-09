// Package probe provides infrastructure adapters for service probing.
package probe

import (
	"errors"
	"time"

	"github.com/kodflow/daemon/internal/domain/probe"
)

// ErrUnknownProberType indicates an unknown prober type was requested.
var ErrUnknownProberType error = errors.New("unknown prober type")

// Factory creates probers based on type.
// It provides a centralized way to create prober instances.
type Factory struct {
	// defaultTimeout is the default timeout for probers.
	defaultTimeout time.Duration
}

// NewFactory creates a new prober factory.
//
// Params:
//   - defaultTimeout: the default timeout for created probers.
//
// Returns:
//   - *Factory: a factory for creating probers.
func NewFactory(defaultTimeout time.Duration) *Factory {
	// Return configured factory.
	return &Factory{
		defaultTimeout: defaultTimeout,
	}
}

// Create creates a prober of the specified type.
//
// Params:
//   - proberType: the type of prober to create (tcp, udp, http, grpc, exec, icmp).
//   - timeout: the timeout for the prober (uses default if zero).
//
// Returns:
//   - probe.Prober: the created prober.
//   - error: ErrUnknownProberType if the type is not recognized.
func (f *Factory) Create(proberType string, timeout time.Duration) (probe.Prober, error) {
	// Use default timeout if not specified.
	if timeout == 0 {
		timeout = f.defaultTimeout
	}

	// Create prober based on type.
	switch proberType {
	// TCP prober for TCP connectivity checks.
	case proberTypeTCP:
		// Return configured TCP prober.
		return NewTCPProber(timeout), nil
	// UDP prober for UDP packet probes.
	case proberTypeUDP:
		// Return configured UDP prober.
		return NewUDPProber(timeout), nil
	// HTTP prober for HTTP endpoint checks.
	case proberTypeHTTP:
		// Return configured HTTP prober.
		return NewHTTPProber(timeout), nil
	// gRPC prober for gRPC health checks.
	case proberTypeGRPC:
		// Return configured gRPC prober.
		return NewGRPCProber(timeout), nil
	// Exec prober for command execution probes.
	case proberTypeExec:
		// Return configured exec prober.
		return NewExecProber(timeout), nil
	// ICMP prober for ping/latency probes.
	case proberTypeICMP:
		// Return configured ICMP prober.
		return NewICMPProber(timeout), nil
	// Unknown prober type requested.
	default:
		// Return error for unrecognized prober type.
		return nil, ErrUnknownProberType
	}
}

// CreateTCP creates a TCP prober.
//
// Params:
//   - timeout: the timeout for the prober (uses default if zero).
//
// Returns:
//   - *TCPProber: the created TCP prober.
func (f *Factory) CreateTCP(timeout time.Duration) *TCPProber {
	// Use default timeout if not specified.
	if timeout == 0 {
		timeout = f.defaultTimeout
	}

	// Return TCP prober configured with timeout.
	return NewTCPProber(timeout)
}

// CreateUDP creates a UDP prober.
//
// Params:
//   - timeout: the timeout for the prober (uses default if zero).
//
// Returns:
//   - *UDPProber: the created UDP prober.
func (f *Factory) CreateUDP(timeout time.Duration) *UDPProber {
	// Use default timeout if not specified.
	if timeout == 0 {
		timeout = f.defaultTimeout
	}

	// Return UDP prober configured with timeout.
	return NewUDPProber(timeout)
}

// CreateHTTP creates an HTTP prober.
//
// Params:
//   - timeout: the timeout for the prober (uses default if zero).
//
// Returns:
//   - *HTTPProber: the created HTTP prober.
func (f *Factory) CreateHTTP(timeout time.Duration) *HTTPProber {
	// Use default timeout if not specified.
	if timeout == 0 {
		timeout = f.defaultTimeout
	}

	// Return HTTP prober configured with timeout.
	return NewHTTPProber(timeout)
}

// CreateGRPC creates a gRPC prober.
//
// Params:
//   - timeout: the timeout for the prober (uses default if zero).
//
// Returns:
//   - *GRPCProber: the created gRPC prober.
func (f *Factory) CreateGRPC(timeout time.Duration) *GRPCProber {
	// Use default timeout if not specified.
	if timeout == 0 {
		timeout = f.defaultTimeout
	}

	// Return gRPC prober configured with timeout.
	return NewGRPCProber(timeout)
}

// CreateExec creates an exec prober.
//
// Params:
//   - timeout: the timeout for the prober (uses default if zero).
//
// Returns:
//   - *ExecProber: the created exec prober.
func (f *Factory) CreateExec(timeout time.Duration) *ExecProber {
	// Use default timeout if not specified.
	if timeout == 0 {
		timeout = f.defaultTimeout
	}

	// Return exec prober configured with timeout.
	return NewExecProber(timeout)
}

// CreateICMP creates an ICMP prober.
//
// Params:
//   - timeout: the timeout for the prober (uses default if zero).
//
// Returns:
//   - *ICMPProber: the created ICMP prober.
func (f *Factory) CreateICMP(timeout time.Duration) *ICMPProber {
	// Use default timeout if not specified.
	if timeout == 0 {
		timeout = f.defaultTimeout
	}

	// Return ICMP prober configured with timeout.
	return NewICMPProber(timeout)
}
