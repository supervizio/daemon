//go:build linux

package collector

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseCPUSample(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fields   []string
		expected cpuSample
	}{
		{
			name:   "full CPU line with steal",
			fields: []string{"cpu", "1000", "100", "500", "8000", "200", "50", "100", "50"},
			expected: cpuSample{
				user:    1000,
				nice:    100,
				system:  500,
				idle:    8000,
				iowait:  200,
				irq:     50,
				softirq: 100,
				steal:   50,
			},
		},
		{
			name:   "CPU line without steal",
			fields: []string{"cpu", "1000", "100", "500", "8000", "200", "50", "100"},
			expected: cpuSample{
				user:    1000,
				nice:    100,
				system:  500,
				idle:    8000,
				iowait:  200,
				irq:     50,
				softirq: 100,
				steal:   0,
			},
		},
		{
			name:   "minimum fields",
			fields: []string{"cpu", "1000", "100", "500", "8000", "200", "50", "100"},
			expected: cpuSample{
				user:    1000,
				nice:    100,
				system:  500,
				idle:    8000,
				iowait:  200,
				irq:     50,
				softirq: 100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := parseCPUSample(tt.fields)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_calculateCPUMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		prevSample    cpuSample
		currentSample cpuSample
		expectCPU     float64
		expectUser    float64
		expectSystem  float64
	}{
		{
			name: "50% CPU usage",
			prevSample: cpuSample{
				user:   1000,
				system: 500,
				idle:   8500,
			},
			currentSample: cpuSample{
				user:   1500,
				system: 1000,
				idle:   12500,
			},
			expectCPU:    20.0, // (500+500) / 5000 * 100
			expectUser:   10.0, // 500 / 5000 * 100
			expectSystem: 10.0, // 500 / 5000 * 100
		},
		{
			name: "100% CPU usage",
			prevSample: cpuSample{
				user:   1000,
				system: 500,
				idle:   8500,
			},
			currentSample: cpuSample{
				user:   2000,
				system: 1500,
				idle:   8500,
			},
			expectCPU:    100.0, // (1000+1000) / 2000 * 100
			expectUser:   50.0,  // 1000 / 2000 * 100
			expectSystem: 50.0,  // 1000 / 2000 * 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewSystemCollector()
			collector.prevCPU = tt.prevSample

			snap := &model.Snapshot{}
			collector.calculateCPUMetrics(tt.currentSample, snap)

			assert.InDelta(t, tt.expectCPU, snap.System.CPUPercent, 0.1)
			assert.InDelta(t, tt.expectUser, snap.System.CPUUser, 0.1)
			assert.InDelta(t, tt.expectSystem, snap.System.CPUSystem, 0.1)
		})
	}

	t.Run("zero delta skips calculation", func(t *testing.T) {
		t.Parallel()

		collector := NewSystemCollector()
		sample := cpuSample{user: 1000, idle: 9000}
		collector.prevCPU = sample

		snap := &model.Snapshot{}
		collector.calculateCPUMetrics(sample, snap)

		// Should not set any values when delta is zero.
		assert.Equal(t, 0.0, snap.System.CPUPercent)
	})
}

func Test_parseMemInfoLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		line          string
		expectedKey   string
		expectedValue uint64
	}{
		{
			name:          "typical meminfo line in kB",
			line:          "MemTotal:       16384000 kB",
			expectedKey:   "MemTotal",
			expectedValue: 16384000 * 1024,
		},
		{
			name:          "line without unit",
			line:          "HugePages_Total:   0",
			expectedKey:   "HugePages_Total",
			expectedValue: 0,
		},
		{
			name:          "line with whitespace",
			line:          "MemFree:           8192000 kB",
			expectedKey:   "MemFree",
			expectedValue: 8192000 * 1024,
		},
		{
			name:          "invalid line without colon",
			line:          "InvalidLine",
			expectedKey:   "",
			expectedValue: 0,
		},
		{
			name:          "line with tabs",
			line:          "SwapTotal:\t\t2048000 kB",
			expectedKey:   "SwapTotal",
			expectedValue: 2048000 * 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			key, value := parseMemInfoLine(tt.line)
			assert.Equal(t, tt.expectedKey, key)
			assert.Equal(t, tt.expectedValue, value)
		})
	}
}

