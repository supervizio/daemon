//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// NetInterfaceJSON contains information about a network interface.
// It includes name, MAC address, MTU, and status flags.
type NetInterfaceJSON struct {
	Name       string   `json:"name"`
	MACAddress string   `json:"mac_address"`
	MTU        uint32   `json:"mtu"`
	IsUp       bool     `json:"is_up"`
	IsLoopback bool     `json:"is_loopback"`
	Flags      []string `json:"flags,omitempty"`
}
