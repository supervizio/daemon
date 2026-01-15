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

// DefaultCgroupV1CPUPath is the default path for cgroup v1 CPU controller.
const DefaultCgroupV1CPUPath = "/sys/fs/cgroup/cpu"

// DefaultCgroupV1MemoryPath is the default path for cgroup v1 memory controller.
const DefaultCgroupV1MemoryPath = "/sys/fs/cgroup/memory"

// V1Reader reads metrics from cgroups v1 (legacy).
type V1Reader struct {
	cpuPath    string
	memoryPath string
}

// NewV1Reader creates a new cgroups v1 reader.
// If paths are empty, it attempts to detect the current cgroup.
func NewV1Reader(cpuPath, memoryPath string) (*V1Reader, error) {
	if cpuPath == "" {
		detected, err := detectV1CPUCgroup()
		if err != nil {
			return nil, fmt.Errorf("detect cpu cgroup: %w", err)
		}
		cpuPath = detected
	}

	if memoryPath == "" {
		detected, err := detectV1MemoryCgroup()
		if err != nil {
			return nil, fmt.Errorf("detect memory cgroup: %w", err)
		}
		memoryPath = detected
	}

	// Verify the paths exist
	if _, err := os.Stat(cpuPath); err != nil {
		return nil, fmt.Errorf("cpu cgroup path not found: %w", err)
	}
	if _, err := os.Stat(memoryPath); err != nil {
		return nil, fmt.Errorf("memory cgroup path not found: %w", err)
	}

	return &V1Reader{cpuPath: cpuPath, memoryPath: memoryPath}, nil
}

// detectV1CPUCgroup reads /proc/self/cgroup to find the current CPU cgroup path.
func detectV1CPUCgroup() (string, error) {
	return detectV1Cgroup("cpu")
}

// detectV1MemoryCgroup reads /proc/self/cgroup to find the current memory cgroup path.
func detectV1MemoryCgroup() (string, error) {
	return detectV1Cgroup("memory")
}

// detectV1Cgroup reads /proc/self/cgroup to find the cgroup path for a controller.
func detectV1Cgroup(controller string) (string, error) {
	data, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		return "", err
	}

	// For cgroup v1, format is "hierarchy-ID:controller-list:cgroup-path"
	// Example: "3:cpu,cpuacct:/docker/abc123"
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 3)
		if len(parts) != 3 {
			continue
		}

		controllers := strings.Split(parts[1], ",")
		for _, c := range controllers {
			if c == controller {
				cgroupPath := strings.TrimSpace(parts[2])
				var basePath string
				switch controller {
				case "cpu", "cpuacct":
					basePath = DefaultCgroupV1CPUPath
				case "memory":
					basePath = DefaultCgroupV1MemoryPath
				default:
					basePath = DefaultCgroupPath + "/" + controller
				}
				return filepath.Join(basePath, cgroupPath), nil
			}
		}
	}

	return "", ErrPathNotFound
}

// CPUUsage returns the total CPU usage in microseconds.
func (r *V1Reader) CPUUsage(ctx context.Context) (uint64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	// Read cpuacct.usage which is in nanoseconds
	data, err := os.ReadFile(filepath.Join(r.cpuPath, "cpuacct.usage"))
	if err != nil {
		return 0, fmt.Errorf("read cpuacct.usage: %w", err)
	}

	nanoseconds, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse cpuacct.usage: %w", err)
	}

	// Convert nanoseconds to microseconds
	return nanoseconds / 1000, nil
}

// CPULimit returns the CPU quota and period.
// Returns (0, 0) if no limit is set.
func (r *V1Reader) CPULimit(ctx context.Context) (quota, period uint64, err error) {
	if err := ctx.Err(); err != nil {
		return 0, 0, err
	}

	// Read cpu.cfs_quota_us
	quotaData, err := os.ReadFile(filepath.Join(r.cpuPath, "cpu.cfs_quota_us"))
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, nil // No limit
		}
		return 0, 0, fmt.Errorf("read cpu.cfs_quota_us: %w", err)
	}

	quotaStr := strings.TrimSpace(string(quotaData))
	if quotaStr == "-1" {
		return 0, 0, nil // Unlimited
	}

	quota, err = strconv.ParseUint(quotaStr, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse quota: %w", err)
	}

	// Read cpu.cfs_period_us
	periodData, err := os.ReadFile(filepath.Join(r.cpuPath, "cpu.cfs_period_us"))
	if err != nil {
		return 0, 0, fmt.Errorf("read cpu.cfs_period_us: %w", err)
	}

	period, err = strconv.ParseUint(strings.TrimSpace(string(periodData)), 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse period: %w", err)
	}

	return quota, period, nil
}

// MemoryUsage returns the current memory usage in bytes.
func (r *V1Reader) MemoryUsage(ctx context.Context) (uint64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	data, err := os.ReadFile(filepath.Join(r.memoryPath, "memory.usage_in_bytes"))
	if err != nil {
		return 0, fmt.Errorf("read memory.usage_in_bytes: %w", err)
	}

	value, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse memory.usage_in_bytes: %w", err)
	}

	return value, nil
}

// MemoryLimit returns the memory limit in bytes.
// Returns 0 if no limit is set.
func (r *V1Reader) MemoryLimit(ctx context.Context) (uint64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	data, err := os.ReadFile(filepath.Join(r.memoryPath, "memory.limit_in_bytes"))
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil // No limit
		}
		return 0, fmt.Errorf("read memory.limit_in_bytes: %w", err)
	}

	value, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse memory.limit_in_bytes: %w", err)
	}

	// Very large values indicate no limit (kernel sets to PAGE_COUNTER_MAX)
	// Typically 9223372036854771712 or similar
	if value > 1<<62 {
		return 0, nil // Effectively unlimited
	}

	return value, nil
}

// ReadMemoryStat returns detailed memory statistics.
func (r *V1Reader) ReadMemoryStat(ctx context.Context) (MemoryStat, error) {
	if err := ctx.Err(); err != nil {
		return MemoryStat{}, err
	}

	statPath := filepath.Join(r.memoryPath, "memory.stat")
	file, err := os.Open(statPath)
	if err != nil {
		return MemoryStat{}, fmt.Errorf("open memory.stat: %w", err)
	}
	defer func() { _ = file.Close() }()

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

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}

		if field, ok := fieldMap[parts[0]]; ok {
			if value, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
				*field = value
			}
		}
	}

	return stat, scanner.Err()
}

// Path returns the CPU cgroup path (primary path for v1).
func (r *V1Reader) Path() string {
	return r.cpuPath
}

// MemoryPath returns the memory cgroup path.
func (r *V1Reader) MemoryPath() string {
	return r.memoryPath
}

// Version returns the cgroup version (always 1 for V1Reader).
func (r *V1Reader) Version() int {
	return 1
}
