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
	platformDarwin = "darwin"
	platformLinux  = "linux"
)

// NewSystemCollector creates a SystemCollector appropriate for the current platform.
// It automatically detects the best available implementation.
//
// Detection order:
// 1. Linux with /proc filesystem -> linux.Probe
// 2. FreeBSD/OpenBSD/NetBSD -> bsd.Probe (TODO)
// 3. Darwin (macOS) -> darwin.Probe (TODO)
// 4. Fallback -> scratch.Probe
func NewSystemCollector() metrics.SystemCollector {
	switch {
	case hasProc() && runtime.GOOS == platformLinux:
		// Linux with /proc filesystem available
		// Note: The linux adapter only partially implements the interface.
		// For now, fall through to scratch as a placeholder.
		// TODO: Return linux.NewProbe() when fully implemented.
		return scratch.NewProbe()

	case isBSD():
		// BSD family (FreeBSD, OpenBSD, NetBSD)
		// TODO: Return bsd.NewProbe() when implemented.
		return scratch.NewProbe()

	case runtime.GOOS == platformDarwin:
		// macOS
		// TODO: Return darwin.NewProbe() when implemented.
		return scratch.NewProbe()

	default:
		// Fallback for scratch containers, Windows, or unknown platforms
		return scratch.NewProbe()
	}
}

// hasProc checks if the /proc filesystem is available and readable.
func hasProc() bool {
	_, err := os.Stat("/proc/stat")
	return err == nil
}

// isBSD returns true if running on a BSD variant.
func isBSD() bool {
	switch runtime.GOOS {
	case "freebsd", "openbsd", "netbsd", "dragonfly":
		return true
	default:
		return false
	}
}

// DetectedPlatform returns a string describing the detected platform.
// This is useful for logging and diagnostics.
func DetectedPlatform() string {
	switch {
	case hasProc() && runtime.GOOS == platformLinux:
		return "linux-proc"
	case isBSD():
		return "bsd-" + runtime.GOOS
	case runtime.GOOS == platformDarwin:
		return "darwin"
	default:
		return "scratch-" + runtime.GOOS
	}
}
