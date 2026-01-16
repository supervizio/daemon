// Package healthcheck provides infrastructure adapters for service probing.
package healthcheck

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	"github.com/kodflow/daemon/internal/domain/health"
)

// proberTypeGRPC is the type identifier for gRPC probers.
const proberTypeGRPC string = "grpc"

var (
	// ErrGRPCNotServing indicates the gRPC service is not serving.
	ErrGRPCNotServing error = errors.New("service not serving")

	// ErrGRPCServiceUnknown indicates the gRPC service is unknown.
	ErrGRPCServiceUnknown error = errors.New("service unknown")

	// ErrGRPCUnknownStatus indicates an unknown health check status was received.
	ErrGRPCUnknownStatus error = errors.New("unknown health status")
)

// GRPCProber performs gRPC health check probes using the standard health/v1 protocol.
// It connects to a gRPC service and uses the standard gRPC Health Checking Protocol
// to determine service health status.
type GRPCProber struct {
	// timeout is the maximum duration for the healthcheck.
	timeout time.Duration
	// insecure enables insecure connections (no TLS).
	insecureMode bool
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
		timeout:      timeout,
		insecureMode: true,
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
		timeout:      timeout,
		insecureMode: false,
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

// Probe performs a gRPC health check using the standard health/v1 protocol.
//
// Params:
//   - ctx: context for cancellation and timeout control.
//   - target: the target to probe, including service name.
//
// Returns:
//   - health.CheckResult: the probe result with latency and status.
func (p *GRPCProber) Probe(ctx context.Context, target health.Target) health.CheckResult {
	start := time.Now()

	// Establish connection with timeout.
	conn, err := p.connect(ctx, target.Address)
	// Check if connection failed.
	if err != nil {
		latency := time.Since(start)
		// Return connection failure result.
		return health.NewFailureCheckResult(
			latency,
			fmt.Sprintf("gRPC connection failed: %v", err),
			err,
		)
	}
	defer func() { _ = conn.Close() }()

	// Perform health check and calculate latency.
	resp, err := p.checkHealth(ctx, conn, target.Service)
	latency := time.Since(start)

	// Check if health check RPC failed.
	if err != nil {
		// Handle RPC error and return result.
		return p.handleRPCError(err, latency, target.Service)
	}

	// Process health status response.
	return p.handleHealthStatus(resp, latency, target)
}

// connect establishes a gRPC connection to the target address.
//
// Params:
//   - ctx: context for cancellation and timeout control.
//   - address: the target address to connect to.
//
// Returns:
//   - *grpc.ClientConn: the established connection.
//   - error: connection error if any.
func (p *GRPCProber) connect(ctx context.Context, address string) (*grpc.ClientConn, error) {
	// Create context with timeout.
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// Build dial options with blocking behavior.
	// Note: WithBlock and DialContext are deprecated but supported throughout gRPC 1.x.
	// They provide blocking behavior required for health check probing.
	opts := []grpc.DialOption{
		grpc.WithBlock(), //nolint:staticcheck // supported throughout 1.x, needed for blocking
	}

	// Check if insecure mode is enabled.
	if p.insecureMode {
		// Add insecure transport credentials.
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Establish gRPC connection and return result.
	return grpc.DialContext(ctx, address, opts...) //nolint:staticcheck // supported throughout 1.x
}

// checkHealth performs the health check RPC call.
//
// Params:
//   - ctx: context for cancellation and timeout control.
//   - conn: the gRPC connection to use.
//   - service: the service name to check.
//
// Returns:
//   - *grpc_health_v1.HealthCheckResponse: the health check response.
//   - error: RPC error if any.
func (p *GRPCProber) checkHealth(
	ctx context.Context,
	conn *grpc.ClientConn,
	service string,
) (*grpc_health_v1.HealthCheckResponse, error) {
	// Create health client from connection.
	client := grpc_health_v1.NewHealthClient(conn)

	// Build health check request with service name.
	req := &grpc_health_v1.HealthCheckRequest{
		Service: service,
	}

	// Perform health check RPC and return result.
	return client.Check(ctx, req)
}

// handleRPCError converts RPC errors to health check results.
//
// Params:
//   - err: the RPC error to handle.
//   - latency: the probe latency.
//   - service: the service name being checked.
//
// Returns:
//   - health.CheckResult: the failure result.
func (p *GRPCProber) handleRPCError(err error, latency time.Duration, service string) health.CheckResult {
	// Try to extract gRPC status from error.
	st, ok := status.FromError(err)
	// Check if status extraction succeeded.
	if !ok {
		// Return generic RPC failure result.
		return health.NewFailureCheckResult(
			latency,
			fmt.Sprintf("gRPC health check failed: %v", err),
			err,
		)
	}

	// Switch on gRPC status code.
	//nolint:exhaustive // codes.Code has 17+ values, default handles all others
	switch st.Code() {
	// Handle service not found.
	case codes.NotFound:
		// Return service unknown result.
		return health.NewFailureCheckResult(
			latency,
			fmt.Sprintf("gRPC service %q unknown", service),
			ErrGRPCServiceUnknown,
		)

	// Handle timeout.
	case codes.DeadlineExceeded:
		// Return timeout result.
		return health.NewFailureCheckResult(
			latency,
			"gRPC health check timeout",
			err,
		)

	// Handle other status codes.
	default:
		// Return generic status failure result.
		return health.NewFailureCheckResult(
			latency,
			fmt.Sprintf("gRPC health check failed: %s", st.Message()),
			err,
		)
	}
}

// handleHealthStatus converts health status responses to check results.
//
// Params:
//   - resp: the health check response.
//   - latency: the probe latency.
//   - target: the target being probed.
//
// Returns:
//   - health.CheckResult: the check result.
func (p *GRPCProber) handleHealthStatus(
	resp *grpc_health_v1.HealthCheckResponse,
	latency time.Duration,
	target health.Target,
) health.CheckResult {
	// Switch on health status.
	//nolint:exhaustive // UNKNOWN handled by default case
	switch resp.Status {
	// Handle serving status.
	case grpc_health_v1.HealthCheckResponse_SERVING:
		service := target.Service
		// Check if service name is empty.
		if service == "" {
			// Use default service name.
			service = "(server)"
		}
		// Return success result.
		return health.NewSuccessCheckResult(
			latency,
			fmt.Sprintf("gRPC %s serving at %s", service, target.Address),
		)

	// Handle not serving status.
	case grpc_health_v1.HealthCheckResponse_NOT_SERVING:
		// Return not serving failure result.
		return health.NewFailureCheckResult(
			latency,
			fmt.Sprintf("gRPC service %q not serving", target.Service),
			ErrGRPCNotServing,
		)

	// Handle service unknown status.
	case grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN:
		// Return service unknown failure result.
		return health.NewFailureCheckResult(
			latency,
			fmt.Sprintf("gRPC service %q unknown", target.Service),
			ErrGRPCServiceUnknown,
		)

	// Handle unknown status values.
	default:
		// Return unknown status failure result.
		return health.NewFailureCheckResult(
			latency,
			fmt.Sprintf("gRPC service %q status unknown: %v", target.Service, resp.Status),
			ErrGRPCUnknownStatus,
		)
	}
}
