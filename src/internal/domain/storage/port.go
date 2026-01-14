// Package storage provides domain interfaces for metrics persistence.
package storage

import (
	"context"
	"time"

	"github.com/kodflow/daemon/internal/domain/probe"
)

// MetricsStore defines the interface for persisting and querying probe.
type MetricsStore interface {
	// WriteSystemCPU persists system CPU probe.
	WriteSystemCPU(ctx context.Context, m *probe.SystemCPU) error
	// WriteSystemMemory persists system memory probe.
	WriteSystemMemory(ctx context.Context, m *probe.SystemMemory) error
	// WriteProcessMetrics persists process probe.
	WriteProcessMetrics(ctx context.Context, m *probe.ProcessMetrics) error

	// GetSystemCPU retrieves system CPU metrics within the time range.
	GetSystemCPU(ctx context.Context, since, until time.Time) ([]probe.SystemCPU, error)
	// GetSystemMemory retrieves system memory metrics within the time range.
	GetSystemMemory(ctx context.Context, since, until time.Time) ([]probe.SystemMemory, error)
	// GetProcessMetrics retrieves process metrics for a service within the time range.
	GetProcessMetrics(ctx context.Context, serviceName string, since, until time.Time) ([]probe.ProcessMetrics, error)

	// GetLatestSystemCPU retrieves the most recent system CPU probe.
	GetLatestSystemCPU(ctx context.Context) (probe.SystemCPU, error)
	// GetLatestSystemMemory retrieves the most recent system memory probe.
	GetLatestSystemMemory(ctx context.Context) (probe.SystemMemory, error)
	// GetLatestProcessMetrics retrieves the most recent process metrics for a service.
	GetLatestProcessMetrics(ctx context.Context, serviceName string) (probe.ProcessMetrics, error)

	// Prune removes metrics older than the specified duration.
	// Returns the number of deleted entries.
	Prune(ctx context.Context, olderThan time.Duration) (int, error)

	// Close closes the store and releases resources.
	Close() error
}

// StoreConfig contains configuration for metrics storage.
type StoreConfig struct {
	// Path is the file path for the database.
	Path string
	// Retention is how long to keep probe.
	Retention time.Duration
	// PruneInterval is how often to run automatic pruning.
	PruneInterval time.Duration
}

// DefaultStoreConfig returns the default storage configuration.
func DefaultStoreConfig() StoreConfig {
	return StoreConfig{
		Path:          "/var/lib/supervizio/probe.db",
		Retention:     24 * time.Hour,
		PruneInterval: time.Hour,
	}
}
