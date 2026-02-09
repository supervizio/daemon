//go:build cgo

// Package probe provides CGO bindings to the Rust probe library.
package probe

/*
#include <stdlib.h>
*/
import "C"
import (
	"hash/fnv"
	"sync"
)

// cStringCacheEntry represents a cached C string conversion.
type cStringCacheEntry struct {
	key    uint64 // FNV-1a hash of C string bytes
	value  string // Cached Go string
	stable bool   // Whether this is a stable string (device name, MAC, etc.)
}

// cStringCache provides thread-safe caching for C string conversions.
// This is particularly effective for stable strings like device names,
// MAC addresses, mount points, and filesystem types that don't change.
type cStringCache struct {
	mu      sync.RWMutex
	entries map[uint64]cStringCacheEntry
	maxSize int
}

// Global cache instance for C string conversions.
// Initialized with reasonable size for typical system (128 entries).
var globalCStringCache = &cStringCache{
	entries: make(map[uint64]cStringCacheEntry),
	maxSize: 128, // Typical: ~20 devices + ~10 mount points + ~5 interfaces + misc
}

// hashCCharArray computes FNV-1a hash of a C char array.
// This is used as the cache key.
//
// Params:
//   - arr: C char array to hash
//
// Returns:
//   - uint64: FNV-1a hash of the array bytes
func hashCCharArray(arr []C.char) uint64 {
	h := fnv.New64a()

	// Convert C.char array to bytes and hash
	// We hash the actual bytes, not the pointer
	for _, c := range arr {
		// Write byte to hash (error always nil for hash.Hash)
		_, _ = h.Write([]byte{byte(c)})
	}

	return h.Sum64()
}

// cCharArrayToStringCached converts a C char array to a Go string with caching.
// This function caches stable strings (device names, MACs, mount points, etc.)
// to reduce allocations. For dynamic strings (IPs, process names), it falls
// back to uncached conversion.
//
// Params:
//   - arr: C char array to convert
//   - stable: whether this is a stable string that should be cached
//
// Returns:
//   - string: the converted Go string
func cCharArrayToStringCached(arr []C.char, stable bool) string {
	// For dynamic strings, skip cache and convert directly
	if !stable {
		return cCharArrayToString(arr)
	}

	// Compute hash for cache lookup
	hash := hashCCharArray(arr)

	// Try read lock first (common case: cache hit)
	globalCStringCache.mu.RLock()
	if entry, found := globalCStringCache.entries[hash]; found {
		globalCStringCache.mu.RUnlock()
		return entry.value
	}
	globalCStringCache.mu.RUnlock()

	// Cache miss - convert string
	str := cCharArrayToString(arr)

	// Write lock to update cache
	globalCStringCache.mu.Lock()
	defer globalCStringCache.mu.Unlock()

	// Check size limit (simple eviction: clear cache if full)
	if len(globalCStringCache.entries) >= globalCStringCache.maxSize {
		// Simple eviction: clear all non-stable entries
		// This is acceptable because stable entries should fill most of the cache
		for k, e := range globalCStringCache.entries {
			if !e.stable {
				delete(globalCStringCache.entries, k)
			}
		}

		// If still full, clear everything (shouldn't happen with stable strings)
		if len(globalCStringCache.entries) >= globalCStringCache.maxSize {
			globalCStringCache.entries = make(map[uint64]cStringCacheEntry)
		}
	}

	// Store in cache
	globalCStringCache.entries[hash] = cStringCacheEntry{
		key:    hash,
		value:  str,
		stable: stable,
	}

	return str
}

// clearCStringCache clears the C string cache.
// This is primarily for testing purposes.
func clearCStringCache() {
	globalCStringCache.mu.Lock()
	defer globalCStringCache.mu.Unlock()
	globalCStringCache.entries = make(map[uint64]cStringCacheEntry)
}

// getCStringCacheStats returns cache statistics for monitoring.
//
// Returns:
//   - size: number of entries in cache
//   - capacity: maximum cache size
func getCStringCacheStats() (size int, capacity int) {
	globalCStringCache.mu.RLock()
	defer globalCStringCache.mu.RUnlock()
	return len(globalCStringCache.entries), globalCStringCache.maxSize
}

// Test helpers - exported functions for testing cache behavior.

// ClearCStringCacheForTest clears the cache (test helper).
func ClearCStringCacheForTest() {
	clearCStringCache()
}

// GetCStringCacheStatsForTest returns cache stats (test helper).
func GetCStringCacheStatsForTest() (size int, capacity int) {
	return getCStringCacheStats()
}

// CCharArrayToStringCachedForTest converts with caching (test helper).
func CCharArrayToStringCachedForTest(arr []C.char, stable bool) string {
	return cCharArrayToStringCached(arr, stable)
}
