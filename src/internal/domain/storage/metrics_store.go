// Package storage provides domain interfaces for metrics persistence.
package storage

import (
	"context"
	"time"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

const (
	// DefaultRetentionHours is the default number of hours to retain metrics.
	DefaultRetentionHours int = 24
)

// MetricsWriter defines the interface for writing metrics to storage.
// Use this interface when a component only needs to persist metrics.
type MetricsWriter interface {
	// WriteSystemCPU persists system CPU metrics.
	WriteSystemCPU(ctx context.Context, m *metrics.SystemCPU) error
	// WriteSystemMemory persists system memory metrics.
	WriteSystemMemory(ctx context.Context, m *metrics.SystemMemory) error
	// WriteProcessMetrics persists process metrics.
	WriteProcessMetrics(ctx context.Context, m *metrics.ProcessMetrics) error
}

// MetricsReader defines the interface for reading metrics from storage.
// Use this interface when a component only needs to query metrics.
type MetricsReader interface {
	// GetSystemCPU retrieves system CPU metrics within the time range.
	GetSystemCPU(ctx context.Context, since, until time.Time) ([]metrics.SystemCPU, error)
	// GetSystemMemory retrieves system memory metrics within the time range.
	GetSystemMemory(ctx context.Context, since, until time.Time) ([]metrics.SystemMemory, error)
	// GetProcessMetrics retrieves process metrics for a service within the time range.
	GetProcessMetrics(ctx context.Context, serviceName string, since, until time.Time) ([]metrics.ProcessMetrics, error)
	// GetLatestSystemCPU retrieves the most recent system CPU metrics.
	GetLatestSystemCPU(ctx context.Context) (metrics.SystemCPU, error)
	// GetLatestSystemMemory retrieves the most recent system memory metrics.
	GetLatestSystemMemory(ctx context.Context) (metrics.SystemMemory, error)
	// GetLatestProcessMetrics retrieves the most recent process metrics for a service.
	GetLatestProcessMetrics(ctx context.Context, serviceName string) (metrics.ProcessMetrics, error)
}

// MetricsMaintainer defines the interface for storage maintenance operations.
// Use this interface when a component only needs maintenance capabilities.
type MetricsMaintainer interface {
	// Prune removes metrics older than the specified duration.
	// Returns the number of deleted entries.
	Prune(ctx context.Context, olderThan time.Duration) (int, error)
	// Close closes the store and releases resources.
	Close() error
}

// MetricsStore defines the complete interface for metrics persistence.
// It composes MetricsWriter, MetricsReader, and MetricsMaintainer.
// Use the smaller interfaces when a component needs only a subset of operations.
type MetricsStore interface {
	MetricsWriter
	MetricsReader
	MetricsMaintainer
}

// StoreConfig contains configuration for metrics storage.
// It defines the persistence location, retention policy, and automatic
// pruning behavior for time-series metrics data.
type StoreConfig struct {
	// Path is the file path for the database.
	Path string
	// Retention is how long to keep metrics.
	Retention time.Duration
	// PruneInterval is how often to run automatic pruning.
	PruneInterval time.Duration
}

// DefaultStoreConfig returns the default storage configuration.
//
// Returns:
//   - StoreConfig: configuration with default database path, 24-hour retention, and hourly pruning
func DefaultStoreConfig() StoreConfig {
	// Returns standard defaults for production use.
	return StoreConfig{
		Path:          "/var/lib/supervizio/metrics.db",
		Retention:     time.Duration(DefaultRetentionHours) * time.Hour,
		PruneInterval: time.Hour,
	}
}
