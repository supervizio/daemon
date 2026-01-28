//go:build linux

// Package cgroup provides adapters for reading cgroup metrics.
package cgroup

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// cgroupV2Prefix is the cgroup v2 path prefix in /proc/self/cgroup.
	cgroupV2Prefix string = "0::"
	// cgroupV2Version is the version identifier for cgroup v2.
	cgroupV2Version int = 2
	// baseDecimal is the base for decimal number parsing.
	baseDecimal int = 10
	// bitSize64 is the bit size for uint64 parsing.
	bitSize64 int = 64
	// cpuMaxUnlimited indicates unlimited CPU quota.
	cpuMaxUnlimited string = "max"
	// memoryMaxUnlimited indicates unlimited memory.
	memoryMaxUnlimited string = "max"
	// expectedCPUMaxFields is the number of fields in cpu.max.
	expectedCPUMaxFields int = 2
	// expectedStatFields is the number of fields per line in stat files.
	expectedStatFields int = 2
)

// V2Reader reads metrics from cgroups v2.
// It provides access to CPU and memory metrics from the unified cgroup hierarchy.
type V2Reader struct {
	path string
}

// NewV2Reader creates a new cgroups v2 reader.
// If path is empty, it attempts to detect the current cgroup.
//
// Params:
//   - path: cgroup filesystem path (empty for auto-detection)
//
// Returns:
//   - *V2Reader: configured reader instance
//   - error: detection or validation errors
func NewV2Reader(path string) (*V2Reader, error) {
	// Auto-detect cgroup path if not provided
	if path == "" {
		detected, err := detectCurrentCgroup()
		// Handle detection error
		if err != nil {
			// Return wrapped error
			return nil, fmt.Errorf("detect cgroup: %w", err)
		}
		path = detected
	}

	// Verify the path exists
	if _, err := os.Stat(path); err != nil {
		// Path validation failed - return error
		return nil, fmt.Errorf("cgroup path not found: %w", err)
	}

	// Return configured reader
	return &V2Reader{path: path}, nil
}

// detectCurrentCgroup reads /proc/self/cgroup to find the current cgroup path.
//
// Returns:
//   - string: detected cgroup filesystem path
//   - error: file read or parsing errors
func detectCurrentCgroup() (string, error) {
	// Read process cgroup information
	data, err := os.ReadFile("/proc/self/cgroup")
	// Handle read error
	if err != nil {
		// Return error to caller
		return "", err
	}

	// Parse cgroup v2 format: "0::/path"
	// Iterate through cgroup entries using splitSeq for efficiency
	for line := range splitSeq(string(data), "\n") {
		// Check for v2 format prefix using CutPrefix
		if cgroupPath, found := strings.CutPrefix(line, cgroupV2Prefix); found {
			cgroupPath = strings.TrimSpace(cgroupPath)
			// Return full filesystem path
			return filepath.Join(DefaultCgroupPath, cgroupPath), nil
		}
	}

	// Path not found in cgroup file
	return "", ErrPathNotFound
}

// CPUUsage returns the total CPU usage in microseconds.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - uint64: CPU usage in microseconds
//   - error: context or read errors
func (r *V2Reader) CPUUsage(ctx context.Context) (uint64, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		// Context was cancelled - return error
		return 0, err
	}

	// Read parsed CPU statistics
	stat, err := r.readCPUStat()
	// Handle read error
	if err != nil {
		// Return error to caller
		return 0, err
	}

	// Return CPU usage from statistics
	return stat.UsageUsec, nil
}

// CPULimit returns the CPU quota and period.
// Returns (0, 0) if no limit is set.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - quota: CPU quota in microseconds (0 if unlimited)
//   - period: CPU period in microseconds
//   - err: context or read errors
func (r *V2Reader) CPULimit(ctx context.Context) (quota, period uint64, err error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		// Context was cancelled - return error
		return 0, 0, err
	}

	// Read cpu.max file
	data, err := os.ReadFile(filepath.Join(r.path, "cpu.max"))
	// Handle read error
	if err != nil {
		// No limit if file doesn't exist
		if os.IsNotExist(err) {
			// Return zero values for unlimited
			return 0, 0, nil
		}
		// Return wrapped error
		return 0, 0, fmt.Errorf("read cpu.max: %w", err)
	}

	// Parse cpu.max format: "quota period"
	content := strings.TrimSpace(string(data))
	parts := strings.Fields(content)
	// Validate field count
	if len(parts) != expectedCPUMaxFields {
		// Invalid format - return error
		return 0, 0, &InvalidFormatError{File: "cpu.max", Content: content, Expected: expectedCPUMaxFields, Got: len(parts)}
	}

	// Parse quota (or "max" for unlimited)
	if parts[0] != cpuMaxUnlimited {
		quota, err = strconv.ParseUint(parts[0], baseDecimal, bitSize64)
		// Handle parse error
		if err != nil {
			// Return wrapped error
			return 0, 0, fmt.Errorf("parse quota: %w", err)
		}
	}

	// Parse period value
	period, err = strconv.ParseUint(parts[1], baseDecimal, bitSize64)
	// Handle parse error
	if err != nil {
		// Return wrapped error
		return 0, 0, fmt.Errorf("parse period: %w", err)
	}

	// Return parsed quota and period
	return quota, period, nil
}

