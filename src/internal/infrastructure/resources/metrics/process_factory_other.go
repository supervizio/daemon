//go:build !linux

// Package metrics provides platform detection and factory for metrics collectors.
package metrics

import (
	"context"

	appmetrics "github.com/kodflow/daemon/internal/application/metrics"
	domainmetrics "github.com/kodflow/daemon/internal/domain/metrics"
)

// stubProcessCollector is a no-op process collector for unsupported platforms.
type stubProcessCollector struct{}

// CollectCPU returns zero CPU metrics.
func (s *stubProcessCollector) CollectCPU(_ context.Context, _ int) (domainmetrics.ProcessCPU, error) {
	// Return empty metrics for unsupported platforms.
	return domainmetrics.ProcessCPU{}, nil
}

// CollectMemory returns zero memory metrics.
func (s *stubProcessCollector) CollectMemory(_ context.Context, _ int) (domainmetrics.ProcessMemory, error) {
	// Return empty metrics for unsupported platforms.
	return domainmetrics.ProcessMemory{}, nil
}

// NewProcessCollector creates a stub ProcessCollector for non-Linux platforms.
// It returns zero values for all metrics.
//
// Returns:
//   - appmetrics.Collector: stub collector that returns zero values
func NewProcessCollector() appmetrics.Collector {
	// Return stub collector for non-Linux platforms.
	return &stubProcessCollector{}
}
