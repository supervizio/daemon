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
	// Match key to corresponding struct field
	switch key {
	// Anonymous memory pages
	case "anon":
		m.Anon = value
	// File-backed memory pages
	case "file":
		m.File = value
	// Kernel memory usage
	case "kernel":
		m.Kernel = value
	// Slab allocator memory
	case "slab":
		m.Slab = value
	// Socket buffer memory
	case "sock":
		m.Sock = value
	// Shared memory
	case "shmem":
		m.Shmem = value
	// Memory-mapped files
	case "mapped":
		m.Mapped = value
	// Dirty page cache
	case "dirty":
		m.Dirty = value
	// Page fault count
	case "pgfault":
		m.Pgfault = value
	// Major page fault count
	case "pgmajfault":
		m.Pgmajfault = value
	// Unknown or unsupported field
	default:
		// Field not recognized
		return false
	}
	// Field was set successfully
	return true
}
