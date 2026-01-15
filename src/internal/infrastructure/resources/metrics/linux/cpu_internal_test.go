//go:build linux

package linux

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_CPUCollector_parseCPULine verifies CPU line parsing with various inputs.
//
// Params:
//   - t: testing instance
func Test_CPUCollector_parseCPULine(t *testing.T) {
	// Create collector
	collector := NewCPUCollector()

	tests := []struct {
		name      string
		line      string
		wantUser  uint64
		wantNice  uint64
		wantIdle  uint64
		expectErr bool
	}{
		{
			name:      "valid full line",
			line:      "cpu  100 50 75 1000 25 10 5 0 0 0",
			wantUser:  100,
			wantNice:  50,
			wantIdle:  1000,
			expectErr: false,
		},
		{
			name:      "minimum fields",
			line:      "cpu  200 100 150 2000",
			wantUser:  200,
			wantNice:  100,
			wantIdle:  2000,
			expectErr: false,
		},
		{
			name:      "too few fields",
			line:      "cpu 100",
			expectErr: true,
		},
		{
			name:      "empty line",
			line:      "",
			expectErr: true,
		},
	}

	// Test each case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse CPU line
			cpu, err := collector.parseCPULine(tt.line)

			// Check error expectation
			if tt.expectErr {
				// Verify error is returned
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidCPULine)
			} else {
				// Verify parsing succeeds
				require.NoError(t, err)
				assert.Equal(t, tt.wantUser, cpu.User)
				assert.Equal(t, tt.wantNice, cpu.Nice)
				assert.Equal(t, tt.wantIdle, cpu.Idle)
				assert.NotZero(t, cpu.Timestamp)
			}
		})
	}
}

// Test_CPUCollector_parseProcessStat verifies process stat parsing.
//
// Params:
//   - t: testing instance
func Test_CPUCollector_parseProcessStat(t *testing.T) {
	// Create collector
	collector := NewCPUCollector()

	tests := []struct {
		name      string
		pid       int
		data      string
		wantName  string
		wantUser  uint64
		expectErr error
	}{
		{
			name:      "valid process stat",
			pid:       1234,
			data:      "1234 (test-process) S 1 1234 1234 0 -1 4194304 100 0 0 0 10 5 0 0 20 0 1 0 12345 1024000 50 18446744073709551615 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0",
			wantName:  "test-process",
			wantUser:  10,
			expectErr: nil,
		},
		{
			name:      "process with spaces in name",
			pid:       5678,
			data:      "5678 (my test process) R 1 5678 5678 0 -1 4194304 200 0 0 0 20 10 5 3 20 0 2 0 23456 2048000 100 18446744073709551615 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0",
			wantName:  "my test process",
			wantUser:  20,
			expectErr: nil,
		},
		{
			name:      "missing parentheses",
			pid:       1234,
			data:      "1234 test S 1",
			expectErr: ErrInvalidStatFormat,
		},
		{
			name:      "insufficient fields",
			pid:       1234,
			data:      "1234 (test) S 1",
			expectErr: ErrInsufficientStatFields,
		},
		{
			name:      "malformed parentheses",
			pid:       1234,
			data:      "1234 )test( S 1",
			expectErr: ErrInvalidStatFormat,
		},
	}

	// Test each case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse process stat
			proc, err := collector.parseProcessStat(tt.pid, tt.data)

			// Check error expectation
			if tt.expectErr != nil {
				// Verify expected error is returned
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				// Verify parsing succeeds
				require.NoError(t, err)
				assert.Equal(t, tt.pid, proc.PID)
				assert.Equal(t, tt.wantName, proc.Name)
				assert.Equal(t, tt.wantUser, proc.User)
				assert.NotZero(t, proc.Timestamp)
			}
		})
	}
}

