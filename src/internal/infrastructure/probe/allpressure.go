//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

import (
	"github.com/kodflow/daemon/internal/domain/metrics"
)

// AllPressure contains all PSI pressure metrics.
// Only available on Linux with PSI support.
type AllPressure struct {
	// CPU contains CPU pressure metrics.
	CPU metrics.CPUPressure `dto:"out,api,pub" json:"cpu"`
	// Memory contains memory pressure metrics.
	Memory metrics.MemoryPressure `dto:"out,api,pub" json:"memory"`
	// IO contains I/O pressure metrics.
	IO metrics.IOPressure `dto:"out,api,pub" json:"io"`
}
