#!/bin/bash
#
# CDRGenerator installation script
#

set -e

# Configuration
INSTALL_DIR="/opt/cdrgenerator"
CONFIG_DIR="/etc/cdrgenerator"
LOG_DIR="/var/log/cdrgenerator"
SERVICE_USER="cdrgenerator"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

echo_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo_error "Please run as root or with sudo"
    exit 1
fi

echo_info "Installing CDRGenerator..."

# Create service user if it doesn't exist
if ! id "$SERVICE_USER" &>/dev/null; then
    echo_info "Creating service user: $SERVICE_USER"
    useradd -r -s /sbin/nologin -d "$INSTALL_DIR" "$SERVICE_USER"
fi

# Add user to dialout group for serial port access
echo_info "Adding $SERVICE_USER to dialout group..."
usermod -a -G dialout "$SERVICE_USER"

# Create directories
echo_info "Creating directories..."
mkdir -p "$INSTALL_DIR"
mkdir -p "$INSTALL_DIR/samples"
mkdir -p "$CONFIG_DIR"
mkdir -p "$LOG_DIR"

# Copy binary
if [ -f "./cdrgenerator" ]; then
    echo_info "Installing binary..."
    cp ./cdrgenerator /usr/local/bin/
    chmod 755 /usr/local/bin/cdrgenerator
else
    echo_error "Binary not found. Please build first: go build -o cdrgenerator ."
    exit 1
fi

# Copy sample files
if [ -d "./samples" ]; then
    echo_info "Installing sample files..."
    cp -r ./samples/* "$INSTALL_DIR/samples/"
fi

# Copy example config if no config exists
if [ ! -f "$CONFIG_DIR/config.json" ]; then
    if [ -f "./configs/example-config.json" ]; then
        echo_info "Installing example configuration..."
        cp ./configs/example-config.json "$CONFIG_DIR/config.json"
        echo_warn "Please edit $CONFIG_DIR/config.json to configure your ports"
    fi
fi

# Install systemd service
if [ -f "./systemd/cdrgenerator.service" ]; then
    echo_info "Installing systemd service..."
    cp ./systemd/cdrgenerator.service /etc/systemd/system/
    systemctl daemon-reload
fi

# Set permissions
echo_info "Setting permissions..."
chown -R "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR"
chown -R "$SERVICE_USER:$SERVICE_USER" "$LOG_DIR"
chown root:root "$CONFIG_DIR"
chmod 755 "$CONFIG_DIR"
if [ -f "$CONFIG_DIR/config.json" ]; then
    chown root:"$SERVICE_USER" "$CONFIG_DIR/config.json"
    chmod 640 "$CONFIG_DIR/config.json"
fi

echo ""
echo_info "Installation complete!"
echo ""
echo "Next steps:"
echo "  1. Edit configuration: sudo nano $CONFIG_DIR/config.json"
echo "  2. Enable service: sudo systemctl enable cdrgenerator"
echo "  3. Start service: sudo systemctl start cdrgenerator"
echo "  4. Check status: sudo systemctl status cdrgenerator"
echo "  5. View logs: sudo journalctl -u cdrgenerator -f"
echo ""
echo "Health check endpoint: http://localhost:8080/health"
echo "Prometheus metrics: http://localhost:8080/metrics"
