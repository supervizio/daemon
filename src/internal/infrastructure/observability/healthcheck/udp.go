// Package healthcheck provides infrastructure adapters for service probing.
package healthcheck

import (
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"time"

	"github.com/kodflow/daemon/internal/domain/health"
)

// proberTypeUDP is the type identifier for UDP probers.
const proberTypeUDP string = "udp"

// udpBufferSize is the size of the buffer for reading UDP responses.
const udpBufferSize int = 1024

// defaultUDPPayload is the default payload sent for UDP probes.
var defaultUDPPayload []byte = []byte("PING")

// udpConn defines the minimal interface for UDP connection operations.
// This interface enables dependency injection and testing without network I/O.
type udpConn interface {
	// Read reads data from the connection.
	Read(b []byte) (int, error)
	// Write writes data to the connection.
	Write(b []byte) (int, error)
}

// UDPProber performs UDP connection probes.
// Note: UDP is connectionless, so probes send a packet and wait for response.
// This is less reliable than TCP probes since no response doesn't mean failure.
type UDPProber struct {
	// timeout is the maximum duration for the healthcheck.
	timeout time.Duration
	// payload is the data to send for the healthcheck.
	payload []byte
}

// NewUDPProber creates a new UDP prober.
//
// Params:
//   - timeout: the maximum duration for the healthcheck.
//
// Returns:
//   - *UDPProber: a configured UDP prober ready to perform probes.
func NewUDPProber(timeout time.Duration) *UDPProber {
	// Defensive copy of default payload to prevent external modifications.
	payloadCopy := slices.Clone(defaultUDPPayload)

	// Return configured UDP prober with copied payload.
	return &UDPProber{
		timeout: timeout,
		payload: payloadCopy,
	}
}

// NewUDPProberWithPayload creates a new UDP prober with custom payload.
//
// Params:
//   - timeout: the maximum duration for the healthcheck.
//   - payload: the data to send for the healthcheck.
//
// Returns:
//   - *UDPProber: a configured UDP prober ready to perform probes.
func NewUDPProberWithPayload(timeout time.Duration, payload []byte) *UDPProber {
	// Use default prober if payload is nil.
	if payload == nil {
		// Return prober with default payload.
		return NewUDPProber(timeout)
	}

	// Defensive copy of payload to prevent external modifications.
	payloadCopy := slices.Clone(payload)

	// Return configured UDP prober with copied payload.
	return &UDPProber{
		timeout: timeout,
		payload: payloadCopy,
	}
}

// Type returns the prober type.
//
// Returns:
//   - string: the constant "udp" identifying the prober type.
func (p *UDPProber) Type() string {
	// Return the UDP prober type identifier.
	return proberTypeUDP
}

// Probe performs a UDP healthcheck.
// It sends a UDP packet and optionally waits for a response.
// Note: UDP is connectionless, so the "connection" always succeeds.
// The probe measures the ability to send and receive data.
//
// Params:
//   - ctx: context for cancellation and timeout control.
//   - target: the target to healthcheck.
//
// Returns:
//   - health.CheckResult: the probe result with latency and status.
func (p *UDPProber) Probe(ctx context.Context, target health.Target) health.CheckResult {
	start := time.Now()

	// Check if context is already cancelled.
	select {
	case <-ctx.Done():
		// Return early with context error.
		return health.NewFailureCheckResult(
			time.Since(start),
			fmt.Sprintf("context cancelled: %v", ctx.Err()),
			ctx.Err(),
		)
	default:
		// Continue with healthcheck.
	}

	// Establish UDP connection to target.
	conn, result := p.dialUDP(ctx, target, start)
	// Check if connection failed.
	if conn == nil {
		// Return failure result from dial.
		return result
	}
	defer func() {
		// Close connection to release resources.
		_ = conn.Close()
	}()

	// Send and receive UDP data.
	return p.sendAndReceive(conn, target.Address, start)
}

