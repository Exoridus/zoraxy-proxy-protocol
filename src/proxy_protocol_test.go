package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
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

// Test Proxy Protocol v1 detection and parsing
func TestProxyProtocolV1Format(t *testing.T) {
	// Create a sample Proxy Protocol v1 header
	// Format: "PROXY TCP4 192.0.2.100 198.51.100.50 45678 443\r\n"
	proxyHeader := "PROXY TCP4 192.0.2.100 198.51.100.50 45678 443\r\n"

	// Add some HTTP data
	httpData := "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"

	testData := []byte(proxyHeader + httpData)

	t.Logf("Created v1 test data with %d bytes", len(testData))

	// Test detection
	detected := detectProxyProtocol(testData)
	if !detected {
		t.Fatalf("Proxy Protocol v1 not detected!")
	}
	t.Logf("✅ Proxy Protocol v1 detected successfully")

	// Test parsing
	processedData, proxyInfo, err := processProxyProtocolData(testData)
	if err != nil {
		t.Fatalf("Parsing error: %v", err)
	}

	if proxyInfo == nil {
		t.Fatalf("No proxy info returned")
	}

	t.Logf("✅ Proxy Protocol v1 parsed successfully")
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

	if proxyInfo.Version != 1 {
		t.Errorf("Expected version 1, got %d", proxyInfo.Version)
	}

	if proxyInfo.TransportProto != "TCP4" {
		t.Errorf("Expected transport TCP4, got %s", proxyInfo.TransportProto)
	}

	// Check that the remaining data is HTTP
	if len(processedData) == 0 {
		t.Error("No remaining data after proxy protocol header removal")
	} else if !strings.HasPrefix(string(processedData), "GET /") {
		t.Errorf("Expected HTTP data, got: %s", string(processedData[:min(50, len(processedData))]))
	} else {
		t.Logf("✅ Remaining data is HTTP request")
	}
}

// Test IPv6 support in Proxy Protocol v2
func TestProxyProtocolV2IPv6(t *testing.T) {
	// Create IPv6 test case
	// Using RFC 3849 documentation IPv6 addresses
	sourceIP := net.ParseIP("2001:db8::1")
	destIP := net.ParseIP("2001:db8::2")
	sourcePort := uint16(12345)
	destPort := uint16(443)

	var buffer bytes.Buffer

	// 1. Signature (12 bytes)
	buffer.Write([]byte(ProxyProtocolV2Prefix))

	// 2. Version and Command (1 byte) - Version 2, Command PROXY
	buffer.WriteByte(0x21) // 0010 0001 (version 2, command 1)

	// 3. Address Family and Protocol (1 byte) - IPv6 TCP
	buffer.WriteByte(0x21) // 0010 0001 (AF_INET6, STREAM)

	// 4. Length (2 bytes) - 36 bytes for IPv6 addresses + ports
	buffer.WriteByte(0x00)
	buffer.WriteByte(0x24) // 36 bytes

	// 5. Addresses and ports (36 bytes for IPv6)
	buffer.Write(sourceIP.To16())             // Source IP (16 bytes)
	buffer.Write(destIP.To16())               // Dest IP (16 bytes)
	buffer.WriteByte(byte(sourcePort >> 8))   // Source port high byte
	buffer.WriteByte(byte(sourcePort & 0xFF)) // Source port low byte
	buffer.WriteByte(byte(destPort >> 8))     // Dest port high byte
	buffer.WriteByte(byte(destPort & 0xFF))   // Dest port low byte

	// 6. Add some TLS handshake data
	tlsData := []byte{0x16, 0x03, 0x01, 0x00, 0x10}
	buffer.Write(tlsData)

	testData := buffer.Bytes()

	// Test detection
	detected := detectProxyProtocol(testData)
	if !detected {
		t.Fatalf("Proxy Protocol v2 IPv6 not detected!")
	}

	// Test parsing
	processedData, proxyInfo, err := processProxyProtocolData(testData)
	if err != nil {
		t.Fatalf("IPv6 parsing error: %v", err)
	}

	if proxyInfo == nil {
		t.Fatalf("No proxy info returned for IPv6")
	}

	if len(processedData) == 0 {
		t.Error("No remaining data after IPv6 proxy protocol header removal")
	}

	// Verify IPv6 addresses
	if proxyInfo.SourceAddr != "2001:db8::1" {
		t.Errorf("Expected source IP 2001:db8::1, got %s", proxyInfo.SourceAddr)
	}

	if proxyInfo.TransportProto != "TCP6" {
		t.Errorf("Expected transport TCP6, got %s", proxyInfo.TransportProto)
	}

	t.Logf("✅ IPv6 support verified: %s:%d -> %s:%d",
		proxyInfo.SourceAddr, proxyInfo.SourcePort,
		proxyInfo.DestAddr, proxyInfo.DestPort)
}

