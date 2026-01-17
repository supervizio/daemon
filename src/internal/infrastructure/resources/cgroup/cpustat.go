//go:build linux

// Package cgroup provides adapters for reading cgroup metrics.
package cgroup

// CPUStat contains parsed CPU statistics.
// It represents CPU usage metrics from cgroup v2 cpu.stat file.
type CPUStat struct {
	// UsageUsec is total CPU usage in microseconds.
	UsageUsec uint64
	// UserUsec is user-mode CPU usage in microseconds.
	UserUsec uint64
	// SystemUsec is system-mode CPU usage in microseconds.
	SystemUsec uint64
}
