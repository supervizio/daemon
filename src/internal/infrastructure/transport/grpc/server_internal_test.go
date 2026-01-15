// Package grpc provides gRPC server implementation for the daemon API.
package grpc

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/lifecycle"
	"github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/domain/process"
)

// mockMetricsProvider provides test metrics.
type mockMetricsProvider struct {
	processMetrics    metrics.ProcessMetrics
	allProcessMetrics []metrics.ProcessMetrics
	err               error
}

func (m *mockMetricsProvider) GetProcessMetrics(_ string) (metrics.ProcessMetrics, error) {
	return m.processMetrics, m.err
}

func (m *mockMetricsProvider) GetAllProcessMetrics() []metrics.ProcessMetrics {
	return m.allProcessMetrics
}

func (m *mockMetricsProvider) Subscribe() <-chan metrics.ProcessMetrics {
	ch := make(chan metrics.ProcessMetrics, 1)
	return ch
}

func (m *mockMetricsProvider) Unsubscribe(_ <-chan metrics.ProcessMetrics) {
}

// mockGetStator provides test daemon state.
type mockGetStator struct {
	state lifecycle.DaemonState
}

func (m *mockGetStator) GetState() lifecycle.DaemonState {
	return m.state
}

// Test_safeInt32 verifies that safeInt32 correctly converts integers with bounds checking.
//
// Params:
//   - t: testing context for assertions
func Test_safeInt32(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    int
		expected int32
	}{
		{name: "within bounds", input: 100, expected: 100},
		{name: "max int32", input: math.MaxInt32, expected: math.MaxInt32},
		{name: "min int32", input: math.MinInt32, expected: math.MinInt32},
		{name: "above max", input: math.MaxInt32 + 1, expected: math.MaxInt32},
		{name: "below min", input: math.MinInt32 - 1, expected: math.MinInt32},
		{name: "zero", input: 0, expected: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := safeInt32(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test_streamLoop verifies that streamLoop correctly handles streaming with context cancellation.
//
// Goroutine lifecycle: The test spawns a cancellation goroutine that sleeps for a configured
// duration then calls cancel(). This goroutine terminates when cancel() completes. The main
// test goroutine blocks on streamLoop() until context cancellation or error occurs.
//
// Params:
//   - t: testing context for assertions
func Test_streamLoop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		emitErr        error
		sendErr        error
		cancelAfter    time.Duration
		expectedErr    error
		minEmitCount   int
		minSendCount   int
		skipAfterFirst bool
	}{
		{
			name:         "cancels on context done",
			cancelAfter:  50 * time.Millisecond,
			expectedErr:  context.Canceled,
			minEmitCount: 1,
			minSendCount: 1,
		},
		{
			name:         "returns error from initial emit",
			emitErr:      errors.New("emit error"),
			expectedErr:  errors.New("emit error"),
			minEmitCount: 0,
			minSendCount: 0,
		},
		{
			name:         "returns error from send",
			sendErr:      errors.New("send error"),
			expectedErr:  errors.New("send error"),
			minEmitCount: 0,
			minSendCount: 0,
		},
		{
			name:           "continues on emit error after initial",
			cancelAfter:    80 * time.Millisecond,
			expectedErr:    context.Canceled,
			minEmitCount:   1,
			minSendCount:   1,
			skipAfterFirst: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			emitCount := 0
			sendCount := 0

			if tt.cancelAfter > 0 {
				// Goroutine lifecycle: Created here, terminates when cancel() is called.
				// This goroutine simply triggers context cancellation after a delay.
				go func() {
					time.Sleep(tt.cancelAfter)
					cancel()
				}()
			}

			err := streamLoop(
				ctx,
				10*time.Millisecond,
				func() (int, error) {
					emitCount++
					if tt.emitErr != nil {
						return 0, tt.emitErr
					}
					if tt.skipAfterFirst && emitCount > 1 {
						return 0, errors.New("transient error")
					}
					return emitCount, nil
				},
				func(_ int) error {
					sendCount++
					if tt.sendErr != nil {
						return tt.sendErr
					}
					return nil
				},
			)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				if errors.Is(tt.expectedErr, context.Canceled) {
					assert.ErrorIs(t, err, context.Canceled)
				}
			}
			assert.GreaterOrEqual(t, emitCount, tt.minEmitCount)
			assert.GreaterOrEqual(t, sendCount, tt.minSendCount)
		})
	}
}

