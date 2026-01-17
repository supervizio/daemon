// Package grpc_test provides black-box tests for the grpc package.
package grpc_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"

	daemonpb "github.com/kodflow/daemon/api/proto/v1/daemon"
	"github.com/kodflow/daemon/internal/domain/lifecycle"
	"github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/infrastructure/transport/grpc"
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

// TestNewServer verifies that NewServer creates a properly configured server.
//
// Params:
//   - t: testing context for assertions
func TestNewServer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		metricsProvider grpc.MetricsProvider
		stateProvider   grpc.GetStator
		expectNotNil    bool
	}{
		{
			name:            "creates server with valid providers",
			metricsProvider: &mockMetricsProvider{},
			stateProvider:   &mockGetStator{},
			expectNotNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := grpc.NewServer(tt.metricsProvider, tt.stateProvider)
			if tt.expectNotNil {
				require.NotNil(t, server)
			}
		})
	}
}

// TestServer_Serve verifies that Serve starts the server and handles errors correctly.
//
// Goroutine lifecycle: Test launches server goroutines that are terminated by
// server.Stop() or context cancellation within each test case.
//
// Params:
//   - t: testing context for assertions
func TestServer_Serve(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func() *grpc.Server
		address     string
		expectError bool
		errorIs     error
	}{
		{
			name: "starts server successfully",
			setup: func() *grpc.Server {
				return grpc.NewServer(&mockMetricsProvider{}, &mockGetStator{})
			},
			address:     "127.0.0.1:0",
			expectError: false,
		},
		{
			name: "returns error when already running",
			setup: func() *grpc.Server {
				server := grpc.NewServer(&mockMetricsProvider{}, &mockGetStator{})
				// Goroutine lifecycle: Starts server, terminated by server.Stop() in cleanup.
				go func() {
					_ = server.Serve("127.0.0.1:0")
				}()
				time.Sleep(100 * time.Millisecond)
				return server
			},
			address:     "127.0.0.1:0",
			expectError: true,
			errorIs:     grpc.ErrServerAlreadyRunning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := tt.setup()
			defer server.Stop()

			if tt.expectError {
				err := server.Serve(tt.address)
				assert.Error(t, err)
				if tt.errorIs != nil {
					assert.ErrorIs(t, err, tt.errorIs)
				}
			} else {
				errCh := make(chan error, 1)
				// Goroutine lifecycle: Starts server, terminated by server.Stop().
				go func() {
					errCh <- server.Serve(tt.address)
				}()

				time.Sleep(100 * time.Millisecond)
				addr := server.Address()
				assert.NotEmpty(t, addr)

				server.Stop()

				select {
				case err := <-errCh:
					assert.NoError(t, err)
				case <-time.After(time.Second):
					t.Fatal("server did not stop")
				}
			}
		})
	}
}

// TestServer_Stop verifies that Stop gracefully shuts down the server.
//
// Goroutine lifecycle: Test launches server goroutines that are terminated by
// server.Stop() call within each test case.
//
// Params:
//   - t: testing context for assertions
func TestServer_Stop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		startServer bool
	}{
		{
			name:        "stops running server",
			startServer: true,
		},
		{
			name:        "does nothing when not running",
			startServer: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := grpc.NewServer(&mockMetricsProvider{}, &mockGetStator{})

			if tt.startServer {
				// Goroutine lifecycle: Starts server, terminated by server.Stop().
				go func() {
					_ = server.Serve("127.0.0.1:0")
				}()
				time.Sleep(100 * time.Millisecond)
			}

			// Should not panic.
			server.Stop()
		})
	}
}

// TestServer_Address verifies that Address returns the correct listening address.
//
// Goroutine lifecycle: Test launches server goroutines that are terminated by
// server.Stop() call within each test case.
//
// Params:
//   - t: testing context for assertions
func TestServer_Address(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		startServer bool
		expectEmpty bool
	}{
		{
			name:        "returns address when running",
			startServer: true,
			expectEmpty: false,
		},
		{
			name:        "returns empty when not running",
			startServer: false,
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := grpc.NewServer(&mockMetricsProvider{}, &mockGetStator{})

			if tt.startServer {
				// Goroutine lifecycle: Starts server, terminated by server.Stop().
				go func() {
					_ = server.Serve("127.0.0.1:0")
				}()
				time.Sleep(100 * time.Millisecond)
				defer server.Stop()
			}

			addr := server.Address()
			if tt.expectEmpty {
				assert.Empty(t, addr)
			} else {
				assert.NotEmpty(t, addr)
			}
		})
	}
}

