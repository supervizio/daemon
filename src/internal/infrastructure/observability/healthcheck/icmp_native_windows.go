//go:build windows

// Package healthcheck provides infrastructure adapters for service probing.
package healthcheck

import (
	"context"
	"time"

	"github.com/kodflow/daemon/internal/domain/health"
)

// nativePing is not supported on Windows, always falls back to TCP.
//
// Params:
//   - ctx: context for cancellation and timeout.
//   - host: target hostname or IP address.
//   - start: probe start time for latency calculation.
//
// Returns:
//   - health.CheckResult: the probe result (always uses TCP fallback).
func (p *ICMPProber) nativePing(ctx context.Context, host string, start time.Time) health.CheckResult {
	// Windows does not support raw ICMP sockets without admin privileges
	// fall back to TCP ping
	return p.tcpPing(ctx, host, start)
}

// detectICMPCapability always returns false on Windows.
//
// Returns:
//   - bool: always false on Windows.
func detectICMPCapability() bool {
	// raw ICMP sockets require admin privileges on Windows
	// always use TCP fallback
	return false
}
