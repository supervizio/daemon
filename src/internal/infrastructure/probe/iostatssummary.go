//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

import (
	"time"
)

// IOStatsSummary contains aggregated I/O statistics.
// This provides a summary of all disk I/O operations.
type IOStatsSummary struct {
	// ReadOps is the total number of read operations.
	ReadOps uint64 `dto:"out,api,pub" json:"readOps"`
	// ReadBytes is the total number of bytes read.
	ReadBytes uint64 `dto:"out,api,pub" json:"readBytes"`
	// WriteOps is the total number of write operations.
	WriteOps uint64 `dto:"out,api,pub" json:"writeOps"`
	// WriteBytes is the total number of bytes written.
	WriteBytes uint64 `dto:"out,api,pub" json:"writeBytes"`
	// Timestamp is when the stats were collected.
	Timestamp time.Time `dto:"out,api,pub" json:"timestamp"`
}
