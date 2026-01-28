// Package supervisor provides the application service for orchestrating multiple services.
package supervisor

// ServiceSnapshotForTUI contains service info for TUI display.
// This struct uses basic types to avoid import cycles with TUI packages.
type ServiceSnapshotForTUI struct {
	Name            string
	StateInt        int
	StateName       string
	HealthInt       int  // Health status as int (maps to health.Status)
	HasHealthChecks bool // Whether health checks are configured
	PID             int
	UptimeSecs      int64
	CPUPercent      float64
	MemoryRSS       uint64
	RestartCount    int
	Ports           []int                    // Listening ports detected (TCP/UDP)
	Listeners       []ListenerSnapshotForTUI // Configured listeners with status
}
