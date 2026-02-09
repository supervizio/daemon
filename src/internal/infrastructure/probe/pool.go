//go:build cgo

// Package probe provides CGO bindings to the Rust probe library.
package probe

import (
	"bytes"
	"sync"
)

// Connection slice pools to reduce allocations.
// These pools reuse slices for frequently allocated connection data.
var (
	// tcpConnPool pools TCP connection slices.
	// Pre-allocates capacity for 256 connections (typical server load).
	tcpConnPool = &sync.Pool{
		New: func() any {
			slice := make([]TcpConnJSON, 0, 256)
			return &slice
		},
	}

	// udpConnPool pools UDP socket slices.
	// Pre-allocates capacity for 64 sockets (typical usage).
	udpConnPool = &sync.Pool{
		New: func() any {
			slice := make([]UdpConnJSON, 0, 64)
			return &slice
		},
	}

	// unixSockPool pools Unix socket slices.
	// Pre-allocates capacity for 64 sockets (typical usage).
	unixSockPool = &sync.Pool{
		New: func() any {
			slice := make([]UnixSockJSON, 0, 64)
			return &slice
		},
	}

	// listenInfoPool pools listening port slices.
	// Pre-allocates capacity for 64 ports (typical server).
	listenInfoPool = &sync.Pool{
		New: func() any {
			slice := make([]ListenInfoJSON, 0, 64)
			return &slice
		},
	}

	// jsonBufferPool pools bytes.Buffer for JSON encoding.
	// Pre-allocates 16KB capacity (typical metrics JSON size).
	jsonBufferPool = &sync.Pool{
		New: func() any {
			// 16KB initial capacity based on typical metrics size
			return bytes.NewBuffer(make([]byte, 0, 16*1024))
		},
	}
)

// getTCPConnSlice retrieves a TCP connection slice from the pool.
// Caller must call putTCPConnSlice when done.
//
// Returns:
//   - *[]TcpConnJSON: pooled slice with zero length
func getTCPConnSlice() *[]TcpConnJSON {
	slice := tcpConnPool.Get().(*[]TcpConnJSON)
	// Reset length to zero while preserving capacity
	*slice = (*slice)[:0]
	return slice
}

// putTCPConnSlice returns a TCP connection slice to the pool.
// The slice should not be used after calling this function.
//
// Params:
//   - slice: the slice to return to the pool
func putTCPConnSlice(slice *[]TcpConnJSON) {
	// Only pool if capacity is reasonable (avoid retaining huge slices)
	if cap(*slice) <= 1024 {
		tcpConnPool.Put(slice)
	}
}

// getUDPConnSlice retrieves a UDP socket slice from the pool.
// Caller must call putUDPConnSlice when done.
//
// Returns:
//   - *[]UdpConnJSON: pooled slice with zero length
func getUDPConnSlice() *[]UdpConnJSON {
	slice := udpConnPool.Get().(*[]UdpConnJSON)
	// Reset length to zero while preserving capacity
	*slice = (*slice)[:0]
	return slice
}

// putUDPConnSlice returns a UDP socket slice to the pool.
// The slice should not be used after calling this function.
//
// Params:
//   - slice: the slice to return to the pool
func putUDPConnSlice(slice *[]UdpConnJSON) {
	// Only pool if capacity is reasonable (avoid retaining huge slices)
	if cap(*slice) <= 512 {
		udpConnPool.Put(slice)
	}
}

// getUnixSockSlice retrieves a Unix socket slice from the pool.
// Caller must call putUnixSockSlice when done.
//
// Returns:
//   - *[]UnixSockJSON: pooled slice with zero length
func getUnixSockSlice() *[]UnixSockJSON {
	slice := unixSockPool.Get().(*[]UnixSockJSON)
	// Reset length to zero while preserving capacity
	*slice = (*slice)[:0]
	return slice
}

