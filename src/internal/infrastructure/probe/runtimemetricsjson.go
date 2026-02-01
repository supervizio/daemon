//go:build cgo

package probe

// RuntimeMetricsJSON contains full runtime detection information.
// It includes container runtime, orchestrator, and available runtimes.
type RuntimeMetricsJSON struct {
	IsContainerized   bool                       `json:"is_containerized"`
	ContainerRuntime  string                     `json:"container_runtime,omitempty"`
	Orchestrator      string                     `json:"orchestrator,omitempty"`
	ContainerID       string                     `json:"container_id,omitempty"`
	WorkloadID        string                     `json:"workload_id,omitempty"`
	WorkloadName      string                     `json:"workload_name,omitempty"`
	Namespace         string                     `json:"namespace,omitempty"`
	AvailableRuntimes []AvailableRuntimeInfoJSON `json:"available_runtimes,omitempty"`
}
