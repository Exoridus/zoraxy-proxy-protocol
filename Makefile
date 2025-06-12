# Makefile for Zoraxy Proxy Protocol Plugin

# Variables
BINARY_NAME=proxy-protocol
DIST_DIR=dist
LDFLAGS=-ldflags="-s -w"

# Default target
.PHONY: all
all: clean build

# Clean build artifacts
.PHONY: clean
clean:
	@if exist "$(BINARY_NAME)*" del /q "$(BINARY_NAME)*" 2>nul || echo.
	@if exist "$(DIST_DIR)" rmdir /s /q "$(DIST_DIR)" 2>nul || echo.

# Download dependencies
.PHONY: deps
deps:
	go mod download
	go mod verify

# Build for current platform
.PHONY: build
build: deps
	@if not exist "$(DIST_DIR)" mkdir "$(DIST_DIR)"
	go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME) .

# Build for Linux AMD64 (most common server architecture)
.PHONY: build-linux
build-linux: deps
	@if not exist "$(DIST_DIR)" mkdir "$(DIST_DIR)"
	set GOOS=linux&& set GOARCH=amd64&& go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .

# Build all platforms
.PHONY: build-all
build-all: deps
	@if not exist "$(DIST_DIR)" mkdir "$(DIST_DIR)"
	@echo Building for Linux AMD64...
	@set GOOS=linux&& set GOARCH=amd64&& go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .
	@echo Building for Linux ARM64...
	@set GOOS=linux&& set GOARCH=arm64&& go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 .
	@echo Building for Windows AMD64...
	@set GOOS=windows&& set GOARCH=amd64&& go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo Building for macOS AMD64...
	@set GOOS=darwin&& set GOARCH=amd64&& go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .
	@echo Building for macOS ARM64...
	@set GOOS=darwin&& set GOARCH=arm64&& go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 .
	@echo Building for FreeBSD AMD64...
	@set GOOS=freebsd&& set GOARCH=amd64&& go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-freebsd-amd64 .

# Run tests
.PHONY: test
test:
	go test -v ./...

# Run linter
.PHONY: lint
lint:
	go vet ./...
	gofmt -s -w .

# Install locally (for development)
.PHONY: install
install: build
	sudo cp $(DIST_DIR)/$(BINARY_NAME) /usr/local/bin/

# Create release (requires git tag)
.PHONY: release
release: clean build-all
	cd $(DIST_DIR) && sha256sum * > checksums.txt



# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all         - Clean and build for current platform"
	@echo "  clean       - Remove build artifacts"
	@echo "  deps        - Download and verify dependencies"
	@echo "  build       - Build for current platform"
	@echo "  build-linux - Build for Linux AMD64"
	@echo "  build-all   - Build for all supported platforms"
	@echo "  test        - Run tests"
	@echo "  lint        - Run linter and formatter"
	@echo "  install     - Install binary locally"
	@echo "  release     - Create release with checksums"

	@echo "  help        - Show this help message"
