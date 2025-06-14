package main

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	plugin "go.codexo.de/exoridus/zoraxy-proxy-protocol/mod/zoraxy_plugin"
)

const (
	PLUGIN_ID = "de.codexo.proxyprotocol"
	UI_PATH   = "/ui"
	WEB_ROOT  = "/www"
)

// Version information - set via ldflags during build
var versionMajor = "1"
var versionMinor = "0"
var versionPatch = "0"

//go:embed www/*
var content embed.FS

// Plugin configuration
type PluginConfig struct {
	Enabled bool `json:"enabled"`
	mu      sync.RWMutex
}

var config = &PluginConfig{
	Enabled: false,
}

// Logger for the plugin
var logger *log.Logger

// Plugin connection registry for active connections
var activeConnections = make(map[string]*proxyProtocolConn)
var connectionsMutex sync.RWMutex

// API response structures
type StatusResponse struct {
	Status  string `json:"status"`
	Enabled bool   `json:"enabled"`
	Version string `json:"version"`
}

type ToggleRequest struct {
	Enabled bool `json:"enabled"`
}

type ToggleResponse struct {
	Result  string `json:"result"`
	Enabled bool   `json:"enabled"`
}

func init() {
	logger = log.New(os.Stdout, "[ProxyProtocol] ", log.LstdFlags)
}

// isPluginHealthy checks if the plugin is functioning correctly
func isPluginHealthy() bool {
	// Basic health checks for the plugin

	// Check if logger is available
	if logger == nil {
		return false
	}

	// Check if config is accessible
	config.mu.RLock()
	defer config.mu.RUnlock()

	// Check if activeConnections map is accessible
	connectionsMutex.RLock()
	defer connectionsMutex.RUnlock()

	// If we reach here, basic components are working
	return true
}

func main() {
	// Check for version flag before doing anything else
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "--version" || arg == "-v" {
			fmt.Printf("proxy-protocol v%s.%s.%s\n", versionMajor, versionMinor, versionPatch)
			os.Exit(0)
		}
	}

	// Convert string flags to integers for the plugin spec
	major, _ := strconv.Atoi(versionMajor)
	minor, _ := strconv.Atoi(versionMinor)
	patch, _ := strconv.Atoi(versionPatch)

	runtimeCfg, err := plugin.ServeAndRecvSpec(&plugin.IntroSpect{
		ID:            PLUGIN_ID,
		Name:          "Proxy Protocol",
		Author:        "Exoridus",
		AuthorContact: "https://github.com/Exoridus",
		Description:   "Adds support for Proxy Protocol",
		URL:           "https://github.com/Exoridus/zoraxy-proxy-protocol",
		Type:          plugin.PluginType_Router,
		VersionMajor:  major,
		VersionMinor:  minor,
		VersionPatch:  patch,

		// No static capture paths as we work at network level
		StaticCapturePaths:   []plugin.StaticCaptureRule{},
		StaticCaptureIngress: "",

		// Dynamic capturing for Proxy Protocol header detection
		DynamicCaptureSniff:   "/proxy_protocol_sniff",
		DynamicCaptureIngress: "/proxy_protocol_handler",

		// UI path for configuration
		UIPath: UI_PATH,

		// No subscription events needed
		SubscriptionPath:    "",
		SubscriptionsEvents: map[string]string{},
	})
	if err != nil {
		fmt.Println("This is a plugin for Zoraxy and should not be run standalone")
		fmt.Println("For installation instructions, see: https://github.com/Exoridus/zoraxy-proxy-protocol")
		panic(err)
	}

	// Register core plugin endpoints (required by Zoraxy)
	http.HandleFunc("/proxy_protocol_sniff", handleProxyProtocolSniff)
	http.HandleFunc("/proxy_protocol_handler", handleProxyProtocolIngress)

	// Register API endpoints BEFORE the embedded router for precedence
	http.HandleFunc(UI_PATH+"/api/status", handleAPIStatus)
	http.HandleFunc(UI_PATH+"/api/toggle", handleAPIToggle)

	// Create embedded web router for UI (this registers /ui/ pattern which is less specific)
	embedWebRouter := plugin.NewPluginEmbedUIRouter(PLUGIN_ID, &content, WEB_ROOT, UI_PATH)
	embedWebRouter.RegisterTerminateHandler(func() {
		fmt.Println("Proxy Protocol Plugin terminated")
	}, nil)
	embedWebRouter.AttachHandlerToMux(nil)

	fmt.Println("Proxy Protocol Plugin started at http://127.0.0.1:" + strconv.Itoa(runtimeCfg.Port))
	err = http.ListenAndServe("127.0.0.1:"+strconv.Itoa(runtimeCfg.Port), nil)
	if err != nil {
		panic(err)
	}
}

