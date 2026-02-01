// Package monitoring provides the application service for external target monitoring.
package monitoring

import (
	"time"

	"github.com/kodflow/daemon/internal/domain/target"
)

// DiscoveryModeConfig configures automatic target discovery.
// When enabled, discoverers run periodically to find new targets.
type DiscoveryModeConfig struct {
	// Enabled indicates if auto-discovery is active.
	Enabled bool

	// Interval is how often to re-run discovery.
	Interval time.Duration

	// Discoverers are the discovery adapters to use.
	Discoverers []target.Discoverer

	// Watchers are the real-time watcher adapters to use.
	Watchers []target.Watcher
}

// NewDiscoveryModeConfig creates a new DiscoveryModeConfig with defaults.
//
// Returns:
//   - DiscoveryModeConfig: a new instance with defaults applied.
func NewDiscoveryModeConfig() DiscoveryModeConfig {
	// construct config with default values
	return DiscoveryModeConfig{
		Enabled:     false,
		Interval:    DefaultDiscoveryInterval,
		Discoverers: nil,
		Watchers:    nil,
	}
}
