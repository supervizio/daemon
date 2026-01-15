// Package lifecycle_test provides external tests for host.go.
package lifecycle_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/lifecycle"
)

func TestHostInfo_Uptime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupHost  func() lifecycle.HostInfo
		validateFn func(t *testing.T, uptime time.Duration)
	}{
		{
			name: "returns positive uptime for started host",
			setupHost: func() lifecycle.HostInfo {
				return lifecycle.HostInfo{
					StartTime: time.Now().Add(-5 * time.Minute),
				}
			},
			validateFn: func(t *testing.T, uptime time.Duration) {
				assert.True(t, uptime >= 5*time.Minute)
				assert.True(t, uptime < 6*time.Minute)
			},
		},
		{
			name: "returns zero for zero start time",
			setupHost: func() lifecycle.HostInfo {
				return lifecycle.HostInfo{}
			},
			validateFn: func(t *testing.T, uptime time.Duration) {
				assert.Equal(t, time.Duration(0), uptime)
			},
		},
		{
			name: "returns uptime for recently started host",
			setupHost: func() lifecycle.HostInfo {
				return lifecycle.HostInfo{
					StartTime: time.Now().Add(-100 * time.Millisecond),
				}
			},
			validateFn: func(t *testing.T, uptime time.Duration) {
				assert.True(t, uptime >= 100*time.Millisecond)
				assert.True(t, uptime < 1*time.Second)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			host := tt.setupHost()
			uptime := host.Uptime()
			tt.validateFn(t, uptime)
		})
	}
}

func TestHostInfo_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		host     lifecycle.HostInfo
		wantHost string
		wantOS   string
		wantArch string
	}{
		{
			name:     "empty host info",
			host:     lifecycle.HostInfo{},
			wantHost: "",
			wantOS:   "",
			wantArch: "",
		},
		{
			name: "full host info",
			host: lifecycle.HostInfo{
				Hostname:      "testhost",
				OS:            "linux",
				Arch:          "amd64",
				KernelVersion: "5.15.0",
				DaemonPID:     1234,
				DaemonVersion: "1.0.0",
			},
			wantHost: "testhost",
			wantOS:   "linux",
			wantArch: "amd64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantHost, tt.host.Hostname)
			assert.Equal(t, tt.wantOS, tt.host.OS)
			assert.Equal(t, tt.wantArch, tt.host.Arch)
		})
	}
}