// API Handlers
func handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("API Status request: %s %s\n", r.Method, r.URL.Path)

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config.mu.RLock()
	enabled := config.Enabled
	config.mu.RUnlock()

	// Determine status based on enabled state and potential errors
	status := "Disabled"
	if enabled {
		// Check if plugin is functioning correctly
		if isPluginHealthy() {
			status = "Enabled"
		} else {
			status = "Error"
			enabled = false // Override enabled to false if there's an error
		}
	}

	// Reconstruct version string from components
	versionString := fmt.Sprintf("%s.%s.%s", versionMajor, versionMinor, versionPatch)

	response := StatusResponse{
		Status:  status,
		Enabled: enabled,
		Version: versionString,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding JSON response: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Status response sent: %+v\n", response)
}

func handleAPIToggle(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("API Toggle request: %s %s\n", r.Method, r.URL.Path)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check for CSRF token
	csrfToken := r.Header.Get("X-CSRF-Token")
	if csrfToken == "" {
		fmt.Printf("CSRF token missing or invalid: %s\n", csrfToken)
		http.Error(w, "Forbidden - CSRF token not found in request", http.StatusForbidden)
		return
	}

	var req ToggleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Printf("Error decoding JSON: %v\n", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	fmt.Printf("Toggle request received: enabled=%t\n", req.Enabled)

	config.mu.Lock()
	config.Enabled = req.Enabled
	config.mu.Unlock()

	fmt.Printf("Proxy Protocol support %s\n", map[bool]string{true: "enabled", false: "disabled"}[req.Enabled])

	response := ToggleResponse{
		Result:  "success",
		Enabled: req.Enabled,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding JSON response: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Toggle response sent: %+v\n", response)
}

// Core plugin functionality - these are the endpoints that Zoraxy calls
func handleProxyProtocolSniff(w http.ResponseWriter, r *http.Request) {
	logger.Printf("=== SNIFF REQUEST RECEIVED ===")
	logger.Printf("Method: %s, URL: %s", r.Method, r.URL.Path)
	logger.Printf("Headers: %+v", r.Header)
	logger.Printf("Remote Addr: %s", r.RemoteAddr)

	config.mu.RLock()
	enabled := config.Enabled
	config.mu.RUnlock()

	logger.Printf("Plugin enabled status: %t", enabled)

	if !enabled {
		// Plugin disabled - let Zoraxy handle normally
		logger.Printf("Plugin disabled, returning UNHANDLED")
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(284) // ControlStatusCode_UNHANDLED
		w.Write([]byte("UNHANDLED"))
		return
	}

	// Get the raw connection data from the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Printf("Error reading request body for sniffing: %v", err)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(580) // ControlStatusCode_ERROR
		w.Write([]byte("ERROR"))
		return
	}

	logger.Printf("Received %d bytes of data for sniffing", len(body))
	if len(body) > 0 {
		// Log first few bytes in hex for debugging
		hexStr := ""
		for i := 0; i < min(32, len(body)); i++ {
			hexStr += fmt.Sprintf("%02x ", body[i])
		}
		logger.Printf("First bytes (hex): %s", hexStr)

		// Also show as string if printable
		printable := ""
		for i := 0; i < min(64, len(body)); i++ {
			if body[i] >= 32 && body[i] <= 126 {
				printable += string(body[i])
			} else {
				printable += fmt.Sprintf("\\x%02x", body[i])
			}
		}
		logger.Printf("First bytes (mixed): %s", printable)
	}

	// Check if this looks like proxy protocol data
	if detectProxyProtocol(body) {
		logger.Printf("✅ Proxy Protocol detected in connection")
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(280) // ControlStatusCode_CAPTURED - Tell Zoraxy we'll handle this
		w.Write([]byte("CAPTURED"))
		return
	}

	// Not proxy protocol data - let Zoraxy handle normally
	logger.Printf("❌ No Proxy Protocol detected, returning UNHANDLED")
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(284) // ControlStatusCode_UNHANDLED
	w.Write([]byte("UNHANDLED"))
}

func handleProxyProtocolIngress(w http.ResponseWriter, r *http.Request) {
	logger.Printf("=== INGRESS REQUEST RECEIVED ===")
	logger.Printf("Method: %s, URL: %s", r.Method, r.URL.Path)
	logger.Printf("Headers: %+v", r.Header)

	config.mu.RLock()
	enabled := config.Enabled
	config.mu.RUnlock()

	if !enabled {
		logger.Printf("Plugin disabled in ingress handler")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Proxy Protocol Handler Disabled"))
		return
	}

	// Get connection identifier from headers
	connID := r.Header.Get("X-Connection-ID")
	if connID == "" {
		logger.Printf("No connection ID provided in proxy protocol ingress")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No Connection ID"))
		return
	}

	logger.Printf("Processing connection ID: %s", connID)

	// Read the raw connection data
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Printf("Error reading proxy protocol data: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Read Error"))
		return
	}

	logger.Printf("Received %d bytes of data for processing", len(body))

	// Process the proxy protocol data
	processedData, proxyInfo, err := processProxyProtocolData(body)
	if err != nil {
		logger.Printf("Error processing proxy protocol: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Parse Error"))
		return
	}

	if proxyInfo != nil {
		logger.Printf("✅ Proxy Protocol parsed: %s:%d -> %s:%d (v%d)",
			proxyInfo.SourceAddr, proxyInfo.SourcePort,
			proxyInfo.DestAddr, proxyInfo.DestPort, proxyInfo.Version)

		// Store the proxy info for later use
		connectionsMutex.Lock()
		if conn, exists := activeConnections[connID]; exists {
			conn.ProxyInfo = proxyInfo
		}
		connectionsMutex.Unlock()

		// Set headers for the processed request with original client IP
		w.Header().Set("X-Original-Remote-Addr", proxyInfo.SourceAddr)
		w.Header().Set("X-Original-Remote-Port", strconv.Itoa(proxyInfo.SourcePort))
		w.Header().Set("X-Forwarded-For", proxyInfo.SourceAddr)
		w.Header().Set("X-Real-IP", proxyInfo.SourceAddr)
		w.Header().Set("X-Forwarded-Port", strconv.Itoa(proxyInfo.SourcePort))

		// Indicate to Zoraxy that it should use the original client IP for further processing
		w.Header().Set("X-Proxy-Protocol-Source", fmt.Sprintf("%s:%d", proxyInfo.SourceAddr, proxyInfo.SourcePort))
	} else {
		logger.Printf("No proxy protocol info found, passing through data unchanged")
	}

	// Return the processed data (without proxy protocol headers)
	// This should be the actual HTTP/HTTPS request that Zoraxy can process
	logger.Printf("Returning %d bytes of processed data", len(processedData))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write(processedData)
}

