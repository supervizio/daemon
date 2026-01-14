// Package state provides domain types for daemon state representation.
package state

import "time"

// HostInfo contains information about the host system.
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
func (h HostInfo) Uptime() time.Duration {
	if h.StartTime.IsZero() {
		return 0
	}
	return time.Since(h.StartTime)
}
