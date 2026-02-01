// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// RawCPUPressure holds raw CPU pressure data.
// Used for Go-only testing of pressure metrics building.
type RawCPUPressure struct {
	// SomeAvg10 represents some-avg10 metric.
	SomeAvg10 float64
	// SomeAvg60 represents some-avg60 metric.
	SomeAvg60 float64
	// SomeAvg300 represents some-avg300 metric.
	SomeAvg300 float64
	// SomeTotalUs represents some-total in microseconds.
	SomeTotalUs uint64
}
