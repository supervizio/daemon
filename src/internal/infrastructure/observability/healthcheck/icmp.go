// Package healthcheck provides infrastructure adapters for service probing.
package healthcheck

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/shared"
)

// proberTypeICMP is the type identifier for ICMP probers.
const proberTypeICMP string = "icmp"

// defaultTCPFallbackPort is the default port for TCP fallback probes.
// Port 80 (HTTP) is commonly open and suitable for connectivity checks.
const defaultTCPFallbackPort int = 80

// ICMPProber performs ICMP ping probes for latency measurement.
// It supports native ICMP (requires CAP_NET_RAW) or TCP fallback mode.
//
// Note: Native ICMP requires elevated privileges (root or CAP_NET_RAW capability).
// In container environments, TCP fallback is commonly used.
type ICMPProber struct {
	// timeout is the maximum duration for the healthcheck.
	timeout time.Duration
	// mode specifies how ICMP probes should operate.
	mode config.ICMPMode
	// hasNativeCapability indicates if native ICMP is available.
	hasNativeCapability bool
	// tcpPort is the port for TCP fallback probes.
	tcpPort int
}

// NewICMPProber creates a new ICMP prober with auto mode.
//
// Params:
//   - timeout: the maximum duration for ping.
//
// Returns:
//   - *ICMPProber: a configured ICMP prober ready to perform probes.
func NewICMPProber(timeout time.Duration) *ICMPProber {
	// auto mode with capability detection
	return &ICMPProber{
		timeout:             timeout,
		mode:                config.ICMPModeAuto,
		hasNativeCapability: detectICMPCapability(),
		tcpPort:             defaultTCPFallbackPort,
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
	// force TCP fallback mode
	return &ICMPProber{
		timeout:             timeout,
		mode:                config.ICMPModeFallback,
		hasNativeCapability: false,
		tcpPort:             tcpPort,
	}
}

// NewICMPProberWithMode creates an ICMP prober with specified mode.
//
// Params:
//   - timeout: the maximum duration for ping.
//   - mode: the ICMP mode (native, fallback, auto).
//
// Returns:
//   - *ICMPProber: a configured ICMP prober with specified mode.
func NewICMPProberWithMode(timeout time.Duration, mode config.ICMPMode) *ICMPProber {
	// create prober with specified mode
	return &ICMPProber{
		timeout:             timeout,
		mode:                mode,
		hasNativeCapability: detectICMPCapability(),
		tcpPort:             defaultTCPFallbackPort,
	}
}

// Type returns the prober type.
//
// Returns:
//   - string: the constant "icmp" identifying the prober type.
func (p *ICMPProber) Type() string {
	// identify this prober as icmp type
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
//   - health.CheckResult: the probe result with latency measurement.
func (p *ICMPProber) Probe(ctx context.Context, target health.Target) health.CheckResult {
	start := time.Now()

	// extract host from address (remove port if present)
	host := target.Address
	// Check if host includes port and extract hostname.
	if hostPart, _, err := net.SplitHostPort(host); err == nil {
		host = hostPart
	}

	// Select probe method based on mode.
	switch p.mode {
	// Handle explicit native ICMP mode.
	case config.ICMPModeNative:
		// Return native ping result.
		return p.nativePing(ctx, host, start)
	// Handle explicit TCP fallback mode.
	case config.ICMPModeFallback:
		// Return TCP fallback result.
		return p.tcpPing(ctx, host, start)
	// Handle auto-detect capability mode.
	case config.ICMPModeAuto:
		// Check if native ICMP is available.
		if p.hasNativeCapability {
			// Return native ping result.
			return p.nativePing(ctx, host, start)
		}
		// Return TCP fallback result.
		return p.tcpPing(ctx, host, start)
	// Handle unknown mode with TCP fallback.
	default:
		// Return TCP fallback result.
		return p.tcpPing(ctx, host, start)
	}
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
//   - health.CheckResult: the probe result with latency.
func (p *ICMPProber) tcpPing(ctx context.Context, host string, start time.Time) health.CheckResult {
	// validate and normalize port number
	tcpPort := p.tcpPort
	// Check if port is valid.
	if tcpPort <= 0 || tcpPort > shared.MaxValidPort {
		tcpPort = defaultTCPFallbackPort
	}
	address := net.JoinHostPort(host, strconv.Itoa(tcpPort))
	dialer := &net.Dialer{Timeout: p.timeout}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	latency := time.Since(start)
	// Check for connection failure.
	if err != nil {
		// Return failure result for unreachable host.
		return health.NewFailureCheckResult(latency, fmt.Sprintf("ping failed: %v", err), err)
	}
	_ = conn.Close()
	// Return success result for reachable host.
	return health.NewSuccessCheckResult(latency, fmt.Sprintf("ping %s: latency=%s (tcp fallback)", host, latency))
}
