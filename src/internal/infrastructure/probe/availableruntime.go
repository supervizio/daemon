//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// AvailableRuntime describes a runtime available on the host.
// It includes connection details and version information.
type AvailableRuntime struct {
	// Runtime type.
	Runtime RuntimeType

	// SocketPath is the Unix socket path (empty if not available).
	SocketPath string

	// Version is the version string (empty if not available).
	Version string

	// IsRunning indicates whether the runtime is currently responsive.
	IsRunning bool
}
