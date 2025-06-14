# Makefile for Zoraxy Proxy Protocol Plugin

# Variables
BINARY_NAME=proxy-protocol
DIST_DIR=dist
SRC_DIR=src

# Test variables
TEST_TIMEOUT=30s
COVERAGE_DIR=$(DIST_DIR)/coverage

# Function to get version (only called when needed)
define get_version
$(if $(VERSION),$(VERSION),$(shell git describe --tags --abbrev=0 2>/dev/null || echo ""))
endef

# Function to validate version (only called when needed)
define validate_version
$(if $(call get_version),,$(error No VERSION provided and no git tags found. Usage: make <target> VERSION=x.y.z or create a git tag))
endef

# Function to parse version components (only called when needed)
define parse_version
$(eval CURRENT_VERSION := $(call get_version))
$(eval VERSION_CLEAN := $(shell echo '$(CURRENT_VERSION)' | sed 's/^v//'))
$(eval VERSION_MAJOR := $(shell echo '$(VERSION_CLEAN)' | cut -d. -f1))
$(eval VERSION_MINOR := $(shell echo '$(VERSION_CLEAN)' | cut -d. -f2))
$(eval VERSION_PATCH := $(shell echo '$(VERSION_CLEAN)' | cut -d. -f3 | cut -d- -f1))
$(eval VERSION_MAJOR := $(if $(VERSION_MAJOR),$(VERSION_MAJOR),1))
$(eval VERSION_MINOR := $(if $(VERSION_MINOR),$(VERSION_MINOR),0))
$(eval VERSION_PATCH := $(if $(VERSION_PATCH),$(VERSION_PATCH),0))
$(eval LDFLAGS := -ldflags="-s -w -X main.versionMajor=$(VERSION_MAJOR) -X main.versionMinor=$(VERSION_MINOR) -X main.versionPatch=$(VERSION_PATCH)")
endef

# Default target - show help when no target specified
.PHONY: all help clean test test-verbose test-coverage test-short bench

# Default target
all: help

clean:
	rm -f $(BINARY_NAME)*
	rm -rf $(DIST_DIR)

help:
	@echo "\033[1;34m╔═════════════════════════════════════════════════════╗\033[0m"
	@echo "\033[1;34m║\033[0m     \033[1;37mZoraxy Proxy Protocol Plugin - Build System\033[0m     \033[1;34m║\033[0m"
	@echo "\033[1;34m╚═════════════════════════════════════════════════════╝\033[0m"
	@echo ""
	@echo "\033[1;33m🧪 TESTING\033[0m"
	@echo "  \033[1;32mtest\033[0m        Run all tests with standard output"
	@echo "  \033[1;32mtest-verbose\033[0m  Run tests with verbose output and logs"
	@echo "  \033[1;32mtest-coverage\033[0m  Run tests with coverage analysis"
	@echo "  \033[1;32mtest-short\033[0m   Run only short tests (skip integration tests)"
	@echo "  \033[1;32mbench\033[0m       Run benchmarks"
	@echo ""
	@echo "\033[1;33m🔨 BUILD\033[0m"
	@echo "  \033[1;32mbuild\033[0m       Build for current platform \033[2m[VERSION, PLATFORM, ARCH]\033[0m"
	@echo "  \033[1;32mbuild-all\033[0m   Build for all supported platforms \033[2m[VERSION]\033[0m"
	@echo ""
	@echo "\033[1;33m📦 RELEASE & INSTALL\033[0m"
	@echo "  \033[1;32mrelease\033[0m     Complete release (test + clean + build-all + package) \033[2m[VERSION]\033[0m"
	@echo "  \033[1;32minstall\033[0m     Install plugin to Zoraxy plugins directory \033[2m[ZORAXY_DIR, VERSION, PLATFORM, ARCH]\033[0m"
	@echo ""
	@echo "\033[1;33m🛠️  UTILITY\033[0m"
	@echo "  \033[1;32mclean\033[0m       Remove all build artifacts"
	@echo "  \033[1;32mhelp\033[0m        Show this help message"
	@echo ""
	@echo "\033[1;33m⚙️  PARAMETERS\033[0m"
	@echo "  \033[1;36mVERSION=1.2.3\033[0m     Version (if omitted, uses latest git tag)"
	@echo "  \033[1;36mPLATFORM=linux\033[0m     Target OS (if omitted, uses current OS)"
	@echo "  \033[1;36mARCH=amd64\033[0m         Target architecture (if omitted, uses current arch)"
	@echo "  \033[1;36mZORAXY_DIR=/opt/zoraxy\033[0m  Zoraxy path (if omitted, tries to auto-detect)"
	@echo ""
	@echo "\033[1;33m💡 EXAMPLES\033[0m"
	@echo "  make test                              \033[2m# Run all tests\033[0m"
	@echo "  make test-coverage                     \033[2m# Run tests with coverage\033[0m"
	@echo "  make build VERSION=1.0.0              \033[2m# Explicit version\033[0m"
	@echo "  make build PLATFORM=linux ARCH=arm64  \033[2m# Cross-compile for ARM64\033[0m"
	@echo "  make install ZORAXY_DIR=/opt/zoraxy    \033[2m# Install to specific path\033[0m"
	@echo "  make release VERSION=2.1.0            \033[2m# Release with explicit version\033[0m"

