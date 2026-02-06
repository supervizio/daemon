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
	payloadCopy := slices.Clone(defaultUDPPayload)

	// use default payload with timeout configuration
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
	// handle nil payload
	if payload == nil {
		// fallback to default when no payload provided
		return NewUDPProber(timeout)
	}

	payloadCopy := slices.Clone(payload)

	// use custom payload with timeout configuration
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
	// identify this prober as udp type
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

	// check for context cancellation
	select {
	case <-ctx.Done():
		// return early if context already cancelled
		return health.NewFailureCheckResult(
			time.Since(start),
			fmt.Sprintf("context cancelled: %v", ctx.Err()),
			ctx.Err(),
		)
	default:
	}

	conn, result := p.dialUDP(ctx, target, start)
	// handle dial failure
	if conn == nil {
		// dial failure means address resolution or connection failed
		return result
	}
	// ensure connection is closed
	defer func() { _ = conn.Close() }()

	// send probe and process response
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
	network := target.Network
	// use default network if not specified
	if network == "" {
		network = "udp"
	}

	addr, err := net.ResolveUDPAddr(network, target.Address)
	// handle address resolution failure
	if err != nil {
		// address resolution failure indicates invalid target
		return nil, health.NewFailureCheckResult(
			time.Since(start),
			fmt.Sprintf("failed to resolve address: %v", err),
			err,
		)
	}

	conn, err := net.DialUDP(network, nil, addr)
	// handle connection failure
	if err != nil {
		// udp dial errors are rare but indicate network issues
		return nil, health.NewFailureCheckResult(
			time.Since(start),
			fmt.Sprintf("failed to dial: %v", err),
			err,
		)
	}

	deadline := p.calculateDeadline(ctx)

	// set deadline for UDP operations
	if err := conn.SetDeadline(deadline); err != nil {
		_ = conn.Close()

		// deadline errors indicate system issues
		return nil, health.NewFailureCheckResult(
			time.Since(start),
			fmt.Sprintf("failed to set deadline: %v", err),
			err,
		)
	}

	// return connection with empty result on success
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

	var proberDeadline time.Time
	// calculate prober timeout deadline
	if p.timeout > 0 {
		proberDeadline = now.Add(p.timeout)
	}

	// Use the earliest deadline between prober timeout and context deadline.
	// check for context deadline
	if ctxDeadline, ok := ctx.Deadline(); ok {
		// use prober deadline if earlier
		if !proberDeadline.IsZero() && proberDeadline.Before(ctxDeadline) {
			// prober timeout takes precedence when earlier
			return proberDeadline
		}

		// context deadline takes precedence when earlier
		return ctxDeadline
	}

	// use prober deadline if set
	if !proberDeadline.IsZero() {
		// prober timeout is the only deadline available
		return proberDeadline
	}

	// fallback to default timeout
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
	_, err := conn.Write(p.payload)
	// handle write failure
	if err != nil {
		// write errors indicate network issues or invalid connection
		return health.NewFailureCheckResult(
			time.Since(start),
			fmt.Sprintf("failed to write: %v", err),
			err,
		)
	}

	buffer := make([]byte, udpBufferSize)
	n, err := conn.Read(buffer)
	latency := time.Since(start)

	// process read result
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
	// handle read error
	if err != nil {
		// For UDP, no response doesn't necessarily mean failure.
		var netErr net.Error
		// treat timeout as success for UDP
		if errors.As(err, &netErr) && netErr.Timeout() {
			// udp timeout is acceptable as packet was sent successfully
			return health.NewSuccessCheckResult(
				latency,
				fmt.Sprintf("sent to %s (no response within timeout)", address),
			)
		}

		// other errors indicate actual failure
		return health.NewFailureCheckResult(
			latency,
			fmt.Sprintf("failed to read response: %v", err),
			err,
		)
	}

	// return success with bytes received
	return health.NewSuccessCheckResult(
		latency,
		fmt.Sprintf("received %d bytes from %s", n, address),
	)
}
