#!/bin/bash

set -euo pipefail

# Generate release notes for GitHub releases
# Usage: generate-release-notes.sh <current_tag> <repository> <go_version>

CURRENT_TAG="${1:-}"
REPOSITORY="${2:-}"
GO_VERSION="${3:-}"

if [[ -z "$CURRENT_TAG" || -z "$REPOSITORY" || -z "$GO_VERSION" ]]; then
    echo "Usage: $0 <current_tag> <repository> <go_version>"
    echo "Example: $0 v1.0.0 owner/repo 1.20"
    exit 1
fi

# Constants
REPO_URL="https://github.com/$REPOSITORY"
PROJECT_NAME=$(echo "$REPOSITORY" | cut -d'/' -f2)
RELEASE_DATE=$(git log -1 --date=short --format="%ad")
OUTPUT_FILE="release-notes.txt"

# Check if CHANGELOG.md exists and use it
if [[ -f "CHANGELOG.md" ]]; then
    echo "Using existing CHANGELOG.md for release notes"
    awk "/^## \\[$CURRENT_TAG\\]/{flag=1; next} /^## \\[/{flag=0} flag" CHANGELOG.md > "$OUTPUT_FILE"
    exit 0
fi

echo "Generating release notes for $CURRENT_TAG..."

# Get previous tag, optimized for performance
get_previous_tag() {
    local current_tag="$1"
    local previous_tag

    # First try to get any previous tag
    previous_tag=$(git describe --tags --abbrev=0 HEAD~ 2>/dev/null || echo "")

    if [[ -z "$previous_tag" ]]; then
        echo ""
        return
    fi

    # If current is not beta but previous is beta, find non-beta previous tag
    if [[ $current_tag != *beta* && $previous_tag == *beta* ]]; then
        previous_tag=$(git describe --tags --abbrev=0 --exclude="*beta*" HEAD~ 2>/dev/null || echo "")
    fi

    echo "$previous_tag"
}

