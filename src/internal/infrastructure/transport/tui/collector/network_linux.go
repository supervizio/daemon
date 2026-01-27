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

	// RX bytes.
	if content, err := os.ReadFile(filepath.Join(basePath, "rx_bytes")); err == nil {
		contentStr := string(content)
		// Parse RX bytes.
		if val, err := strconv.ParseUint(strings.TrimSpace(contentStr), networkDecimalBase, networkBitSize64); err == nil {
			rxBytes = val
		}
	}

	// TX bytes.
	if content, err := os.ReadFile(filepath.Join(basePath, "tx_bytes")); err == nil {
		contentStr := string(content)
		// Parse TX bytes.
		if val, err := strconv.ParseUint(strings.TrimSpace(contentStr), networkDecimalBase, networkBitSize64); err == nil {
			txBytes = val
		}
	}

	// Speed (in Mbps from sysfs, convert to bps).
	// Virtual interfaces (lo, veth, docker0) don't have speed - use adaptive fallback.
	speedPath := filepath.Join("/sys/class/net", name, "speed")
	// Try to read speed file.
	if content, err := os.ReadFile(speedPath); err == nil {
		// Parse speed in Mbps.
		if val, err := strconv.ParseUint(strings.TrimSpace(string(content)), networkDecimalBase, networkBitSize64); err == nil {
			// sysfs reports in Mbps, convert to bps.
			// Some virtual interfaces report -1 or very large invalid values.
			// Validate speed is within sane range.
			if val > 0 && val < maxSaneMbps { // Sanity check (< 1 Tbps)
				speed = val * mbpsMultiplier
				// Return hardware-reported speed.
				return rxBytes, txBytes, speed
			}
		}
	}

	// No hardware speed - use adaptive estimation for virtual interfaces.
	adaptiveSpeedMu.RLock()
	speed = adaptiveSpeed[name]
	adaptiveSpeedMu.RUnlock()

	// Check if speed needs initialization.
	if speed == 0 {
		// Initialize to 1 Gbps.
		speed = speed1Gbps
		adaptiveSpeedMu.Lock()
		adaptiveSpeed[name] = speed
		adaptiveSpeedMu.Unlock()
	}

	// Return with adaptive speed.
	return rxBytes, txBytes, speed
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
