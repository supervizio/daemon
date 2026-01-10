// Package probe provides infrastructure adapters for service probing.
// It implements the probe.Prober interface for different protocols.
package probe

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/kodflow/daemon/internal/domain/probe"
)

// proberTypeTCP is the type identifier for TCP probers.
const proberTypeTCP string = "tcp"

// TCPProber performs TCP connection probes.
// It verifies service availability by establishing TCP connections.
type TCPProber struct {
	// timeout is the maximum duration for connection attempts.
	timeout time.Duration
}

// NewTCPProber creates a new TCP prober.
//
// Params:
//   - timeout: the maximum duration for connection attempts.
//
// Returns:
//   - *TCPProber: a configured TCP prober ready to perform probes.
func NewTCPProber(timeout time.Duration) *TCPProber {
	// Return configured TCP prober.
	return &TCPProber{
		timeout: timeout,
	}
}

// Type returns the prober type.
//
// Returns:
//   - string: the constant "tcp" identifying the prober type.
func (p *TCPProber) Type() string {
	// Return the TCP prober type identifier.
	return proberTypeTCP
}

// Probe performs a TCP connection probe.
// It attempts to establish a TCP connection to verify the target is accepting connections.
//
// Params:
//   - ctx: context for cancellation and timeout control.
//   - target: the target to probe.
//
// Returns:
//   - probe.Result: the probe result with latency and connection status.
func (p *TCPProber) Probe(ctx context.Context, target probe.Target) probe.Result {
	start := time.Now()

	// Determine the network type.
	network := target.Network
	// Check if network type was specified.
	if network == "" {
		// Default to TCP if not specified.
		network = "tcp"
	}

	// Create dialer with configured timeout.
	dialer := &net.Dialer{
		Timeout: p.timeout,
	}

	// Attempt to establish connection.
	conn, err := dialer.DialContext(ctx, network, target.Address)
	latency := time.Since(start)

	// Handle connection failure.
	if err != nil {
		// Return failure result with connection error.
		return probe.NewFailureResult(
			latency,
			fmt.Sprintf("connection failed: %v", err),
			err,
		)
	}
	// Close the connection.
	_ = conn.Close()

	// Return success result with connection details.
	return probe.NewSuccessResult(
		latency,
		fmt.Sprintf("connected to %s", target.Address),
	)
}
