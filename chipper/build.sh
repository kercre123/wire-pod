#!/bin/bash
UNAME=$(uname -a)
echo "Building chipper..."
if [[ "${UNAME}" == *"aarch64"* ]]; then
   ARCH=arm64
elif [[ "${UNAME}" == *"armv7l"* ]]; then
   ARCH=arm
elif [[ "${UNAME}" == *"x86_64"* ]]; then
   ARCH=amd64
fi
CGO_ENABLED=1 GOOS=linux GOARCH=${ARCH} /usr/local/go/bin/go build \
-tags vosk
-ldflags "-w -s -extldflags "-static"" \
-trimpath \
-o chipper cmd/main.go
echo "Built chipper!"