// putUnixSockSlice returns a Unix socket slice to the pool.
// The slice should not be used after calling this function.
//
// Params:
//   - slice: the slice to return to the pool
func putUnixSockSlice(slice *[]UnixSockJSON) {
	// Only pool if capacity is reasonable (avoid retaining huge slices)
	if cap(*slice) <= 512 {
		unixSockPool.Put(slice)
	}
}

// getListenInfoSlice retrieves a listening port slice from the pool.
// Caller must call putListenInfoSlice when done.
//
// Returns:
//   - *[]ListenInfoJSON: pooled slice with zero length
func getListenInfoSlice() *[]ListenInfoJSON {
	slice := listenInfoPool.Get().(*[]ListenInfoJSON)
	// Reset length to zero while preserving capacity
	*slice = (*slice)[:0]
	return slice
}

// putListenInfoSlice returns a listening port slice to the pool.
// The slice should not be used after calling this function.
//
// Params:
//   - slice: the slice to return to the pool
func putListenInfoSlice(slice *[]ListenInfoJSON) {
	// Only pool if capacity is reasonable (avoid retaining huge slices)
	if cap(*slice) <= 512 {
		listenInfoPool.Put(slice)
	}
}

// getJSONBuffer retrieves a bytes.Buffer from the pool.
// Caller must call putJSONBuffer when done.
//
// Returns:
//   - *bytes.Buffer: pooled buffer, reset and ready for use
func getJSONBuffer() *bytes.Buffer {
	buf := jsonBufferPool.Get().(*bytes.Buffer)
	// Reset buffer to zero length while preserving capacity
	buf.Reset()
	return buf
}

// putJSONBuffer returns a bytes.Buffer to the pool.
// The buffer should not be used after calling this function.
//
// Params:
//   - buf: the buffer to return to the pool
func putJSONBuffer(buf *bytes.Buffer) {
	// Only pool if capacity is reasonable (avoid retaining huge buffers)
	// 1MB max to prevent memory bloat
	if buf.Cap() <= 1024*1024 {
		jsonBufferPool.Put(buf)
	}
}

// Test helpers - exported functions for testing pool behavior.
// These should only be used in tests.

// GetTCPConnSliceForTest retrieves a TCP connection slice from the pool (test helper).
func GetTCPConnSliceForTest() *[]TcpConnJSON {
	return getTCPConnSlice()
}

// PutTCPConnSliceForTest returns a TCP connection slice to the pool (test helper).
func PutTCPConnSliceForTest(slice *[]TcpConnJSON) {
	putTCPConnSlice(slice)
}

// GetUDPConnSliceForTest retrieves a UDP socket slice from the pool (test helper).
func GetUDPConnSliceForTest() *[]UdpConnJSON {
	return getUDPConnSlice()
}

// PutUDPConnSliceForTest returns a UDP socket slice to the pool (test helper).
func PutUDPConnSliceForTest(slice *[]UdpConnJSON) {
	putUDPConnSlice(slice)
}

// GetUnixSockSliceForTest retrieves a Unix socket slice from the pool (test helper).
func GetUnixSockSliceForTest() *[]UnixSockJSON {
	return getUnixSockSlice()
}

// PutUnixSockSliceForTest returns a Unix socket slice to the pool (test helper).
func PutUnixSockSliceForTest(slice *[]UnixSockJSON) {
	putUnixSockSlice(slice)
}

// GetListenInfoSliceForTest retrieves a listening port slice from the pool (test helper).
func GetListenInfoSliceForTest() *[]ListenInfoJSON {
	return getListenInfoSlice()
}

// PutListenInfoSliceForTest returns a listening port slice to the pool (test helper).
func PutListenInfoSliceForTest(slice *[]ListenInfoJSON) {
	putListenInfoSlice(slice)
}

// GetJSONBufferForTest retrieves a bytes.Buffer from the pool (test helper).
func GetJSONBufferForTest() *bytes.Buffer {
	return getJSONBuffer()
}

// PutJSONBufferForTest returns a bytes.Buffer to the pool (test helper).
func PutJSONBufferForTest(buf *bytes.Buffer) {
	putJSONBuffer(buf)
}
