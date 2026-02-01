//go:build cgo

package probe

// ContainerMetricsJSON contains container detection information.
// It indicates containerization status, runtime, and container ID.
type ContainerMetricsJSON struct {
	IsContainerized bool   `json:"is_containerized"`
	Runtime         string `json:"runtime,omitempty"`
	ContainerID     string `json:"container_id,omitempty"`
}
