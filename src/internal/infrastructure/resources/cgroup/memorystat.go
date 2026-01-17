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

// fieldMap returns the field mapping for MemoryStat parsing.
//
// Returns:
//   - map[string]*uint64: field name to struct field pointer mapping
func (m *MemoryStat) fieldMap() map[string]*uint64 {
	// Return mapping of memory.stat keys to struct fields
	return map[string]*uint64{
		"anon":       &m.Anon,
		"file":       &m.File,
		"kernel":     &m.Kernel,
		"slab":       &m.Slab,
		"sock":       &m.Sock,
		"shmem":      &m.Shmem,
		"mapped":     &m.Mapped,
		"dirty":      &m.Dirty,
		"pgfault":    &m.Pgfault,
		"pgmajfault": &m.Pgmajfault,
	}
}