// readCPUStat parses cpu.stat.
//
// Returns:
//   - CPUStat: parsed CPU statistics
//   - error: file read or parsing errors
func (r *V2Reader) readCPUStat() (CPUStat, error) {
	// Open cpu.stat file
	statPath := filepath.Join(r.path, "cpu.stat")
	file, err := os.Open(statPath)
	// Handle open error
	if err != nil {
		// Return wrapped error
		return CPUStat{}, fmt.Errorf("open cpu.stat: %w", err)
	}
	defer func() { _ = file.Close() }()

	var stat CPUStat
	scanner := bufio.NewScanner(file)
	// Parse each line: "key value"
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		// Skip malformed lines
		if len(parts) != expectedStatFields {
			continue
		}

		// Parse numeric value
		value, err := strconv.ParseUint(parts[1], baseDecimal, bitSize64)
		// Skip invalid numeric values
		if err != nil {
			continue
		}

		// Map to struct field by key
		switch parts[0] {
		// Total usage in microseconds
		case "usage_usec":
			stat.UsageUsec = value
		// User mode usage in microseconds
		case "user_usec":
			stat.UserUsec = value
		// System mode usage in microseconds
		case "system_usec":
			stat.SystemUsec = value
		}
	}

	// Return parsed statistics and scanner error
	return stat, scanner.Err()
}

// MemoryUsage returns the current memory usage in bytes.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - uint64: current memory usage in bytes
//   - error: context or read errors
func (r *V2Reader) MemoryUsage(ctx context.Context) (uint64, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		// Context was cancelled - return error
		return 0, err
	}

	// Read memory.current file
	data, err := os.ReadFile(filepath.Join(r.path, "memory.current"))
	// Handle read error
	if err != nil {
		// Return wrapped error
		return 0, fmt.Errorf("read memory.current: %w", err)
	}

	// Parse numeric value
	value, err := strconv.ParseUint(strings.TrimSpace(string(data)), baseDecimal, bitSize64)
	// Handle parse error
	if err != nil {
		// Return wrapped error
		return 0, fmt.Errorf("parse memory.current: %w", err)
	}

	// Return parsed memory usage
	return value, nil
}

// MemoryLimit returns the memory limit in bytes.
// Returns 0 if no limit is set (or limit is "max").
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - uint64: memory limit in bytes (0 if unlimited)
//   - error: context or read errors
func (r *V2Reader) MemoryLimit(ctx context.Context) (uint64, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		// Context was cancelled - return error
		return 0, err
	}

	// Read memory.max file
	data, err := os.ReadFile(filepath.Join(r.path, "memory.max"))
	// Handle read error
	if err != nil {
		// No limit if file doesn't exist
		if os.IsNotExist(err) {
			// Return zero for unlimited
			return 0, nil
		}
		// Return wrapped error
		return 0, fmt.Errorf("read memory.max: %w", err)
	}

	// Check for unlimited indicator
	content := strings.TrimSpace(string(data))
	// Return zero for unlimited
	if content == memoryMaxUnlimited {
		// Unlimited memory
		return 0, nil
	}

	// Parse numeric limit
	value, err := strconv.ParseUint(content, baseDecimal, bitSize64)
	// Handle parse error
	if err != nil {
		// Return wrapped error
		return 0, fmt.Errorf("parse memory.max: %w", err)
	}

	// Return parsed memory limit
	return value, nil
}

// ReadMemoryStat parses memory.stat.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - MemoryStat: parsed memory statistics
//   - error: context or read errors
func (r *V2Reader) ReadMemoryStat(ctx context.Context) (MemoryStat, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		// Context was cancelled - return error
		return MemoryStat{}, err
	}

	// Open memory.stat file
	statPath := filepath.Join(r.path, "memory.stat")
	file, err := os.Open(statPath)
	// Handle open error
	if err != nil {
		// Return wrapped error
		return MemoryStat{}, fmt.Errorf("open memory.stat: %w", err)
	}
	defer func() { _ = file.Close() }()

	var stat MemoryStat

	scanner := bufio.NewScanner(file)
	// Parse each line: "key value"
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		// Skip malformed lines
		if len(parts) != expectedStatFields {
			continue
		}

		// Parse and store numeric value if key is recognized
		if value, err := strconv.ParseUint(parts[1], baseDecimal, bitSize64); err == nil {
			stat.setField(parts[0], value)
		}
	}

	// Return parsed statistics and scanner error
	return stat, scanner.Err()
}

// Path returns the cgroup path.
//
// Returns:
//   - string: cgroup filesystem path
func (r *V2Reader) Path() string {
	// Return stored path
	return r.path
}

// Version returns the cgroup version (always 2 for V2Reader).
//
// Returns:
//   - int: cgroup version (2)
func (r *V2Reader) Version() int {
	// Return v2 version identifier
	return cgroupV2Version
}
