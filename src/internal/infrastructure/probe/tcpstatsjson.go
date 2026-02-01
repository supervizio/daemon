//go:build cgo

package probe

// TcpStatsJSON contains aggregated TCP statistics.
// It tracks connection counts by state for monitoring and diagnostics.
type TcpStatsJSON struct {
	Established uint32 `json:"established"`
	SynSent     uint32 `json:"syn_sent"`
	SynRecv     uint32 `json:"syn_recv"`
	FinWait1    uint32 `json:"fin_wait1"`
	FinWait2    uint32 `json:"fin_wait2"`
	TimeWait    uint32 `json:"time_wait"`
	Close       uint32 `json:"close"`
	CloseWait   uint32 `json:"close_wait"`
	LastAck     uint32 `json:"last_ack"`
	Listen      uint32 `json:"listen"`
	Closing     uint32 `json:"closing"`
	Total       uint32 `json:"total"`
}
