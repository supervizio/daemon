// Package lifecycle provides domain types for daemon lifecycle management.
package lifecycle

import "time"

// HostInfo contains information about the host system.
//
// This provides metadata about the daemon's execution environment,
// including OS details, daemon identity, and start time.
type HostInfo struct {
	// Hostname is the system hostname.
	Hostname string `json:"hostname"`
	// OS is the operating system name (e.g., "linux").
	OS string `json:"os"`
	// Arch is the CPU architecture (e.g., "amd64").
	Arch string `json:"arch"`
	// KernelVersion is the kernel version string.
	KernelVersion string `json:"kernel_version"`
	// DaemonPID is the process ID of the daemon.
	DaemonPID int `json:"daemon_pid"`
	// DaemonVersion is the version of the daemon.
	DaemonVersion string `json:"daemon_version"`
	// StartTime is when the daemon started.
	StartTime time.Time `json:"start_time"`
}

// Uptime returns the duration since the daemon started.
//
// Returns:
//   - time.Duration: time elapsed since StartTime, or 0 if StartTime is zero
func (h HostInfo) Uptime() time.Duration {
	if h.StartTime.IsZero() {
		return 0
	}
	return time.Since(h.StartTime)
}
