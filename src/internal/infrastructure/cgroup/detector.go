//go:build linux

// Package cgroup provides adapters for reading cgroup metrics.
package cgroup

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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
	return len(content) > 0 && (strings.Contains(content, "/docker/") ||
		strings.Contains(content, "/kubepods/") ||
		strings.Contains(content, "/lxc/") ||
		strings.Contains(content, "/containerd/"))
}

// Reader is the interface for reading cgroup metrics.
type Reader interface {
	CPUUsage(ctx context.Context) (uint64, error)
	CPULimit(ctx context.Context) (quota, period uint64, err error)
	MemoryUsage(ctx context.Context) (uint64, error)
	MemoryLimit(ctx context.Context) (uint64, error)
	Path() string
}

// NewReader creates a cgroup reader based on the detected version.
// Returns an error if the cgroup version is not supported.
func NewReader() (Reader, error) {
	return NewReaderWithPath("")
}

// NewReaderWithPath creates a cgroup reader for the specified path.
// If path is empty, it auto-detects the current cgroup.
func NewReaderWithPath(path string) (Reader, error) {
	version := Detect()
	switch version {
	case VersionV2, VersionHybrid:
		return NewV2Reader(path)
	case VersionV1:
		return nil, ErrV1NotSupported
	case VersionUnknown:
		return nil, ErrUnknownVersion
	}
	return nil, ErrUnknownVersion
}
