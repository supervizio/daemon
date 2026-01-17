// Package main provides the entry point for the superviz.io process supervisor.
// All dependency injection and application logic is handled by the bootstrap package.
package main

import (
	"os"

	"github.com/kodflow/daemon/internal/bootstrap"
)

// main is the entry point for the superviz.io process supervisor.
func main() {
	os.Exit(bootstrap.Run())
}
