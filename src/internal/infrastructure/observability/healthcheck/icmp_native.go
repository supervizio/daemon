//go:build unix

// Package healthcheck provides infrastructure adapters for service probing.
package healthcheck

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"

	"github.com/kodflow/daemon/internal/domain/health"
)

// icmpProtocolNumber is the protocol number for ICMP.
const icmpProtocolNumber int = 1

// icmpEchoDataSize is the size of the echo request payload.
const icmpEchoDataSize int = 32

// icmpMaxPacketSize is the maximum expected ICMP packet size.
const icmpMaxPacketSize int = 1500

// nativePing sends a real ICMP echo request.
//
// Params:
//   - ctx: context for cancellation and timeout.
//   - host: target hostname or IP address.
//   - start: probe start time for latency calculation.
//
// Returns:
//   - health.CheckResult: the probe result with latency measurement.
func (p *ICMPProber) nativePing(ctx context.Context, host string, start time.Time) health.CheckResult {
	// resolve hostname to IP address
	addr, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		// resolution failed, return error
		return health.NewFailureCheckResult(
			time.Since(start),
			fmt.Sprintf("resolve failed: %s", host),
			fmt.Errorf("resolving %s: %w", host, err),
		)
	}

	// create ICMP socket (requires CAP_NET_RAW)
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		// fallback to TCP if ICMP socket creation fails
		return p.tcpPing(ctx, host, start)
	}
	defer func() { _ = conn.Close() }()

	// build echo request message
	echoData := make([]byte, icmpEchoDataSize)
	for i := range echoData {
		echoData[i] = byte(i & 0xff)
	}

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: echoData,
		},
	}

	// marshal message to bytes
	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		// marshal failed
		return health.NewFailureCheckResult(
			time.Since(start),
			"marshal failed",
			fmt.Errorf("marshaling ICMP message: %w", err),
		)
	}

	// set deadline from context or timeout
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(p.timeout)
	}
	if err := conn.SetDeadline(deadline); err != nil {
		// deadline setting failed
		return health.NewFailureCheckResult(
			time.Since(start),
			"deadline failed",
			fmt.Errorf("setting deadline: %w", err),
		)
	}

	// send echo request
	if _, err := conn.WriteTo(msgBytes, addr); err != nil {
		// send failed
		return health.NewFailureCheckResult(
			time.Since(start),
			"send failed",
			fmt.Errorf("sending ICMP echo to %s: %w", addr, err),
		)
	}

	// wait for echo reply
	reply := make([]byte, icmpMaxPacketSize)
	n, _, err := conn.ReadFrom(reply)
	if err != nil {
		// receive failed (timeout or network error)
		return health.NewFailureCheckResult(
			time.Since(start),
			"receive failed",
			fmt.Errorf("receiving ICMP reply from %s: %w", addr, err),
		)
	}

	// parse reply message
	rm, err := icmp.ParseMessage(icmpProtocolNumber, reply[:n])
	if err != nil {
		// parse failed
		return health.NewFailureCheckResult(
			time.Since(start),
			"parse failed",
			fmt.Errorf("parsing ICMP reply: %w", err),
		)
	}

	// verify reply type
	if rm.Type != ipv4.ICMPTypeEchoReply {
		// unexpected reply type
		return health.NewFailureCheckResult(
			time.Since(start),
			fmt.Sprintf("unexpected reply type: %v", rm.Type),
			fmt.Errorf("expected echo reply, got %v", rm.Type),
		)
	}

	// success - calculate final latency
	latency := time.Since(start)
	return health.NewSuccessCheckResult(
		latency,
		fmt.Sprintf("ping %s: latency=%s (native icmp)", addr, latency),
	)
}

// detectICMPCapability checks if native ICMP is available.
//
// Returns:
//   - bool: true if ICMP socket creation succeeds, false otherwise.
func detectICMPCapability() bool {
	// try to create ICMP socket
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		// ICMP not available (no CAP_NET_RAW)
		return false
	}
	// close immediately - we only tested capability
	_ = conn.Close()
	// ICMP is available
	return true
}
