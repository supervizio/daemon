//go:build cgo

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
