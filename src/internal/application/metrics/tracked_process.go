// Package metrics provides application services for process metrics tracking.
package metrics

import (
	"time"

	domainmetrics "github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/domain/process"
)

// trackedProcess holds the state for a single tracked process.
type trackedProcess struct {
	serviceName  string
	pid          int
	state        process.State
	healthy      bool
	startTime    time.Time
	restartCount int
	lastError    string
	lastMetrics  domainmetrics.ProcessMetrics
}
