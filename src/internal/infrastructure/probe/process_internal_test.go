//go:build cgo

package probe

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewProcessCollector_Internal verifies constructor creates valid instance.
func TestNewProcessCollector_Internal(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates non-nil collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewProcessCollector()
			assert.NotNil(t, collector)
		})
	}
}

// TestProcessCollector_StructType verifies the collector type.
func TestProcessCollector_StructType(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "struct type is not nil"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &ProcessCollector{}
			assert.NotNil(t, collector)
		})
	}
}

// TestProcessFDs_Structure verifies ProcessFDs struct fields.
func TestProcessFDs_Structure(t *testing.T) {
	tests := []struct {
		name     string
		pid      int
		count    uint32
		wantPID  int
		wantCnt  uint32
	}{
		{
			name:    "with values",
			pid:     1234,
			count:   42,
			wantPID: 1234,
			wantCnt: 42,
		},
		{
			name:    "zero values",
			pid:     0,
			count:   0,
			wantPID: 0,
			wantCnt: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fds := ProcessFDs{
				PID:   tt.pid,
				Count: tt.count,
			}

			assert.Equal(t, tt.wantPID, fds.PID)
			assert.Equal(t, tt.wantCnt, fds.Count)
		})
	}
}

// TestProcessIO_Structure verifies ProcessIO struct fields.
func TestProcessIO_Structure(t *testing.T) {
	tests := []struct {
		name      string
		pid       int
		readBPS   uint64
		writeBPS  uint64
		wantPID   int
		wantRead  uint64
		wantWrite uint64
	}{
		{
			name:      "with values",
			pid:       1234,
			readBPS:   1000,
			writeBPS:  2000,
			wantPID:   1234,
			wantRead:  1000,
			wantWrite: 2000,
		},
		{
			name:      "zero values",
			pid:       0,
			readBPS:   0,
			writeBPS:  0,
			wantPID:   0,
			wantRead:  0,
			wantWrite: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			io := ProcessIO{
				PID:              tt.pid,
				ReadBytesPerSec:  tt.readBPS,
				WriteBytesPerSec: tt.writeBPS,
			}

			assert.Equal(t, tt.wantPID, io.PID)
			assert.Equal(t, tt.wantRead, io.ReadBytesPerSec)
			assert.Equal(t, tt.wantWrite, io.WriteBytesPerSec)
		})
	}
}

// TestProcessCollector_CollectCPU verifies process CPU collection.
func TestProcessCollector_CollectCPU(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		pid         int
		expectError bool
	}{
		{
			name:        "with initialized probe current pid",
			initProbe:   true,
			pid:         os.Getpid(),
			expectError: false,
		},
		{
			name:        "without initialized probe",
			initProbe:   false,
			pid:         os.Getpid(),
			expectError: true,
		},
		{
			name:        "invalid pid",
			initProbe:   true,
			pid:         99999,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			collector := NewProcessCollector()
			ctx := context.Background()

			cpu, err := collector.CollectCPU(ctx, tt.pid)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.pid, cpu.PID)
			}
		})
	}
}

// TestProcessCollector_CollectMemory verifies process memory collection.
func TestProcessCollector_CollectMemory(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		pid         int
		expectError bool
	}{
		{
			name:        "with initialized probe current pid",
			initProbe:   true,
			pid:         os.Getpid(),
			expectError: false,
		},
		{
			name:        "without initialized probe",
			initProbe:   false,
			pid:         os.Getpid(),
			expectError: true,
		},
		{
			name:        "invalid pid",
			initProbe:   true,
			pid:         99999,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			collector := NewProcessCollector()
			ctx := context.Background()

			mem, err := collector.CollectMemory(ctx, tt.pid)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.pid, mem.PID)
			}
		})
	}
}
