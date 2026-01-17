// Package health provides health monitoring for services.
package health

// ProbeType identifies the type of health probe.
type ProbeType string

const (
	// ProbeTCP is a TCP connection probe.
	ProbeTCP ProbeType = "tcp"
	// ProbeUDP is a UDP probe.
	ProbeUDP ProbeType = "udp"
	// ProbeHTTP is an HTTP request probe.
	ProbeHTTP ProbeType = "http"
	// ProbeGRPC is a gRPC health check probe.
	ProbeGRPC ProbeType = "grpc"
	// ProbeExec is a command execution probe.
	ProbeExec ProbeType = "exec"
	// ProbeICMP is an ICMP ping probe.
	ProbeICMP ProbeType = "icmp"
)
