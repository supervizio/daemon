//go:build cgo

package probe

// AvailableRuntimeInfoJSON contains info about an available runtime on the host.
// It includes runtime name, socket path, version, and running status.
type AvailableRuntimeInfoJSON struct {
	Runtime    string `json:"runtime"`
	SocketPath string `json:"socket_path,omitempty"`
	Version    string `json:"version,omitempty"`
	IsRunning  bool   `json:"is_running"`
}
