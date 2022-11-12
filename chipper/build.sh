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

if [[ ! -f ./source.sh ]]; then
  echo "You need to make a source.sh file. This can be done with the setup.sh script."
  exit 0
fi

source source.sh

CGO_ENABLED=1 GOOS=linux GOARCH=${ARCH} /usr/local/go/bin/go build \
-tags ${STT_SERVICE} \
-ldflags "-w -s -extldflags "-static"" \
-trimpath \
-o chipper cmd/main.go
echo "Built chipper!"
