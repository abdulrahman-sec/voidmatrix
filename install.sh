#!/bin/sh
# voidmatrix installer script
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/abdulrahman-sec/voidmatrix/main/install.sh | sh

set -e

# Evaluate ANSI escape codes portably using printf
BOLD=$(printf '\033[1m')
GREEN=$(printf '\033[32m')
RED=$(printf '\033[31m')
CYAN=$(printf '\033[36m')
RESET=$(printf '\033[0m')

echo "${BOLD}=== VoidMatrix Installer ===${RESET}"
echo ""

# 1. Check if Go is installed
if ! command -v go >/dev/null 2>&1; then
    echo "${RED}Error: Go is not installed on this system.${RESET}"
    echo "Please download and install Go (Go 1.16+) from https://golang.org and rerun this script."
    exit 1
fi

# 2. Build the binary from source
echo "▶ Compiling from source..."
if ! go build -o voidmatrix .; then
    echo ""
    echo "${RED}❌ Build failed!${RESET}"
    exit 1
fi
echo "${GREEN}✔ Build complete${RESET}"
echo ""

# 3. Resolve installation directory
INSTALL_DIR="/usr/local/bin"
echo "▶ Installing to $INSTALL_DIR..."

if [ -w "$INSTALL_DIR" ]; then
    cp voidmatrix "$INSTALL_DIR/voidmatrix"
    chmod +x "$INSTALL_DIR/voidmatrix"
else
    echo "Requesting sudo permissions to copy binary..."
    sudo cp voidmatrix "$INSTALL_DIR/voidmatrix"
    sudo chmod +x "$INSTALL_DIR/voidmatrix"
fi
echo "${GREEN}✔ Installed successfully${RESET}"
echo ""

# 4. Display confirmation card
echo "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo "${BOLD}🚀 VoidMatrix is ready!${RESET}"
echo ""
echo "Run:"
echo "  ${BOLD}voidmatrix${RESET}"
echo ""
echo "Optional:"
echo "  voidmatrix --help"
echo ""
echo "Uninstall:"
echo "  sudo rm $INSTALL_DIR/voidmatrix"
echo "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