// Test_CPUCollector_collectFromEntries verifies collection from directory entries.
//
// Params:
//   - t: testing instance
func Test_CPUCollector_collectFromEntries(t *testing.T) {
	tests := []struct {
		name          string
		setupMockFS   func(t *testing.T) string
		expectCount   int
		expectPIDs    []int
		expectErr     bool
		cancelContext bool
	}{
		{
			name: "valid processes",
			setupMockFS: func(t *testing.T) string {
				// Create temporary directory for mock /proc
				tmpDir := t.TempDir()

				// Create two mock process directories
				require.NoError(t, os.MkdirAll(tmpDir+"/1234", 0o755))
				require.NoError(t, os.MkdirAll(tmpDir+"/5678", 0o755))

				// Create stat files with valid content
				stat1 := "1234 (proc1) S 1 1234 1234 0 -1 4194304 100 0 0 0 10 5 0 0 20 0 1 0 12345 1024000 50 18446744073709551615 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0"
				stat2 := "5678 (proc2) S 1 5678 5678 0 -1 4194304 200 0 0 0 20 10 0 0 20 0 1 0 23456 2048000 100 18446744073709551615 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0"

				// Write stat files
				require.NoError(t, os.WriteFile(tmpDir+"/1234/stat", []byte(stat1), 0o644))
				require.NoError(t, os.WriteFile(tmpDir+"/5678/stat", []byte(stat2), 0o644))

				// Return temp directory path
				return tmpDir
			},
			expectCount: 2,
			expectPIDs:  []int{1234, 5678},
			expectErr:   false,
		},
		{
			name: "mixed valid and invalid",
			setupMockFS: func(t *testing.T) string {
				// Create temporary directory
				tmpDir := t.TempDir()

				// Create one valid process
				require.NoError(t, os.MkdirAll(tmpDir+"/9999", 0o755))
				stat := "9999 (valid) S 1 9999 9999 0 -1 4194304 100 0 0 0 10 5 0 0 20 0 1 0 12345 1024000 50 18446744073709551615 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0"
				require.NoError(t, os.WriteFile(tmpDir+"/9999/stat", []byte(stat), 0o644))

				// Create non-numeric directory (should be skipped)
				require.NoError(t, os.MkdirAll(tmpDir+"/self", 0o755))

				// Create file (not directory, should be skipped)
				require.NoError(t, os.WriteFile(tmpDir+"/version", []byte("test"), 0o644))

				// Return temp directory
				return tmpDir
			},
			expectCount: 1,
			expectPIDs:  []int{9999},
			expectErr:   false,
		},
		{
			name: "context cancelled",
			setupMockFS: func(t *testing.T) string {
				// Create temporary directory with one process
				tmpDir := t.TempDir()
				require.NoError(t, os.MkdirAll(tmpDir+"/8888", 0o755))
				stat := "8888 (proc) S 1 8888 8888 0 -1 4194304 100 0 0 0 10 5 0 0 20 0 1 0 12345 1024000 50 18446744073709551615 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0"
				require.NoError(t, os.WriteFile(tmpDir+"/8888/stat", []byte(stat), 0o644))
				// Return temp directory
				return tmpDir
			},
			cancelContext: true,
			expectErr:     true,
		},
	}

	// Test each case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock filesystem
			tmpDir := tt.setupMockFS(t)

			// Create collector with mock path
			collector := NewCPUCollectorWithPath(tmpDir)

			// Create context
			ctx := context.Background()
			// Check if context should be cancelled
			if tt.cancelContext {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				// Cancel immediately
				cancel()
			}

			// Read directory entries
			entries, err := os.ReadDir(tmpDir)
			require.NoError(t, err)

			// Collect from entries
			processes, err := collector.collectFromEntries(ctx, entries)

			// Check error expectation
			if tt.expectErr {
				// Verify error is returned
				require.Error(t, err)
			} else {
				// Verify collection succeeds
				require.NoError(t, err)
				assert.Len(t, processes, tt.expectCount)

				// Verify expected PIDs are present
				for i, expectedPID := range tt.expectPIDs {
					// Check if PID matches (within bounds)
					if i < len(processes) {
						assert.Equal(t, expectedPID, processes[i].PID, fmt.Sprintf("process %d PID mismatch", i))
					}
				}
			}
		})
	}
}
