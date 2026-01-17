// Package scratch provides a minimal metrics collector for environments
// without /proc filesystem access (e.g., scratch containers, Windows).
package scratch

import (
	"context"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// NetworkCollector provides minimal network metrics collection.
// All methods return ErrNotSupported as network metrics are unavailable
// in scratch environments without /proc/net or netlink access.
type NetworkCollector struct{}

// NewNetworkCollector creates a new network collector for scratch environments.
//
// Returns:
//   - *NetworkCollector: initialized network collector
func NewNetworkCollector() *NetworkCollector {
	// Return zero-value network collector
	return &NetworkCollector{}
}

// ListInterfaces returns an error as interface enumeration is not available.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - []metrics.NetInterface: nil slice
//   - error: ErrNotSupported or context error
func (n *NetworkCollector) ListInterfaces(ctx context.Context) ([]metrics.NetInterface, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return nil, ctx.Err()
	}
	// Not supported in scratch mode.
	return nil, ErrNotSupported
}

// CollectStats returns an error as interface statistics are not available.
//
// Params:
//   - ctx: context for cancellation
//   - iface: interface name to collect stats for
//
// Returns:
//   - metrics.NetStats: zero value with Interface set for context
//   - error: ErrNotSupported, ErrEmptyInterface, or context error
func (n *NetworkCollector) CollectStats(ctx context.Context, iface string) (metrics.NetStats, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.NetStats{}, ctx.Err()
	}
	// Validate interface name before returning not supported.
	if iface == "" {
		// Return error for empty interface name.
		return metrics.NetStats{}, ErrEmptyInterface
	}
	// Not supported in scratch mode, return interface for context.
	return metrics.NetStats{Interface: iface}, ErrNotSupported
}

// CollectAllStats returns an error as interface enumeration is not available.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - []metrics.NetStats: nil slice
//   - error: ErrNotSupported or context error
func (n *NetworkCollector) CollectAllStats(ctx context.Context) ([]metrics.NetStats, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return nil, ctx.Err()
	}
	// Not supported in scratch mode.
	return nil, ErrNotSupported
}
