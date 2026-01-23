package behavioral_test

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// defaultStartupTimeout is the maximum time to wait for container startup.
	defaultStartupTimeout = 30 * time.Second
	// defaultPollInterval is the interval between status checks.
	defaultPollInterval = 100 * time.Millisecond
)

// testContainer wraps a testcontainers container with helper methods.
type testContainer struct {
	container testcontainers.Container
	ctx       context.Context
	t         *testing.T
}

// startContainer creates and starts a container with the specified config file.
func startContainer(t *testing.T, configFile string) *testContainer {
	t.Helper()
	ctx := context.Background()

	// Get absolute paths for the build context
	absPath, err := filepath.Abs("../..")
	require.NoError(t, err, "failed to get absolute path")

	configPath := fmt.Sprintf("e2e/behavioral/testdata/%s", configFile)

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    absPath,
			Dockerfile: "e2e/behavioral/Dockerfile.behavioral",
		},
		Cmd: []string{"--config", "/etc/daemon/config.yaml"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(absPath, configPath),
				ContainerFilePath: "/etc/daemon/config.yaml",
				FileMode:          0644,
			},
		},
		// Wait for container to start and supervizio to be running as PID 1
		WaitingFor: wait.ForExec([]string{"pgrep", "-x", "supervizio"}).
			WithStartupTimeout(defaultStartupTimeout).
			WithPollInterval(500 * time.Millisecond),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "failed to start container")

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	return &testContainer{
		container: container,
		ctx:       ctx,
		t:         t,
	}
}

// startContainerWithPorts creates and starts a container with exposed ports.
func startContainerWithPorts(t *testing.T, configFile string, ports ...string) *testContainer {
	t.Helper()
	ctx := context.Background()

	absPath, err := filepath.Abs("../..")
	require.NoError(t, err, "failed to get absolute path")

	configPath := fmt.Sprintf("e2e/behavioral/testdata/%s", configFile)

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    absPath,
			Dockerfile: "e2e/behavioral/Dockerfile.behavioral",
		},
		Cmd: []string{"--config", "/etc/daemon/config.yaml"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(absPath, configPath),
				ContainerFilePath: "/etc/daemon/config.yaml",
				FileMode:          0644,
			},
		},
		ExposedPorts: ports,
		// Wait for container to start and supervizio to be running as PID 1
		WaitingFor: wait.ForExec([]string{"pgrep", "-x", "supervizio"}).
			WithStartupTimeout(defaultStartupTimeout).
			WithPollInterval(500 * time.Millisecond),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "failed to start container")

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	return &testContainer{
		container: container,
		ctx:       ctx,
		t:         t,
	}
}

// exec executes a command in the container and returns the exit code and output.
func (tc *testContainer) exec(args ...string) (int, string, error) {
	tc.t.Helper()
	code, reader, err := tc.container.Exec(tc.ctx, args)
	if err != nil {
		return code, "", err
	}

	// Read all output and demux docker exec multiplexed stream
	buf := new(strings.Builder)
	rawBuf := make([]byte, 4096)
	for {
		n, readErr := reader.Read(rawBuf)
		if n > 0 {
			// Docker exec output has 8-byte header per frame
			// Format: [STREAM_TYPE, 0, 0, 0, SIZE1, SIZE2, SIZE3, SIZE4]
			// STREAM_TYPE: 1 = stdout, 2 = stderr
			data := rawBuf[:n]
			for len(data) > 0 {
				if len(data) < 8 {
					// Partial header or no more frames
					buf.Write(data)
					break
				}
				// Skip 8-byte header and extract size
				frameSize := int(data[4])<<24 | int(data[5])<<16 | int(data[6])<<8 | int(data[7])
				data = data[8:]
				if frameSize > len(data) {
					frameSize = len(data)
				}
				buf.Write(data[:frameSize])
				data = data[frameSize:]
			}
		}
		if readErr != nil {
			break
		}
	}

	return code, buf.String(), nil
}

