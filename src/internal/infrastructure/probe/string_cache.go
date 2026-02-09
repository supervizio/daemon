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
var globalCStringCache *cStringCache = &cStringCache{
	entries: make(map[uint64]cStringCacheEntry, maxCStringCacheSize),
	maxSize: maxCStringCacheSize, // Typical: ~20 devices + ~10 mount points + ~5 interfaces + misc
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
	// create FNV-1a hasher instance
	hasher := fnv.New64a()

	// convert C.char array to bytes and hash
	// we hash the actual bytes, not the pointer
	for _, c := range arr {
		// write byte to hash (error always nil for hash.Hash)
		_, _ = hasher.Write([]byte{byte(c)})
	}

	// return computed hash value
	return hasher.Sum64()
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
	// for dynamic strings, skip cache and convert directly
	if !stable {
		// return uncached string conversion
		return cCharArrayToString(arr)
	}

	// compute hash for cache lookup
	hash := hashCCharArray(arr)

	// try read lock first (common case: cache hit)
	globalCStringCache.mu.RLock()
	// check if entry exists in cache
	if entry, found := globalCStringCache.entries[hash]; found {
		globalCStringCache.mu.RUnlock()
		// return cached string value
		return entry.value
	}
	globalCStringCache.mu.RUnlock()

	// cache miss - convert string
	str := cCharArrayToString(arr)

	// write lock to update cache
	globalCStringCache.mu.Lock()
	defer globalCStringCache.mu.Unlock()

	// evict non-stable entries if cache is full.
	if len(globalCStringCache.entries) >= globalCStringCache.maxSize {
		evictNonStableEntries()
	}

	// store in cache
	globalCStringCache.entries[hash] = cStringCacheEntry{
		key:    hash,
		value:  str,
		stable: stable,
	}

	// return converted string
	return str
}

// evictNonStableEntries removes non-stable entries from the cache.
// If the cache is still full after evicting non-stable entries,
// all entries are cleared. Must be called with write lock held.
func evictNonStableEntries() {
	// remove all non-stable (dynamic) entries first.
	for k, e := range globalCStringCache.entries {
		// skip stable entries that should remain cached.
		if !e.stable {
			delete(globalCStringCache.entries, k)
		}
	}
	// if still full after eviction, reset entire cache.
	if len(globalCStringCache.entries) >= globalCStringCache.maxSize {
		globalCStringCache.entries = make(map[uint64]cStringCacheEntry, maxCStringCacheSize)
	}
}

// clearCStringCache clears the C string cache.
// This is primarily for testing purposes.
func clearCStringCache() {
	globalCStringCache.mu.Lock()
	defer globalCStringCache.mu.Unlock()
	globalCStringCache.entries = make(map[uint64]cStringCacheEntry, maxCStringCacheSize)
}

// getCStringCacheStats returns cache statistics for monitoring.
//
// Returns:
//   - size: number of entries in cache
//   - capacity: maximum cache size
func getCStringCacheStats() (size, capacity int) {
	globalCStringCache.mu.RLock()
	defer globalCStringCache.mu.RUnlock()
	// return current size and max capacity
	return len(globalCStringCache.entries), globalCStringCache.maxSize
}

// Test helpers - exported functions for testing cache behavior.

// ClearCStringCacheForTest clears the cache (test helper).
func ClearCStringCacheForTest() {
	clearCStringCache()
}

// GetCStringCacheStatsForTest returns cache stats (test helper).
//
// Returns:
//   - size: number of entries in cache
//   - capacity: maximum cache size
func GetCStringCacheStatsForTest() (size, capacity int) {
	// return cache statistics for testing
	return getCStringCacheStats()
}

// CCharArrayToStringCachedForTest converts with caching (test helper).
//
// Params:
//   - arr: C char array to convert
//   - stable: whether this is a stable string that should be cached
//
// Returns:
//   - string: the converted Go string
func CCharArrayToStringCachedForTest(arr []C.char, stable bool) string {
	return cCharArrayToStringCached(arr, stable)
}
