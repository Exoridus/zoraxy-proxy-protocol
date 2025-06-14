package main

import (
	"bytes"
	"net"
	"testing"
)

// Test Proxy Protocol v2 detection and parsing with HAProxy format
func TestProxyProtocolV2HAProxyFormat(t *testing.T) {
	// Create a sample Proxy Protocol v2 header like HAProxy would send
	// Using RFC 5737 documentation IP addresses for testing
	// This simulates: 192.0.2.100:45678 -> 198.51.100.50:443
	sourceIP := net.ParseIP("192.0.2.100") // RFC 5737 TEST-NET-1
	destIP := net.ParseIP("198.51.100.50") // RFC 5737 TEST-NET-2
	sourcePort := uint16(45678)
	destPort := uint16(443)

	// Build Proxy Protocol v2 header manually
	var buffer bytes.Buffer

	// 1. Signature (12 bytes)
	buffer.Write([]byte(ProxyProtocolV2Prefix))

	// 2. Version and Command (1 byte) - Version 2, Command PROXY
	buffer.WriteByte(0x21) // 0010 0001 (version 2, command 1)

	// 3. Address Family and Protocol (1 byte) - IPv4 TCP
	buffer.WriteByte(0x11) // 0001 0001 (AF_INET, STREAM)

	// 4. Length (2 bytes) - 12 bytes for IPv4 addresses + ports
	buffer.WriteByte(0x00)
	buffer.WriteByte(0x0C) // 12 bytes

	// 5. Addresses and ports (12 bytes for IPv4)
	buffer.Write(sourceIP.To4())              // Source IP (4 bytes)
	buffer.Write(destIP.To4())                // Dest IP (4 bytes)
	buffer.WriteByte(byte(sourcePort >> 8))   // Source port high byte
	buffer.WriteByte(byte(sourcePort & 0xFF)) // Source port low byte
	buffer.WriteByte(byte(destPort >> 8))     // Dest port high byte
	buffer.WriteByte(byte(destPort & 0xFF))   // Dest port low byte

	// 6. Add some TLS handshake data (ClientHello)
	tlsClientHello := []byte{
		0x16, 0x03, 0x01, 0x00, 0xF4, // TLS Handshake record
		0x01, 0x00, 0x00, 0xF0, // ClientHello message
		0x03, 0x03, // TLS version 1.2
		// Random bytes would follow...
	}
	buffer.Write(tlsClientHello)

	testData := buffer.Bytes()

	t.Logf("Created test data with %d bytes", len(testData))

	// Test detection
	detected := detectProxyProtocol(testData)
	if !detected {
		t.Fatalf("Proxy Protocol v2 not detected!")
	}
	t.Logf("✅ Proxy Protocol v2 detected successfully")

	// Test parsing
	processedData, proxyInfo, err := processProxyProtocolData(testData)
	if err != nil {
		t.Fatalf("Parsing error: %v", err)
	}

	if proxyInfo == nil {
		t.Fatalf("No proxy info returned")
	}

	t.Logf("✅ Proxy Protocol v2 parsed successfully")
	t.Logf("   Source: %s:%d", proxyInfo.SourceAddr, proxyInfo.SourcePort)
	t.Logf("   Dest: %s:%d", proxyInfo.DestAddr, proxyInfo.DestPort)
	t.Logf("   Version: %d", proxyInfo.Version)
	t.Logf("   Transport: %s", proxyInfo.TransportProto)
	t.Logf("   Remaining data: %d bytes", len(processedData))

	// Verify the values
	if proxyInfo.SourceAddr != "192.0.2.100" {
		t.Errorf("Expected source IP 192.0.2.100, got %s", proxyInfo.SourceAddr)
	}

	if proxyInfo.SourcePort != 45678 {
		t.Errorf("Expected source port 45678, got %d", proxyInfo.SourcePort)
	}

	if proxyInfo.DestAddr != "198.51.100.50" {
		t.Errorf("Expected dest IP 198.51.100.50, got %s", proxyInfo.DestAddr)
	}

	if proxyInfo.DestPort != 443 {
		t.Errorf("Expected dest port 443, got %d", proxyInfo.DestPort)
	}

	if proxyInfo.Version != 2 {
		t.Errorf("Expected version 2, got %d", proxyInfo.Version)
	}

	if proxyInfo.TransportProto != "TCP4" {
		t.Errorf("Expected transport TCP4, got %s", proxyInfo.TransportProto)
	}

	// Check that the remaining data is TLS
	if len(processedData) == 0 {
		t.Error("No remaining data after proxy protocol header removal")
	} else if processedData[0] != 0x16 {
		t.Errorf("Expected TLS handshake (0x16), got 0x%02x", processedData[0])
	} else {
		t.Logf("✅ Remaining data is TLS handshake")
	}
}

// Test with invalid data to ensure no false positives
func TestProxyProtocolDetectionFalsePositives(t *testing.T) {
	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "Plain HTTP request",
			data: []byte("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"),
		},
		{
			name: "TLS handshake without proxy protocol",
			data: []byte{0x16, 0x03, 0x01, 0x00, 0xF4, 0x01, 0x00, 0x00, 0xF0},
		},
		{
			name: "Random binary data",
			data: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
		},
		{
			name: "Empty data",
			data: []byte{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			detected := detectProxyProtocol(tc.data)
			if detected {
				t.Errorf("False positive: %s was detected as proxy protocol", tc.name)
			}
		})
	}
}
