#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="${BUILD_DIR:-${ROOT}/build/go}"
OUTPUT_DIR="${OUTPUT_DIR:-${ROOT}/dist/releases}"
BINARY_NAME="mrkto"
VERSION="${VERSION:-dev}"
COMMIT="${COMMIT:-$(git -C "${ROOT}" rev-parse --short HEAD 2>/dev/null || echo unknown)}"
DATE="${DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"

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

detect_goarch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *) error "Unsupported architecture: $(uname -m)" ;;
    esac
}

detect_archive_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "x64" ;;
        arm64|aarch64) echo "arm64" ;;
        *) error "Unsupported architecture: $(uname -m)" ;;
    esac
}

write_checksums() {
    local output_dir="$1"
    local asset_name="$2"
    local checksum_path="${output_dir}/checksums.txt"

    if command -v shasum >/dev/null 2>&1; then
        (cd "$output_dir" && shasum -a 256 "$asset_name" > "$(basename "$checksum_path")")
        return
    fi
    if command -v sha256sum >/dev/null 2>&1; then
        (cd "$output_dir" && sha256sum "$asset_name" > "$(basename "$checksum_path")")
        return
    fi
    error "No checksum tool found"
}

command -v go >/dev/null 2>&1 || error "Go is not installed"

OS="$(detect_os)"
GOARCH="$(detect_goarch)"
ARCHIVE_ARCH="$(detect_archive_arch)"
ASSET_NAME="${BINARY_NAME}-${OS}-${ARCHIVE_ARCH}.tar.gz"

rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR" "$OUTPUT_DIR"

info "Building ${BINARY_NAME} for ${OS}/${GOARCH}"
(
    cd "$ROOT"
    CGO_ENABLED=0 GOOS="$OS" GOARCH="$GOARCH" \
        go build \
        -trimpath \
        -ldflags "-s -w -X github.com/itodca/marketo-cli/internal/version.Version=${VERSION} -X github.com/itodca/marketo-cli/internal/version.Commit=${COMMIT} -X github.com/itodca/marketo-cli/internal/version.Date=${DATE}" \
        -o "${BUILD_DIR}/${BINARY_NAME}" \
        ./cmd/mrkto
)

info "Smoke testing ${BINARY_NAME}"
"${BUILD_DIR}/${BINARY_NAME}" --help >/dev/null
"${BUILD_DIR}/${BINARY_NAME}" version >/dev/null

info "Packaging ${ASSET_NAME}"
tar -C "${BUILD_DIR}" -czf "${OUTPUT_DIR}/${ASSET_NAME}" "${BINARY_NAME}"
write_checksums "${OUTPUT_DIR}" "${ASSET_NAME}"

info "Release artifacts written to ${OUTPUT_DIR}"
echo "  ${OUTPUT_DIR}/${ASSET_NAME}"
echo "  ${OUTPUT_DIR}/checksums.txt"
