// Package grpc provides gRPC server implementation for the daemon API.
package grpc

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	daemonpb "github.com/kodflow/daemon/api/proto/v1/daemon"
	"github.com/kodflow/daemon/internal/domain/lifecycle"
	"github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/domain/process"
)

// DefaultStreamInterval is the default interval for streaming updates.
const DefaultStreamInterval time.Duration = 5 * time.Second

// ErrServerAlreadyRunning indicates the server is already running.
var ErrServerAlreadyRunning error = errors.New("server already running")

// safeInt32 converts an int to int32 with bounds checking.
//
// Params:
//   - v: integer value to convert.
//
// Returns:
//   - int32: converted value clamped to int32 bounds.
func safeInt32(v int) int32 {
	// Check upper bound.
	if v > math.MaxInt32 {
		// Return maximum int32 value.
		return math.MaxInt32
	}
	// Check lower bound.
	if v < math.MinInt32 {
		// Return minimum int32 value.
		return math.MinInt32
	}
	// Return value within bounds.
	return int32(v)
}

// streamLoop runs a ticker loop that calls emit on each tick until context is done.
//
// Params:
//   - ctx: context for cancellation.
//   - interval: duration between ticks.
//   - emit: function to generate values.
//   - send: function to send values to stream.
//
// Returns:
//   - error: if emit or send fails, or context is canceled.
func streamLoop[T any](ctx context.Context, interval time.Duration, emit func() (T, error), send func(T) error) error {
	// Send initial value immediately.
	initial, err := emit()
	// Check if emit failed.
	if err != nil {
		// Return error from emit.
		return err
	}
	// Send initial value to stream.
	if err := send(initial); err != nil {
		// Return error from send.
		return err
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Loop until context is done.
	for {
		select {
		// Check if context is canceled.
		case <-ctx.Done():
			// Return context error.
			return ctx.Err()
		// Wait for next tick.
		case <-ticker.C:
			value, err := emit()
			// Check if emit failed.
			if err != nil {
				// Skip this tick on error.
				continue
			}
			// Send value to stream.
			if err := send(value); err != nil {
				// Return error from send.
				return err
			}
		}
	}
}

// MetricsProvider provides access to current metrics.
type MetricsProvider interface {
	// GetProcessMetrics returns metrics for a specific process.
	GetProcessMetrics(serviceName string) (metrics.ProcessMetrics, error)
	// GetAllProcessMetrics returns metrics for all processes.
	GetAllProcessMetrics() []metrics.ProcessMetrics
	// Subscribe returns a channel that receives metrics updates.
	Subscribe() <-chan metrics.ProcessMetrics
	// Unsubscribe removes a subscription channel.
	Unsubscribe(ch <-chan metrics.ProcessMetrics)
}

// GetStator provides access to daemon lifecycle state.
type GetStator interface {
	// GetState returns the current daemon lifecycle.
	GetState() lifecycle.DaemonState
}

// Server implements the gRPC daemon services.
//
// Server provides gRPC endpoints for daemon control and monitoring.
// It exposes DaemonService and MetricsService with health check support.
type Server struct {
	daemonpb.UnimplementedDaemonServiceServer
	daemonpb.UnimplementedMetricsServiceServer

	grpcServer      *grpc.Server
	healthServer    *health.Server
	metricsProvider MetricsProvider
	stateProvider   GetStator
	listener        net.Listener
	mu              sync.Mutex
	running         bool
}

// NewServer creates a new gRPC server.
//
// Params:
//   - metricsProvider: provider for process metrics.
//   - stateProvider: provider for daemon lifecycle.
//
// Returns:
//   - *Server: configured gRPC server.
func NewServer(metricsProvider MetricsProvider, stateProvider GetStator) *Server {
	grpcServer := grpc.NewServer()
	healthServer := health.NewServer()

	s := &Server{
		grpcServer:      grpcServer,
		healthServer:    healthServer,
		metricsProvider: metricsProvider,
		stateProvider:   stateProvider,
	}

	// Register services.
	daemonpb.RegisterDaemonServiceServer(grpcServer, s)
	daemonpb.RegisterMetricsServiceServer(grpcServer, s)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	// Set initial health status.
	healthServer.SetServingStatus("daemon.v1.DaemonService", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("daemon.v1.MetricsService", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Return configured server.
	return s
}

// Serve starts the gRPC server on the specified address.
//
// Params:
//   - address: network address to listen on (e.g., ":50051").
//
// Returns:
//   - error: if the server fails to start.
func (s *Server) Serve(address string) error {
	s.mu.Lock()
	// Check if server is already running.
	if s.running {
		s.mu.Unlock()
		// Return sentinel error for already running server.
		return fmt.Errorf("serve: %w", ErrServerAlreadyRunning)
	}

	listener, err := net.Listen("tcp", address)
	// Check if listen failed.
	if err != nil {
		s.mu.Unlock()
		// Return wrapped error.
		return fmt.Errorf("listen: %w", err)
	}
	// Listener will be closed by GracefulStop in Stop method.
	defer func() {
		// Close listener if Serve returns early.
		if listener != nil {
			_ = listener.Close()
		}
	}()

	s.listener = listener
	s.running = true
	s.mu.Unlock()

	// Start serving gRPC requests.
	return s.grpcServer.Serve(listener)
}

// Stop gracefully stops the gRPC server.
//
// Returns:
//   - none: stops server gracefully.
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if server is not running.
	if !s.running {
		// Nothing to stop.
		return
	}

	// Mark health as not serving.
	s.healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	s.grpcServer.GracefulStop()
	s.running = false
}

// Address returns the server's listening address.
//
// Returns:
//   - string: listening address, or empty if not running.
func (s *Server) Address() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if listener exists.
	if s.listener == nil {
		// Return empty string for no listener.
		return ""
	}
	// Return listener address.
	return s.listener.Addr().String()
}

// GetState implements DaemonService.GetState.
//
// Params:
//   - ctx: request context for cancellation.
//   - req: empty request (required by gRPC interface).
//
// Returns:
//   - *daemonpb.DaemonState: current daemon state.
//   - error: context error if cancelled.
func (s *Server) GetState(ctx context.Context, req *emptypb.Empty) (*daemonpb.DaemonState, error) {
	// Check request is valid (interface requirement).
	if req == nil {
		// Handle nil request gracefully.
		return nil, nil
	}
	// Check for context cancellation.
	if ctx.Err() != nil {
		// Return context error.
		return nil, ctx.Err()
	}
	ds := s.stateProvider.GetState()
	// Return converted state.
	return s.convertDaemonState(&ds), nil
}

// StreamState implements DaemonService.StreamState.
//
// Params:
//   - req: stream request with interval.
//   - stream: bidirectional stream for state updates.
//
// Returns:
//   - error: if stream fails.
func (s *Server) StreamState(req *daemonpb.StreamStateRequest, stream daemonpb.DaemonService_StreamStateServer) error {
	interval := DefaultStreamInterval
	// Check if custom interval provided.
	if req.Interval != nil {
		// Use custom interval.
		interval = req.Interval.AsDuration()
	}

	// Stream daemon state updates.
	return streamLoop(
		stream.Context(),
		interval,
		func() (*daemonpb.DaemonState, error) {
			ds := s.stateProvider.GetState()
			// Return converted state.
			return s.convertDaemonState(&ds), nil
		},
		stream.Send,
	)
}

// ListProcesses implements DaemonService.ListProcesses.
//
// Params:
//   - ctx: request context for cancellation.
//   - req: empty request (required by gRPC interface).
//
// Returns:
//   - *daemonpb.ListProcessesResponse: list of all processes.
//   - error: context error if cancelled.
func (s *Server) ListProcesses(ctx context.Context, req *emptypb.Empty) (*daemonpb.ListProcessesResponse, error) {
	// Check request is valid (interface requirement).
	if req == nil {
		// Handle nil request gracefully.
		return nil, nil
	}
	// Check for context cancellation.
	if ctx.Err() != nil {
		// Return context error.
		return nil, ctx.Err()
	}
	allMetrics := s.metricsProvider.GetAllProcessMetrics()
	processes := make([]*daemonpb.ProcessMetrics, 0, len(allMetrics))

	// Convert all process metrics.
	for i := range allMetrics {
		processes = append(processes, s.convertProcessMetrics(&allMetrics[i]))
	}

	// Return process list.
	return &daemonpb.ListProcessesResponse{
		Processes: processes,
	}, nil
}

// GetProcess implements DaemonService.GetProcess.
//
// Params:
//   - ctx: request context for cancellation.
//   - req: request with service name.
//
// Returns:
//   - *daemonpb.ProcessMetrics: metrics for requested process.
//   - error: if process not found or context cancelled.
func (s *Server) GetProcess(ctx context.Context, req *daemonpb.GetProcessRequest) (*daemonpb.ProcessMetrics, error) {
	// Check for context cancellation.
	if ctx.Err() != nil {
		// Return context error.
		return nil, ctx.Err()
	}
	m, err := s.metricsProvider.GetProcessMetrics(req.ServiceName)
	// Check if metrics retrieval failed.
	if err != nil {
		// Return wrapped error.
		return nil, fmt.Errorf("get process: %w", err)
	}
	// Return converted metrics.
	return s.convertProcessMetrics(&m), nil
}

// StreamProcessMetrics implements DaemonService.StreamProcessMetrics and MetricsService.StreamProcessMetrics.
//
// Params:
//   - req: stream request with service name and interval.
//   - stream: bidirectional stream for metrics updates.
//
// Returns:
//   - error: if stream fails.
func (s *Server) StreamProcessMetrics(req *daemonpb.StreamProcessMetricsRequest, stream daemonpb.DaemonService_StreamProcessMetricsServer) error {
	interval := DefaultStreamInterval
	// Check if custom interval provided.
	if req.Interval != nil {
		// Use custom interval.
		interval = req.Interval.AsDuration()
	}

	// Stream process metrics updates.
	return streamLoop(
		stream.Context(),
		interval,
		func() (*daemonpb.ProcessMetrics, error) {
			m, err := s.metricsProvider.GetProcessMetrics(req.ServiceName)
			// Check if retrieval failed.
			if err != nil {
				// Return wrapped error.
				return nil, fmt.Errorf("get process metrics: %w", err)
			}
			// Return converted metrics.
			return s.convertProcessMetrics(&m), nil
		},
		stream.Send,
	)
}

// GetSystemMetrics implements MetricsService.GetSystemMetrics.
//
// Params:
//   - ctx: request context for cancellation.
//   - req: empty request (required by gRPC interface).
//
// Returns:
//   - *daemonpb.SystemMetrics: current system metrics.
//   - error: context error if cancelled.
func (s *Server) GetSystemMetrics(ctx context.Context, req *emptypb.Empty) (*daemonpb.SystemMetrics, error) {
	// Check request is valid (interface requirement).
	if req == nil {
		// Handle nil request gracefully.
		return nil, nil
	}
	// Check for context cancellation.
	if ctx.Err() != nil {
		// Return context error.
		return nil, ctx.Err()
	}
	ds := s.stateProvider.GetState()
	// Return converted system metrics.
	return s.convertSystemMetrics(&ds), nil
}

// StreamSystemMetrics implements MetricsService.StreamSystemMetrics.
//
// Params:
//   - req: stream request with interval.
//   - stream: bidirectional stream for system metrics updates.
//
// Returns:
//   - error: if stream fails.
func (s *Server) StreamSystemMetrics(req *daemonpb.StreamMetricsRequest, stream daemonpb.MetricsService_StreamSystemMetricsServer) error {
	interval := DefaultStreamInterval
	// Check if custom interval provided.
	if req.Interval != nil {
		// Use custom interval.
		interval = req.Interval.AsDuration()
	}

	// Stream system metrics updates.
	return streamLoop(
		stream.Context(),
		interval,
		func() (*daemonpb.SystemMetrics, error) {
			ds := s.stateProvider.GetState()
			// Return converted system metrics.
			return s.convertSystemMetrics(&ds), nil
		},
		stream.Send,
	)
}

// StreamAllProcessMetrics implements MetricsService.StreamAllProcessMetrics.
//
// Params:
//   - req: stream request with interval.
//   - stream: bidirectional stream for all process metrics.
//
// Returns:
//   - error: if stream fails.
func (s *Server) StreamAllProcessMetrics(req *daemonpb.StreamMetricsRequest, stream daemonpb.MetricsService_StreamAllProcessMetricsServer) error {
	interval := DefaultStreamInterval
	// Check if custom interval provided.
	if req.Interval != nil {
		// Use custom interval.
		interval = req.Interval.AsDuration()
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Loop until context is done.
	for {
		select {
		// Check if context is canceled.
		case <-stream.Context().Done():
			// Return context error.
			return stream.Context().Err()
		// Wait for next tick.
		case <-ticker.C:
			allMetrics := s.metricsProvider.GetAllProcessMetrics()
			// Send metrics for all processes.
			for i := range allMetrics {
				// Send individual process metrics.
				if err := stream.Send(s.convertProcessMetrics(&allMetrics[i])); err != nil {
					// Return error from send.
					return err
				}
			}
		}
	}
}

// convertDaemonState converts domain state to protobuf.
//
// Params:
//   - ds: domain daemon state.
//
// Returns:
//   - *daemonpb.DaemonState: protobuf daemon state.
func (s *Server) convertDaemonState(ds *lifecycle.DaemonState) *daemonpb.DaemonState {
	processes := make([]*daemonpb.ProcessMetrics, 0, len(ds.Processes))
	// Convert all process metrics.
	for i := range ds.Processes {
		processes = append(processes, s.convertProcessMetrics(&ds.Processes[i]))
	}

	// Compute overall healthy status.
	healthy := ds.HealthyProcessCount() == ds.ProcessCount() && ds.ProcessCount() > 0

	// Return protobuf state.
	return &daemonpb.DaemonState{
		Version:    ds.Host.DaemonVersion,
		StartTime:  timestamppb.New(ds.Host.StartTime),
		Uptime:     durationpb.New(ds.Host.Uptime()),
		Healthy:    healthy,
		Processes:  processes,
		System:     s.convertSystemMetrics(ds),
		Host:       s.convertHostInfo(&ds.Host),
		Kubernetes: s.convertKubernetesInfo(ds.Kubernetes),
	}
}

// convertProcessMetrics converts domain metrics to protobuf.
//
// Params:
//   - m: domain process metrics.
//
// Returns:
//   - *daemonpb.ProcessMetrics: protobuf process metrics.
func (s *Server) convertProcessMetrics(m *metrics.ProcessMetrics) *daemonpb.ProcessMetrics {
	// Return protobuf metrics.
	return &daemonpb.ProcessMetrics{
		ServiceName:  m.ServiceName,
		Pid:          safeInt32(m.PID),
		State:        s.convertProcessState(m.State),
		Healthy:      m.Healthy,
		Cpu:          s.convertProcessCPU(&m.CPU),
		Memory:       s.convertProcessMemory(&m.Memory),
		StartTime:    timestamppb.New(m.StartTime),
		Uptime:       durationpb.New(m.Uptime),
		RestartCount: safeInt32(m.RestartCount),
		LastError:    m.LastError,
		Timestamp:    timestamppb.New(m.Timestamp),
	}
}

// convertProcessState converts domain process state to protobuf.
//
// Params:
//   - ps: domain process state.
//
// Returns:
//   - daemonpb.ProcessState: protobuf process state.
func (s *Server) convertProcessState(ps process.State) daemonpb.ProcessState {
	// Match domain state to protobuf state.
	switch ps {
	// Process is stopped.
	case process.StateStopped:
		// Return stopped state.
		return daemonpb.ProcessState_PROCESS_STATE_STOPPED
	// Process is starting.
	case process.StateStarting:
		// Return starting state.
		return daemonpb.ProcessState_PROCESS_STATE_STARTING
	// Process is running.
	case process.StateRunning:
		// Return running state.
		return daemonpb.ProcessState_PROCESS_STATE_RUNNING
	// Process is stopping.
	case process.StateStopping:
		// Return stopping state.
		return daemonpb.ProcessState_PROCESS_STATE_STOPPING
	// Process has failed.
	case process.StateFailed:
		// Return failed state.
		return daemonpb.ProcessState_PROCESS_STATE_FAILED
	// Unknown state.
	default:
		// Return unspecified state.
		return daemonpb.ProcessState_PROCESS_STATE_UNSPECIFIED
	}
}

// convertProcessCPU converts domain CPU metrics to protobuf.
//
// Params:
//   - cpu: domain process CPU metrics.
//
// Returns:
//   - *daemonpb.ProcessCPU: protobuf CPU metrics.
func (s *Server) convertProcessCPU(cpu *metrics.ProcessCPU) *daemonpb.ProcessCPU {
	// Return protobuf CPU metrics.
	return &daemonpb.ProcessCPU{
		UserTimeNs:   cpu.User,
		SystemTimeNs: cpu.System,
		TotalTimeNs:  cpu.User + cpu.System,
	}
}

// convertProcessMemory converts domain memory metrics to protobuf.
//
// Params:
//   - mem: domain process memory metrics.
//
// Returns:
//   - *daemonpb.ProcessMemory: protobuf memory metrics.
func (s *Server) convertProcessMemory(mem *metrics.ProcessMemory) *daemonpb.ProcessMemory {
	// Return protobuf memory metrics.
	return &daemonpb.ProcessMemory{
		RssBytes:    mem.RSS,
		VmsBytes:    mem.VMS,
		SwapBytes:   mem.Swap,
		SharedBytes: mem.Shared,
		DataBytes:   mem.Data,
		StackBytes:  mem.Stack,
	}
}

// convertSystemMetrics converts daemon state to system metrics protobuf.
//
// Params:
//   - ds: domain daemon state.
//
// Returns:
//   - *daemonpb.SystemMetrics: protobuf system metrics.
func (s *Server) convertSystemMetrics(ds *lifecycle.DaemonState) *daemonpb.SystemMetrics {
	// Return protobuf system metrics.
	return &daemonpb.SystemMetrics{
		Cpu: &daemonpb.SystemCPU{
			UserNs:    ds.System.CPU.User,
			NiceNs:    ds.System.CPU.Nice,
			SystemNs:  ds.System.CPU.System,
			IdleNs:    ds.System.CPU.Idle,
			IowaitNs:  ds.System.CPU.IOWait,
			IrqNs:     ds.System.CPU.IRQ,
			SoftirqNs: ds.System.CPU.SoftIRQ,
			StealNs:   ds.System.CPU.Steal,
		},
		Memory: &daemonpb.SystemMemory{
			TotalBytes:     ds.System.Memory.Total,
			AvailableBytes: ds.System.Memory.Available,
			UsedBytes:      ds.System.Memory.Used,
			FreeBytes:      ds.System.Memory.Free,
			BuffersBytes:   ds.System.Memory.Buffers,
			CachedBytes:    ds.System.Memory.Cached,
			SharedBytes:    ds.System.Memory.Shared,
			SwapTotalBytes: ds.System.Memory.SwapTotal,
			SwapUsedBytes:  ds.System.Memory.SwapUsed,
			SwapFreeBytes:  ds.System.Memory.SwapFree,
		},
		Timestamp: timestamppb.New(ds.Timestamp),
	}
}

// convertHostInfo converts domain host info to protobuf.
//
// Params:
//   - host: domain host information.
//
// Returns:
//   - *daemonpb.HostInfo: protobuf host information.
func (s *Server) convertHostInfo(host *lifecycle.HostInfo) *daemonpb.HostInfo {
	// Return protobuf host info.
	return &daemonpb.HostInfo{
		Hostname: host.Hostname,
		Os:       host.OS,
		Arch:     host.Arch,
	}
}

// convertKubernetesInfo converts domain k8s state to protobuf.
//
// Params:
//   - k8s: domain Kubernetes state, may be nil.
//
// Returns:
//   - *daemonpb.KubernetesInfo: protobuf k8s info, or nil if not in k8s.
func (s *Server) convertKubernetesInfo(k8s *lifecycle.KubernetesState) *daemonpb.KubernetesInfo {
	// Check if running in Kubernetes.
	if k8s == nil || k8s.PodName == "" {
		// Return nil for non-Kubernetes environment.
		return nil
	}
	// Return protobuf Kubernetes info.
	return &daemonpb.KubernetesInfo{
		PodName:   k8s.PodName,
		Namespace: k8s.Namespace,
		NodeName:  k8s.NodeName,
	}
}
