//go:build linux

package linux

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/metrics"
)
const (
	// testMeminfoStandardFormat contains sample /proc/meminfo content for testing.
	testMeminfoStandardFormat string = `MemTotal:       16384000 kB
MemFree:         4096000 kB
MemAvailable:    8192000 kB
Buffers:          512000 kB
`

	// testStatusStandardFormat contains sample /proc/[pid]/status content for testing.
	testStatusStandardFormat string = `Name:	test-process
State:	S (sleeping)
VmSize:	  100000 kB
VmRSS:	   50000 kB
VmSwap:	    1000 kB
`
)


// TestCollectSystemContextCancellation verifies context cancellation handling.
//
// Params:
//   - t: testing context
//
// Returns:
//   - none
func TestCollectSystemContextCancellation(t *testing.T) {
	tests := []struct {
		name    string
		wantErr error
	}{
		{
			name:    "returns error with cancelled context",
			wantErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewMemoryCollector()

			// Create cancelled context.
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			// Attempt collection with cancelled context.
			_, err := collector.CollectSystem(ctx)

			// Verify error is context cancellation.
			if err == nil {
				t.Error("expected error with cancelled context")
			}

			// Verify error is context.Canceled.
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

// TestMemoryCollectProcessInvalidPID verifies invalid PID handling.
//
// Params:
//   - t: testing context
//
// Returns:
//   - none
func TestMemoryCollectProcessInvalidPID(t *testing.T) {
	tests := []struct {
		name string
		pid  int
	}{
		{name: "zero PID", pid: 0},
		{name: "negative PID", pid: -1},
		{name: "large negative PID", pid: -9999},
	}

	collector := NewMemoryCollector()
	ctx := context.Background()

	// Test each invalid PID.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Attempt to collect with invalid PID.
			_, err := collector.CollectProcess(ctx, tt.pid)

			// Verify error is returned.
			if err == nil {
				t.Errorf("expected error for PID %d", tt.pid)
			}

			// Verify error is InvalidPIDError type.
			var pidErr *InvalidPIDError
			if !errors.As(err, &pidErr) {
				t.Errorf("expected InvalidPIDError, got %T", err)
			}
		})
	}
}

// TestInvalidPIDError verifies InvalidPIDError formatting.
//
// Params:
//   - t: testing context
//
// Returns:
//   - none
func TestInvalidPIDError(t *testing.T) {
	tests := []struct {
		name     string
		pid      int
		expected string
	}{
		{
			name:     "formats negative PID error message",
			pid:      -5,
			expected: "invalid pid: -5",
		},
		{
			name:     "formats zero PID error message",
			pid:      0,
			expected: "invalid pid: 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create error with specific PID.
			err := &InvalidPIDError{PID: tt.pid}

			// Verify error message format.
			if err.Error() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, err.Error())
			}
		})
	}
}

// TestProcPath verifies internal procPath field is correctly set.
//
// Params:
//   - t: testing context
//
// Returns:
//   - none
func TestProcPath(t *testing.T) {
	tests := []struct {
		name         string
		customPath   string
		wantProcPath string
	}{
		{
			name:         "default collector uses /proc",
			wantProcPath: "/proc",
		},
		{
			name:         "custom collector uses custom path",
			customPath:   "/tmp/mockproc",
			wantProcPath: "/tmp/mockproc",
		},
		{
			name:         "custom collector uses alternative path",
			customPath:   "/var/test/proc",
			wantProcPath: "/var/test/proc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var collector *MemoryCollector
			if tt.customPath != "" {
				collector = NewMemoryCollectorWithPath(tt.customPath)
			} else {
				collector = NewMemoryCollector()
			}

			// Verify collector is not nil.
			if collector == nil {
				t.Fatal("expected non-nil collector")
			}

			// Verify proc path is set correctly (internal field access).
			if collector.procPath != tt.wantProcPath {
				t.Errorf("expected procPath=%s, got %s", tt.wantProcPath, collector.procPath)
			}
		})
	}
}

