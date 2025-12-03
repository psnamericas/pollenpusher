#!/bin/bash

# Build and Deploy CDR Generator GUI
# This script builds the GUI for Linux and deploys it to the production server

set -e

# Configuration
TARGET_SERVER="root@100.88.226.119"
TARGET_PATH="/usr/local/bin/cdrgenerator-gui"
BUILD_DIR="gui"

echo "========================================="
echo "CDR Generator GUI - Build & Deploy"
echo "========================================="
echo ""

# Check if we're in the right directory
if [ ! -d "$BUILD_DIR" ]; then
    echo "Error: Must run from cdrgenerator project root"
    exit 1
fi

# Copy source to server
echo "[1/4] Copying source code to server $TARGET_SERVER..."
rsync -av --exclude='cdrgenerator-gui*' "$BUILD_DIR/" "$TARGET_SERVER:/tmp/gui-build/"
rsync -av "config/" "$TARGET_SERVER:/tmp/gui-build-config/" || true
rsync -av "go.mod" "go.sum" "$TARGET_SERVER:/tmp/gui-build-root/"

# Build and install on server
echo "[2/4] Building on server..."
ssh "$TARGET_SERVER" << 'EOF'
    # Stop any running GUI instances
    pkill -f cdrgenerator-gui || true

    # Create build directory if it doesn't exist
    mkdir -p /opt/cdrgenerator-gui-build
    cd /opt/cdrgenerator-gui-build

    # Copy files
    cp -r /tmp/gui-build/* .
    cp /tmp/gui-build-root/go.* .
    cp -r /tmp/gui-build-config config/ || mkdir -p config

    # Install dependencies if needed
    if ! command -v go &> /dev/null; then
        echo "Error: Go is not installed on the server"
        exit 1
    fi

    # Install build dependencies for Fyne (if not already installed)
    echo "Checking GUI dependencies..."
    apt-get update -qq || true
    apt-get install -y -qq libgl1-mesa-dev xorg-dev 2>/dev/null || true

    # Build
    echo "Building GUI..."
    go mod tidy
    go build -o /usr/local/bin/cdrgenerator-gui

    # Verify installation
    ls -lh /usr/local/bin/cdrgenerator-gui

    # Clean up
    cd /
    rm -rf /tmp/gui-build /tmp/gui-build-root /tmp/gui-build-config
EOF

echo ""
echo "[3/4] Installation complete!"

echo ""
echo "[4/4] Deployment complete!"
echo ""
echo "========================================="
echo "GUI installed at: $TARGET_PATH"
echo "========================================="
echo ""
echo "To run the GUI on the server:"
echo "  1. SSH with X11 forwarding:"
echo "     ssh -X $TARGET_SERVER"
echo "     cdrgenerator-gui"
echo ""
echo "  2. Or use VNC/remote desktop"
echo ""
echo "Done!"
