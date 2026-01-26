//go:build linux

// Package cgroup provides internal tests for MemoryStat functionality.
package cgroup

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test_MemoryStat_setField verifies the setField method sets correct fields.
//
// Params:
//   - t: testing instance
func Test_MemoryStat_setField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		expectedFields []string
	}{
		{
			name: "recognizes all expected memory stat fields",
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

			// Verify all expected fields are recognized
			for _, field := range tt.expectedFields {
				// Check if field is recognized by setField
				recognized := stat.setField(field, 1)
				assert.True(t, recognized, "field %q should be recognized by setField", field)
			}

			// Verify unknown field returns false
			recognized := stat.setField("unknown_field", 1)
			assert.False(t, recognized, "unknown field should return false")
		})
	}
}

// Test_MemoryStat_setField_Values verifies setField updates struct fields correctly.
//
// Params:
//   - t: testing instance
func Test_MemoryStat_setField_Values(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		field    string
		setValue uint64
	}{
		{
			name:     "anon field is set correctly",
			field:    "anon",
			setValue: 1024,
		},
		{
			name:     "file field is set correctly",
			field:    "file",
			setValue: 2048,
		},
		{
			name:     "kernel field is set correctly",
			field:    "kernel",
			setValue: 4096,
		},
		{
			name:     "slab field is set correctly",
			field:    "slab",
			setValue: 8192,
		},
		{
			name:     "sock field is set correctly",
			field:    "sock",
			setValue: 16384,
		},
		{
			name:     "shmem field is set correctly",
			field:    "shmem",
			setValue: 32768,
		},
		{
			name:     "mapped field is set correctly",
			field:    "mapped",
			setValue: 65536,
		},
		{
			name:     "dirty field is set correctly",
			field:    "dirty",
			setValue: 131072,
		},
		{
			name:     "pgfault field is set correctly",
			field:    "pgfault",
			setValue: 262144,
		},
		{
			name:     "pgmajfault field is set correctly",
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

			// Set value through setField
			ok := stat.setField(tt.field, tt.setValue)
			assert.True(t, ok, "setField should return true for known field")

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
			assert.Equal(t, tt.setValue, actual, "value should be set correctly through setField")
		})
	}
}
