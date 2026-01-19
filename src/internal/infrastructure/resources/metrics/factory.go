// Package metrics provides platform detection and factory for metrics collectors.
package metrics

import (
	"os"
	"runtime"

	"github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/infrastructure/resources/metrics/scratch"
)

// Platform constants.
const (
	platformDarwin string = "darwin"
	platformLinux  string = "linux"
)

// Platform detection functions that can be overridden in tests.
var (
	// currentGOOS returns the current operating system.
	// This can be overridden in tests to simulate different platforms.
	currentGOOS func() string = func() string {
		return runtime.GOOS
	}

	// statProc checks if /proc/stat exists.
	// This can be overridden in tests to simulate presence/absence of /proc.
	statProc func() error = func() error {
		_, err := os.Stat("/proc/stat")
		return err
	}
)

// NewSystemCollector creates a SystemCollector appropriate for the current platform.
// It automatically detects the best available implementation.
// Detection order:
//  1. Linux with /proc filesystem -> linux.Probe
//  2. FreeBSD/OpenBSD/NetBSD -> bsd.Probe (TODO)
//  3. Darwin (macOS) -> darwin.Probe (TODO)
//  4. Fallback -> scratch.Probe
//
// Returns:
//   - metrics.SystemCollector: platform-specific metrics collector
func NewSystemCollector() metrics.SystemCollector {
	// Select appropriate collector based on platform detection.
	switch {
	// Check for Linux with /proc filesystem.
	case hasProc() && currentGOOS() == platformLinux:
		// Linux with /proc filesystem available
		// Note: The linux adapter only partially implements the interface.
		// For now, fall through to scratch as a placeholder.
		// TODO: Return linux.NewProbe() when fully implemented.
		return scratch.NewProbe()

	// Check for BSD family operating systems.
	case isBSD():
		// BSD family (FreeBSD, OpenBSD, NetBSD)
		// TODO: Return bsd.NewProbe() when implemented.
		return scratch.NewProbe()

	// Check for macOS.
	case currentGOOS() == platformDarwin:
		// macOS
		// TODO: Return darwin.NewProbe() when implemented.
		return scratch.NewProbe()

	// Use fallback for all other platforms.
	default:
		// Fallback for scratch containers, Windows, or unknown platforms
		return scratch.NewProbe()
	}
}

// hasProc checks if the /proc filesystem is available and readable.
//
// Returns:
//   - bool: true if /proc/stat exists and is readable
func hasProc() bool {
	// Attempt to stat the /proc/stat file.
	err := statProc()
	// Return true if no error occurred.
	return err == nil
}

// isBSD returns true if running on a BSD variant.
//
// Returns:
//   - bool: true if running on FreeBSD, OpenBSD, NetBSD, or DragonFly BSD
func isBSD() bool {
	// Check runtime OS against known BSD variants.
	switch currentGOOS() {
	// Recognized BSD operating systems.
	case "freebsd", "openbsd", "netbsd", "dragonfly":
		// OS is a BSD variant.
		return true
	// All other operating systems.
	default:
		// OS is not a BSD variant.
		return false
	}
}

// DetectedPlatform returns a string describing the detected platform.
// This is useful for logging and diagnostics.
//
// Returns:
//   - string: platform identifier (e.g., "linux-proc", "bsd-freebsd", "darwin", "scratch-windows")
func DetectedPlatform() string {
	goos := currentGOOS()
	// Determine platform identifier based on detection logic.
	switch {
	// Check for Linux with /proc filesystem.
	case hasProc() && goos == platformLinux:
		// Return Linux with /proc indicator.
		return "linux-proc"
	// Check for BSD family operating systems.
	case isBSD():
		// Return BSD variant with specific OS name.
		return "bsd-" + goos
	// Check for macOS.
	case goos == platformDarwin:
		// Return Darwin (macOS) indicator.
		return "darwin"
	// All other platforms use scratch fallback.
	default:
		// Return scratch mode with OS name.
		return "scratch-" + goos
	}
}