// Test Proxy Protocol v1 detection and parsing
func TestProxyProtocolV1Format(t *testing.T) {
	// Create a sample Proxy Protocol v1 header
	// Format: "PROXY TCP4 192.0.2.100 198.51.100.50 45678 443\r\n"
	proxyHeader := "PROXY TCP4 192.0.2.100 198.51.100.50 45678 443\r\n"

	// Add some HTTP data
	httpData := "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"

	testData := []byte(proxyHeader + httpData)

	t.Logf("Created v1 test data with %d bytes", len(testData))

	// Test detection
	detected := detectProxyProtocol(testData)
	if !detected {
		t.Fatalf("Proxy Protocol v1 not detected!")
	}
	t.Logf("✅ Proxy Protocol v1 detected successfully")

	// Test parsing
	processedData, proxyInfo, err := processProxyProtocolData(testData)
	if err != nil {
		t.Fatalf("Parsing error: %v", err)
	}

	if proxyInfo == nil {
		t.Fatalf("No proxy info returned")
	}

	t.Logf("✅ Proxy Protocol v1 parsed successfully")
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

	if proxyInfo.Version != 1 {
		t.Errorf("Expected version 1, got %d", proxyInfo.Version)
	}

	if proxyInfo.TransportProto != "TCP4" {
		t.Errorf("Expected transport TCP4, got %s", proxyInfo.TransportProto)
	}

	// Check that the remaining data is HTTP
	if len(processedData) == 0 {
		t.Error("No remaining data after proxy protocol header removal")
	} else if !strings.HasPrefix(string(processedData), "GET /") {
		t.Errorf("Expected HTTP data, got: %s", string(processedData[:min(50, len(processedData))]))
	} else {
		t.Logf("✅ Remaining data is HTTP request")
	}
}

