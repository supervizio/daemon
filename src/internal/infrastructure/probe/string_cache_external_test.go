//go:build cgo

package probe_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
#include <stdlib.h>
*/
import "C"

// TestCStringCachingBasic verifies basic caching behavior.
func TestCStringCachingBasic(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "stable strings are cached"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear cache before test
			probe.ClearCStringCacheForTest()

			// Create a C char array
			input := []C.char{C.char('t'), C.char('e'), C.char('s'), C.char('t')}

			// First call should cache
			result1 := probe.CCharArrayToStringCachedForTest(input, true)
			assert.Equal(t, "test", result1)

			// Check cache size increased
			size, _ := probe.GetCStringCacheStatsForTest()
			assert.Equal(t, 1, size, "Cache should have 1 entry")

			// Second call should hit cache (same string)
			result2 := probe.CCharArrayToStringCachedForTest(input, true)
			assert.Equal(t, "test", result2)

			// Cache size should remain the same
			size, _ = probe.GetCStringCacheStatsForTest()
			assert.Equal(t, 1, size, "Cache should still have 1 entry")

			// Different string should add new entry
			input2 := []C.char{C.char('f'), C.char('o'), C.char('o')}
			result3 := probe.CCharArrayToStringCachedForTest(input2, true)
			assert.Equal(t, "foo", result3)

			// Cache size should increase
			size, _ = probe.GetCStringCacheStatsForTest()
			assert.Equal(t, 2, size, "Cache should have 2 entries")
		})
	}
}

// TestCStringCachingUnstable verifies unstable strings bypass cache.
func TestCStringCachingUnstable(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "unstable strings bypass cache"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear cache before test
			probe.ClearCStringCacheForTest()

			// Create a C char array
			input := []C.char{C.char('d'), C.char('y'), C.char('n')}

			// Call with stable=false should bypass cache
			result1 := probe.CCharArrayToStringCachedForTest(input, false)
			assert.Equal(t, "dyn", result1)

			// Check cache is still empty
			size, _ := probe.GetCStringCacheStatsForTest()
			assert.Equal(t, 0, size, "Cache should be empty for unstable strings")

			// Multiple calls should not cache
			result2 := probe.CCharArrayToStringCachedForTest(input, false)
			assert.Equal(t, "dyn", result2)

			size, _ = probe.GetCStringCacheStatsForTest()
			assert.Equal(t, 0, size, "Cache should still be empty")
		})
	}
}

// TestCStringCachingConcurrency verifies thread-safe cache access.
func TestCStringCachingConcurrency(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "concurrent cache access is safe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear cache before test
			probe.ClearCStringCacheForTest()

			const goroutines = 10
			const iterations = 100

			done := make(chan bool, goroutines)

			for i := 0; i < goroutines; i++ {
				go func(id int) {
					for j := 0; j < iterations; j++ {
						// Create different strings per goroutine
						input := []C.char{
							C.char('g'),
							C.char('o'),
							C.char(byte('0' + id)),
						}

						// Cache the string
						result := probe.CCharArrayToStringCachedForTest(input, true)
						require.NotEmpty(t, result)
					}
					done <- true
				}(i)
			}

			// Wait for all goroutines
			for i := 0; i < goroutines; i++ {
				<-done
			}

			// Cache should have entries (one per unique string)
			size, capacity := probe.GetCStringCacheStatsForTest()
			assert.Greater(t, size, 0, "Cache should have entries")
			assert.LessOrEqual(t, size, capacity, "Cache size should not exceed capacity")
		})
	}
}

// TestCStringCacheEviction verifies cache eviction when full.
func TestCStringCacheEviction(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "cache evicts when full"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear cache before test
			probe.ClearCStringCacheForTest()

			_, capacity := probe.GetCStringCacheStatsForTest()

			// Fill cache to capacity with stable strings
			for i := 0; i < capacity; i++ {
				input := []C.char{
					C.char('s'),
					C.char(byte('0' + (i / 10))),
					C.char(byte('0' + (i % 10))),
				}
				probe.CCharArrayToStringCachedForTest(input, true)
			}

			// Check cache is full
			size, _ := probe.GetCStringCacheStatsForTest()
			assert.Equal(t, capacity, size, "Cache should be at capacity")

			// Add one more entry (should trigger eviction)
			overflow := []C.char{C.char('o'), C.char('v'), C.char('f')}
			probe.CCharArrayToStringCachedForTest(overflow, true)

			// Cache should have been cleared or partially evicted
			sizeAfter, _ := probe.GetCStringCacheStatsForTest()
			assert.LessOrEqual(t, sizeAfter, capacity, "Cache should not exceed capacity")
		})
	}
}

// TestCStringCacheStats verifies cache statistics.
func TestCStringCacheStats(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "cache stats are accurate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear cache before test
			probe.ClearCStringCacheForTest()

			// Get initial stats
			size, capacity := probe.GetCStringCacheStatsForTest()
			assert.Equal(t, 0, size, "Cache should be empty initially")
			assert.Greater(t, capacity, 0, "Capacity should be positive")

			// Add entries
			for i := 0; i < 5; i++ {
				input := []C.char{C.char('x'), C.char(byte('0' + i))}
				probe.CCharArrayToStringCachedForTest(input, true)
			}

			// Check stats updated
			size, capacity = probe.GetCStringCacheStatsForTest()
			assert.Equal(t, 5, size, "Cache should have 5 entries")
			assert.Greater(t, capacity, size, "Capacity should be greater than size")

			// Clear and verify
			probe.ClearCStringCacheForTest()
			size, capacity = probe.GetCStringCacheStatsForTest()
			assert.Equal(t, 0, size, "Cache should be empty after clear")
		})
	}
}

// TestCStringCacheEmpty verifies empty string handling.
func TestCStringCacheEmpty(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "empty strings are handled correctly"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear cache before test
			probe.ClearCStringCacheForTest()

			// Empty array
			input := []C.char{}

			result := probe.CCharArrayToStringCachedForTest(input, true)
			assert.Equal(t, "", result)

			// Cache behavior for empty strings is implementation-defined
			// Just verify no panic occurred
		})
	}
}

// TestCStringCacheConsistency verifies cache returns consistent results.
func TestCStringCacheConsistency(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "cache returns consistent results"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear cache before test
			probe.ClearCStringCacheForTest()

			// Create input
			input := []C.char{
				C.char('d'), C.char('e'), C.char('v'),
				C.char('s'), C.char('d'), C.char('a'),
			}

			// Call multiple times
			results := make([]string, 10)
			for i := 0; i < 10; i++ {
				results[i] = probe.CCharArrayToStringCachedForTest(input, true)
			}

			// All results should be identical
			for i := 1; i < len(results); i++ {
				assert.Equal(t, results[0], results[i], "All cached results should be identical")
			}

			// Should all equal expected value
			assert.Equal(t, "devsda", results[0])
		})
	}
}
