//go:build cgo

// Package probe provides CGO bindings to the Rust probe library.
package probe

// Constants for probe package to avoid magic numbers.
const (
	// defaultTCPConnCapacity is the initial capacity for TCP connection slices.
	defaultTCPConnCapacity int = 256

	// defaultUDPConnCapacity is the initial capacity for UDP connection slices.
	defaultUDPConnCapacity int = 64

	// defaultUnixSockCapacity is the initial capacity for Unix socket slices.
	defaultUnixSockCapacity int = 64

	// defaultListenInfoCapacity is the initial capacity for listening port slices.
	defaultListenInfoCapacity int = 64

	// defaultJSONBufferSize is the initial capacity for JSON buffer (16KB).
	defaultJSONBufferSize int = 16 * 1024

	// maxTCPConnPoolSize is the maximum capacity before discarding from pool.
	maxTCPConnPoolSize int = 1024

	// maxGenericPoolSize is the maximum capacity for UDP/Unix/Listen pools.
	maxGenericPoolSize int = 512

	// maxJSONBufferPoolSize is the maximum buffer size before discarding (1MB).
	maxJSONBufferPoolSize int = 1024 * 1024

	// maxCStringCacheSize is the maximum number of cached C string conversions.
	maxCStringCacheSize int = 128

	// minListeningCapacity is the minimum capacity for listening port slices.
	minListeningCapacity int = 8

	// minEstablishedCapacity is the minimum capacity for established connection slices.
	minEstablishedCapacity int = 16

	// minProcessConnCapacity is the minimum capacity for process connection slices.
	minProcessConnCapacity int = 4

	// listeningPortPercentage is the divisor for estimating listening port capacity (5%).
	listeningPortPercentage int = 20

	// establishedConnPercentage is the divisor for estimating established connection capacity (70%).
	establishedConnPercentage int = 10

	// processConnPercentage is the divisor for estimating process connection capacity (10%).
	processConnPercentage int = 10
)
