//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// NetInterfaceInfo contains network interface information.
// Used for JSON output in the --probe command.
type NetInterfaceInfo struct {
	// Name is the interface name (e.g., "eth0", "en0").
	Name string `dto:"out,api,pub" json:"name"`
	// MACAddress is the hardware MAC address.
	MACAddress string `dto:"out,api,pub" json:"mac_address"`
	// MTU is the maximum transmission unit.
	MTU uint32 `dto:"out,api,pub" json:"mtu"`
	// IsUp indicates if the interface is up.
	IsUp bool `dto:"out,api,pub" json:"is_up"`
	// IsLoopback indicates if this is a loopback interface.
	IsLoopback bool `dto:"out,api,pub" json:"is_loopback"`
}
