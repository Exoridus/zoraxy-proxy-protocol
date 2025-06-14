# Proxy Protocol Plugin for Zoraxy

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

1. Download the binary for your platform from [releases](https://github.com/Exoridus/zoraxy-proxy-protocol/releases)
2. Go to your Zoraxy installation directory (where the `zoraxy` executable is located)
3. Create the directory: `mkdir -p plugins/proxy-protocol`
4. Copy and rename the binary: `cp proxy-protocol-* plugins/proxy-protocol/proxy-protocol`
5. Make it executable (Linux/macOS): `chmod +x plugins/proxy-protocol/proxy-protocol`
6. Restart Zoraxy and configure via **Plugins** → **Proxy Protocol**

> **Note:** Place the executable at `{zoraxy-directory}/plugins/proxy-protocol/proxy-protocol`

### Directory Structure

After installation, your Zoraxy directory will look like this:
```
/path/to/zoraxy/
├── zoraxy                ← Main Zoraxy executable
├── plugins/
│   └── proxy-protocol/
│       └── proxy-protocol ← Plugin executable goes here
└── ... (other files)
```

### From Source

```bash
git clone https://github.com/Exoridus/zoraxy-proxy-protocol.git
cd zoraxy-proxy-protocol
make build-linux  # or make build for current platform
# Then follow the installation steps above
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
