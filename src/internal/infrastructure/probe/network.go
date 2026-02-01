//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for unified
// cross-platform system metrics and resource quota management.
package probe

/*
#include "probe.h"
*/
import "C"

import (
	"context"
	"time"
	"unsafe"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// NetworkCollector provides network metrics via the Rust probe library.
// It implements the metrics.NetworkCollector interface for network statistics.
type NetworkCollector struct{}

// NewNetworkCollector creates a new network collector.
//
// Returns:
//   - *NetworkCollector: new network collector instance
func NewNetworkCollector() *NetworkCollector {
	// Return a new empty collector instance.
	return &NetworkCollector{}
}

// ListInterfaces returns all network interfaces.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//
// Returns:
//   - []metrics.NetInterface: list of network interfaces
//   - error: nil on success, error if probe not initialized or collection fails
func (n *NetworkCollector) ListInterfaces(ctx context.Context) ([]metrics.NetInterface, error) {
	// Check if context has been cancelled before expensive FFI call.
	if err := checkContext(ctx); err != nil {
		// Return nil slice with context error.
		return nil, err
	}
	// Verify probe library is initialized before collecting.
	if err := checkInitialized(); err != nil {
		// Return nil slice with initialization error.
		return nil, err
	}

	var list C.NetInterfaceList
	result := C.probe_list_net_interfaces(&list)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return nil slice with collection error.
		return nil, err
	}
	defer C.probe_free_net_interface_list(&list)

	count := int(list.count)
	ifaces := make([]metrics.NetInterface, 0, count)
	items := unsafe.Slice(list.items, count)
	// Iterate through each network interface from the Rust library.
	for _, item := range items {
		var flags []string
		// Check if interface is up.
		if item.is_up {
			flags = append(flags, "up")
		}
		// Check if interface is loopback.
		if item.is_loopback {
			flags = append(flags, "loopback")
		}

		ifaces = append(ifaces, metrics.NetInterface{
			Name:         cCharArrayToString(item.name[:]),
			HardwareAddr: cCharArrayToString(item.mac_address[:]),
			MTU:          int(item.mtu),
			Flags:        flags,
		})
	}

	// Return collected network interfaces.
	return ifaces, nil
}

// CollectStats collects statistics for a specific interface.
//
// Params:
//   - ctx: context for cancellation
//   - iface: name of the interface to collect stats for
//
// Returns:
//   - metrics.NetStats: statistics for the specified interface
//   - error: nil on success, ErrNotFound if interface not found
func (n *NetworkCollector) CollectStats(ctx context.Context, iface string) (metrics.NetStats, error) {
	all, err := n.CollectAllStats(ctx)
	// Check if collecting all stats failed.
	if err != nil {
		// Return empty stats with collection error.
		return metrics.NetStats{}, err
	}

	// Search for the requested interface in the collected stats.
	for _, stat := range all {
		// Check if this is the requested interface.
		if stat.Interface == iface {
			// Return the matching interface stats.
			return stat, nil
		}
	}

	// Return error if interface was not found.
	return metrics.NetStats{}, ErrNotFound
}

// CollectAllStats collects statistics for all interfaces.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//
// Returns:
//   - []metrics.NetStats: statistics for all network interfaces
//   - error: nil on success, error if probe not initialized or collection fails
func (n *NetworkCollector) CollectAllStats(ctx context.Context) ([]metrics.NetStats, error) {
	// Check if context has been cancelled before expensive FFI call.
	if err := checkContext(ctx); err != nil {
		// Return nil slice with context error.
		return nil, err
	}
	// Verify probe library is initialized before collecting.
	if err := checkInitialized(); err != nil {
		// Return nil slice with initialization error.
		return nil, err
	}

	var list C.NetStatsList
	result := C.probe_collect_net_stats(&list)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return nil slice with collection error.
		return nil, err
	}
	defer C.probe_free_net_stats_list(&list)

	count := int(list.count)
	stats := make([]metrics.NetStats, 0, count)
	items := unsafe.Slice(list.items, count)
	// Iterate through each interface's statistics.
	for _, item := range items {
		// C struct uses 'interface' which is renamed to '_interface' by CGO
		stats = append(stats, metrics.NetStats{
			Interface:   cCharArrayToString(item._interface[:]),
			BytesRecv:   uint64(item.rx_bytes),
			BytesSent:   uint64(item.tx_bytes),
			PacketsRecv: uint64(item.rx_packets),
			PacketsSent: uint64(item.tx_packets),
			ErrorsIn:    uint64(item.rx_errors),
			ErrorsOut:   uint64(item.tx_errors),
			DropsIn:     uint64(item.rx_drops),
			DropsOut:    uint64(item.tx_drops),
			Timestamp:   time.Now(),
		})
	}

	// Return collected network statistics.
	return stats, nil
}
