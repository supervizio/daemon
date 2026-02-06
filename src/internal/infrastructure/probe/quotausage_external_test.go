//go:build cgo

package probe_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
)

func TestNewQuotaUsage_External(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ReturnsZeroInitializedUsage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usage := probe.NewQuotaUsage()
			assert.NotNil(t, usage)
			assert.Equal(t, uint64(0), usage.MemoryBytes)
			assert.Equal(t, uint64(0), usage.MemoryLimitBytes)
		})
	}
}

func TestQuotaUsage_MemoryUsagePercent_External(t *testing.T) {
	tests := []struct {
		name  string
		usage *probe.QuotaUsage
		want  float64
	}{
		{
			name:  "ZeroWhenNoLimit",
			usage: &probe.QuotaUsage{},
			want:  0,
		},
		{
			name: "CorrectPercentage",
			usage: &probe.QuotaUsage{
				MemoryBytes:      50 * 1024 * 1024,
				MemoryLimitBytes: 100 * 1024 * 1024,
			},
			want: 50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.usage.MemoryUsagePercent()
			assert.Equal(t, tt.want, got)
		})
	}
}
