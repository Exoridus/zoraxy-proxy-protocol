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

func main() {
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

	status := "Disabled"
	if enabled {
		status = "Enabled"
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
	config.mu.RLock()
	enabled := config.Enabled
	config.mu.RUnlock()

	if !enabled {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("DISABLED"))
		return
	}

	// Get the raw connection data from the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Printf("Error reading request body for sniffing: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ERROR"))
		return
	}

	// Check if this looks like proxy protocol data
	if detectProxyProtocol(body) {
		logger.Printf("Proxy Protocol detected in connection")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("PROXY_PROTOCOL"))
		return
	}

	// Not proxy protocol data
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("NORMAL"))
}

func handleProxyProtocolIngress(w http.ResponseWriter, r *http.Request) {
	config.mu.RLock()
	enabled := config.Enabled
	config.mu.RUnlock()

	if !enabled {
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

	// Read the raw connection data
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Printf("Error reading proxy protocol data: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Read Error"))
		return
	}

	// Process the proxy protocol data
	processedData, proxyInfo, err := processProxyProtocolData(body)
	if err != nil {
		logger.Printf("Error processing proxy protocol: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Parse Error"))
		return
	}

	if proxyInfo != nil {
		logger.Printf("Proxy Protocol parsed: %s:%d -> %s:%d",
			proxyInfo.SourceAddr, proxyInfo.SourcePort,
			proxyInfo.DestAddr, proxyInfo.DestPort)

		// Store the proxy info for later use
		connectionsMutex.Lock()
		if conn, exists := activeConnections[connID]; exists {
			conn.ProxyInfo = proxyInfo
		}
		connectionsMutex.Unlock()

		// Set headers for the processed request
		w.Header().Set("X-Forwarded-For", proxyInfo.SourceAddr)
		w.Header().Set("X-Real-IP", proxyInfo.SourceAddr)
		w.Header().Set("X-Forwarded-Port", strconv.Itoa(proxyInfo.SourcePort))
	}

	// Return the processed data (without proxy protocol headers)
	w.WriteHeader(http.StatusOK)
	w.Write(processedData)
}

// detectProxyProtocol checks if the data starts with proxy protocol headers
func detectProxyProtocol(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// Check for Proxy Protocol v1 (text-based)
	if bytes.HasPrefix(data, []byte(ProxyProtocolV1Prefix)) {
		return true
	}

	// Check for Proxy Protocol v2 (binary)
	if len(data) >= len(ProxyProtocolV2Prefix) &&
		bytes.HasPrefix(data, []byte(ProxyProtocolV2Prefix)) {
		return true
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
	if bytes.HasPrefix(data, []byte(ProxyProtocolV1Prefix)) {
		// Parse v1
		proxyInfo, err := parseProxyProtocolV1(reader)
		if err != nil {
			return nil, nil, err
		}

		// Find where the proxy protocol header ends
		headerEnd := bytes.Index(data, []byte("\r\n"))
		if headerEnd == -1 {
			return nil, nil, fmt.Errorf("malformed proxy protocol v1 header")
		}

		// Return remaining data after the header
		remainingData := data[headerEnd+2:]
		return remainingData, proxyInfo, nil

	} else if bytes.HasPrefix(data, []byte(ProxyProtocolV2Prefix)) {
		// Parse v2
		proxyInfo, err := parseProxyProtocolV2(reader)
		if err != nil {
			return nil, nil, err
		}

		// For v2, we need to calculate the header length
		if len(data) < 16 {
			return nil, nil, fmt.Errorf("proxy protocol v2 header too short")
		}

		// Read the length field (bytes 14-15)
		headerLen := 16 + int(data[14])<<8 + int(data[15])
		if len(data) < headerLen {
			return nil, nil, fmt.Errorf("proxy protocol v2 header incomplete")
		}

		// Return remaining data after the header
		remainingData := data[headerLen:]
		return remainingData, proxyInfo, nil

	}

	// No proxy protocol found
	return data, nil, nil
}
