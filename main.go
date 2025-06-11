package main

import (
	"embed"
	"fmt"
	"net/http"
	"strconv"

	plugin "plugins.zoraxy.aroz.org/zoraxy/proxy-protocol/mod/zoraxy_plugin"
)

const (
	PLUGIN_ID = "de.codexo.proxyprotocol"
	UI_PATH   = "/ui"
	WEB_ROOT  = "/www"
)

//go:embed www/*
var content embed.FS

func main() {
	runtimeCfg, err := plugin.ServeAndRecvSpec(&plugin.IntroSpect{
		ID:            PLUGIN_ID,
		Name:          "Proxy Protocol Support",
		Author:        "Zoraxy Community",
		AuthorContact: "info@example.com",
		Description:   "Adds support for Proxy Protocol (HAProxy compatible)",
		URL:           "https://github.com/zoraxy-proxy-protocol",
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

	fmt.Println("Proxy Protocol Plugin started at http://127.0.0.1:" + strconv.Itoa(runtimeCfg.Port))
	err = http.ListenAndServe("127.0.0.1:"+strconv.Itoa(runtimeCfg.Port), nil)
	if err != nil {
		panic(err)
	}
}

// Core plugin functionality - these are the only endpoints that matter
func handleProxyProtocolSniff(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement actual proxy protocol detection logic
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleProxyProtocolIngress(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement actual proxy protocol processing logic
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Proxy Protocol Handler"))
}
