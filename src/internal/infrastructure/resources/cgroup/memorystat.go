//go:build linux

// Package cgroup provides adapters for reading cgroup metrics.
package cgroup

// MemoryStat contains parsed memory statistics.
// It represents memory usage metrics from cgroup v2 memory.stat file.
type MemoryStat struct {
	// Anon is anonymous memory in bytes.
	Anon uint64
	// File is file-backed memory in bytes.
	File uint64
	// Kernel is kernel memory in bytes.
	Kernel uint64
	// Slab is slab memory in bytes.
	Slab uint64
	// Sock is socket buffer memory in bytes.
	Sock uint64
	// Shmem is shared memory in bytes.
	Shmem uint64
	// Mapped is mapped memory in bytes.
	Mapped uint64
	// Dirty is dirty pages in bytes.
	Dirty uint64
	// Pgfault is page fault count.
	Pgfault uint64
	// Pgmajfault is major page fault count.
	Pgmajfault uint64
}

// setField sets the appropriate field based on the key.
// Uses switch instead of map to avoid heap allocations on every call.
//
// Params:
//   - key: field name from memory.stat
//   - value: parsed uint64 value
//
// Returns:
//   - bool: true if key was recognized and field was set
func (m *MemoryStat) setField(key string, value uint64) bool {
	switch key {
	case "anon":
		m.Anon = value
	case "file":
		m.File = value
	case "kernel":
		m.Kernel = value
	case "slab":
		m.Slab = value
	case "sock":
		m.Sock = value
	case "shmem":
		m.Shmem = value
	case "mapped":
		m.Mapped = value
	case "dirty":
		m.Dirty = value
	case "pgfault":
		m.Pgfault = value
	case "pgmajfault":
		m.Pgmajfault = value
	default:
		return false
	}
	return true
}
