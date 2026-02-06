//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// ListenInfoJSON contains information about a listening port.
// It includes protocol, address, port, and owning process.
type ListenInfoJSON struct {
	Protocol    string `json:"protocol"`
	Address     string `json:"address"`
	Port        uint16 `json:"port"`
	PID         int32  `json:"pid"`
	ProcessName string `json:"process_name,omitempty"`
}
