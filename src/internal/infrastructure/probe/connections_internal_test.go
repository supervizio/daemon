//go:build cgo

package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotFoundPIDConstant(t *testing.T) {
	tests := []struct {
		name string
		want int32
	}{
		{
			name: "NotFoundPIDIsNegativeOne",
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, notFoundPID)
		})
	}
}

// TestConvertTCPConnections tests the convertTCPConnections function exists.
func TestConvertTCPConnections(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
	}{
		{name: "function exists and compiles"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// convertTCPConnections requires C.TcpConnectionList, tested via CollectTCP integration.
			assert.NotNil(t, convertTCPConnections)
		})
	}
}

// TestConvertUDPConnections tests the convertUDPConnections function exists.
func TestConvertUDPConnections(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
	}{
		{name: "function exists and compiles"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// convertUDPConnections requires C.UdpConnectionList, tested via CollectUDP integration.
			assert.NotNil(t, convertUDPConnections)
		})
	}
}
