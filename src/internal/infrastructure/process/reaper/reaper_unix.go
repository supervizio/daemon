//go:build unix

// Package reaper provides platform-specific implementations of kernel interfaces.
// This file implements zombie process reaping for Unix systems.
package reaper

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Reaper implements ZombieReaper for Unix systems.
// It handles automatic reaping of zombie child processes using SIGCHLD signals.
type Reaper struct {
	// mu protects the running state from concurrent access.
	mu sync.Mutex
	// running indicates whether the reaper loop is currently active.
	running bool
	// stopCh is used to signal the reaper loop to stop.
	stopCh chan struct{}
	// doneCh is closed when the reaper loop has fully stopped.
	doneCh chan struct{}
}

// NewReaper returns a Reaper for orphan zombie cleanup.
//
// Returns:
//   - *Reaper: initialized reaper ready to start
func NewReaper() *Reaper {
	return &Reaper{stopCh: make(chan struct{}), doneCh: make(chan struct{})}
}

// New returns a Reaper for orphan zombie cleanup.
//
// Returns:
//   - *Reaper: initialized reaper ready to start
func New() *Reaper {
	return &Reaper{stopCh: make(chan struct{}), doneCh: make(chan struct{})}
}

// Start launches the SIGCHLD-driven reaping goroutine.
func (r *Reaper) Start() {
	r.mu.Lock()
	// Already running; prevent duplicate goroutines.
	if r.running {
		r.mu.Unlock()
		return
	}
	r.running = true
	// Fresh channels for this run cycle.
	r.stopCh = make(chan struct{})
	r.doneCh = make(chan struct{})
	r.mu.Unlock()
	go r.reapLoop()
}

// Stop terminates the reaping loop and waits for completion.
func (r *Reaper) Stop() {
	r.mu.Lock()
	// Not running; nothing to stop.
	if !r.running {
		r.mu.Unlock()
		return
	}
	r.running = false
	r.mu.Unlock()
	close(r.stopCh)
	// Wait for final reap cycle to complete.
	<-r.doneCh
}

// reapLoop waits for SIGCHLD and reaps zombies until stopped.
func (r *Reaper) reapLoop() {
	defer close(r.doneCh)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGCHLD)
	defer signal.Stop(sigCh)
	for {
		select {
		// Shutdown requested; do final cleanup.
		case <-r.stopCh:
			r.reapAll()
			return
		// Child terminated; reap all pending zombies.
		case <-sigCh:
			r.reapAll()
		}
	}
}

// reapAll uses non-blocking waitpid to collect all terminated children.
func (r *Reaper) reapAll() {
	// Loop until no more zombies remain.
	for {
		pid, err := syscall.Wait4(-1, nil, syscall.WNOHANG, nil)
		// No more children or error (ECHILD when no children exist).
		if err != nil || pid <= 0 {
			break
		}
	}
}

// ReapOnce collects zombies once and returns the count.
//
// Returns:
//   - int: number of zombie processes reaped in this call
func (r *Reaper) ReapOnce() int {
	count := 0
	// Loop until no more zombies remain.
	for {
		pid, err := syscall.Wait4(-1, nil, syscall.WNOHANG, nil)
		// No more children or error (ECHILD when no children exist).
		if err != nil || pid <= 0 {
			break
		}
		count++
	}
	return count
}

// IsPID1 checks if running as init process (needed for subreaper role).
//
// Returns:
//   - bool: true when running as PID 1 (container init or system init)
func (r *Reaper) IsPID1() bool { return os.Getpid() == 1 }
