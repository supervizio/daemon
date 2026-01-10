// Package probe provides infrastructure adapters for service probing.
package probe

import (
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"time"

	"github.com/kodflow/daemon/internal/domain/probe"
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
	// timeout is the maximum duration for the probe.
	timeout time.Duration
	// payload is the data to send for the probe.
	payload []byte
}

// NewUDPProber creates a new UDP prober.
//
// Params:
//   - timeout: the maximum duration for the probe.
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
//   - timeout: the maximum duration for the probe.
//   - payload: the data to send for the probe.
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

// Probe performs a UDP probe.
// It sends a UDP packet and optionally waits for a response.
// Note: UDP is connectionless, so the "connection" always succeeds.
// The probe measures the ability to send and receive data.
//
// Params:
//   - ctx: context for cancellation and timeout control.
//   - target: the target to probe.
//
// Returns:
//   - probe.Result: the probe result with latency and status.
func (p *UDPProber) Probe(ctx context.Context, target probe.Target) probe.Result {
	start := time.Now()

	// Check if context is already cancelled.
	select {
	case <-ctx.Done():
		// Return early with context error.
		return probe.NewFailureResult(
			time.Since(start),
			fmt.Sprintf("context cancelled: %v", ctx.Err()),
			ctx.Err(),
		)
	default:
		// Continue with probe.
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
//   - probe.Result: failure result if connection failed.
func (p *UDPProber) dialUDP(ctx context.Context, target probe.Target, start time.Time) (*net.UDPConn, probe.Result) {
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
		return nil, probe.NewFailureResult(
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
		return nil, probe.NewFailureResult(
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
		return nil, probe.NewFailureResult(
			time.Since(start),
			fmt.Sprintf("failed to set deadline: %v", err),
			err,
		)
	}

	// Return established connection with empty result.
	return conn, probe.Result{}
}

// calculateDeadline determines the deadline from timeout or context.
//
// Params:
//   - ctx: context that may contain a deadline.
//
// Returns:
//   - time.Time: the calculated deadline.
func (p *UDPProber) calculateDeadline(ctx context.Context) time.Time {
	// Use prober timeout if positive.
	if p.timeout > 0 {
		// Return deadline based on configured timeout.
		return time.Now().Add(p.timeout)
	}

	// Use context deadline if available.
	if ctxDeadline, ok := ctx.Deadline(); ok {
		// Return context-provided deadline.
		return ctxDeadline
	}

	// Fall back to default timeout.
	return time.Now().Add(probe.DefaultTimeout)
}

// sendAndReceive sends the probe packet and reads the response.
//
// Params:
//   - conn: the UDP connection.
//   - address: the target address for logging.
//   - start: the start time for latency measurement.
//
// Returns:
//   - probe.Result: the probe result.
func (p *UDPProber) sendAndReceive(conn udpConn, address string, start time.Time) probe.Result {
	// Send probe packet.
	_, err := conn.Write(p.payload)
	// Check if write operation failed.
	if err != nil {
		// Return failure for write error.
		return probe.NewFailureResult(
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
//   - probe.Result: the probe result.
func (p *UDPProber) handleReadResult(err error, n int, address string, latency time.Duration) probe.Result {
	// Handle read timeout gracefully for UDP.
	if err != nil {
		// For UDP, no response doesn't necessarily mean failure.
		// We consider the probe successful if we could send the packet.
		var netErr net.Error
		// Check if error is a timeout.
		if errors.As(err, &netErr) && netErr.Timeout() {
			// Return success with timeout note - packet was sent.
			return probe.NewSuccessResult(
				latency,
				fmt.Sprintf("sent to %s (no response within timeout)", address),
			)
		}
		// Other errors indicate actual failures.
		return probe.NewFailureResult(
			latency,
			fmt.Sprintf("failed to read response: %v", err),
			err,
		)
	}

	// Return success with response details.
	return probe.NewSuccessResult(
		latency,
		fmt.Sprintf("received %d bytes from %s", n, address),
	)
}
