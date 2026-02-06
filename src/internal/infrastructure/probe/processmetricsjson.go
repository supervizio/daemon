//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// ProcessMetricsJSON contains information about the current process and system processes.
// It includes PID, process count, and resource usage for top processes.
type ProcessMetricsJSON struct {
	CurrentPID   int32             `json:"current_pid"`
	ProcessCount int               `json:"process_count"`
	TopProcesses []ProcessInfoJSON `json:"top_processes,omitempty"`
}
