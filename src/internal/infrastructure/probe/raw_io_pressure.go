// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// RawIOPressure holds raw I/O pressure data.
// Used for Go-only testing of pressure metrics building.
type RawIOPressure struct {
	// SomeAvg10 represents some-avg10 metric.
	SomeAvg10 float64
	// SomeAvg60 represents some-avg60 metric.
	SomeAvg60 float64
	// SomeAvg300 represents some-avg300 metric.
	SomeAvg300 float64
	// SomeTotalUs represents some-total in microseconds.
	SomeTotalUs uint64
	// FullAvg10 represents full-avg10 metric.
	FullAvg10 float64
	// FullAvg60 represents full-avg60 metric.
	FullAvg60 float64
	// FullAvg300 represents full-avg300 metric.
	FullAvg300 float64
	// FullTotalUs represents full-total in microseconds.
	FullTotalUs uint64
}