// TestMemoryCollector_ReadMemInfo verifies meminfo file parsing.
//
// Params:
//   - t: testing context
//
// Returns:
//   - none
func TestMemoryCollector_ReadMemInfo(t *testing.T) {
	tests := []struct {
		name           string
		meminfoContent string
		wantKeys       []string
		wantValues     map[string]uint64
		wantErr        bool
	}{
		{
			name: "parses standard meminfo format",
			meminfoContent: testMeminfoStandardFormat,
			wantKeys: []string{"MemTotal", "MemFree", "MemAvailable", "Buffers"},
			wantValues: map[string]uint64{
				"MemTotal":     16384000,
				"MemFree":      4096000,
				"MemAvailable": 8192000,
				"Buffers":      512000,
			},
			wantErr: false,
		},
		{
			name: "handles empty lines gracefully",
			meminfoContent: `MemTotal:       1000 kB

MemFree:         500 kB
`,
			wantKeys: []string{"MemTotal", "MemFree"},
			wantValues: map[string]uint64{
				"MemTotal": 1000,
				"MemFree":  500,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary mock proc directory.
			mockProc := t.TempDir()
			err := os.WriteFile(filepath.Join(mockProc, "meminfo"), []byte(tt.meminfoContent), 0o644)
			if err != nil {
				t.Fatalf("failed to create mock meminfo: %v", err)
			}

			// Create collector with mock path.
			collector := NewMemoryCollectorWithPath(mockProc)

			// Call readMemInfo.
			values, err := collector.readMemInfo()

			// Verify error expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("readMemInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify parsed values.
			if !tt.wantErr {
				for key, expectedValue := range tt.wantValues {
					if values[key] != expectedValue {
						t.Errorf("key %s: expected %d, got %d", key, expectedValue, values[key])
					}
				}
			}
		})
	}
}

// TestMemoryCollector_ReadMemInfoMissingFile verifies error handling for missing meminfo file.
//
// Params:
//   - t: testing context
//
// Returns:
//   - none
func TestMemoryCollector_ReadMemInfoMissingFile(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "returns error for missing meminfo file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create empty mock proc directory without meminfo.
			mockProc := t.TempDir()
			collector := NewMemoryCollectorWithPath(mockProc)

			// Call readMemInfo.
			_, err := collector.readMemInfo()

			// Verify error is returned.
			if err == nil {
				t.Error("expected error for missing meminfo file")
			}
		})
	}
}

// TestMemoryCollector_BuildSystemMemory verifies system memory construction from parsed values.
//
// Params:
//   - t: testing context
//
// Returns:
//   - none
func TestMemoryCollector_BuildSystemMemory(t *testing.T) {
	tests := []struct {
		name              string
		values            map[string]uint64
		wantTotal         uint64
		wantUsed          uint64
		wantUsagePercent  float64
		wantSwapUsed      uint64
		checkUsagePercent bool
	}{
		{
			name: "calculates derived values correctly",
			values: map[string]uint64{
				"MemTotal":     10000,
				"MemFree":      2000,
				"MemAvailable": 5000,
				"SwapTotal":    4000,
				"SwapFree":     3000,
				"Buffers":      100,
				"Cached":       200,
				"Shmem":        50,
			},
			wantTotal:         10000 * 1024,
			wantUsed:          5000 * 1024,
			wantUsagePercent:  50.0,
			wantSwapUsed:      1000 * 1024,
			checkUsagePercent: true,
		},
		{
			name: "handles zero total memory",
			values: map[string]uint64{
				"MemTotal":     0,
				"MemAvailable": 0,
			},
			wantTotal:         0,
			wantUsed:          0,
			wantUsagePercent:  0.0,
			checkUsagePercent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewMemoryCollector()

			// Call buildSystemMemory.
			mem := collector.buildSystemMemory(tt.values)

			// Verify total.
			if mem.Total != tt.wantTotal {
				t.Errorf("Total: expected %d, got %d", tt.wantTotal, mem.Total)
			}

			// Verify used.
			if mem.Used != tt.wantUsed {
				t.Errorf("Used: expected %d, got %d", tt.wantUsed, mem.Used)
			}

			// Verify swap used.
			if mem.SwapUsed != tt.wantSwapUsed {
				t.Errorf("SwapUsed: expected %d, got %d", tt.wantSwapUsed, mem.SwapUsed)
			}

			// Verify usage percentage if specified.
			if tt.checkUsagePercent {
				diff := mem.UsagePercent - tt.wantUsagePercent
				if diff < -0.01 || diff > 0.01 {
					t.Errorf("UsagePercent: expected %f, got %f", tt.wantUsagePercent, mem.UsagePercent)
				}
			}

			// Verify timestamp is set.
			if mem.Timestamp.IsZero() {
				t.Error("Timestamp should be set")
			}
		})
	}
}

