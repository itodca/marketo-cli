#!/usr/bin/env bash
set -euo pipefail

# marketo-cli installer
# curl -fsSL https://raw.githubusercontent.com/itodca/marketo-cli/main/install.sh | bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}✓${NC} $1"; }
warn()  { echo -e "${YELLOW}!${NC} $1"; }
error() { echo -e "${RED}✗${NC} $1"; exit 1; }

# --- Check Python ---
PYTHON=""
for cmd in python3 python; do
    if command -v "$cmd" &>/dev/null; then
        version=$("$cmd" --version 2>&1 | grep -oE '[0-9]+\.[0-9]+')
        major=$(echo "$version" | cut -d. -f1)
        minor=$(echo "$version" | cut -d. -f2)
        if [ "$major" -ge 3 ] && [ "$minor" -ge 10 ]; then
            PYTHON="$cmd"
            break
        fi
    fi
done

if [ -z "$PYTHON" ]; then
    echo ""
    error "Python 3.10+ is required but not found."
    echo ""
    echo "Install Python:"
    echo "  macOS:   brew install python"
    echo "  Ubuntu:  sudo apt install python3"
    echo "  Windows: https://python.org/downloads"
    echo ""
    exit 1
fi

info "Found $($PYTHON --version)"

# --- Check pip/uv ---
if command -v uv &>/dev/null; then
    info "Installing with uv..."
    uv pip install --system "marketo-cli @ git+https://github.com/itodca/marketo-cli.git"
elif command -v pipx &>/dev/null; then
    info "Installing with pipx..."
    pipx install "git+https://github.com/itodca/marketo-cli.git"
elif $PYTHON -m pip --version &>/dev/null; then
    info "Installing with pip..."
    $PYTHON -m pip install --user "marketo-cli @ git+https://github.com/itodca/marketo-cli.git"
else
    error "No package installer found. Install one of: uv, pipx, pip"
fi

# --- Verify ---
if command -v mrkto &>/dev/null; then
    echo ""
    info "marketo-cli installed successfully!"
    echo ""
    echo "  mrkto --help          Show all commands"
    echo "  mrkto auth check      Verify your credentials"
    echo "  mrkto lead user@co    Look up a lead by email"
    echo ""
    echo "Set credentials in .env or environment:"
    echo "  MARKETO_MUNCHKIN_ID=your-munchkin-id"
    echo "  MARKETO_CLIENT_ID=your-client-id"
    echo "  MARKETO_CLIENT_SECRET=your-client-secret"
else
    warn "Installed, but 'mrkto' not on PATH. You may need to restart your shell."
fi
