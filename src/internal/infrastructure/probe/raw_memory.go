// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// RawMemoryData holds raw memory data for Go-only building.
// This struct mirrors C data without requiring CGO for testing.
type RawMemoryData struct {
	// TotalBytes represents total memory in bytes.
	TotalBytes uint64
	// AvailableBytes represents available memory in bytes.
	AvailableBytes uint64
	// UsedBytes represents used memory in bytes.
	UsedBytes uint64
	// CachedBytes represents cached memory in bytes.
	CachedBytes uint64
	// BuffersBytes represents buffer memory in bytes.
	BuffersBytes uint64
	// SwapTotalBytes represents total swap in bytes.
	SwapTotalBytes uint64
	// SwapUsedBytes represents used swap in bytes.
	SwapUsedBytes uint64
}
