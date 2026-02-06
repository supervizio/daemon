//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// IOPressureJSON contains PSI pressure metrics for I/O.
// It tracks I/O contention using Linux Pressure Stall Information.
type IOPressureJSON struct {
	SomeAvg10   float64 `json:"some_avg10"`
	SomeAvg60   float64 `json:"some_avg60"`
	SomeAvg300  float64 `json:"some_avg300"`
	SomeTotalUs uint64  `json:"some_total_us"`
	FullAvg10   float64 `json:"full_avg10"`
	FullAvg60   float64 `json:"full_avg60"`
	FullAvg300  float64 `json:"full_avg300"`
	FullTotalUs uint64  `json:"full_total_us"`
}
