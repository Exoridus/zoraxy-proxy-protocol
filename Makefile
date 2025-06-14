# Makefile for Zoraxy Proxy Protocol Plugin

# Variables
BINARY_NAME=proxy-protocol
DIST_DIR=dist
SRC_DIR=src

# Auto-detect version from git tags if not provided
ifndef VERSION
GIT_TAG=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "")
ifneq ($(GIT_TAG),)
    VERSION=$(GIT_TAG)
    $(info Auto-detected version from git tag: $(VERSION))
endif
endif

# Check if VERSION is provided (except for utility targets)
UTILITY_TARGETS=clean help
ifeq ($(filter $(UTILITY_TARGETS),$(MAKECMDGOALS)),)
ifndef VERSION
    $(error No VERSION provided and no git tags found. Usage: make <target> VERSION=x.y.z or create a git tag)
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

help:
	@echo "Zoraxy Proxy Protocol Plugin - Build System"
	@echo "==========================================="
	@echo ""
	@echo "Version handling:"
	@echo "  - Auto-detects version from latest git tag"
	@echo "  - Or specify explicitly: VERSION=x.y.z"
	@echo ""
	@echo "Build targets:"
	@echo "  build           - Build for current platform (auto-detected)"
	@echo "                    Optional: PLATFORM=linux ARCH=amd64"
	@echo "  build-all       - Build for all platforms"
	@echo ""
	@echo "Package and release:"
	@echo "  package         - Create ZIP packages"
	@echo "  release         - Complete release build"
	@echo ""
	@echo "Utility targets:"
	@echo "  clean           - Remove build artifacts"
	@echo "  help            - Show this help"
	@echo ""
	@echo "Supported platforms:"
	@echo "  linux/amd64, linux/arm64, windows/amd64"
	@echo "  darwin/amd64, darwin/arm64, freebsd/amd64"
	@echo ""
	@echo "Usage examples:"
	@echo "  make build                                             # Auto-detect version & platform"
	@echo "  make build VERSION=1.0.0                              # Explicit version, auto-detect platform"
	@echo "  make build PLATFORM=linux ARCH=amd64                  # Auto-detect version, explicit platform"
	@echo "  make build VERSION=1.0.0 PLATFORM=linux ARCH=amd64   # All explicit"
	@echo "  make build-all                                         # Auto-detect version, all platforms"
	@echo "  make release VERSION=1.2.3                            # Explicit version release"

# Generic build target - auto-detects platform and version if not specified
.PHONY: build build-all package release
build:
	$(eval PLATFORM := $(or $(PLATFORM),$(shell cd $(SRC_DIR) && go env GOOS 2>/dev/null || echo linux)))
	$(eval ARCH := $(or $(ARCH),$(shell cd $(SRC_DIR) && go env GOARCH 2>/dev/null || echo amd64)))
	@echo "→ Building $(PLATFORM)/$(ARCH) version $(VERSION)"
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-$(PLATFORM)-$(ARCH)/$(BINARY_NAME)
	@cd $(SRC_DIR) && GOOS=$(PLATFORM) GOARCH=$(ARCH) go build $(LDFLAGS) -o ../$(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-$(PLATFORM)-$(ARCH)/$(BINARY_NAME)/$(BINARY_NAME)$(if $(filter windows,$(PLATFORM)),.exe) .
	@cp $(SRC_DIR)/icon.png $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-$(PLATFORM)-$(ARCH)/$(BINARY_NAME)/
	@echo "✓ $(PLATFORM)/$(ARCH) version $(VERSION) build completed"

# Build all platforms using the generic build target
build-all:
	@$(MAKE) build VERSION=$(VERSION) PLATFORM=linux ARCH=amd64
	@$(MAKE) build VERSION=$(VERSION) PLATFORM=linux ARCH=arm64
	@$(MAKE) build VERSION=$(VERSION) PLATFORM=windows ARCH=amd64
	@$(MAKE) build VERSION=$(VERSION) PLATFORM=darwin ARCH=amd64
	@$(MAKE) build VERSION=$(VERSION) PLATFORM=darwin ARCH=arm64
	@$(MAKE) build VERSION=$(VERSION) PLATFORM=freebsd ARCH=amd64
	@echo "✓ All builds completed for version $(VERSION)"

package: build-all
	@echo "Creating plugin packages for version $(VERSION)..."
	@cd $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64 && zip -r ../$(BINARY_NAME)-$(VERSION)-linux-amd64.zip $(BINARY_NAME)
	@cd $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64 && zip -r ../$(BINARY_NAME)-$(VERSION)-linux-arm64.zip $(BINARY_NAME)
	@cd $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64 && zip -r ../$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)
	@cd $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64 && zip -r ../$(BINARY_NAME)-$(VERSION)-darwin-amd64.zip $(BINARY_NAME)
	@cd $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64 && zip -r ../$(BINARY_NAME)-$(VERSION)-darwin-arm64.zip $(BINARY_NAME)
	@cd $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-freebsd-amd64 && zip -r ../$(BINARY_NAME)-$(VERSION)-freebsd-amd64.zip $(BINARY_NAME)
	@echo "✓ All packages created for version $(VERSION)"

release: clean package
	@echo "✓ Release $(VERSION) ready in $(DIST_DIR)/"
