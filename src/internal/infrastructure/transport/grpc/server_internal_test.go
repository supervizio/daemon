// Package grpc provides gRPC server implementation for the daemon API.
package grpc

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"

	daemonpb "github.com/kodflow/daemon/api/proto/v1/daemon"
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

// mockMetricsProviderFailAfter fails after N successful calls.
// Uses a call log slice to track invocations without direct increment in getter.
type mockMetricsProviderFailAfter struct {
	processMetrics    metrics.ProcessMetrics
	allProcessMetrics []metrics.ProcessMetrics
	failAfter         int
	callLog           []time.Time
	err               error
}

// trackCall records a call timestamp and returns the current count.
func (m *mockMetricsProviderFailAfter) trackCall() int {
	m.callLog = append(m.callLog, time.Now())
	return len(m.callLog)
}

func (m *mockMetricsProviderFailAfter) GetProcessMetrics(_ string) (metrics.ProcessMetrics, error) {
	// Track this call and check if should fail.
	callCount := m.trackCall()
	if callCount > m.failAfter {
		return metrics.ProcessMetrics{}, m.err
	}
	return m.processMetrics, nil
}

func (m *mockMetricsProviderFailAfter) GetAllProcessMetrics() []metrics.ProcessMetrics {
	return m.allProcessMetrics
}

func (m *mockMetricsProviderFailAfter) Subscribe() <-chan metrics.ProcessMetrics {
	ch := make(chan metrics.ProcessMetrics, 1)
	close(ch)
	return ch
}

func (m *mockMetricsProviderFailAfter) Unsubscribe(_ <-chan metrics.ProcessMetrics) {}

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

// Test_Server_RequestHandling verifies request handling edge cases for all endpoints.
//
// Params:
//   - t: testing context for assertions
func Test_Server_RequestHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		endpoint        string
		testType        string
		metricsProvider *mockMetricsProvider
		stateProvider   *mockGetStator
		wantErr         bool
		wantNilResult   bool
		errorContains   string
	}{
		{
			name:            "GetState_nilRequest",
			endpoint:        "GetState",
			testType:        "nil",
			metricsProvider: &mockMetricsProvider{},
			stateProvider:   &mockGetStator{},
			wantErr:         false,
			wantNilResult:   true,
		},
		{
			name:            "GetState_cancelledContext",
			endpoint:        "GetState",
			testType:        "cancelled",
			metricsProvider: &mockMetricsProvider{},
			stateProvider:   &mockGetStator{},
			wantErr:         true,
		},
		{
			name:     "ListProcesses_nilRequest",
			endpoint: "ListProcesses",
			testType: "nil",
			metricsProvider: &mockMetricsProvider{
				allProcessMetrics: []metrics.ProcessMetrics{
					{ServiceName: "test-service", PID: 1234},
				},
			},
			stateProvider: &mockGetStator{},
			wantErr:       false,
			wantNilResult: true,
		},
		{
			name:     "ListProcesses_cancelledContext",
			endpoint: "ListProcesses",
			testType: "cancelled",
			metricsProvider: &mockMetricsProvider{
				allProcessMetrics: []metrics.ProcessMetrics{
					{ServiceName: "test-service", PID: 1234},
				},
			},
			stateProvider: &mockGetStator{},
			wantErr:       true,
		},
		{
			name:            "GetSystemMetrics_nilRequest",
			endpoint:        "GetSystemMetrics",
			testType:        "nil",
			metricsProvider: &mockMetricsProvider{},
			stateProvider:   &mockGetStator{},
			wantErr:         false,
			wantNilResult:   true,
		},
		{
			name:            "GetSystemMetrics_cancelledContext",
			endpoint:        "GetSystemMetrics",
			testType:        "cancelled",
			metricsProvider: &mockMetricsProvider{},
			stateProvider: &mockGetStator{
				state: lifecycle.DaemonState{
					System: lifecycle.SystemState{
						CPU:    metrics.SystemCPU{User: 100},
						Memory: metrics.SystemMemory{Total: 1024},
					},
				},
			},
			wantErr: true,
		},
		{
			name:     "GetProcess_error",
			endpoint: "GetProcess",
			testType: "error",
			metricsProvider: &mockMetricsProvider{
				err: errors.New("process not found"),
			},
			stateProvider: &mockGetStator{},
			wantErr:       true,
			errorContains: "process not found",
		},
		{
			name:            "GetProcess_cancelledContext",
			endpoint:        "GetProcess",
			testType:        "cancelled",
			metricsProvider: &mockMetricsProvider{},
			stateProvider:   &mockGetStator{},
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := NewServer(tt.metricsProvider, tt.stateProvider)

			var err error
			var result any
			ctx := context.Background()

			if tt.testType == "cancelled" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			}

			switch tt.endpoint {
			case "GetState":
				if tt.testType == "nil" {
					result, err = server.GetState(ctx, nil)
				} else {
					result, err = server.GetState(ctx, &emptypb.Empty{})
				}
			case "ListProcesses":
				if tt.testType == "nil" {
					result, err = server.ListProcesses(ctx, nil)
				} else {
					result, err = server.ListProcesses(ctx, &emptypb.Empty{})
				}
			case "GetSystemMetrics":
				if tt.testType == "nil" {
					result, err = server.GetSystemMetrics(ctx, nil)
				} else {
					result, err = server.GetSystemMetrics(ctx, &emptypb.Empty{})
				}
			case "GetProcess":
				req := &daemonpb.GetProcessRequest{ServiceName: "test-service"}
				result, err = server.GetProcess(ctx, req)
			}

			if tt.wantErr {
				assert.Error(t, err)
				if tt.testType == "cancelled" {
					assert.ErrorIs(t, err, context.Canceled)
				}
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.wantNilResult {
					assert.Nil(t, result)
				}
			}
		})
	}
}

