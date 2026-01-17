//go:build linux

// Package linux provides Linux-specific metric collectors using /proc filesystem.
// It implements system and process-level memory monitoring by parsing kernel data.
package linux

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

const (
	// percentageMultiplier converts fractions to percentages.
	percentageMultiplier float64 = 100

	// bytesPerKilobyte converts kilobyte values to bytes.
	bytesPerKilobyte uint64 = 1024

	// meminfoFieldCount is the typical number of fields in /proc/meminfo.
	meminfoFieldCount int = 50

	// statusFieldCount is the typical number of fields in /proc/[pid]/status.
	statusFieldCount int = 10

	// colonFieldCount is the expected number of parts when splitting by colon.
	colonFieldCount int = 2

	// decimalBase is the base for decimal number parsing.
	decimalBase int = 10

	// uint64BitSize is the bit size for uint64 parsing.
	uint64BitSize int = 64
)

// MemoryCollector implements metrics.MemoryCollector by reading from /proc filesystem.
// It provides system-wide and per-process memory statistics from kernel interfaces.
type MemoryCollector struct {
	procPath string
}

// NewMemoryCollector creates a new memory collector using the default /proc path.
//
// Returns:
//   - *MemoryCollector: configured collector instance
func NewMemoryCollector() *MemoryCollector {
	// Use standard proc filesystem location.
	return &MemoryCollector{procPath: "/proc"}
}

// NewMemoryCollectorWithPath creates a new memory collector with a custom proc path.
// This is useful for testing with mock /proc filesystems.
//
// Params:
//   - procPath: custom path to proc filesystem (e.g., "/tmp/mockproc")
//
// Returns:
//   - *MemoryCollector: configured collector instance
func NewMemoryCollectorWithPath(procPath string) *MemoryCollector {
	// Allow custom proc path for testing scenarios.
	return &MemoryCollector{procPath: procPath}
}

// CollectSystem collects system-wide memory metrics from /proc/meminfo.
//
// Params:
//   - ctx: context for cancellation and timeout control
//
// Returns:
//   - metrics.SystemMemory: current system memory statistics
//   - error: context cancellation or filesystem errors
func (c *MemoryCollector) CollectSystem(ctx context.Context) (metrics.SystemMemory, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.SystemMemory{}, ctx.Err()
	}

	// Parse meminfo file into key-value map.
	values, err := c.readMemInfo()
	// Check if meminfo reading failed.
	if err != nil {
		// Failed to read or parse /proc/meminfo.
		return metrics.SystemMemory{}, err
	}

	// Build structured metrics from parsed values.
	return c.buildSystemMemory(values), nil
}

