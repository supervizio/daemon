//go:build linux

// Package collector provides data collection for the TUI.
package collector

// cpuSample is a private helper struct for SystemCollector.
type cpuSample struct {
	user    uint64
	nice    uint64
	system  uint64
	idle    uint64
	iowait  uint64
	irq     uint64
	softirq uint64
	steal   uint64
}

// total returns the total CPU time.
//
// Returns:
//   - uint64: sum of all CPU time fields
func (s cpuSample) total() uint64 {
	// Sum all CPU time categories.
	return s.user + s.nice + s.system + s.idle + s.iowait + s.irq + s.softirq + s.steal
}

// busy returns the busy CPU time (non-idle).
//
// Returns:
//   - uint64: sum of non-idle CPU time fields
func (s cpuSample) busy() uint64 {
	// Sum non-idle CPU time categories.
	return s.user + s.nice + s.system + s.irq + s.softirq + s.steal
}
