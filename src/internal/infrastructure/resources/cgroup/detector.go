//go:build linux

// Package cgroup provides adapters for reading cgroup metrics.
package cgroup

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/kodflow/daemon/internal/domain/shared"
)

// DefaultCgroupPath is the default cgroup filesystem path.
const DefaultCgroupPath string = "/sys/fs/cgroup"

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
//
// Returns:
//   - string: human-readable version name ("unknown", "v1", "v2", "hybrid")
func (v Version) String() string {
	// Match version constant to its string representation
	switch v {
	// Handle unknown version
	case VersionUnknown:
		// Return unknown identifier
		return "unknown"
	// Handle v1 legacy cgroups
	case VersionV1:
		// Return v1 identifier
		return "v1"
	// Handle v2 unified hierarchy
	case VersionV2:
		// Return v2 identifier
		return "v2"
	// Handle hybrid mode (v1 + v2 coexistence)
	case VersionHybrid:
		// Return hybrid identifier
		return "hybrid"
	}
	// Return default for unrecognized values
	return "unknown"
}

// Detect detects the cgroup version in use.
//
// Returns:
//   - Version: detected cgroup version (VersionV1, VersionV2, VersionHybrid, or VersionUnknown)
func Detect() Version {
	// Use default cgroup path for detection
	return DetectWithPath(DefaultCgroupPath)
}

// DetectWithPath detects the cgroup version using a custom path.
//
// Params:
//   - cgroupPath: filesystem path to cgroup root
//
// Returns:
//   - Version: detected cgroup version
func DetectWithPath(cgroupPath string) Version {
	// Check for cgroup v2 (unified hierarchy)
	// In v2, /sys/fs/cgroup/cgroup.controllers exists
	controllersPath := filepath.Join(cgroupPath, "cgroup.controllers")
	// Test if controllers file exists (v2 marker)
	if _, err := os.Stat(controllersPath); err == nil {
		// Check if this is pure v2 or hybrid
		// In hybrid mode, v1 controllers exist alongside v2
		cpuPath := filepath.Join(cgroupPath, "cpu")
		memoryPath := filepath.Join(cgroupPath, "memory")

		cpuInfo, cpuErr := os.Stat(cpuPath)
		memInfo, memErr := os.Stat(memoryPath)

		// Both v1 controllers exist as directories means hybrid mode
		if cpuErr == nil && cpuInfo.IsDir() && memErr == nil && memInfo.IsDir() {
			// Return hybrid version (v1 + v2 coexist)
			return VersionHybrid
		}
		// Only v2 controllers exist (pure unified hierarchy)
		return VersionV2
	}

	// Check for cgroup v1
	// In v1, /sys/fs/cgroup/cpu and /sys/fs/cgroup/memory exist
	cpuPath := filepath.Join(cgroupPath, "cpu")
	// Test if cpu controller directory exists (v1 marker)
	if _, err := os.Stat(cpuPath); err == nil {
		// Return v1 legacy version
		return VersionV1
	}

	// No recognizable cgroup structure found
	return VersionUnknown
}

// IsContainerized attempts to detect if we're running in a container.
//
// Returns:
//   - bool: true if running in a container, false otherwise
func IsContainerized() bool {
	// Delegate to injectable version with default filesystem
	return IsContainerizedWithFS(shared.DefaultFileSystem)
}

// IsContainerizedWithFS detects if running in a container using the provided filesystem.
// This function allows dependency injection for testing.
//
// Params:
//   - fs: filesystem interface for file operations
//
// Returns:
//   - bool: true if running in a container, false otherwise
func IsContainerizedWithFS(fs shared.FileSystem) bool {
	// Check for /.dockerenv
	// Docker marker file exists in all Docker containers
	if _, err := fs.Stat("/.dockerenv"); err == nil {
		// Found Docker marker file
		return true
	}

	// Check /proc/1/cgroup for container indicators
	data, err := fs.ReadFile("/proc/1/cgroup")
	// File read failed (not fatal for detection)
	if err != nil {
		// Assume not containerized if can't read cgroup info
		return false
	}

	content := string(data)
	// Docker/containerd typically have paths like /docker/<id> or /kubepods/<id>
	// Return true if container runtime markers found in cgroup path
	return content != "" && (strings.Contains(content, "/docker/") ||
		strings.Contains(content, "/kubepods/") ||
		strings.Contains(content, "/lxc/") ||
		strings.Contains(content, "/containerd/"))
}

// Reader is the interface for reading cgroup metrics.
// Implementations: V1Reader (legacy), V2Reader (unified).
type Reader interface {
	// CPUUsage returns the total CPU usage in microseconds.
	CPUUsage(ctx context.Context) (uint64, error)
	// CPULimit returns the CPU quota and period. Returns (0, 0) if unlimited.
	CPULimit(ctx context.Context) (quota, period uint64, err error)
	// MemoryUsage returns the current memory usage in bytes.
	MemoryUsage(ctx context.Context) (uint64, error)
	// MemoryLimit returns the memory limit in bytes. Returns 0 if unlimited.
	MemoryLimit(ctx context.Context) (uint64, error)
	// ReadMemoryStat returns detailed memory statistics.
	ReadMemoryStat(ctx context.Context) (MemoryStat, error)
	// Path returns the primary cgroup path.
	Path() string
	// Version returns the cgroup version (1 or 2).
	Version() int
}

// NewReader creates a cgroup reader based on the detected version.
// Returns an error if the cgroup version is not supported.
//
// Returns:
//   - Reader: cgroup reader instance
//   - error: ErrUnknownVersion if cgroup version cannot be detected
func NewReader() (Reader, error) {
	// Use auto-detection with empty path
	return NewReaderWithPath("")
}

// NewReaderWithPath creates a cgroup reader for the specified path.
// If path is empty, it auto-detects the current cgroup.
//
// Params:
//   - path: cgroup path (empty for auto-detection)
//
// Returns:
//   - Reader: cgroup reader instance
//   - error: ErrUnknownVersion if cgroup version is not supported
func NewReaderWithPath(path string) (Reader, error) {
	version := Detect()
	// Select reader implementation based on detected version
	switch version {
	// Unified hierarchy (v2) or hybrid mode
	case VersionV2, VersionHybrid:
		// Create v2 reader (handles both v2 and hybrid)
		return NewV2Reader(path)
	// Legacy cgroups (v1)
	case VersionV1:
		// For V1, path is ignored - we auto-detect CPU and memory paths
		// Create v1 reader with auto-detected paths
		return NewV1Reader("", "")
	// Unknown version cannot be handled
	case VersionUnknown:
		// Return error for unknown version
		return nil, ErrUnknownVersion
	}
	// Fallback return (should never reach here)
	return nil, ErrUnknownVersion
}
