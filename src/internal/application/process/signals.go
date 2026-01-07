// Package process provides the application service for managing process lifecycle.
package process

import (
	"os"
	"syscall"
)

// signalHUP is the SIGHUP signal for reload operations.
var signalHUP os.Signal = syscall.SIGHUP
