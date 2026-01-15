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
)

// GRPCProber performs gRPC health check probes using the standard health/v1 protocol.
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

	// Create context with timeout.
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// Build dial options.
	// Note: WithBlock and DialContext are deprecated but supported throughout gRPC 1.x.
	// They provide blocking behavior required for health check probing.
	opts := []grpc.DialOption{
		grpc.WithBlock(), //nolint:staticcheck // supported throughout 1.x, needed for blocking
	}

	// Add transport credentials.
	if p.insecureMode {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Establish gRPC connection.
	conn, err := grpc.DialContext(ctx, target.Address, opts...) //nolint:staticcheck // supported throughout 1.x
	if err != nil {
		latency := time.Since(start)
		return health.NewFailureCheckResult(
			latency,
			fmt.Sprintf("gRPC connection failed: %v", err),
			err,
		)
	}
	defer func() { _ = conn.Close() }()

	// Create health client.
	client := grpc_health_v1.NewHealthClient(conn)

	// Build health check request.
	req := &grpc_health_v1.HealthCheckRequest{
		Service: target.Service,
	}

	// Perform health check.
	resp, err := client.Check(ctx, req)
	latency := time.Since(start)

	// Handle RPC errors.
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				return health.NewFailureCheckResult(
					latency,
					fmt.Sprintf("gRPC service %q unknown", target.Service),
					ErrGRPCServiceUnknown,
				)
			case codes.DeadlineExceeded:
				return health.NewFailureCheckResult(
					latency,
					"gRPC health check timeout",
					err,
				)
			default:
				return health.NewFailureCheckResult(
					latency,
					fmt.Sprintf("gRPC health check failed: %s", st.Message()),
					err,
				)
			}
		}
		return health.NewFailureCheckResult(
			latency,
			fmt.Sprintf("gRPC health check failed: %v", err),
			err,
		)
	}

	// Check response status.
	switch resp.Status {
	case grpc_health_v1.HealthCheckResponse_SERVING:
		service := target.Service
		if service == "" {
			service = "(server)"
		}
		return health.NewSuccessCheckResult(
			latency,
			fmt.Sprintf("gRPC %s serving at %s", service, target.Address),
		)
	case grpc_health_v1.HealthCheckResponse_NOT_SERVING:
		return health.NewFailureCheckResult(
			latency,
			fmt.Sprintf("gRPC service %q not serving", target.Service),
			ErrGRPCNotServing,
		)
	case grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN:
		return health.NewFailureCheckResult(
			latency,
			fmt.Sprintf("gRPC service %q unknown", target.Service),
			ErrGRPCServiceUnknown,
		)
	default:
		return health.NewFailureCheckResult(
			latency,
			fmt.Sprintf("gRPC service %q status unknown: %v", target.Service, resp.Status),
			errors.New("unknown health status"),
		)
	}
}
