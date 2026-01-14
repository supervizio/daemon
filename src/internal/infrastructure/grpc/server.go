// Package grpc provides gRPC server implementation for the daemon API.
package grpc

import (
	"context"
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
	"github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/domain/state"
)

// DefaultStreamInterval is the default interval for streaming updates.
const DefaultStreamInterval = 5 * time.Second

// safeInt32 converts an int to int32 with bounds checking.
func safeInt32(v int) int32 {
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	if v < math.MinInt32 {
		return math.MinInt32
	}
	return int32(v)
}

// streamLoop runs a ticker loop that calls emit on each tick until context is done.
func streamLoop[T any](ctx context.Context, interval time.Duration, emit func() (T, error), send func(T) error) error {
	// Send initial value immediately.
	initial, err := emit()
	if err != nil {
		return err
	}
	if err := send(initial); err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			value, err := emit()
			if err != nil {
				continue // Skip this tick on error.
			}
			if err := send(value); err != nil {
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

// StateProvider provides access to daemon state.
type StateProvider interface {
	// GetState returns the current daemon state.
	GetState() state.DaemonState
}

// Server implements the gRPC daemon services.
type Server struct {
	daemonpb.UnimplementedDaemonServiceServer
	daemonpb.UnimplementedMetricsServiceServer

	grpcServer      *grpc.Server
	healthServer    *health.Server
	metricsProvider MetricsProvider
	stateProvider   StateProvider
	listener        net.Listener
	mu              sync.Mutex
	running         bool
}

// NewServer creates a new gRPC server.
//
// Params:
//   - metricsProvider: provider for process metrics.
//   - stateProvider: provider for daemon state.
//
// Returns:
//   - *Server: configured gRPC server.
func NewServer(metricsProvider MetricsProvider, stateProvider StateProvider) *Server {
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
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server already running")
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("listen: %w", err)
	}

	s.listener = listener
	s.running = true
	s.mu.Unlock()

	return s.grpcServer.Serve(listener)
}

// Stop gracefully stops the gRPC server.
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	s.grpcServer.GracefulStop()
	s.running = false
}

// Address returns the server's listening address.
// Returns empty string if server is not running.
func (s *Server) Address() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}

// GetState implements DaemonService.GetState.
func (s *Server) GetState(_ context.Context, _ *emptypb.Empty) (*daemonpb.DaemonState, error) {
	ds := s.stateProvider.GetState()
	return s.convertDaemonState(&ds), nil
}

// StreamState implements DaemonService.StreamState.
func (s *Server) StreamState(req *daemonpb.StreamStateRequest, stream daemonpb.DaemonService_StreamStateServer) error {
	interval := DefaultStreamInterval
	if req.Interval != nil {
		interval = req.Interval.AsDuration()
	}

	return streamLoop(
		stream.Context(),
		interval,
		func() (*daemonpb.DaemonState, error) {
			ds := s.stateProvider.GetState()
			return s.convertDaemonState(&ds), nil
		},
		stream.Send,
	)
}

// ListProcesses implements DaemonService.ListProcesses.
func (s *Server) ListProcesses(_ context.Context, _ *emptypb.Empty) (*daemonpb.ListProcessesResponse, error) {
	allMetrics := s.metricsProvider.GetAllProcessMetrics()
	processes := make([]*daemonpb.ProcessMetrics, 0, len(allMetrics))

	for i := range allMetrics {
		processes = append(processes, s.convertProcessMetrics(&allMetrics[i]))
	}

	return &daemonpb.ListProcessesResponse{
		Processes: processes,
	}, nil
}

// GetProcess implements DaemonService.GetProcess.
func (s *Server) GetProcess(_ context.Context, req *daemonpb.GetProcessRequest) (*daemonpb.ProcessMetrics, error) {
	m, err := s.metricsProvider.GetProcessMetrics(req.ServiceName)
	if err != nil {
		return nil, err
	}
	return s.convertProcessMetrics(&m), nil
}

// StreamProcessMetrics implements DaemonService.StreamProcessMetrics and MetricsService.StreamProcessMetrics.
func (s *Server) StreamProcessMetrics(req *daemonpb.StreamProcessMetricsRequest, stream daemonpb.DaemonService_StreamProcessMetricsServer) error {
	interval := DefaultStreamInterval
	if req.Interval != nil {
		interval = req.Interval.AsDuration()
	}

	return streamLoop(
		stream.Context(),
		interval,
		func() (*daemonpb.ProcessMetrics, error) {
			m, err := s.metricsProvider.GetProcessMetrics(req.ServiceName)
			if err != nil {
				return nil, err
			}
			return s.convertProcessMetrics(&m), nil
		},
		stream.Send,
	)
}

