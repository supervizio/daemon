//go:build unix

// Package healthcheck provides infrastructure adapters for service probing.
package healthcheck

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"

	"github.com/kodflow/daemon/internal/domain/health"
)

// ICMP constants.
const (
	// icmpProtocolNumber is the protocol number for ICMP.
	icmpProtocolNumber int = 1

	// icmpEchoDataSize is the size of the echo request payload.
	icmpEchoDataSize int = 32

	// icmpMaxPacketSize is the maximum expected ICMP packet size.
	icmpMaxPacketSize int = 1500

	// icmpByteMask masks a value to 8 bits (single byte).
	icmpByteMask int = 0xff

	// icmpPIDMask masks process ID to 16 bits for ICMP echo ID.
	icmpPIDMask int = 0xffff
)

// errUnexpectedReplyType is returned when ICMP reply type is not EchoReply.
var errUnexpectedReplyType error = errors.New("unexpected ICMP reply type")

// packetConn abstracts ICMP packet connection operations.
type packetConn interface {
	WriteTo(b []byte, dst net.Addr) (int, error)
	SetDeadline(t time.Time) error
	ReadFrom(b []byte) (int, net.Addr, error)
}

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
	// Resolve hostname to IP address.
	addr, err := net.ResolveIPAddr("ip4", host)
	// Check for hostname resolution error.
	if err != nil {
		// Return failure result for resolution error.
		return health.NewFailureCheckResult(
			time.Since(start),
			fmt.Sprintf("resolve failed: %s", host),
			fmt.Errorf("resolving %s: %w", host, err),
		)
	}

	// Create ICMP socket (requires CAP_NET_RAW).
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	// Check for socket creation error.
	if err != nil {
		// Return TCP fallback if ICMP unavailable.
		return p.tcpPing(ctx, host, start)
	}
	defer func() { _ = conn.Close() }()

	// Send echo request and receive reply.
	if err := p.sendAndReceiveICMP(ctx, conn, addr); err != nil {
		// Return failure result for ICMP error.
		return health.NewFailureCheckResult(time.Since(start), err.Error(), err)
	}

	// Success - calculate final latency.
	latency := time.Since(start)
	// Return success result with latency measurement.
	return health.NewSuccessCheckResult(
		latency,
		fmt.Sprintf("ping %s: latency=%s (native icmp)", addr, latency),
	)
}

// sendAndReceiveICMP sends an ICMP echo and waits for reply.
//
// Params:
//   - ctx: context for deadline.
//   - conn: ICMP packet connection.
//   - addr: target IP address.
//
// Returns:
//   - error: any error during send or receive.
func (p *ICMPProber) sendAndReceiveICMP(ctx context.Context, conn *icmp.PacketConn, addr *net.IPAddr) error {
	// Build and send echo request.
	if err := p.sendEchoRequest(ctx, conn, addr); err != nil {
		// Return error from send.
		return err
	}

	// Receive and validate echo reply.
	return p.receiveEchoReply(conn)
}

// sendEchoRequest builds and sends an ICMP echo request.
//
// Params:
//   - ctx: context for deadline.
//   - conn: ICMP packet connection.
//   - addr: target IP address.
//
// Returns:
//   - error: any error during send.
func (p *ICMPProber) sendEchoRequest(ctx context.Context, conn packetConn, addr *net.IPAddr) error {
	// Build echo request message.
	msg := p.buildEchoMessage()

	// Marshal message to bytes.
	msgBytes, err := msg.Marshal(nil)
	// Check for marshal error.
	if err != nil {
		// Return error with marshal context.
		return fmt.Errorf("marshal failed: %w", err)
	}

	// Set deadline from context or timeout.
	if err := p.setConnectionDeadline(ctx, conn); err != nil {
		// Return deadline error.
		return err
	}

	// Send echo request.
	if _, err := conn.WriteTo(msgBytes, addr); err != nil {
		// Return error with send context.
		return fmt.Errorf("send failed: %w", err)
	}

	// Return nil on successful send.
	return nil
}

// buildEchoMessage creates an ICMP echo request message.
//
// Returns:
//   - icmp.Message: the echo request message.
func (p *ICMPProber) buildEchoMessage() icmp.Message {
	var echoData [icmpEchoDataSize]byte
	// Fill echo data with sequential bytes.
	for i := range echoData {
		echoData[i] = byte(i & icmpByteMask)
	}

	// Return constructed ICMP echo message.
	return icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & icmpPIDMask,
			Seq:  1,
			Data: echoData[:],
		},
	}
}

// setConnectionDeadline sets the deadline on the ICMP connection.
//
// Params:
//   - ctx: context for deadline extraction.
//   - conn: ICMP packet connection.
//
// Returns:
//   - error: any error during deadline setting.
func (p *ICMPProber) setConnectionDeadline(ctx context.Context, conn packetConn) error {
	deadline, ok := ctx.Deadline()
	// Use timeout if context has no deadline.
	if !ok {
		deadline = time.Now().Add(p.timeout)
	}
	// Check for deadline setting error.
	if err := conn.SetDeadline(deadline); err != nil {
		// Return error with deadline context.
		return fmt.Errorf("deadline failed: %w", err)
	}
	// Return nil on success.
	return nil
}

// receiveEchoReply waits for and validates an ICMP echo reply.
//
// Params:
//   - conn: ICMP packet connection.
//
// Returns:
//   - error: any error during receive or validation.
func (p *ICMPProber) receiveEchoReply(conn packetConn) error {
	// Wait for echo reply.
	var reply [icmpMaxPacketSize]byte
	n, _, err := conn.ReadFrom(reply[:])
	// Check for receive error.
	if err != nil {
		// Return error with receive context.
		return fmt.Errorf("receive failed: %w", err)
	}

	// Parse reply message.
	rm, err := icmp.ParseMessage(icmpProtocolNumber, reply[:n])
	// Check for parse error.
	if err != nil {
		// Return error with parse context.
		return fmt.Errorf("parse failed: %w", err)
	}

	// Verify reply type.
	if rm.Type != ipv4.ICMPTypeEchoReply {
		// Return error for unexpected reply type.
		return fmt.Errorf("unexpected reply type %v: %w", rm.Type, errUnexpectedReplyType)
	}

	// Return nil on success.
	return nil
}

// detectICMPCapability checks if native ICMP is available.
//
// Returns:
//   - bool: true if ICMP socket creation succeeds, false otherwise.
func detectICMPCapability() bool {
	// try to create ICMP socket
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	// Check for socket creation error.
	if err != nil {
		// Return false when ICMP is not available.
		return false
	}
	// close immediately - we only tested capability
	_ = conn.Close()
	// Return true when ICMP is available.
	return true
}
