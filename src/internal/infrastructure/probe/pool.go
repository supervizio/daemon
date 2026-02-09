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
	tcpConnPool *sync.Pool = &sync.Pool{
		New: func() any {
			slice := make([]TcpConnJSON, 0, defaultTCPConnCapacity)
			return &slice
		},
	}

	// udpConnPool pools UDP socket slices.
	// Pre-allocates capacity for 64 sockets (typical usage).
	udpConnPool *sync.Pool = &sync.Pool{
		New: func() any {
			slice := make([]UdpConnJSON, 0, defaultUDPConnCapacity)
			return &slice
		},
	}

	// unixSockPool pools Unix socket slices.
	// Pre-allocates capacity for 64 sockets (typical usage).
	unixSockPool *sync.Pool = &sync.Pool{
		New: func() any {
			slice := make([]UnixSockJSON, 0, defaultUnixSockCapacity)
			return &slice
		},
	}

	// listenInfoPool pools listening port slices.
	// Pre-allocates capacity for 64 ports (typical server).
	listenInfoPool *sync.Pool = &sync.Pool{
		New: func() any {
			slice := make([]ListenInfoJSON, 0, defaultListenInfoCapacity)
			return &slice
		},
	}

	// jsonBufferPool pools bytes.Buffer for JSON encoding.
	// Pre-allocates 16KB capacity (typical metrics JSON size).
	jsonBufferPool *sync.Pool = &sync.Pool{
		New: func() any {
			// 16KB initial capacity based on typical metrics size
			return bytes.NewBuffer(make([]byte, 0, defaultJSONBufferSize))
		},
	}
)

// getTCPConnSlice retrieves a TCP connection slice from the pool.
// Caller must call putTCPConnSlice when done.
//
// Returns:
//   - *[]TcpConnJSON: pooled slice with zero length
func getTCPConnSlice() *[]TcpConnJSON {
	// get slice from pool with type assertion.
	slice, ok := tcpConnPool.Get().(*[]TcpConnJSON)
	// fallback to new slice if type assertion fails.
	if !ok {
		temp := make([]TcpConnJSON, 0, defaultTCPConnCapacity)
		// return freshly allocated slice.
		return &temp
	}
	// reset length to zero while preserving capacity
	*slice = (*slice)[:0]
	// return pooled slice ready for use
	return slice
}

// putTCPConnSlice returns a TCP connection slice to the pool.
// The slice should not be used after calling this function.
//
// Params:
//   - slice: the slice to return to the pool
func putTCPConnSlice(slice *[]TcpConnJSON) {
	// only pool if capacity is reasonable (avoid retaining huge slices)
	if cap(*slice) <= maxTCPConnPoolSize {
		tcpConnPool.Put(slice)
	}
}

// getUDPConnSlice retrieves a UDP socket slice from the pool.
// Caller must call putUDPConnSlice when done.
//
// Returns:
//   - *[]UdpConnJSON: pooled slice with zero length
func getUDPConnSlice() *[]UdpConnJSON {
	// get slice from pool with type assertion.
	slice, ok := udpConnPool.Get().(*[]UdpConnJSON)
	// fallback to new slice if type assertion fails.
	if !ok {
		temp := make([]UdpConnJSON, 0, defaultUDPConnCapacity)
		// return freshly allocated slice.
		return &temp
	}
	// reset length to zero while preserving capacity
	*slice = (*slice)[:0]
	// return pooled slice ready for use
	return slice
}

// putUDPConnSlice returns a UDP socket slice to the pool.
// The slice should not be used after calling this function.
//
// Params:
//   - slice: the slice to return to the pool
func putUDPConnSlice(slice *[]UdpConnJSON) {
	// only pool if capacity is reasonable (avoid retaining huge slices)
	if cap(*slice) <= maxGenericPoolSize {
		udpConnPool.Put(slice)
	}
}

// getUnixSockSlice retrieves a Unix socket slice from the pool.
// Caller must call putUnixSockSlice when done.
//
// Returns:
//   - *[]UnixSockJSON: pooled slice with zero length
func getUnixSockSlice() *[]UnixSockJSON {
	// get slice from pool with type assertion.
	slice, ok := unixSockPool.Get().(*[]UnixSockJSON)
	// fallback to new slice if type assertion fails.
	if !ok {
		temp := make([]UnixSockJSON, 0, defaultUnixSockCapacity)
		// return freshly allocated slice.
		return &temp
	}
	// reset length to zero while preserving capacity
	*slice = (*slice)[:0]
	// return pooled slice ready for use
	return slice
}

// putUnixSockSlice returns a Unix socket slice to the pool.
// The slice should not be used after calling this function.
//
// Params:
//   - slice: the slice to return to the pool
func putUnixSockSlice(slice *[]UnixSockJSON) {
	// only pool if capacity is reasonable (avoid retaining huge slices)
	if cap(*slice) <= maxGenericPoolSize {
		unixSockPool.Put(slice)
	}
}

