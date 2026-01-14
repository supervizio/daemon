// Package healthcheck provides infrastructure adapters for service probing.
package healthcheck

import (
	"errors"
	"time"

	"github.com/kodflow/daemon/internal/domain/healthcheck"
)

// proberConstructor is a function that creates a prober with a given timeout.
type proberConstructor func(timeout time.Duration) healthcheck.Prober

var (
	// ErrUnknownProberType indicates an unknown prober type was requested.
	ErrUnknownProberType error = errors.New("unknown prober type")

	// proberConstructors maps prober types to their constructor functions.
	proberConstructors map[string]proberConstructor = map[string]proberConstructor{
		proberTypeTCP:  func(t time.Duration) healthcheck.Prober { return NewTCPProber(t) },
		proberTypeUDP:  func(t time.Duration) healthcheck.Prober { return NewUDPProber(t) },
		proberTypeHTTP: func(t time.Duration) healthcheck.Prober { return NewHTTPProber(t) },
		proberTypeGRPC: func(t time.Duration) healthcheck.Prober { return NewGRPCProber(t) },
		proberTypeExec: func(t time.Duration) healthcheck.Prober { return NewExecProber(t) },
		proberTypeICMP: func(t time.Duration) healthcheck.Prober { return NewICMPProber(t) },
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
//   - healthcheck.Prober: the created prober.
//   - error: ErrUnknownProberType if the type is not recognized.
func (f *Factory) Create(proberType string, timeout time.Duration) (healthcheck.Prober, error) {
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
	// Return TCP prober with normalized timeout.
	return NewTCPProber(f.normalizeTimeout(timeout))
}

// CreateUDP creates a UDP prober.
//
// Params:
//   - timeout: the timeout for the prober (uses default if zero).
//
// Returns:
//   - *UDPProber: the created UDP prober.
func (f *Factory) CreateUDP(timeout time.Duration) *UDPProber {
	// Return UDP prober with normalized timeout.
	return NewUDPProber(f.normalizeTimeout(timeout))
}

// CreateHTTP creates an HTTP prober.
//
// Params:
//   - timeout: the timeout for the prober (uses default if zero).
//
// Returns:
//   - *HTTPProber: the created HTTP prober.
func (f *Factory) CreateHTTP(timeout time.Duration) *HTTPProber {
	// Return HTTP prober with normalized timeout.
	return NewHTTPProber(f.normalizeTimeout(timeout))
}

// CreateGRPC creates a gRPC prober.
//
// Params:
//   - timeout: the timeout for the prober (uses default if zero).
//
// Returns:
//   - *GRPCProber: the created gRPC prober.
func (f *Factory) CreateGRPC(timeout time.Duration) *GRPCProber {
	// Return gRPC prober with normalized timeout.
	return NewGRPCProber(f.normalizeTimeout(timeout))
}

// CreateExec creates an exec prober.
//
// Params:
//   - timeout: the timeout for the prober (uses default if zero).
//
// Returns:
//   - *ExecProber: the created exec prober.
func (f *Factory) CreateExec(timeout time.Duration) *ExecProber {
	// Return exec prober with normalized timeout.
	return NewExecProber(f.normalizeTimeout(timeout))
}

// CreateICMP creates an ICMP prober.
//
// Params:
//   - timeout: the timeout for the prober (uses default if zero).
//
// Returns:
//   - *ICMPProber: the created ICMP prober.
func (f *Factory) CreateICMP(timeout time.Duration) *ICMPProber {
	// Return ICMP prober with normalized timeout.
	return NewICMPProber(f.normalizeTimeout(timeout))
}