// TestServer_GetState verifies that GetState returns the current daemon state.
//
// Params:
//   - t: testing context for assertions
func TestServer_GetState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		state           lifecycle.DaemonState
		expectedVersion string
	}{
		{
			name: "returns daemon state",
			state: lifecycle.DaemonState{
				Host: lifecycle.HostInfo{
					DaemonVersion: "1.0.0",
					StartTime:     time.Now(),
					Hostname:      "test-host",
					OS:            "linux",
					Arch:          "amd64",
				},
			},
			expectedVersion: "1.0.0",
		},
		{
			name: "returns different version",
			state: lifecycle.DaemonState{
				Host: lifecycle.HostInfo{
					DaemonVersion: "2.0.0",
					StartTime:     time.Now(),
				},
			},
			expectedVersion: "2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stateProvider := &mockGetStator{state: tt.state}
			server := grpc.NewServer(&mockMetricsProvider{}, stateProvider)

			state, err := server.GetState(context.Background(), &emptypb.Empty{})
			require.NoError(t, err)
			require.NotNil(t, state)
			assert.Equal(t, tt.expectedVersion, state.Version)
		})
	}
}

// TestServer_ListProcesses verifies that ListProcesses returns all process metrics.
//
// Params:
//   - t: testing context for assertions
func TestServer_ListProcesses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		allProcessMetrics []metrics.ProcessMetrics
		expectedCount     int
		expectedFirst     string
	}{
		{
			name: "returns single process",
			allProcessMetrics: []metrics.ProcessMetrics{
				{ServiceName: "test-service", PID: 1234, Timestamp: time.Now()},
			},
			expectedCount: 1,
			expectedFirst: "test-service",
		},
		{
			name: "returns multiple processes",
			allProcessMetrics: []metrics.ProcessMetrics{
				{ServiceName: "service-1", PID: 1234, Timestamp: time.Now()},
				{ServiceName: "service-2", PID: 5678, Timestamp: time.Now()},
			},
			expectedCount: 2,
			expectedFirst: "service-1",
		},
		{
			name:              "returns empty list",
			allProcessMetrics: []metrics.ProcessMetrics{},
			expectedCount:     0,
			expectedFirst:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			metricsProvider := &mockMetricsProvider{allProcessMetrics: tt.allProcessMetrics}
			server := grpc.NewServer(metricsProvider, &mockGetStator{})

			resp, err := server.ListProcesses(context.Background(), &emptypb.Empty{})
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Len(t, resp.Processes, tt.expectedCount)
			if tt.expectedCount > 0 {
				assert.Equal(t, tt.expectedFirst, resp.Processes[0].ServiceName)
			}
		})
	}
}

// TestServer_GetProcess verifies that GetProcess returns metrics for a specific process.
//
// Params:
//   - t: testing context for assertions
func TestServer_GetProcess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		processMetrics metrics.ProcessMetrics
		err            error
		serviceName    string
		expectError    bool
		expectedPID    int32
	}{
		{
			name:           "returns process metrics",
			processMetrics: metrics.ProcessMetrics{ServiceName: "test-service", PID: 1234, Timestamp: time.Now()},
			serviceName:    "test-service",
			expectError:    false,
			expectedPID:    1234,
		},
		{
			name:        "returns error when not found",
			err:         errors.New("process not found"),
			serviceName: "unknown-service",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			metricsProvider := &mockMetricsProvider{
				processMetrics: tt.processMetrics,
				err:            tt.err,
			}
			server := grpc.NewServer(metricsProvider, &mockGetStator{})

			resp, err := server.GetProcess(context.Background(), &daemonpb.GetProcessRequest{
				ServiceName: tt.serviceName,
			})

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.expectedPID, resp.Pid)
			}
		})
	}
}

