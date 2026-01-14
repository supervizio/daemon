// Package healthcheck provides infrastructure adapters for service probing.
package healthcheck

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/kodflow/daemon/internal/domain/healthcheck"
	"github.com/kodflow/daemon/internal/domain/shared"
)

// proberTypeICMP is the type identifier for ICMP probers.
const proberTypeICMP string = "icmp"

// defaultTCPFallbackPort is the default port for TCP fallback probes.
// Port 80 (HTTP) is commonly open and suitable for connectivity checks.
const defaultTCPFallbackPort int = 80

// ICMPProber performs ICMP ping probes for latency measurement.
// It falls back to TCP probe when ICMP is not available (no CAP_NET_RAW).
//
// Note: ICMP requires elevated privileges (root or CAP_NET_RAW capability).
// In container environments, TCP fallback is commonly used.
type ICMPProber struct {
	// timeout is the maximum duration for the healthcheck.
	timeout time.Duration
	// useTCPFallback forces TCP fallback mode.
	useTCPFallback bool
	// tcpPort is the port for TCP fallback probes.
	tcpPort int
}

// NewICMPProber creates a new ICMP prober.
//
// Params:
//   - timeout: the maximum duration for ping.
//
// Returns:
//   - *ICMPProber: a configured ICMP prober ready to perform probes.
func NewICMPProber(timeout time.Duration) *ICMPProber {
	// Return configured ICMP prober with TCP fallback enabled by default.
	// ICMP requires CAP_NET_RAW which is often unavailable in containers.
	return &ICMPProber{
		timeout:        timeout,
		useTCPFallback: true,
		tcpPort:        defaultTCPFallbackPort,
	}
}

// NewICMPProberWithTCPFallback creates an ICMP prober with specific TCP fallback port.
//
// Params:
//   - timeout: the maximum duration for ping.
//   - tcpPort: the port to use for TCP fallback.
//
// Returns:
//   - *ICMPProber: a configured ICMP prober with TCP fallback.
func NewICMPProberWithTCPFallback(timeout time.Duration, tcpPort int) *ICMPProber {
	// Return configured ICMP prober with TCP fallback.
	return &ICMPProber{
		timeout:        timeout,
		useTCPFallback: true,
		tcpPort:        tcpPort,
	}
}

// Type returns the prober type.
//
// Returns:
//   - string: the constant "icmp" identifying the prober type.
func (p *ICMPProber) Type() string {
	// Return the ICMP prober type identifier.
	return proberTypeICMP
}

// Probe performs an ICMP ping or TCP fallback healthcheck.
// It measures network latency to the target host.
//
// Params:
//   - ctx: context for cancellation and timeout control.
//   - target: the target to healthcheck.
//
// Returns:
//   - healthcheck.Result: the probe result with latency measurement.
func (p *ICMPProber) Probe(ctx context.Context, target healthcheck.Target) healthcheck.Result {
	start := time.Now()

	// Resolve the target address.
	host := target.Address
	// Check if address contains port (e.g., host:port format).
	if hostPart, _, err := net.SplitHostPort(host); err == nil {
		// Extract just the host portion for ICMP/TCP ping.
		host = hostPart
	}

	// Use TCP fallback when CAP_NET_RAW is unavailable.
	if p.useTCPFallback {
		// Return TCP-based latency measurement.
		return p.tcpPing(ctx, host, start)
	}

	// TODO: Implement real ICMP ping using golang.org/x/net/icmp
	// For now, always use TCP fallback.
	return p.tcpPing(ctx, host, start)
}

// tcpPing performs a TCP-based ping for latency measurement.
// It measures the time to establish a TCP connection.
//
// Params:
//   - ctx: context for cancellation.
//   - host: the target host.
//   - start: the start time for latency measurement.
//
// Returns:
//   - healthcheck.Result: the probe result with latency.
func (p *ICMPProber) tcpPing(ctx context.Context, host string, start time.Time) healthcheck.Result {
	// Validate and normalize TCP port.
	tcpPort := p.tcpPort
	// Apply default port when not configured or invalid.
	if tcpPort <= 0 || tcpPort > shared.MaxValidPort {
		tcpPort = defaultTCPFallbackPort
	}

	// Build address with TCP port (handles IPv6 correctly).
	address := net.JoinHostPort(host, strconv.Itoa(tcpPort))

	// Create dialer with configured timeout.
	dialer := &net.Dialer{
		Timeout: p.timeout,
	}

	// Attempt to establish connection.
	conn, err := dialer.DialContext(ctx, "tcp", address)
	latency := time.Since(start)

	// Handle connection failure.
	if err != nil {
		// Return failure result.
		return healthcheck.NewFailureResult(
			latency,
			fmt.Sprintf("ping failed: %v", err),
			err,
		)
	}
	// Close the connection.
	_ = conn.Close()

	// Return success result with latency.
	return healthcheck.NewSuccessResult(
		latency,
		fmt.Sprintf("ping %s: latency=%s (tcp fallback)", host, latency),
	)
}
