//go:build linux

// Package cgroup provides adapters for reading cgroup metrics.
package cgroup

import (
	"os"
	"path/filepath"
)

// Version represents the cgroup version.
type Version int

// Cgroup versions.
const (
	VersionUnknown Version = iota
	VersionV1
	VersionV2
	VersionHybrid
)

// String returns the string representation of the version.
func (v Version) String() string {
	switch v {
	case VersionUnknown:
		return "unknown"
	case VersionV1:
		return "v1"
	case VersionV2:
		return "v2"
	case VersionHybrid:
		return "hybrid"
	}
	return "unknown"
}

// DefaultCgroupPath is the default cgroup filesystem path.
const DefaultCgroupPath = "/sys/fs/cgroup"

// Detect detects the cgroup version in use.
func Detect() Version {
	return DetectWithPath(DefaultCgroupPath)
}

// DetectWithPath detects the cgroup version using a custom path.
func DetectWithPath(cgroupPath string) Version {
	// Check for cgroup v2 (unified hierarchy)
	// In v2, /sys/fs/cgroup/cgroup.controllers exists
	controllersPath := filepath.Join(cgroupPath, "cgroup.controllers")
	if _, err := os.Stat(controllersPath); err == nil {
		// Check if this is pure v2 or hybrid
		// In hybrid mode, v1 controllers exist alongside v2
		cpuPath := filepath.Join(cgroupPath, "cpu")
		memoryPath := filepath.Join(cgroupPath, "memory")

		cpuInfo, cpuErr := os.Stat(cpuPath)
		memInfo, memErr := os.Stat(memoryPath)

		if cpuErr == nil && cpuInfo.IsDir() && memErr == nil && memInfo.IsDir() {
			return VersionHybrid
		}
		return VersionV2
	}

	// Check for cgroup v1
	// In v1, /sys/fs/cgroup/cpu and /sys/fs/cgroup/memory exist
	cpuPath := filepath.Join(cgroupPath, "cpu")
	if _, err := os.Stat(cpuPath); err == nil {
		return VersionV1
	}

	return VersionUnknown
}

// IsContainerized attempts to detect if we're running in a container.
func IsContainerized() bool {
	// Check for /.dockerenv
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check /proc/1/cgroup for container indicators
	data, err := os.ReadFile("/proc/1/cgroup")
	if err != nil {
		return false
	}

	content := string(data)
	// Docker/containerd typically have paths like /docker/<id> or /kubepods/<id>
	return len(content) > 0 && (contains(content, "/docker/") ||
		contains(content, "/kubepods/") ||
		contains(content, "/lxc/") ||
		contains(content, "/containerd/"))
}

// contains is a simple string contains check.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr) >= 0
}

// searchString finds substr in s.
func searchString(s, substr string) int {
	n := len(substr)
	for i := 0; i+n <= len(s); i++ {
		if s[i:i+n] == substr {
			return i
		}
	}
	return -1
}