// TestServer_GetSystemMetrics verifies that GetSystemMetrics returns system-wide metrics.
//
// Params:
//   - t: testing context for assertions
func TestServer_GetSystemMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		state         lifecycle.DaemonState
		expectedUser  uint64
		expectedTotal uint64
	}{
		{
			name: "returns system metrics",
			state: lifecycle.DaemonState{
				System: lifecycle.SystemState{
					CPU:    metrics.SystemCPU{User: 1000, System: 2000},
					Memory: metrics.SystemMemory{Total: 1024 * 1024 * 1024, Used: 512 * 1024 * 1024},
				},
				Timestamp: time.Now(),
			},
			expectedUser:  1000,
			expectedTotal: 1024 * 1024 * 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stateProvider := &mockGetStator{state: tt.state}
			server := grpc.NewServer(&mockMetricsProvider{}, stateProvider)

			resp, err := server.GetSystemMetrics(context.Background(), &emptypb.Empty{})
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tt.expectedUser, resp.Cpu.UserNs)
			assert.Equal(t, tt.expectedTotal, resp.Memory.TotalBytes)
		})
	}
}

// mockStreamStateServer mocks the gRPC stream for StreamState.
type mockStreamStateServer struct {
	daemonpb.DaemonService_StreamStateServer
	ctx     context.Context
	sent    []*daemonpb.DaemonState
	sendErr error
	mu      sync.Mutex
}

func (m *mockStreamStateServer) Context() context.Context {
	return m.ctx
}

func (m *mockStreamStateServer) Send(state *daemonpb.DaemonState) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sent = append(m.sent, state)
	return nil
}

// TestServer_StreamState verifies that StreamState sends state updates to clients.
//
// Goroutine lifecycle: Test launches cancel goroutines that are terminated when
// the specified delay completes and context is cancelled.
//
// Params:
//   - t: testing context for assertions
func TestServer_StreamState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		state           lifecycle.DaemonState
		cancelAfter     time.Duration
		expectedMinSent int
	}{
		{
			name: "streams state updates",
			state: lifecycle.DaemonState{
				Host: lifecycle.HostInfo{DaemonVersion: "1.0.0", StartTime: time.Now()},
			},
			cancelAfter:     50 * time.Millisecond,
			expectedMinSent: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stateProvider := &mockGetStator{state: tt.state}
			server := grpc.NewServer(&mockMetricsProvider{}, stateProvider)

			ctx, cancel := context.WithCancel(context.Background())
			mockStream := &mockStreamStateServer{ctx: ctx}

			// Goroutine lifecycle: Cancels context after delay, terminated when delay completes.
			go func() {
				time.Sleep(tt.cancelAfter)
				cancel()
			}()

			err := server.StreamState(&daemonpb.StreamStateRequest{}, mockStream)
			assert.ErrorIs(t, err, context.Canceled)

			mockStream.mu.Lock()
			assert.GreaterOrEqual(t, len(mockStream.sent), tt.expectedMinSent)
			mockStream.mu.Unlock()
		})
	}
}

// mockStreamProcessMetricsServer mocks the gRPC stream for StreamProcessMetrics.
type mockStreamProcessMetricsServer struct {
	daemonpb.DaemonService_StreamProcessMetricsServer
	ctx     context.Context
	sent    []*daemonpb.ProcessMetrics
	sendErr error
	mu      sync.Mutex
}

func (m *mockStreamProcessMetricsServer) Context() context.Context {
	return m.ctx
}

func (m *mockStreamProcessMetricsServer) Send(pm *daemonpb.ProcessMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sent = append(m.sent, pm)
	return nil
}

// TestServer_StreamProcessMetrics verifies that StreamProcessMetrics sends metrics updates.
//
// Goroutine lifecycle: Test launches cancel goroutines that are terminated when
// the specified delay completes and context is cancelled.
//
// Params:
//   - t: testing context for assertions
func TestServer_StreamProcessMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		processMetrics  metrics.ProcessMetrics
		serviceName     string
		cancelAfter     time.Duration
		expectedMinSent int
	}{
		{
			name:            "streams process metrics",
			processMetrics:  metrics.ProcessMetrics{ServiceName: "test-service", PID: 1234, Timestamp: time.Now()},
			serviceName:     "test-service",
			cancelAfter:     50 * time.Millisecond,
			expectedMinSent: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			metricsProvider := &mockMetricsProvider{processMetrics: tt.processMetrics}
			server := grpc.NewServer(metricsProvider, &mockGetStator{})

			ctx, cancel := context.WithCancel(context.Background())
			mockStream := &mockStreamProcessMetricsServer{ctx: ctx}

			// Goroutine lifecycle: Cancels context after delay, terminated when delay completes.
			go func() {
				time.Sleep(tt.cancelAfter)
				cancel()
			}()

			err := server.StreamProcessMetrics(
				&daemonpb.StreamProcessMetricsRequest{ServiceName: tt.serviceName},
				mockStream,
			)
			assert.ErrorIs(t, err, context.Canceled)

			mockStream.mu.Lock()
			assert.GreaterOrEqual(t, len(mockStream.sent), tt.expectedMinSent)
			mockStream.mu.Unlock()
		})
	}
}

