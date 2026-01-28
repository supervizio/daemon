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
	if payload == nil {
		return NewUDPProber(timeout)
	}

	payloadCopy := slices.Clone(payload)

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

	select {
	case <-ctx.Done():
		return health.NewFailureCheckResult(
			time.Since(start),
			fmt.Sprintf("context cancelled: %v", ctx.Err()),
			ctx.Err(),
		)
	default:
	}

	conn, result := p.dialUDP(ctx, target, start)
	if conn == nil {
		return result
	}
	defer func() { _ = conn.Close() }()

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
	if network == "" {
		network = "udp"
	}

	addr, err := net.ResolveUDPAddr(network, target.Address)
	if err != nil {
		return nil, health.NewFailureCheckResult(
			time.Since(start),
			fmt.Sprintf("failed to resolve address: %v", err),
			err,
		)
	}

	conn, err := net.DialUDP(network, nil, addr)
	if err != nil {
		return nil, health.NewFailureCheckResult(
			time.Since(start),
			fmt.Sprintf("failed to dial: %v", err),
			err,
		)
	}

	deadline := p.calculateDeadline(ctx)

	if err := conn.SetDeadline(deadline); err != nil {
		_ = conn.Close()

		return nil, health.NewFailureCheckResult(
			time.Since(start),
			fmt.Sprintf("failed to set deadline: %v", err),
			err,
		)
	}

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
	if p.timeout > 0 {
		proberDeadline = now.Add(p.timeout)
	}

	// Use the earliest deadline between prober timeout and context deadline.
	if ctxDeadline, ok := ctx.Deadline(); ok {
		if !proberDeadline.IsZero() && proberDeadline.Before(ctxDeadline) {
			return proberDeadline
		}

		return ctxDeadline
	}

	if !proberDeadline.IsZero() {
		return proberDeadline
	}

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
	if err != nil {
		return health.NewFailureCheckResult(
			time.Since(start),
			fmt.Sprintf("failed to write: %v", err),
			err,
		)
	}

	buffer := make([]byte, udpBufferSize)
	n, err := conn.Read(buffer)
	latency := time.Since(start)

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
	if err != nil {
		// For UDP, no response doesn't necessarily mean failure.
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return health.NewSuccessCheckResult(
				latency,
				fmt.Sprintf("sent to %s (no response within timeout)", address),
			)
		}

		return health.NewFailureCheckResult(
			latency,
			fmt.Sprintf("failed to read response: %v", err),
			err,
		)
	}

	return health.NewSuccessCheckResult(
		latency,
		fmt.Sprintf("received %d bytes from %s", n, address),
	)
}
