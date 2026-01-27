//go:build linux

// Package collector provides Linux-specific cgroup limit collection.
package collector

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

const (
	// cgroupMax is the value used in cgroups to indicate unlimited.
	cgroupMax string = "max"

	// minCPUPartsCount is the minimum number of parts expected in cpu.max format.
	minCPUPartsCount int = 2

	// decimalBase is the base for decimal number parsing.
	decimalBase int = 10

	// bitSize64 is the bit size for 64-bit integers.
	bitSize64 int = 64

	// memoryUnlimitedThreshold is the threshold to detect unlimited memory (< 4 EiB).
	memoryUnlimitedThreshold uint64 = 1 << 62

	// cgroupPathParts is the number of parts in cgroup path format (hierarchy:controller:path).
	cgroupPathParts int = 3

	// cgroupPathIndex is the index of the cgroup path in split parts.
	cgroupPathIndex int = 2

	// cgroupHierarchyIndex is the index of the hierarchy ID in split parts.
	cgroupHierarchyIndex int = 0

	// cgroupControllerIndex is the index of the controller list in split parts.
	cgroupControllerIndex int = 1

	// defaultCgroupV1Capacity is the initial capacity for cgroup v1 controller map.
	defaultCgroupV1Capacity int = 8
)

// Cached cgroup paths - computed once per process lifetime.
var (
	cachedCgroupV2Path  func() string            = sync.OnceValue(getCgroupV2Path)
	cachedCgroupV1Paths func() map[string]string = sync.OnceValue(getCgroupV1Paths)
)

// collectCgroupLimits gathers cgroup v1/v2 limits.
//
// Params:
//   - limits: target resource limits struct to populate
func collectCgroupLimits(limits *model.ResourceLimits) {
	// Try cgroup v2 first (preferred on modern systems).
	if collectCgroupV2(limits) {
		// Successfully collected from cgroup v2.
		return
	}

	// Fall back to cgroup v1.
	collectCgroupV1(limits)
}

// collectCgroupV2 reads cgroup v2 limits.
// Uses cached cgroup path for efficiency.
//
// Params:
//   - limits: target resource limits struct to populate
//
// Returns:
//   - bool: true if cgroup v2 was available and read successfully
func collectCgroupV2(limits *model.ResourceLimits) bool {
	// Find cgroup path (cached).
	cgroupPath := cachedCgroupV2Path()
	// Check if cgroup v2 is available.
	if cgroupPath == "" {
		// Cgroup v2 not available.
		return false
	}

	// CPU limits (cpu.max).
	if content, err := os.ReadFile(filepath.Join(cgroupPath, "cpu.max")); err == nil {
		contentStr := string(content)
		parts := strings.Fields(contentStr)
		// Parse CPU quota/period if both parts are present.
		if len(parts) >= minCPUPartsCount {
			// Check if quota is not unlimited.
			if parts[0] != cgroupMax {
				// Parse CPU quota.
				if quota, err := strconv.ParseInt(parts[0], decimalBase, bitSize64); err == nil {
					// Parse CPU period.
					if period, err := strconv.ParseInt(parts[1], decimalBase, bitSize64); err == nil {
						limits.CPUQuotaRaw = quota
						limits.CPUPeriod = period
						limits.CPUQuota = float64(quota) / float64(period)
					}
				}
			}
		}
	}

	// CPU set (cpuset.cpus.effective or cpuset.cpus).
	// Try each cpuset file in order of preference.
	for _, name := range []string{"cpuset.cpus.effective", "cpuset.cpus"} {
		// Attempt to read cpuset file.
		if content, err := os.ReadFile(filepath.Join(cgroupPath, name)); err == nil {
			limits.CPUSet = strings.TrimSpace(string(content))
			// Check if we got a valid value.
			if limits.CPUSet != "" {
				// Found valid cpuset.
				break
			}
		}
	}

	// Memory max.
	if content, err := os.ReadFile(filepath.Join(cgroupPath, "memory.max")); err == nil {
		s := strings.TrimSpace(string(content))
		// Check if memory is not unlimited.
		if s != cgroupMax {
			// Parse memory limit.
			if val, err := strconv.ParseUint(s, decimalBase, bitSize64); err == nil {
				limits.MemoryMax = val
			}
		}
	}

	// Memory current.
	if content, err := os.ReadFile(filepath.Join(cgroupPath, "memory.current")); err == nil {
		// Parse memory current value.
		if val, err := strconv.ParseUint(strings.TrimSpace(string(content)), decimalBase, bitSize64); err == nil {
			limits.MemoryCurrent = val
		}
	}

	// PIDs max.
	if content, err := os.ReadFile(filepath.Join(cgroupPath, "pids.max")); err == nil {
		s := strings.TrimSpace(string(content))
		// Check if PIDs is not unlimited.
		if s != cgroupMax {
			// Parse PIDs limit.
			if val, err := strconv.ParseInt(s, decimalBase, bitSize64); err == nil {
				limits.PIDsMax = val
			}
		}
	}

	// PIDs current.
	if content, err := os.ReadFile(filepath.Join(cgroupPath, "pids.current")); err == nil {
		// Parse PIDs current value.
		if val, err := strconv.ParseInt(strings.TrimSpace(string(content)), decimalBase, bitSize64); err == nil {
			limits.PIDsCurrent = val
		}
	}

	// Successfully collected cgroup v2 limits.
	return true
}

