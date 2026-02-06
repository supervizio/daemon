//go:build cgo

package probe_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
)

func TestNewQuotaLimits_External(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ReturnsZeroInitializedLimits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limits := probe.NewQuotaLimits()
			assert.NotNil(t, limits)
			assert.Equal(t, uint64(0), limits.CPUQuotaUS)
			assert.Equal(t, uint64(0), limits.MemoryLimitBytes)
			assert.Equal(t, uint32(0), limits.Flags)
		})
	}
}

func TestQuotaLimits_HasCPULimit_External(t *testing.T) {
	tests := []struct {
		name   string
		limits *probe.QuotaLimits
		want   bool
	}{
		{
			name:   "NoLimitWhenZero",
			limits: &probe.QuotaLimits{},
			want:   false,
		},
		{
			name: "NoLimitWhenFlagNotSet",
			limits: &probe.QuotaLimits{
				CPUQuotaUS:  100000,
				CPUPeriodUS: 100000,
			},
			want: false,
		},
		{
			name: "HasLimitWhenSet",
			limits: &probe.QuotaLimits{
				CPUQuotaUS:  100000,
				CPUPeriodUS: 100000,
				Flags:       probe.QuotaFlagCPU,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.limits.HasCPULimit()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestQuotaLimits_HasMemoryLimit_External(t *testing.T) {
	tests := []struct {
		name   string
		limits *probe.QuotaLimits
		want   bool
	}{
		{
			name:   "NoLimitWhenZero",
			limits: &probe.QuotaLimits{},
			want:   false,
		},
		{
			name: "NoLimitWhenFlagNotSet",
			limits: &probe.QuotaLimits{
				MemoryLimitBytes: 1024 * 1024 * 1024,
			},
			want: false,
		},
		{
			name: "HasLimitWhenSet",
			limits: &probe.QuotaLimits{
				MemoryLimitBytes: 1024 * 1024 * 1024,
				Flags:            probe.QuotaFlagMemory,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.limits.HasMemoryLimit()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestQuotaLimits_CPULimitPercent_External(t *testing.T) {
	tests := []struct {
		name   string
		limits *probe.QuotaLimits
		want   float64
	}{
		{
			name:   "ZeroWhenNoLimit",
			limits: &probe.QuotaLimits{},
			want:   0,
		},
		{
			name: "CorrectPercentage",
			limits: &probe.QuotaLimits{
				CPUQuotaUS:  50000,
				CPUPeriodUS: 100000,
				Flags:       probe.QuotaFlagCPU,
			},
			want: 50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.limits.CPULimitPercent()
			assert.Equal(t, tt.want, got)
		})
	}
}
