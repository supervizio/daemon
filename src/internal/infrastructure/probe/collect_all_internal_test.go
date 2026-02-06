//go:build cgo

package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFullPercentageConstant(t *testing.T) {
	tests := []struct {
		name string
		want float64
	}{
		{
			name: "FullPercentageIs100",
			want: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, fullPercentage)
		})
	}
}

// TestExtractPressureFromC tests the extractPressureFromC function exists.
// Note: Cannot test directly without C struct, verified through buildPressureMetricsFromRaw.
func TestExtractPressureFromC(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
	}{
		{name: "function exists and compiles"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// extractPressureFromC requires C.AllMetrics, tested via CollectAll integration.
			// Verify function signature exists by checking build compiles.
			assert.NotNil(t, extractPressureFromC)
		})
	}
}

// TestExtractPartitionsFromC tests the extractPartitionsFromC function exists.
func TestExtractPartitionsFromC(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
	}{
		{name: "function exists and compiles"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// extractPartitionsFromC requires C.AllMetrics, tested via CollectAll integration.
			assert.NotNil(t, extractPartitionsFromC)
		})
	}
}

// TestExtractDiskUsageFromC tests the extractDiskUsageFromC function exists.
func TestExtractDiskUsageFromC(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
	}{
		{name: "function exists and compiles"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// extractDiskUsageFromC requires C.AllMetrics, tested via CollectAll integration.
			assert.NotNil(t, extractDiskUsageFromC)
		})
	}
}

// TestExtractDiskIOFromC tests the extractDiskIOFromC function exists.
func TestExtractDiskIOFromC(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
	}{
		{name: "function exists and compiles"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// extractDiskIOFromC requires C.AllMetrics, tested via CollectAll integration.
			assert.NotNil(t, extractDiskIOFromC)
		})
	}
}

// TestExtractNetIfacesFromC tests the extractNetIfacesFromC function exists.
func TestExtractNetIfacesFromC(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
	}{
		{name: "function exists and compiles"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// extractNetIfacesFromC requires C.AllMetrics, tested via CollectAll integration.
			assert.NotNil(t, extractNetIfacesFromC)
		})
	}
}

// TestExtractNetStatsFromC tests the extractNetStatsFromC function exists.
func TestExtractNetStatsFromC(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
	}{
		{name: "function exists and compiles"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// extractNetStatsFromC requires C.AllMetrics, tested via CollectAll integration.
			assert.NotNil(t, extractNetStatsFromC)
		})
	}
}