// Test_Server_convertProcessState verifies that convertProcessState correctly maps domain states.
//
// Params:
//   - t: testing context for assertions
func Test_Server_convertProcessState(t *testing.T) {
	t.Parallel()

	metricsProvider := &mockMetricsProvider{}
	stateProvider := &mockGetStator{}
	server := NewServer(metricsProvider, stateProvider)

	tests := []struct {
		name     string
		state    process.State
		expected string
	}{
		{name: "stopped", state: process.StateStopped, expected: "PROCESS_STATE_STOPPED"},
		{name: "starting", state: process.StateStarting, expected: "PROCESS_STATE_STARTING"},
		{name: "running", state: process.StateRunning, expected: "PROCESS_STATE_RUNNING"},
		{name: "stopping", state: process.StateStopping, expected: "PROCESS_STATE_STOPPING"},
		{name: "failed", state: process.StateFailed, expected: "PROCESS_STATE_FAILED"},
		{name: "unknown", state: process.State(99), expected: "PROCESS_STATE_UNSPECIFIED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := server.convertProcessState(tt.state)
			assert.Contains(t, result.String(), tt.expected)
		})
	}
}

// Test_Server_convertDaemonState verifies that convertDaemonState correctly converts domain state.
//
// Params:
//   - t: testing context for assertions
func Test_Server_convertDaemonState(t *testing.T) {
	t.Parallel()

	metricsProvider := &mockMetricsProvider{}
	stateProvider := &mockGetStator{}
	server := NewServer(metricsProvider, stateProvider)

	startTime := time.Now().Add(-time.Hour)

	tests := []struct {
		name            string
		daemonState     *lifecycle.DaemonState
		expectedVersion string
		expectedHost    string
		expectedHealthy bool
		processCount    int
	}{
		{
			name: "healthy state with processes",
			daemonState: &lifecycle.DaemonState{
				Host: lifecycle.HostInfo{
					DaemonVersion: "1.0.0",
					StartTime:     startTime,
					Hostname:      "test-host",
					OS:            "linux",
					Arch:          "amd64",
				},
				Processes: []metrics.ProcessMetrics{
					{
						ServiceName: "test-service",
						PID:         1234,
						State:       process.StateRunning,
						Healthy:     true,
					},
				},
				System: lifecycle.SystemState{
					CPU:    metrics.SystemCPU{User: 1000, System: 2000},
					Memory: metrics.SystemMemory{Total: 1024 * 1024 * 1024},
				},
			},
			expectedVersion: "1.0.0",
			expectedHost:    "test-host",
			expectedHealthy: true,
			processCount:    1,
		},
		{
			name: "empty processes",
			daemonState: &lifecycle.DaemonState{
				Host: lifecycle.HostInfo{
					DaemonVersion: "2.0.0",
					StartTime:     startTime,
					Hostname:      "empty-host",
				},
				Processes: []metrics.ProcessMetrics{},
			},
			expectedVersion: "2.0.0",
			expectedHost:    "empty-host",
			expectedHealthy: false,
			processCount:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := server.convertDaemonState(tt.daemonState)

			assert.Equal(t, tt.expectedVersion, result.Version)
			assert.Equal(t, tt.expectedHost, result.Host.Hostname)
			assert.Equal(t, tt.expectedHealthy, result.Healthy)
			assert.Len(t, result.Processes, tt.processCount)
		})
	}
}

