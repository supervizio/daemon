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
	// insecure mode is default for testing and internal services
	return &GRPCProber{timeout: timeout, insecureMode: true}
}

// NewGRPCProberSecure creates a new gRPC health prober with TLS.
//
// Params:
//   - timeout: the maximum duration for health checks.
//
// Returns:
//   - *GRPCProber: a configured gRPC prober with TLS enabled.
func NewGRPCProberSecure(timeout time.Duration) *GRPCProber {
	// secure mode for production services with TLS
	return &GRPCProber{timeout: timeout, insecureMode: false}
}

// Type returns the prober type.
//
// Returns:
//   - string: the constant "grpc" identifying the prober type.
func (p *GRPCProber) Type() string { return proberTypeGRPC }

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
	conn, err := p.connect(ctx, target.Address)
	// handle connection failure
	if err != nil {
		// connection errors indicate service is unreachable
		return health.NewFailureCheckResult(time.Since(start), fmt.Sprintf("gRPC connection failed: %v", err), err)
	}
	// ensure connection is closed after use
	defer func() { _ = conn.Close() }()

	resp, err := p.checkHealth(ctx, conn, target.Service)
	latency := time.Since(start)
	// handle RPC call failure
	if err != nil {
		// convert grpc errors to health check results
		return p.handleRPCError(err, latency, target.Service)
	}
	// process health status response
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
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// WithBlock provides blocking behavior required for health check probing.
	opts := []grpc.DialOption{
		grpc.WithBlock(),
	}
	// add insecure credentials if needed
	if p.insecureMode {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	// establish blocking connection with timeout
	return grpc.DialContext(ctx, address, opts...)
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
	client := grpc_health_v1.NewHealthClient(conn)
	req := &grpc_health_v1.HealthCheckRequest{Service: service}
	// execute health check rpc call
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
	st, ok := status.FromError(err)
	// handle non-grpc errors
	if !ok {
		// non-grpc errors should still be reported as failures
		return health.NewFailureCheckResult(latency, fmt.Sprintf("gRPC health check failed: %v", err), err)
	}

	// handle specific error codes
	switch st.Code() {
	// service not found means misconfiguration or service not registered
	case codes.NotFound:
		// service name not registered with health service
		return health.NewFailureCheckResult(latency, fmt.Sprintf("gRPC service %q unknown", service), ErrGRPCServiceUnknown)
	// timeout means service exists but is too slow
	case codes.DeadlineExceeded:
		// service took too long to respond
		return health.NewFailureCheckResult(latency, "gRPC health check timeout", err)
	// all other codes indicate service failure
	default:
		// unexpected grpc error code
		return health.NewFailureCheckResult(latency, fmt.Sprintf("gRPC health check failed: %s", st.Message()), err)
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
	// map health status to check result
	switch resp.Status {
	// service is healthy and accepting requests
	case grpc_health_v1.HealthCheckResponse_SERVING:
		service := target.Service
		// use default name for server check
		if service == "" {
			service = "(server)"
		}
		// service is confirmed healthy
		return health.NewSuccessCheckResult(latency, fmt.Sprintf("gRPC %s serving at %s", service, target.Address))
	// service exists but is not accepting requests
	case grpc_health_v1.HealthCheckResponse_NOT_SERVING:
		// service is down or degraded
		return health.NewFailureCheckResult(latency, fmt.Sprintf("gRPC service %q not serving", target.Service), ErrGRPCNotServing)
	// service name is not recognized by the server
	case grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN:
		// service name unknown to health endpoint
		return health.NewFailureCheckResult(latency, fmt.Sprintf("gRPC service %q unknown", target.Service), ErrGRPCServiceUnknown)
	// unknown status should be treated as failure
	default:
		// unexpected health status value
		return health.NewFailureCheckResult(latency, fmt.Sprintf("gRPC service %q status unknown: %v", target.Service, resp.Status), ErrGRPCUnknownStatus)
	}
}
