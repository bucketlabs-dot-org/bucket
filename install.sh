#!/usr/bin/env bash
# bucket CLI installation 

# Exit immediately if a command exits with a non-zero status
set -euo pipefail

# --- Variables ---
BIN_NAME="bucket"
BIN_PATH="/usr/local/bin/$BIN_NAME"
VERSION="0.0.3"

URI="https://github.com/bucketlabs-dot-org/bucket/releases/download"
RELEASE_LINUX="${URI}/Linux/bucket-cli-${VERSION}-linux_x86_64"
RELEASE_MACOS="${URI}/MacOS/bucket-cli-${VERSION}-darwin_x86_64"

# --- Functions ---

info() { echo -e "\033[0;32m[INFO]\033[0m $1"; }
error() { echo -e "\033[0;31m[ERROR]\033[0m $1" >&2; exit 1; }

# --- Logic ---

# 1. Detect OS
OS_TYPE="$(uname -s)"
case "${OS_TYPE}" in
    Linux*)     DOWNLOAD_URL=$RELEASE_LINUX;;
    Darwin*)    DOWNLOAD_URL=$RELEASE_MACOS;;
    *)          error "Unsupported operating system: ${OS_TYPE}";;
esac

info "Detected OS: ${OS_TYPE}"
info "Downloading ${BIN_NAME} v${VERSION}..."

# 2. Create a temporary directory for the download
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT
TMP_BIN="$TMP_DIR/$BIN_NAME"

# 3. Download the binary
if ! curl -sSL "$DOWNLOAD_URL" -o "$TMP_BIN"; then
    error "Failed to download binary from $DOWNLOAD_URL"
fi

# 4. Make it executable
chmod +x "$TMP_BIN"

# 5. Move to /usr/local/bin (handling permissions)
info "Installing to ${BIN_PATH}..."

if [ -w "/usr/local/bin" ]; then
    mv "$TMP_BIN" "$BIN_PATH"
else
    info "Elevated permissions required to install to /usr/local/bin"
    sudo mv "$TMP_BIN" "$BIN_PATH"
fi

info "Installation complete! Try running: ${BIN_NAME} --help"
