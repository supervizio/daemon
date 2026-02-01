//go:build unix

// Package discovery provides infrastructure adapters for target discovery.
package discovery

// dockerPort represents a port mapping from Docker API response.
// This is an internal DTO for JSON unmarshaling from the Docker Engine API.
type dockerPort struct {
	IP          string `json:"IP"`
	PrivatePort int    `json:"PrivatePort"`
	PublicPort  int    `json:"PublicPort"`
	Type        string `json:"Type"`
}