// getCgroupV2Path returns the cgroup v2 path for the current process.
//
// Returns:
//   - string: cgroup v2 path or empty string if not available
func getCgroupV2Path() string {
	// Check if cgroup v2 is mounted.
	if _, err := os.Stat("/sys/fs/cgroup/cgroup.controllers"); err != nil {
		// Cgroup v2 not available.
		return ""
	}

	// Read our cgroup path.
	content, err := os.ReadFile("/proc/self/cgroup")
	// Handle read error.
	if err != nil {
		// Cannot read cgroup info.
		return ""
	}

	// Format: 0::/path
	// Parse each line to find cgroup v2 entry.
	for _, line := range strings.Split(string(content), "\n") {
		parts := strings.SplitN(line, ":", cgroupPathParts)
		// Check for valid cgroup v2 entry (hierarchy 0).
		if len(parts) == cgroupPathParts && parts[cgroupHierarchyIndex] == "0" {
			path := strings.TrimSpace(parts[cgroupPathIndex])
			// Handle root cgroup.
			if path == "" || path == "/" {
				// Use base cgroup path.
				return "/sys/fs/cgroup"
			}
			// Return full cgroup path.
			return filepath.Join("/sys/fs/cgroup", path)
		}
	}

	// Default to base cgroup path.
	return "/sys/fs/cgroup"
}

