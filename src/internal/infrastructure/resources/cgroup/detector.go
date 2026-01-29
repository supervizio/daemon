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

// String formats the version for display.
//
// Returns:
//   - string: human-readable version name ("v1", "v2", "hybrid", or "unknown")
func (v Version) String() string {
	// Map enum values to display strings.
	switch v {
	// version is unknown or undetectable.
	case VersionUnknown:
		// return unknown label.
		return "unknown"
	// version is cgroup v2 unified hierarchy.
	case VersionV2:
		// return v2 label.
		return "v2"
	// version is legacy cgroup v1.
	case VersionV1:
		// return v1 label.
		return "v1"
	// version is hybrid mode with both v1 and v2.
	case VersionHybrid:
		// return hybrid label.
		return "hybrid"
	// defensive case for future enum additions.
	default:
		// Defensive case for future enum additions.
		return "unknown"
	}
}

// Detect auto-detects the cgroup version from the default path.
//
// Returns:
//   - Version: detected cgroup version (V1, V2, Hybrid, or Unknown)
func Detect() Version { return DetectWithPath(DefaultCgroupPath) }

// DetectWithPath probes filesystem markers to determine cgroup version.
//
// Params:
//   - cgroupPath: root path to cgroup filesystem (typically /sys/fs/cgroup)
//
// Returns:
//   - Version: detected cgroup version (V1, V2, Hybrid, or Unknown)
func DetectWithPath(cgroupPath string) Version {
	// V2 marker file: cgroup.controllers exists at root.
	controllersPath := filepath.Join(cgroupPath, "cgroup.controllers")
	// cgroup.controllers present indicates v2 or hybrid.
	if _, err := os.Stat(controllersPath); err == nil {
		// Check for legacy v1 controller directories coexisting with v2.
		cpuPath := filepath.Join(cgroupPath, "cpu")
		memoryPath := filepath.Join(cgroupPath, "memory")
		cpuInfo, cpuErr := os.Stat(cpuPath)
		memInfo, memErr := os.Stat(memoryPath)
		// Legacy cpu and memory dirs indicate hybrid mode.
		if cpuErr == nil && cpuInfo.IsDir() && memErr == nil && memInfo.IsDir() {
			// found both v1 and v2 hierarchies.
			return VersionHybrid
		}
		// Pure v2: only unified hierarchy.
		return VersionV2
	}
	// V1 detection: legacy cpu controller directory exists.
	cpuPath := filepath.Join(cgroupPath, "cpu")
	// check if v1 cpu controller exists.
	if _, err := os.Stat(cpuPath); err == nil {
		// found v1 cpu controller.
		return VersionV1
	}
	// No cgroup markers found (non-Linux or misconfigured).
	return VersionUnknown
}

// IsContainerized checks for container runtime markers.
//
// Returns:
//   - bool: true when running inside Docker, Kubernetes, LXC, or containerd
func IsContainerized() bool { return IsContainerizedWithFS(shared.DefaultFileSystem) }

// IsContainerizedWithFS allows testing with a mock filesystem.
//
// Params:
//   - fs: filesystem interface for file operations
//
// Returns:
//   - bool: true when container markers are detected
func IsContainerizedWithFS(fs shared.FileSystem) bool {
	// Docker creates /.dockerenv in container root.
	if _, err := fs.Stat("/.dockerenv"); err == nil {
		// docker marker file found.
		return true
	}
	// Check cgroup path of init process for container identifiers.
	data, err := fs.ReadFile("/proc/1/cgroup")
	// Cannot read cgroup info; assume not containerized.
	if err != nil {
		// failed to read proc cgroup.
		return false
	}
	// Look for known container runtime path patterns.
	content := string(data)
	// check for container runtime identifiers in cgroup path.
	return content != "" && (strings.Contains(content, "/docker/") ||
		strings.Contains(content, "/kubepods/") ||
		strings.Contains(content, "/lxc/") ||
		strings.Contains(content, "/containerd/"))
}

// Reader abstracts cgroup v1/v2 metric collection.
type Reader interface {
	CPUUsage(ctx context.Context) (uint64, error)
	CPULimit(ctx context.Context) (quota, period uint64, err error)
	MemoryUsage(ctx context.Context) (uint64, error)
	MemoryLimit(ctx context.Context) (uint64, error)
	ReadMemoryStat(ctx context.Context) (MemoryStat, error)
	Path() string
	Version() int
}

// NewReader auto-detects version and returns an appropriate Reader.
//
// Returns:
//   - Reader: cgroup reader appropriate for detected version
//   - error: ErrUnknownVersion if cgroup type cannot be determined
func NewReader() (Reader, error) { return NewReaderWithPath("") }

// NewReaderWithPath returns a Reader for the specified cgroup path.
//
// Params:
//   - path: cgroup path (empty string for auto-detection)
//
// Returns:
//   - Reader: version-appropriate cgroup reader
//   - error: ErrUnknownVersion if detection fails
func NewReaderWithPath(path string) (Reader, error) {
	version := Detect()
	// Select reader implementation based on detected version.
	switch version {
	// V2 and hybrid both use unified hierarchy for metrics.
	case VersionV2, VersionHybrid:
		// use v2 reader for unified hierarchy.
		return NewV2Reader(path)
	// Legacy v1 uses separate controller hierarchies.
	case VersionV1:
		// use v1 reader for separate hierarchies.
		return NewV1Reader("", "")
	// Unknown version means no cgroup support detected.
	case VersionUnknown:
		// no cgroup support detected.
		return nil, ErrUnknownVersion
	// unreachable but required for exhaustive switch.
	default:
		// Unreachable but required for exhaustive.
		return nil, ErrUnknownVersion
	}
}