// Test_Server_convertProcessMetrics verifies that convertProcessMetrics correctly converts domain metrics.
//
// Params:
//   - t: testing context for assertions
func Test_Server_convertProcessMetrics(t *testing.T) {
	t.Parallel()

	metricsProvider := &mockMetricsProvider{}
	stateProvider := &mockGetStator{}
	server := NewServer(metricsProvider, stateProvider)

	startTime := time.Now().Add(-time.Hour)
	timestamp := time.Now()

	tests := []struct {
		name            string
		metrics         *metrics.ProcessMetrics
		expectedName    string
		expectedPID     int32
		expectedHealthy bool
		expectedRestart int32
	}{
		{
			name: "healthy running process",
			metrics: &metrics.ProcessMetrics{
				ServiceName:  "test-service",
				PID:          1234,
				State:        process.StateRunning,
				Healthy:      true,
				CPU:          metrics.ProcessCPU{User: 1000, System: 2000},
				Memory:       metrics.ProcessMemory{RSS: 1024, VMS: 2048},
				StartTime:    startTime,
				Uptime:       time.Hour,
				RestartCount: 3,
				LastError:    "none",
				Timestamp:    timestamp,
			},
			expectedName:    "test-service",
			expectedPID:     1234,
			expectedHealthy: true,
			expectedRestart: 3,
		},
		{
			name: "unhealthy failed process",
			metrics: &metrics.ProcessMetrics{
				ServiceName:  "failed-service",
				PID:          0,
				State:        process.StateFailed,
				Healthy:      false,
				RestartCount: 10,
				LastError:    "exit code 1",
				Timestamp:    timestamp,
			},
			expectedName:    "failed-service",
			expectedPID:     0,
			expectedHealthy: false,
			expectedRestart: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := server.convertProcessMetrics(tt.metrics)

			assert.Equal(t, tt.expectedName, result.ServiceName)
			assert.Equal(t, tt.expectedPID, result.Pid)
			assert.Equal(t, tt.expectedHealthy, result.Healthy)
			assert.Equal(t, tt.expectedRestart, result.RestartCount)
		})
	}
}

// Test_Server_convertProcessCPU verifies that convertProcessCPU correctly converts CPU metrics.
//
// Params:
//   - t: testing context for assertions
func Test_Server_convertProcessCPU(t *testing.T) {
	t.Parallel()

	metricsProvider := &mockMetricsProvider{}
	stateProvider := &mockGetStator{}
	server := NewServer(metricsProvider, stateProvider)

	tests := []struct {
		name           string
		cpu            *metrics.ProcessCPU
		expectedUser   uint64
		expectedSystem uint64
		expectedTotal  uint64
	}{
		{
			name:           "normal values",
			cpu:            &metrics.ProcessCPU{User: 1000, System: 2000},
			expectedUser:   1000,
			expectedSystem: 2000,
			expectedTotal:  3000,
		},
		{
			name:           "zero values",
			cpu:            &metrics.ProcessCPU{User: 0, System: 0},
			expectedUser:   0,
			expectedSystem: 0,
			expectedTotal:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := server.convertProcessCPU(tt.cpu)

			assert.Equal(t, tt.expectedUser, result.UserTimeNs)
			assert.Equal(t, tt.expectedSystem, result.SystemTimeNs)
			assert.Equal(t, tt.expectedTotal, result.TotalTimeNs)
		})
	}
}

// Test_Server_convertProcessMemory verifies that convertProcessMemory correctly converts memory metrics.
//
// Params:
//   - t: testing context for assertions
func Test_Server_convertProcessMemory(t *testing.T) {
	t.Parallel()

	metricsProvider := &mockMetricsProvider{}
	stateProvider := &mockGetStator{}
	server := NewServer(metricsProvider, stateProvider)

	tests := []struct {
		name        string
		mem         *metrics.ProcessMemory
		expectedRSS uint64
		expectedVMS uint64
	}{
		{
			name: "all fields populated",
			mem: &metrics.ProcessMemory{
				RSS:    1024,
				VMS:    2048,
				Swap:   512,
				Shared: 256,
				Data:   128,
				Stack:  64,
			},
			expectedRSS: 1024,
			expectedVMS: 2048,
		},
		{
			name:        "zero values",
			mem:         &metrics.ProcessMemory{},
			expectedRSS: 0,
			expectedVMS: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := server.convertProcessMemory(tt.mem)

			assert.Equal(t, tt.expectedRSS, result.RssBytes)
			assert.Equal(t, tt.expectedVMS, result.VmsBytes)
		})
	}
}

