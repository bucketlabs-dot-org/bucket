#!/usr/bin/env bash
# bucket CLI installation 

set -euo pipefail

# webhook on version upload so file won't have to be changed
VERSION_URL="https://raw.githubusercontent.com/bucketlabs-dot-org/bucket/refs/heads/main/install/VERSION"
VERSION="$(curl -sSL ${VERSION_URL})"
BIN_NAME="bucket"
BIN_PATH="/usr/local/bin/$BIN_NAME"
URI="https://github.com/bucketlabs-dot-org/bucket/releases/download"

info() { echo -e "\033[0;32m[INFO]\033[0m $1"; }
error() { echo -e "\033[0;31m[ERROR]\033[0m $1" >&2; exit 1; }

OS_TYPE="$(uname -s)"
case "${OS_TYPE}" in
    Linux*)     DOWNLOAD_URL=$RELEASE_LINUX;;
    Darwin*)    DOWNLOAD_URL=$RELEASE_MACOS;;
    *)          error "Unsupported operating system: ${OS_TYPE}";;
esac

ARCH=$(uname -m | tr '[:upper:]' '[:lower:]')
case "${ARCH}" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    i*86|i386) ARCH="386" ;;  
    *) error "Unsupported architecture: ${ARCH} (only amd64/arm64 supported)" ;;
esac

info "Detected OS: ${OS_TYPE}"
info "Detected ARCH: ${ARCH}"
RELEASE_LINUX="${URI}/v${VERSION}/bucket-cli-${VERSION}-linux_${ARCH}"
RELEASE_MACOS="${URI}/v${VERSION}/bucket-cli-${VERSION}-darwin_${ARCH}"
info "Downloading ${BIN_NAME} v${VERSION}..."

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT
TMP_BIN="$TMP_DIR/$BIN_NAME"

if ! curl -sSL "$DOWNLOAD_URL" -o "$TMP_BIN"; then
    error "Failed to download binary from $DOWNLOAD_URL"
fi

chmod +x "$TMP_BIN"

info "Installing to ${BIN_PATH}..."

if [ -w "/usr/local/bin" ]; then
    mv "$TMP_BIN" "$BIN_PATH"
else
    info "Elevated permissions required to install to /usr/local/bin"
    sudo mv "$TMP_BIN" "$BIN_PATH"
fi

info "Installation complete! Try running: ${BIN_NAME} --help"
