// Package health provides infrastructure adapters for health checking.
package health

import (
	"context"
	"fmt"
	"net"
	"time"

	domain "github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/service"
)

// checkerTypeTCP is the type identifier for TCP health checkers.
const checkerTypeTCP string = "tcp"

// networkTCP is the network type for TCP connections.
const networkTCP string = "tcp"

// TCPChecker performs TCP health checks by attempting connections.
// It verifies service availability by establishing TCP connections to the target.
type TCPChecker struct {
	// name holds the identifier for this health checker instance.
	name string
	// address stores the target address in host:port format.
	address string
	// timeout defines the maximum duration for connection attempts.
	timeout time.Duration
}

// NewTCPChecker creates a new TCP health checker.
// It configures the checker with the provided health check settings.
//
// Params:
//   - cfg: configuration containing host, port, timeout, and optional name
//
// Returns:
//   - *TCPChecker: a configured TCP health checker ready to perform checks
func NewTCPChecker(cfg *service.HealthCheckConfig) *TCPChecker {
	// Use provided name or generate one from host and port.
	name := cfg.Name
	// Generate default name from host and port when no custom name is provided.
	if name == "" {
		name = fmt.Sprintf("tcp-%s:%d", cfg.Host, cfg.Port)
	}

	// Build and return the configured TCP checker.
	return &TCPChecker{
		name:    name,
		address: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		timeout: cfg.Timeout.Duration(),
	}
}

// Name returns the checker name.
// It provides the identifier for this health checker instance.
//
// Returns:
//   - string: the name of this health checker
func (c *TCPChecker) Name() string {
	// Return the configured checker name.
	return c.name
}

// Type returns the checker type.
// It identifies this as a TCP-based health checker.
//
// Returns:
//   - string: the constant "tcp" identifying the checker type
func (c *TCPChecker) Type() string {
	// Return the TCP checker type identifier.
	return checkerTypeTCP
}

// Check performs a TCP health check by attempting to connect.
// It tries to establish a TCP connection to verify the target is accepting connections.
//
// Params:
//   - ctx: context for cancellation and timeout control
//
// Returns:
//   - domain.Result: the health check result indicating healthy or unhealthy status
func (c *TCPChecker) Check(ctx context.Context) domain.Result {
	start := time.Now()

	dialer := &net.Dialer{
		Timeout: c.timeout,
	}

	conn, err := dialer.DialContext(ctx, networkTCP, c.address)
	// Handle connection failure by returning unhealthy result.
	if err != nil {
		// Return unhealthy result with connection error details.
		return domain.NewUnhealthyResult(
			fmt.Sprintf("connection failed: %v", err),
			time.Since(start),
			err,
		)
	}
	_ = conn.Close()

	// Return healthy result indicating successful connection.
	return domain.NewHealthyResult(
		fmt.Sprintf("connected to %s", c.address),
		time.Since(start),
	)
}