// GetSystemMetrics implements MetricsService.GetSystemMetrics.
func (s *Server) GetSystemMetrics(_ context.Context, _ *emptypb.Empty) (*daemonpb.SystemMetrics, error) {
	ds := s.stateProvider.GetState()
	return s.convertSystemMetrics(&ds), nil
}

// StreamSystemMetrics implements MetricsService.StreamSystemMetrics.
func (s *Server) StreamSystemMetrics(req *daemonpb.StreamMetricsRequest, stream daemonpb.MetricsService_StreamSystemMetricsServer) error {
	interval := DefaultStreamInterval
	if req.Interval != nil {
		interval = req.Interval.AsDuration()
	}

	return streamLoop(
		stream.Context(),
		interval,
		func() (*daemonpb.SystemMetrics, error) {
			ds := s.stateProvider.GetState()
			return s.convertSystemMetrics(&ds), nil
		},
		stream.Send,
	)
}

// StreamAllProcessMetrics implements MetricsService.StreamAllProcessMetrics.
func (s *Server) StreamAllProcessMetrics(req *daemonpb.StreamMetricsRequest, stream daemonpb.MetricsService_StreamAllProcessMetricsServer) error {
	interval := DefaultStreamInterval
	if req.Interval != nil {
		interval = req.Interval.AsDuration()
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case <-ticker.C:
			allMetrics := s.metricsProvider.GetAllProcessMetrics()
			for i := range allMetrics {
				if err := stream.Send(s.convertProcessMetrics(&allMetrics[i])); err != nil {
					return err
				}
			}
		}
	}
}

// convertDaemonState converts domain state to protobuf.
func (s *Server) convertDaemonState(ds *state.DaemonState) *daemonpb.DaemonState {
	processes := make([]*daemonpb.ProcessMetrics, 0, len(ds.Processes))
	for i := range ds.Processes {
		processes = append(processes, s.convertProcessMetrics(&ds.Processes[i]))
	}

	// Compute overall healthy status.
	healthy := ds.HealthyProcessCount() == ds.ProcessCount() && ds.ProcessCount() > 0

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
func (s *Server) convertProcessMetrics(m *metrics.ProcessMetrics) *daemonpb.ProcessMetrics {
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
func (s *Server) convertProcessState(ps process.State) daemonpb.ProcessState {
	switch ps {
	case process.StateStopped:
		return daemonpb.ProcessState_PROCESS_STATE_STOPPED
	case process.StateStarting:
		return daemonpb.ProcessState_PROCESS_STATE_STARTING
	case process.StateRunning:
		return daemonpb.ProcessState_PROCESS_STATE_RUNNING
	case process.StateStopping:
		return daemonpb.ProcessState_PROCESS_STATE_STOPPING
	case process.StateFailed:
		return daemonpb.ProcessState_PROCESS_STATE_FAILED
	default:
		return daemonpb.ProcessState_PROCESS_STATE_UNSPECIFIED
	}
}

// convertProcessCPU converts domain CPU metrics to protobuf.
func (s *Server) convertProcessCPU(cpu *metrics.ProcessCPU) *daemonpb.ProcessCPU {
	return &daemonpb.ProcessCPU{
		UserTimeNs:   cpu.User,
		SystemTimeNs: cpu.System,
		TotalTimeNs:  cpu.User + cpu.System,
	}
}

// convertProcessMemory converts domain memory metrics to protobuf.
func (s *Server) convertProcessMemory(mem *metrics.ProcessMemory) *daemonpb.ProcessMemory {
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
func (s *Server) convertSystemMetrics(ds *state.DaemonState) *daemonpb.SystemMetrics {
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
func (s *Server) convertHostInfo(host *state.HostInfo) *daemonpb.HostInfo {
	return &daemonpb.HostInfo{
		Hostname: host.Hostname,
		Os:       host.OS,
		Arch:     host.Arch,
	}
}

// convertKubernetesInfo converts domain k8s state to protobuf.
func (s *Server) convertKubernetesInfo(k8s *state.KubernetesState) *daemonpb.KubernetesInfo {
	if k8s == nil || k8s.PodName == "" {
		return nil
	}
	return &daemonpb.KubernetesInfo{
		PodName:   k8s.PodName,
		Namespace: k8s.Namespace,
		NodeName:  k8s.NodeName,
	}
}
