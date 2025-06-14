package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	// Proxy Protocol v1 Signature
	ProxyProtocolV1Prefix = "PROXY "
	// Proxy Protocol v2 Signature (binary version)
	ProxyProtocolV2Prefix = "\x0D\x0A\x0D\x0A\x00\x0D\x0A\x51\x55\x49\x54\x0A"
)

// ProxyProtocolInfo contains information from the Proxy Protocol header
type ProxyProtocolInfo struct {
	SourceAddr      string
	DestAddr        string
	SourcePort      int
	DestPort        int
	Version         int    // 1 or 2
	TransportProto  string // "TCP4", "TCP6", or "UNKNOWN"
	OriginalRequest *http.Request
}

// Listener implements the Proxy Protocol support
type ProxyProtocolListener struct {
	Listener        net.Listener
	Logger          *log.Logger
	ReadTimeout     time.Duration // Timeout for reading the Proxy Protocol header
	OriginalHandler http.Handler  // Original HTTP Handler
}

// NewProxyProtocolListener creates a new listener with Proxy Protocol support
func NewProxyProtocolListener(listener net.Listener, handler http.Handler, logger *log.Logger) *ProxyProtocolListener {
	return &ProxyProtocolListener{
		Listener:        listener,
		Logger:          logger,
		ReadTimeout:     5 * time.Second, // Default timeout
		OriginalHandler: handler,
	}
}

// Accept accepts a connection and reads the Proxy Protocol header
func (l *ProxyProtocolListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	// Set timeout
	if l.ReadTimeout > 0 {
		conn.SetReadDeadline(time.Now().Add(l.ReadTimeout))
	}

	// Buffer for reading the header
	br := bufio.NewReader(conn)

	// Read first bytes (without consuming)
	peek, err := br.Peek(14) // Enough to detect the signature (v1 or v2)
	if err != nil {
		l.Logger.Printf("Error reading Proxy Protocol header: %v", err)
		return conn, nil // Accept connection normally if header cannot be read
	}

	var proxyInfo *ProxyProtocolInfo

	// Check if it's a Proxy Protocol header
	if bytes.HasPrefix(peek, []byte(ProxyProtocolV1Prefix)) {
		// Read V1 header
		proxyInfo, err = parseProxyProtocolV1(br)
	} else if bytes.HasPrefix(peek, []byte(ProxyProtocolV2Prefix)) {
		// Read V2 header
		proxyInfo, err = parseProxyProtocolV2(br)
	} else {
		// No Proxy Protocol header
		return conn, nil
	}

	if err != nil {
		l.Logger.Printf("Error parsing Proxy Protocol header: %v", err)
		return conn, nil
	}

	// Reset timeout
	conn.SetReadDeadline(time.Time{})

	// Return connection with Proxy Protocol information
	return &proxyProtocolConn{
		Conn:            conn,
		ProxyInfo:       proxyInfo,
		BufReader:       br,
		proxyRemoteAddr: &proxyProtocolAddr{proxyInfo.SourceAddr, proxyInfo.SourcePort},
		proxyLocalAddr:  &proxyProtocolAddr{proxyInfo.DestAddr, proxyInfo.DestPort},
	}, nil
}

// Close closes the listener
func (l *ProxyProtocolListener) Close() error {
	return l.Listener.Close()
}

// Addr returns the listener's address
func (l *ProxyProtocolListener) Addr() net.Addr {
	return l.Listener.Addr()
}

// Structure for a connection with Proxy Protocol information
type proxyProtocolConn struct {
	net.Conn
	ProxyInfo       *ProxyProtocolInfo
	BufReader       *bufio.Reader
	proxyRemoteAddr net.Addr
	proxyLocalAddr  net.Addr
}

// Read reads data from the reader
func (c *proxyProtocolConn) Read(b []byte) (int, error) {
	return c.BufReader.Read(b)
}

// LocalAddr returns the local address
func (c *proxyProtocolConn) LocalAddr() net.Addr {
	return c.proxyLocalAddr
}

// RemoteAddr returns the remote address
func (c *proxyProtocolConn) RemoteAddr() net.Addr {
	return c.proxyRemoteAddr
}

// Address for Proxy Protocol
type proxyProtocolAddr struct {
	address string
	port    int
}

func (a *proxyProtocolAddr) Network() string {
	return "tcp"
}

func (a *proxyProtocolAddr) String() string {
	return fmt.Sprintf("%s:%d", a.address, a.port)
}

