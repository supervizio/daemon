// Package collector provides data collection for the TUI.
package collector

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

// SystemCollector collects system-wide metrics.
type SystemCollector struct {
	prevCPU     cpuSample
	prevSampled time.Time
}

// cpuSample holds a CPU sample from /proc/stat.
type cpuSample struct {
	user    uint64
	nice    uint64
	system  uint64
	idle    uint64
	iowait  uint64
	irq     uint64
	softirq uint64
	steal   uint64
}

// total returns the total CPU time.
func (s cpuSample) total() uint64 {
	return s.user + s.nice + s.system + s.idle + s.iowait + s.irq + s.softirq + s.steal
}

// busy returns the busy CPU time (non-idle).
func (s cpuSample) busy() uint64 {
	return s.user + s.nice + s.system + s.irq + s.softirq + s.steal
}

// NewSystemCollector creates a new system collector.
func NewSystemCollector() *SystemCollector {
	return &SystemCollector{}
}

// CollectInto gathers system metrics.
func (c *SystemCollector) CollectInto(snap *model.Snapshot) error {
	// Collect CPU.
	c.collectCPU(snap)

	// Collect memory.
	c.collectMemory(snap)

	// Collect load average.
	c.collectLoadAvg(snap)

	// Collect disk usage.
	c.collectDisk(snap)

	return nil
}

// collectCPU reads /proc/stat for CPU usage.
func (c *SystemCollector) collectCPU(snap *model.Snapshot) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 8 {
			return
		}

		sample := cpuSample{
			user:    parseUint64(fields[1]),
			nice:    parseUint64(fields[2]),
			system:  parseUint64(fields[3]),
			idle:    parseUint64(fields[4]),
			iowait:  parseUint64(fields[5]),
			irq:     parseUint64(fields[6]),
			softirq: parseUint64(fields[7]),
		}
		if len(fields) > 8 {
			sample.steal = parseUint64(fields[8])
		}

		now := time.Now()

		// Calculate delta if we have a previous sample.
		if !c.prevSampled.IsZero() {
			totalDelta := sample.total() - c.prevCPU.total()
			if totalDelta > 0 {
				busyDelta := sample.busy() - c.prevCPU.busy()
				idleDelta := sample.idle - c.prevCPU.idle
				userDelta := sample.user + sample.nice - c.prevCPU.user - c.prevCPU.nice
				systemDelta := sample.system + sample.irq + sample.softirq - c.prevCPU.system - c.prevCPU.irq - c.prevCPU.softirq
				iowaitDelta := sample.iowait - c.prevCPU.iowait

				snap.System.CPUPercent = float64(busyDelta) / float64(totalDelta) * 100
				snap.System.CPUIdle = float64(idleDelta) / float64(totalDelta) * 100
				snap.System.CPUUser = float64(userDelta) / float64(totalDelta) * 100
				snap.System.CPUSystem = float64(systemDelta) / float64(totalDelta) * 100
				snap.System.CPUIOWait = float64(iowaitDelta) / float64(totalDelta) * 100
			}
		}

		c.prevCPU = sample
		c.prevSampled = now
		break
	}
}

// collectMemory reads /proc/meminfo for memory usage.
func (c *SystemCollector) collectMemory(snap *model.Snapshot) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	mem := make(map[string]uint64)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")
		value := parseUint64(fields[1])

		// Values in meminfo are in kB.
		if len(fields) >= 3 && fields[2] == "kB" {
			value *= 1024
		}

		mem[key] = value
	}

	snap.System.MemoryTotal = mem["MemTotal"]
	snap.System.MemoryAvailable = mem["MemAvailable"]
	snap.System.MemoryCached = mem["Cached"]
	snap.System.MemoryBuffers = mem["Buffers"]
	snap.System.SwapTotal = mem["SwapTotal"]
	snap.System.SwapUsed = mem["SwapTotal"] - mem["SwapFree"]

	// Calculate used memory (excluding buffers/cache).
	if snap.System.MemoryTotal > 0 {
		snap.System.MemoryUsed = snap.System.MemoryTotal - snap.System.MemoryAvailable
		snap.System.MemoryPercent = float64(snap.System.MemoryUsed) / float64(snap.System.MemoryTotal) * 100
	}

	// Calculate swap percent.
	if snap.System.SwapTotal > 0 {
		snap.System.SwapPercent = float64(snap.System.SwapUsed) / float64(snap.System.SwapTotal) * 100
	}
}

// collectLoadAvg reads /proc/loadavg.
func (c *SystemCollector) collectLoadAvg(snap *model.Snapshot) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return
	}

	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return
	}

	snap.System.LoadAvg1 = parseFloat64(fields[0])
	snap.System.LoadAvg5 = parseFloat64(fields[1])
	snap.System.LoadAvg15 = parseFloat64(fields[2])
}

// parseUint64 parses a string to uint64, returning 0 on error.
func parseUint64(s string) uint64 {
	v, _ := strconv.ParseUint(s, 10, 64)
	return v
}

// parseFloat64 parses a string to float64, returning 0 on error.
func parseFloat64(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// collectDisk reads root filesystem usage using statfs.
func (c *SystemCollector) collectDisk(snap *model.Snapshot) {
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err != nil {
		return
	}

	// Calculate sizes in bytes.
	blockSize := uint64(stat.Bsize)
	snap.System.DiskTotal = stat.Blocks * blockSize
	snap.System.DiskAvailable = stat.Bavail * blockSize
	snap.System.DiskUsed = (stat.Blocks - stat.Bfree) * blockSize
	snap.System.DiskPath = "/"

	// Calculate percentage.
	if snap.System.DiskTotal > 0 {
		snap.System.DiskPercent = float64(snap.System.DiskUsed) / float64(snap.System.DiskTotal) * 100
	}
}
