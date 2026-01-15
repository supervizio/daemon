//go:build linux

// Package cgroup_test provides external tests for MemoryStat functionality.
package cgroup_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/infrastructure/resources/cgroup"
)

// TestMemoryStat_Fields validates MemoryStat struct field access.
// It ensures the struct correctly holds memory statistics.
//
// Params:
//   - t: testing instance
func TestMemoryStat_Fields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		stat       cgroup.MemoryStat
		wantAnon   uint64
		wantFile   uint64
		wantKernel uint64
	}{
		{
			name:       "zero value MemoryStat",
			stat:       cgroup.MemoryStat{},
			wantAnon:   0,
			wantFile:   0,
			wantKernel: 0,
		},
		{
			name: "MemoryStat with values",
			stat: cgroup.MemoryStat{
				Anon:       1048576,
				File:       2097152,
				Kernel:     524288,
				Slab:       262144,
				Sock:       131072,
				Shmem:      65536,
				Mapped:     32768,
				Dirty:      16384,
				Pgfault:    1000,
				Pgmajfault: 10,
			},
			wantAnon:   1048576,
			wantFile:   2097152,
			wantKernel: 524288,
		},
		{
			name: "large memory values",
			stat: cgroup.MemoryStat{
				Anon:   17179869184, // 16 GiB
				File:   8589934592,  // 8 GiB
				Kernel: 4294967296,  // 4 GiB
			},
			wantAnon:   17179869184,
			wantFile:   8589934592,
			wantKernel: 4294967296,
		},
	}

	// Test each case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Verify field values
			assert.Equal(t, tt.wantAnon, tt.stat.Anon, "Anon should match")
			assert.Equal(t, tt.wantFile, tt.stat.File, "File should match")
			assert.Equal(t, tt.wantKernel, tt.stat.Kernel, "Kernel should match")
		})
	}
}

// TestMemoryStat_AllFields validates all MemoryStat struct fields.
//
// Params:
//   - t: testing instance
func TestMemoryStat_AllFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		stat cgroup.MemoryStat
	}{
		{
			name: "all fields populated",
			stat: cgroup.MemoryStat{
				Anon:       1024,
				File:       2048,
				Kernel:     4096,
				Slab:       8192,
				Sock:       16384,
				Shmem:      32768,
				Mapped:     65536,
				Dirty:      131072,
				Pgfault:    262144,
				Pgmajfault: 524288,
			},
		},
	}

	// Test each case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Verify all fields are accessible
			assert.Equal(t, uint64(1024), tt.stat.Anon)
			assert.Equal(t, uint64(2048), tt.stat.File)
			assert.Equal(t, uint64(4096), tt.stat.Kernel)
			assert.Equal(t, uint64(8192), tt.stat.Slab)
			assert.Equal(t, uint64(16384), tt.stat.Sock)
			assert.Equal(t, uint64(32768), tt.stat.Shmem)
			assert.Equal(t, uint64(65536), tt.stat.Mapped)
			assert.Equal(t, uint64(131072), tt.stat.Dirty)
			assert.Equal(t, uint64(262144), tt.stat.Pgfault)
			assert.Equal(t, uint64(524288), tt.stat.Pgmajfault)
		})
	}
}
