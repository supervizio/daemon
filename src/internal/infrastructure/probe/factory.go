// Package probe provides platform detection and factory for metrics collectors.
package probe

import (
	"os"
	"runtime"

	"github.com/kodflow/daemon/internal/domain/probe"
	"github.com/kodflow/daemon/internal/infrastructure/probe/scratch"
)

// NewSystemCollector creates a SystemCollector appropriate for the current platform.
// It automatically detects the best available implementation.
//
// Detection order:
// 1. Linux with /proc filesystem -> linux.LinuxProbe
// 2. FreeBSD/OpenBSD/NetBSD -> bsd.BSDProbe (TODO)
// 3. Darwin (macOS) -> darwin.DarwinProbe (TODO)
// 4. Fallback -> scratch.ScratchProbe
func NewSystemCollector() probe.SystemCollector {
	switch {
	case hasProc() && runtime.GOOS == "linux":
		// Linux with /proc filesystem available
		// Note: The linux adapter only partially implements the interface.
		// For now, fall through to scratch as a placeholder.
		// TODO: Return linux.NewLinuxProbe() when fully implemented.
		return scratch.NewScratchProbe()

	case isBSD():
		// BSD family (FreeBSD, OpenBSD, NetBSD)
		// TODO: Return bsd.NewBSDProbe() when implemented.
		return scratch.NewScratchProbe()

	case runtime.GOOS == "darwin":
		// macOS
		// TODO: Return darwin.NewDarwinProbe() when implemented.
		return scratch.NewScratchProbe()

	default:
		// Fallback for scratch containers, Windows, or unknown platforms
		return scratch.NewScratchProbe()
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
	case hasProc() && runtime.GOOS == "linux":
		return "linux-proc"
	case isBSD():
		return "bsd-" + runtime.GOOS
	case runtime.GOOS == "darwin":
		return "darwin"
	default:
		return "scratch-" + runtime.GOOS
	}
}