// getListenInfoSlice retrieves a listening port slice from the pool.
// Caller must call putListenInfoSlice when done.
//
// Returns:
//   - *[]ListenInfoJSON: pooled slice with zero length
func getListenInfoSlice() *[]ListenInfoJSON {
	// get slice from pool with type assertion.
	slice, ok := listenInfoPool.Get().(*[]ListenInfoJSON)
	// fallback to new slice if type assertion fails.
	if !ok {
		temp := make([]ListenInfoJSON, 0, defaultListenInfoCapacity)
		// return freshly allocated slice.
		return &temp
	}
	// reset length to zero while preserving capacity
	*slice = (*slice)[:0]
	// return pooled slice ready for use
	return slice
}

// putListenInfoSlice returns a listening port slice to the pool.
// The slice should not be used after calling this function.
//
// Params:
//   - slice: the slice to return to the pool
func putListenInfoSlice(slice *[]ListenInfoJSON) {
	// only pool if capacity is reasonable (avoid retaining huge slices)
	if cap(*slice) <= maxGenericPoolSize {
		listenInfoPool.Put(slice)
	}
}

// getJSONBuffer retrieves a bytes.Buffer from the pool.
// Caller must call putJSONBuffer when done.
//
// Returns:
//   - *bytes.Buffer: pooled buffer, reset and ready for use
func getJSONBuffer() *bytes.Buffer {
	// get buffer from pool with type assertion.
	buf, ok := jsonBufferPool.Get().(*bytes.Buffer)
	// fallback to new buffer if type assertion fails.
	if !ok {
		// return freshly allocated buffer.
		return bytes.NewBuffer(make([]byte, 0, defaultJSONBufferSize))
	}
	// reset buffer to zero length while preserving capacity
	buf.Reset()
	// return pooled buffer ready for use
	return buf
}

// putJSONBuffer returns a bytes.Buffer to the pool.
// The buffer should not be used after calling this function.
//
// Params:
//   - buf: the buffer to return to the pool
func putJSONBuffer(buf *bytes.Buffer) {
	// only pool if capacity is reasonable (avoid retaining huge buffers)
	// 1MB max to prevent memory bloat
	if buf.Cap() <= maxJSONBufferPoolSize {
		jsonBufferPool.Put(buf)
	}
}

// Test helpers - exported functions for testing pool behavior.
// These should only be used in tests.

// GetTCPConnSliceForTest retrieves a TCP connection slice from the pool (test helper).
//
// Returns:
//   - *[]TcpConnJSON: pooled slice for testing
func GetTCPConnSliceForTest() *[]TcpConnJSON {
	// delegate to internal function
	return getTCPConnSlice()
}

// PutTCPConnSliceForTest returns a TCP connection slice to the pool (test helper).
//
// Params:
//   - slice: the slice to return to the pool
func PutTCPConnSliceForTest(slice *[]TcpConnJSON) {
	putTCPConnSlice(slice)
}

// GetUDPConnSliceForTest retrieves a UDP socket slice from the pool (test helper).
//
// Returns:
//   - *[]UdpConnJSON: pooled slice for testing
func GetUDPConnSliceForTest() *[]UdpConnJSON {
	// delegate to internal function
	return getUDPConnSlice()
}

// PutUDPConnSliceForTest returns a UDP socket slice to the pool (test helper).
//
// Params:
//   - slice: the slice to return to the pool
func PutUDPConnSliceForTest(slice *[]UdpConnJSON) {
	putUDPConnSlice(slice)
}

// GetUnixSockSliceForTest retrieves a Unix socket slice from the pool (test helper).
//
// Returns:
//   - *[]UnixSockJSON: pooled slice for testing
func GetUnixSockSliceForTest() *[]UnixSockJSON {
	// delegate to internal function
	return getUnixSockSlice()
}

// PutUnixSockSliceForTest returns a Unix socket slice to the pool (test helper).
//
// Params:
//   - slice: the slice to return to the pool
func PutUnixSockSliceForTest(slice *[]UnixSockJSON) {
	putUnixSockSlice(slice)
}

// GetListenInfoSliceForTest retrieves a listening port slice from the pool (test helper).
//
// Returns:
//   - *[]ListenInfoJSON: pooled slice for testing
func GetListenInfoSliceForTest() *[]ListenInfoJSON {
	// delegate to internal function
	return getListenInfoSlice()
}

// PutListenInfoSliceForTest returns a listening port slice to the pool (test helper).
//
// Params:
//   - slice: the slice to return to the pool
func PutListenInfoSliceForTest(slice *[]ListenInfoJSON) {
	putListenInfoSlice(slice)
}

// GetJSONBufferForTest retrieves a bytes.Buffer from the pool (test helper).
//
// Returns:
//   - *bytes.Buffer: pooled buffer for testing
func GetJSONBufferForTest() *bytes.Buffer {
	// delegate to internal function
	return getJSONBuffer()
}

// PutJSONBufferForTest returns a bytes.Buffer to the pool (test helper).
//
// Params:
//   - buf: the buffer to return to the pool
func PutJSONBufferForTest(buf *bytes.Buffer) {
	putJSONBuffer(buf)
}
