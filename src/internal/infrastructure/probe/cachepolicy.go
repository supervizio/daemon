//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for unified
// cross-platform system metrics and resource quota management.
package probe

// CachePolicy represents a cache policy preset.
// It defines the TTL behavior for cached metrics.
type CachePolicy uint32

// Cache policy presets for different collection frequencies.
const (
	// CachePolicyDefault provides balanced TTLs for general use.
	CachePolicyDefault CachePolicy = 0
	// CachePolicyHighFreq provides shorter TTLs for frequent collection.
	CachePolicyHighFreq CachePolicy = 1
	// CachePolicyLowFreq provides longer TTLs for infrequent collection.
	CachePolicyLowFreq CachePolicy = 2
	// CachePolicyNoCache disables caching (TTL=0).
	CachePolicyNoCache CachePolicy = 3
)