// dialUDP establishes a UDP connection to the target.
//
// Params:
//   - ctx: context for cancellation (used to derive deadline if timeout is zero).
//   - target: the probe target.
//   - start: the start time for latency measurement.
//
// Returns:
//   - *net.UDPConn: the established connection (nil if failed).
//   - health.CheckResult: failure result if connection failed.
func (p *UDPProber) dialUDP(ctx context.Context, target health.Target, start time.Time) (*net.UDPConn, health.CheckResult) {
	// Determine the network type.
	network := target.Network
	// Check if network type was specified.
	if network == "" {
		// Default to UDP if not specified.
		network = "udp"
	}

	// Resolve UDP address.
	addr, err := net.ResolveUDPAddr(network, target.Address)
	// Check if address resolution failed.
	if err != nil {
		// Return failure for address resolution error.
		return nil, health.NewFailureCheckResult(
			time.Since(start),
			fmt.Sprintf("failed to resolve address: %v", err),
			err,
		)
	}

	// Create UDP connection.
	conn, err := net.DialUDP(network, nil, addr)
	// Check if connection creation failed.
	if err != nil {
		// Return failure for connection error.
		return nil, health.NewFailureCheckResult(
			time.Since(start),
			fmt.Sprintf("failed to dial: %v", err),
			err,
		)
	}

	// Determine deadline from timeout or context.
	deadline := p.calculateDeadline(ctx)

	// Set deadlines for read/write operations.
	if err := conn.SetDeadline(deadline); err != nil {
		// Close connection before returning.
		_ = conn.Close()
		// Return failure for deadline error.
		return nil, health.NewFailureCheckResult(
			time.Since(start),
			fmt.Sprintf("failed to set deadline: %v", err),
			err,
		)
	}

	// Return established connection with empty result.
	return conn, health.CheckResult{}
}

// calculateDeadline determines the deadline from timeout or context.
//
// Params:
//   - ctx: context that may contain a deadline.
//
// Returns:
//   - time.Time: the calculated deadline.
func (p *UDPProber) calculateDeadline(ctx context.Context) time.Time {
	now := time.Now()

	// Compute prober deadline if configured.
	var proberDeadline time.Time
	// Set prober deadline only when timeout is positive.
	if p.timeout > 0 {
		proberDeadline = now.Add(p.timeout)
	}

	// Prefer the earliest deadline between prober timeout and context deadline.
	if ctxDeadline, ok := ctx.Deadline(); ok {
		// Check if prober deadline is earlier than context deadline.
		if !proberDeadline.IsZero() && proberDeadline.Before(ctxDeadline) {
			// Return prober deadline as it expires sooner.
			return proberDeadline
		}
		// Return context-provided deadline.
		return ctxDeadline
	}

	// Use prober deadline if set.
	if !proberDeadline.IsZero() {
		// Return prober-configured deadline.
		return proberDeadline
	}

	// Fall back to default timeout when neither is set.
	return now.Add(health.DefaultTimeout)
}

// sendAndReceive sends the probe packet and reads the response.
//
// Params:
//   - conn: the UDP connection.
//   - address: the target address for logging.
//   - start: the start time for latency measurement.
//
// Returns:
//   - health.CheckResult: the probe result.
func (p *UDPProber) sendAndReceive(conn udpConn, address string, start time.Time) health.CheckResult {
	// Send probe packet.
	_, err := conn.Write(p.payload)
	// Check if write operation failed.
	if err != nil {
		// Return failure for write error.
		return health.NewFailureCheckResult(
			time.Since(start),
			fmt.Sprintf("failed to write: %v", err),
			err,
		)
	}

	// Try to read response (optional for UDP).
	buffer := make([]byte, udpBufferSize)
	n, err := conn.Read(buffer)
	latency := time.Since(start)

	// Handle read result.
	return p.handleReadResult(err, n, address, latency)
}

// handleReadResult processes the UDP read result.
//
// Params:
//   - err: any error from the read operation.
//   - n: number of bytes read.
//   - address: the target address for logging.
//   - latency: the measured latency.
//
// Returns:
//   - health.CheckResult: the probe result.
func (p *UDPProber) handleReadResult(err error, n int, address string, latency time.Duration) health.CheckResult {
	// Handle read timeout gracefully for UDP.
	if err != nil {
		// For UDP, no response doesn't necessarily mean failure.
		// We consider the probe successful if we could send the packet.
		var netErr net.Error
		// Check if error is a timeout.
		if errors.As(err, &netErr) && netErr.Timeout() {
			// Return success with timeout note - packet was sent.
			return health.NewSuccessCheckResult(
				latency,
				fmt.Sprintf("sent to %s (no response within timeout)", address),
			)
		}
		// Other errors indicate actual failures.
		return health.NewFailureCheckResult(
			latency,
			fmt.Sprintf("failed to read response: %v", err),
			err,
		)
	}

	// Return success with response details.
	return health.NewSuccessCheckResult(
		latency,
		fmt.Sprintf("received %d bytes from %s", n, address),
	)
}
