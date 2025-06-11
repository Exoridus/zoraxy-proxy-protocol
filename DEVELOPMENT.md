# Development Guide

## Local Development Setup

### Prerequisites

- **Go 1.20+** installed on your system
- **Make** (optional, but recommended for easier commands)
- **Git** for version control

### Quick Start

1. **Clone and setup:**
```bash
git clone https://github.com/your-username/zoraxy-proxy-protocol.git
cd zoraxy-proxy-protocol
go mod tidy
```

2. **Build for your development platform:**
```bash
# Quick build for current platform
make build

# Or build specifically for Linux (if developing on Windows/Mac but deploying to Linux)
make build-linux
```

3. **Test your changes:**
```bash
make test lint
```

## Development Workflow

### Making Changes

1. **Edit the source code** (mainly `main.go` and `proxy_protocol_utils.go`)

2. **Test locally:**
```bash
# Run tests
make test

# Check code quality
make lint

# Build for testing
make build-linux
```

3. **Test with Zoraxy:**
```bash
# Copy to your test Zoraxy instance
scp proxy-protocol-linux-amd64 user@your-server:/path/to/zoraxy/plugins/proxy-protocol

# SSH to server and restart Zoraxy plugin system
```

### Release Process

**Local releases are not needed** - the GitHub workflow handles this automatically:

1. **Commit your changes:**
```bash
git add .
git commit -m "Your change description"
git push origin main
```

2. **Create a version tag when ready for release:**
```bash
git tag v1.0.1
git push origin v1.0.1
```

3. **GitHub automatically:**
   - Builds binaries for all platforms
   - Runs tests
   - Creates a GitHub release
   - Uploads binaries with checksums

### Useful Make Targets

```bash
make build       # Build for current platform
make build-linux # Build for Linux (most common deployment)
make test        # Run all tests
make lint        # Run code quality checks
make clean       # Remove build artifacts
make help        # Show all available commands
```

### Project Structure

```
zoraxy-proxy-protocol/
├── main.go                    # Plugin entry point and configuration
├── proxy_protocol_utils.go    # Core Proxy Protocol implementation
├── go.mod                     # Go module definition
├── www/                       # Web UI files
├── mod/zoraxy_plugin/         # Zoraxy plugin interface
├── Makefile                   # Build automation
├── .github/workflows/         # CI/CD pipeline
└── README.md                  # User documentation
```

### Testing Your Plugin

1. **Unit Testing:**
```bash
make test
```

2. **Integration Testing:**
   - Deploy to a test Zoraxy instance
   - Configure HAProxy or nginx with proxy_protocol enabled
   - Test that client IPs are properly forwarded

3. **Manual Testing:**
   - Check plugin loads in Zoraxy admin interface
   - Verify configuration UI works
   - Test with real proxy protocol traffic

### Debugging Tips

- **Check Zoraxy logs** for plugin loading errors
- **Use the plugin's debug endpoints** for status information
- **Test proxy protocol headers** with tools like `socat` or `nc`
- **Verify plugin permissions** and executable bit

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes and test
4. Commit: `git commit -am 'Add some feature'`
5. Push: `git push origin feature-name`
6. Create a Pull Request

The maintainers will review and merge if appropriate. 
