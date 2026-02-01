//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// NetStatsInfo contains network interface statistics.
// Used for JSON output in the --probe command.
type NetStatsInfo struct {
	// Interface is the interface name.
	Interface string `dto:"out,api,pub" json:"interface"`
	// RxBytes is the number of bytes received.
	RxBytes uint64 `dto:"out,api,pub" json:"rx_bytes"`
	// RxPackets is the number of packets received.
	RxPackets uint64 `dto:"out,api,pub" json:"rx_packets"`
	// RxErrors is the number of receive errors.
	RxErrors uint64 `dto:"out,api,pub" json:"rx_errors"`
	// RxDrops is the number of received packets dropped.
	RxDrops uint64 `dto:"out,api,pub" json:"rx_drops"`
	// TxBytes is the number of bytes transmitted.
	TxBytes uint64 `dto:"out,api,pub" json:"tx_bytes"`
	// TxPackets is the number of packets transmitted.
	TxPackets uint64 `dto:"out,api,pub" json:"tx_packets"`
	// TxErrors is the number of transmit errors.
	TxErrors uint64 `dto:"out,api,pub" json:"tx_errors"`
	// TxDrops is the number of transmitted packets dropped.
	TxDrops uint64 `dto:"out,api,pub" json:"tx_drops"`
}