# Test targets
test:
	@echo "→ Running tests..."
	@cd $(SRC_DIR) && go test -timeout=$(TEST_TIMEOUT) ./... && echo "✓ All tests passed"

test-verbose:
	@echo "→ Running tests with verbose output..."
	@cd $(SRC_DIR) && go test -v -timeout=$(TEST_TIMEOUT) ./...
	@echo "✓ All tests passed"

test-short:
	@echo "→ Running short tests..."
	@cd $(SRC_DIR) && go test -short -timeout=$(TEST_TIMEOUT) ./... && echo "✓ Short tests passed"

test-coverage:
	@echo "→ Running tests with coverage analysis..."
	@mkdir -p $(COVERAGE_DIR)
	@cd $(SRC_DIR) && go test -timeout=$(TEST_TIMEOUT) -coverprofile=../$(COVERAGE_DIR)/coverage.out -covermode=count $$(go list ./... | grep -v './mod/zoraxy_plugin')
	@cd $(SRC_DIR) && go tool cover -html=../$(COVERAGE_DIR)/coverage.out -o ../$(COVERAGE_DIR)/coverage.html
	@cd $(SRC_DIR) && go tool cover -func=../$(COVERAGE_DIR)/coverage.out
	@echo "✓ Coverage report generated at $(COVERAGE_DIR)/coverage.html"

bench:
	@echo "→ Running benchmarks..."
	@cd $(SRC_DIR) && go test -bench=. -benchmem ./...
	@echo "✓ Benchmarks completed"

# Generic build target - auto-detects platform and version if not specified
.PHONY: build build-all release install
build:
	$(call validate_version)
	$(call parse_version)
	$(eval PLATFORM := $(or $(PLATFORM),$(shell cd $(SRC_DIR) && go env GOOS 2>/dev/null || echo linux)))
	$(eval ARCH := $(or $(ARCH),$(shell cd $(SRC_DIR) && go env GOARCH 2>/dev/null || echo amd64)))
	@echo "→ Building $(PLATFORM)/$(ARCH) version $(CURRENT_VERSION)"
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-$(CURRENT_VERSION)-$(PLATFORM)-$(ARCH)/$(BINARY_NAME)
	@cd $(SRC_DIR) && GOOS=$(PLATFORM) GOARCH=$(ARCH) go build $(LDFLAGS) -o ../$(DIST_DIR)/$(BINARY_NAME)-$(CURRENT_VERSION)-$(PLATFORM)-$(ARCH)/$(BINARY_NAME)/$(BINARY_NAME)$(if $(filter windows,$(PLATFORM)),.exe) .
	@cp $(SRC_DIR)/icon.png $(DIST_DIR)/$(BINARY_NAME)-$(CURRENT_VERSION)-$(PLATFORM)-$(ARCH)/$(BINARY_NAME)/
	@# Ensure binary is executable (important for Unix-like systems)
	@chmod +x $(DIST_DIR)/$(BINARY_NAME)-$(CURRENT_VERSION)-$(PLATFORM)-$(ARCH)/$(BINARY_NAME)/$(BINARY_NAME)$(if $(filter windows,$(PLATFORM)),.exe) 2>/dev/null || true
	@echo "✓ $(PLATFORM)/$(ARCH) version $(CURRENT_VERSION) build completed"

