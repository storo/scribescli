#!/usr/bin/env bash
set -euo pipefail

# Vosk Installation Script for ScribesAI
# Installs libvosk for Linux and macOS (Intel + Apple Silicon)
# Version: 0.3.45

VOSK_VERSION="0.3.45"
INSTALL_DIR="/usr/local"
LIB_DIR="${INSTALL_DIR}/lib"
INCLUDE_DIR="${INSTALL_DIR}/include"
TEMP_DIR=$(mktemp -d)

cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

echo "==== Vosk Installation Script ===="
echo "Version: $VOSK_VERSION"
echo ""

# Detect OS and architecture
OS=$(uname -s)
ARCH=$(uname -m)

case "$OS" in
    Linux)
        case "$ARCH" in
            x86_64)
                PLATFORM="linux-x86_64"
                LIB_NAME="libvosk.so"
                ;;
            aarch64|arm64)
                PLATFORM="linux-aarch64"
                LIB_NAME="libvosk.so"
                ;;
            *)
                echo "Error: Unsupported Linux architecture: $ARCH"
                exit 1
                ;;
        esac
        ;;
    Darwin)
        case "$ARCH" in
            x86_64)
                PLATFORM="osx-x86_64"
                LIB_NAME="libvosk.dylib"
                ;;
            arm64)
                PLATFORM="osx-arm64"
                LIB_NAME="libvosk.dylib"
                ;;
            *)
                echo "Error: Unsupported macOS architecture: $ARCH"
                exit 1
                ;;
        esac
        ;;
    *)
        echo "Error: Unsupported operating system: $OS"
        exit 1
        ;;
esac

echo "Detected platform: $PLATFORM"
echo "Install directory: $INSTALL_DIR"
echo ""

# Check if already installed
if [ -f "$LIB_DIR/$LIB_NAME" ]; then
    echo "Vosk is already installed at $LIB_DIR/$LIB_NAME"
    read -p "Do you want to reinstall? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Installation cancelled."
        exit 0
    fi
fi

# Download libvosk
ARCHIVE_NAME="vosk-${PLATFORM}-${VOSK_VERSION}.zip"
DOWNLOAD_URL="https://github.com/alphacep/vosk-api/releases/download/v${VOSK_VERSION}/${ARCHIVE_NAME}"

echo "Downloading Vosk..."
echo "URL: $DOWNLOAD_URL"
echo ""

cd "$TEMP_DIR"
if ! curl -L -o "$ARCHIVE_NAME" "$DOWNLOAD_URL"; then
    echo "Error: Failed to download Vosk"
    exit 1
fi

echo "Extracting archive..."
if ! unzip -q "$ARCHIVE_NAME"; then
    echo "Error: Failed to extract archive"
    exit 1
fi

# Find the extracted directory (varies by platform)
EXTRACTED_DIR=$(find . -maxdepth 1 -type d -name "vosk-*" | head -n 1)
if [ -z "$EXTRACTED_DIR" ]; then
    echo "Error: Could not find extracted directory"
    exit 1
fi

echo "Installing to $INSTALL_DIR..."
echo ""

# Install library
echo "Installing library..."
sudo mkdir -p "$LIB_DIR"
sudo cp "$EXTRACTED_DIR/$LIB_NAME" "$LIB_DIR/"
sudo chmod 755 "$LIB_DIR/$LIB_NAME"

# Install headers
echo "Installing headers..."
sudo mkdir -p "$INCLUDE_DIR/vosk"
if [ -f "$EXTRACTED_DIR/vosk_api.h" ]; then
    sudo cp "$EXTRACTED_DIR/vosk_api.h" "$INCLUDE_DIR/vosk/"
elif [ -f "$EXTRACTED_DIR/include/vosk_api.h" ]; then
    sudo cp "$EXTRACTED_DIR/include/vosk_api.h" "$INCLUDE_DIR/vosk/"
else
    echo "Warning: Could not find vosk_api.h"
fi

# Update library cache (Linux only)
if [ "$OS" = "Linux" ]; then
    echo "Updating library cache..."
    sudo ldconfig
fi

# Create pkg-config file
echo "Creating pkg-config file..."
PKG_CONFIG_DIR="${INSTALL_DIR}/lib/pkgconfig"
sudo mkdir -p "$PKG_CONFIG_DIR"

cat << EOF | sudo tee "$PKG_CONFIG_DIR/vosk.pc" > /dev/null
prefix=$INSTALL_DIR
exec_prefix=\${prefix}
libdir=\${exec_prefix}/lib
includedir=\${prefix}/include

Name: Vosk
Description: Offline speech recognition API
Version: $VOSK_VERSION
Libs: -L\${libdir} -lvosk
Cflags: -I\${includedir}
EOF

# Verify installation
echo ""
echo "Verifying installation..."
if [ -f "$LIB_DIR/$LIB_NAME" ]; then
    echo "✓ Library installed: $LIB_DIR/$LIB_NAME"
else
    echo "✗ Library NOT found"
    exit 1
fi

if [ -f "$INCLUDE_DIR/vosk/vosk_api.h" ]; then
    echo "✓ Header installed: $INCLUDE_DIR/vosk/vosk_api.h"
else
    echo "✗ Header NOT found"
fi

if pkg-config --exists vosk 2>/dev/null; then
    echo "✓ pkg-config configured"
    echo ""
    echo "pkg-config output:"
    pkg-config --cflags --libs vosk
else
    echo "⚠ pkg-config not configured (non-fatal)"
    echo "You may need to set PKG_CONFIG_PATH=$PKG_CONFIG_DIR"
fi

echo ""
echo "==== Installation Complete ===="
echo ""
echo "Next steps:"
echo "1. Download a Vosk model from https://alphacephei.com/vosk/models"
echo "2. Extract it to ./models/"
echo "3. Or use ScribesAI's built-in model manager"
echo ""
echo "Example:"
echo "  mkdir -p models"
echo "  cd models"
echo "  curl -L https://alphacephei.com/vosk/models/vosk-model-small-en-us-0.15.zip -o model.zip"
echo "  unzip model.zip"
echo ""