// Test IPv6 support in Proxy Protocol v2
func TestProxyProtocolV2IPv6(t *testing.T) {
	// Create IPv6 test case
	// Using RFC 3849 documentation IPv6 addresses
	sourceIP := net.ParseIP("2001:db8::1")
	destIP := net.ParseIP("2001:db8::2")
	sourcePort := uint16(12345)
	destPort := uint16(443)

	var buffer bytes.Buffer

	// 1. Signature (12 bytes)
	buffer.Write([]byte(ProxyProtocolV2Prefix))

	// 2. Version and Command (1 byte) - Version 2, Command PROXY
	buffer.WriteByte(0x21) // 0010 0001 (version 2, command 1)

	// 3. Address Family and Protocol (1 byte) - IPv6 TCP
	buffer.WriteByte(0x21) // 0010 0001 (AF_INET6, STREAM)

	// 4. Length (2 bytes) - 36 bytes for IPv6 addresses + ports
	buffer.WriteByte(0x00)
	buffer.WriteByte(0x24) // 36 bytes

	// 5. Addresses and ports (36 bytes for IPv6)
	buffer.Write(sourceIP.To16())             // Source IP (16 bytes)
	buffer.Write(destIP.To16())               // Dest IP (16 bytes)
	buffer.WriteByte(byte(sourcePort >> 8))   // Source port high byte
	buffer.WriteByte(byte(sourcePort & 0xFF)) // Source port low byte
	buffer.WriteByte(byte(destPort >> 8))     // Dest port high byte
	buffer.WriteByte(byte(destPort & 0xFF))   // Dest port low byte

	// 6. Add some TLS handshake data
	tlsData := []byte{0x16, 0x03, 0x01, 0x00, 0x10}
	buffer.Write(tlsData)

	testData := buffer.Bytes()

	// Test detection
	detected := detectProxyProtocol(testData)
	if !detected {
		t.Fatalf("Proxy Protocol v2 IPv6 not detected!")
	}

	// Test parsing
	processedData, proxyInfo, err := processProxyProtocolData(testData)
	if err != nil {
		t.Fatalf("IPv6 parsing error: %v", err)
	}

	if proxyInfo == nil {
		t.Fatalf("No proxy info returned for IPv6")
	}

	if len(processedData) == 0 {
		t.Error("No remaining data after IPv6 proxy protocol header removal")
	}

	// Verify IPv6 addresses
	if proxyInfo.SourceAddr != "2001:db8::1" {
		t.Errorf("Expected source IP 2001:db8::1, got %s", proxyInfo.SourceAddr)
	}

	if proxyInfo.TransportProto != "TCP6" {
		t.Errorf("Expected transport TCP6, got %s", proxyInfo.TransportProto)
	}

	t.Logf("✅ IPv6 support verified: %s:%d -> %s:%d",
		proxyInfo.SourceAddr, proxyInfo.SourcePort,
		proxyInfo.DestAddr, proxyInfo.DestPort)
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
		{
			name: "Partial proxy protocol v1",
			data: []byte("PROXY TCP4 192.0.2.1"),
		},
		{
			name: "Invalid proxy protocol v1",
			data: []byte("PROXY TCP4 192.0.2.1 198.51.100.1\r\n"),
		},
		{
			name: "Partial proxy protocol v2",
			data: []byte(ProxyProtocolV2Prefix[:8]),
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

// Test edge cases for robust error handling
func TestProxyProtocolEdgeCases(t *testing.T) {
	t.Run("Malformed v1 header", func(t *testing.T) {
		// Missing CRLF
		testData := []byte("PROXY TCP4 192.0.2.100 198.51.100.50 45678 443")

		detected := detectProxyProtocol(testData)
		if detected {
			t.Error("Malformed v1 header should not be detected")
		}
	})

	t.Run("Insufficient v2 data", func(t *testing.T) {
		// Only signature, no version/command
		testData := []byte(ProxyProtocolV2Prefix)

		detected := detectProxyProtocol(testData)
		if detected {
			t.Error("Insufficient v2 data should not be detected")
		}
	})

	t.Run("v2 with UNKNOWN connection", func(t *testing.T) {
		var buffer bytes.Buffer
		buffer.Write([]byte(ProxyProtocolV2Prefix))
		buffer.WriteByte(0x20) // Version 2, Command LOCAL
		buffer.WriteByte(0x00) // AF_UNSPEC
		buffer.WriteByte(0x00) // Length high
		buffer.WriteByte(0x00) // Length low

		testData := buffer.Bytes()

		detected := detectProxyProtocol(testData)
		if !detected {
			t.Error("v2 with UNKNOWN connection should be detected")
		}

		// Test parsing
		_, proxyInfo, err := processProxyProtocolData(testData)
		if err != nil {
			t.Fatalf("Should handle UNKNOWN connection: %v", err)
		}

		if proxyInfo.Version != 2 {
			t.Errorf("Expected version 2, got %d", proxyInfo.Version)
		}
	})
}

// Benchmark tests for performance
func BenchmarkProxyProtocolV1Detection(b *testing.B) {
	testData := []byte("PROXY TCP4 192.0.2.100 198.51.100.50 45678 443\r\nGET / HTTP/1.1\r\n\r\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detectProxyProtocol(testData)
	}
}

func BenchmarkProxyProtocolV2Detection(b *testing.B) {
	var buffer bytes.Buffer
	buffer.Write([]byte(ProxyProtocolV2Prefix))
	buffer.WriteByte(0x21) // Version 2, Command PROXY
	buffer.WriteByte(0x11) // AF_INET, STREAM
	buffer.WriteByte(0x00) // Length high
	buffer.WriteByte(0x0C) // Length low

	// Add dummy IPv4 addresses and ports
	buffer.Write([]byte{192, 0, 2, 100})   // Source IP
	buffer.Write([]byte{198, 51, 100, 50}) // Dest IP
	buffer.WriteByte(0xB2)                 // Source port high (45678)
	buffer.WriteByte(0x6E)                 // Source port low
	buffer.WriteByte(0x01)                 // Dest port high (443)
	buffer.WriteByte(0xBB)                 // Dest port low

	testData := buffer.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detectProxyProtocol(testData)
	}
}

func BenchmarkProxyProtocolV1Parsing(b *testing.B) {
	testData := []byte("PROXY TCP4 192.0.2.100 198.51.100.50 45678 443\r\nGET / HTTP/1.1\r\n\r\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processProxyProtocolData(testData)
	}
}

func BenchmarkProxyProtocolV2Parsing(b *testing.B) {
	var buffer bytes.Buffer
	buffer.Write([]byte(ProxyProtocolV2Prefix))
	buffer.WriteByte(0x21) // Version 2, Command PROXY
	buffer.WriteByte(0x11) // AF_INET, STREAM
	buffer.WriteByte(0x00) // Length high
	buffer.WriteByte(0x0C) // Length low

	// Add dummy IPv4 addresses and ports
	buffer.Write([]byte{192, 0, 2, 100})   // Source IP
	buffer.Write([]byte{198, 51, 100, 50}) // Dest IP
	buffer.WriteByte(0xB2)                 // Source port high (45678)
	buffer.WriteByte(0x6E)                 // Source port low
	buffer.WriteByte(0x01)                 // Dest port high (443)
	buffer.WriteByte(0xBB)                 // Dest port low

	testData := buffer.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processProxyProtocolData(testData)
	}
}