// readMemInfo reads and parses /proc/meminfo into a key-value map.
//
// Returns:
//   - map[string]uint64: memory metrics keyed by field name
//   - error: file open or scan errors
func (c *MemoryCollector) readMemInfo() (map[string]uint64, error) {
	// Open meminfo file from proc filesystem.
	file, err := os.Open(filepath.Join(c.procPath, "meminfo"))
	// Check if file open failed.
	if err != nil {
		// File doesn't exist or insufficient permissions.
		return nil, fmt.Errorf("open /proc/meminfo: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Pre-allocate map with typical number of meminfo fields.
	values := make(map[string]uint64, meminfoFieldCount)
	scanner := bufio.NewScanner(file)

	// Parse each line into key-value pairs.
	for scanner.Scan() {
		key, value := c.parseMemInfoLine(scanner.Text())
		// Skip invalid or empty lines.
		if key != "" {
			values[key] = value
		}
	}

	// Check for scanner errors during reading.
	if err := scanner.Err(); err != nil {
		// I/O error while reading file.
		return nil, fmt.Errorf("scan /proc/meminfo: %w", err)
	}

	// Return complete parsed metrics.
	return values, nil
}

// buildSystemMemory constructs SystemMemory from parsed meminfo values.
//
// Params:
//   - values: parsed key-value pairs from /proc/meminfo
//
// Returns:
//   - metrics.SystemMemory: structured memory metrics with calculations
func (c *MemoryCollector) buildSystemMemory(values map[string]uint64) metrics.SystemMemory {
	mem := metrics.SystemMemory{Timestamp: time.Now()}

	// Map values to struct (values are in kB, convert to bytes).
	mem.Total = values["MemTotal"] * bytesPerKilobyte
	mem.Free = values["MemFree"] * bytesPerKilobyte
	mem.Available = values["MemAvailable"] * bytesPerKilobyte
	mem.Buffers = values["Buffers"] * bytesPerKilobyte
	mem.Cached = values["Cached"] * bytesPerKilobyte
	mem.SwapTotal = values["SwapTotal"] * bytesPerKilobyte
	mem.SwapFree = values["SwapFree"] * bytesPerKilobyte
	mem.Shared = values["Shmem"] * bytesPerKilobyte

	// Calculate swap usage with underflow protection.
	if mem.SwapTotal >= mem.SwapFree {
		mem.SwapUsed = mem.SwapTotal - mem.SwapFree
	}

	// Calculate memory usage with underflow protection.
	if mem.Total >= mem.Available {
		mem.Used = mem.Total - mem.Available
	}

	// Calculate usage percentage to avoid division by zero.
	if mem.Total > 0 {
		mem.UsagePercent = float64(mem.Used) / float64(mem.Total) * percentageMultiplier
	}

	// Return complete metrics with all derived values.
	return mem
}

// parseMemInfoLine parses a single line from /proc/meminfo.
// Format: "FieldName:       12345 kB" or "FieldName:       12345"
//
// Params:
//   - line: single line from /proc/meminfo
//
// Returns:
//   - key: field name (e.g., "MemTotal")
//   - value: numeric value in kilobytes
func (c *MemoryCollector) parseMemInfoLine(line string) (key string, value uint64) {
	parts := strings.SplitN(line, ":", colonFieldCount)

	// Require colon separator for valid lines.
	if len(parts) != colonFieldCount {
		// Invalid format, skip this line.
		return "", 0
	}

	key = strings.TrimSpace(parts[0])
	valueStr := strings.TrimSpace(parts[1])

	// Remove "kB" suffix if present.
	valueStr = strings.TrimSuffix(valueStr, " kB")
	valueStr = strings.TrimSpace(valueStr)

	var err error
	value, err = strconv.ParseUint(valueStr, decimalBase, uint64BitSize)

	// Return key even on parse error, with zero value.
	if err != nil {
		// Non-numeric value, return zero.
		return key, 0
	}

	// Return parsed key-value pair.
	return key, value
}

// CollectProcess collects memory metrics for a specific process from /proc/[pid]/status.
//
// Params:
//   - ctx: context for cancellation and timeout control
//   - pid: process ID to collect metrics for
//
// Returns:
//   - metrics.ProcessMemory: process memory statistics
//   - error: invalid PID, context cancellation, or filesystem errors
func (c *MemoryCollector) CollectProcess(ctx context.Context, pid int) (metrics.ProcessMemory, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.ProcessMemory{}, ctx.Err()
	}

	// Validate PID before proceeding.
	if pid <= 0 {
		// Return error for invalid process ID.
		return metrics.ProcessMemory{}, NewInvalidPIDError(pid)
	}

	// Read and parse process status file.
	proc, values, err := c.readProcessStatus(pid)
	// Check if status file reading failed.
	if err != nil {
		// Failed to read /proc/[pid]/status.
		return metrics.ProcessMemory{}, err
	}

	// Map memory values from status file.
	c.mapProcessMemoryValues(&proc, values)

	// Return complete process metrics.
	return proc, nil
}

// readProcessStatus reads and parses /proc/[pid]/status file.
//
// Params:
//   - pid: process ID to read
//
// Returns:
//   - metrics.ProcessMemory: partially populated metrics (PID, Timestamp, Name)
//   - map[string]uint64: parsed memory field values
//   - error: file open or scan errors
func (c *MemoryCollector) readProcessStatus(pid int) (metrics.ProcessMemory, map[string]uint64, error) {
	statusPath := filepath.Join(c.procPath, strconv.Itoa(pid), "status")

	// Open process status file.
	file, err := os.Open(statusPath)
	// Check if file open failed.
	if err != nil {
		// Process doesn't exist or insufficient permissions.
		return metrics.ProcessMemory{}, nil, fmt.Errorf("open /proc/%d/status: %w", pid, err)
	}
	defer func() { _ = file.Close() }()

	proc := metrics.ProcessMemory{
		PID:       pid,
		Timestamp: time.Now(),
	}
	values := make(map[string]uint64, statusFieldCount)

	scanner := bufio.NewScanner(file)

	// Parse each line for process information.
	for scanner.Scan() {
		line := scanner.Text()

		// Extract process name using CutPrefix for efficiency.
		if name, found := strings.CutPrefix(line, "Name:"); found {
			// Found process name, extract and trim.
			proc.Name = strings.TrimSpace(name)
		} else {
			// Parse memory-related fields.
			key, value := c.parseStatusLine(line)
			// Skip non-memory fields.
			if key != "" {
				values[key] = value
			}
		}
	}

	// Check for scanner errors during reading.
	if err := scanner.Err(); err != nil {
		// I/O error while reading file.
		return metrics.ProcessMemory{}, nil, fmt.Errorf("scan /proc/%d/status: %w", pid, err)
	}

	// Return parsed data.
	return proc, values, nil
}

// mapProcessMemoryValues maps parsed status values to ProcessMemory fields.
//
// Params:
//   - proc: process memory structure to populate (modified in place)
//   - values: parsed memory field values in kilobytes
func (c *MemoryCollector) mapProcessMemoryValues(proc *metrics.ProcessMemory, values map[string]uint64) {
	// Map values to struct (values are in kB, convert to bytes).
	proc.RSS = values["VmRSS"] * bytesPerKilobyte
	proc.VMS = values["VmSize"] * bytesPerKilobyte
	proc.Swap = values["VmSwap"] * bytesPerKilobyte
	proc.Data = values["VmData"] * bytesPerKilobyte
	proc.Stack = values["VmStk"] * bytesPerKilobyte

	// Shared = RssShmem + RssFile (if available).
	proc.Shared = (values["RssShmem"] + values["RssFile"]) * bytesPerKilobyte
}

// parseStatusLine parses a single line from /proc/[pid]/status for memory values.
// Format: "VmRSS:       12345 kB"
//
// Params:
//   - line: single line from /proc/[pid]/status
//
// Returns:
//   - key: memory field name (e.g., "VmRSS")
//   - value: numeric value in kilobytes
func (c *MemoryCollector) parseStatusLine(line string) (key string, value uint64) {
	parts := strings.SplitN(line, ":", colonFieldCount)

	// Require colon separator for valid lines.
	if len(parts) != colonFieldCount {
		// Invalid format, skip this line.
		return "", 0
	}

	key = strings.TrimSpace(parts[0])

	// Only parse memory-related fields to avoid unnecessary work.
	if !strings.HasPrefix(key, "Vm") && !strings.HasPrefix(key, "Rss") {
		// Not a memory field, skip.
		return "", 0
	}

	valueStr := strings.TrimSpace(parts[1])
	valueStr = strings.TrimSuffix(valueStr, " kB")
	valueStr = strings.TrimSpace(valueStr)

	var err error
	value, err = strconv.ParseUint(valueStr, decimalBase, uint64BitSize)

	// Return key even on parse error, with zero value.
	if err != nil {
		// Non-numeric value, return zero.
		return key, 0
	}

	// Return parsed memory metric.
	return key, value
}

// CollectAllProcesses collects memory metrics for all visible processes.
//
// Params:
//   - ctx: context for cancellation and timeout control
//
// Returns:
//   - []metrics.ProcessMemory: memory metrics for all processes
//   - error: context cancellation or filesystem errors
func (c *MemoryCollector) CollectAllProcesses(ctx context.Context) ([]metrics.ProcessMemory, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return nil, ctx.Err()
	}

	// Get total system memory for percentage calculation.
	sysMem, err := c.CollectSystem(ctx)
	// Check if system memory collection failed.
	if err != nil {
		// Cannot collect system memory for percentage calculations.
		return nil, fmt.Errorf("collect system memory: %w", err)
	}

	// Read all entries in /proc directory.
	entries, err := os.ReadDir(c.procPath)
	// Check if reading /proc failed.
	if err != nil {
		// Cannot access /proc filesystem.
		return nil, fmt.Errorf("read /proc: %w", err)
	}

	// Delegate to helper for cleaner separation.
	return c.collectProcessesFromEntries(ctx, entries, sysMem.Total)
}

