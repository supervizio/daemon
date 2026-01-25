//go:build linux

package collector

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// getInterfaceStats reads network stats from sysfs.
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
	speedPath := filepath.Join("/sys/class/net", name, "speed")
	if content, err := os.ReadFile(speedPath); err == nil {
		if val, err := strconv.ParseUint(strings.TrimSpace(string(content)), 10, 64); err == nil {
			// sysfs reports in Mbps, convert to bps.
			speed = val * 1000 * 1000
		}
	}

	return rxBytes, txBytes, speed
}