func Test_parseUint64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected uint64
	}{
		{
			name:     "valid number",
			input:    "12345",
			expected: 12345,
		},
		{
			name:     "zero",
			input:    "0",
			expected: 0,
		},
		{
			name:     "max uint64",
			input:    "18446744073709551615",
			expected: 18446744073709551615,
		},
		{
			name:     "invalid input returns zero",
			input:    "invalid",
			expected: 0,
		},
		{
			name:     "empty string returns zero",
			input:    "",
			expected: 0,
		},
		{
			name:     "negative returns zero",
			input:    "-1",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := parseUint64(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_parseFloat64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name:     "valid float",
			input:    "3.14",
			expected: 3.14,
		},
		{
			name:     "integer",
			input:    "42",
			expected: 42.0,
		},
		{
			name:     "zero",
			input:    "0",
			expected: 0.0,
		},
		{
			name:     "negative",
			input:    "-1.5",
			expected: -1.5,
		},
		{
			name:     "scientific notation",
			input:    "1e10",
			expected: 1e10,
		},
		{
			name:     "invalid returns zero",
			input:    "invalid",
			expected: 0.0,
		},
		{
			name:     "empty returns zero",
			input:    "",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := parseFloat64(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_populateMemoryMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		memValues         map[string]uint64
		expectedTotal     uint64
		expectedAvailable uint64
		expectedPercent   float64
	}{
		{
			name: "typical memory values",
			memValues: map[string]uint64{
				"MemTotal":      16384000 * 1024,
				"MemAvailable":  12288000 * 1024,
				"Cached":        2048000 * 1024,
				"Buffers":       512000 * 1024,
				"SwapTotal":     4096000 * 1024,
				"SwapFree":      4096000 * 1024,
			},
			expectedTotal:     16384000 * 1024,
			expectedAvailable: 12288000 * 1024,
			expectedPercent:   25.0, // (16384000 - 12288000) / 16384000 * 100
		},
		{
			name:              "no memory data",
			memValues:         map[string]uint64{},
			expectedTotal:     0,
			expectedAvailable: 0,
			expectedPercent:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewSystemCollector()
			collector.memValues = tt.memValues

			snap := &model.Snapshot{}
			collector.populateMemoryMetrics(snap)

			assert.Equal(t, tt.expectedTotal, snap.System.MemoryTotal)
			assert.Equal(t, tt.expectedAvailable, snap.System.MemoryAvailable)
			assert.InDelta(t, tt.expectedPercent, snap.System.MemoryPercent, 0.1)
		})
	}
}

// Test_SystemCollector_populateMemoryMetrics tests the populateMemoryMetrics method.
func Test_SystemCollector_populateMemoryMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		memValues       map[string]uint64
		expectedPercent float64
		expectedSwap    float64
	}{
		{
			name: "full memory data with swap",
			memValues: map[string]uint64{
				"MemTotal":     8192 * 1024,
				"MemAvailable": 4096 * 1024,
				"Cached":       1024 * 1024,
				"Buffers":      512 * 1024,
				"SwapTotal":    2048 * 1024,
				"SwapFree":     1024 * 1024,
			},
			expectedPercent: 50.0,
			expectedSwap:    50.0,
		},
		{
			name: "no swap configured",
			memValues: map[string]uint64{
				"MemTotal":     8192 * 1024,
				"MemAvailable": 6144 * 1024,
				"SwapTotal":    0,
				"SwapFree":     0,
			},
			expectedPercent: 25.0,
			expectedSwap:    0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := NewSystemCollector()
			c.memValues = tt.memValues

			snap := &model.Snapshot{}
			c.populateMemoryMetrics(snap)

			assert.InDelta(t, tt.expectedPercent, snap.System.MemoryPercent, 0.1)
			assert.InDelta(t, tt.expectedSwap, snap.System.SwapPercent, 0.1)
		})
	}
}

// Test_SystemCollector_collectCPU tests the collectCPU method reads from /proc/stat.
func Test_SystemCollector_collectCPU(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		iterations int
	}{
		{
			name:       "single collection",
			iterations: 1,
		},
		{
			name:       "multiple collections for delta",
			iterations: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := NewSystemCollector()

			// Collect CPU stats.
			for i := range tt.iterations {
				snap := &model.Snapshot{}
				c.collectCPU(snap)

				// First call sets baseline, no percentage.
				if i == 0 {
					// No assertion on CPU percent for first call.
					continue
				}

				// After delta, we should have valid percentages.
				assert.GreaterOrEqual(t, snap.System.CPUPercent, 0.0)
				assert.LessOrEqual(t, snap.System.CPUPercent, 100.0)
			}

			// Verify prevSampled was updated.
			assert.False(t, c.prevSampled.IsZero())
		})
	}
}

// Test_SystemCollector_processCPULine tests the processCPULine method.
func Test_SystemCollector_processCPULine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		line          string
		hasPrevSample bool
		expectMetrics bool
	}{
		{
			name:          "valid line with previous sample",
			line:          "cpu  1000 100 500 8000 200 50 100 50",
			hasPrevSample: true,
			expectMetrics: true,
		},
		{
			name:          "valid line without previous sample",
			line:          "cpu  1000 100 500 8000 200 50 100 50",
			hasPrevSample: false,
			expectMetrics: false,
		},
		{
			name:          "line with insufficient fields",
			line:          "cpu  1000 100 500",
			hasPrevSample: true,
			expectMetrics: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := NewSystemCollector()
			if tt.hasPrevSample {
				c.prevCPU = cpuSample{user: 500, system: 250, idle: 4000}
				c.prevSampled = time.Now().Add(-time.Second)
			}

			snap := &model.Snapshot{}
			c.processCPULine(tt.line, snap)

			if tt.expectMetrics {
				// Should have computed metrics.
				assert.GreaterOrEqual(t, snap.System.CPUPercent, 0.0)
			}

			// prevSampled should be updated for valid lines.
			if len(strings.Fields(tt.line)) >= minCPUStatFields {
				assert.False(t, c.prevSampled.IsZero())
			}
		})
	}
}

