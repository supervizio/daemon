//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// ContainerMetricsJSON contains container detection information.
// It indicates containerization status, runtime, and container ID.
type ContainerMetricsJSON struct {
	IsContainerized bool   `json:"is_containerized"`
	Runtime         string `json:"runtime,omitempty"`
	ContainerID     string `json:"container_id,omitempty"`
}
