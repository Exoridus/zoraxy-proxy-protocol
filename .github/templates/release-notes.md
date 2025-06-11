# 🚀 ${PROJECT_NAME} ${CURRENT_TAG}

**Release Date:** ${RELEASE_DATE}

## 📦 What's Included

This release contains pre-built binaries for multiple platforms:

- **Linux** (AMD64, ARM64)
- **Windows** (AMD64) 
- **macOS** (Intel & Apple Silicon)
- **FreeBSD** (AMD64)

## 🔧 Installation

1. Download the appropriate binary for your platform
2. Make it executable: \`chmod +x proxy-protocol-*\`
3. Follow the setup instructions in the [README](${REPO_URL}/blob/main/README.md)

## 🔐 Verification

All binaries are provided with SHA256 checksums in \`checksums.txt\`. Verify your download:

\`\`\`bash
sha256sum -c checksums.txt
\`\`\`

## 📝 Changes in this Release

[Full Changelog](${REPO_URL}/compare/${PREVIOUS_TAG}...${CURRENT_TAG}) | [All Releases](${REPO_URL}/releases)

${COMMIT_LIST}

---

**Built with Go ${GO_VERSION}** | **Automated release from GitHub Actions** 
