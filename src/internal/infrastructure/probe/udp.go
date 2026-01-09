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
	// Return configured UDP prober with default payload.
	return &UDPProber{
		timeout: timeout,
		payload: defaultUDPPayload,
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
	// Return configured UDP prober with custom payload.
	return &UDPProber{
		timeout: timeout,
		payload: payload,
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
//   - _ctx: context for cancellation (unused, required by interface).
//   - target: the target to probe.
//
// Returns:
//   - probe.Result: the probe result with latency and status.
func (p *UDPProber) Probe(_ctx context.Context, target probe.Target) probe.Result {
	start := time.Now()

	// Establish UDP connection to target.
	conn, result := p.dialUDP(target, start)
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
//   - target: the probe target.
//   - start: the start time for latency measurement.
//
// Returns:
//   - *net.UDPConn: the established connection (nil if failed).
//   - probe.Result: failure result if connection failed.
func (p *UDPProber) dialUDP(target probe.Target, start time.Time) (*net.UDPConn, probe.Result) {
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

	// Set deadlines for read/write operations.
	deadline := time.Now().Add(p.timeout)
	// Check if deadline configuration failed.
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