// Test HTTP API handlers
func TestAPIHandlers(t *testing.T) {
	t.Run("Status API GET", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/ui/api/status", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handleAPIStatus)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
		}

		var response StatusResponse
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to parse JSON response: %v", err)
		}

		if response.Status == "" {
			t.Error("Status field should not be empty")
		}

		if response.Version == "" {
			t.Error("Version field should not be empty")
		}
	})

	t.Run("Status API POST - Method Not Allowed", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/ui/api/status", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handleAPIStatus)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusMethodNotAllowed {
			t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, status)
		}
	})

	t.Run("Toggle API POST - Enable", func(t *testing.T) {
		reqBody := `{"enabled": true}`
		req, err := http.NewRequest("POST", "/ui/api/toggle", strings.NewReader(reqBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", "test-token") // Add required CSRF token

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handleAPIToggle)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
		}

		var response ToggleResponse
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to parse JSON response: %v", err)
		}

		if response.Result != "success" {
			t.Errorf("Expected result 'success', got '%s'", response.Result)
		}

		if !response.Enabled {
			t.Error("Expected enabled to be true")
		}
	})

	t.Run("Toggle API POST - Disable", func(t *testing.T) {
		reqBody := `{"enabled": false}`
		req, err := http.NewRequest("POST", "/ui/api/toggle", strings.NewReader(reqBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", "test-token") // Add required CSRF token

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handleAPIToggle)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
		}

		var response ToggleResponse
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to parse JSON response: %v", err)
		}

		if response.Enabled {
			t.Error("Expected enabled to be false")
		}
	})

	t.Run("Toggle API GET - Method Not Allowed", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/ui/api/toggle", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handleAPIToggle)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusMethodNotAllowed {
			t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, status)
		}
	})

	t.Run("Toggle API POST - Invalid JSON", func(t *testing.T) {
		reqBody := `{invalid json}`
		req, err := http.NewRequest("POST", "/ui/api/toggle", strings.NewReader(reqBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", "test-token") // Add required CSRF token

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handleAPIToggle)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, status)
		}
	})

	t.Run("Toggle API POST - Missing CSRF Token", func(t *testing.T) {
		reqBody := `{"enabled": true}`
		req, err := http.NewRequest("POST", "/ui/api/toggle", strings.NewReader(reqBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		// Intentionally not setting CSRF token

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handleAPIToggle)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusForbidden {
			t.Errorf("Expected status code %d, got %d", http.StatusForbidden, status)
		}
	})
}

// Test utility functions
func TestUtilityFunctions(t *testing.T) {
	t.Run("min function", func(t *testing.T) {
		if min(5, 3) != 3 {
			t.Error("min(5, 3) should return 3")
		}
		if min(2, 8) != 2 {
			t.Error("min(2, 8) should return 2")
		}
		if min(4, 4) != 4 {
			t.Error("min(4, 4) should return 4")
		}
	})

	t.Run("isPluginHealthy function", func(t *testing.T) {
		healthy := isPluginHealthy()
		if !healthy {
			t.Error("Plugin should be healthy in test environment")
		}
	})
}

// Test proxy protocol parsing error cases
func TestProxyProtocolParsingErrors(t *testing.T) {
	t.Run("processProxyProtocolData with empty data", func(t *testing.T) {
		data, info, err := processProxyProtocolData([]byte{})
		if err != nil {
			t.Errorf("Empty data should not cause error: %v", err)
		}
		if info != nil {
			t.Error("Empty data should not return proxy info")
		}
		if len(data) != 0 {
			t.Error("Empty data should return empty data")
		}
	})

	t.Run("processProxyProtocolData with invalid v1 format", func(t *testing.T) {
		invalidV1 := []byte("PROXY TCP4 invalid_format\r\n")
		returnedData, info, err := processProxyProtocolData(invalidV1)
		if err == nil {
			t.Error("Invalid v1 format should cause error")
		}
		if info != nil {
			t.Error("Invalid v1 format should not return proxy info")
		}
		if returnedData != nil {
			t.Error("Invalid v1 format should return nil data when error occurs")
		}
	})

	t.Run("processProxyProtocolData with malformed v2 header", func(t *testing.T) {
		var buffer bytes.Buffer
		buffer.Write([]byte(ProxyProtocolV2Prefix))
		buffer.WriteByte(0x21) // Version 2, Command PROXY
		buffer.WriteByte(0x11) // AF_INET, STREAM
		buffer.WriteByte(0xFF) // Invalid length high
		buffer.WriteByte(0xFF) // Invalid length low

		malformedV2 := buffer.Bytes()
		returnedData, info, err := processProxyProtocolData(malformedV2)
		if err == nil {
			t.Error("Malformed v2 header should cause error")
		}
		if info != nil {
			t.Error("Malformed v2 header should not return proxy info")
		}
		if returnedData != nil {
			t.Error("Malformed v2 header should return nil data when error occurs")
		}
	})

	t.Run("processProxyProtocolData with v1 missing ports", func(t *testing.T) {
		invalidV1 := []byte("PROXY TCP4 192.0.2.1 198.51.100.1\r\n")
		_, info, err := processProxyProtocolData(invalidV1)
		if err == nil {
			t.Error("V1 missing ports should cause error")
		}
		if info != nil {
			t.Error("V1 missing ports should not return proxy info")
		}
	})

	t.Run("processProxyProtocolData with v2 unsupported address family", func(t *testing.T) {
		var buffer bytes.Buffer
		buffer.Write([]byte(ProxyProtocolV2Prefix))
		buffer.WriteByte(0x21) // Version 2, Command PROXY
		buffer.WriteByte(0x31) // AF_UNIX, STREAM (unsupported)
		buffer.WriteByte(0x00) // Length high
		buffer.WriteByte(0x00) // Length low

		unsupportedV2 := buffer.Bytes()
		_, info, err := processProxyProtocolData(unsupportedV2)
		if err == nil {
			t.Error("Unsupported address family should cause error")
		}
		if info != nil {
			t.Error("Unsupported address family should not return proxy info")
		}
	})
}