// detectProxyProtocol checks if the data starts with proxy protocol headers
func detectProxyProtocol(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// Check for Proxy Protocol v1 (text-based)
	// Format: "PROXY TCP4 255.255.255.255 255.255.255.255 65535 65535\r\n"
	if len(data) >= len(ProxyProtocolV1Prefix) &&
		bytes.HasPrefix(data, []byte(ProxyProtocolV1Prefix)) {
		// Additional validation: check for proper line ending
		if idx := bytes.Index(data, []byte("\r\n")); idx != -1 {
			line := string(data[:idx])
			parts := strings.Fields(line)
			// Valid proxy protocol v1 should have at least 6 parts
			if len(parts) >= 6 && parts[0] == "PROXY" {
				return true
			}
		}
	}

	// Check for Proxy Protocol v2 (binary)
	// Must have the exact 12-byte signature
	if len(data) >= len(ProxyProtocolV2Prefix) &&
		bytes.Equal(data[:len(ProxyProtocolV2Prefix)], []byte(ProxyProtocolV2Prefix)) {
		// Additional validation: check version and command fields
		if len(data) >= 16 {
			versionCmd := data[12]
			version := versionCmd >> 4
			command := versionCmd & 0xF
			// Version should be 2, command should be 0 (LOCAL) or 1 (PROXY)
			if version == 2 && (command == 0 || command == 1) {
				return true
			}
		}
	}

	return false
}

