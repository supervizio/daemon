// Package metrics provides domain types for system and process metrics collection.
package metrics

import "slices"

// NetInterface represents a network interface.
//
// Provides information about network interface configuration including
// hardware address, MTU, operational state, and assigned IP addresses.
type NetInterface struct {
	// Name is the interface name (e.g., "eth0", "en0", "lo").
	Name string
	// Index is the interface index.
	Index int
	// HardwareAddr is the MAC address (e.g., "00:11:22:33:44:55").
	HardwareAddr string
	// MTU is the Maximum Transmission Unit in bytes.
	MTU int
	// Flags describes the interface state (e.g., "up", "broadcast", "multicast").
	Flags []string
	// Addresses are the IP addresses assigned to this interface.
	Addresses []string
}

// NewNetInterface creates a new NetInterface with essential fields.
//
// Params:
//   - name: interface name (e.g., "eth0", "en0", "lo")
//   - index: interface index
//
// Returns:
//   - NetInterface: new network interface instance
func NewNetInterface(name string, index int) NetInterface {
	// Return interface with essential fields set.
	return NetInterface{
		Name:  name,
		Index: index,
	}
}

// IsUp returns true if the interface is up.
//
// Returns:
//   - bool: true if interface is up, false otherwise
func (n NetInterface) IsUp() bool {
	// Check if "up" flag is present in the flags list.
	return slices.Contains(n.Flags, "up")
}

// IsLoopback returns true if this is a loopback interface.
//
// Returns:
//   - bool: true if interface is loopback, false otherwise
func (n NetInterface) IsLoopback() bool {
	// Check if "loopback" flag is present in the flags list.
	return slices.Contains(n.Flags, "loopback")
}