// Test proxy protocol listener functionality
func TestProxyProtocolListener(t *testing.T) {
	t.Run("NewProxyProtocolListener", func(t *testing.T) {
		// Create a mock listener
		listener := &mockListener{}
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		ppListener := NewProxyProtocolListener(listener, handler, logger)

		if ppListener.Listener != listener {
			t.Error("Listener should be set correctly")
		}
		if ppListener.OriginalHandler == nil {
			t.Error("Handler should be set")
		}
		if ppListener.Logger != logger {
			t.Error("Logger should be set correctly")
		}
		if ppListener.ReadTimeout != 5*time.Second {
			t.Error("Default timeout should be 5 seconds")
		}
	})

	t.Run("proxyProtocolAddr", func(t *testing.T) {
		addr := &proxyProtocolAddr{
			address: "192.0.2.1",
			port:    8080,
		}

		if addr.Network() != "tcp" {
			t.Error("Network should return 'tcp'")
		}

		expected := "192.0.2.1:8080"
		if addr.String() != expected {
			t.Errorf("Expected address string '%s', got '%s'", expected, addr.String())
		}
	})
}

// Mock listener for testing
type mockListener struct {
	closed bool
}

func (m *mockListener) Accept() (net.Conn, error) {
	if m.closed {
		return nil, fmt.Errorf("listener closed")
	}
	return &mockConn{}, nil
}

func (m *mockListener) Close() error {
	m.closed = true
	return nil
}

func (m *mockListener) Addr() net.Addr {
	return &mockAddr{}
}

// Mock connection for testing
type mockConn struct {
	data []byte
	pos  int
}

func (m *mockConn) Read(b []byte) (int, error) {
	if m.pos >= len(m.data) {
		return 0, io.EOF
	}
	n := copy(b, m.data[m.pos:])
	m.pos += n
	return n, nil
}

func (m *mockConn) Write(b []byte) (int, error) {
	return len(b), nil
}

func (m *mockConn) Close() error {
	return nil
}

func (m *mockConn) LocalAddr() net.Addr {
	return &mockAddr{}
}

func (m *mockConn) RemoteAddr() net.Addr {
	return &mockAddr{}
}

func (m *mockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// Mock address for testing
type mockAddr struct{}

func (m *mockAddr) Network() string {
	return "tcp"
}

func (m *mockAddr) String() string {
	return "127.0.0.1:8080"
}

// Test additional proxy protocol v2 edge cases
func TestProxyProtocolV2EdgeCases(t *testing.T) {
	t.Run("v2 with LOCAL command", func(t *testing.T) {
		var buffer bytes.Buffer
		buffer.Write([]byte(ProxyProtocolV2Prefix))
		buffer.WriteByte(0x20) // Version 2, Command LOCAL
		buffer.WriteByte(0x00) // AF_UNSPEC
		buffer.WriteByte(0x00) // Length high
		buffer.WriteByte(0x00) // Length low

		testData := buffer.Bytes()
		processedData, proxyInfo, err := processProxyProtocolData(testData)

		if err != nil {
			t.Errorf("LOCAL command should not cause error: %v", err)
		}
		if proxyInfo.Version != 2 {
			t.Errorf("Expected version 2, got %d", proxyInfo.Version)
		}
		if len(processedData) != 0 {
			t.Error("LOCAL command should consume all header data")
		}
	})

	t.Run("v2 with truncated header", func(t *testing.T) {
		var buffer bytes.Buffer
		buffer.Write([]byte(ProxyProtocolV2Prefix))
		buffer.WriteByte(0x21) // Version 2, Command PROXY
		buffer.WriteByte(0x11) // AF_INET, STREAM
		buffer.WriteByte(0x00) // Length high
		buffer.WriteByte(0x0C) // Length low (12 bytes)
		// Only provide 6 bytes instead of 12
		buffer.Write([]byte{192, 0, 2, 100, 198, 51})

		testData := buffer.Bytes()
		_, _, err := processProxyProtocolData(testData)

		if err == nil {
			t.Error("Truncated v2 header should cause error")
		}
	})
}

// Test version flag handling
func TestVersionFlag(t *testing.T) {
	t.Run("Version components should be set", func(t *testing.T) {
		if versionMajor == "" {
			t.Error("versionMajor should not be empty")
		}
		if versionMinor == "" {
			t.Error("versionMinor should not be empty")
		}
		if versionPatch == "" {
			t.Error("versionPatch should not be empty")
		}
	})
}

// Test concurrent access to config
func TestConfigConcurrency(t *testing.T) {
	t.Run("Concurrent config access", func(t *testing.T) {
		var wg sync.WaitGroup
		const numGoroutines = 10

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(enabled bool) {
				defer wg.Done()
				config.mu.Lock()
				config.Enabled = enabled
				config.mu.Unlock()

				config.mu.RLock()
				_ = config.Enabled
				config.mu.RUnlock()
			}(i%2 == 0)
		}

		wg.Wait()
	})
}

