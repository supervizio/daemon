//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// ProcessInfoJSON contains comprehensive information about a process.
// It includes CPU, memory, I/O, thread, and file descriptor statistics.
type ProcessInfoJSON struct {
	PID              int32   `json:"pid"`
	CPUPercent       float64 `json:"cpu_percent"`
	MemoryRSSBytes   uint64  `json:"memory_rss_bytes"`
	MemoryVMSBytes   uint64  `json:"memory_vms_bytes"`
	MemoryPercent    float64 `json:"memory_percent"`
	NumThreads       uint32  `json:"num_threads"`
	NumFDs           uint32  `json:"num_fds"`
	ReadBytesPerSec  uint64  `json:"read_bytes_per_sec"`
	WriteBytesPerSec uint64  `json:"write_bytes_per_sec"`
	State            string  `json:"state"`
}