// TestMemoryCollector_ParseMemInfoLine verifies single line parsing from meminfo.
//
// Params:
//   - t: testing context
//
// Returns:
//   - none
func TestMemoryCollector_ParseMemInfoLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantKey   string
		wantValue uint64
	}{
		{
			name:      "parses standard kB format",
			line:      "MemTotal:       16384000 kB",
			wantKey:   "MemTotal",
			wantValue: 16384000,
		},
		{
			name:      "parses without kB suffix",
			line:      "HugePages_Total:       0",
			wantKey:   "HugePages_Total",
			wantValue: 0,
		},
		{
			name:      "handles missing colon",
			line:      "InvalidLine",
			wantKey:   "",
			wantValue: 0,
		},
		{
			name:      "handles non-numeric value",
			line:      "Invalid:    notanumber kB",
			wantKey:   "Invalid",
			wantValue: 0,
		},
		{
			name:      "handles extra whitespace",
			line:      "MemFree:           1234567 kB",
			wantKey:   "MemFree",
			wantValue: 1234567,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewMemoryCollector()

			// Call parseMemInfoLine.
			key, value := collector.parseMemInfoLine(tt.line)

			// Verify key.
			if key != tt.wantKey {
				t.Errorf("key: expected %q, got %q", tt.wantKey, key)
			}

			// Verify value.
			if value != tt.wantValue {
				t.Errorf("value: expected %d, got %d", tt.wantValue, value)
			}
		})
	}
}

// TestMemoryCollector_ReadProcessStatus verifies process status file parsing.
//
// Params:
//   - t: testing context
//
// Returns:
//   - none
func TestMemoryCollector_ReadProcessStatus(t *testing.T) {
	tests := []struct {
		name          string
		statusContent string
		pid           int
		wantName      string
		wantVMRSS     uint64
		wantErr       bool
	}{
		{
			name: "parses standard status format",
			statusContent: testStatusStandardFormat,
			pid:       1234,
			wantName:  "test-process",
			wantVMRSS: 50000,
			wantErr:   false,
		},
		{
			name: "handles missing memory fields",
			statusContent: `Name:	minimal-process
State:	R (running)
`,
			pid:       5678,
			wantName:  "minimal-process",
			wantVMRSS: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary mock proc directory.
			mockProc := t.TempDir()
			pidDir := filepath.Join(mockProc, "1234")
			if tt.pid != 1234 {
				pidDir = filepath.Join(mockProc, "5678")
			}
			err := os.Mkdir(pidDir, 0o755)
			if err != nil {
				t.Fatalf("failed to create pid directory: %v", err)
			}
			err = os.WriteFile(filepath.Join(pidDir, "status"), []byte(tt.statusContent), 0o644)
			if err != nil {
				t.Fatalf("failed to create mock status: %v", err)
			}

			// Create collector with mock path.
			collector := NewMemoryCollectorWithPath(mockProc)

			// Call readProcessStatus.
			proc, values, err := collector.readProcessStatus(tt.pid)

			// Verify error expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("readProcessStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify parsed values.
			if !tt.wantErr {
				if proc.Name != tt.wantName {
					t.Errorf("Name: expected %q, got %q", tt.wantName, proc.Name)
				}
				if proc.PID != tt.pid {
					t.Errorf("PID: expected %d, got %d", tt.pid, proc.PID)
				}
				if values["VmRSS"] != tt.wantVMRSS {
					t.Errorf("VmRSS: expected %d, got %d", tt.wantVMRSS, values["VmRSS"])
				}
				if proc.Timestamp.IsZero() {
					t.Error("Timestamp should be set")
				}
			}
		})
	}
}

// TestMemoryCollector_ReadProcessStatusMissingFile verifies error handling for missing status file.
//
// Params:
//   - t: testing context
//
// Returns:
//   - none
func TestMemoryCollector_ReadProcessStatusMissingFile(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "returns error for missing status file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create empty mock proc directory.
			mockProc := t.TempDir()
			collector := NewMemoryCollectorWithPath(mockProc)

			// Call readProcessStatus.
			_, _, err := collector.readProcessStatus(9999)

			// Verify error is returned.
			if err == nil {
				t.Error("expected error for missing status file")
			}
		})
	}
}

