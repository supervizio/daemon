//go:build linux

// Package collector provides Linux-specific network statistics collection.
package collector

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const (
	// typicalInterfaceMapCap is the typical capacity for interface maps.
	typicalInterfaceMapCap int = 8

	// networkDecimalBase is the base for decimal number parsing.
	networkDecimalBase int = 10

	// networkBitSize64 is the bit size for 64-bit integers.
	networkBitSize64 int = 64

	// percentageThreshold is the throughput percentage threshold to trigger speed scaling.
	percentageThreshold uint64 = 80

	// percentageTotal is the total percentage value.
	percentageTotal uint64 = 100

	// maxSaneMbps is the maximum sane Mbps value (< 1 Tbps).
	maxSaneMbps uint64 = 1000000

	// mbpsMultiplier converts Mbps to bps.
	mbpsMultiplier uint64 = 1000 * 1000
)

// Standard speed tiers in bps.
const (
	speed1Gbps   uint64 = 1_000_000_000
	speed25Gbps  uint64 = 2_500_000_000
	speed5Gbps   uint64 = 5_000_000_000
	speed10Gbps  uint64 = 10_000_000_000
	speed25GbpsV uint64 = 25_000_000_000
	speed40Gbps  uint64 = 40_000_000_000
	speed100Gbps uint64 = 100_000_000_000
)

var (
	// adaptiveSpeedMu protects access to adaptiveSpeed map.
	adaptiveSpeedMu sync.RWMutex
	// adaptiveSpeed tracks estimated bandwidth per interface.
	// Starts at 1 Gbps and auto-scales up based on observed throughput.
	adaptiveSpeed map[string]uint64 = make(map[string]uint64, typicalInterfaceMapCap)

	// speedTiers defines the auto-scaling tiers.
	speedTiers []uint64 = []uint64{
		speed1Gbps,
		speed25Gbps,
		speed5Gbps,
		speed10Gbps,
		speed25GbpsV,
		speed40Gbps,
		speed100Gbps,
	}
)

// getInterfaceStats reads network stats from sysfs.
// For interfaces without hardware speed, uses adaptive estimation starting at 1 Gbps.
//
// Params:
//   - name: interface name
//
// Returns:
//   - rxBytes: received bytes counter
//   - txBytes: transmitted bytes counter
//   - speed: interface speed in bits per second
func getInterfaceStats(name string) (rxBytes, txBytes, speed uint64) {
	basePath := filepath.Join("/sys/class/net", name, "statistics")

	// Read RX and TX bytes.
	rxBytes = readSysfsCounter(filepath.Join(basePath, "rx_bytes"))
	txBytes = readSysfsCounter(filepath.Join(basePath, "tx_bytes"))

	// Get speed (hardware or adaptive).
	speed = getInterfaceSpeed(name)

	// Return all stats.
	return rxBytes, txBytes, speed
}

// readSysfsCounter reads a single counter from sysfs.
//
// Params:
//   - path: path to the sysfs file
//
// Returns:
//   - uint64: counter value or 0 on error
func readSysfsCounter(path string) uint64 {
	content, err := os.ReadFile(path)
	// Handle read error.
	if err != nil {
		// Cannot read counter.
		return 0
	}
	// Parse counter value.
	val, _ := strconv.ParseUint(strings.TrimSpace(string(content)), networkDecimalBase, networkBitSize64)
	// Return parsed value (0 on parse error).
	return val
}

// getInterfaceSpeed returns the interface speed in bps.
// Uses hardware speed if available, otherwise adaptive estimation.
//
// Params:
//   - name: interface name
//
// Returns:
//   - uint64: speed in bits per second
func getInterfaceSpeed(name string) uint64 {
	// Try hardware speed first.
	if speed := readHardwareSpeed(name); speed > 0 {
		// Hardware speed available.
		return speed
	}

	// Fall back to adaptive estimation for virtual interfaces.
	return getAdaptiveSpeed(name)
}

// readHardwareSpeed reads the hardware speed from sysfs.
//
// Params:
//   - name: interface name
//
// Returns:
//   - uint64: speed in bps or 0 if not available
func readHardwareSpeed(name string) uint64 {
	speedPath := filepath.Join("/sys/class/net", name, "speed")
	content, err := os.ReadFile(speedPath)
	// Handle read error.
	if err != nil {
		// Speed file not available.
		return 0
	}

	// Parse speed in Mbps.
	val, err := strconv.ParseUint(strings.TrimSpace(string(content)), networkDecimalBase, networkBitSize64)
	// Handle parse error or invalid value.
	if err != nil || val == 0 || val >= maxSaneMbps {
		// Invalid or out of range speed.
		return 0
	}

	// Convert Mbps to bps.
	return val * mbpsMultiplier
}

// getAdaptiveSpeed returns or initializes the adaptive speed estimate.
//
// Params:
//   - name: interface name
//
// Returns:
//   - uint64: estimated speed in bps
func getAdaptiveSpeed(name string) uint64 {
	adaptiveSpeedMu.RLock()
	speed := adaptiveSpeed[name]
	adaptiveSpeedMu.RUnlock()

	// Check if speed needs initialization.
	if speed == 0 {
		// Initialize to 1 Gbps.
		speed = speed1Gbps
		adaptiveSpeedMu.Lock()
		adaptiveSpeed[name] = speed
		adaptiveSpeedMu.Unlock()
	}

	// Return adaptive speed.
	return speed
}

// UpdateAdaptiveSpeed adjusts the estimated speed based on observed throughput.
// Call this after calculating bytes/sec from the collector.
//
// Params:
//   - name: interface name
//   - throughputBps: observed throughput in bits per second
func UpdateAdaptiveSpeed(name string, throughputBps uint64) {
	adaptiveSpeedMu.Lock()
	defer adaptiveSpeedMu.Unlock()

	currentSpeed := adaptiveSpeed[name]
	// Initialize if not set.
	if currentSpeed == 0 {
		currentSpeed = speed1Gbps
	}

	// If throughput exceeds 80% of current estimate, scale up.
	threshold := currentSpeed * percentageThreshold / percentageTotal
	// Check if scaling is needed.
	if throughputBps > threshold {
		// Find next tier.
		// Iterate through speed tiers.
		for _, tier := range speedTiers {
			// Check if tier is higher than current.
			if tier > currentSpeed {
				adaptiveSpeed[name] = tier
				// Stop after finding next tier.
				break
			}
		}
	}
}
