// Package probe provides infrastructure adapters for service probing.
package probe

import (
	"errors"
	"time"

	"github.com/kodflow/daemon/internal/domain/probe"
)

// proberConstructor is a function that creates a prober with a given timeout.
type proberConstructor func(timeout time.Duration) probe.Prober

var (
	// ErrUnknownProberType indicates an unknown prober type was requested.
	ErrUnknownProberType error = errors.New("unknown prober type")

	// proberConstructors maps prober types to their constructor functions.
	proberConstructors map[string]proberConstructor = map[string]proberConstructor{
		proberTypeTCP:  func(t time.Duration) probe.Prober { return NewTCPProber(t) },
		proberTypeUDP:  func(t time.Duration) probe.Prober { return NewUDPProber(t) },
		proberTypeHTTP: func(t time.Duration) probe.Prober { return NewHTTPProber(t) },
		proberTypeGRPC: func(t time.Duration) probe.Prober { return NewGRPCProber(t) },
		proberTypeExec: func(t time.Duration) probe.Prober { return NewExecProber(t) },
		proberTypeICMP: func(t time.Duration) probe.Prober { return NewICMPProber(t) },
	}
)

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
	// Normalize timeout using helper.
	timeout = f.normalizeTimeout(timeout)

	// Look up constructor in map.
	constructor, exists := proberConstructors[proberType]
	// Check if prober type is recognized.
	if !exists {
		// Return error for unrecognized prober type.
		return nil, ErrUnknownProberType
	}

	// Create prober using constructor.
	return constructor(timeout), nil
}

// normalizeTimeout returns a valid timeout value.
//
// Params:
//   - timeout: the input timeout duration.
//
// Returns:
//   - time.Duration: the input timeout or factory default if zero/negative.
func (f *Factory) normalizeTimeout(timeout time.Duration) time.Duration {
	// Use factory default timeout if not specified or invalid.
	if timeout <= 0 {
		// Return factory default.
		return f.defaultTimeout
	}
	// Return input timeout.
	return timeout
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
