//go:build linux

// Package linux provides CPU metrics collection from Linux /proc filesystem.
// It implements domain.CPUCollector by parsing /proc/stat for system-wide metrics
// and /proc/[pid]/stat for per-process metrics.
package linux

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// /proc/stat field indices (after "cpu " prefix)
const (
	// cpuFieldUser is the user mode CPU time field index.
	cpuFieldUser int = 1
	// cpuFieldNice is the nice mode CPU time field index.
	cpuFieldNice int = 2
	// cpuFieldSystem is the system mode CPU time field index.
	cpuFieldSystem int = 3
	// cpuFieldIdle is the idle CPU time field index.
	cpuFieldIdle int = 4
	// cpuFieldIOWait is the I/O wait CPU time field index.
	cpuFieldIOWait int = 5
	// cpuFieldIRQ is the IRQ servicing CPU time field index.
	cpuFieldIRQ int = 6
	// cpuFieldSoftIRQ is the soft IRQ servicing CPU time field index.
	cpuFieldSoftIRQ int = 7
	// cpuFieldSteal is the stolen CPU time field index.
	cpuFieldSteal int = 8
	// cpuFieldGuest is the guest CPU time field index.
	cpuFieldGuest int = 9
	// cpuFieldGuestNice is the guest nice CPU time field index.
	cpuFieldGuestNice int = 10

	// minCPUFields is the minimum number of fields required in a cpu line.
	minCPUFields int = 5
)

// /proc/[pid]/stat field indices (after command name extraction)
const (
	// statFieldUTime is the user mode CPU time offset in /proc/[pid]/stat after comm.
	statFieldUTime int = 11
	// statFieldSTime is the system mode CPU time offset.
	statFieldSTime int = 12
	// statFieldCUTime is the children user mode CPU time offset.
	statFieldCUTime int = 13
	// statFieldCSTime is the children system mode CPU time offset.
	statFieldCSTime int = 14
	// statFieldStartTime is the process start time offset.
	statFieldStartTime int = 19

	// minStatFields is the minimum number of fields required after comm in /proc/[pid]/stat.
	minStatFields int = 20
)

// strconv.ParseUint parameters
const (
	// parseBase is the base for parsing numeric strings.
	parseBase int = 10
	// parseBitSize is the bit size for parsing uint64.
	parseBitSize int = 64
)

// /proc/[pid]/stat parsing offsets
const (
	// statCommEndOffset is the offset from the closing paren to the next field.
	statCommEndOffset int = 2
)

// Sentinel errors for CPU collection failures.
var (
	// ErrInvalidPID indicates an invalid process ID was provided.
	ErrInvalidPID error = errors.New("invalid pid")
	// ErrInvalidCPULine indicates malformed CPU line in /proc/stat.
	ErrInvalidCPULine error = errors.New("invalid cpu line format")
	// ErrInvalidStatFormat indicates malformed /proc/[pid]/stat.
	ErrInvalidStatFormat error = errors.New("invalid stat format")
	// ErrInsufficientStatFields indicates too few fields in /proc/[pid]/stat.
	ErrInsufficientStatFields error = errors.New("insufficient fields in stat")
)

// CPUCollector implements metrics.CPUCollector by reading from /proc filesystem.
// It provides methods to collect system-wide CPU metrics and per-process CPU metrics
// by parsing Linux /proc virtual filesystem.
type CPUCollector struct {
	procPath string
}

// NewCPUCollector creates a new CPU collector using the default /proc path.
//
// Returns:
//   - *CPUCollector: configured collector with /proc as the base path
func NewCPUCollector() *CPUCollector {
	// Return collector with default /proc path
	return &CPUCollector{procPath: "/proc"}
}

// NewCPUCollectorWithPath creates a new CPU collector with a custom proc path.
// This is useful for testing with mock /proc filesystems.
//
// Params:
//   - procPath: custom path to proc filesystem root
//
// Returns:
//   - *CPUCollector: configured collector with custom path
func NewCPUCollectorWithPath(procPath string) *CPUCollector {
	// Return collector with custom proc path for testing
	return &CPUCollector{procPath: procPath}
}

