#!/bin/bash

export BUILDFILES="./cmd/wire-pod-installer"

export ARCHITECTURE=amd64

set -e

if [[ ! -f .aptDone ]]; then
    sudo apt install mingw-w64 zip build-essential autoconf unzip upx
    touch .aptDone
fi

export GOOS=windows
export ARCHITECTURE=amd64
export CGO_ENABLED=1
export CC=/usr/bin/x86_64-w64-mingw32-gcc
export CXX=/usr/bin/x86_64-w64-mingw32-g++


cd ..

x86_64-w64-mingw32-windres cmd/wire-pod-installer/rc/app.rc -O coff -o cmd/wire-pod-installer/app.syso

go build \
-ldflags "-H=windowsgui" \
-o windows/wire-pod-installer.exe \
${BUILDFILES}

#upx windows/wire-pod-installer.exe
