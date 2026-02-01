//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.

// Package probe provides CGO bindings to the Rust probe library for unified
// cross-platform system metrics and resource quota management.
package probe

// ContainerInfo represents container detection results.
// It provides information about the containerized execution environment.
type ContainerInfo struct {
	// IsContainerized indicates whether running in a container.
	IsContainerized bool

	// Runtime is the detected container runtime.
	Runtime ContainerRuntime

	// ContainerID is the container ID if available.
	ContainerID string
}
