// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// RawNetInterfaceData holds raw network interface data for Go-only building.
// This struct mirrors C data without requiring CGO for testing.
type RawNetInterfaceData struct {
	// Name is the interface name.
	Name string
	// MACAddress is the hardware address.
	MACAddress string
	// MTU is the maximum transmission unit.
	MTU uint32
	// IsUp indicates if interface is up.
	IsUp bool
	// IsLoopback indicates if interface is loopback.
	IsLoopback bool
}
