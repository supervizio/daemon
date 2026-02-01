//go:build linux

// Package discovery provides infrastructure adapters for target discovery.
// This file defines the listeningPort struct for port scan discovery.
package discovery

// listeningPort represents a port in LISTEN state from /proc/net/tcp.
type listeningPort struct {
	Protocol  string `dto:"out,priv,pub" json:"protocol"`  // "tcp" or "udp"
	LocalAddr string `dto:"out,priv,pub" json:"localAddr"` // IP address
	LocalPort int    `dto:"out,priv,pub" json:"localPort"` // Port number
	State     string `dto:"out,priv,pub" json:"state"`     // Socket state (hex)
}
