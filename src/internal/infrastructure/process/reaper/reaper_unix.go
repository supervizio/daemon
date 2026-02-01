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
	// return new reaper instance.
	return &Reaper{stopCh: make(chan struct{}), doneCh: make(chan struct{})}
}

// New returns a Reaper for orphan zombie cleanup.
//
// Returns:
//   - *Reaper: initialized reaper ready to start
func New() *Reaper {
	// return new reaper instance.
	return &Reaper{stopCh: make(chan struct{}), doneCh: make(chan struct{})}
}

// Start launches the SIGCHLD-driven reaping goroutine.
func (r *Reaper) Start() {
	r.mu.Lock()
	// Already running; prevent duplicate goroutines.
	if r.running {
		// release lock and skip duplicate start.
		r.mu.Unlock()
		// exit early to avoid duplicate reaper loops.
		return
	}
	r.running = true
	// Fresh channels for this run cycle.
	// create new stop signal channel.
	r.stopCh = make(chan struct{})
	// create new completion notification channel.
	r.doneCh = make(chan struct{})
	r.mu.Unlock()
	// start background reaping loop.
	go r.reapLoop()
}

// Stop terminates the reaping loop and waits for completion.
func (r *Reaper) Stop() {
	r.mu.Lock()
	// Not running; nothing to stop.
	if !r.running {
		// release lock and skip stop.
		r.mu.Unlock()
		// exit early when already stopped.
		return
	}
	r.running = false
	r.mu.Unlock()
	// signal reaper loop to terminate.
	close(r.stopCh)
	// Wait for final reap cycle to complete.
	<-r.doneCh
}

// reapLoop waits for SIGCHLD and reaps zombies until stopped.
func (r *Reaper) reapLoop() {
	// signal completion on function exit.
	defer close(r.doneCh)
	// create buffered channel for SIGCHLD notifications.
	sigCh := make(chan os.Signal, 1)
	// register for child termination signals.
	signal.Notify(sigCh, syscall.SIGCHLD)
	// unregister signal handler on exit.
	defer signal.Stop(sigCh)
	// loop until stop requested.
	for {
		// wait for termination or child signal.
		select {
		// Shutdown requested; do final cleanup.
		case <-r.stopCh:
			// collect any remaining zombies before exit.
			r.reapAll()
			// terminate reaper loop.
			return
		// Child terminated; reap all pending zombies.
		case <-sigCh:
			// collect all terminated child processes.
			r.reapAll()
		}
	}
}

// reapAll uses non-blocking waitpid to collect all terminated children.
func (r *Reaper) reapAll() {
	// Loop until no more zombies remain.
	// repeatedly collect terminated children.
	for {
		// attempt non-blocking wait for any child process.
		pid, err := syscall.Wait4(-1, nil, syscall.WNOHANG, nil)
		// No more children or error (ECHILD when no children exist).
		if err != nil || pid <= 0 {
			// exit loop when no more zombies to reap.
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
	// repeatedly collect terminated children.
	for {
		// attempt non-blocking wait for any child process.
		pid, err := syscall.Wait4(-1, nil, syscall.WNOHANG, nil)
		// No more children or error (ECHILD when no children exist).
		if err != nil || pid <= 0 {
			// exit loop when no more zombies to reap.
			break
		}
		// increment count of reaped processes.
		count++
	}
	// return total number of reaped zombies.
	return count
}

// IsPID1 checks if running as init process (needed for subreaper role).
//
// Returns:
//   - bool: true when running as PID 1 (container init or system init)
func (r *Reaper) IsPID1() bool {
	// check if current process ID is 1.
	return os.Getpid() == 1
}
