# Makefile for Zoraxy Proxy Protocol Plugin

# Variables
BINARY_NAME=proxy-protocol
LDFLAGS=-ldflags="-s -w"

# Default target
.PHONY: all
all: clean build

# Clean build artifacts
.PHONY: clean
clean:
	rm -f $(BINARY_NAME)*
	rm -rf dist/

# Download dependencies
.PHONY: deps
deps:
	go mod download
	go mod verify

# Build for current platform
.PHONY: build
build: deps
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Build for Linux AMD64 (most common server architecture)
.PHONY: build-linux
build-linux: deps
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 .

# Build all platforms
.PHONY: build-all
build-all: deps
	mkdir -p dist
	# Linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 .
	# Windows
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .
	# macOS
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .
	# FreeBSD
	GOOS=freebsd GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-freebsd-amd64 .

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
	sudo cp $(BINARY_NAME) /usr/local/bin/

# Create release (requires git tag)
.PHONY: release
release: clean build-all
	cd dist && sha256sum * > checksums.txt



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
