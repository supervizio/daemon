package config_test

import (
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/shared"
	"github.com/stretchr/testify/assert"
)

func TestTargetConfig(t *testing.T) {
	tests := []struct {
		name     string
		target   config.TargetConfig
		wantName string
		wantType string
		wantAddr string
	}{
		{
			name: "remote target",
			target: config.TargetConfig{
				Name:    "web-api",
				Type:    "remote",
				Address: "api.example.com:443",
			},
			wantName: "web-api",
			wantType: "remote",
			wantAddr: "api.example.com:443",
		},
		{
			name: "docker target",
			target: config.TargetConfig{
				Name:      "nginx-container",
				Type:      "docker",
				Container: "nginx_1",
			},
			wantName: "nginx-container",
			wantType: "docker",
		},
		{
			name: "kubernetes target",
			target: config.TargetConfig{
				Name:      "k8s-service",
				Type:      "kubernetes",
				Namespace: "production",
				Service:   "web-service",
			},
			wantName: "k8s-service",
			wantType: "kubernetes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantName, tt.target.Name)
			assert.Equal(t, tt.wantType, tt.target.Type)
			if tt.wantAddr != "" {
				assert.Equal(t, tt.wantAddr, tt.target.Address)
			}
		})
	}
}

func TestTargetConfig_WithIntervalAndTimeout(t *testing.T) {
	tests := []struct {
		name         string
		target       config.TargetConfig
		wantInterval time.Duration
		wantTimeout  time.Duration
	}{
		{
			name: "custom interval and timeout",
			target: config.TargetConfig{
				Name:     "test-target",
				Type:     "remote",
				Address:  "localhost:8080",
				Interval: shared.Duration(15 * time.Second),
				Timeout:  shared.Duration(3 * time.Second),
			},
			wantInterval: 15 * time.Second,
			wantTimeout:  3 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantInterval, tt.target.Interval.Duration())
			assert.Equal(t, tt.wantTimeout, tt.target.Timeout.Duration())
		})
	}
}

func TestTargetConfig_WithLabels(t *testing.T) {
	tests := []struct {
		name       string
		target     config.TargetConfig
		wantLabels map[string]string
	}{
		{
			name: "target with labels",
			target: config.TargetConfig{
				Name: "labeled-target",
				Type: "custom",
				Labels: map[string]string{
					"env":  "production",
					"team": "platform",
				},
			},
			wantLabels: map[string]string{
				"env":  "production",
				"team": "platform",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantLabels["env"], tt.target.Labels["env"])
			assert.Equal(t, tt.wantLabels["team"], tt.target.Labels["team"])
			assert.Len(t, tt.target.Labels, len(tt.wantLabels))
		})
	}
}

func TestTargetConfig_WithProbe(t *testing.T) {
	tests := []struct {
		name      string
		target    config.TargetConfig
		wantProbe config.ProbeConfig
	}{
		{
			name: "target with http probe",
			target: config.TargetConfig{
				Name:    "probed-target",
				Type:    "remote",
				Address: "api.example.com:443",
				Probe: config.ProbeConfig{
					Type:             "http",
					Path:             "/health",
					Interval:         shared.Duration(10 * time.Second),
					Timeout:          shared.Duration(2 * time.Second),
					SuccessThreshold: 1,
					FailureThreshold: 3,
				},
			},
			wantProbe: config.ProbeConfig{
				Type: "http",
				Path: "/health",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantProbe.Type, tt.target.Probe.Type)
			assert.Equal(t, tt.wantProbe.Path, tt.target.Probe.Path)
		})
	}
}

func TestTargetConfig_ZeroValue(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "zero value struct"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target config.TargetConfig

			assert.Empty(t, target.Name)
			assert.Empty(t, target.Type)
			assert.Empty(t, target.Address)
			assert.Nil(t, target.Labels)
		})
	}
}