// Test_Server_ErrorHandling verifies server error handling for various scenarios.
//
// Params:
//   - t: testing context for assertions
func Test_Server_ErrorHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		testType      string
		errorContains string
	}{
		{
			name:          "Serve_listenError",
			testType:      "serve",
			errorContains: "listen",
		},
		{
			name:          "streamLoop_emitErrorAfterInitial",
			testType:      "streamLoop",
			errorContains: "stop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var err error

			switch tt.testType {
			case "serve":
				metricsProvider := &mockMetricsProvider{}
				stateProvider := &mockGetStator{}
				server := NewServer(metricsProvider, stateProvider)
				err = server.Serve(context.Background(), "invalid-address-no-port")

			case "streamLoop":
				var callCount int
				emit := func() (string, error) {
					callCount++
					if callCount == 1 {
						return "initial", nil
					}
					if callCount <= 3 {
						return "", errors.New("emit error")
					}
					return "recovered", nil
				}

				var sent []string
				send := func(val string) error {
					sent = append(sent, val)
					if len(sent) >= 2 {
						return errors.New("stop")
					}
					return nil
				}

				ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
				defer cancel()
				err = streamLoop(ctx, 10*time.Millisecond, emit, send)
			}

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorContains)
		})
	}
}

// mockInternalStreamStateServer is a simple mock for StreamState internal tests.
type mockInternalStreamStateServer struct {
	daemonpb.DaemonService_StreamStateServer
	ctx context.Context
}

func (m *mockInternalStreamStateServer) Context() context.Context {
	return m.ctx
}

func (m *mockInternalStreamStateServer) Send(_ *daemonpb.DaemonState) error {
	return nil
}

// Test_Server_StreamCustomInterval verifies all stream methods use custom intervals correctly.
//
// Params:
//   - t: testing context for assertions
func Test_Server_StreamCustomInterval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		streamMethod string
	}{
		{name: "StreamState_customInterval", streamMethod: "StreamState"},
		{name: "StreamProcessMetrics_customInterval", streamMethod: "StreamProcessMetrics"},
		{name: "StreamSystemMetrics_customInterval", streamMethod: "StreamSystemMetrics"},
		{name: "StreamAllProcessMetrics_customInterval", streamMethod: "StreamAllProcessMetrics"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			metricsProvider := &mockMetricsProvider{
				processMetrics: metrics.ProcessMetrics{ServiceName: "test", PID: 123},
				allProcessMetrics: []metrics.ProcessMetrics{
					{ServiceName: "test", PID: 123},
				},
			}
			stateProvider := &mockGetStator{}
			server := NewServer(metricsProvider, stateProvider)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			var err error
			switch tt.streamMethod {
			case "StreamState":
				req := &daemonpb.StreamStateRequest{
					Interval: durationpb.New(100 * time.Millisecond),
				}
				mockStream := &mockInternalStreamStateServer{ctx: ctx}
				err = server.StreamState(req, mockStream)
			case "StreamProcessMetrics":
				req := &daemonpb.StreamProcessMetricsRequest{
					ServiceName: "test",
					Interval:    durationpb.New(100 * time.Millisecond),
				}
				mockStream := &mockInternalStreamProcessMetricsServer{ctx: ctx}
				err = server.StreamProcessMetrics(req, mockStream)
			case "StreamSystemMetrics":
				req := &daemonpb.StreamMetricsRequest{
					Interval: durationpb.New(100 * time.Millisecond),
				}
				mockStream := &mockInternalStreamSystemMetricsServer{ctx: ctx}
				err = server.StreamSystemMetrics(req, mockStream)
			case "StreamAllProcessMetrics":
				req := &daemonpb.StreamMetricsRequest{
					Interval: durationpb.New(100 * time.Millisecond),
				}
				mockStream := &mockInternalStreamAllProcessMetricsServer{ctx: ctx}
				err = server.StreamAllProcessMetrics(req, mockStream)
			}

			assert.Error(t, err)
		})
	}
}

