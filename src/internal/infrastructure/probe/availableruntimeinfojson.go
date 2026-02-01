//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// AvailableRuntimeInfoJSON contains info about an available runtime on the host.
// It includes runtime name, socket path, version, and running status.
type AvailableRuntimeInfoJSON struct {
	Runtime    string `json:"runtime"`
	SocketPath string `json:"socket_path,omitempty"`
	Version    string `json:"version,omitempty"`
	IsRunning  bool   `json:"is_running"`
}
