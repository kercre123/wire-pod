#!/usr/bin/env bash

set -euo pipefail

# Check for required commands early
for cmd in uname sudo git curl sync; do
    if ! command -v "${cmd}" > /dev/null 2>&1; then
        echo "Error: Required command '${cmd}' not found."
        exit 1
    fi
done

# Detect OS type and set TARGET
if [[ "$(uname -s)" == "Darwin" ]]; then
    TARGET="darwin"
    echo "macOS detected."
elif [[ -f /usr/bin/apt ]]; then
    TARGET="debian"
    echo "Debian-based Linux detected."
elif [[ -f /usr/bin/pacman ]]; then
    TARGET="arch"
    echo "Arch Linux detected."
elif [[ -f /usr/bin/dnf ]]; then
    TARGET="fedora"
    echo "Fedora/openSUSE detected."
else
    echo "Error: Unable to detect a supported operating system."
    exit 1
fi

# Install packages based on TARGET
case "${TARGET}" in
    debian)
        if ! command -v sudo > /dev/null 2>&1; then
            echo "Error: sudo not found. Please install sudo and rerun."
            exit 1
        fi
        sudo apt update -y || { echo "Error: apt update failed."; exit 1; }
        sudo apt install -y wget openssl net-tools libsox-dev libopus-dev make iproute2 xz-utils libopusfile-dev pkg-config gcc curl g++ unzip avahi-daemon git libasound2-dev libsodium-dev || { echo "Error: apt install failed."; exit 1; }
        ;;
    arch)
        sudo pacman -Sy --noconfirm || { echo "Error: pacman -Sy failed."; exit 1; }
        sudo pacman -S --noconfirm wget openssl net-tools sox opus make iproute2 opusfile curl unzip avahi git libsodium || { echo "Error: pacman install failed."; exit 1; }
        ;;
    fedora)
        sudo dnf update -y || { echo "Error: dnf update failed."; exit 1; }
        sudo dnf install -y wget openssl net-tools sox opus make opusfile curl unzip avahi git libsodium-devel || { echo "Error: dnf install failed."; exit 1; }
        ;;
    darwin)
        if ! command -v brew > /dev/null 2>&1; then
            echo "Error: Homebrew not found. Please install Homebrew and rerun."
            exit 1
        fi
        SUDO_USER="${SUDO_USER:-$(whoami)}"
        sudo -u "${SUDO_USER}" brew update || { echo "Error: brew update failed."; exit 1; }
        sudo -u "${SUDO_USER}" brew install wget pkg-config opus opusfile || { echo "Error: brew install failed."; exit 1; }
        ;;
    *)
        echo "Error: Unknown TARGET '${TARGET}'"
        exit 1
        ;;
esac

# Check if we are in the correct directory
if [[ ! -d ./chipper ]]; then
    echo "Error: This must be run in the wire-pod/ directory, which must contain 'chipper/'."
    exit 1
fi

# Update code
git fetch --all || { echo "Error: git fetch failed."; exit 1; }
git reset --hard origin/main || { echo "Error: git reset failed."; exit 1; }

# Proceed only if chipper binary or directory is present
if [[ -f ./chipper/chipper ]]; then
    cd chipper || { echo "Error: Failed to cd into chipper directory."; exit 1; }

    if [[ ! -f source.sh ]]; then
        echo "Error: source.sh not found in chipper directory."
        exit 1
    fi
    # shellcheck source=/dev/null
    source source.sh

    # Check STT_SERVICE variable
    if [[ -z "${STT_SERVICE:-}" ]]; then
        echo "Error: STT_SERVICE environment variable not set. Please define it in source.sh."
        exit 1
    fi

    # Check for required commands before build
    for cmd in go systemctl; do
        if ! command -v "${cmd}" > /dev/null 2>&1; then
            echo "Error: Required command '${cmd}' not found. Please install it and rerun."
            exit 1
        fi
    done

    # Attempt to stop the service if it's running
    sudo systemctl stop wire-pod || echo "Warning: Could not stop wire-pod service (it may not be running)."

    # Build based on STT_SERVICE
    case "${STT_SERVICE}" in
        leopard)
            echo "Building chipper with Picovoice (Leopard) STT service..."
            sudo /usr/local/go/bin/go build cmd/leopard/main.go || { echo "Error: go build (leopard) failed."; exit 1; }
            ;;
        vosk)
            echo "Building chipper with VOSK STT service..."
            export CGO_ENABLED=1
            sudo LD_LIBRARY_PATH="/root/.vosk/libvosk:$LD_LIBRARY_PATH" \
                 CGO_LDFLAGS="-L /root/.vosk/libvosk -lvosk -ldl -lpthread" \
                 CGO_CFLAGS="-I/root/.vosk/libvosk" \
                 /usr/local/go/bin/go build cmd/vosk/main.go || { echo "Error: go build (vosk) failed."; exit 1; }
            ;;
        coqui)
            echo "Building chipper with Coqui STT service..."
            sudo LD_LIBRARY_PATH="/root/.coqui/:$LD_LIBRARY_PATH" \
                 CGO_CXXFLAGS="-I/root/.coqui/" \
                 CGO_LDFLAGS="-L/root/.coqui/" \
                 /usr/local/go/bin/go build cmd/coqui/main.go || { echo "Error: go build (coqui) failed."; exit 1; }
            ;;
        *)
            echo "Error: Unsupported STT_SERVICE '${STT_SERVICE}'. Please update source.sh and rerun."
            exit 1
            ;;
    esac

    echo "Syncing..."
    sync || { echo "Warning: sync command failed."; }

    sudo systemctl daemon-reload || { echo "Error: systemctl daemon-reload failed."; exit 1; }
    sudo systemctl start wire-pod || { echo "Error: Could not start wire-pod service."; exit 1; }

    echo "wire-pod is now running with the updated code!"
else
    echo "No existing chipper binary found. Update completed, but build was not performed."
fi

echo
echo "Updated successfully!"
