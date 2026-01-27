// Package tui provides terminal user interface for superviz.io.
// Supports two modes: raw (static MOTD) and interactive (real-time TUI).
package tui

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/collector"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/terminal"
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

// Timing constants for TUI operations.
const (
	// defaultRefreshInterval is the default update frequency for interactive mode (10 FPS).
	defaultRefreshInterval time.Duration = 100 * time.Millisecond
	// startupPortWait is the delay before rendering raw mode to allow ports to bind.
	startupPortWait time.Duration = 500 * time.Millisecond
)

// Config holds TUI configuration.
// It specifies the operating mode, refresh interval, version, and output writer.
type Config struct {
	// Mode is the operating mode (raw or interactive).
	Mode Mode
	// RefreshInterval is the update frequency for interactive mode.
	// Minimum: 1 second.
	RefreshInterval time.Duration
	// Version is the daemon version string.
	Version string
	// Output is the writer for raw mode output.
	Output io.Writer
}

// DefaultConfig returns the default configuration.
//
// Params:
//   - version: the daemon version string.
//
// Returns:
//   - Config: the default configuration.
func DefaultConfig(version string) Config {
	// Return config with default values.
	return Config{
		Mode:            ModeRaw,
		RefreshInterval: defaultRefreshInterval,
		Version:         version,
		Output:          os.Stdout,
	}
}

// TUI is the main terminal user interface.
// It coordinates data collection, snapshot management, and rendering for both raw and interactive modes.
type TUI struct {
	config     Config
	collectors *collector.Collectors
	snapshot   *model.Snapshot

	// Data providers (set externally).
	serviceProvider ServiceProvider
	metricsProvider MetricsProvider
	healthProvider  HealthProvider
}

// ServiceProvider provides service information.
type ServiceProvider interface {
	// Services returns all service snapshots.
	Services() []model.ServiceSnapshot
}

// MetricsProvider provides system metrics.
type MetricsProvider interface {
	// SystemMetrics returns current system metrics.
	SystemMetrics() model.SystemMetrics
}

// HealthProvider provides health information.
type HealthProvider interface {
	// LogSummary returns log summary.
	LogSummary() model.LogSummary
}

// New creates a new TUI.
//
// Params:
//   - config: the TUI configuration.
//
// Returns:
//   - *TUI: the created TUI instance.
func New(config Config) *TUI {
	// Return new TUI instance with initialized collectors and snapshot.
	return &TUI{
		config:     config,
		collectors: collector.DefaultCollectors(config.Version),
		snapshot:   model.NewSnapshot(),
	}
}

// SetServiceProvider sets the service data provider.
//
// Params:
//   - p: the service provider.
func (t *TUI) SetServiceProvider(p ServiceProvider) {
	t.serviceProvider = p
}

// SetMetricsProvider sets the metrics data provider.
//
// Params:
//   - p: the metrics provider.
func (t *TUI) SetMetricsProvider(p MetricsProvider) {
	t.metricsProvider = p
}

// SetHealthProvider sets the health data provider.
//
// Params:
//   - p: the health provider.
func (t *TUI) SetHealthProvider(p HealthProvider) {
	t.healthProvider = p
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

	// Collect service data if provider is available.
	if t.serviceProvider != nil {
		t.snapshot.Services = t.serviceProvider.Services()
	}

	// Collect metrics if provider is available.
	if t.metricsProvider != nil {
		t.snapshot.System = t.metricsProvider.SystemMetrics()
	}

	// Collect health/logs if provider is available.
	if t.healthProvider != nil {
		t.snapshot.Logs = t.healthProvider.LogSummary()
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

// ShouldUseInteractive determines if interactive mode should be used.
//
// Returns:
//   - bool: true if interactive mode should be used.
func ShouldUseInteractive() bool {
	// Check if terminal supports interactive mode.
	return terminal.IsTTY()
}
