//go:build linux

// Package cgroup provides internal tests for MemoryStat functionality.
package cgroup

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test_MemoryStat_fieldMap verifies the fieldMap method returns correct mappings.
//
// Params:
//   - t: testing instance
func Test_MemoryStat_fieldMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		expectedFields []string
	}{
		{
			name: "returns all expected memory stat fields",
			expectedFields: []string{
				"anon",
				"file",
				"kernel",
				"slab",
				"sock",
				"shmem",
				"mapped",
				"dirty",
				"pgfault",
				"pgmajfault",
			},
		},
	}

	// Test each case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create MemoryStat instance
			stat := &MemoryStat{}

			// Get field map
			fm := stat.fieldMap()

			// Verify all expected fields are present
			for _, field := range tt.expectedFields {
				// Check if field exists in map
				_, exists := fm[field]
				assert.True(t, exists, "field %q should exist in fieldMap", field)
			}

			// Verify correct number of fields
			assert.Len(t, fm, len(tt.expectedFields), "fieldMap should have correct number of fields")
		})
	}
}

// Test_MemoryStat_fieldMap_PointerValidity verifies fieldMap pointers update struct fields.
//
// Params:
//   - t: testing instance
func Test_MemoryStat_fieldMap_PointerValidity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		field    string
		setValue uint64
	}{
		{
			name:     "anon field pointer is valid",
			field:    "anon",
			setValue: 1024,
		},
		{
			name:     "file field pointer is valid",
			field:    "file",
			setValue: 2048,
		},
		{
			name:     "kernel field pointer is valid",
			field:    "kernel",
			setValue: 4096,
		},
		{
			name:     "slab field pointer is valid",
			field:    "slab",
			setValue: 8192,
		},
		{
			name:     "sock field pointer is valid",
			field:    "sock",
			setValue: 16384,
		},
		{
			name:     "shmem field pointer is valid",
			field:    "shmem",
			setValue: 32768,
		},
		{
			name:     "mapped field pointer is valid",
			field:    "mapped",
			setValue: 65536,
		},
		{
			name:     "dirty field pointer is valid",
			field:    "dirty",
			setValue: 131072,
		},
		{
			name:     "pgfault field pointer is valid",
			field:    "pgfault",
			setValue: 262144,
		},
		{
			name:     "pgmajfault field pointer is valid",
			field:    "pgmajfault",
			setValue: 524288,
		},
	}

	// Test each case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create MemoryStat instance
			stat := &MemoryStat{}

			// Get field map
			fm := stat.fieldMap()

			// Set value through pointer
			*fm[tt.field] = tt.setValue

			// Verify value was set in struct
			var actual uint64
			switch tt.field {
			case "anon":
				actual = stat.Anon
			case "file":
				actual = stat.File
			case "kernel":
				actual = stat.Kernel
			case "slab":
				actual = stat.Slab
			case "sock":
				actual = stat.Sock
			case "shmem":
				actual = stat.Shmem
			case "mapped":
				actual = stat.Mapped
			case "dirty":
				actual = stat.Dirty
			case "pgfault":
				actual = stat.Pgfault
			case "pgmajfault":
				actual = stat.Pgmajfault
			}

			// Verify value
			assert.Equal(t, tt.setValue, actual, "value should be set correctly through pointer")
		})
	}
}
