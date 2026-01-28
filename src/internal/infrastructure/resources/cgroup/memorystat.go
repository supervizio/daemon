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

// fieldSetters maps field names to setter functions.
// Using a map avoids deep switch/case complexity.
var fieldSetters map[string]func(*MemoryStat, uint64) = map[string]func(*MemoryStat, uint64){
	"anon":       func(m *MemoryStat, v uint64) { m.Anon = v },
	"file":       func(m *MemoryStat, v uint64) { m.File = v },
	"kernel":     func(m *MemoryStat, v uint64) { m.Kernel = v },
	"slab":       func(m *MemoryStat, v uint64) { m.Slab = v },
	"sock":       func(m *MemoryStat, v uint64) { m.Sock = v },
	"shmem":      func(m *MemoryStat, v uint64) { m.Shmem = v },
	"mapped":     func(m *MemoryStat, v uint64) { m.Mapped = v },
	"dirty":      func(m *MemoryStat, v uint64) { m.Dirty = v },
	"pgfault":    func(m *MemoryStat, v uint64) { m.Pgfault = v },
	"pgmajfault": func(m *MemoryStat, v uint64) { m.Pgmajfault = v },
}

// setField applies a value to the field matching key via map lookup.
//
// Params:
//   - key: field name from memory.stat (e.g., "anon", "file", "kernel")
//   - value: parsed numeric value in bytes
//
// Returns:
//   - bool: true if key matched a known field, false for unrecognized keys
func (m *MemoryStat) setField(key string, value uint64) bool {
	setter, ok := fieldSetters[key]
	// Unknown field; skip silently for forward compatibility.
	if !ok {
		return false
	}
	setter(m, value)
	return true
}