// Test_Server_convertSystemMetrics verifies that convertSystemMetrics correctly converts system metrics.
//
// Params:
//   - t: testing context for assertions
func Test_Server_convertSystemMetrics(t *testing.T) {
	t.Parallel()

	metricsProvider := &mockMetricsProvider{}
	stateProvider := &mockGetStator{}
	server := NewServer(metricsProvider, stateProvider)

	timestamp := time.Now()

	tests := []struct {
		name          string
		daemonState   *lifecycle.DaemonState
		expectedUser  uint64
		expectedTotal uint64
	}{
		{
			name: "populated metrics",
			daemonState: &lifecycle.DaemonState{
				System: lifecycle.SystemState{
					CPU: metrics.SystemCPU{
						User:    1000,
						Nice:    100,
						System:  2000,
						Idle:    5000,
						IOWait:  300,
						IRQ:     50,
						SoftIRQ: 25,
						Steal:   10,
					},
					Memory: metrics.SystemMemory{
						Total:     1024 * 1024 * 1024,
						Available: 512 * 1024 * 1024,
					},
				},
				Timestamp: timestamp,
			},
			expectedUser:  1000,
			expectedTotal: 1024 * 1024 * 1024,
		},
		{
			name: "zero metrics",
			daemonState: &lifecycle.DaemonState{
				System:    lifecycle.SystemState{},
				Timestamp: timestamp,
			},
			expectedUser:  0,
			expectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := server.convertSystemMetrics(tt.daemonState)

			assert.Equal(t, tt.expectedUser, result.Cpu.UserNs)
			assert.Equal(t, tt.expectedTotal, result.Memory.TotalBytes)
		})
	}
}

// Test_Server_convertHostInfo verifies that convertHostInfo correctly converts host information.
//
// Params:
//   - t: testing context for assertions
func Test_Server_convertHostInfo(t *testing.T) {
	t.Parallel()

	metricsProvider := &mockMetricsProvider{}
	stateProvider := &mockGetStator{}
	server := NewServer(metricsProvider, stateProvider)

	tests := []struct {
		name         string
		host         *lifecycle.HostInfo
		expectedHost string
		expectedOS   string
		expectedArch string
	}{
		{
			name: "linux host",
			host: &lifecycle.HostInfo{
				Hostname: "test-host",
				OS:       "linux",
				Arch:     "amd64",
			},
			expectedHost: "test-host",
			expectedOS:   "linux",
			expectedArch: "amd64",
		},
		{
			name: "darwin host",
			host: &lifecycle.HostInfo{
				Hostname: "mac-host",
				OS:       "darwin",
				Arch:     "arm64",
			},
			expectedHost: "mac-host",
			expectedOS:   "darwin",
			expectedArch: "arm64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := server.convertHostInfo(tt.host)

			assert.Equal(t, tt.expectedHost, result.Hostname)
			assert.Equal(t, tt.expectedOS, result.Os)
			assert.Equal(t, tt.expectedArch, result.Arch)
		})
	}
}

// Test_Server_convertKubernetesInfo verifies that convertKubernetesInfo correctly handles K8s state.
//
// Params:
//   - t: testing context for assertions
func Test_Server_convertKubernetesInfo(t *testing.T) {
	t.Parallel()

	metricsProvider := &mockMetricsProvider{}
	stateProvider := &mockGetStator{}
	server := NewServer(metricsProvider, stateProvider)

	tests := []struct {
		name        string
		k8s         *lifecycle.KubernetesState
		expectedNil bool
		expectedPod string
		expectedNS  string
	}{
		{
			name: "valid kubernetes state",
			k8s: &lifecycle.KubernetesState{
				PodName:   "test-pod",
				Namespace: "test-namespace",
				NodeName:  "test-node",
			},
			expectedNil: false,
			expectedPod: "test-pod",
			expectedNS:  "test-namespace",
		},
		{
			name:        "nil kubernetes state",
			k8s:         nil,
			expectedNil: true,
		},
		{
			name: "empty pod name",
			k8s: &lifecycle.KubernetesState{
				PodName: "",
			},
			expectedNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := server.convertKubernetesInfo(tt.k8s)

			if tt.expectedNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedPod, result.PodName)
				assert.Equal(t, tt.expectedNS, result.Namespace)
			}
		})
	}
}
