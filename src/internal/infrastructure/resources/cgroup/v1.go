//go:build linux

// Package cgroup provides adapters for reading cgroup metrics.
package cgroup

import (
	"bufio"
	"context"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// === Constants ===

// DefaultCgroupV1CPUPath is the default path for cgroup v1 CPU controller.
const DefaultCgroupV1CPUPath string = "/sys/fs/cgroup/cpu"

// DefaultCgroupV1MemoryPath is the default path for cgroup v1 memory controller.
const DefaultCgroupV1MemoryPath string = "/sys/fs/cgroup/memory"

// parseCgroupFieldCount is the expected number of fields in a cgroup line.
const parseCgroupFieldCount int = 3

// memoryStatFieldCount is the expected number of fields in memory.stat line.
const memoryStatFieldCount int = 2

// cgroupPathIndex is the index of the cgroup path in the parsed line.
const cgroupPathIndex int = 2

// nanosecondsToMicroseconds is the conversion factor from nanoseconds to microseconds.
const nanosecondsToMicroseconds uint64 = 1000

// unlimitedQuotaValue represents unlimited CPU quota in cgroup v1.
const unlimitedQuotaValue string = "-1"

// unlimitedMemoryThreshold is the threshold above which memory is considered unlimited.
// This is 2^62, used to detect kernel's PAGE_COUNTER_MAX.
const unlimitedMemoryThreshold uint64 = 1 << 62

// parseBase is the numeric base for parsing cgroup values.
const parseBase int = 10

// parseBitSize is the bit size for parsing uint64 values.
const parseBitSize int = 64

// === Types ===

// V1Reader reads metrics from cgroups v1 (legacy).
// It provides access to CPU and memory statistics from the cgroup v1 hierarchy,
// where CPU and memory controllers have separate mount points.
type V1Reader struct {
	path       string
	memoryPath string
}

// === Constructors ===

// NewV1Reader creates a new cgroups v1 reader.
// If paths are empty, it attempts to detect the current cgroup.
//
// Params:
//   - cpuPath: path to CPU cgroup (empty for auto-detect)
//   - memoryPath: path to memory cgroup (empty for auto-detect)
//
// Returns:
//   - *V1Reader: the cgroup v1 reader instance
//   - error: any error during path detection or validation
func NewV1Reader(cpuPath, memoryPath string) (*V1Reader, error) {
	// Auto-detect CPU cgroup path if not provided
	if cpuPath == "" {
		// Detect CPU cgroup from /proc/self/cgroup
		detected, err := detectV1CPUCgroup()
		// Return error if detection fails
		if err != nil {
			// Wrap error with context
			return nil, fmt.Errorf("detect cpu cgroup: %w", err)
		}
		// Use detected path
		cpuPath = detected
	}

	// Auto-detect memory cgroup path if not provided
	if memoryPath == "" {
		// Detect memory cgroup from /proc/self/cgroup
		detected, err := detectV1MemoryCgroup()
		// Return error if detection fails
		if err != nil {
			// Wrap error with context
			return nil, fmt.Errorf("detect memory cgroup: %w", err)
		}
		// Use detected path
		memoryPath = detected
	}

	// Verify paths exist
	return validateAndCreateReader(cpuPath, memoryPath)
}

// validateAndCreateReader validates paths and creates the reader.
//
// Params:
//   - cpuPath: path to CPU cgroup
//   - memoryPath: path to memory cgroup
//
// Returns:
//   - *V1Reader: the cgroup v1 reader instance
//   - error: any error during path validation
func validateAndCreateReader(cpuPath, memoryPath string) (*V1Reader, error) {
	// Verify CPU path exists
	if _, err := os.Stat(cpuPath); err != nil {
		// Return error if CPU path not found
		return nil, fmt.Errorf("cpu cgroup path not found: %w", err)
	}

	// Verify memory path exists
	if _, err := os.Stat(memoryPath); err != nil {
		// Return error if memory path not found
		return nil, fmt.Errorf("memory cgroup path not found: %w", err)
	}

	// Return configured reader
	return &V1Reader{path: cpuPath, memoryPath: memoryPath}, nil
}

// === Detection Functions ===

// detectV1CPUCgroup reads /proc/self/cgroup to find the current CPU cgroup path.
//
// Returns:
//   - string: the detected CPU cgroup path
//   - error: ErrPathNotFound if CPU cgroup not found
func detectV1CPUCgroup() (string, error) {
	// Delegate to generic controller detection
	return detectV1Cgroup("cpu")
}

// detectV1MemoryCgroup reads /proc/self/cgroup to find the current memory cgroup path.
//
// Returns:
//   - string: the detected memory cgroup path
//   - error: ErrPathNotFound if memory cgroup not found
func detectV1MemoryCgroup() (string, error) {
	// Delegate to generic controller detection
	return detectV1Cgroup("memory")
}

// detectV1Cgroup reads /proc/self/cgroup to find the cgroup path for a controller.
//
// Params:
//   - controller: the cgroup controller name (cpu, memory, etc.)
//
// Returns:
//   - string: the detected cgroup path
//   - error: ErrPathNotFound if controller not found
func detectV1Cgroup(controller string) (string, error) {
	// Read cgroup file for current process
	data, err := os.ReadFile("/proc/self/cgroup")
	// Return error if file read fails
	if err != nil {
		// Return read error
		return "", err
	}

	// Parse and find controller path
	return parseV1CgroupData(string(data), controller)
}

// parseV1CgroupData parses cgroup data to find the path for a controller.
//
// Params:
//   - data: raw cgroup file content
//   - controller: the controller name to find
//
// Returns:
//   - string: the detected cgroup path
//   - error: ErrPathNotFound if controller not found
func parseV1CgroupData(data, controller string) (string, error) {
	// Iterate over each line using SplitSeq for efficiency
	for line := range splitSeq(data, "\n") {
		// Try to find controller in this line
		path, found := parseV1CgroupLine(line, controller)
		// Return path if found
		if found {
			// Return discovered path
			return path, nil
		}
	}

	// Controller not found in cgroup file
	return "", ErrPathNotFound
}

// parseV1CgroupLine parses a single cgroup line for the controller.
//
// Params:
//   - line: a single line from /proc/self/cgroup
//   - controller: the controller name to find
//
// Returns:
//   - string: the cgroup path if found
//   - bool: true if controller was found
func parseV1CgroupLine(line, controller string) (string, bool) {
	// Split line into hierarchy, controllers, and path
	parts := strings.SplitN(line, ":", parseCgroupFieldCount)
	// Skip malformed lines
	if len(parts) != parseCgroupFieldCount {
		// Return not found
		return "", false
	}

	// Check each controller in the list
	for c := range splitSeq(parts[1], ",") {
		// Check if this is the requested controller
		if c == controller {
			// Build and return the full path
			return buildV1CgroupPath(controller, parts[cgroupPathIndex]), true
		}
	}

	// Controller not in this line
	return "", false
}

// buildV1CgroupPath constructs the full cgroup path for a controller.
//
// Params:
//   - controller: the cgroup controller name
//   - cgroupPath: the relative cgroup path from /proc/self/cgroup
//
// Returns:
//   - string: the full filesystem path
func buildV1CgroupPath(controller, cgroupPath string) string {
	// Trim whitespace from cgroup path
	cgroupPath = strings.TrimSpace(cgroupPath)
	// Select base path for controller
	var basePath string
	// Determine base path based on controller type
	switch controller {
	// CPU and cpuacct share the same hierarchy
	case "cpu", "cpuacct":
		// Use CPU base path
		basePath = DefaultCgroupV1CPUPath
	// Memory controller has its own hierarchy
	case "memory":
		// Use memory base path
		basePath = DefaultCgroupV1MemoryPath
	// Other controllers use generic path
	default:
		// Construct path from default cgroup path
		basePath = DefaultCgroupPath + "/" + controller
	}

	// Return full path by joining base and cgroup path
	return filepath.Join(basePath, cgroupPath)
}

// splitSeq returns an iterator over substrings separated by sep.
//
// Params:
//   - s: the string to split
//   - sep: the separator string
//
// Returns:
//   - iter.Seq[string]: iterator over substrings
func splitSeq(s, sep string) iter.Seq[string] {
	// Return iterator function
	return func(yield func(string) bool) {
		// Iterate while string has content
		for {
			// Find next separator
			i := strings.Index(s, sep)
			// No more separators found
			if i < 0 {
				// Yield final part and exit
				_ = yield(s)
				// Exit loop
				return
			}
			// Yield part before separator
			if !yield(s[:i]) {
				// Exit if yield returns false
				return
			}
			// Move past separator
			s = s[i+len(sep):]
		}
	}
}

// === CPU Methods ===

// CPUUsage returns the total CPU usage in microseconds.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - uint64: CPU usage in microseconds
//   - error: any error reading or parsing the value
func (r *V1Reader) CPUUsage(ctx context.Context) (uint64, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		// Return context error
		return 0, err
	}

	// Read cpuacct.usage which is in nanoseconds
	data, err := os.ReadFile(filepath.Join(r.path, "cpuacct.usage"))
	// Return error if file read fails
	if err != nil {
		// Wrap error with context
		return 0, fmt.Errorf("read cpuacct.usage: %w", err)
	}

	// Parse nanoseconds value
	nanoseconds, err := strconv.ParseUint(strings.TrimSpace(string(data)), parseBase, parseBitSize)
	// Return error if parsing fails
	if err != nil {
		// Wrap error with context
		return 0, fmt.Errorf("parse cpuacct.usage: %w", err)
	}

	// Convert nanoseconds to microseconds and return
	return nanoseconds / nanosecondsToMicroseconds, nil
}

