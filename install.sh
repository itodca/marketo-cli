#!/usr/bin/env bash
set -euo pipefail

OWNER="itodca"
REPO="marketo-cli"
BINARY="mrkto"
VERSION="${MRKTO_VERSION:-latest}"
INSTALL_DIR="${MRKTO_INSTALL_DIR:-$HOME/.local/bin}"
LEGACY_APP_DIR="${MRKTO_LEGACY_APP_DIR:-$HOME/.local/share/mrkto}"
RELEASE_BASE_URL="${MRKTO_RELEASE_BASE_URL:-}"
MODIFY_PATH=1
UNINSTALL=0

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}OK${NC} $1"; }
warn()  { echo -e "${YELLOW}!!${NC} $1"; }
error() { echo -e "${RED}XX${NC} $1" >&2; exit 1; }

usage() {
    cat <<EOF
Install the mrkto native binary from GitHub Releases.

Usage:
  curl -fsSL https://raw.githubusercontent.com/${OWNER}/${REPO}/main/install.sh | bash
  curl -fsSL https://raw.githubusercontent.com/${OWNER}/${REPO}/main/install.sh | bash -s -- --version vX.Y.Z

Options:
  --version <tag>       Release tag to install. Default: latest
  --install-dir <dir>   Install directory. Default: ${INSTALL_DIR}
  --uninstall           Remove the installed command and legacy app bundle
  --no-modify-path      Do not update your shell startup file
  --help                Show this help
EOF
}

require_cmd() {
    command -v "$1" >/dev/null 2>&1 || error "Required command not found: $1"
}

while [ $# -gt 0 ]; do
    case "$1" in
        --version)
            [ $# -ge 2 ] || error "--version requires a value"
            VERSION="$2"
            shift 2
            ;;
        --install-dir)
            [ $# -ge 2 ] || error "--install-dir requires a value"
            INSTALL_DIR="$2"
            shift 2
            ;;
        --uninstall)
            UNINSTALL=1
            shift
            ;;
        --no-modify-path)
            MODIFY_PATH=0
            shift
            ;;
        --help|-h)
            usage
            exit 0
            ;;
        *)
            error "Unknown argument: $1"
            ;;
    esac
done

detect_os() {
    case "$(uname -s)" in
        Darwin) echo "darwin" ;;
        Linux) echo "linux" ;;
        *) error "Unsupported operating system: $(uname -s)" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "x64" ;;
        arm64|aarch64) echo "arm64" ;;
        *) error "Unsupported architecture: $(uname -m)" ;;
    esac
}

release_base_url() {
    if [ -n "$RELEASE_BASE_URL" ]; then
        echo "$RELEASE_BASE_URL"
        return
    fi

    if [ "$VERSION" = "latest" ]; then
        echo "https://github.com/${OWNER}/${REPO}/releases/latest/download"
    else
        echo "https://github.com/${OWNER}/${REPO}/releases/download/${VERSION}"
    fi
}

path_contains() {
    case ":$PATH:" in
        *":$1:"*) return 0 ;;
        *) return 1 ;;
    esac
}

