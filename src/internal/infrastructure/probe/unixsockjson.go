//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// UnixSockJSON contains information about a Unix socket.
// It includes path, type, state, and owning process.
type UnixSockJSON struct {
	Path        string `json:"path"`
	Type        string `json:"type"`
	State       string `json:"state"`
	PID         int32  `json:"pid"`
	ProcessName string `json:"process_name,omitempty"`
}
