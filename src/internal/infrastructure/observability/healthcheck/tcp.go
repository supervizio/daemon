// Package healthcheck provides infrastructure adapters for service probing.
// It implements the healthcheck.Prober interface for different protocols.
package healthcheck

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/kodflow/daemon/internal/domain/health"
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
	return &TCPProber{
		timeout: timeout,
	}
}

// Type returns the prober type.
//
// Returns:
//   - string: the constant "tcp" identifying the prober type.
func (p *TCPProber) Type() string {
	return proberTypeTCP
}

// Probe performs a TCP connection healthcheck.
// It attempts to establish a TCP connection to verify the target is accepting connections.
//
// Params:
//   - ctx: context for cancellation and timeout control.
//   - target: the target to healthcheck.
//
// Returns:
//   - health.CheckResult: the probe result with latency and connection status.
func (p *TCPProber) Probe(ctx context.Context, target health.Target) health.CheckResult {
	start := time.Now()

	network := target.Network
	if network == "" {
		network = "tcp"
	}

	dialer := &net.Dialer{
		Timeout: p.timeout,
	}

	conn, err := dialer.DialContext(ctx, network, target.Address)
	latency := time.Since(start)

	if err != nil {
		return health.NewFailureCheckResult(
			latency,
			fmt.Sprintf("connection failed: %v", err),
			err,
		)
	}
	_ = conn.Close()

	return health.NewSuccessCheckResult(
		latency,
		fmt.Sprintf("connected to %s", target.Address),
	)
}