// CPULimit returns the CPU quota and period.
// Returns (0, 0) if no limit is set.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - quota: CPU quota in microseconds (0 if unlimited)
//   - period: CPU period in microseconds (0 if unlimited)
//   - err: any error reading or parsing the values
func (r *V1Reader) CPULimit(ctx context.Context) (quota, period uint64, err error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		// Return context error
		return 0, 0, err
	}

	// Read and parse quota
	quota, unlimited, err := r.readCPUQuota()
	// Return on error or unlimited
	if err != nil || unlimited {
		// Return current values
		return 0, 0, err
	}

	// Read and parse period
	period, err = r.readCPUPeriod()
	// Return error if reading fails
	if err != nil {
		// Return with error
		return 0, 0, err
	}

	// Return parsed values
	return quota, period, nil
}

// readCPUQuota reads the CPU quota from cfs_quota_us.
//
// Returns:
//   - uint64: the quota value
//   - bool: true if unlimited
//   - error: any error reading the file
func (r *V1Reader) readCPUQuota() (uint64, bool, error) {
	// Read cpu.cfs_quota_us
	quotaData, err := os.ReadFile(filepath.Join(r.path, "cpu.cfs_quota_us"))
	// Handle file read error
	if err != nil {
		// Check if file does not exist
		if os.IsNotExist(err) {
			// No limit file means unlimited
			return 0, true, nil
		}
		// Wrap error with context
		return 0, false, fmt.Errorf("read cpu.cfs_quota_us: %w", err)
	}

	// Parse quota string
	quotaStr := strings.TrimSpace(string(quotaData))
	// Check for unlimited quota value
	if quotaStr == unlimitedQuotaValue {
		// Return unlimited flag
		return 0, true, nil
	}

	// Parse quota value
	quota, err := strconv.ParseUint(quotaStr, parseBase, parseBitSize)
	// Return error if parsing fails
	if err != nil {
		// Wrap error with context
		return 0, false, fmt.Errorf("parse quota: %w", err)
	}

	// Return quota value
	return quota, false, nil
}

