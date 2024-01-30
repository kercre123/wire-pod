#!/bin/bash

# ensure all required packages are installed
if [[ $(uname -a) == *"Darwin"* ]]; then
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
fi

if [[ ${TARGET} == "debian" ]]; then
    sudo apt update -y
    sudo apt install -y wget openssl net-tools libsox-dev libopus-dev make iproute2 xz-utils libopusfile-dev pkg-config gcc curl g++ unzip avahi-daemon git libasound2-dev libsodium-dev
elif [[ ${TARGET} == "arch" ]]; then
    sudo pacman -Sy --noconfirm
    sudo pacman -S --noconfirm wget openssl net-tools sox opus make iproute2 opusfile curl unzip avahi git libsodium
elif [[ ${TARGET} == "fedora" ]]; then
    sudo dnf update
    sudo dnf install -y wget openssl net-tools sox opus make opusfile curl unzip avahi git libsodium-devel
elif [[ ${TARGET} == "darwin" ]]; then
    sudo -u $SUDO_USER brew update
    sudo -u $SUDO_USER brew install wget pkg-config opus opusfile
fi

if [[ ! -d ./chipper ]]; then
  echo "This must be run in the wire-pod/ directory."
  exit 1
fi

git fetch --all
git reset --hard origin/main
if [[ -f ./chipper/chipper ]]; then
    cd chipper
    source source.sh
    sudo systemctl stop wire-pod
    if [[ ${STT_SERVICE} == "leopard" ]]; then
        echo "wire-pod.service created, building chipper with Picovoice STT service..."
        sudo /usr/local/go/bin/go build cmd/leopard/main.go
    elif [[ ${STT_SERVICE} == "vosk" ]]; then
        echo "wire-pod.service created, building chipper with VOSK STT service..."
        export CGO_ENABLED=1
        sudo LD_LIBRARY_PATH="/root/.vosk/libvosk:$LD_LIBRARY_PATH" CGO_LDFLAGS="-L /root/.vosk/libvosk -lvosk -ldl -lpthread" CGO_CFLAGS="-I/root/.vosk/libvosk" CGO_ENABLED=1 /usr/local/go/bin/go build cmd/vosk/main.go
    elif [[ ${STT_SERVICE} == "coqui" ]]; then
        echo "wire-pod.service created, building chipper with Coqui STT service..."
        sudo LD_LIBRARY_PATH="/root/.coqui/:$LD_LIBRARY_PATH" CGO_CXXFLAGS="-I/root/.coqui/" CGO_LDFLAGS="-L/root/.coqui/" /usr/local/go/bin/go build cmd/coqui/main.go
    else
	echo "Unsupported STT ${STT_SERVICE}. You must build this manually. The code has been updated, though."
	exit 1
    fi
    echo "Syncing..."
    sync
    sudo systemctl daemon-reload
    sudo systemctl start wire-pod
    echo "wire-pod is now running with the updated code!"
fi
echo
echo "Updated successfully!"
echo
