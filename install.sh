#!/bin/bash
set -e

# Craft CLI Installer
# Usage: curl -sSL https://raw.githubusercontent.com/nerveband/craft-cli/main/install.sh | bash

REPO="nerveband/craft-cli"
BINARY_NAME="craft"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Detect OS
detect_os() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$OS" in
        darwin) OS="Darwin" ;;
        linux) OS="Linux" ;;
        mingw*|msys*|cygwin*) OS="Windows" ;;
        *) error "Unsupported operating system: $OS" ;;
    esac
}

# Detect architecture
detect_arch() {
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64|amd64) ARCH="x86_64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac
}

# Get latest release version from GitHub
get_latest_version() {
    LATEST_VERSION=$(curl -sS "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$LATEST_VERSION" ]; then
        error "Failed to get latest version. Check your internet connection."
    fi
}

# Download and install
install() {
    detect_os
    detect_arch
    get_latest_version

    info "Installing Craft CLI ${LATEST_VERSION} for ${OS}/${ARCH}..."

    # Construct download URL
    if [ "$OS" = "Windows" ]; then
        ARCHIVE_NAME="craft-cli_${OS}_${ARCH}.zip"
    else
        ARCHIVE_NAME="craft-cli_${OS}_${ARCH}.tar.gz"
    fi

    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_VERSION}/${ARCHIVE_NAME}"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    info "Downloading from ${DOWNLOAD_URL}..."
    if ! curl -sSL "$DOWNLOAD_URL" -o "${TMP_DIR}/${ARCHIVE_NAME}"; then
        error "Failed to download. Release may not exist yet."
    fi

    # Extract
    info "Extracting..."
    cd "$TMP_DIR"
    if [ "$OS" = "Windows" ]; then
        unzip -q "$ARCHIVE_NAME"
    else
        tar -xzf "$ARCHIVE_NAME"
    fi

    # Install
    if [ -w "$INSTALL_DIR" ]; then
        mv "$BINARY_NAME" "$INSTALL_DIR/"
    else
        info "Installing to ${INSTALL_DIR} (requires sudo)..."
        sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
    fi

    # Verify installation
    if command -v "$BINARY_NAME" &> /dev/null; then
        info "Successfully installed Craft CLI!"
        echo ""
        "$BINARY_NAME" version
        echo ""
        info "Run 'craft --help' to get started"
        info "Run 'craft upgrade' to update to the latest version"
    else
        warn "Installation complete, but '$BINARY_NAME' not found in PATH"
        warn "You may need to add ${INSTALL_DIR} to your PATH"
        echo ""
        echo "Add this to your shell profile:"
        echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    fi
}

# Run installer
install
