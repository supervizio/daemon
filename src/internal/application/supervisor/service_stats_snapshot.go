// Package supervisor provides the application service for orchestrating multiple services.
// It manages the lifecycle of services including start, stop, restart, and reload operations.
package supervisor

// ServiceStatsSnapshot is an immutable copy of ServiceStats counters.
// Used for passing stats to callbacks without race conditions.
type ServiceStatsSnapshot struct {
	StartCount   int `dto:"out,priv,pub" json:"startCount"`
	StopCount    int `dto:"out,priv,pub" json:"stopCount"`
	FailCount    int `dto:"out,priv,pub" json:"failCount"`
	RestartCount int `dto:"out,priv,pub" json:"restartCount"`
}
