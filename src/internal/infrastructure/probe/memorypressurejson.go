//go:build cgo

package probe

// MemoryPressureJSON contains PSI pressure metrics for memory.
// It tracks memory contention using Linux Pressure Stall Information.
type MemoryPressureJSON struct {
	SomeAvg10   float64 `json:"some_avg10"`
	SomeAvg60   float64 `json:"some_avg60"`
	SomeAvg300  float64 `json:"some_avg300"`
	SomeTotalUs uint64  `json:"some_total_us"`
	FullAvg10   float64 `json:"full_avg10"`
	FullAvg60   float64 `json:"full_avg60"`
	FullAvg300  float64 `json:"full_avg300"`
	FullTotalUs uint64  `json:"full_total_us"`
}
