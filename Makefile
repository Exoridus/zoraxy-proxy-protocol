# Makefile for Zoraxy Proxy Protocol Plugin

# Variables
BINARY_NAME=proxy-protocol
DIST_DIR=dist
SRC_DIR=src

# Check if VERSION is provided (except for utility targets)
UTILITY_TARGETS=clean deps test lint help
ifeq ($(filter $(UTILITY_TARGETS),$(MAKECMDGOALS)),)
ifndef VERSION
$(error VERSION is required. Usage: make <target> VERSION=x.y.z)
endif
endif

# Parse version components (only if VERSION is set)
ifneq ($(VERSION),)
VERSION_CLEAN=$(shell echo '$(VERSION)' | sed 's/^v//')
VERSION_MAJOR=$(shell echo '$(VERSION_CLEAN)' | cut -d. -f1)
VERSION_MINOR=$(shell echo '$(VERSION_CLEAN)' | cut -d. -f2)
VERSION_PATCH=$(shell echo '$(VERSION_CLEAN)' | cut -d. -f3 | cut -d- -f1)

# Ensure version components are not empty
VERSION_MAJOR:=$(if $(VERSION_MAJOR),$(VERSION_MAJOR),1)
VERSION_MINOR:=$(if $(VERSION_MINOR),$(VERSION_MINOR),0)
VERSION_PATCH:=$(if $(VERSION_PATCH),$(VERSION_PATCH),0)

LDFLAGS=-ldflags="-s -w -X main.versionMajor=$(VERSION_MAJOR) -X main.versionMinor=$(VERSION_MINOR) -X main.versionPatch=$(VERSION_PATCH)"
endif

# Default target
.PHONY: all $(UTILITY_TARGETS)
all: clean build-all

# Utility targets (no VERSION required)
clean:
	rm -f $(BINARY_NAME)*
	rm -rf $(DIST_DIR)

deps:
	cd $(SRC_DIR) && go mod download && go mod verify

test:
	cd $(SRC_DIR) && go test -v ./...

lint:
	cd $(SRC_DIR) && go vet ./... && gofmt -s -w .

help:
	@echo "Zoraxy Proxy Protocol Plugin - Build System"
	@echo "==========================================="
	@echo ""
	@echo "Build targets (require VERSION):"
	@echo "  build       - Build for current platform"
	@echo "  build-linux - Build for Linux AMD64"
	@echo "  build-all   - Build for all platforms"
	@echo "  package     - Create ZIP packages"
	@echo "  install     - Install locally"
	@echo "  release     - Complete release build"
	@echo ""
	@echo "Utility targets (no VERSION required):"
	@echo "  clean       - Remove build artifacts"
	@echo "  deps        - Download dependencies"
	@echo "  test        - Run tests"
	@echo "  lint        - Run linter and formatter"
	@echo "  help        - Show this help"
	@echo ""
	@echo "Usage examples:"
	@echo "  make build VERSION=1.0.0"
	@echo "  make release VERSION=1.2.3"
	@echo "  make clean && make test"

# Build targets
.PHONY: build build-linux build-all package install release

build: deps
	mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-current/$(BINARY_NAME)
	cd $(SRC_DIR) && go build $(LDFLAGS) -o ../$(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-current/$(BINARY_NAME)/$(BINARY_NAME) .
	cp $(SRC_DIR)/icon.png $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-current/$(BINARY_NAME)/

build-linux: deps
	mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64/$(BINARY_NAME)
	cd $(SRC_DIR) && GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o ../$(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64/$(BINARY_NAME)/$(BINARY_NAME) .
	cp $(SRC_DIR)/icon.png $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64/$(BINARY_NAME)/

build-all: deps
	@echo "Building for all platforms..."

	@echo "→ Linux AMD64"
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64/$(BINARY_NAME)
	@cd $(SRC_DIR) && GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o ../$(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64/$(BINARY_NAME)/$(BINARY_NAME) .
	@cp $(SRC_DIR)/icon.png $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64/$(BINARY_NAME)/

	@echo "→ Linux ARM64"
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64/$(BINARY_NAME)
	@cd $(SRC_DIR) && GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o ../$(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64/$(BINARY_NAME)/$(BINARY_NAME) .
	@cp $(SRC_DIR)/icon.png $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64/$(BINARY_NAME)/

	@echo "→ Windows AMD64"
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64/$(BINARY_NAME)
	@cd $(SRC_DIR) && GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o ../$(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64/$(BINARY_NAME)/$(BINARY_NAME).exe .
	@cp $(SRC_DIR)/icon.png $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64/$(BINARY_NAME)/

	@echo "→ macOS AMD64"
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64/$(BINARY_NAME)
	@cd $(SRC_DIR) && GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o ../$(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64/$(BINARY_NAME)/$(BINARY_NAME) .
	@cp $(SRC_DIR)/icon.png $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64/$(BINARY_NAME)/

	@echo "→ macOS ARM64"
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64/$(BINARY_NAME)
	@cd $(SRC_DIR) && GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o ../$(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64/$(BINARY_NAME)/$(BINARY_NAME) .
	@cp $(SRC_DIR)/icon.png $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64/$(BINARY_NAME)/

	@echo "→ FreeBSD AMD64"
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-freebsd-amd64/$(BINARY_NAME)
	@cd $(SRC_DIR) && GOOS=freebsd GOARCH=amd64 go build $(LDFLAGS) -o ../$(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-freebsd-amd64/$(BINARY_NAME)/$(BINARY_NAME) .
	@cp $(SRC_DIR)/icon.png $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-freebsd-amd64/$(BINARY_NAME)/

	@echo "✓ All builds completed"

package: build-all
	@echo "Creating plugin packages..."
	@cd $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64 && zip -r ../$(BINARY_NAME)-$(VERSION)-linux-amd64.zip $(BINARY_NAME)
	@cd $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64 && zip -r ../$(BINARY_NAME)-$(VERSION)-linux-arm64.zip $(BINARY_NAME)
	@cd $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64 && zip -r ../$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)
	@cd $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64 && zip -r ../$(BINARY_NAME)-$(VERSION)-darwin-amd64.zip $(BINARY_NAME)
	@cd $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64 && zip -r ../$(BINARY_NAME)-$(VERSION)-darwin-arm64.zip $(BINARY_NAME)
	@cd $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-freebsd-amd64 && zip -r ../$(BINARY_NAME)-$(VERSION)-freebsd-amd64.zip $(BINARY_NAME)
	@echo "✓ All packages created"

install: build
	sudo cp $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-current/$(BINARY_NAME)/$(BINARY_NAME) /usr/local/bin/

release: clean package
	@echo "✓ Release $(VERSION) ready in $(DIST_DIR)/"