// readCPUPeriod reads the CPU period from cfs_period_us.
//
// Returns:
//   - uint64: the period value
//   - error: any error reading the file
func (r *V1Reader) readCPUPeriod() (uint64, error) {
	// Read cpu.cfs_period_us
	periodData, err := os.ReadFile(filepath.Join(r.path, "cpu.cfs_period_us"))
	// Return error if file read fails
	if err != nil {
		// Wrap error with context
		return 0, fmt.Errorf("read cpu.cfs_period_us: %w", err)
	}

	// Parse period value
	period, err := strconv.ParseUint(strings.TrimSpace(string(periodData)), parseBase, parseBitSize)
	// Return error if parsing fails
	if err != nil {
		// Wrap error with context
		return 0, fmt.Errorf("parse period: %w", err)
	}

	// Return parsed period
	return period, nil
}

// === Memory Methods ===

// MemoryUsage returns the current memory usage in bytes.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - uint64: memory usage in bytes
//   - error: any error reading or parsing the value
func (r *V1Reader) MemoryUsage(ctx context.Context) (uint64, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		// Return context error
		return 0, err
	}

	// Read memory.usage_in_bytes
	data, err := os.ReadFile(filepath.Join(r.memoryPath, "memory.usage_in_bytes"))
	// Return error if file read fails
	if err != nil {
		// Wrap error with context
		return 0, fmt.Errorf("read memory.usage_in_bytes: %w", err)
	}

	// Parse memory usage value
	value, err := strconv.ParseUint(strings.TrimSpace(string(data)), parseBase, parseBitSize)
	// Return error if parsing fails
	if err != nil {
		// Wrap error with context
		return 0, fmt.Errorf("parse memory.usage_in_bytes: %w", err)
	}

	// Return parsed value
	return value, nil
}