// isProcessRunning checks if a process with the given name is running.
func (tc *testContainer) isProcessRunning(name string) bool {
	tc.t.Helper()
	code, _, _ := tc.exec("pgrep", "-x", name)
	return code == 0
}

// countProcesses returns the number of running processes with the given name.
func (tc *testContainer) countProcesses(name string) int {
	tc.t.Helper()
	code, output, _ := tc.exec("pgrep", "-cx", name)
	if code != 0 {
		return 0
	}
	var count int
	fmt.Sscanf(strings.TrimSpace(output), "%d", &count)
	return count
}

// waitForProcess waits for a process to appear within the timeout.
func (tc *testContainer) waitForProcess(name string, timeout time.Duration) bool {
	tc.t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if tc.isProcessRunning(name) {
			return true
		}
		time.Sleep(defaultPollInterval)
	}
	return false
}

// waitForProcessExit waits for a process to exit within the timeout.
func (tc *testContainer) waitForProcessExit(name string, timeout time.Duration) bool {
	tc.t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !tc.isProcessRunning(name) {
			return true
		}
		time.Sleep(defaultPollInterval)
	}
	return false
}

// killProcess sends SIGKILL to a process by name.
func (tc *testContainer) killProcess(name string) error {
	tc.t.Helper()
	_, _, err := tc.exec("pkill", "-9", "-x", name)
	return err
}

// getZombieCount returns the number of zombie processes in the container.
func (tc *testContainer) getZombieCount() int {
	tc.t.Helper()
	_, output, err := tc.exec("sh", "-c", "ps aux | awk '$8 ~ /Z/ {count++} END {print count+0}'")
	if err != nil {
		return -1
	}
	var count int
	fmt.Sscanf(strings.TrimSpace(output), "%d", &count)
	return count
}

// getLogs returns the container logs.
func (tc *testContainer) getLogs() string {
	tc.t.Helper()
	logs, err := tc.container.Logs(tc.ctx)
	if err != nil {
		return ""
	}
	buf := make([]byte, 32768)
	n, _ := logs.Read(buf)
	return string(buf[:n])
}

// isRunning checks if the container is still running.
func (tc *testContainer) isRunning() bool {
	tc.t.Helper()
	state, err := tc.container.State(tc.ctx)
	if err != nil {
		return false
	}
	return state.Running
}

// getMappedPort returns the host port mapped to a container port.
func (tc *testContainer) getMappedPort(containerPort string) (string, error) {
	tc.t.Helper()
	port, err := tc.container.MappedPort(tc.ctx, nat.Port(containerPort))
	if err != nil {
		return "", err
	}
	return port.Port(), nil
}

// getHost returns the container host.
func (tc *testContainer) getHost() (string, error) {
	tc.t.Helper()
	return tc.container.Host(tc.ctx)
}

// getProcessPID returns the PID of a process by name.
func (tc *testContainer) getProcessPID(name string) int {
	tc.t.Helper()
	_, output, _ := tc.exec("pgrep", "-x", name)
	pid, _ := strconv.Atoi(strings.TrimSpace(output))
	return pid
}

// getLogsReader returns the container logs as an io.ReadCloser.
func (tc *testContainer) getLogsReader() (io.ReadCloser, error) {
	tc.t.Helper()
	return tc.container.Logs(tc.ctx)
}

// waitForLogPattern waits for a log pattern to appear within the timeout.
func (tc *testContainer) waitForLogPattern(pattern string, timeout time.Duration) bool {
	tc.t.Helper()
	deadline := time.Now().Add(timeout)
	re := regexp.MustCompile(pattern)
	for time.Now().Before(deadline) {
		reader, err := tc.getLogsReader()
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		buf := new(strings.Builder)
		_, _ = io.Copy(buf, reader)
		reader.Close()
		if re.MatchString(buf.String()) {
			return true
		}
		time.Sleep(200 * time.Millisecond)
	}
	return false
}
