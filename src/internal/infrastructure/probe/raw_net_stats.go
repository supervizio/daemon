// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// RawNetStatsData holds raw network stats data for Go-only building.
// This struct mirrors C data without requiring CGO for testing.
type RawNetStatsData struct {
	// Interface is the interface name.
	Interface string
	// RxBytes is received bytes.
	RxBytes uint64
	// RxPackets is received packets.
	RxPackets uint64
	// RxErrors is receive errors.
	RxErrors uint64
	// RxDrops is receive drops.
	RxDrops uint64
	// TxBytes is transmitted bytes.
	TxBytes uint64
	// TxPackets is transmitted packets.
	TxPackets uint64
	// TxErrors is transmit errors.
	TxErrors uint64
	// TxDrops is transmit drops.
	TxDrops uint64
}
