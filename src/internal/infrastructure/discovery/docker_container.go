//go:build unix

// Package discovery provides infrastructure adapters for target discovery.
package discovery

// dockerContainer represents a container from Docker API response.
// This is an internal DTO for JSON unmarshaling from the Docker Engine API.
type dockerContainer struct {
	ID     string            `json:"Id"`
	Names  []string          `json:"Names"`
	State  string            `json:"State"`
	Status string            `json:"Status"`
	Labels map[string]string `json:"Labels"`
	Ports  []dockerPort      `json:"Ports"`
}