// Parser for Proxy Protocol v1
func parseProxyProtocolV1(reader *bufio.Reader) (*ProxyProtocolInfo, error) {
	// Read a line (header ends with \r\n)
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	// Remove \r\n at the end
	line = strings.TrimSpace(line)

	// Parse header
	parts := strings.Split(line, " ")
	if len(parts) < 6 {
		return nil, fmt.Errorf("invalid Proxy Protocol v1 header: %s", line)
	}

	// Format: "PROXY TCP4/TCP6 SOURCE_IP DEST_IP SOURCE_PORT DEST_PORT"
	if parts[0] != "PROXY" {
		return nil, fmt.Errorf("invalid Proxy Protocol v1 prefix: %s", parts[0])
	}

	proto := parts[1]
	sourceAddr := parts[2]
	destAddr := parts[3]
	sourcePort := 0
	destPort := 0

	fmt.Sscanf(parts[4], "%d", &sourcePort)
	fmt.Sscanf(parts[5], "%d", &destPort)

	return &ProxyProtocolInfo{
		SourceAddr:     sourceAddr,
		DestAddr:       destAddr,
		SourcePort:     sourcePort,
		DestPort:       destPort,
		Version:        1,
		TransportProto: proto,
	}, nil
}

// Parser for Proxy Protocol v2 (binary header)
func parseProxyProtocolV2(reader *bufio.Reader) (*ProxyProtocolInfo, error) {
	// Read and discard signature (13 bytes)
	signature := make([]byte, 12)
	if _, err := reader.Read(signature); err != nil {
		return nil, err
	}

	// Read version info and command (1 byte)
	versionCmd, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	// Check version
	version := versionCmd >> 4
	if version != 2 {
		return nil, fmt.Errorf("invalid Proxy Protocol version: %d", version)
	}

	// Check command
	command := versionCmd & 0xF
	if command != 1 {
		// Local commands (COMMAND_LOCAL) or other unsupported commands
		// In this case, just skip the rest of the header
		return &ProxyProtocolInfo{
			Version:        2,
			TransportProto: "UNKNOWN",
		}, nil
	}

	// Read address family and protocol (1 byte)
	afProto, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	// Extract address family (4 highest bits)
	af := afProto >> 4

	// Read length (2 bytes)
	lenBytes := make([]byte, 2)
	if _, err := reader.Read(lenBytes); err != nil {
		return nil, err
	}
	addrLen := int(lenBytes[0])<<8 | int(lenBytes[1])

	// Read address data
	addrData := make([]byte, addrLen)
	if _, err := reader.Read(addrData); err != nil {
		return nil, err
	}

	// Parse address and ports based on address family
	var sourceAddr, destAddr string
	var sourcePort, destPort int
	var proto string

	switch af {
	case 1: // AF_INET (IPv4)
		if addrLen < 12 {
			return nil, fmt.Errorf("IPv4 address data too short: %d bytes", addrLen)
		}
		sourceAddr = fmt.Sprintf("%d.%d.%d.%d", addrData[0], addrData[1], addrData[2], addrData[3])
		destAddr = fmt.Sprintf("%d.%d.%d.%d", addrData[4], addrData[5], addrData[6], addrData[7])
		sourcePort = int(addrData[8])<<8 | int(addrData[9])
		destPort = int(addrData[10])<<8 | int(addrData[11])
		proto = "TCP4"

	case 2: // AF_INET6 (IPv6)
		if addrLen < 36 {
			return nil, fmt.Errorf("IPv6 address data too short: %d bytes", addrLen)
		}
		// Format IPv6 addresses
		srcIP := net.IP(addrData[0:16])
		dstIP := net.IP(addrData[16:32])
		sourceAddr = srcIP.String()
		destAddr = dstIP.String()
		sourcePort = int(addrData[32])<<8 | int(addrData[33])
		destPort = int(addrData[34])<<8 | int(addrData[35])
		proto = "TCP6"

	default:
		return nil, fmt.Errorf("unsupported address family: %d", af)
	}

	return &ProxyProtocolInfo{
		SourceAddr:     sourceAddr,
		DestAddr:       destAddr,
		SourcePort:     sourcePort,
		DestPort:       destPort,
		Version:        2,
		TransportProto: proto,
	}, nil
}

// ProxyProtocolMiddleware is an HTTP middleware that inserts Proxy Protocol information into the request
func ProxyProtocolMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the connection comes via Proxy Protocol
		if pc, ok := r.Context().Value("proxy_protocol_conn").(*proxyProtocolConn); ok {
			// Use remote address from Proxy Protocol
			r.RemoteAddr = pc.RemoteAddr().String()

			// Set X-Forwarded-For header if not present
			if r.Header.Get("X-Forwarded-For") == "" {
				r.Header.Set("X-Forwarded-For", pc.ProxyInfo.SourceAddr)
			}

			// Set X-Real-IP header
			r.Header.Set("X-Real-IP", pc.ProxyInfo.SourceAddr)

			// Set X-Forwarded-Proto header if not present
			if r.Header.Get("X-Forwarded-Proto") == "" {
				if r.TLS != nil {
					r.Header.Set("X-Forwarded-Proto", "https")
				} else {
					r.Header.Set("X-Forwarded-Proto", "http")
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}