// Test_SystemCollector_calculateCPUMetrics tests the calculateCPUMetrics method.
func Test_SystemCollector_calculateCPUMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		prevSample   cpuSample
		currSample   cpuSample
		expectCPU    float64
		expectIOWait float64
	}{
		{
			name: "with I/O wait",
			prevSample: cpuSample{
				user:   1000,
				system: 500,
				idle:   7500,
				iowait: 1000,
			},
			currSample: cpuSample{
				user:   1500,
				system: 750,
				idle:   9500,
				iowait: 2250,
			},
			// totalDelta = 14000 - 10000 = 4000
			// busyDelta = 2250 - 1500 = 750
			// iowaitDelta = 2250 - 1000 = 1250
			expectCPU:    18.75, // 750 / 4000 * 100
			expectIOWait: 31.25, // 1250 / 4000 * 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := NewSystemCollector()
			c.prevCPU = tt.prevSample

			snap := &model.Snapshot{}
			c.calculateCPUMetrics(tt.currSample, snap)

			assert.InDelta(t, tt.expectCPU, snap.System.CPUPercent, 0.5)
			assert.InDelta(t, tt.expectIOWait, snap.System.CPUIOWait, 0.5)
		})
	}
}

// Test_SystemCollector_collectMemory tests the collectMemory method reads from /proc/meminfo.
func Test_SystemCollector_collectMemory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "reads system memory info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := NewSystemCollector()
			snap := &model.Snapshot{}

			c.collectMemory(snap)

			// On a real Linux system, memory total should be positive.
			assert.Greater(t, snap.System.MemoryTotal, uint64(0))
			assert.GreaterOrEqual(t, snap.System.MemoryPercent, 0.0)
			assert.LessOrEqual(t, snap.System.MemoryPercent, 100.0)
		})
	}
}

// Test_SystemCollector_parseMemInfo tests the parseMemInfo method.
func Test_SystemCollector_parseMemInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		content      string
		expectedKeys []string
	}{
		{
			name: "typical meminfo content",
			content: `MemTotal:       16384000 kB
MemFree:         8192000 kB
MemAvailable:   12288000 kB
Buffers:          512000 kB
Cached:          2048000 kB`,
			expectedKeys: []string{"MemTotal", "MemFree", "MemAvailable", "Buffers", "Cached"},
		},
		{
			name:         "empty content",
			content:      "",
			expectedKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := NewSystemCollector()

			// Create temp file with content using t.TempDir().
			tmpDir := t.TempDir()
			tmpPath := tmpDir + "/meminfo_test"
			err := os.WriteFile(tmpPath, []byte(tt.content), 0600)
			require.NoError(t, err)

			tmpFile, err := os.Open(tmpPath)
			require.NoError(t, err)
			defer func() { _ = tmpFile.Close() }()

			c.parseMemInfo(tmpFile)

			for _, key := range tt.expectedKeys {
				_, exists := c.memValues[key]
				assert.True(t, exists, "expected key %s to exist", key)
			}
		})
	}
}

// Test_SystemCollector_collectLoadAvg tests the collectLoadAvg method.
func Test_SystemCollector_collectLoadAvg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "reads system load average",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := NewSystemCollector()
			snap := &model.Snapshot{}

			c.collectLoadAvg(snap)

			// Load averages should be non-negative on a real system.
			assert.GreaterOrEqual(t, snap.System.LoadAvg1, 0.0)
			assert.GreaterOrEqual(t, snap.System.LoadAvg5, 0.0)
			assert.GreaterOrEqual(t, snap.System.LoadAvg15, 0.0)
		})
	}
}

// Test_SystemCollector_collectDisk tests the collectDisk method.
func Test_SystemCollector_collectDisk(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "reads root filesystem stats",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := NewSystemCollector()
			snap := &model.Snapshot{}

			c.collectDisk(snap)

			// On a real Linux system, disk total should be positive.
			assert.Greater(t, snap.System.DiskTotal, uint64(0))
			assert.Equal(t, "/", snap.System.DiskPath)
			assert.GreaterOrEqual(t, snap.System.DiskPercent, 0.0)
			assert.LessOrEqual(t, snap.System.DiskPercent, 100.0)
		})
	}
}

