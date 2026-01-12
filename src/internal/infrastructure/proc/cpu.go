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

// CPUCollector implements metrics.CPUCollector by reading from /proc filesystem.
type CPUCollector struct {
	procPath string
}

// NewCPUCollector creates a new CPU collector using the default /proc path.
func NewCPUCollector() *CPUCollector {
	return &CPUCollector{procPath: "/proc"}
}

// NewCPUCollectorWithPath creates a new CPU collector with a custom proc path.
// This is useful for testing with mock /proc filesystems.
func NewCPUCollectorWithPath(procPath string) *CPUCollector {
	return &CPUCollector{procPath: procPath}
}

// CollectSystem collects system-wide CPU metrics from /proc/stat.
func (c *CPUCollector) CollectSystem(ctx context.Context) (metrics.SystemCPU, error) {
	select {
	case <-ctx.Done():
		return metrics.SystemCPU{}, ctx.Err()
	default:
	}

	file, err := os.Open(filepath.Join(c.procPath, "stat"))
	if err != nil {
		return metrics.SystemCPU{}, fmt.Errorf("open /proc/stat: %w", err)
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			return c.parseCPULine(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return metrics.SystemCPU{}, fmt.Errorf("scan /proc/stat: %w", err)
	}

	return metrics.SystemCPU{}, fmt.Errorf("cpu line not found in /proc/stat")
}

// parseCPULine parses a cpu line from /proc/stat.
// Format: cpu user nice system idle iowait irq softirq steal guest guest_nice
func (c *CPUCollector) parseCPULine(line string) (metrics.SystemCPU, error) {
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return metrics.SystemCPU{}, fmt.Errorf("invalid cpu line format: %s", line)
	}

	cpu := metrics.SystemCPU{Timestamp: time.Now()}

	parseField := func(index int) uint64 {
		if index >= len(fields) {
			return 0
		}
		val, _ := strconv.ParseUint(fields[index], 10, 64)
		return val
	}

	cpu.User = parseField(1)
	cpu.Nice = parseField(2)
	cpu.System = parseField(3)
	cpu.Idle = parseField(4)
	cpu.IOWait = parseField(5)
	cpu.IRQ = parseField(6)
	cpu.SoftIRQ = parseField(7)
	cpu.Steal = parseField(8)
	cpu.Guest = parseField(9)
	cpu.GuestNice = parseField(10)

	// Note: UsagePercent requires comparing two snapshots and should be
	// calculated at the application layer, not from a single sample.
	// A single sample only gives average usage since boot.

	return cpu, nil
}

// CollectProcess collects CPU metrics for a specific process from /proc/[pid]/stat.
func (c *CPUCollector) CollectProcess(ctx context.Context, pid int) (metrics.ProcessCPU, error) {
	select {
	case <-ctx.Done():
		return metrics.ProcessCPU{}, ctx.Err()
	default:
	}

	if pid <= 0 {
		return metrics.ProcessCPU{}, fmt.Errorf("invalid pid: %d", pid)
	}

	statPath := filepath.Join(c.procPath, strconv.Itoa(pid), "stat")
	data, err := os.ReadFile(statPath)
	if err != nil {
		return metrics.ProcessCPU{}, fmt.Errorf("read /proc/%d/stat: %w", pid, err)
	}

	return c.parseProcessStat(pid, string(data))
}

// parseProcessStat parses /proc/[pid]/stat content.
// The format is complex because the command name (field 2) can contain spaces and parentheses.
// Fields after (comm): state(3), ppid(4), ..., utime(14), stime(15), cutime(16), cstime(17), ..., starttime(22)
func (c *CPUCollector) parseProcessStat(pid int, data string) (metrics.ProcessCPU, error) {
	// Find the command name between parentheses
	start := strings.Index(data, "(")
	end := strings.LastIndex(data, ")")
	if start == -1 || end == -1 || end <= start {
		return metrics.ProcessCPU{}, fmt.Errorf("invalid stat format for pid %d", pid)
	}

	name := data[start+1 : end]
	// Fields after the closing parenthesis
	rest := strings.Fields(data[end+2:])

	if len(rest) < 20 {
		return metrics.ProcessCPU{}, fmt.Errorf("insufficient fields in stat for pid %d", pid)
	}

	parseField := func(index int) uint64 {
		if index >= len(rest) {
			return 0
		}
		val, _ := strconv.ParseUint(rest[index], 10, 64)
		return val
	}

	proc := metrics.ProcessCPU{
		PID:            pid,
		Name:           name,
		User:           parseField(11), // utime (field 14, 0-indexed from rest is 11)
		System:         parseField(12), // stime (field 15)
		ChildrenUser:   parseField(13), // cutime (field 16)
		ChildrenSystem: parseField(14), // cstime (field 17)
		StartTime:      parseField(19), // starttime (field 22)
		Timestamp:      time.Now(),
	}

	return proc, nil
}

// CollectAllProcesses collects CPU metrics for all visible processes.
func (c *CPUCollector) CollectAllProcesses(ctx context.Context) ([]metrics.ProcessCPU, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	entries, err := os.ReadDir(c.procPath)
	if err != nil {
		return nil, fmt.Errorf("read /proc: %w", err)
	}

	var results []metrics.ProcessCPU
	for _, entry := range entries {
		// Check for context cancellation to allow early abort
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

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

		results = append(results, proc)
	}

	return results, nil
}
