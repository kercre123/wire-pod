#!/bin/bash

export CHAIN="$(pwd)/vic-toolchain/arm-linux-gnueabi/bin/arm-linux-gnueabi-"

export CC=${CHAIN}gcc
export CXX=${CHAIN}g++

export CGO_LDFLAGS="-L$(pwd)/built-libs/opus -L$(pwd)/built-libs/vosk -L$(pwd)/built-libs/ogg"
export CGO_CFLAGS="-I$(pwd)/built-libs/opus/include -I$(pwd)/built-libs/ogg/include -I$(pwd)/built-libs/vosk"

export GOARM=7
export GOARCH=arm
export CGO_ENABLED=1

cd ../

echo "Building chipper for Vector..."

if [[ ! $1 == "-s" ]]; then
go build \
-tags nolibopusfile \
-o vbuild/chipper \
cmd/vosk/main.go

#upx vbuild/chipper
fi
