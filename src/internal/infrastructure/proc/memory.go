//go:build linux

// Package proc provides Linux /proc filesystem adapters for metrics collection.
package proc

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

// MemoryCollector implements metrics.MemoryCollector by reading from /proc filesystem.
type MemoryCollector struct {
	procPath string
}

// NewMemoryCollector creates a new memory collector using the default /proc path.
func NewMemoryCollector() *MemoryCollector {
	return &MemoryCollector{procPath: "/proc"}
}

// NewMemoryCollectorWithPath creates a new memory collector with a custom proc path.
// This is useful for testing with mock /proc filesystems.
func NewMemoryCollectorWithPath(procPath string) *MemoryCollector {
	return &MemoryCollector{procPath: procPath}
}

// CollectSystem collects system-wide memory metrics from /proc/meminfo.
func (c *MemoryCollector) CollectSystem(ctx context.Context) (metrics.SystemMemory, error) {
	select {
	case <-ctx.Done():
		return metrics.SystemMemory{}, ctx.Err()
	default:
	}

	file, err := os.Open(filepath.Join(c.procPath, "meminfo"))
	if err != nil {
		return metrics.SystemMemory{}, fmt.Errorf("open /proc/meminfo: %w", err)
	}
	defer func() { _ = file.Close() }()

	mem := metrics.SystemMemory{Timestamp: time.Now()}
	values := make(map[string]uint64)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		key, value := c.parseMemInfoLine(line)
		if key != "" {
			values[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return metrics.SystemMemory{}, fmt.Errorf("scan /proc/meminfo: %w", err)
	}

	// Map values to struct (values are in kB, convert to bytes)
	mem.Total = values["MemTotal"] * 1024
	mem.Free = values["MemFree"] * 1024
	mem.Available = values["MemAvailable"] * 1024
	mem.Buffers = values["Buffers"] * 1024
	mem.Cached = values["Cached"] * 1024
	mem.SwapTotal = values["SwapTotal"] * 1024
	mem.SwapFree = values["SwapFree"] * 1024
	mem.Shared = values["Shmem"] * 1024

	// Calculate derived values
	mem.SwapUsed = mem.SwapTotal - mem.SwapFree
	mem.Used = mem.Total - mem.Available

	// Calculate usage percentage
	if mem.Total > 0 {
		mem.UsagePercent = float64(mem.Used) / float64(mem.Total) * 100
	}

	return mem, nil
}

// parseMemInfoLine parses a single line from /proc/meminfo.
// Format: "FieldName:       12345 kB" or "FieldName:       12345"
func (c *MemoryCollector) parseMemInfoLine(line string) (string, uint64) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", 0
	}

	key := strings.TrimSpace(parts[0])
	valueStr := strings.TrimSpace(parts[1])

	// Remove "kB" suffix if present
	valueStr = strings.TrimSuffix(valueStr, " kB")
	valueStr = strings.TrimSpace(valueStr)

	value, err := strconv.ParseUint(valueStr, 10, 64)
	if err != nil {
		return key, 0
	}

	return key, value
}

// CollectProcess collects memory metrics for a specific process from /proc/[pid]/status.
func (c *MemoryCollector) CollectProcess(ctx context.Context, pid int) (metrics.ProcessMemory, error) {
	select {
	case <-ctx.Done():
		return metrics.ProcessMemory{}, ctx.Err()
	default:
	}

	statusPath := filepath.Join(c.procPath, strconv.Itoa(pid), "status")
	file, err := os.Open(statusPath)
	if err != nil {
		return metrics.ProcessMemory{}, fmt.Errorf("open /proc/%d/status: %w", pid, err)
	}
	defer func() { _ = file.Close() }()

	proc := metrics.ProcessMemory{
		PID:       pid,
		Timestamp: time.Now(),
	}
	values := make(map[string]uint64)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Name:") {
			proc.Name = strings.TrimSpace(strings.TrimPrefix(line, "Name:"))
		} else {
			key, value := c.parseStatusLine(line)
			if key != "" {
				values[key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return metrics.ProcessMemory{}, fmt.Errorf("scan /proc/%d/status: %w", pid, err)
	}

	// Map values to struct (values are in kB, convert to bytes)
	proc.RSS = values["VmRSS"] * 1024
	proc.VMS = values["VmSize"] * 1024
	proc.Swap = values["VmSwap"] * 1024
	proc.Data = values["VmData"] * 1024
	proc.Stack = values["VmStk"] * 1024

	// Shared = RssShmem + RssFile (if available)
	proc.Shared = (values["RssShmem"] + values["RssFile"]) * 1024

	return proc, nil
}

// parseStatusLine parses a single line from /proc/[pid]/status for memory values.
// Format: "VmRSS:       12345 kB"
func (c *MemoryCollector) parseStatusLine(line string) (string, uint64) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", 0
	}

	key := strings.TrimSpace(parts[0])

	// Only parse memory-related fields
	if !strings.HasPrefix(key, "Vm") && !strings.HasPrefix(key, "Rss") {
		return "", 0
	}

	valueStr := strings.TrimSpace(parts[1])
	valueStr = strings.TrimSuffix(valueStr, " kB")
	valueStr = strings.TrimSpace(valueStr)

	value, err := strconv.ParseUint(valueStr, 10, 64)
	if err != nil {
		return key, 0
	}

	return key, value
}

// CollectAllProcesses collects memory metrics for all visible processes.
func (c *MemoryCollector) CollectAllProcesses(ctx context.Context) ([]metrics.ProcessMemory, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Get total system memory for percentage calculation
	sysMem, err := c.CollectSystem(ctx)
	if err != nil {
		return nil, fmt.Errorf("collect system memory: %w", err)
	}

	entries, err := os.ReadDir(c.procPath)
	if err != nil {
		return nil, fmt.Errorf("read /proc: %w", err)
	}

	var results []metrics.ProcessMemory
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue // Not a PID directory
		}

		proc, err := c.CollectProcess(ctx, pid)
		if err != nil {
			continue // Process may have exited
		}

		// Calculate usage percentage relative to total system memory
		if sysMem.Total > 0 {
			proc.UsagePercent = float64(proc.RSS) / float64(sysMem.Total) * 100
		}

		results = append(results, proc)
	}

	return results, nil
}