# Categorize commits by gitmoji with optimized regex matching
categorize_commits() {
    local commit_range="$1"
    local repo_url="$2"

    # Pre-compile regex patterns for better performance
    local feature_pattern='^✨'
    local fix_pattern='^🐛'
    local docs_pattern='^📝'
    local perf_pattern='^⚡'
    local refactor_pattern='^♻️'
    local test_pattern='^✅'
    local chore_pattern='^(🔧|🔨|📦|⬆️|⬇️)'

    # Initialize arrays for better memory management
    declare -a features=()
    declare -a fixes=()
    declare -a docs=()
    declare -a performance=()
    declare -a refactor=()
    declare -a tests=()
    declare -a chores=()
    declare -a others=()

    # Single git call to get all commits
    local all_commits
    all_commits=$(git log --pretty="format:%s|%h|%H" --no-merges "$commit_range" 2>/dev/null || echo "")

    if [[ -z "$all_commits" ]]; then
        echo "- Initial release"
        return
    fi

    # Process commits in a single pass
    while IFS='|' read -r msg hash full_hash; do
        [[ -z "$msg" ]] && continue

        local commit_link="[\`$hash\`]($repo_url/commit/$full_hash)"
        local commit_entry="- $msg ($commit_link)"

        # Use case statement for better performance than multiple if-elif
        case "$msg" in
            $feature_pattern*)
                features+=("- ${msg#✨ } ($commit_link)")
                ;;
            $fix_pattern*)
                fixes+=("- ${msg#🐛 } ($commit_link)")
                ;;
            $docs_pattern*)
                docs+=("- ${msg#📝 } ($commit_link)")
                ;;
            $perf_pattern*)
                performance+=("- ${msg#⚡ } ($commit_link)")
                ;;
            $refactor_pattern*)
                refactor+=("- ${msg#♻️ } ($commit_link)")
                ;;
            $test_pattern*)
                tests+=("- ${msg#✅ } ($commit_link)")
                ;;
            🔧*|🔨*|📦*|⬆️*|⬇️*)
                chores+=("$commit_entry")
                ;;
            *)
                others+=("$commit_entry")
                ;;
        esac
    done <<< "$all_commits"

    # Build sections efficiently
    local commit_list=""

    [[ ${#features[@]} -gt 0 ]] && {
        commit_list+="### ✨ Features"$'\n\n'
        printf '%s\n' "${features[@]}" >> temp_features
        commit_list+="$(cat temp_features)"$'\n\n'
        rm -f temp_features
    }

    [[ ${#fixes[@]} -gt 0 ]] && {
        commit_list+="### 🐛 Bug Fixes"$'\n\n'
        printf '%s\n' "${fixes[@]}" >> temp_fixes
        commit_list+="$(cat temp_fixes)"$'\n\n'
        rm -f temp_fixes
    }

    [[ ${#performance[@]} -gt 0 ]] && {
        commit_list+="### ⚡ Performance"$'\n\n'
        printf '%s\n' "${performance[@]}" >> temp_performance
        commit_list+="$(cat temp_performance)"$'\n\n'
        rm -f temp_performance
    }

    [[ ${#refactor[@]} -gt 0 ]] && {
        commit_list+="### ♻️ Refactoring"$'\n\n'
        printf '%s\n' "${refactor[@]}" >> temp_refactor
        commit_list+="$(cat temp_refactor)"$'\n\n'
        rm -f temp_refactor
    }

    [[ ${#tests[@]} -gt 0 ]] && {
        commit_list+="### ✅ Tests"$'\n\n'
        printf '%s\n' "${tests[@]}" >> temp_tests
        commit_list+="$(cat temp_tests)"$'\n\n'
        rm -f temp_tests
    }

    [[ ${#docs[@]} -gt 0 ]] && {
        commit_list+="### 📝 Documentation"$'\n\n'
        printf '%s\n' "${docs[@]}" >> temp_docs
        commit_list+="$(cat temp_docs)"$'\n\n'
        rm -f temp_docs
    }

    [[ ${#chores[@]} -gt 0 ]] && {
        commit_list+="### 🔧 Maintenance"$'\n\n'
        printf '%s\n' "${chores[@]}" >> temp_chores
        commit_list+="$(cat temp_chores)"$'\n\n'
        rm -f temp_chores
    }

    [[ ${#others[@]} -gt 0 ]] && {
        commit_list+="### 🔀 Other Changes"$'\n\n'
        printf '%s\n' "${others[@]}" >> temp_others
        commit_list+="$(cat temp_others)"$'\n\n'
        rm -f temp_others
    }

    # If no categorized commits, show all
    if [[ -z "$commit_list" ]]; then
        commit_list=$(git log --pretty="format:- %s ([\`%h\`]($repo_url/commit/%H))" --no-merges "$commit_range" 2>/dev/null || echo "- Initial release")
    fi

    echo "$commit_list"
}

# Main logic
PREVIOUS_TAG=$(get_previous_tag "$CURRENT_TAG")

if [[ -n "$PREVIOUS_TAG" && "$PREVIOUS_TAG" != "$CURRENT_TAG" ]]; then
    echo "Generating changelog from $PREVIOUS_TAG to $CURRENT_TAG"
    COMMIT_LIST=$(categorize_commits "${PREVIOUS_TAG}..HEAD" "$REPO_URL")
else
    echo "No previous tag found, generating initial release notes"
    COMMIT_LIST="- Initial release"
fi

# Check if template exists, create basic one if not
TEMPLATE_FILE=".github/templates/release-notes.md"
if [[ ! -f "$TEMPLATE_FILE" ]]; then
    echo "Creating basic release notes template..."
    mkdir -p "$(dirname "$TEMPLATE_FILE")"
    cat > "$TEMPLATE_FILE" << 'EOF'
# \${PROJECT_NAME} \${CURRENT_TAG}

**Release Date:** \${RELEASE_DATE}

## Changes in this Release

\${COMMIT_LIST}

## Installation

Download the appropriate binary for your platform from the assets below.

## Requirements

- Go \${GO_VERSION}+ (for building from source)

---

**Full Changelog:** [View Changes](\${REPO_URL}/compare/\${PREVIOUS_TAG}...\${CURRENT_TAG}) | [All Releases](\${REPO_URL}/releases)
EOF
fi

# Use template and replace variables efficiently
cp "$TEMPLATE_FILE" "$OUTPUT_FILE"

# Use sed for efficient string replacement
sed -i "s|\${PROJECT_NAME}|$PROJECT_NAME|g" "$OUTPUT_FILE"
sed -i "s|\${CURRENT_TAG}|$CURRENT_TAG|g" "$OUTPUT_FILE"
sed -i "s|\${RELEASE_DATE}|$RELEASE_DATE|g" "$OUTPUT_FILE"
sed -i "s|\${REPO_URL}|$REPO_URL|g" "$OUTPUT_FILE"
sed -i "s|\${GO_VERSION}|$GO_VERSION|g" "$OUTPUT_FILE"

# Handle changelog link based on whether we have a previous tag
if [[ -n "$PREVIOUS_TAG" ]]; then
    sed -i "s|\${PREVIOUS_TAG}|$PREVIOUS_TAG|g" "$OUTPUT_FILE"
else
    sed -i "s|\\[View Changes\\](\${REPO_URL}/compare/\${PREVIOUS_TAG}...\${CURRENT_TAG}) \\| ||g" "$OUTPUT_FILE"
fi

# Replace commit list using a temporary file for better performance
echo "$COMMIT_LIST" > commits.tmp
sed -i "/\${COMMIT_LIST}/r commits.tmp" "$OUTPUT_FILE"
sed -i "/\${COMMIT_LIST}/d" "$OUTPUT_FILE"
rm -f commits.tmp

echo "Release notes generated successfully: $OUTPUT_FILE"
