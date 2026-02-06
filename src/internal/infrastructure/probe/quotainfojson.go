//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// QuotaInfoJSON contains resource limit information.
// It includes CPU, memory, process, and file descriptor limits.
type QuotaInfoJSON struct {
	CPUQuotaUs       uint64 `json:"cpu_quota_us,omitempty"`
	CPUPeriodUs      uint64 `json:"cpu_period_us,omitempty"`
	MemoryLimitBytes uint64 `json:"memory_limit_bytes,omitempty"`
	PidsLimit        uint64 `json:"pids_limit,omitempty"`
	NofileLimit      uint64 `json:"nofile_limit,omitempty"`
}