// TestMemoryCollector_MapProcessMemoryValues verifies memory value mapping to struct fields.
//
// Params:
//   - t: testing context
//
// Returns:
//   - none
func TestMemoryCollector_MapProcessMemoryValues(t *testing.T) {
	tests := []struct {
		name       string
		values     map[string]uint64
		wantRSS    uint64
		wantVMS    uint64
		wantSwap   uint64
		wantData   uint64
		wantStack  uint64
		wantShared uint64
	}{
		{
			name: "maps all memory values correctly",
			values: map[string]uint64{
				"VmRSS":    80000,
				"VmSize":   450000,
				"VmSwap":   1000,
				"VmData":   200000,
				"VmStk":    136,
				"RssShmem": 5000,
				"RssFile":  15000,
			},
			wantRSS:    80000 * 1024,
			wantVMS:    450000 * 1024,
			wantSwap:   1000 * 1024,
			wantData:   200000 * 1024,
			wantStack:  136 * 1024,
			wantShared: 20000 * 1024, // RssShmem + RssFile
		},
		{
			name:       "handles empty values map",
			values:     map[string]uint64{},
			wantRSS:    0,
			wantVMS:    0,
			wantSwap:   0,
			wantData:   0,
			wantStack:  0,
			wantShared: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewMemoryCollector()
			proc := &metrics.ProcessMemory{}

			// Call mapProcessMemoryValues.
			collector.mapProcessMemoryValues(proc, tt.values)

			// Verify mapped values.
			if proc.RSS != tt.wantRSS {
				t.Errorf("RSS: expected %d, got %d", tt.wantRSS, proc.RSS)
			}
			if proc.VMS != tt.wantVMS {
				t.Errorf("VMS: expected %d, got %d", tt.wantVMS, proc.VMS)
			}
			if proc.Swap != tt.wantSwap {
				t.Errorf("Swap: expected %d, got %d", tt.wantSwap, proc.Swap)
			}
			if proc.Data != tt.wantData {
				t.Errorf("Data: expected %d, got %d", tt.wantData, proc.Data)
			}
			if proc.Stack != tt.wantStack {
				t.Errorf("Stack: expected %d, got %d", tt.wantStack, proc.Stack)
			}
			if proc.Shared != tt.wantShared {
				t.Errorf("Shared: expected %d, got %d", tt.wantShared, proc.Shared)
			}
		})
	}
}

// TestMemoryCollector_ParseStatusLine verifies single line parsing from process status.
//
// Params:
//   - t: testing context
//
// Returns:
//   - none
func TestMemoryCollector_ParseStatusLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantKey   string
		wantValue uint64
	}{
		{
			name:      "parses VmRSS field",
			line:      "VmRSS:	   80000 kB",
			wantKey:   "VmRSS",
			wantValue: 80000,
		},
		{
			name:      "parses RssShmem field",
			line:      "RssShmem:	    5000 kB",
			wantKey:   "RssShmem",
			wantValue: 5000,
		},
		{
			name:      "skips non-memory field",
			line:      "State:	S (sleeping)",
			wantKey:   "",
			wantValue: 0,
		},
		{
			name:      "handles missing colon",
			line:      "InvalidLine",
			wantKey:   "",
			wantValue: 0,
		},
		{
			name:      "handles non-numeric memory value",
			line:      "VmPeak:	invalid kB",
			wantKey:   "VmPeak",
			wantValue: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewMemoryCollector()

			// Call parseStatusLine.
			key, value := collector.parseStatusLine(tt.line)

			// Verify key.
			if key != tt.wantKey {
				t.Errorf("key: expected %q, got %q", tt.wantKey, key)
			}

			// Verify value.
			if value != tt.wantValue {
				t.Errorf("value: expected %d, got %d", tt.wantValue, value)
			}
		})
	}
}

// TestMemoryCollector_CollectProcessesFromEntries verifies process collection from directory entries.
//
// Params:
//   - t: testing context
//
// Returns:
//   - none
func TestMemoryCollector_CollectProcessesFromEntries(t *testing.T) {
	tests := []struct {
		name          string
		processes     []string
		dirs          []string
		totalMemory   uint64
		wantCount     int
		wantErr       bool
		cancelContext bool
	}{
		{
			name:        "collects processes from valid entries",
			processes:   []string{"1:init:1000:2000", "2:kworker:500:1000"},
			totalMemory: 10000000 * 1024,
			wantCount:   2,
		},
		{
			name:        "skips non-numeric directories",
			processes:   []string{"1:init:1000:2000"},
			dirs:        []string{"self", "sys"},
			totalMemory: 10000000 * 1024,
			wantCount:   1,
		},
		{
			name:          "handles context cancellation",
			processes:     []string{"1:init:1000:2000"},
			totalMemory:   10000000 * 1024,
			wantErr:       true,
			cancelContext: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProc := t.TempDir()
			setupMockProc(t, mockProc, tt.processes, tt.dirs)

			collector := NewMemoryCollectorWithPath(mockProc)
			entries, err := os.ReadDir(mockProc)
			if err != nil {
				t.Fatalf("failed to read directory: %v", err)
			}

			ctx := context.Background()
			if tt.cancelContext {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			}

			results, err := collector.collectProcessesFromEntries(ctx, entries, tt.totalMemory)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(results) != tt.wantCount {
				t.Errorf("expected %d results, got %d", tt.wantCount, len(results))
			}
		})
	}
}