// collectProcessesFromEntries iterates over /proc entries and collects memory metrics.
//
// Params:
//   - ctx: context for cancellation and timeout control
//   - entries: directory entries from /proc filesystem
//   - totalMemory: total system memory for percentage calculations
//
// Returns:
//   - []metrics.ProcessMemory: memory metrics for valid processes
//   - error: context cancellation
func (c *MemoryCollector) collectProcessesFromEntries(
	ctx context.Context,
	entries []os.DirEntry,
	totalMemory uint64,
) ([]metrics.ProcessMemory, error) {
	var results []metrics.ProcessMemory

	// Iterate through all /proc entries.
	for _, entry := range entries {
		// Check for context cancellation to allow early abort.
		if ctx.Err() != nil {
			// Return context error when cancelled during iteration.
			return nil, ctx.Err()
		}

		proc, ok := c.tryCollectProcessEntry(ctx, entry, totalMemory)

		// Add successfully collected process metrics.
		if ok {
			results = append(results, proc)
		}
	}

	// Return all collected process metrics.
	return results, nil
}

// tryCollectProcessEntry attempts to collect memory metrics for a single /proc entry.
// Returns the metrics and true if successful, zero value and false otherwise.
//
// Params:
//   - ctx: context for cancellation and timeout control
//   - entry: directory entry from /proc filesystem
//   - totalMemory: total system memory for percentage calculations
//
// Returns:
//   - metrics.ProcessMemory: process memory metrics if successful
//   - bool: true if collection succeeded, false otherwise
func (c *MemoryCollector) tryCollectProcessEntry(
	ctx context.Context,
	entry os.DirEntry,
	totalMemory uint64,
) (metrics.ProcessMemory, bool) {
	// Only process directories (skip files like /proc/meminfo).
	if !entry.IsDir() {
		// Not a directory, skip.
		return metrics.ProcessMemory{}, false
	}

	// Convert directory name to PID.
	pid, err := strconv.Atoi(entry.Name())
	// Check if PID conversion failed.
	if err != nil {
		// Not a numeric directory name, skip.
		return metrics.ProcessMemory{}, false
	}

	// Collect metrics for this process.
	proc, err := c.CollectProcess(ctx, pid)
	// Check if process collection failed.
	if err != nil {
		// Process collection failed (process may have exited), skip.
		return metrics.ProcessMemory{}, false
	}

	// Calculate memory usage percentage relative to system total.
	if totalMemory > 0 {
		proc.UsagePercent = float64(proc.RSS) / float64(totalMemory) * percentageMultiplier
	}

	// Return successful collection.
	return proc, true
}
