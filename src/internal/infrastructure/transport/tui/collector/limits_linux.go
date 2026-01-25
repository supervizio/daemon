//go:build linux

package collector

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

// cgroupMax is the value used in cgroups to indicate unlimited.
const cgroupMax = "max"

// collectCgroupLimits gathers cgroup v1/v2 limits.
func collectCgroupLimits(limits *model.ResourceLimits) {
	// Try cgroup v2 first.
	if collectCgroupV2(limits) {
		return
	}

	// Fall back to cgroup v1.
	collectCgroupV1(limits)
}

// collectCgroupV2 reads cgroup v2 limits.
func collectCgroupV2(limits *model.ResourceLimits) bool {
	// Find cgroup path.
	cgroupPath := getCgroupV2Path()
	if cgroupPath == "" {
		return false
	}

	// CPU limits (cpu.max).
	if content, err := os.ReadFile(filepath.Join(cgroupPath, "cpu.max")); err == nil {
		parts := strings.Fields(string(content))
		if len(parts) >= 2 {
			if parts[0] != cgroupMax {
				if quota, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
					if period, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
						limits.CPUQuotaRaw = quota
						limits.CPUPeriod = period
						limits.CPUQuota = float64(quota) / float64(period)
					}
				}
			}
		}
	}

	// CPU set (cpuset.cpus.effective or cpuset.cpus).
	for _, name := range []string{"cpuset.cpus.effective", "cpuset.cpus"} {
		if content, err := os.ReadFile(filepath.Join(cgroupPath, name)); err == nil {
			limits.CPUSet = strings.TrimSpace(string(content))
			if limits.CPUSet != "" {
				break
			}
		}
	}

	// Memory max.
	if content, err := os.ReadFile(filepath.Join(cgroupPath, "memory.max")); err == nil {
		s := strings.TrimSpace(string(content))
		if s != cgroupMax {
			if val, err := strconv.ParseUint(s, 10, 64); err == nil {
				limits.MemoryMax = val
			}
		}
	}

	// Memory current.
	if content, err := os.ReadFile(filepath.Join(cgroupPath, "memory.current")); err == nil {
		if val, err := strconv.ParseUint(strings.TrimSpace(string(content)), 10, 64); err == nil {
			limits.MemoryCurrent = val
		}
	}

	// PIDs max.
	if content, err := os.ReadFile(filepath.Join(cgroupPath, "pids.max")); err == nil {
		s := strings.TrimSpace(string(content))
		if s != cgroupMax {
			if val, err := strconv.ParseInt(s, 10, 64); err == nil {
				limits.PIDsMax = val
			}
		}
	}

	// PIDs current.
	if content, err := os.ReadFile(filepath.Join(cgroupPath, "pids.current")); err == nil {
		if val, err := strconv.ParseInt(strings.TrimSpace(string(content)), 10, 64); err == nil {
			limits.PIDsCurrent = val
		}
	}

	return true
}

// getCgroupV2Path returns the cgroup v2 path for the current process.
func getCgroupV2Path() string {
	// Check if cgroup v2 is mounted.
	if _, err := os.Stat("/sys/fs/cgroup/cgroup.controllers"); err != nil {
		return ""
	}

	// Read our cgroup path.
	content, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		return ""
	}

	// Format: 0::/path
	for _, line := range strings.Split(string(content), "\n") {
		parts := strings.SplitN(line, ":", 3)
		if len(parts) == 3 && parts[0] == "0" {
			path := strings.TrimSpace(parts[2])
			if path == "" || path == "/" {
				return "/sys/fs/cgroup"
			}
			return filepath.Join("/sys/fs/cgroup", path)
		}
	}

	return "/sys/fs/cgroup"
}

// collectCgroupV1 reads cgroup v1 limits.
func collectCgroupV1(limits *model.ResourceLimits) {
	// Find cgroup paths.
	cgroupPaths := getCgroupV1Paths()

	// CPU quota (cpu.cfs_quota_us / cpu.cfs_period_us).
	if cpuPath, ok := cgroupPaths["cpu"]; ok {
		if content, err := os.ReadFile(filepath.Join(cpuPath, "cpu.cfs_quota_us")); err == nil {
			if quota, err := strconv.ParseInt(strings.TrimSpace(string(content)), 10, 64); err == nil && quota > 0 {
				limits.CPUQuotaRaw = quota
				if content, err := os.ReadFile(filepath.Join(cpuPath, "cpu.cfs_period_us")); err == nil {
					if period, err := strconv.ParseInt(strings.TrimSpace(string(content)), 10, 64); err == nil && period > 0 {
						limits.CPUPeriod = period
						limits.CPUQuota = float64(quota) / float64(period)
					}
				}
			}
		}
	}

	// CPU set.
	if cpusetPath, ok := cgroupPaths["cpuset"]; ok {
		if content, err := os.ReadFile(filepath.Join(cpusetPath, "cpuset.cpus")); err == nil {
			limits.CPUSet = strings.TrimSpace(string(content))
		}
	}

	// Memory limit.
	if memPath, ok := cgroupPaths["memory"]; ok {
		if content, err := os.ReadFile(filepath.Join(memPath, "memory.limit_in_bytes")); err == nil {
			if val, err := strconv.ParseUint(strings.TrimSpace(string(content)), 10, 64); err == nil {
				// Check for "unlimited" (very large value).
				if val < 1<<62 {
					limits.MemoryMax = val
				}
			}
		}
		if content, err := os.ReadFile(filepath.Join(memPath, "memory.usage_in_bytes")); err == nil {
			if val, err := strconv.ParseUint(strings.TrimSpace(string(content)), 10, 64); err == nil {
				limits.MemoryCurrent = val
			}
		}
	}

	// PIDs limit.
	if pidsPath, ok := cgroupPaths["pids"]; ok {
		if content, err := os.ReadFile(filepath.Join(pidsPath, "pids.max")); err == nil {
			s := strings.TrimSpace(string(content))
			if s != cgroupMax {
				if val, err := strconv.ParseInt(s, 10, 64); err == nil {
					limits.PIDsMax = val
				}
			}
		}
		if content, err := os.ReadFile(filepath.Join(pidsPath, "pids.current")); err == nil {
			if val, err := strconv.ParseInt(strings.TrimSpace(string(content)), 10, 64); err == nil {
				limits.PIDsCurrent = val
			}
		}
	}
}

// getCgroupV1Paths returns controller -> path mapping for cgroup v1.
func getCgroupV1Paths() map[string]string {
	paths := make(map[string]string)

	content, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		return paths
	}

	// Format: hierarchy-ID:controller-list:cgroup-path
	for _, line := range strings.Split(string(content), "\n") {
		parts := strings.SplitN(line, ":", 3)
		if len(parts) != 3 {
			continue
		}

		controllers := strings.Split(parts[1], ",")
		cgroupPath := strings.TrimSpace(parts[2])

		for _, controller := range controllers {
			if controller == "" {
				continue
			}

			// Build full path.
			fullPath := filepath.Join("/sys/fs/cgroup", controller, cgroupPath)
			if _, err := os.Stat(fullPath); err == nil {
				paths[controller] = fullPath
			}
		}
	}

	return paths
}
