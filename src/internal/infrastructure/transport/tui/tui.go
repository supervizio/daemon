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

// Config holds TUI configuration.
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
func DefaultConfig(version string) Config {
	return Config{
		Mode:            ModeRaw,
		RefreshInterval: 100 * time.Millisecond, // 10 FPS for smooth updates.
		Version:         version,
		Output:          os.Stdout,
	}
}

// TUI is the main terminal user interface.
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
func New(config Config) *TUI {
	return &TUI{
		config:     config,
		collectors: collector.DefaultCollectors(config.Version),
		snapshot:   model.NewSnapshot(),
	}
}

// SetServiceProvider sets the service data provider.
func (t *TUI) SetServiceProvider(p ServiceProvider) {
	t.serviceProvider = p
}

// SetMetricsProvider sets the metrics data provider.
func (t *TUI) SetMetricsProvider(p MetricsProvider) {
	t.metricsProvider = p
}

// SetHealthProvider sets the health data provider.
func (t *TUI) SetHealthProvider(p HealthProvider) {
	t.healthProvider = p
}

// SetConfigPath sets the configuration file path for display.
func (t *TUI) SetConfigPath(path string) {
	t.collectors.SetConfigPath(path)
}

// Run executes the TUI.
func (t *TUI) Run(ctx context.Context) error {
	switch t.config.Mode {
	case ModeRaw:
		return t.runRaw()
	case ModeInteractive:
		return t.runInteractive(ctx)
	}
	// Fallback to raw mode for unknown modes.
	return t.runRaw()
}

// runRaw outputs a single snapshot.
func (t *TUI) runRaw() error {
	// Wait briefly for services to start and bind to ports.
	// This allows port status to be accurate in the startup banner.
	time.Sleep(500 * time.Millisecond)

	// Collect data.
	t.collectData()

	// Render.
	renderer := NewRawRenderer(t.config.Output)

	size := terminal.GetSize()
	layout := terminal.GetLayout(size)

	if layout == terminal.LayoutCompact {
		return renderer.RenderCompact(t.snapshot)
	}
	return renderer.Render(t.snapshot)
}

// runInteractive runs the real-time TUI.
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

	// Collect service data.
	if t.serviceProvider != nil {
		t.snapshot.Services = t.serviceProvider.Services()
	}

	// Collect metrics.
	if t.metricsProvider != nil {
		t.snapshot.System = t.metricsProvider.SystemMetrics()
	}

	// Collect health/logs.
	if t.healthProvider != nil {
		t.snapshot.Logs = t.healthProvider.LogSummary()
	}
}

// Snapshot returns the current snapshot (for testing).
func (t *TUI) Snapshot() *model.Snapshot {
	return t.snapshot
}

// ShouldUseInteractive determines if interactive mode should be used.
func ShouldUseInteractive() bool {
	return terminal.IsTTY()
}
