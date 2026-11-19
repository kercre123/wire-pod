#!/bin/bash

POD_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GO_VERSION="$(tr -d '[:space:]' < "${POD_DIR}/GO_VERSION")"

if [[ "$(uname -s)" == "Darwin" ]]; then
    GO_OS="darwin"
else
    GO_OS="linux"
fi

case "$(uname -m)" in
    x86_64) GO_ARCH="amd64" ;;
    aarch64|arm64) GO_ARCH="arm64" ;;
    armv7l|armv6l) GO_ARCH="armv6l" ;;
    *) GO_ARCH="unknown" ;;
esac

GO_DIR="${POD_DIR}/go/${GO_VERSION}-${GO_ARCH}"
GO="${GO_DIR}/bin/go"

VOSK_DIR="${POD_DIR}/vosk/libvosk"
if [[ ! -d "${VOSK_DIR}" ]] && [[ -d /root/.vosk/libvosk ]]; then
    VOSK_DIR="/root/.vosk/libvosk"
fi

function getGo() {
    if [[ -x "${GO}" ]]; then
        return 0
    fi
    if [[ ${GO_ARCH} == "unknown" ]]; then
        echo "Unsupported architecture for Go: $(uname -m)"
        return 1
    fi
    local tarball="go${GO_VERSION}.${GO_OS}-${GO_ARCH}.tar.gz"
    local tmp="${POD_DIR}/go/tmp-${GO_VERSION}-${GO_ARCH}"
    echo "Downloading Go ${GO_VERSION} (${GO_OS}-${GO_ARCH}) to ${GO_DIR}"
    rm -rf "${tmp}"
    mkdir -p "${tmp}"
    if ! curl -fL --progress-bar -o "${tmp}/${tarball}" "https://go.dev/dl/${tarball}"; then
        echo "Failed to download ${tarball}"
        rm -rf "${tmp}"
        return 1
    fi
    local want have
    want="$(curl -fsSL "https://dl.google.com/go/${tarball}.sha256")"
    if command -v sha256sum >/dev/null; then
        have="$(sha256sum "${tmp}/${tarball}" | cut -d' ' -f1)"
    else
        have="$(shasum -a 256 "${tmp}/${tarball}" | cut -d' ' -f1)"
    fi
    if [[ -z "${want}" ]] || [[ "${want}" != "${have}" ]]; then
        echo "Checksum mismatch for ${tarball}"
        rm -rf "${tmp}"
        return 1
    fi
    tar -C "${tmp}" -xzf "${tmp}/${tarball}"
    rm -rf "${GO_DIR}"
    mv "${tmp}/go" "${GO_DIR}"
    rm -rf "${tmp}"
    echo "Go ${GO_VERSION} installed to ${GO_DIR}"
}
