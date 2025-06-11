# Building Proxy Protocol Plugin for Zoraxy

## Local Build on Debian 12 LXC Container

### Prerequisites

1. **Install Go 1.20 or later:**
```bash
# Update package list
sudo apt update

# Install Go (Debian 12 includes Go 1.19+, but we need 1.20+)
wget https://golang.org/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export GOBIN=$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

2. **Install build dependencies:**
```bash
sudo apt install -y git build-essential
```

### Building the Plugin

1. **Clone and build:**
```bash
# Navigate to project directory
cd zoraxy-proxy-protocol

# Download dependencies
go mod tidy

# Build for Linux amd64 (your LXC container)
go build -o proxy-protocol-linux-amd64 .

# Or build with optimization flags
go build -ldflags="-s -w" -o proxy-protocol-linux-amd64 .
```

2. **Cross-compile for other platforms (optional):**
```bash
# Linux ARM64 (for ARM-based servers)
GOOS=linux GOARCH=arm64 go build -o proxy-protocol-linux-arm64 .

# Windows amd64
GOOS=windows GOARCH=amd64 go build -o proxy-protocol-windows-amd64.exe .

# macOS amd64
GOOS=darwin GOARCH=amd64 go build -o proxy-protocol-darwin-amd64 .

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o proxy-protocol-darwin-arm64 .
```

### Installation in Zoraxy

1. **Copy the binary to your Zoraxy plugins directory:**
```bash
# Assuming your Zoraxy is installed in /opt/zoraxy
sudo cp proxy-protocol-linux-amd64 /opt/zoraxy/plugins/proxy-protocol

# Make it executable
sudo chmod +x /opt/zoraxy/plugins/proxy-protocol

# Set proper ownership (adjust user as needed)
sudo chown zoraxy:zoraxy /opt/zoraxy/plugins/proxy-protocol
```

2. **Restart Zoraxy and enable the plugin through the web interface**

### Build Flags Explanation

- `-ldflags="-s -w"`: Strip debugging info and symbol table (smaller binary)
- `-o filename`: Specify output filename
- `GOOS=linux GOARCH=amd64`: Target Linux on x86_64 architecture

### Troubleshooting

- **Permission denied**: Ensure the binary has execute permissions
- **Plugin not loading**: Check Zoraxy logs for error messages
- **Network issues**: Verify the LXC container has proper network access 
