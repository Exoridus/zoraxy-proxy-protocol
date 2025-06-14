# Proxy Protocol Plugin for Zoraxy

[![Latest Version](https://img.shields.io/github/v/release/Exoridus/zoraxy-proxy-protocol?style=for-the-badge&label=Latest&logo=github&color=44cc11)](https://github.com/Exoridus/zoraxy-proxy-protocol/releases/latest)
[![Build Status](https://img.shields.io/github/actions/workflow/status/Exoridus/zoraxy-proxy-protocol/build-release.yml?style=for-the-badge&label=Build&logo=github)](https://github.com/Exoridus/zoraxy-proxy-protocol/actions)
[![Code Coverage](https://img.shields.io/codecov/c/github/Exoridus/zoraxy-proxy-protocol?style=for-the-badge&label=Coverage&logo=codecov)](https://codecov.io/gh/Exoridus/zoraxy-proxy-protocol)
![Go Version](https://img.shields.io/github/go-mod/go-version/Exoridus/zoraxy-proxy-protocol/master?filename=src%2Fgo.mod&style=for-the-badge&label=&logo=go&logoColor=%23fff&logoSize=auto)
[![Sponsor](https://img.shields.io/badge/Sponsor-1a1e23?style=for-the-badge&logo=githubsponsors)](https://github.com/sponsors/Exoridus)

A Zoraxy plugin that adds support for the Proxy Protocol (HAProxy compatible), preserving original client IP information when traffic passes through Layer 4 proxies or load balancers.

## ⚠️ **DISCLAIMER - WORK IN PROGRESS**

> **🚧 This project is currently under active development and should be considered experimental.**
> 
> - **AI-Generated Code**: Large portions of this codebase have been generated using AI assistance and may require additional testing and refinement.
> - **Pre-Production**: Not recommended for production environments without thorough testing.
> - **Active Development**: Features, APIs, and functionality may change without notice.
> 
> **Use at your own risk** and always test thoroughly in non-production environments first.

## ✨ Features

- **Proxy Protocol v1 & v2** support (text and binary formats)
- **Automatic detection** of Proxy Protocol headers
- **Client IP preservation** through proxy chains
- **Web-based configuration** interface
- **Real-time enable/disable** without restarts
- **Compatible** with HAProxy, nginx, AWS NLB, and more

## 🚀 Installation

### From GitHub Releases (Recommended)

**Installation Steps:**
1. Download the ZIP package for your platform from [releases](https://github.com/Exoridus/zoraxy-proxy-protocol/releases)
2. Extract the ZIP file to your Zoraxy plugins directory:
   ```bash
   unzip proxy-protocol.zip -d /path/to/zoraxy/plugins/
   ```
3. The plugin is automatically ready - no restart required!

**Platform-Specific Downloads:**

**Linux AMD64:**
```bash
wget -qO- "https://github.com/Exoridus/zoraxy-proxy-protocol/releases/latest/download/proxy-protocol-linux-amd64.zip" | unzip - -d "/path/to/zoraxy/plugins/"
```

**Linux ARM64:**
```bash
wget -qO- "https://github.com/Exoridus/zoraxy-proxy-protocol/releases/latest/download/proxy-protocol-linux-arm64.zip" | unzip - -d "/path/to/zoraxy/plugins/"
```

**Windows AMD64:**
```bash
wget -qO- "https://github.com/Exoridus/zoraxy-proxy-protocol/releases/latest/download/proxy-protocol-windows-amd64.zip" | unzip - -d "/path/to/zoraxy/plugins/"
```

**macOS Intel:**
```bash
wget -qO- "https://github.com/Exoridus/zoraxy-proxy-protocol/releases/latest/download/proxy-protocol-darwin-amd64.zip" | unzip - -d "/path/to/zoraxy/plugins/"
```

**macOS Apple Silicon:**
```bash
wget -qO- "https://github.com/Exoridus/zoraxy-proxy-protocol/releases/latest/download/proxy-protocol-darwin-arm64.zip" | unzip - -d "/path/to/zoraxy/plugins/"
```

**FreeBSD AMD64:**
```bash
wget -qO- "https://github.com/Exoridus/zoraxy-proxy-protocol/releases/latest/download/proxy-protocol-freebsd-amd64.zip" | unzip - -d "/path/to/zoraxy/plugins/"
```

> **Note:** The ZIP contains the correct directory structure with executable permissions preserved.

### Directory Structure

After installation, your Zoraxy directory will look like this:
```
/path/to/zoraxy/
├── zoraxy                ← Main Zoraxy executable
├── plugins/
│   └── proxy-protocol/
│       ├── proxy-protocol ← Plugin executable (auto-extracted)
│       └── icon.png      ← Plugin icon (auto-extracted)
└── ... (other files)
```

### Using the Makefile (Advanced)

If you have the source code, you can use the built-in install command:
```bash
git clone https://github.com/Exoridus/zoraxy-proxy-protocol.git
cd zoraxy-proxy-protocol

# Auto-detect Zoraxy installation and install
make install

# Or specify custom Zoraxy directory
make install ZORAXY_DIR=/path/to/zoraxy
```

### From Source

```bash
git clone https://github.com/Exoridus/zoraxy-proxy-protocol.git
cd zoraxy-proxy-protocol
make release  # Creates ZIP packages in dist/
# Then use the ZIP installation method above
```

## 📖 How It Works

The plugin intercepts incoming connections to detect Proxy Protocol headers, extracts the original client information (IP address, port), and makes it available to Zoraxy's reverse proxy engine through standard HTTP headers.

### Supported Headers

When Proxy Protocol is detected, the plugin sets:
- `X-Forwarded-For`: Original client IP
- `X-Real-IP`: Original client IP  
- `X-Forwarded-Port`: Original client port

## ⚙️ Configuration

Access the plugin configuration through the Zoraxy admin interface:

1. Navigate to **Plugins** → **Proxy Protocol**
2. Toggle **Enable Proxy Protocol Support**
3. Monitor status and connections

### API Endpoints

The plugin exposes REST endpoints for programmatic control:

#### GET `/ui/api/status`
Returns current plugin status and configuration.

**Response:**
```json
{
  "status": "Enabled|Disabled",
  "enabled": true,
  "version": "1.0.0"
}
```

#### POST `/ui/api/toggle`
Enable or disable proxy protocol processing.

**Request:**
```json
{
  "enabled": true
}
```

**Response:**
```json
{
  "result": "success",
  "enabled": true
}
```

## 🔧 Proxy Configuration Examples

### HAProxy
```haproxy
backend zoraxy
    mode tcp
    option tcp-check
    server zoraxy1 192.168.1.100:80 send-proxy check
```

### nginx
```nginx
upstream zoraxy {
    server 192.168.1.100:80;
}

server {
    listen 80;
    proxy_protocol on;
    location / {
        proxy_pass http://zoraxy;
        proxy_protocol on;
    }
}
```

### AWS Network Load Balancer
Enable "Proxy Protocol v2" in the target group settings.

## 🔍 Compatibility

- **Zoraxy**: v3.1.9+ (tested with v3.2.3)
- **Go**: 1.20+ (for building from source)
- **Platforms**: Linux, Windows, macOS, FreeBSD (amd64, arm64)

## 🐛 Troubleshooting

### Plugin Not Loading
- Check the binary is at: `{zoraxy-directory}/plugins/proxy-protocol/proxy-protocol`
- Verify binary permissions: `chmod +x plugins/proxy-protocol/proxy-protocol`
- Check Zoraxy logs for error messages
- Ensure binary and directory are both named `proxy-protocol`

### Proxy Protocol Not Working
- Confirm upstream proxy sends Proxy Protocol headers
- Check plugin is enabled in the web interface
- Verify network connectivity between proxy and Zoraxy

### Client IP Not Preserved
- Ensure Proxy Protocol is enabled on upstream proxy
- Check that traffic is actually passing through the proxy
- Verify HTTP headers are being set correctly

## 📄 License

This project is open source under the MIT License.

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, coding conventions, and the contribution process.

## 🎯 About Zoraxy

[Zoraxy](https://github.com/tobychui/zoraxy) is a general-purpose HTTP reverse proxy and forwarding tool written in Go, designed for simplicity and ease of use in homelab and small business environments.
