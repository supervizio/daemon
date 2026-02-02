//go:build linux

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

const (
	// cpuFieldUser is the index of user CPU time in /proc/stat fields.
	cpuFieldUser int = 1
	// cpuFieldNice is the index of nice CPU time in /proc/stat fields.
	cpuFieldNice int = 2
	// cpuFieldSystem is the index of system CPU time in /proc/stat fields.
	cpuFieldSystem int = 3
	// cpuFieldIdle is the index of idle CPU time in /proc/stat fields.
	cpuFieldIdle int = 4
	// cpuFieldIOWait is the index of I/O wait time in /proc/stat fields.
	cpuFieldIOWait int = 5
	// cpuFieldIRQ is the index of IRQ time in /proc/stat fields.
	cpuFieldIRQ int = 6
	// cpuFieldSoftIRQ is the index of soft IRQ time in /proc/stat fields.
	cpuFieldSoftIRQ int = 7
	// cpuFieldSteal is the index of steal time in /proc/stat fields.
	cpuFieldSteal int = 8

	// minCPUStatFields is the minimum number of fields required (label + 7 values).
	minCPUStatFields int = 8
	// minFieldsForSteal is the minimum number of fields to include steal time.
	minFieldsForSteal int = 9
	// percentageScale is the multiplier to convert fractions to percentage.
	percentageScale float64 = 100.0

	// bytesPerKB is the conversion factor from kilobytes to bytes.
	bytesPerKB uint64 = 1024
	// minLoadAvgFields is the minimum number of fields expected in /proc/loadavg.
	minLoadAvgFields int = 3
	// decimalBase10 is the base for decimal number parsing.
	decimalBase10 int = 10

	// memInfoPreallocFactor is the multiplier for pre-allocating meminfo map capacity.
	memInfoPreallocFactor int = 2

	// loadAvgIndex1 is the index for 1-minute load average in /proc/loadavg.
	loadAvgIndex1 int = 0
	// loadAvgIndex5 is the index for 5-minute load average in /proc/loadavg.
	loadAvgIndex5 int = 1
	// loadAvgIndex15 is the index for 15-minute load average in /proc/loadavg.
	loadAvgIndex15 int = 2
)

// SystemCollector collects system-wide metrics.
// Fields are reused across calls to minimize allocations.
type SystemCollector struct {
	prevCPU     cpuSample
	prevSampled time.Time

	// Reusable buffers to avoid allocations on every tick.
	memValues map[string]uint64 // Reused for /proc/meminfo parsing.
}

// NewSystemCollector creates a new system collector.
// Pre-allocates buffers for zero-allocation collection.
//
// Returns:
//   - *SystemCollector: configured system collector with pre-allocated buffers
func NewSystemCollector() *SystemCollector {
	// Return new collector with pre-allocated map.
	return &SystemCollector{
		// Pre-allocate for common meminfo keys.
		memValues: make(map[string]uint64, minCPUStatFields*memInfoPreallocFactor),
	}
}

// Gather collects system metrics.
//
// Params:
//   - snap: target snapshot to populate with system metrics
//
// Returns:
//   - error: always returns nil as collection errors are handled internally
func (c *SystemCollector) Gather(snap *model.Snapshot) error {
	// Collect CPU.
	c.collectCPU(snap)

	// Collect memory.
	c.collectMemory(snap)

	// Collect load average.
	c.collectLoadAvg(snap)

	// Collect disk usage.
	c.collectDisk(snap)

	// Return nil for graceful operation.
	return nil
}

