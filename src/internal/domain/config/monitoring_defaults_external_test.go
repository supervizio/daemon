package config_test

import (
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/shared"
	"github.com/stretchr/testify/assert"
)

func TestMonitoringDefaults(t *testing.T) {
	tests := []struct {
		name             string
		interval         shared.Duration
		timeout          shared.Duration
		successThreshold int
		failureThreshold int
		wantInterval     time.Duration
		wantTimeout      time.Duration
	}{
		{
			name:             "standard defaults",
			interval:         shared.Duration(30 * time.Second),
			timeout:          shared.Duration(5 * time.Second),
			successThreshold: 1,
			failureThreshold: 3,
			wantInterval:     30 * time.Second,
			wantTimeout:      5 * time.Second,
		},
		{
			name:             "custom intervals",
			interval:         shared.Duration(1 * time.Minute),
			timeout:          shared.Duration(10 * time.Second),
			successThreshold: 2,
			failureThreshold: 5,
			wantInterval:     1 * time.Minute,
			wantTimeout:      10 * time.Second,
		},
		{
			name:             "zero values",
			interval:         shared.Duration(0),
			timeout:          shared.Duration(0),
			successThreshold: 0,
			failureThreshold: 0,
			wantInterval:     0,
			wantTimeout:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaults := config.MonitoringDefaults{
				Interval:         tt.interval,
				Timeout:          tt.timeout,
				SuccessThreshold: tt.successThreshold,
				FailureThreshold: tt.failureThreshold,
			}

			assert.Equal(t, tt.wantInterval, defaults.Interval.Duration())
			assert.Equal(t, tt.wantTimeout, defaults.Timeout.Duration())
			assert.Equal(t, tt.successThreshold, defaults.SuccessThreshold)
			assert.Equal(t, tt.failureThreshold, defaults.FailureThreshold)
		})
	}
}

func TestMonitoringDefaults_ZeroValue(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "zero value struct"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var defaults config.MonitoringDefaults

			assert.Zero(t, defaults.Interval.Duration())
			assert.Zero(t, defaults.Timeout.Duration())
			assert.Zero(t, defaults.SuccessThreshold)
			assert.Zero(t, defaults.FailureThreshold)
		})
	}
}
