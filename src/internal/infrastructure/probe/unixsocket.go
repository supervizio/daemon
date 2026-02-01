//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// UnixSocket represents a Unix domain socket with process information.
// It includes socket path, type, state, and owning process details.
type UnixSocket struct {
	Path        string
	SocketType  string
	State       SocketState
	PID         int32
	ProcessName string
	Inode       uint64
}