// mockStreamSystemMetricsServer mocks the gRPC stream for StreamSystemMetrics.
type mockStreamSystemMetricsServer struct {
	daemonpb.MetricsService_StreamSystemMetricsServer
	ctx     context.Context
	sent    []*daemonpb.SystemMetrics
	sendErr error
	mu      sync.Mutex
}

func (m *mockStreamSystemMetricsServer) Context() context.Context {
	return m.ctx
}

func (m *mockStreamSystemMetricsServer) Send(sm *daemonpb.SystemMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sent = append(m.sent, sm)
	return nil
}

// TestServer_StreamSystemMetrics verifies that StreamSystemMetrics sends system metrics updates.
//
// Goroutine lifecycle: Test launches cancel goroutines that are terminated when
// the specified delay completes and context is cancelled.
//
// Params:
//   - t: testing context for assertions
func TestServer_StreamSystemMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		state           lifecycle.DaemonState
		cancelAfter     time.Duration
		expectedMinSent int
	}{
		{
			name: "streams system metrics",
			state: lifecycle.DaemonState{
				System:    lifecycle.SystemState{CPU: metrics.SystemCPU{User: 1000, System: 2000}},
				Timestamp: time.Now(),
			},
			cancelAfter:     50 * time.Millisecond,
			expectedMinSent: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stateProvider := &mockGetStator{state: tt.state}
			server := grpc.NewServer(&mockMetricsProvider{}, stateProvider)

			ctx, cancel := context.WithCancel(context.Background())
			mockStream := &mockStreamSystemMetricsServer{ctx: ctx}

			// Goroutine lifecycle: Cancels context after delay, terminated when delay completes.
			go func() {
				time.Sleep(tt.cancelAfter)
				cancel()
			}()

			err := server.StreamSystemMetrics(&daemonpb.StreamMetricsRequest{}, mockStream)
			assert.ErrorIs(t, err, context.Canceled)

			mockStream.mu.Lock()
			assert.GreaterOrEqual(t, len(mockStream.sent), tt.expectedMinSent)
			mockStream.mu.Unlock()
		})
	}
}

// mockStreamAllProcessMetricsServer mocks the gRPC stream for StreamAllProcessMetrics.
type mockStreamAllProcessMetricsServer struct {
	daemonpb.MetricsService_StreamAllProcessMetricsServer
	ctx     context.Context
	sent    []*daemonpb.ProcessMetrics
	sendErr error
	mu      sync.Mutex
}

func (m *mockStreamAllProcessMetricsServer) Context() context.Context {
	return m.ctx
}

func (m *mockStreamAllProcessMetricsServer) Send(pm *daemonpb.ProcessMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sent = append(m.sent, pm)
	return nil
}

// TestServer_StreamAllProcessMetrics verifies that StreamAllProcessMetrics sends all process metrics.
//
// Goroutine lifecycle: Test launches cancel goroutines that are terminated when
// the specified delay completes and context is cancelled.
//
// Params:
//   - t: testing context for assertions
func TestServer_StreamAllProcessMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		allProcessMetrics []metrics.ProcessMetrics
		cancelAfter       time.Duration
	}{
		{
			name: "streams all process metrics",
			allProcessMetrics: []metrics.ProcessMetrics{
				{ServiceName: "service-1", PID: 1234, Timestamp: time.Now()},
				{ServiceName: "service-2", PID: 5678, Timestamp: time.Now()},
			},
			cancelAfter: 50 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			metricsProvider := &mockMetricsProvider{allProcessMetrics: tt.allProcessMetrics}
			server := grpc.NewServer(metricsProvider, &mockGetStator{})

			ctx, cancel := context.WithCancel(context.Background())
			mockStream := &mockStreamAllProcessMetricsServer{ctx: ctx}

			// Goroutine lifecycle: Cancels context after delay, terminated when delay completes.
			go func() {
				time.Sleep(tt.cancelAfter)
				cancel()
			}()

			err := server.StreamAllProcessMetrics(&daemonpb.StreamMetricsRequest{}, mockStream)
			assert.ErrorIs(t, err, context.Canceled)
		})
	}
}
