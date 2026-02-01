//go:build cgo

package probe

// CPUPressureJSON contains PSI pressure metrics for CPU.
// It tracks CPU contention using Linux Pressure Stall Information.
type CPUPressureJSON struct {
	SomeAvg10   float64 `json:"some_avg10"`
	SomeAvg60   float64 `json:"some_avg60"`
	SomeAvg300  float64 `json:"some_avg300"`
	SomeTotalUs uint64  `json:"some_total_us"`
}