// collectCPU reads /proc/stat for CPU usage.
//
// Params:
//   - snap: target snapshot to populate with CPU metrics
func (c *SystemCollector) collectCPU(snap *model.Snapshot) {
	f, err := os.Open("/proc/stat")
	// Handle file open error.
	if err != nil {
		// Cannot read CPU stats.
		return
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	// Scan through /proc/stat lines.
	for scanner.Scan() {
		line := scanner.Text()
		// Skip non-aggregate CPU lines.
		if !strings.HasPrefix(line, "cpu ") {
			// Not the aggregate CPU line.
			continue
		}

		// Parse and process CPU sample.
		c.processCPULine(line, snap)
		break
	}
}

// processCPULine parses a CPU stat line and calculates metrics.
//
// Params:
//   - line: raw line from /proc/stat
//   - snap: target snapshot to populate
func (c *SystemCollector) processCPULine(line string, snap *model.Snapshot) {
	fields := strings.Fields(line)
	// Validate field count.
	if len(fields) < minCPUStatFields {
		// Insufficient fields.
		return
	}

	sample := parseCPUSample(fields)
	now := time.Now()

	// Calculate delta if we have a previous sample.
	if !c.prevSampled.IsZero() {
		c.calculateCPUMetrics(sample, snap)
	}

	c.prevCPU = sample
	c.prevSampled = now
}

// parseCPUSample parses CPU fields into a sample.
//
// Params:
//   - fields: fields from /proc/stat CPU line
//
// Returns:
//   - cpuSample: parsed CPU sample
func parseCPUSample(fields []string) cpuSample {
	sample := cpuSample{
		user:    parseUint64(fields[cpuFieldUser]),
		nice:    parseUint64(fields[cpuFieldNice]),
		system:  parseUint64(fields[cpuFieldSystem]),
		idle:    parseUint64(fields[cpuFieldIdle]),
		iowait:  parseUint64(fields[cpuFieldIOWait]),
		irq:     parseUint64(fields[cpuFieldIRQ]),
		softirq: parseUint64(fields[cpuFieldSoftIRQ]),
	}
	// Check if steal field is available.
	if len(fields) >= minFieldsForSteal {
		sample.steal = parseUint64(fields[cpuFieldSteal])
	}
	// Return parsed sample.
	return sample
}

// calculateCPUMetrics calculates CPU percentages from sample delta.
//
// Params:
//   - sample: current CPU sample
//   - snap: target snapshot to populate
func (c *SystemCollector) calculateCPUMetrics(sample cpuSample, snap *model.Snapshot) {
	totalDelta := sample.total() - c.prevCPU.total()
	// Ensure positive delta to avoid division issues.
	if totalDelta == 0 {
		// No delta, skip calculation.
		return
	}

	busyDelta := sample.busy() - c.prevCPU.busy()
	idleDelta := sample.idle - c.prevCPU.idle
	userDelta := sample.user + sample.nice - c.prevCPU.user - c.prevCPU.nice
	systemDelta := sample.system + sample.irq + sample.softirq - c.prevCPU.system - c.prevCPU.irq - c.prevCPU.softirq
	iowaitDelta := sample.iowait - c.prevCPU.iowait

	snap.System.CPUPercent = float64(busyDelta) / float64(totalDelta) * percentageScale
	snap.System.CPUIdle = float64(idleDelta) / float64(totalDelta) * percentageScale
	snap.System.CPUUser = float64(userDelta) / float64(totalDelta) * percentageScale
	snap.System.CPUSystem = float64(systemDelta) / float64(totalDelta) * percentageScale
	snap.System.CPUIOWait = float64(iowaitDelta) / float64(totalDelta) * percentageScale
}

// collectMemory reads /proc/meminfo for memory usage.
// Reuses c.memValues map to avoid allocations.
//
// Params:
//   - snap: target snapshot to populate with memory metrics
func (c *SystemCollector) collectMemory(snap *model.Snapshot) {
	f, err := os.Open("/proc/meminfo")
	// Handle file open error.
	if err != nil {
		// Cannot read memory info.
		return
	}
	defer func() { _ = f.Close() }()

	// Clear and reuse the map instead of allocating new one.
	clear(c.memValues)

	// Parse meminfo into map.
	c.parseMemInfo(f)

	// Populate snapshot from parsed values.
	c.populateMemoryMetrics(snap)
}

// parseMemInfo reads and parses /proc/meminfo into c.memValues.
//
// Params:
//   - f: file handle for /proc/meminfo
func (c *SystemCollector) parseMemInfo(f *os.File) {
	scanner := bufio.NewScanner(f)
	// Parse each line of meminfo.
	for scanner.Scan() {
		key, value := parseMemInfoLine(scanner.Text())
		// Store value if key is valid.
		if key != "" {
			c.memValues[key] = value
		}
	}
}

// parseMemInfoLine parses a single meminfo line.
//
// Params:
//   - line: raw line from /proc/meminfo
//
// Returns:
//   - string: key name or empty on error
//   - uint64: parsed value in bytes
func parseMemInfoLine(line string) (key string, value uint64) {
	// Parse manually to avoid strings.Fields allocation.
	// Format: "KeyName:     12345 kB"
	key, rest, found := strings.Cut(line, ":")
	// Skip lines without colon.
	if !found {
		// Invalid line format.
		return "", 0
	}

	rest = strings.TrimLeft(rest, " \t")

	// Find the numeric value (first field after colon).
	spaceIdx := strings.IndexByte(rest, ' ')
	var valueStr string
	// Extract value string based on space position.
	if spaceIdx > 0 {
		valueStr = rest[:spaceIdx]
	} else {
		// No space found, use entire rest.
		valueStr = rest
	}

	value = parseUint64(valueStr)

	// Values in meminfo are in kB.
	if strings.HasSuffix(rest, " kB") {
		value *= bytesPerKB
	}

	// Return parsed key and value.
	return key, value
}

// populateMemoryMetrics populates snapshot with memory metrics.
//
// Params:
//   - snap: target snapshot to populate
func (c *SystemCollector) populateMemoryMetrics(snap *model.Snapshot) {
	mem := c.memValues

	snap.System.MemoryTotal = mem["MemTotal"]
	snap.System.MemoryAvailable = mem["MemAvailable"]
	snap.System.MemoryCached = mem["Cached"]
	snap.System.MemoryBuffers = mem["Buffers"]
	snap.System.SwapTotal = mem["SwapTotal"]
	snap.System.SwapUsed = mem["SwapTotal"] - mem["SwapFree"]

	// Calculate used memory (excluding buffers/cache).
	if snap.System.MemoryTotal > 0 {
		snap.System.MemoryUsed = snap.System.MemoryTotal - snap.System.MemoryAvailable
		snap.System.MemoryPercent = float64(snap.System.MemoryUsed) / float64(snap.System.MemoryTotal) * percentageScale
	}

	// Calculate swap percent.
	if snap.System.SwapTotal > 0 {
		snap.System.SwapPercent = float64(snap.System.SwapUsed) / float64(snap.System.SwapTotal) * percentageScale
	}
}

// collectLoadAvg reads /proc/loadavg.
//
// Params:
//   - snap: target snapshot to populate with load average metrics
func (c *SystemCollector) collectLoadAvg(snap *model.Snapshot) {
	data, err := os.ReadFile("/proc/loadavg")
	// Handle read error.
	if err != nil {
		// Cannot read load average.
		return
	}

	fields := strings.Fields(string(data))
	// Validate field count.
	if len(fields) < minLoadAvgFields {
		// Insufficient fields.
		return
	}

	snap.System.LoadAvg1 = parseFloat64(fields[loadAvgIndex1])
	snap.System.LoadAvg5 = parseFloat64(fields[loadAvgIndex5])
	snap.System.LoadAvg15 = parseFloat64(fields[loadAvgIndex15])
}

// parseUint64 parses a string to uint64, returning 0 on error.
//
// Params:
//   - s: string to parse
//
// Returns:
//   - uint64: parsed value or 0 on error
func parseUint64(s string) uint64 {
	val, _ := strconv.ParseUint(s, decimalBase10, bitSize64)
	// Return parsed value or zero on error.
	return val
}

// parseFloat64 parses a string to float64, returning 0 on error.
//
// Params:
//   - s: string to parse
//
// Returns:
//   - float64: parsed value or 0 on error
func parseFloat64(s string) float64 {
	val, _ := strconv.ParseFloat(s, bitSize64)
	// Return parsed value or zero on error.
	return val
}

// collectDisk reads root filesystem usage using statfs.
//
// Params:
//   - snap: target snapshot to populate with disk metrics
func (c *SystemCollector) collectDisk(snap *model.Snapshot) {
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	// Handle statfs error.
	if err != nil {
		// Cannot read disk stats.
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
		snap.System.DiskPercent = float64(snap.System.DiskUsed) / float64(snap.System.DiskTotal) * percentageScale
	}
}
