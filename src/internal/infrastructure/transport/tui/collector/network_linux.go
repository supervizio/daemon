//go:build linux

package collector

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// adaptiveSpeed tracks estimated bandwidth per interface.
// Starts at 1 Gbps and auto-scales up based on observed throughput.
var (
	adaptiveSpeedMu sync.RWMutex
	adaptiveSpeed   = make(map[string]uint64)
)

// Standard speed tiers in bps.
const (
	speed1Gbps   = 1_000_000_000
	speed2_5Gbps = 2_500_000_000
	speed5Gbps   = 5_000_000_000
	speed10Gbps  = 10_000_000_000
	speed25Gbps  = 25_000_000_000
	speed40Gbps  = 40_000_000_000
	speed100Gbps = 100_000_000_000
)

// speedTiers defines the auto-scaling tiers.
var speedTiers = []uint64{
	speed1Gbps,
	speed2_5Gbps,
	speed5Gbps,
	speed10Gbps,
	speed25Gbps,
	speed40Gbps,
	speed100Gbps,
}

// getInterfaceStats reads network stats from sysfs.
// For interfaces without hardware speed, uses adaptive estimation starting at 1 Gbps.
func getInterfaceStats(name string) (rxBytes, txBytes, speed uint64) {
	basePath := filepath.Join("/sys/class/net", name, "statistics")

	// RX bytes.
	if content, err := os.ReadFile(filepath.Join(basePath, "rx_bytes")); err == nil {
		if val, err := strconv.ParseUint(strings.TrimSpace(string(content)), 10, 64); err == nil {
			rxBytes = val
		}
	}

	// TX bytes.
	if content, err := os.ReadFile(filepath.Join(basePath, "tx_bytes")); err == nil {
		if val, err := strconv.ParseUint(strings.TrimSpace(string(content)), 10, 64); err == nil {
			txBytes = val
		}
	}

	// Speed (in Mbps from sysfs, convert to bps).
	// Virtual interfaces (lo, veth, docker0) don't have speed - use adaptive fallback.
	speedPath := filepath.Join("/sys/class/net", name, "speed")
	if content, err := os.ReadFile(speedPath); err == nil {
		if val, err := strconv.ParseUint(strings.TrimSpace(string(content)), 10, 64); err == nil {
			// sysfs reports in Mbps, convert to bps.
			// Some virtual interfaces report -1 or very large invalid values.
			if val > 0 && val < 1000000 { // Sanity check (< 1 Tbps)
				speed = val * 1000 * 1000
				return rxBytes, txBytes, speed
			}
		}
	}

	// No hardware speed - use adaptive estimation for virtual interfaces.
	adaptiveSpeedMu.RLock()
	speed = adaptiveSpeed[name]
	adaptiveSpeedMu.RUnlock()

	if speed == 0 {
		// Initialize to 1 Gbps.
		speed = speed1Gbps
		adaptiveSpeedMu.Lock()
		adaptiveSpeed[name] = speed
		adaptiveSpeedMu.Unlock()
	}

	return rxBytes, txBytes, speed
}

// UpdateAdaptiveSpeed adjusts the estimated speed based on observed throughput.
// Call this after calculating bytes/sec from the collector.
func UpdateAdaptiveSpeed(name string, throughputBps uint64) {
	adaptiveSpeedMu.Lock()
	defer adaptiveSpeedMu.Unlock()

	currentSpeed := adaptiveSpeed[name]
	if currentSpeed == 0 {
		currentSpeed = speed1Gbps
	}

	// If throughput exceeds 80% of current estimate, scale up.
	threshold := currentSpeed * 80 / 100
	if throughputBps > threshold {
		// Find next tier.
		for _, tier := range speedTiers {
			if tier > currentSpeed {
				adaptiveSpeed[name] = tier
				break
			}
		}
	}
}
