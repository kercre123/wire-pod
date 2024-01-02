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

if [[ $EUID != "0" ]]; then
  echo "This must be run as root."
  exit 1
fi

git fetch --all
git reset --hard origin/main
echo
echo "Updated!"
echo
