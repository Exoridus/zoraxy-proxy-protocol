# Proxy Protocol Plugin for Zoraxy

This plugin adds support for the Proxy Protocol (HAProxy compatible) to Zoraxy, allowing it to preserve client connection information when traffic passes through a Layer 4 proxy or load balancer.

## Features

- Support for Proxy Protocol v1 (text-based) and v2 (binary)
- Automatic detection of Proxy Protocol headers
- Extraction of original client IP and port information
- Setting of appropriate X-Forwarded-For and X-Real-IP headers
- Compatible with HAProxy, nginx, and other proxy implementations

## Installation

### From GitHub Releases (Recommended)

1. Download the appropriate binary for your platform from the [releases page](https://github.com/your-username/zoraxy-proxy-protocol/releases)
2. Place the binary in your Zoraxy plugins directory
3. Make it executable: `chmod +x proxy-protocol`
4. Enable the plugin through the Zoraxy admin interface

### From Source

See [BUILD.md](BUILD.md) for detailed build instructions.

## Local Development

This project uses a Makefile for common development tasks:

```bash
# Quick development build for current platform
make build

# Build specifically for Linux (most common server platform)
make build-linux

# Build for all supported platforms
make build-all

# Run tests and linting
make test lint

# Install locally for testing
make install

# Clean build artifacts
make clean

# Show all available targets
make help
```

### Development Workflow

1. **Clone the repository:**
```bash
git clone https://github.com/your-username/zoraxy-proxy-protocol.git
cd zoraxy-proxy-protocol
```

2. **Make your changes** to the source code

3. **Test locally:**
```bash
make test lint
make build-linux
```

4. **Test the plugin** with your Zoraxy instance

5. **Commit and push** your changes

6. **Create a release** by pushing a version tag:
```bash
git tag v1.0.1
git push origin v1.0.1
```

The GitHub workflow will automatically build multi-platform binaries and create a release.

## How it Works

The plugin implements the Proxy Protocol specification, which allows a proxy or load balancer to convey connection information to the backend server. When enabled, the plugin:

1. Detects incoming Proxy Protocol headers
2. Parses the client connection information (IP, port)
3. Sets the appropriate HTTP headers for downstream processing
4. Preserves the original client information through the proxy chain

## Configuration

The plugin provides a web interface for configuration accessible through the Zoraxy admin panel. You can:

- Enable/disable Proxy Protocol support
- View the current status of the plugin
- Monitor processed connections

## Compatibility

**Zoraxy Versions:** Tested with 3.1.9 and 3.2.3

### Proxy Protocol Compatibility

This plugin is compatible with:
- HAProxy
- nginx (with proxy_protocol)
- AWS Network Load Balancer
- Other Proxy Protocol v1/v2 implementations

## Technical Details

The plugin works at the network level by:
- Intercepting incoming connections
- Reading and parsing Proxy Protocol headers
- Modifying the connection context to include original client information
- Passing the enhanced connection information to Zoraxy's reverse proxy engine

## License

This project is open source and available under the MIT License.

## 🎨 Contributing with Gitmoji

This project uses [Gitmoji](https://gitmoji.dev/) for commit messages to keep the git history clean and visually appealing.

### Quick Setup

```bash
# Install gitmoji-cli
npm install -g gitmoji-cli

# Or use without installation
npx gitmoji -c
```

### Common Emojis

| Emoji | Description | Example |
|-------|-------------|---------|
| 🎉 | Initial commit | `🎉 initial commit` |
| ✨ | New feature | `✨ add proxy protocol v2 support` |
| 🐛 | Bug fix | `🐛 fix header parsing issue` |
| 📝 | Documentation | `📝 update installation guide` |
| ♻️ | Refactor | `♻️ extract parsing logic` |
| ⚡ | Performance | `⚡ optimize connection handling` |
| ✅ | Tests | `✅ add parser unit tests` |
| 🔧 | Config | `🔧 update Makefile` |

### Usage

```bash
# Interactive mode (recommended)
gitmoji -c

# Quick commit
gitmoji -m "✨ add new feature"
```

Release notes are automatically organized by emoji type! 🚀
