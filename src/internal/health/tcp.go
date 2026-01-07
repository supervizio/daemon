package health

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/kodflow/daemon/internal/config"
)

// TCPChecker performs TCP health checks.
type TCPChecker struct {
	name    string
	address string
	timeout time.Duration
}

// NewTCPChecker creates a new TCP health checker.
func NewTCPChecker(cfg *config.HealthCheckConfig) *TCPChecker {
	name := cfg.Name
	if name == "" {
		name = fmt.Sprintf("tcp-%s:%d", cfg.Host, cfg.Port)
	}

	return &TCPChecker{
		name:    name,
		address: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		timeout: cfg.Timeout.Duration(),
	}
}

// Name returns the checker name.
func (c *TCPChecker) Name() string {
	return c.name
}

// Type returns the checker type.
func (c *TCPChecker) Type() string {
	return "tcp"
}

// Check performs a TCP health check by attempting to connect.
func (c *TCPChecker) Check(ctx context.Context) Result {
	start := time.Now()

	dialer := &net.Dialer{
		Timeout: c.timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", c.address)
	if err != nil {
		return Result{
			Status:    StatusUnhealthy,
			Message:   fmt.Sprintf("connection failed: %v", err),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
			Error:     err,
		}
	}
	conn.Close()

	return Result{
		Status:    StatusHealthy,
		Message:   fmt.Sprintf("connected to %s", c.address),
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	}
}
