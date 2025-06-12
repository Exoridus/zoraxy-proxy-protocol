# ${PROJECT_NAME} v${CURRENT_TAG} (${RELEASE_DATE})

This release contains pre-built binaries for multiple platforms:

- **Linux** ([AMD64](${REPO_URL}/releases/download/${CURRENT_TAG}/proxy-protocol-linux-amd64), [ARM64](${REPO_URL}/releases/download/${CURRENT_TAG}/proxy-protocol-linux-arm64))
- **Windows** ([AMD64](${REPO_URL}/releases/download/${CURRENT_TAG}/proxy-protocol-windows-amd64.exe))
- **macOS** ([Intel](${REPO_URL}/releases/download/${CURRENT_TAG}/proxy-protocol-darwin-amd64), [Apple Silicon](${REPO_URL}/releases/download/${CURRENT_TAG}/proxy-protocol-darwin-arm64))
- **FreeBSD** ([AMD64](${REPO_URL}/releases/download/${CURRENT_TAG}/proxy-protocol-freebsd-amd64))

## 🔧 Installation

1. Download the appropriate binary for your platform
2. Go to your Zoraxy installation directory (where the `zoraxy` executable is located)
3. Create directory: `mkdir -p plugins/proxy-protocol`
4. Copy binary: `cp proxy-protocol-* plugins/proxy-protocol/proxy-protocol`
5. Make executable (Linux/macOS): `chmod +x plugins/proxy-protocol/proxy-protocol`
6. Restart Zoraxy and configure via **Plugins** → **Proxy Protocol**

> **Path:** `{zoraxy-directory}/plugins/proxy-protocol/proxy-protocol`

For detailed instructions, see the [README](${REPO_URL}/blob/master/README.md)

## 🔐 Verification

All binaries are provided with SHA256 checksums in `checksums.txt`. Verify your download:

```bash
sha256sum -c checksums.txt
```

## 📝 Changes in this Release

[View on GitHub](${REPO_URL}/releases/tag/${CURRENT_TAG}) | [All Releases](${REPO_URL}/releases)

${COMMIT_LIST}

---

**Built with Go ${GO_VERSION}** | **Automated release from GitHub Actions** 
