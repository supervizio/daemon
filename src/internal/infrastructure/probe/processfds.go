//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// ProcessFDs contains file descriptor count for a process.
//
// This value object captures the number of open file descriptors for a specific
// process. File descriptors include open files, sockets, pipes, and other I/O resources.
type ProcessFDs struct {
	// PID is the process identifier.
	PID int
	// Count is the number of open file descriptors.
	Count uint32
}