// mockInternalStreamProcessMetricsServer is a simple mock for StreamProcessMetrics internal tests.
type mockInternalStreamProcessMetricsServer struct {
	daemonpb.DaemonService_StreamProcessMetricsServer
	ctx       context.Context
	sendCount int
}

func (m *mockInternalStreamProcessMetricsServer) Context() context.Context {
	return m.ctx
}

func (m *mockInternalStreamProcessMetricsServer) Send(_ *daemonpb.ProcessMetrics) error {
	m.sendCount++
	return nil
}

// mockInternalStreamSystemMetricsServer is a simple mock for StreamSystemMetrics internal tests.
type mockInternalStreamSystemMetricsServer struct {
	daemonpb.MetricsService_StreamSystemMetricsServer
	ctx context.Context
}

func (m *mockInternalStreamSystemMetricsServer) Context() context.Context {
	return m.ctx
}

func (m *mockInternalStreamSystemMetricsServer) Send(_ *daemonpb.SystemMetrics) error {
	return nil
}

// mockInternalStreamAllProcessMetricsServer is a simple mock for StreamAllProcessMetrics internal tests.
type mockInternalStreamAllProcessMetricsServer struct {
	daemonpb.MetricsService_StreamAllProcessMetricsServer
	ctx context.Context
}

func (m *mockInternalStreamAllProcessMetricsServer) Context() context.Context {
	return m.ctx
}

func (m *mockInternalStreamAllProcessMetricsServer) Send(_ *daemonpb.ProcessMetrics) error {
	return nil
}

// mockInternalStreamAllProcessMetricsServerWithSend is a mock that allows ticker to fire.
type mockInternalStreamAllProcessMetricsServerWithSend struct {
	daemonpb.MetricsService_StreamAllProcessMetricsServer
	ctx       context.Context
	sendCount int
	sendErr   error
}

func (m *mockInternalStreamAllProcessMetricsServerWithSend) Context() context.Context {
	return m.ctx
}

func (m *mockInternalStreamAllProcessMetricsServerWithSend) Send(_ *daemonpb.ProcessMetrics) error {
	m.sendCount++
	return m.sendErr
}

// Test_Server_StreamBehaviors verifies streaming behaviors for various scenarios.
//
// Params:
//   - t: testing context for assertions
func Test_Server_StreamBehaviors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		testType        string
		sendErr         error
		minSendCount    int
		errorContains   string
		useFailProvider bool
	}{
		{
			name:         "StreamAllProcessMetrics_tickerFires",
			testType:     "tickerFires",
			minSendCount: 1,
		},
		{
			name:          "StreamAllProcessMetrics_sendError",
			testType:      "sendError",
			sendErr:       errors.New("send failed"),
			errorContains: "send failed",
		},
		{
			name:            "StreamProcessMetrics_emitErrorDuringStreaming",
			testType:        "emitError",
			minSendCount:    1,
			useFailProvider: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stateProvider = &mockGetStator{}
			var err error

			switch tt.testType {
			case "tickerFires":
				metricsProvider := &mockMetricsProvider{
					allProcessMetrics: []metrics.ProcessMetrics{
						{ServiceName: "test1", PID: 123},
						{ServiceName: "test2", PID: 456},
					},
				}
				server := NewServer(metricsProvider, stateProvider)

				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()

				req := &daemonpb.StreamMetricsRequest{
					Interval: durationpb.New(10 * time.Millisecond),
				}
				mockStream := &mockInternalStreamAllProcessMetricsServerWithSend{ctx: ctx}

				err = server.StreamAllProcessMetrics(req, mockStream)
				assert.Error(t, err)
				assert.Greater(t, mockStream.sendCount, tt.minSendCount-1)

			case "sendError":
				metricsProvider := &mockMetricsProvider{
					allProcessMetrics: []metrics.ProcessMetrics{
						{ServiceName: "test", PID: 123},
					},
				}
				server := NewServer(metricsProvider, stateProvider)

				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()

				req := &daemonpb.StreamMetricsRequest{
					Interval: durationpb.New(10 * time.Millisecond),
				}
				mockStream := &mockInternalStreamAllProcessMetricsServerWithSend{
					ctx:     ctx,
					sendErr: tt.sendErr,
				}

				err = server.StreamAllProcessMetrics(req, mockStream)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)

			case "emitError":
				metricsProvider := &mockMetricsProviderFailAfter{
					processMetrics: metrics.ProcessMetrics{ServiceName: "test", PID: 123},
					failAfter:      1,
					err:            errors.New("metrics unavailable"),
				}
				server := NewServer(metricsProvider, stateProvider)

				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()

				req := &daemonpb.StreamProcessMetricsRequest{
					ServiceName: "test",
					Interval:    durationpb.New(10 * time.Millisecond),
				}
				mockStream := &mockInternalStreamProcessMetricsServer{ctx: ctx}

				err = server.StreamProcessMetrics(req, mockStream)
				assert.Error(t, err)
				assert.GreaterOrEqual(t, mockStream.sendCount, tt.minSendCount)
			}
		})
	}
}
