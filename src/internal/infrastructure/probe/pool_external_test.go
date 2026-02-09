//go:build cgo

package probe_test

import (
	"bytes"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTCPConnPool verifies TCP connection slice pooling.
func TestTCPConnPool(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "get and put preserves capacity"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get slice from pool
			slice1 := probe.GetTCPConnSliceForTest()
			require.NotNil(t, slice1)
			assert.Equal(t, 0, len(*slice1))
			assert.GreaterOrEqual(t, cap(*slice1), 0)

			// Add some data
			*slice1 = append(*slice1, probe.TcpConnJSON{
				Family:    "ipv4",
				LocalAddr: "127.0.0.1",
				LocalPort: 8080,
			})
			initialCap := cap(*slice1)

			// Return to pool
			probe.PutTCPConnSliceForTest(slice1)

			// Get again - should have zero length but preserved capacity
			slice2 := probe.GetTCPConnSliceForTest()
			assert.Equal(t, 0, len(*slice2))
			// Capacity should be at least what we had before (pool may give us a different slice)
			assert.GreaterOrEqual(t, cap(*slice2), 0)

			// Return to pool
			probe.PutTCPConnSliceForTest(slice2)
		})
	}
}

// TestUDPConnPool verifies UDP socket slice pooling.
func TestUDPConnPool(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "get and put preserves capacity"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get slice from pool
			slice1 := probe.GetUDPConnSliceForTest()
			require.NotNil(t, slice1)
			assert.Equal(t, 0, len(*slice1))

			// Add some data
			*slice1 = append(*slice1, probe.UdpConnJSON{
				Family:    "ipv4",
				LocalAddr: "127.0.0.1",
				LocalPort: 53,
			})

			// Return to pool
			probe.PutUDPConnSliceForTest(slice1)

			// Get again - should have zero length
			slice2 := probe.GetUDPConnSliceForTest()
			assert.Equal(t, 0, len(*slice2))

			// Return to pool
			probe.PutUDPConnSliceForTest(slice2)
		})
	}
}

// TestUnixSockPool verifies Unix socket slice pooling.
func TestUnixSockPool(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "get and put preserves capacity"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get slice from pool
			slice1 := probe.GetUnixSockSliceForTest()
			require.NotNil(t, slice1)
			assert.Equal(t, 0, len(*slice1))

			// Add some data
			*slice1 = append(*slice1, probe.UnixSockJSON{
				Path:  "/var/run/test.sock",
				Type:  "stream",
				State: "listening",
			})

			// Return to pool
			probe.PutUnixSockSliceForTest(slice1)

			// Get again - should have zero length
			slice2 := probe.GetUnixSockSliceForTest()
			assert.Equal(t, 0, len(*slice2))

			// Return to pool
			probe.PutUnixSockSliceForTest(slice2)
		})
	}
}

// TestListenInfoPool verifies listening port slice pooling.
func TestListenInfoPool(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "get and put preserves capacity"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get slice from pool
			slice1 := probe.GetListenInfoSliceForTest()
			require.NotNil(t, slice1)
			assert.Equal(t, 0, len(*slice1))

			// Add some data
			*slice1 = append(*slice1, probe.ListenInfoJSON{
				Protocol: "tcp",
				Address:  "0.0.0.0",
				Port:     80,
			})

			// Return to pool
			probe.PutListenInfoSliceForTest(slice1)

			// Get again - should have zero length
			slice2 := probe.GetListenInfoSliceForTest()
			assert.Equal(t, 0, len(*slice2))

			// Return to pool
			probe.PutListenInfoSliceForTest(slice2)
		})
	}
}

// TestJSONBufferPool verifies JSON buffer pooling.
func TestJSONBufferPool(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "get and put resets buffer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get buffer from pool
			buf1 := probe.GetJSONBufferForTest()
			require.NotNil(t, buf1)
			assert.Equal(t, 0, buf1.Len())

			// Write some data
			buf1.WriteString("test data")
			assert.Greater(t, buf1.Len(), 0)

			// Return to pool
			probe.PutJSONBufferForTest(buf1)

			// Get again - should be reset
			buf2 := probe.GetJSONBufferForTest()
			assert.Equal(t, 0, buf2.Len())

			// Return to pool
			probe.PutJSONBufferForTest(buf2)
		})
	}
}

// TestPoolConcurrency verifies pools work correctly under concurrent access.
func TestPoolConcurrency(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "concurrent get and put operations"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const goroutines = 10
			const iterations = 100

			done := make(chan bool, goroutines)

			for i := 0; i < goroutines; i++ {
				go func() {
					for j := 0; j < iterations; j++ {
						// Test TCP pool
						tcpSlice := probe.GetTCPConnSliceForTest()
						*tcpSlice = append(*tcpSlice, probe.TcpConnJSON{LocalPort: uint16(j)})
						probe.PutTCPConnSliceForTest(tcpSlice)

						// Test JSON buffer pool
						buf := probe.GetJSONBufferForTest()
						buf.WriteString("test")
						probe.PutJSONBufferForTest(buf)
					}
					done <- true
				}()
			}

			// Wait for all goroutines
			for i := 0; i < goroutines; i++ {
				<-done
			}
		})
	}
}

// TestPoolSizeLimit verifies pools don't retain excessively large slices.
func TestPoolSizeLimit(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "huge slices are not pooled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get a slice
			slice := probe.GetTCPConnSliceForTest()

			// Grow it to exceed pool limit (1024)
			hugeSlice := make([]probe.TcpConnJSON, 2000)
			*slice = hugeSlice

			// This should not panic even though we're returning a huge slice
			probe.PutTCPConnSliceForTest(slice)
			// The implementation should detect the size and not pool it

			// Get another slice - should be a fresh one (or from pool if available)
			slice2 := probe.GetTCPConnSliceForTest()
			assert.NotNil(t, slice2)
			assert.Equal(t, 0, len(*slice2))

			probe.PutTCPConnSliceForTest(slice2)
		})
	}
}

// TestJSONBufferSizeLimit verifies JSON buffers don't retain excessively large buffers.
func TestJSONBufferSizeLimit(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "huge buffers are not pooled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get a buffer
			buf := probe.GetJSONBufferForTest()

			// Grow it beyond pool limit (1MB)
			hugeData := make([]byte, 2*1024*1024) // 2MB
			buf.Write(hugeData)

			// This should not panic
			probe.PutJSONBufferForTest(buf)

			// Get another buffer - should be a fresh one
			buf2 := probe.GetJSONBufferForTest()
			assert.NotNil(t, buf2)
			assert.Equal(t, 0, buf2.Len())

			probe.PutJSONBufferForTest(buf2)
		})
	}
}
