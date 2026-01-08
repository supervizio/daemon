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

// defaultUDPPayload is the default payload sent for UDP probes.
var defaultUDPPayload = []byte("PING")

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
//   - ctx: context for cancellation and timeout control.
//   - target: the target to probe.
//
// Returns:
//   - probe.Result: the probe result with latency and status.
func (p *UDPProber) Probe(ctx context.Context, target probe.Target) probe.Result {
	start := time.Now()

	// Determine the network type.
	network := target.Network
	if network == "" {
		// Default to UDP if not specified.
		network = "udp"
	}

	// Resolve UDP address.
	addr, err := net.ResolveUDPAddr(network, target.Address)
	if err != nil {
		// Return failure for address resolution error.
		return probe.NewFailureResult(
			time.Since(start),
			fmt.Sprintf("failed to resolve address: %v", err),
			err,
		)
	}

	// Create UDP connection.
	conn, err := net.DialUDP(network, nil, addr)
	if err != nil {
		// Return failure for connection error.
		return probe.NewFailureResult(
			time.Since(start),
			fmt.Sprintf("failed to dial: %v", err),
			err,
		)
	}
	defer func() {
		// Close connection.
		_ = conn.Close()
	}()

	// Set deadlines.
	deadline := time.Now().Add(p.timeout)
	if err := conn.SetDeadline(deadline); err != nil {
		// Return failure for deadline error.
		return probe.NewFailureResult(
			time.Since(start),
			fmt.Sprintf("failed to set deadline: %v", err),
			err,
		)
	}

	// Send probe packet.
	_, err = conn.Write(p.payload)
	if err != nil {
		// Return failure for write error.
		return probe.NewFailureResult(
			time.Since(start),
			fmt.Sprintf("failed to write: %v", err),
			err,
		)
	}

	// Try to read response (optional for UDP).
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	latency := time.Since(start)

	// Handle read timeout gracefully for UDP.
	if err != nil {
		// For UDP, no response doesn't necessarily mean failure.
		// We consider the probe successful if we could send the packet.
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			// Return success with timeout note.
			return probe.NewSuccessResult(
				latency,
				fmt.Sprintf("sent to %s (no response within timeout)", target.Address),
			)
		}
		// Other errors are failures.
		return probe.NewFailureResult(
			latency,
			fmt.Sprintf("failed to read response: %v", err),
			err,
		)
	}

	// Return success with response details.
	return probe.NewSuccessResult(
		latency,
		fmt.Sprintf("received %d bytes from %s", n, target.Address),
	)
}