// setupMockProc creates mock process directories and empty directories.
func setupMockProc(t *testing.T, baseDir string, processes, dirs []string) {
	t.Helper()
	for _, p := range processes {
		// Create process directory using first character as PID.
		pidDir := filepath.Join(baseDir, p[:1])
		if err := os.Mkdir(pidDir, 0o755); err != nil {
			t.Fatal(err)
		}
		status := "Name:\tproc\nVmRSS:\t1000 kB\n"
		if err := os.WriteFile(filepath.Join(pidDir, "status"), []byte(status), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, d := range dirs {
		if err := os.Mkdir(filepath.Join(baseDir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

// TestMemoryCollector_TryCollectProcessEntry verifies single process entry collection.
//
// Params:
//   - t: testing context
//
// Returns:
//   - none
func TestMemoryCollector_TryCollectProcessEntry(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, dir string)
		entryName   string
		totalMemory uint64
		wantOK      bool
		wantPID     int
	}{
		{
			name: "collects valid process entry",
			setupFunc: func(t *testing.T, dir string) {
				t.Helper()
				createProcessDir(t, dir, "1234", "Name:\ttest-app\nVmRSS:\t5000 kB\n")
			},
			entryName:   "1234",
			totalMemory: 10000000 * 1024,
			wantOK:      true,
			wantPID:     1234,
		},
		{
			name: "skips non-directory entry",
			setupFunc: func(t *testing.T, dir string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(dir, "version"), []byte("test"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			entryName:   "version",
			totalMemory: 10000000 * 1024,
			wantOK:      false,
		},
		{
			name: "skips non-numeric directory",
			setupFunc: func(t *testing.T, dir string) {
				t.Helper()
				if err := os.Mkdir(filepath.Join(dir, "self"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			entryName:   "self",
			totalMemory: 10000000 * 1024,
			wantOK:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProc := t.TempDir()
			tt.setupFunc(t, mockProc)

			collector := NewMemoryCollectorWithPath(mockProc)
			entries, err := os.ReadDir(mockProc)
			if err != nil {
				t.Fatalf("failed to read directory: %v", err)
			}

			var targetEntry os.DirEntry
			for _, e := range entries {
				if e.Name() == tt.entryName {
					targetEntry = e
					break
				}
			}
			if targetEntry == nil {
				t.Fatalf("target entry %q not found", tt.entryName)
			}

			ctx := context.Background()
			proc, ok := collector.tryCollectProcessEntry(ctx, targetEntry, tt.totalMemory)

			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if tt.wantOK && proc.PID != tt.wantPID {
				t.Errorf("PID: expected %d, got %d", tt.wantPID, proc.PID)
			}
		})
	}
}

// createProcessDir creates a mock process directory with status file.
func createProcessDir(t *testing.T, baseDir, pid, status string) {
	t.Helper()
	pidDir := filepath.Join(baseDir, pid)
	if err := os.Mkdir(pidDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pidDir, "status"), []byte(status), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestMemoryCollector_TryCollectProcessEntryMissingStatus verifies handling of missing status file.
//
// Params:
//   - t: testing context
//
// Returns:
//   - none
func TestMemoryCollector_TryCollectProcessEntryMissingStatus(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "returns false for process with missing status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProc := t.TempDir()

			// Create directory without status file.
			pidDir := filepath.Join(mockProc, "9999")
			if err := os.Mkdir(pidDir, 0o755); err != nil {
				t.Fatal(err)
			}

			collector := NewMemoryCollectorWithPath(mockProc)

			// Read directory entries.
			entries, err := os.ReadDir(mockProc)
			if err != nil {
				t.Fatalf("failed to read directory: %v", err)
			}

			ctx := context.Background()

			// Call tryCollectProcessEntry.
			_, ok := collector.tryCollectProcessEntry(ctx, entries[0], 10000000*1024)

			// Verify failure.
			if ok {
				t.Error("expected false for process with missing status file")
			}
		})
	}
}

// TestMemoryCollector_BuildSystemMemoryTimestamp verifies timestamp is set during build.
//
// Params:
//   - t: testing context
//
// Returns:
//   - none
func TestMemoryCollector_BuildSystemMemoryTimestamp(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "sets timestamp during build",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewMemoryCollector()
			values := map[string]uint64{"MemTotal": 1000}

			before := time.Now()
			mem := collector.buildSystemMemory(values)
			after := time.Now()

			// Verify timestamp is within bounds.
			if mem.Timestamp.Before(before) {
				t.Error("Timestamp should be after test start")
			}
			if mem.Timestamp.After(after) {
				t.Error("Timestamp should be before test end")
			}
		})
	}
}
