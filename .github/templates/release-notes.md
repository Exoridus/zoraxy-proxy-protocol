# ${PROJECT_NAME} v${CURRENT_TAG} (${RELEASE_DATE})

This release contains pre-built ZIP packages for multiple platforms with the correct directory structure and executable permissions:

- **Linux** ([AMD64](${REPO_URL}/releases/download/${CURRENT_TAG}/proxy-protocol-linux-amd64.zip), [ARM64](${REPO_URL}/releases/download/${CURRENT_TAG}/proxy-protocol-linux-arm64.zip))
- **Windows** ([AMD64](${REPO_URL}/releases/download/${CURRENT_TAG}/proxy-protocol-windows-amd64.zip))
- **macOS** ([Intel](${REPO_URL}/releases/download/${CURRENT_TAG}/proxy-protocol-darwin-amd64.zip), [Apple Silicon](${REPO_URL}/releases/download/${CURRENT_TAG}/proxy-protocol-darwin-arm64.zip))
- **FreeBSD** ([AMD64](${REPO_URL}/releases/download/${CURRENT_TAG}/proxy-protocol-freebsd-amd64.zip))

## 🚀 Installation

**Installation Steps:**
1. Download the ZIP package for your platform from the links above
2. Extract the ZIP file to your Zoraxy plugins directory:
   ```bash
   unzip proxy-protocol.zip -d /path/to/zoraxy/plugins/
   ```
3. The plugin is automatically ready - no restart required!

**Platform-Specific One-liner Commands:**

**Linux AMD64:**
```bash
wget -qO- "${REPO_URL}/releases/latest/download/proxy-protocol-linux-amd64.zip" | unzip - -d "/path/to/zoraxy/plugins/"
```

**Linux ARM64:**
```bash
wget -qO- "${REPO_URL}/releases/latest/download/proxy-protocol-linux-arm64.zip" | unzip - -d "/path/to/zoraxy/plugins/"
```

**Windows AMD64:**
```bash
wget -qO- "${REPO_URL}/releases/latest/download/proxy-protocol-windows-amd64.zip" | unzip - -d "/path/to/zoraxy/plugins/"
```

**macOS Intel:**
```bash
wget -qO- "${REPO_URL}/releases/latest/download/proxy-protocol-darwin-amd64.zip" | unzip - -d "/path/to/zoraxy/plugins/"
```

**macOS Apple Silicon:**
```bash
wget -qO- "${REPO_URL}/releases/latest/download/proxy-protocol-darwin-arm64.zip" | unzip - -d "/path/to/zoraxy/plugins/"
```

**FreeBSD AMD64:**
```bash
wget -qO- "${REPO_URL}/releases/latest/download/proxy-protocol-freebsd-amd64.zip" | unzip - -d "/path/to/zoraxy/plugins/"
```

> **Note:** The ZIP contains the correct directory structure with executable permissions preserved.

## 📁 What's Included

Each ZIP package contains:
- `proxy-protocol/proxy-protocol` - Plugin executable (with correct permissions)
- `proxy-protocol/icon.png` - Plugin icon for the web interface

The ZIP automatically creates the correct directory structure when extracted to `{zoraxy-dir}/plugins/`.

## 📁 Directory Structure

After installation, your Zoraxy directory will look like this:
```
/path/to/zoraxy/
├── zoraxy                ← Main Zoraxy executable
├── plugins/
│   └── proxy-protocol/
│       ├── proxy-protocol ← Plugin executable (auto-extracted)
│       └── icon.png      ← Plugin icon (auto-extracted)
└── ... (other files)
```

For detailed instructions and troubleshooting, see the [README](${REPO_URL}/blob/master/README.md)

## 📝 Changes in this Release

[View on GitHub](${REPO_URL}/releases/tag/${CURRENT_TAG}) | [All Releases](${REPO_URL}/releases)

${COMMIT_LIST}

---

**Built with Go ${GO_VERSION}** | **Automated release from GitHub Actions** 
