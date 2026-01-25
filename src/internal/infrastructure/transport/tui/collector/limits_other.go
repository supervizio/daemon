//go:build !linux

package collector

import (
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

// collectCgroupLimits is a no-op on non-Linux platforms.
func collectCgroupLimits(limits *model.ResourceLimits) {
	// Cgroups are Linux-specific.
	// On other platforms, limits are not collected.
	_ = limits
}
