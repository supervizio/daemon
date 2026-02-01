//go:build cgo

package probe

// IOMetricsJSON contains I/O-related metrics for JSON output.
// It includes read/write operations, bytes, and optional pressure metrics.
type IOMetricsJSON struct {
	ReadOps    uint64          `json:"read_ops"`
	ReadBytes  uint64          `json:"read_bytes"`
	WriteOps   uint64          `json:"write_ops"`
	WriteBytes uint64          `json:"write_bytes"`
	Pressure   *IOPressureJSON `json:"pressure,omitempty"`
}