// Test CORS headers
func TestCORSHeaders(t *testing.T) {
	t.Run("Status API has CORS headers", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/ui/api/status", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handleAPIStatus)
		handler.ServeHTTP(rr, req)

		corsHeader := rr.Header().Get("Access-Control-Allow-Origin")
		if corsHeader != "*" {
			t.Errorf("Expected CORS header '*', got '%s'", corsHeader)
		}

		contentType := rr.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}
	})
}

// Test core plugin handlers
func TestProxyProtocolHandlers(t *testing.T) {
	t.Run("handleProxyProtocolSniff - Plugin Disabled", func(t *testing.T) {
		// Ensure plugin is disabled
		config.mu.Lock()
		config.Enabled = false
		config.mu.Unlock()

		req, err := http.NewRequest("POST", "/proxy_protocol_sniff", strings.NewReader("test data"))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handleProxyProtocolSniff)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != 284 {
			t.Errorf("Expected status code 284 (UNHANDLED), got %d", status)
		}

		if body := rr.Body.String(); body != "UNHANDLED" {
			t.Errorf("Expected body 'UNHANDLED', got '%s'", body)
		}
	})

	t.Run("handleProxyProtocolSniff - Plugin Enabled, No Proxy Protocol", func(t *testing.T) {
		// Enable plugin
		config.mu.Lock()
		config.Enabled = true
		config.mu.Unlock()

		httpData := "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"
		req, err := http.NewRequest("POST", "/proxy_protocol_sniff", strings.NewReader(httpData))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handleProxyProtocolSniff)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != 284 {
			t.Errorf("Expected status code 284 (UNHANDLED), got %d", status)
		}

		if body := rr.Body.String(); body != "UNHANDLED" {
			t.Errorf("Expected body 'UNHANDLED', got '%s'", body)
		}
	})

	t.Run("handleProxyProtocolSniff - Plugin Enabled, With Proxy Protocol", func(t *testing.T) {
		// Enable plugin
		config.mu.Lock()
		config.Enabled = true
		config.mu.Unlock()

		proxyData := "PROXY TCP4 192.0.2.100 198.51.100.50 45678 443\r\nGET / HTTP/1.1\r\n\r\n"
		req, err := http.NewRequest("POST", "/proxy_protocol_sniff", strings.NewReader(proxyData))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handleProxyProtocolSniff)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != 280 {
			t.Errorf("Expected status code 280 (CAPTURED), got %d", status)
		}

		if body := rr.Body.String(); body != "CAPTURED" {
			t.Errorf("Expected body 'CAPTURED', got '%s'", body)
		}
	})

	t.Run("handleProxyProtocolSniff - Read Error", func(t *testing.T) {
		// Enable plugin
		config.mu.Lock()
		config.Enabled = true
		config.mu.Unlock()

		// Create a request with a body that will cause read error
		req, err := http.NewRequest("POST", "/proxy_protocol_sniff", &errorReader{})
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handleProxyProtocolSniff)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != 580 {
			t.Errorf("Expected status code 580 (ERROR), got %d", status)
		}

		if body := rr.Body.String(); body != "ERROR" {
			t.Errorf("Expected body 'ERROR', got '%s'", body)
		}
	})

	t.Run("handleProxyProtocolIngress - Plugin Disabled", func(t *testing.T) {
		// Disable plugin
		config.mu.Lock()
		config.Enabled = false
		config.mu.Unlock()

		req, err := http.NewRequest("POST", "/proxy_protocol_handler", strings.NewReader("test"))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handleProxyProtocolIngress)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusServiceUnavailable {
			t.Errorf("Expected status code %d, got %d", http.StatusServiceUnavailable, status)
		}
	})

	t.Run("handleProxyProtocolIngress - Missing Connection ID", func(t *testing.T) {
		// Enable plugin
		config.mu.Lock()
		config.Enabled = true
		config.mu.Unlock()

		req, err := http.NewRequest("POST", "/proxy_protocol_handler", strings.NewReader("test"))
		if err != nil {
			t.Fatal(err)
		}
		// Don't set X-Connection-ID header

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handleProxyProtocolIngress)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, status)
		}
	})

	t.Run("handleProxyProtocolIngress - Valid Proxy Protocol", func(t *testing.T) {
		// Enable plugin
		config.mu.Lock()
		config.Enabled = true
		config.mu.Unlock()

		proxyData := "PROXY TCP4 192.0.2.100 198.51.100.50 45678 443\r\nGET / HTTP/1.1\r\n\r\n"
		req, err := http.NewRequest("POST", "/proxy_protocol_handler", strings.NewReader(proxyData))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("X-Connection-ID", "test-conn-123")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handleProxyProtocolIngress)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
		}

		// Check for proxy protocol headers
		if header := rr.Header().Get("X-Original-Remote-Addr"); header != "192.0.2.100" {
			t.Errorf("Expected X-Original-Remote-Addr '192.0.2.100', got '%s'", header)
		}

		if header := rr.Header().Get("X-Real-IP"); header != "192.0.2.100" {
			t.Errorf("Expected X-Real-IP '192.0.2.100', got '%s'", header)
		}
	})

	t.Run("handleProxyProtocolIngress - Read Error", func(t *testing.T) {
		// Enable plugin
		config.mu.Lock()
		config.Enabled = true
		config.mu.Unlock()

		req, err := http.NewRequest("POST", "/proxy_protocol_handler", &errorReader{})
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("X-Connection-ID", "test-conn-123")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handleProxyProtocolIngress)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, status)
		}
	})

	t.Run("handleProxyProtocolIngress - Parse Error", func(t *testing.T) {
		// Enable plugin
		config.mu.Lock()
		config.Enabled = true
		config.mu.Unlock()

		// Send malformed proxy protocol data
		malformedData := "PROXY TCP4 invalid_format\r\n"
		req, err := http.NewRequest("POST", "/proxy_protocol_handler", strings.NewReader(malformedData))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("X-Connection-ID", "test-conn-123")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handleProxyProtocolIngress)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, status)
		}
	})
}

