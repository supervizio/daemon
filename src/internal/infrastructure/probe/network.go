//go:build cgo

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
type NetworkCollector struct{}

// NewNetworkCollector creates a new network collector.
func NewNetworkCollector() *NetworkCollector {
	return &NetworkCollector{}
}

// ListInterfaces returns all network interfaces.
//
//nolint:gocritic // dupSubExpr false positive from CGO list operations
func (n *NetworkCollector) ListInterfaces(_ context.Context) ([]metrics.NetInterface, error) {
	if err := checkInitialized(); err != nil {
		return nil, err
	}

	var list C.NetInterfaceList
	result := C.probe_list_net_interfaces(&list)
	if err := resultToError(result); err != nil {
		return nil, err
	}
	defer C.probe_free_net_interface_list(&list)

	ifaces := make([]metrics.NetInterface, list.count)
	items := unsafe.Slice(list.items, list.count)
	for i, item := range items {
		var flags []string
		if item.is_up {
			flags = append(flags, "up")
		}
		if item.is_loopback {
			flags = append(flags, "loopback")
		}

		ifaces[i] = metrics.NetInterface{
			Name:         cCharArrayToString(item.name[:]),
			HardwareAddr: cCharArrayToString(item.mac_address[:]),
			MTU:          int(item.mtu),
			Flags:        flags,
		}
	}

	return ifaces, nil
}

// CollectStats collects statistics for a specific interface.
func (n *NetworkCollector) CollectStats(ctx context.Context, iface string) (metrics.NetStats, error) {
	all, err := n.CollectAllStats(ctx)
	if err != nil {
		return metrics.NetStats{}, err
	}

	for _, stat := range all {
		if stat.Interface == iface {
			return stat, nil
		}
	}

	return metrics.NetStats{}, ErrNotFound
}

// CollectAllStats collects statistics for all interfaces.
//
//nolint:gocritic // dupSubExpr false positive from CGO list operations
func (n *NetworkCollector) CollectAllStats(_ context.Context) ([]metrics.NetStats, error) {
	if err := checkInitialized(); err != nil {
		return nil, err
	}

	var list C.NetStatsList
	result := C.probe_collect_net_stats(&list)
	if err := resultToError(result); err != nil {
		return nil, err
	}
	defer C.probe_free_net_stats_list(&list)

	stats := make([]metrics.NetStats, list.count)
	items := unsafe.Slice(list.items, list.count)
	for i, item := range items {
		// C struct uses 'interface' which is renamed to '_interface' by CGO
		stats[i] = metrics.NetStats{
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
		}
	}

	return stats, nil
}