// CollectSystem collects system-wide CPU metrics from /proc/stat.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - metrics.SystemCPU: system-wide CPU metrics
//   - error: error if collection fails
func (c *CPUCollector) CollectSystem(ctx context.Context) (metrics.SystemCPU, error) {
	// Check if context is already cancelled before proceeding
	select {
	case <-ctx.Done():
		// Return context error if cancelled
		return metrics.SystemCPU{}, ctx.Err()
	default:
	}

	// Open /proc/stat file for reading
	file, err := os.Open(filepath.Join(c.procPath, "stat"))
	// Check if file open failed
	if err != nil {
		// Return error with context
		return metrics.SystemCPU{}, fmt.Errorf("open /proc/stat: %w", err)
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	// Scan through lines looking for the cpu line
	for scanner.Scan() {
		line := scanner.Text()
		// Check if this is the aggregate cpu line
		if strings.HasPrefix(line, "cpu ") {
			// Parse and return CPU metrics from this line
			return c.parseCPULine(line)
		}
	}

	// Check if scanner encountered an error
	if err := scanner.Err(); err != nil {
		// Return scanner error with context
		return metrics.SystemCPU{}, fmt.Errorf("scan /proc/stat: %w", err)
	}

	// Return error if cpu line was not found
	return metrics.SystemCPU{}, fmt.Errorf("cpu line not found in /proc/stat: %w", os.ErrNotExist)
}

// parseCPULine parses a cpu line from /proc/stat.
// Format: cpu user nice system idle iowait irq softirq steal guest guest_nice
//
// Params:
//   - line: raw line from /proc/stat starting with "cpu "
//
// Returns:
//   - metrics.SystemCPU: parsed CPU metrics
//   - error: error if parsing fails
func (c *CPUCollector) parseCPULine(line string) (metrics.SystemCPU, error) {
	fields := strings.Fields(line)
	// Validate minimum number of fields are present
	if len(fields) < minCPUFields {
		// Return error if line format is invalid
		return metrics.SystemCPU{}, fmt.Errorf("%w: %s", ErrInvalidCPULine, line)
	}

	cpu := metrics.SystemCPU{Timestamp: time.Now()}

	// Define helper to safely parse field values
	parseField := func(index int) uint64 {
		// Check if index is out of bounds
		if index >= len(fields) {
			// Return zero for missing optional fields
			return 0
		}
		val, _ := strconv.ParseUint(fields[index], parseBase, parseBitSize)
		// Return parsed value (errors default to 0)
		return val
	}

	cpu.User = parseField(cpuFieldUser)
	cpu.Nice = parseField(cpuFieldNice)
	cpu.System = parseField(cpuFieldSystem)
	cpu.Idle = parseField(cpuFieldIdle)
	cpu.IOWait = parseField(cpuFieldIOWait)
	cpu.IRQ = parseField(cpuFieldIRQ)
	cpu.SoftIRQ = parseField(cpuFieldSoftIRQ)
	cpu.Steal = parseField(cpuFieldSteal)
	cpu.Guest = parseField(cpuFieldGuest)
	cpu.GuestNice = parseField(cpuFieldGuestNice)

	// Note: UsagePercent requires comparing two snapshots and should be
	// calculated at the application layer, not from a single sample.
	// A single sample only gives average usage since boot.

	// Return populated CPU metrics
	return cpu, nil
}

// CollectProcess collects CPU metrics for a specific process from /proc/[pid]/stat.
//
// Params:
//   - ctx: context for cancellation
//   - pid: process ID to collect metrics for
//
// Returns:
//   - metrics.ProcessCPU: process CPU metrics
//   - error: error if collection fails
func (c *CPUCollector) CollectProcess(ctx context.Context, pid int) (metrics.ProcessCPU, error) {
	// Check if context is already cancelled before proceeding
	select {
	case <-ctx.Done():
		// Return context error if cancelled
		return metrics.ProcessCPU{}, ctx.Err()
	default:
	}

	// Validate PID is positive
	if pid <= 0 {
		// Return error for invalid PID
		return metrics.ProcessCPU{}, fmt.Errorf("%w: %d", ErrInvalidPID, pid)
	}

	statPath := filepath.Join(c.procPath, strconv.Itoa(pid), "stat")
	// Read the entire stat file
	data, err := os.ReadFile(statPath)
	// Check if file read failed
	if err != nil {
		// Return error with context
		return metrics.ProcessCPU{}, fmt.Errorf("read /proc/%d/stat: %w", pid, err)
	}

	// Parse the stat file content
	return c.parseProcessStat(pid, string(data))
}

// parseProcessStat parses /proc/[pid]/stat content.
// The format is complex because the command name (field 2) can contain spaces and parentheses.
// Fields after (comm): state(3), ppid(4), ..., utime(14), stime(15), cutime(16), cstime(17), ..., starttime(22)
//
// Params:
//   - pid: process ID for error reporting
//   - data: raw content from /proc/[pid]/stat
//
// Returns:
//   - metrics.ProcessCPU: parsed process CPU metrics
//   - error: error if parsing fails
func (c *CPUCollector) parseProcessStat(pid int, data string) (metrics.ProcessCPU, error) {
	// Find the command name between parentheses
	start := strings.Index(data, "(")
	end := strings.LastIndex(data, ")")
	// Validate parentheses are present and properly ordered
	if start == -1 || end == -1 || end <= start {
		// Return error for malformed stat file
		return metrics.ProcessCPU{}, fmt.Errorf("%w for pid %d", ErrInvalidStatFormat, pid)
	}

	name := data[start+1 : end]
	// Fields after the closing parenthesis
	rest := strings.Fields(data[end+statCommEndOffset:])

	// Validate sufficient fields are present
	if len(rest) < minStatFields {
		// Return error if stat file has insufficient fields
		return metrics.ProcessCPU{}, fmt.Errorf("%w for pid %d", ErrInsufficientStatFields, pid)
	}

	// Define helper to safely parse field values
	parseField := func(index int) uint64 {
		// Check if index is out of bounds
		if index >= len(rest) {
			// Return zero for missing fields
			return 0
		}
		val, _ := strconv.ParseUint(rest[index], parseBase, parseBitSize)
		// Return parsed value (errors default to 0)
		return val
	}

	proc := metrics.ProcessCPU{
		PID:            pid,
		Name:           name,
		User:           parseField(statFieldUTime),     // utime (field 14, 0-indexed from rest is 11)
		System:         parseField(statFieldSTime),     // stime (field 15)
		ChildrenUser:   parseField(statFieldCUTime),    // cutime (field 16)
		ChildrenSystem: parseField(statFieldCSTime),    // cstime (field 17)
		StartTime:      parseField(statFieldStartTime), // starttime (field 22)
		Timestamp:      time.Now(),
	}

	// Return populated process metrics
	return proc, nil
}

// CollectAllProcesses collects CPU metrics for all visible processes.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - []metrics.ProcessCPU: slice of process CPU metrics
//   - error: error if collection fails
func (c *CPUCollector) CollectAllProcesses(ctx context.Context) ([]metrics.ProcessCPU, error) {
	// Check if context is already cancelled before proceeding
	select {
	case <-ctx.Done():
		// Return context error if cancelled
		return nil, ctx.Err()
	default:
	}

	// Read /proc directory to enumerate processes
	entries, err := os.ReadDir(c.procPath)
	// Check if directory read failed
	if err != nil {
		// Return error with context
		return nil, fmt.Errorf("read /proc: %w", err)
	}

	// Collect metrics from all process entries
	return c.collectFromEntries(ctx, entries)
}

// collectFromEntries collects CPU metrics from directory entries.
//
// Params:
//   - ctx: context for cancellation
//   - entries: directory entries from /proc
//
// Returns:
//   - []metrics.ProcessCPU: slice of process CPU metrics
//   - error: error if context is cancelled
func (c *CPUCollector) collectFromEntries(ctx context.Context, entries []os.DirEntry) ([]metrics.ProcessCPU, error) {
	results := make([]metrics.ProcessCPU, 0, len(entries))
	// Iterate through all entries in /proc
	for _, entry := range entries {
		// Check for context cancellation to allow early abort
		select {
		case <-ctx.Done():
			// Return context error if cancelled during iteration
			return nil, ctx.Err()
		default:
		}

		// Skip non-directory entries
		if !entry.IsDir() {
			continue
		}

		// Try to parse entry name as PID
		pid, err := strconv.Atoi(entry.Name())
		// Skip if not a valid PID (e.g., "self", "thread-self")
		if err != nil {
			continue // Not a PID directory
		}

		// Collect metrics for this process
		proc, err := c.CollectProcess(ctx, pid)
		// Skip if process collection failed (e.g., process exited)
		if err != nil {
			continue // Process may have exited
		}

		results = append(results, proc)
	}

	// Return all collected process metrics
	return results, nil
}
