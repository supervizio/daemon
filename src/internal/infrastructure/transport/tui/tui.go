// Package tui provides terminal user interface for superviz.io.
// Supports two modes: raw (static MOTD) and interactive (real-time TUI).
package tui

import (
	"context"
	"time"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/collector"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/terminal"
)

// Timing constants for TUI operations.
const (
	// defaultRefreshInterval is the default update frequency for interactive mode (10 FPS).
	defaultRefreshInterval time.Duration = 100 * time.Millisecond
	// startupPortWait is the delay before rendering raw mode to allow ports to bind.
	startupPortWait time.Duration = 500 * time.Millisecond
)

// Mode represents the TUI operating mode.
type Mode int

// Mode constants.
const (
	// ModeRaw outputs a static MOTD snapshot and exits.
	ModeRaw Mode = iota
	// ModeInteractive runs a real-time updating TUI.
	ModeInteractive
)

// TUI is the main terminal user interface.
// It coordinates data collection, snapshot management, and rendering for both raw and interactive modes.
type TUI struct {
	config     Config
	collectors *collector.Collectors
	snapshot   *model.Snapshot

	// Data providers (set externally).
	serviceLister ListServicesser
	metricser     Metricser
	summarizeer   Summarizeer
}

// ListServicesser provides service information.
type ListServicesser interface {
	// ListServices returns all service snapshots.
	ListServices() []model.ServiceSnapshot
}

// Metricser provides system metrics.
type Metricser interface {
	// Metrics returns current system metrics.
	Metrics() model.SystemMetrics
}

// Summarizeer provides health information.
type Summarizeer interface {
	// Summarize returns log summary.
	Summarize() model.LogSummary
}

// NewTUI creates a new TUI.
//
// Params:
//   - config: the TUI configuration.
//
// Returns:
//   - *TUI: the created TUI instance.
func NewTUI(config Config) *TUI {
	// Return new TUI instance with initialized collectors and snapshot.
	return &TUI{
		config:     config,
		collectors: collector.DefaultCollectors(config.Version),
		snapshot:   model.NewSnapshot(),
	}
}

// SetServiceLister sets the service data provider.
//
// Params:
//   - l: the service lister.
func (t *TUI) SetServiceLister(l ListServicesser) {
	t.serviceLister = l
}

// SetMetricser sets the metrics data provider.
//
// Params:
//   - m: the metricser.
func (t *TUI) SetMetricser(m Metricser) {
	t.metricser = m
}

// SetSummarizeer sets the health data provider.
//
// Params:
//   - s: the summarizeer.
func (t *TUI) SetSummarizeer(s Summarizeer) {
	t.summarizeer = s
}

// SetConfigPath sets the configuration file path for display.
//
// Params:
//   - path: the configuration file path.
func (t *TUI) SetConfigPath(path string) {
	t.collectors.SetConfigPath(path)
}

// Run executes the TUI.
//
// Params:
//   - ctx: the context for cancellation.
//
// Returns:
//   - error: nil on success, error on failure.
func (t *TUI) Run(ctx context.Context) error {
	// Select mode based on configuration.
	switch t.config.Mode {
	// Handle raw mode: output static snapshot.
	case ModeRaw:
		// Execute raw mode rendering.
		return t.runRaw()
	// Handle interactive mode: real-time TUI.
	case ModeInteractive:
		// Execute interactive mode rendering.
		return t.runInteractive(ctx)
	}
	// Fallback to raw mode for unknown modes.
	return t.runRaw()
}

// runRaw outputs a single snapshot.
//
// Returns:
//   - error: nil on success, error on failure.
func (t *TUI) runRaw() error {
	// Wait briefly for services to start and bind to ports.
	// This allows port status to be accurate in the startup banner.
	time.Sleep(startupPortWait)

	// Collect data.
	t.collectData()

	// Render.
	renderer := NewRawRenderer(t.config.Output)

	size := terminal.GetSize()
	layout := terminal.GetLayout(size)

	// Use compact layout for small terminals.
	if layout == terminal.LayoutCompact {
		// Render compact output for narrow terminals.
		return renderer.RenderCompact(t.snapshot)
	}
	// Render normal output for standard terminals.
	return renderer.Render(t.snapshot)
}

// runInteractive runs the real-time TUI.
//
// Params:
//   - ctx: the context for cancellation.
//
// Returns:
//   - error: nil on success, error on failure.
func (t *TUI) runInteractive(ctx context.Context) error {
	// Check if terminal is available.
	if !terminal.IsTTY() {
		// Fall back to raw mode.
		return t.runRaw()
	}

	// Run interactive mode.
	return t.runBubbleTea(ctx)
}

// collectData gathers all snapshot data.
func (t *TUI) collectData() {
	// Reset snapshot.
	t.snapshot = model.NewSnapshot()

	// Collect system data.
	_ = t.collectors.CollectAll(t.snapshot)

	// Collect service data if lister is available.
	if t.serviceLister != nil {
		t.snapshot.Services = t.serviceLister.ListServices()
	}

	// Collect metrics if metricser is available.
	if t.metricser != nil {
		t.snapshot.System = t.metricser.Metrics()
	}

	// Collect health/logs if summarizer is available.
	if t.summarizeer != nil {
		t.snapshot.Logs = t.summarizeer.Summarize()
	}
}

// Snapshot returns the current snapshot (for testing).
//
// Returns:
//   - *model.Snapshot: the current snapshot.
func (t *TUI) Snapshot() *model.Snapshot {
	// Return current snapshot reference.
	return t.snapshot
}

// SetSnapshot sets the snapshot (for testing).
//
// Params:
//   - snap: the snapshot to set.
func (t *TUI) SetSnapshot(snap *model.Snapshot) {
	// Set snapshot reference.
	t.snapshot = snap
}

// ShouldUseInteractive determines if interactive mode should be used.
//
// Returns:
//   - bool: true if interactive mode should be used.
func ShouldUseInteractive() bool {
	// Check if terminal supports interactive mode.
	return terminal.IsTTY()
}