// Test ProxyProtocolListener methods
func TestProxyProtocolListenerMethods(t *testing.T) {
	t.Run("ProxyProtocolListener Accept", func(t *testing.T) {
		mockListener := &mockListener{}
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		ppListener := NewProxyProtocolListener(mockListener, handler, logger)

		conn, err := ppListener.Accept()
		if err != nil {
			t.Errorf("Accept should not error: %v", err)
		}
		if conn == nil {
			t.Error("Accept should return a connection")
		}
	})

	t.Run("ProxyProtocolListener Close", func(t *testing.T) {
		mockListener := &mockListener{}
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		ppListener := NewProxyProtocolListener(mockListener, handler, logger)

		err := ppListener.Close()
		if err != nil {
			t.Errorf("Close should not error: %v", err)
		}

		if !mockListener.closed {
			t.Error("Underlying listener should be closed")
		}
	})

	t.Run("ProxyProtocolListener Addr", func(t *testing.T) {
		mockListener := &mockListener{}
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		ppListener := NewProxyProtocolListener(mockListener, handler, logger)

		addr := ppListener.Addr()
		if addr == nil {
			t.Error("Addr should return an address")
		}
	})
}

// Test proxyProtocolConn methods
func TestProxyProtocolConnMethods(t *testing.T) {
	t.Run("proxyProtocolConn Read", func(t *testing.T) {
		testData := []byte("test data")
		mockConn := &mockConn{data: testData}

		// Create proxy protocol connection with a bufio reader
		proxyInfo := &ProxyProtocolInfo{
			SourceAddr: "192.0.2.1",
			SourcePort: 12345,
			DestAddr:   "198.51.100.1",
			DestPort:   80,
		}

		reader := bufio.NewReader(strings.NewReader(string(testData)))
		ppConn := &proxyProtocolConn{
			Conn:            mockConn,
			ProxyInfo:       proxyInfo,
			BufReader:       reader,
			proxyRemoteAddr: &proxyProtocolAddr{"192.0.2.1", 12345},
			proxyLocalAddr:  &proxyProtocolAddr{"198.51.100.1", 80},
		}

		buffer := make([]byte, 100)
		n, err := ppConn.Read(buffer)
		if err != nil {
			t.Errorf("Read should not error: %v", err)
		}
		if n != len(testData) {
			t.Errorf("Expected to read %d bytes, got %d", len(testData), n)
		}
		if string(buffer[:n]) != string(testData) {
			t.Errorf("Expected to read '%s', got '%s'", string(testData), string(buffer[:n]))
		}
	})

	t.Run("proxyProtocolConn LocalAddr", func(t *testing.T) {
		ppConn := &proxyProtocolConn{
			proxyLocalAddr: &proxyProtocolAddr{"198.51.100.1", 80},
		}

		addr := ppConn.LocalAddr()
		if addr == nil {
			t.Error("LocalAddr should return an address")
		}
		if addr.String() != "198.51.100.1:80" {
			t.Errorf("Expected address '198.51.100.1:80', got '%s'", addr.String())
		}
	})

	t.Run("proxyProtocolConn RemoteAddr", func(t *testing.T) {
		ppConn := &proxyProtocolConn{
			proxyRemoteAddr: &proxyProtocolAddr{"192.0.2.1", 12345},
		}

		addr := ppConn.RemoteAddr()
		if addr == nil {
			t.Error("RemoteAddr should return an address")
		}
		if addr.String() != "192.0.2.1:12345" {
			t.Errorf("Expected address '192.0.2.1:12345', got '%s'", addr.String())
		}
	})
}

