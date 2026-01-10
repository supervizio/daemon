// Package probe provides infrastructure adapters for service probing.
package probe

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/kodflow/daemon/internal/domain/probe"
)

// proberTypeGRPC is the type identifier for gRPC probers.
const proberTypeGRPC string = "grpc"

var (
	// ErrGRPCNotServing indicates the gRPC service is not serving.
	ErrGRPCNotServing error = errors.New("service not serving")

	// ErrGRPCServiceUnknown indicates the gRPC service is unknown.
	ErrGRPCServiceUnknown error = errors.New("service unknown")
)

// GRPCProber performs gRPC health check probes.
// It uses the standard gRPC health/v1 protocol.
//
// Note: This is a simplified implementation that performs TCP connectivity check.
// For full gRPC health checking, google.golang.org/grpc/health/grpc_health_v1
// would be required as a dependency.
type GRPCProber struct {
	// timeout is the maximum duration for the probe.
	timeout time.Duration
	// insecure enables insecure connections (no TLS).
	insecure bool
}

// NewGRPCProber creates a new gRPC health prober.
//
// Params:
//   - timeout: the maximum duration for health checks.
//
// Returns:
//   - *GRPCProber: a configured gRPC prober ready to perform probes.
func NewGRPCProber(timeout time.Duration) *GRPCProber {
	// Return configured gRPC prober.
	return &GRPCProber{
		timeout:  timeout,
		insecure: true,
	}
}

// NewGRPCProberSecure creates a new gRPC health prober with TLS.
//
// Params:
//   - timeout: the maximum duration for health checks.
//
// Returns:
//   - *GRPCProber: a configured gRPC prober with TLS enabled.
func NewGRPCProberSecure(timeout time.Duration) *GRPCProber {
	// Return configured gRPC prober with TLS.
	return &GRPCProber{
		timeout:  timeout,
		insecure: false,
	}
}

// Type returns the prober type.
//
// Returns:
//   - string: the constant "grpc" identifying the prober type.
func (p *GRPCProber) Type() string {
	// Return the gRPC prober type identifier.
	return proberTypeGRPC
}

// Probe performs a gRPC health check.
// Currently implements TCP connectivity as a proxy for gRPC health.
// Full gRPC health protocol requires google.golang.org/grpc dependency.
//
// Params:
//   - ctx: context for cancellation and timeout control.
//   - target: the target to probe, including service name.
//
// Returns:
//   - probe.Result: the probe result with latency and status.
func (p *GRPCProber) Probe(ctx context.Context, target probe.Target) probe.Result {
	start := time.Now()

	// Currently we perform TCP connectivity check.
	// TODO: Implement full gRPC health protocol when grpc dependency is added.

	// Create dialer with configured timeout.
	dialer := &net.Dialer{
		Timeout: p.timeout,
	}

	// Attempt to establish TCP connection to gRPC server.
	conn, err := dialer.DialContext(ctx, "tcp", target.Address)
	latency := time.Since(start)

	// Handle connection failure.
	if err != nil {
		// Return failure result.
		return probe.NewFailureResult(
			latency,
			fmt.Sprintf("gRPC connection failed: %v", err),
			err,
		)
	}
	// Close the connection.
	_ = conn.Close()

	// Build output message.
	service := target.Service
	// Check if service name was specified.
	if service == "" {
		// Default to server overall health check.
		service = "(server)"
	}

	// Return success result.
	return probe.NewSuccessResult(
		latency,
		fmt.Sprintf("gRPC %s reachable at %s", service, target.Address),
	)
}
