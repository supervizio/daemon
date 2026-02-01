//go:build cgo

package probe

// ProcessMetricsJSON contains information about the current process and system processes.
// It includes PID, process count, and resource usage for top processes.
type ProcessMetricsJSON struct {
	CurrentPID   int32             `json:"current_pid"`
	ProcessCount int               `json:"process_count"`
	TopProcesses []ProcessInfoJSON `json:"top_processes,omitempty"`
}