// Test middleware function
func TestProxyProtocolMiddleware(t *testing.T) {
	t.Run("ProxyProtocolMiddleware", func(t *testing.T) {
		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test response"))
		})

		// Wrap with middleware
		middlewareHandler := ProxyProtocolMiddleware(testHandler)

		req, err := http.NewRequest("GET", "/test", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		middlewareHandler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
		}

		if body := rr.Body.String(); body != "test response" {
			t.Errorf("Expected body 'test response', got '%s'", body)
		}
	})
}

// Helper type for testing read errors
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated read error")
}

// Test additional parsing v1 edge cases
func TestParseProxyProtocolV1EdgeCases(t *testing.T) {
	t.Run("parseProxyProtocolV1 with TCP6", func(t *testing.T) {
		reader := bufio.NewReader(strings.NewReader("PROXY TCP6 2001:db8::1 2001:db8::2 12345 443\r\n"))

		info, err := parseProxyProtocolV1(reader)
		if err != nil {
			t.Errorf("TCP6 parsing should not error: %v", err)
		}
		if info == nil {
			t.Fatal("Should return proxy info")
		}
		if info.TransportProto != "TCP6" {
			t.Errorf("Expected TCP6, got %s", info.TransportProto)
		}
		if info.SourceAddr != "2001:db8::1" {
			t.Errorf("Expected source addr '2001:db8::1', got '%s'", info.SourceAddr)
		}
	})

	t.Run("parseProxyProtocolV1 with UNKNOWN", func(t *testing.T) {
		// According to proxy protocol v1 spec, UNKNOWN format should include placeholders
		reader := bufio.NewReader(strings.NewReader("PROXY UNKNOWN 0.0.0.0 0.0.0.0 0 0\r\n"))

		info, err := parseProxyProtocolV1(reader)
		if err != nil {
			t.Errorf("UNKNOWN parsing should not error: %v", err)
		}
		if info == nil {
			t.Fatal("Should return proxy info")
		}
		if info.TransportProto != "UNKNOWN" {
			t.Errorf("Expected UNKNOWN, got %s", info.TransportProto)
		}
		if info.SourceAddr != "0.0.0.0" {
			t.Errorf("Expected source addr '0.0.0.0', got '%s'", info.SourceAddr)
		}
		if info.SourcePort != 0 {
			t.Errorf("Expected source port 0, got %d", info.SourcePort)
		}
	})
}

// Test additional parsing v2 edge cases
func TestParseProxyProtocolV2EdgeCases(t *testing.T) {
	t.Run("parseProxyProtocolV2 with IPv6", func(t *testing.T) {
		var buffer bytes.Buffer
		buffer.Write([]byte(ProxyProtocolV2Prefix))
		buffer.WriteByte(0x21) // Version 2, Command PROXY
		buffer.WriteByte(0x21) // AF_INET6, STREAM
		buffer.WriteByte(0x00) // Length high
		buffer.WriteByte(0x24) // Length low (36 bytes for IPv6)

		// Add IPv6 addresses (16 bytes each)
		sourceIP := net.ParseIP("2001:db8::1")
		destIP := net.ParseIP("2001:db8::2")
		buffer.Write(sourceIP.To16())
		buffer.Write(destIP.To16())
		buffer.WriteByte(0x30) // Source port high (12345)
		buffer.WriteByte(0x39) // Source port low
		buffer.WriteByte(0x01) // Dest port high (443)
		buffer.WriteByte(0xBB) // Dest port low

		reader := bufio.NewReader(&buffer)

		info, err := parseProxyProtocolV2(reader)
		if err != nil {
			t.Errorf("IPv6 v2 parsing should not error: %v", err)
		}
		if info == nil {
			t.Fatal("Should return proxy info")
		}
		if info.TransportProto != "TCP6" {
			t.Errorf("Expected TCP6, got %s", info.TransportProto)
		}
		if info.SourceAddr != "2001:db8::1" {
			t.Errorf("Expected source addr '2001:db8::1', got '%s'", info.SourceAddr)
		}
	})

	t.Run("parseProxyProtocolV2 with unsupported family", func(t *testing.T) {
		var buffer bytes.Buffer
		buffer.Write([]byte(ProxyProtocolV2Prefix))
		buffer.WriteByte(0x21) // Version 2, Command PROXY
		buffer.WriteByte(0x31) // AF_UNIX, STREAM (unsupported)
		buffer.WriteByte(0x00) // Length high
		buffer.WriteByte(0x00) // Length low

		reader := bufio.NewReader(&buffer)

		_, err := parseProxyProtocolV2(reader)
		if err == nil {
			t.Error("Unsupported address family should cause error")
		}
	})
}
