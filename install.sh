#!/bin/sh
# voidmatrix installer script
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/yourusername/voidmatrix/main/install.sh | sh

set -e

# ANSI escape codes for styling
BOLD="\033[1m"
GREEN="\033[32m"
RED="\033[31m"
RESET="\033[0m"

echo "${BOLD}=== voidmatrix Installer ===${RESET}"

# 1. Check if Go is installed
if ! command -v go >/dev/null 2>&1; then
    echo "${RED}Error: Go is not installed on this system.${RESET}"
    echo "Please download and install Go (Go 1.16+) from https://golang.org and rerun this script."
    exit 1
fi

# 2. Build the binary from source
echo "Compiling voidmatrix from source..."
go build -o voidmatrix .

# 3. Resolve installation directory
INSTALL_DIR="/usr/local/bin"
if [ -w "$INSTALL_DIR" ]; then
    echo "Installing binary to $INSTALL_DIR..."
    cp voidmatrix "$INSTALL_DIR/voidmatrix"
    chmod +x "$INSTALL_DIR/voidmatrix"
else
    echo "Sudo permissions required to write to $INSTALL_DIR."
    sudo cp voidmatrix "$INSTALL_DIR/voidmatrix"
    sudo chmod +x "$INSTALL_DIR/voidmatrix"
fi

echo "${GREEN}${BOLD}=== Installation Successful! ===${RESET}"
echo "Type ${BOLD}voidmatrix${RESET} in your terminal to enter the matrix."
