# Contributing to Zoraxy Proxy Protocol Plugin

Thank you for your interest in contributing! This guide covers the development workflow, coding conventions, and technical details for this Zoraxy plugin.

## Development Setup

### Prerequisites

- **Go 1.20+** installed on your system
- **Make** (optional, but recommended)
- **Git** for version control

### Quick Start

1. **Clone and setup:**
```bash
git clone https://github.com/Exoridus/zoraxy-proxy-protocol.git
cd zoraxy-proxy-protocol
go mod tidy
```

2. **Build for testing:**
```bash
make build-linux  # For most server deployments
# or
make build        # For current platform
```

3. **Run tests:**
```bash
make test lint
```

## Development Workflow

### Making Changes

1. **Create a feature branch:**
```bash
git checkout -b feature-name
```

2. **Edit source code** (mainly `main.go` and `proxy_protocol_utils.go`)

3. **Test your changes:**
```bash
make test        # Run unit tests
make lint        # Check code quality
make build-linux # Build for testing
```

4. **Test with Zoraxy:**
   - Copy binary to test Zoraxy instance: `plugins/proxy-protocol/proxy-protocol`
   - Restart Zoraxy to load the plugin
   - Test functionality through the UI

### Code Conventions

#### Go Style Guide
- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Write self-documenting code with minimal comments
- Handle errors explicitly

#### Plugin Architecture
- **Thread Safety**: Use mutexes for shared state
- **Error Handling**: Return proper HTTP status codes
- **Graceful Degradation**: Handle disabled state properly
- **Logging**: Use `fmt.Printf` for important state changes

#### UI/API Conventions
- Follow REST principles for API endpoints
- Use JSON for all API communication
- Prefix plugin endpoints with `/ui/api/`
- Return consistent response structures

### Testing Guidelines

1. **Unit Testing:**
```bash
make test
```

2. **Integration Testing:**
   - Deploy to test Zoraxy instance
   - Test with real proxy protocol traffic
   - Verify client IP forwarding works

3. **Manual Testing:**
   - Plugin loads in Zoraxy admin interface
   - Configuration UI is functional
   - Enable/disable toggle works
   - Status reporting is accurate

## Plugin Technical Architecture

### Key Components

1. **Plugin Specification** - Metadata and endpoint definitions
2. **UI Integration** - Embedded web interface 
3. **API Endpoints** - Configuration and status management
4. **Core Handlers** - Proxy protocol detection and processing

### File Structure

```
zoraxy-proxy-protocol/
├── main.go                 # Plugin entry point and API handlers
├── proxy_protocol_utils.go # Core proxy protocol implementation
├── go.mod                 # Go module definition  
├── www/index.html         # Plugin web UI
├── mod/zoraxy_plugin/     # Zoraxy plugin SDK
├── Makefile              # Build automation
└── .github/workflows/    # CI/CD pipeline
```

### Plugin Integration Points

**UI_PATH (`"/ui"`)** is essential for:
- Zoraxy integration (mounting point)
- Embedded router setup
- API endpoint namespacing

**Core Endpoints:**
- `/proxy_protocol_sniff` - Detection endpoint called by Zoraxy
- `/proxy_protocol_handler` - Processing endpoint for traffic
- `/ui/api/status` - Status information for UI
- `/ui/api/toggle` - Enable/disable functionality

## Release Process

### Semantic Versioning

This project follows [SemVer](https://semver.org/) without the "v" prefix:

- **MAJOR** (1.0.0 → 2.0.0): Breaking changes
- **MINOR** (1.0.0 → 1.1.0): New features (backward compatible)
- **PATCH** (1.0.0 → 1.0.1): Bug fixes (backward compatible)

### Creating Releases

1. **Commit and push changes:**
```bash
git add .
git commit -m "🐛 fix parsing issue"  # Use gitmoji
git push origin feature-name
```

2. **Create pull request** and get it merged

3. **Tag release from master branch:**
```bash
git checkout master
git pull origin master
git tag 1.0.1
git push origin 1.0.1
```

4. **GitHub Actions automatically:**
   - Builds binaries for all platforms
   - Runs tests and checks
   - Creates GitHub release
   - Uploads binaries with checksums

## Commit Message Conventions

We use [Gitmoji](https://gitmoji.dev/) for clean, visual commit messages:

### Common Patterns
- 🎉 `:tada:` Initial commit
- ✨ `:sparkles:` New feature
- 🐛 `:bug:` Bug fix
- 📝 `:memo:` Documentation
- ♻️ `:recycle:` Refactoring
- ⚡ `:zap:` Performance improvements
- ✅ `:white_check_mark:` Tests
- 🔧 `:wrench:` Configuration

### Usage
```bash
# Interactive mode (recommended)
npx gitmoji -c

# Direct commit
git commit -m "✨ add proxy protocol v2 support"
```

## Useful Make Targets

```bash
make build       # Build for current platform
make build-linux # Build for Linux deployment
make build-all   # Build for all platforms  
make test        # Run tests
make lint        # Code quality checks
make clean       # Remove build artifacts
make install     # Install locally for testing
make help        # Show all targets
```

## Pull Request Guidelines

1. **Describe your changes** clearly in the PR description
2. **Reference issues** if applicable (`Fixes #123`)
3. **Test thoroughly** before submitting
4. **Follow code conventions** outlined above
5. **Update documentation** if needed

### PR Checklist
- [ ] Tests pass (`make test`)
- [ ] Linting passes (`make lint`)
- [ ] Plugin builds successfully
- [ ] Manual testing completed
- [ ] Documentation updated if needed
- [ ] Commit messages follow gitmoji convention

## Getting Help

- **Issues**: Use GitHub issues for bugs and feature requests
- **Discussions**: Use GitHub discussions for questions
- **Documentation**: Check existing docs in the repository

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (MIT). 
