//go:build cgo

package probe_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
)

func TestNewTcpStats(t *testing.T) {
	tests := []struct {
		name string
		want *probe.TcpStats
	}{
		{
			name: "ReturnsZeroInitializedStats",
			want: &probe.TcpStats{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := probe.NewTcpStats()
			assert.NotNil(t, got)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTcpStats_Total(t *testing.T) {
	tests := []struct {
		name  string
		stats *probe.TcpStats
		want  uint32
	}{
		{
			name:  "ZeroStats",
			stats: &probe.TcpStats{},
			want:  0,
		},
		{
			name: "SingleState",
			stats: &probe.TcpStats{
				Established: 5,
			},
			want: 5,
		},
		{
			name: "AllStates",
			stats: &probe.TcpStats{
				Established: 1,
				SynSent:     2,
				SynRecv:     3,
				FinWait1:    4,
				FinWait2:    5,
				TimeWait:    6,
				Close:       7,
				CloseWait:   8,
				LastAck:     9,
				Listen:      10,
				Closing:     11,
			},
			want: 66,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stats.Total()
			assert.Equal(t, tt.want, got)
		})
	}
}