# Build all platforms using the generic build target
build-all: test
	$(call validate_version)
	$(call parse_version)
	@$(MAKE) build VERSION=$(CURRENT_VERSION) PLATFORM=linux ARCH=amd64
	@$(MAKE) build VERSION=$(CURRENT_VERSION) PLATFORM=linux ARCH=arm64
	@$(MAKE) build VERSION=$(CURRENT_VERSION) PLATFORM=windows ARCH=amd64
	@$(MAKE) build VERSION=$(CURRENT_VERSION) PLATFORM=darwin ARCH=amd64
	@$(MAKE) build VERSION=$(CURRENT_VERSION) PLATFORM=darwin ARCH=arm64
	@$(MAKE) build VERSION=$(CURRENT_VERSION) PLATFORM=freebsd ARCH=amd64
	@echo "✓ All builds completed for version $(CURRENT_VERSION)"

release: test clean build-all
	$(call parse_version)
	@echo "Creating plugin packages for version $(CURRENT_VERSION)..."
	@cd $(DIST_DIR)/$(BINARY_NAME)-$(CURRENT_VERSION)-linux-amd64 && zip -r ../$(BINARY_NAME)-linux-amd64.zip $(BINARY_NAME)
	@cd $(DIST_DIR)/$(BINARY_NAME)-$(CURRENT_VERSION)-linux-arm64 && zip -r ../$(BINARY_NAME)-linux-arm64.zip $(BINARY_NAME)
	@cd $(DIST_DIR)/$(BINARY_NAME)-$(CURRENT_VERSION)-windows-amd64 && zip -r ../$(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME)
	@cd $(DIST_DIR)/$(BINARY_NAME)-$(CURRENT_VERSION)-darwin-amd64 && zip -r ../$(BINARY_NAME)-darwin-amd64.zip $(BINARY_NAME)
	@cd $(DIST_DIR)/$(BINARY_NAME)-$(CURRENT_VERSION)-darwin-arm64 && zip -r ../$(BINARY_NAME)-darwin-arm64.zip $(BINARY_NAME)
	@cd $(DIST_DIR)/$(BINARY_NAME)-$(CURRENT_VERSION)-freebsd-amd64 && zip -r ../$(BINARY_NAME)-freebsd-amd64.zip $(BINARY_NAME)
	@echo "✓ All packages created for version $(CURRENT_VERSION)"
	@echo "✓ Release $(CURRENT_VERSION) ready in $(DIST_DIR)/"

install: build
	$(call validate_version)
	$(call parse_version)
	$(eval PLATFORM := $(or $(PLATFORM),$(shell cd $(SRC_DIR) && go env GOOS 2>/dev/null || echo linux)))
	$(eval ARCH := $(or $(ARCH),$(shell cd $(SRC_DIR) && go env GOARCH 2>/dev/null || echo amd64)))
	@echo "→ Installing $(BINARY_NAME) version $(CURRENT_VERSION) for $(PLATFORM)/$(ARCH)"
	@# Determine Zoraxy installation path
	$(eval ZORAXY_INSTALL_PATH := $(if $(ZORAXY_DIR),$(ZORAXY_DIR),$(shell \
		ZORAXY_BIN=$$(which zoraxy 2>/dev/null); \
		if [ -n "$$ZORAXY_BIN" ]; then \
			if [ -L "$$ZORAXY_BIN" ]; then \
				readlink -f "$$ZORAXY_BIN" | xargs dirname; \
			else \
				dirname "$$ZORAXY_BIN"; \
			fi; \
		fi \
	)))
	@# Validate path exists and is accessible
	@if [ -z "$(ZORAXY_INSTALL_PATH)" ] || [ ! -d "$(ZORAXY_INSTALL_PATH)" ]; then \
		echo "❌ Error: Could not detect or access Zoraxy installation path."; \
		echo "   Please specify: make install ZORAXY_DIR=/path/to/zoraxy"; \
		exit 1; \
	fi
	@echo "   Using Zoraxy path: $(ZORAXY_INSTALL_PATH)"
	@# Create plugins directory if it doesn't exist
	@mkdir -p "$(ZORAXY_INSTALL_PATH)/plugins/$(BINARY_NAME)"
	@# Copy plugin files
	@cp -r "$(DIST_DIR)/$(BINARY_NAME)-$(CURRENT_VERSION)-$(PLATFORM)-$(ARCH)/$(BINARY_NAME)/"* "$(ZORAXY_INSTALL_PATH)/plugins/$(BINARY_NAME)/"
	@echo "✓ Plugin installed to $(ZORAXY_INSTALL_PATH)/plugins/$(BINARY_NAME)/"