// MemoryLimit returns the memory limit in bytes.
// Returns 0 if no limit is set.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - uint64: memory limit in bytes (0 if unlimited)
//   - error: any error reading or parsing the value
func (r *V1Reader) MemoryLimit(ctx context.Context) (uint64, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		// Return context error
		return 0, err
	}

	// Read memory.limit_in_bytes
	data, err := os.ReadFile(filepath.Join(r.memoryPath, "memory.limit_in_bytes"))
	// Handle file read error
	if err != nil {
		// Check if file does not exist
		if os.IsNotExist(err) {
			// No limit file means unlimited
			return 0, nil
		}
		// Wrap error with context
		return 0, fmt.Errorf("read memory.limit_in_bytes: %w", err)
	}

	// Parse and check limit value
	return parseMemoryLimit(data)
}

// parseMemoryLimit parses and validates a memory limit value.
//
// Params:
//   - data: raw bytes from the limit file
//
// Returns:
//   - uint64: memory limit in bytes (0 if unlimited)
//   - error: any error parsing the value
func parseMemoryLimit(data []byte) (uint64, error) {
	// Parse memory limit value
	value, err := strconv.ParseUint(strings.TrimSpace(string(data)), parseBase, parseBitSize)
	// Return error if parsing fails
	if err != nil {
		// Wrap error with context
		return 0, fmt.Errorf("parse memory.limit_in_bytes: %w", err)
	}

	// Check for effectively unlimited value
	if value > unlimitedMemoryThreshold {
		// Return zero to indicate unlimited
		return 0, nil
	}

	// Return parsed value
	return value, nil
}

// ReadMemoryStat returns detailed memory statistics.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - MemoryStat: parsed memory statistics
//   - error: any error reading or parsing the file
func (r *V1Reader) ReadMemoryStat(ctx context.Context) (MemoryStat, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		// Return empty stats with context error
		return MemoryStat{}, err
	}

	// Open memory.stat file
	statPath := filepath.Join(r.memoryPath, "memory.stat")
	file, err := os.Open(statPath)
	// Return error if file open fails
	if err != nil {
		// Wrap error with context
		return MemoryStat{}, fmt.Errorf("open memory.stat: %w", err)
	}
	// Ensure file is closed
	defer func() { _ = file.Close() }()

	// Parse and return stats
	return parseV1MemoryStat(file)
}

// parseV1MemoryStat parses memory statistics from a reader.
//
// Params:
//   - file: file to read from
//
// Returns:
//   - MemoryStat: parsed memory statistics
//   - error: any error during parsing
func parseV1MemoryStat(file *os.File) (MemoryStat, error) {
	// Initialize stats struct
	var stat MemoryStat
	// V1 has different field names than v2
	fieldMap := map[string]*uint64{
		"total_rss":         &stat.Anon,
		"total_cache":       &stat.File,
		"total_shmem":       &stat.Shmem,
		"total_mapped_file": &stat.Mapped,
		"total_dirty":       &stat.Dirty,
		"total_pgfault":     &stat.Pgfault,
		"total_pgmajfault":  &stat.Pgmajfault,
	}

	// Scan file line by line
	scanner := bufio.NewScanner(file)
	// Iterate over each line
	for scanner.Scan() {
		// Parse single line
		parseV1MemoryStatLine(scanner.Text(), fieldMap)
	}

	// Return stats and any scanner error
	return stat, scanner.Err()
}

// parseV1MemoryStatLine parses a single memory.stat line.
//
// Params:
//   - line: the line to parse
//   - fieldMap: map of field names to struct field pointers
func parseV1MemoryStatLine(line string, fieldMap map[string]*uint64) {
	// Split line into key and value
	parts := strings.Fields(line)
	// Skip malformed lines
	if len(parts) != memoryStatFieldCount {
		// Skip this line
		return
	}

	// Check if this field is in our map
	if field, ok := fieldMap[parts[0]]; ok {
		// Parse and assign value if parsing succeeds
		if value, err := strconv.ParseUint(parts[1], parseBase, parseBitSize); err == nil {
			// Assign parsed value to field
			*field = value
		}
	}
}

// === Accessor Methods ===

// Path returns the CPU cgroup path (primary path for v1).
// This method implements the Reader interface.
//
// Returns:
//   - string: the CPU cgroup filesystem path
func (r *V1Reader) Path() string {
	// Return CPU path
	return r.path
}

// MemoryPath returns the memory cgroup path.
//
// Returns:
//   - string: the memory cgroup filesystem path
func (r *V1Reader) MemoryPath() string {
	// Return memory path
	return r.memoryPath
}

// Version returns the cgroup version (always 1 for V1Reader).
//
// Returns:
//   - int: the cgroup version number (1)
func (r *V1Reader) Version() int {
	// Return version 1 for V1Reader
	return 1
}
