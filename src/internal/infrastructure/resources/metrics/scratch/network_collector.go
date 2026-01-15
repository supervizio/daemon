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
//   - _: context (unused in scratch mode)
//
// Returns:
//   - []metrics.NetInterface: nil slice
//   - error: ErrNotSupported
func (n *NetworkCollector) ListInterfaces(_ context.Context) ([]metrics.NetInterface, error) {
	// Not supported in scratch mode
	return nil, ErrNotSupported
}

// CollectStats returns an error as interface statistics are not available.
//
// Params:
//   - _: context (unused in scratch mode)
//   - _: interface name (unused in scratch mode)
//
// Returns:
//   - metrics.NetStats: zero value
//   - error: ErrNotSupported
func (n *NetworkCollector) CollectStats(_ context.Context, _ string) (metrics.NetStats, error) {
	// Not supported in scratch mode
	return metrics.NetStats{}, ErrNotSupported
}

// CollectAllStats returns an error as interface enumeration is not available.
//
// Params:
//   - _: context (unused in scratch mode)
//
// Returns:
//   - []metrics.NetStats: nil slice
//   - error: ErrNotSupported
func (n *NetworkCollector) CollectAllStats(_ context.Context) ([]metrics.NetStats, error) {
	// Not supported in scratch mode
	return nil, ErrNotSupported
}
