// Package metrics provides domain types for system and process metrics collection.
package metrics

// SystemMemoryInput contains the input parameters for creating SystemMemory.
//
// This struct groups the parameters needed to construct a SystemMemory value object.
type SystemMemoryInput struct {
	// Total is the total physical RAM in bytes.
	Total uint64
	// Available is the memory available for starting new applications in bytes.
	Available uint64
	// Free is the free memory in bytes (MemFree from /proc/meminfo).
	Free uint64
	// Cached is the page cache memory in bytes.
	Cached uint64
	// Buffers is the buffer memory in bytes.
	Buffers uint64
	// SwapTotal is the total swap space in bytes.
	SwapTotal uint64
	// SwapUsed is the used swap space in bytes.
	SwapUsed uint64
	// SwapFree is the free swap space in bytes.
	SwapFree uint64
	// Shared is the shared memory in bytes (Shmem from /proc/meminfo).
	Shared uint64
}
