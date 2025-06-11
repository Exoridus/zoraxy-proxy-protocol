package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	plugin "plugins.zoraxy.aroz.org/zoraxy/proxy-protocol/mod/zoraxy_plugin"
)

const (
	PLUGIN_ID = "de.codexo.proxyprotocol"
	UI_PATH   = "/ui"
	WEB_ROOT  = "/www"
)

// Version information - set via ldflags during build
var version = "dev"

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

func main() {
	runtimeCfg, err := plugin.ServeAndRecvSpec(&plugin.IntroSpect{
		ID:            PLUGIN_ID,
		Name:          "Proxy Protocol Support",
		Author:        "Exoridus",
		AuthorContact: "https://github.com/Exoridus",
		Description:   "Adds support for Proxy Protocol (HAProxy compatible)",
		URL:           "https://github.com/Exoridus/zoraxy-proxy-protocol",
		Type:          plugin.PluginType_Router,
		VersionMajor:  1,
		VersionMinor:  0,
		VersionPatch:  0,

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
		fmt.Println("This is a plugin for Zoraxy and should not be run standalone\nVisit zoraxy.aroz.org to download Zoraxy.")
		panic(err)
	}

	// Create embedded web router for UI
	embedWebRouter := plugin.NewPluginEmbedUIRouter(PLUGIN_ID, &content, WEB_ROOT, UI_PATH)
	embedWebRouter.RegisterTerminateHandler(func() {
		fmt.Println("Proxy Protocol Plugin terminated")
	}, nil)
	embedWebRouter.AttachHandlerToMux(nil)

	// Register core plugin endpoints (required by Zoraxy)
	http.HandleFunc("/proxy_protocol_sniff", handleProxyProtocolSniff)
	http.HandleFunc("/proxy_protocol_handler", handleProxyProtocolIngress)

	// Register UI API endpoints
	http.HandleFunc("/ui/api/status", handleAPIStatus)
	http.HandleFunc("/ui/api/toggle", handleAPIToggle)

	fmt.Println("Proxy Protocol Plugin started at http://127.0.0.1:" + strconv.Itoa(runtimeCfg.Port))
	err = http.ListenAndServe("127.0.0.1:"+strconv.Itoa(runtimeCfg.Port), nil)
	if err != nil {
		panic(err)
	}
}

// API Handlers
func handleAPIStatus(w http.ResponseWriter, r *http.Request) {
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

	response := StatusResponse{
		Status:  status,
		Enabled: enabled,
		Version: version,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleAPIToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ToggleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	config.mu.Lock()
	config.Enabled = req.Enabled
	config.mu.Unlock()

	fmt.Printf("Proxy Protocol support %s\n", map[bool]string{true: "enabled", false: "disabled"}[req.Enabled])

	response := ToggleResponse{
		Result:  "success",
		Enabled: req.Enabled,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Core plugin functionality - these are the only endpoints that matter
func handleProxyProtocolSniff(w http.ResponseWriter, r *http.Request) {
	config.mu.RLock()
	enabled := config.Enabled
	config.mu.RUnlock()

	if !enabled {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("DISABLED"))
		return
	}

	// TODO: Implement actual proxy protocol detection logic
	// For now, just indicate that we're enabled and ready
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
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

	// TODO: Implement actual proxy protocol processing logic
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Proxy Protocol Handler"))
}
