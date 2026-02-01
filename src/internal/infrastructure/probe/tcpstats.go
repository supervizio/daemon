//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// TcpStats contains aggregated TCP connection statistics.
// It tracks connection counts by state for system monitoring.
type TcpStats struct {
	Established uint32
	SynSent     uint32
	SynRecv     uint32
	FinWait1    uint32
	FinWait2    uint32
	TimeWait    uint32
	Close       uint32
	CloseWait   uint32
	LastAck     uint32
	Listen      uint32
	Closing     uint32
}

// NewTcpStats creates a new empty TcpStats instance.
//
// Returns:
//   - *TcpStats: a new zero-initialized TCP statistics instance
func NewTcpStats() *TcpStats {
	// Return zero-initialized struct.
	return &TcpStats{}
}

// Total returns the total number of TCP connections.
//
// Returns:
//   - uint32: sum of all connection state counts
func (s *TcpStats) Total() uint32 {
	// Sum all connection states
	return s.Established + s.SynSent + s.SynRecv + s.FinWait1 + s.FinWait2 +
		s.TimeWait + s.Close + s.CloseWait + s.LastAck + s.Listen + s.Closing
}
