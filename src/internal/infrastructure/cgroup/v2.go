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

// V2Reader reads metrics from cgroups v2.
type V2Reader struct {
	path string
}

// NewV2Reader creates a new cgroups v2 reader.
// If path is empty, it attempts to detect the current cgroup.
func NewV2Reader(path string) (*V2Reader, error) {
	if path == "" {
		detected, err := detectCurrentCgroup()
		if err != nil {
			return nil, fmt.Errorf("detect cgroup: %w", err)
		}
		path = detected
	}

	// Verify the path exists
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("cgroup path not found: %w", err)
	}

	return &V2Reader{path: path}, nil
}

// detectCurrentCgroup reads /proc/self/cgroup to find the current cgroup path.
func detectCurrentCgroup() (string, error) {
	data, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		return "", err
	}

	// For cgroup v2, the format is "0::/path"
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "0::") {
			cgroupPath := strings.TrimPrefix(line, "0::")
			cgroupPath = strings.TrimSpace(cgroupPath)
			return filepath.Join(DefaultCgroupPath, cgroupPath), nil
		}
	}

	return "", ErrPathNotFound
}

// CPUUsage returns the total CPU usage in microseconds.
func (r *V2Reader) CPUUsage(ctx context.Context) (uint64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	stat, err := r.readCPUStat()
	if err != nil {
		return 0, err
	}

	return stat.UsageUsec, nil
}

// CPULimit returns the CPU quota and period.
// Returns (0, 0) if no limit is set.
func (r *V2Reader) CPULimit(ctx context.Context) (quota, period uint64, err error) {
	if err := ctx.Err(); err != nil {
		return 0, 0, err
	}

	data, err := os.ReadFile(filepath.Join(r.path, "cpu.max"))
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, nil // No limit
		}
		return 0, 0, fmt.Errorf("read cpu.max: %w", err)
	}

	content := strings.TrimSpace(string(data))
	parts := strings.Fields(content)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid cpu.max format: %s", content)
	}

	// First field is quota (or "max" for unlimited)
	if parts[0] != "max" {
		quota, err = strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("parse quota: %w", err)
		}
	}

	// Second field is period
	period, err = strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse period: %w", err)
	}

	return quota, period, nil
}

// CPUStat contains parsed CPU statistics.
type CPUStat struct {
	UsageUsec  uint64
	UserUsec   uint64
	SystemUsec uint64
}

// readCPUStat parses cpu.stat.
func (r *V2Reader) readCPUStat() (CPUStat, error) {
	statPath := filepath.Join(r.path, "cpu.stat")
	file, err := os.Open(statPath) //nolint:gosec // Path is constructed from validated cgroup path
	if err != nil {
		return CPUStat{}, fmt.Errorf("open cpu.stat: %w", err)
	}
	defer func() { _ = file.Close() }()

	var stat CPUStat
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}

		value, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			continue
		}

		switch parts[0] {
		case "usage_usec":
			stat.UsageUsec = value
		case "user_usec":
			stat.UserUsec = value
		case "system_usec":
			stat.SystemUsec = value
		}
	}

	return stat, scanner.Err()
}

// MemoryUsage returns the current memory usage in bytes.
func (r *V2Reader) MemoryUsage(ctx context.Context) (uint64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	data, err := os.ReadFile(filepath.Join(r.path, "memory.current"))
	if err != nil {
		return 0, fmt.Errorf("read memory.current: %w", err)
	}

	value, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse memory.current: %w", err)
	}

	return value, nil
}

// MemoryLimit returns the memory limit in bytes.
// Returns 0 if no limit is set (or limit is "max").
func (r *V2Reader) MemoryLimit(ctx context.Context) (uint64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	data, err := os.ReadFile(filepath.Join(r.path, "memory.max"))
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil // No limit
		}
		return 0, fmt.Errorf("read memory.max: %w", err)
	}

	content := strings.TrimSpace(string(data))
	if content == "max" {
		return 0, nil // Unlimited
	}

	value, err := strconv.ParseUint(content, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse memory.max: %w", err)
	}

	return value, nil
}

// MemoryStat contains parsed memory statistics.
type MemoryStat struct {
	Anon       uint64
	File       uint64
	Kernel     uint64
	Slab       uint64
	Sock       uint64
	Shmem      uint64
	Mapped     uint64
	Dirty      uint64
	Pgfault    uint64
	Pgmajfault uint64
}

// ReadMemoryStat parses memory.stat.
func (r *V2Reader) ReadMemoryStat(ctx context.Context) (MemoryStat, error) {
	if err := ctx.Err(); err != nil {
		return MemoryStat{}, err
	}

	statPath := filepath.Join(r.path, "memory.stat")
	file, err := os.Open(statPath) //nolint:gosec // Path is constructed from validated cgroup path
	if err != nil {
		return MemoryStat{}, fmt.Errorf("open memory.stat: %w", err)
	}
	defer func() { _ = file.Close() }()

	var stat MemoryStat
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}

		value, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			continue
		}

		switch parts[0] {
		case "anon":
			stat.Anon = value
		case "file":
			stat.File = value
		case "kernel":
			stat.Kernel = value
		case "slab":
			stat.Slab = value
		case "sock":
			stat.Sock = value
		case "shmem":
			stat.Shmem = value
		case "mapped":
			stat.Mapped = value
		case "dirty":
			stat.Dirty = value
		case "pgfault":
			stat.Pgfault = value
		case "pgmajfault":
			stat.Pgmajfault = value
		}
	}

	return stat, scanner.Err()
}

// Path returns the cgroup path.
func (r *V2Reader) Path() string {
	return r.path
}