// collectCgroupV1 reads cgroup v1 limits.
// Uses cached cgroup paths for efficiency.
//
// Params:
//   - limits: target resource limits struct to populate
func collectCgroupV1(limits *model.ResourceLimits) {
	// Find cgroup paths (cached).
	cgroupPaths := cachedCgroupV1Paths()

	// CPU quota (cpu.cfs_quota_us / cpu.cfs_period_us).
	// Check if CPU controller path exists.
	if cpuPath, ok := cgroupPaths["cpu"]; ok {
		// Read CPU quota file.
		if content, err := os.ReadFile(filepath.Join(cpuPath, "cpu.cfs_quota_us")); err == nil {
			// Parse and validate quota value.
			if quota, err := strconv.ParseInt(strings.TrimSpace(string(content)), decimalBase, bitSize64); err == nil && quota > 0 {
				limits.CPUQuotaRaw = quota
				// Read CPU period file.
				if content, err := os.ReadFile(filepath.Join(cpuPath, "cpu.cfs_period_us")); err == nil {
					// Parse and validate period value.
					if period, err := strconv.ParseInt(strings.TrimSpace(string(content)), decimalBase, bitSize64); err == nil && period > 0 {
						limits.CPUPeriod = period
						limits.CPUQuota = float64(quota) / float64(period)
					}
				}
			}
		}
	}

	// CPU set.
	// Check if cpuset controller path exists.
	if cpusetPath, ok := cgroupPaths["cpuset"]; ok {
		// Read cpuset.cpus file.
		if content, err := os.ReadFile(filepath.Join(cpusetPath, "cpuset.cpus")); err == nil {
			limits.CPUSet = strings.TrimSpace(string(content))
		}
	}

	// Memory limit.
	// Check if memory controller path exists.
	if memPath, ok := cgroupPaths["memory"]; ok {
		// Read memory limit file.
		if content, err := os.ReadFile(filepath.Join(memPath, "memory.limit_in_bytes")); err == nil {
			// Parse memory limit value.
			if val, err := strconv.ParseUint(strings.TrimSpace(string(content)), decimalBase, bitSize64); err == nil {
				// Check for "unlimited" (very large value).
				if val < memoryUnlimitedThreshold {
					limits.MemoryMax = val
				}
			}
		}
		// Read memory usage file.
		if content, err := os.ReadFile(filepath.Join(memPath, "memory.usage_in_bytes")); err == nil {
			// Parse memory usage value.
			if val, err := strconv.ParseUint(strings.TrimSpace(string(content)), decimalBase, bitSize64); err == nil {
				limits.MemoryCurrent = val
			}
		}
	}

	// PIDs limit.
	// Check if pids controller path exists.
	if pidsPath, ok := cgroupPaths["pids"]; ok {
		// Read PIDs max file.
		if content, err := os.ReadFile(filepath.Join(pidsPath, "pids.max")); err == nil {
			s := strings.TrimSpace(string(content))
			// Check if not unlimited.
			if s != cgroupMax {
				// Parse PIDs max value.
				if val, err := strconv.ParseInt(s, decimalBase, bitSize64); err == nil {
					limits.PIDsMax = val
				}
			}
		}
		// Read PIDs current file.
		if content, err := os.ReadFile(filepath.Join(pidsPath, "pids.current")); err == nil {
			// Parse PIDs current value.
			if val, err := strconv.ParseInt(strings.TrimSpace(string(content)), decimalBase, bitSize64); err == nil {
				limits.PIDsCurrent = val
			}
		}
	}
}

// getCgroupV1Paths returns controller -> path mapping for cgroup v1.
//
// Returns:
//   - map[string]string: controller name to filesystem path mapping
func getCgroupV1Paths() map[string]string {
	// Pre-allocate for typical cgroup v1 controllers.
	paths := make(map[string]string, defaultCgroupV1Capacity)

	content, err := os.ReadFile("/proc/self/cgroup")
	// Handle read error.
	if err != nil {
		// Return empty map on error.
		return paths
	}

	// Format: hierarchy-ID:controller-list:cgroup-path
	// Parse each line of cgroup file.
	for _, line := range strings.Split(string(content), "\n") {
		parts := strings.SplitN(line, ":", cgroupPathParts)
		// Validate part count.
		if len(parts) != cgroupPathParts {
			// Skip malformed lines.
			continue
		}

		controllers := strings.Split(parts[cgroupControllerIndex], ",")
		cgroupPath := strings.TrimSpace(parts[cgroupPathIndex])

		// Process each controller.
		for _, controller := range controllers {
			// Skip empty controller names.
			if controller == "" {
				// Ignore empty entries.
				continue
			}

			// Build full path.
			fullPath := filepath.Join("/sys/fs/cgroup", controller, cgroupPath)
			// Verify path exists.
			if _, err := os.Stat(fullPath); err == nil {
				paths[controller] = fullPath
			}
		}
	}

	// Return controller to path mapping.
	return paths
}
