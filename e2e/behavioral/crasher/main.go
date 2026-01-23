// Package main provides a controllable test binary for E2E behavioral testing.
// It simulates various process behaviors (crashes, health endpoints, orphans)
// that supervizio must handle correctly.
package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

var (
	exitCode   = flag.Int("exit", 0, "Exit code to return")
	delay      = flag.Duration("delay", 0, "Delay before exit (0 = immediate)")
	port       = flag.Int("port", 0, "TCP port to listen on (0 = no listener)")
	httpHealth = flag.Bool("http", false, "Serve HTTP /health endpoint")
	spawnOrphan = flag.Bool("orphan", false, "Spawn an orphan process before exit")
	ignoreTerm = flag.Bool("ignore-term", false, "Ignore SIGTERM signal")
	logFile    = flag.String("log", "", "Log file path (empty = stdout)")
	crashAfter = flag.Int("crash-after", 0, "Crash after N seconds (0 = use delay)")
	healthy    = flag.Bool("healthy", true, "Health endpoint returns 200 (false = 503)")
)

func main() {
	flag.Parse()

	// Setup logging
	logOutput := os.Stdout
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open log file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		logOutput = f
	}

	log := func(format string, args ...any) {
		fmt.Fprintf(logOutput, "[crasher] %s %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(format, args...))
	}

	log("starting with exit=%d delay=%v port=%d http=%v orphan=%v ignore-term=%v crash-after=%d healthy=%v",
		*exitCode, *delay, *port, *httpHealth, *spawnOrphan, *ignoreTerm, *crashAfter, *healthy)

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	if *ignoreTerm {
		log("ignoring SIGTERM")
		signal.Ignore(syscall.SIGTERM)
	} else {
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		go func() {
			sig := <-sigCh
			log("received signal: %v, exiting gracefully", sig)
			os.Exit(0)
		}()
	}

	// Start HTTP server if requested
	if *httpHealth && *port > 0 {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			if *healthy {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "OK")
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
				fmt.Fprintln(w, "UNHEALTHY")
			}
		})
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "crasher")
		})

		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", *port),
			Handler: mux,
		}

		go func() {
			log("starting HTTP server on port %d", *port)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log("HTTP server error: %v", err)
			}
		}()
	} else if *port > 0 {
		// TCP-only listener (no HTTP)
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
		if err != nil {
			log("failed to listen on port %d: %v", *port, err)
			os.Exit(1)
		}
		defer listener.Close()

		go func() {
			log("starting TCP listener on port %d", *port)
			for {
				conn, err := listener.Accept()
				if err != nil {
					return
				}
				conn.Write([]byte("OK\n"))
				conn.Close()
			}
		}()
	}

	// Spawn orphan process if requested
	if *spawnOrphan {
		log("spawning orphan process")
		cmd := exec.Command("sleep", "60")
		if err := cmd.Start(); err != nil {
			log("failed to spawn orphan: %v", err)
		} else {
			log("orphan process started with PID %d", cmd.Process.Pid)
		}
		// Intentionally not waiting for the orphan - it becomes a zombie/orphan
	}

	// Determine wait duration
	waitDuration := *delay
	if *crashAfter > 0 {
		waitDuration = time.Duration(*crashAfter) * time.Second
	}

	// Wait before exiting
	if waitDuration > 0 {
		log("waiting %v before exit", waitDuration)
		time.Sleep(waitDuration)
	}

	log("exiting with code %d", *exitCode)
	os.Exit(*exitCode)
}
