#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PYTHON_BIN="${PYTHON_BIN:-python3}"
BUILD_ROOT="${ROOT}/build/pyinstaller"
OUTPUT_DIR="${OUTPUT_DIR:-${ROOT}/dist/releases}"
BINARY_NAME="mrkto"

info() {
    printf 'OK %s\n' "$1"
}

error() {
    printf 'XX %s\n' "$1" >&2
    exit 1
}

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

checksum_file() {
    local file="$1"
    local dir
    local base
    dir="$(dirname "$file")"
    base="$(basename "$file")"
    if command -v shasum >/dev/null 2>&1; then
        (cd "$dir" && shasum -a 256 "$base" > "${base}.sha256")
        return
    fi
    if command -v sha256sum >/dev/null 2>&1; then
        (cd "$dir" && sha256sum "$base" > "${base}.sha256")
        return
    fi
    error "No checksum tool found"
}

command -v "$PYTHON_BIN" >/dev/null 2>&1 || error "Python not found: ${PYTHON_BIN}"
"$PYTHON_BIN" -m PyInstaller --version >/dev/null 2>&1 || error "PyInstaller is not installed. Run: ${PYTHON_BIN} -m pip install '.[build]'"

OS="$(detect_os)"
ARCH="$(detect_arch)"
ASSET_NAME="${BINARY_NAME}-${OS}-${ARCH}.tar.gz"

rm -rf "$BUILD_ROOT"
mkdir -p "$OUTPUT_DIR"

info "Building ${BINARY_NAME} with PyInstaller"
"$PYTHON_BIN" -m PyInstaller \
    --noconfirm \
    --clean \
    --onefile \
    --name "$BINARY_NAME" \
    --paths "${ROOT}/src" \
    --distpath "${BUILD_ROOT}/dist" \
    --workpath "${BUILD_ROOT}/work" \
    --specpath "${BUILD_ROOT}/spec" \
    "${ROOT}/src/mrkto/__main__.py"

info "Packaging ${ASSET_NAME}"
tar -C "${BUILD_ROOT}/dist" -czf "${OUTPUT_DIR}/${ASSET_NAME}" "$BINARY_NAME"
checksum_file "${OUTPUT_DIR}/${ASSET_NAME}"

info "Release artifacts written to ${OUTPUT_DIR}"
echo "  ${OUTPUT_DIR}/${ASSET_NAME}"
echo "  ${OUTPUT_DIR}/${ASSET_NAME}.sha256"