// processProxyProtocolData parses proxy protocol headers and returns the remaining data
func processProxyProtocolData(data []byte) ([]byte, *ProxyProtocolInfo, error) {
	if len(data) == 0 {
		return data, nil, nil
	}

	reader := bufio.NewReader(bytes.NewReader(data))

	// Try to parse proxy protocol
	if len(data) >= len(ProxyProtocolV1Prefix) &&
		bytes.HasPrefix(data, []byte(ProxyProtocolV1Prefix)) {
		// Parse v1
		proxyInfo, err := parseProxyProtocolV1(reader)
		if err != nil {
			return nil, nil, fmt.Errorf("proxy protocol v1 parse error: %w", err)
		}

		// Find where the proxy protocol header ends
		headerEnd := bytes.Index(data, []byte("\r\n"))
		if headerEnd == -1 {
			return nil, nil, fmt.Errorf("malformed proxy protocol v1 header: missing CRLF")
		}

		// Return remaining data after the header
		remainingData := data[headerEnd+2:]

		// Log what we're returning for debugging
		logger.Printf("Proxy Protocol v1 processed, returning %d bytes of data", len(remainingData))
		if len(remainingData) > 0 {
			// Check if remaining data looks like TLS handshake
			if remainingData[0] == 0x16 { // TLS handshake record type
				logger.Printf("Remaining data appears to be TLS handshake")
			} else if remainingData[0] >= 0x20 && remainingData[0] <= 0x7E {
				// Looks like printable ASCII (HTTP)
				logger.Printf("Remaining data appears to be HTTP: %s", string(remainingData[:min(50, len(remainingData))]))
			}
		}

		return remainingData, proxyInfo, nil

	} else if len(data) >= len(ProxyProtocolV2Prefix) &&
		bytes.Equal(data[:len(ProxyProtocolV2Prefix)], []byte(ProxyProtocolV2Prefix)) {
		// Parse v2
		proxyInfo, err := parseProxyProtocolV2(reader)
		if err != nil {
			return nil, nil, fmt.Errorf("proxy protocol v2 parse error: %w", err)
		}

		// For v2, we need to calculate the header length
		if len(data) < 16 {
			return nil, nil, fmt.Errorf("proxy protocol v2 header too short: %d bytes", len(data))
		}

		// Read the length field (bytes 14-15)
		addrLen := int(data[14])<<8 + int(data[15])
		headerLen := 16 + addrLen

		if len(data) < headerLen {
			return nil, nil, fmt.Errorf("proxy protocol v2 header incomplete: expected %d bytes, got %d", headerLen, len(data))
		}

		// Return remaining data after the header
		remainingData := data[headerLen:]

		// Log what we're returning for debugging
		logger.Printf("Proxy Protocol v2 processed, returning %d bytes of data", len(remainingData))
		if len(remainingData) > 0 {
			// Check if remaining data looks like TLS handshake
			if remainingData[0] == 0x16 { // TLS handshake record type
				logger.Printf("Remaining data appears to be TLS handshake")
			} else if remainingData[0] >= 0x20 && remainingData[0] <= 0x7E {
				// Looks like printable ASCII (HTTP)
				logger.Printf("Remaining data appears to be HTTP: %s", string(remainingData[:min(50, len(remainingData))]))
			}
		}

		return remainingData, proxyInfo, nil
	}

	// No proxy protocol found
	return data, nil, nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