path_expr_for_shell() {
    if [ "$INSTALL_DIR" = "$HOME" ]; then
        printf '%s' '$HOME'
        return
    fi

    case "$INSTALL_DIR" in
        "$HOME"/*)
            printf '\$HOME/%s' "${INSTALL_DIR#"$HOME"/}"
            ;;
        *)
            printf '%s' "$INSTALL_DIR"
            ;;
    esac
}

shell_name() {
    basename "${SHELL:-}"
}

shell_rc_path() {
    case "$(shell_name)" in
        zsh) echo "$HOME/.zshrc" ;;
        bash) echo "$HOME/.bashrc" ;;
        fish) echo "$HOME/.config/fish/config.fish" ;;
        *) return 1 ;;
    esac
}

shell_path_line() {
    local path_expr
    path_expr="$(path_expr_for_shell)"
    case "$(shell_name)" in
        zsh|bash)
            printf 'export PATH="%s:$PATH"' "$path_expr"
            ;;
        fish)
            printf 'fish_add_path "%s"' "$path_expr"
            ;;
        *)
            return 1
            ;;
    esac
}

shell_reload_hint() {
    case "$(shell_name)" in
        zsh) echo "source ~/.zshrc" ;;
        bash) echo "source ~/.bashrc" ;;
        fish) echo "source ~/.config/fish/config.fish" ;;
        *) return 1 ;;
    esac
}

ensure_path_configured() {
    local rc_path line

    if path_contains "$INSTALL_DIR"; then
        info "${INSTALL_DIR} is already on PATH"
        return 0
    fi

    if [ "$MODIFY_PATH" -ne 1 ]; then
        warn "${INSTALL_DIR} is not on PATH"
        return 1
    fi

    rc_path="$(shell_rc_path)" || {
        warn "Could not determine which shell startup file to update"
        return 1
    }
    line="$(shell_path_line)" || {
        warn "Could not build a PATH update line for your shell"
        return 1
    }

    mkdir -p "$(dirname "$rc_path")"
    touch "$rc_path"

    if grep -F "$line" "$rc_path" >/dev/null 2>&1; then
        info "PATH already configured in ${rc_path}"
        return 0
    fi

    printf '\n%s\n' "$line" >>"$rc_path"
    info "Added ${INSTALL_DIR} to PATH in ${rc_path}"
    return 0
}

print_manual_path_help() {
    local path_expr
    path_expr="$(path_expr_for_shell)"

    case "$(shell_name)" in
        zsh|bash)
            echo "Add this line to your shell startup file:"
            echo "  export PATH=\"${path_expr}:\$PATH\""
            ;;
        fish)
            echo "Add this line to your fish config:"
            echo "  fish_add_path \"${path_expr}\""
            ;;
        *)
            echo "Add ${INSTALL_DIR} to your PATH manually."
            ;;
    esac
}

verify_checksum() {
    local workdir="$1"
    local checksum_file="$2"
    local archive_name="$3"
    local filtered_file="${workdir}/${archive_name}.sha256"

    grep -E "[[:space:]]${archive_name}\$" "${workdir}/${checksum_file}" > "${filtered_file}" || error "Could not find checksum for ${archive_name}"

    if command -v shasum >/dev/null 2>&1; then
        (cd "$workdir" && shasum -a 256 -c "$(basename "$filtered_file")")
        return
    fi

    if command -v sha256sum >/dev/null 2>&1; then
        (cd "$workdir" && sha256sum -c "$(basename "$filtered_file")")
        return
    fi

    warn "No checksum tool found; skipping checksum verification"
}

require_safe_dir() {
    case "$1" in
        ""|"/")
            error "Refusing to operate on unsafe directory path: $1"
            ;;
    esac
}

install_binary() {
    local extract_dir="$1"
    local artifact_path

    artifact_path="$(find "${extract_dir}" -type f -name "${BINARY}" | head -n 1)"
    [ -n "$artifact_path" ] || error "Extracted archive did not contain ${BINARY}"

    mkdir -p "$INSTALL_DIR"

    if [ -d "${INSTALL_DIR}/${BINARY}" ] && [ ! -L "${INSTALL_DIR}/${BINARY}" ]; then
        error "Cannot overwrite directory at ${INSTALL_DIR}/${BINARY}"
    fi

    rm -f "${INSTALL_DIR}/${BINARY}"
    install -m 0755 "${artifact_path}" "${INSTALL_DIR}/${BINARY}"

    info "Installed ${INSTALL_DIR}/${BINARY}"

    if [ -d "$LEGACY_APP_DIR" ]; then
        warn "Legacy app bundle still exists at ${LEGACY_APP_DIR}"
        warn "Run the installer with --uninstall if you want that old bundle removed."
    fi
}

uninstall_app() {
    require_safe_dir "$INSTALL_DIR"
    require_safe_dir "$LEGACY_APP_DIR"

    if [ -d "${INSTALL_DIR}/${BINARY}" ] && [ ! -L "${INSTALL_DIR}/${BINARY}" ]; then
        error "Refusing to remove directory at ${INSTALL_DIR}/${BINARY}"
    fi

    if [ -e "${INSTALL_DIR}/${BINARY}" ] || [ -L "${INSTALL_DIR}/${BINARY}" ]; then
        rm -f "${INSTALL_DIR}/${BINARY}"
        info "Removed ${INSTALL_DIR}/${BINARY}"
    fi

    if [ -e "$LEGACY_APP_DIR" ] || [ -L "$LEGACY_APP_DIR" ]; then
        rm -rf "$LEGACY_APP_DIR"
        info "Removed legacy bundle at ${LEGACY_APP_DIR}"
    fi

    echo ""
    info "mrkto uninstalled"
    exit 0
}

require_cmd curl
require_cmd tar
require_cmd find

if [ "$UNINSTALL" -eq 1 ]; then
    uninstall_app
fi

OS="$(detect_os)"
ARCH="$(detect_arch)"
ARCHIVE_NAME="${BINARY}-${OS}-${ARCH}.tar.gz"
CHECKSUM_NAME="checksums.txt"
BASE_URL="$(release_base_url)"
ARCHIVE_URL="${BASE_URL}/${ARCHIVE_NAME}"
CHECKSUM_URL="${BASE_URL}/${CHECKSUM_NAME}"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

info "Downloading ${ARCHIVE_NAME}"
curl -fsSL "$ARCHIVE_URL" -o "${TMPDIR}/${ARCHIVE_NAME}" || error "Failed to download ${ARCHIVE_URL}"
curl -fsSL "$CHECKSUM_URL" -o "${TMPDIR}/${CHECKSUM_NAME}" || error "Failed to download ${CHECKSUM_URL}"

info "Verifying checksum"
verify_checksum "$TMPDIR" "$CHECKSUM_NAME" "$ARCHIVE_NAME"

info "Extracting ${ARCHIVE_NAME}"
mkdir -p "${TMPDIR}/extract"
tar -xzf "${TMPDIR}/${ARCHIVE_NAME}" -C "${TMPDIR}/extract"

install_binary "${TMPDIR}/extract"

"${INSTALL_DIR}/${BINARY}" --help >/dev/null 2>&1 || error "Installed binary did not start correctly"

echo ""
if ensure_path_configured; then
    if path_contains "$INSTALL_DIR"; then
        info "Open a new terminal or run the command below if this shell cannot find ${BINARY} yet:"
    else
        info "Run the command below, or open a new terminal, before using ${BINARY}:"
    fi
    if shell_reload_hint >/dev/null 2>&1; then
        echo "  $(shell_reload_hint)"
    fi
else
    warn "${BINARY} may not be available until ${INSTALL_DIR} is added to PATH"
    print_manual_path_help
fi

echo ""
info "mrkto installed successfully"
echo ""
echo "  mrkto --help"
echo "  mrkto auth check"
echo "  mrkto lead list --email user@example.com"
