//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// UsageInfoJSON contains current resource usage.
// It includes memory, process count, and CPU utilization.
type UsageInfoJSON struct {
	MemoryBytes      uint64  `json:"memory_bytes"`
	MemoryLimitBytes uint64  `json:"memory_limit_bytes,omitempty"`
	PidsCurrent      uint64  `json:"pids_current"`
	PidsLimit        uint64  `json:"pids_limit,omitempty"`
	CPUPercent       float64 `json:"cpu_percent"`
	CPULimitPercent  float64 `json:"cpu_limit_percent,omitempty"`
}
