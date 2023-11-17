#!/bin/bash

export BUILDFILES="./cmd/windows"

export CC=/usr/bin/x86_64-w64-mingw32-gcc
export CXX=/usr/bin/x86_64-w64-mingw32-g++
export PODHOST=x86_64-w64-mingw32
export ARCHITECTURE=amd64

set -e

if [[ ! -f .aptDone ]]; then
    sudo apt install mingw-w64 zip build-essential autoconf unzip
    touch .aptDone
fi

export ORIGDIR="$(pwd)"
export PODLIBS="${ORIGDIR}/libs"

mkdir -p "${PODLIBS}"

if [[ ! -d ogg ]] || [[ ! -d "${PODLIBS}/ogg" ]]; then
    echo "ogg directory doesn't exist. cloning and building"
    rm -rf ogg
    git clone https://github.com/xiph/ogg --depth=1
    cd ogg
    ./autogen.sh
    ./configure --host=${PODHOST} --prefix="${PODLIBS}/ogg"
    make -j
    make install
    cd "${ORIGDIR}"
fi

if [[ ! -d opus ]] || [[ ! -d "${PODLIBS}/opus" ]]; then
    echo "opus directory doesn't exist. cloning and building"
    rm -rf opus
    git clone https://github.com/xiph/opus --depth=1
    cd opus
    ./autogen.sh
    ./configure --host=${PODHOST} --prefix="${PODLIBS}/opus"
    make -j
    make install
    cd "${ORIGDIR}"
fi

if [[ ! -d ${PODLIBS}/vosk ]]; then
    echo "getting vosk from alphacep releases page"
    cd "${PODLIBS}"
    wget https://github.com/alphacep/vosk-api/releases/download/v0.3.45/vosk-win64-0.3.45.zip
    unzip vosk-win64-0.3.45.zip
    mv vosk-win64-0.3.45 vosk
    cd "${ORIGDIR}"
fi

export GOOS=windows
export ARCHITECTURE=amd64
export GO_TAGS="nolibopusfile"

export CGO_ENABLED=1
export CGO_LDFLAGS="-L${PODLIBS}/ogg/lib -L${PODLIBS}/opus/lib -L${PODLIBS}/vosk"
export CGO_CFLAGS="-I${PODLIBS}/ogg/include -I${PODLIBS}/opus/include -I${PODLIBS}/vosk"

cd ..

x86_64-w64-mingw32-windres cmd/windows/rc/app.rc -O coff -o cmd/windows/app.syso

go build \
-tags ${GO_TAGS} \
-ldflags "-H=windowsgui" \
-o windows/chipper.exe \
${BUILDFILES}

go build \
-tags ${GO_TAGS} \
-ldflags "-H=windowsgui" \
-o windows/uninstall.exe \
./cmd/wire-pod-installer/uninstall/main.go

cd windows

rm -rf tmp
mkdir -p tmp/wire-pod/chipper
mkdir -p tmp/wire-pod/vector-cloud/build

cp -r ../intent-data tmp/wire-pod/chipper/
cp ../weather-map.json tmp/wire-pod/chipper/
cp -r ../webroot tmp/wire-pod/chipper/
cp -r ../epod tmp/wire-pod/chipper/
cp ../stttest.pcm tmp/wire-pod/chipper/
cp ../../vector-cloud/build/vic-cloud tmp/wire-pod/vector-cloud/build/
cp ../../vector-cloud/pod-bot-install.sh tmp/wire-pod/vector-cloud/
cp -r icons tmp/wire-pod/chipper/

# echo "export DEBUG_LOGGING=true" > tmp/botpack/wire-pod/chipper/source.sh
# echo "export STT_SERVICE=vosk" >> tmp/botpack/wire-pod/chipper/source.sh

cp uninstall.exe tmp/wire-pod/
cp chipper.exe tmp/wire-pod/chipper/

cp ${PODLIBS}/opus/bin/libopus-0.dll tmp/wire-pod/chipper/
cp ${PODLIBS}/ogg/bin/libogg-0.dll tmp/wire-pod/chipper/
cp ${PODLIBS}/vosk/* tmp/wire-pod/chipper/
rm tmp/wire-pod/chipper/libvosk.lib

cd tmp

rm -rf ../wire-pod-win-${ARCHITECTURE}.zip

zip -r ../wire-pod-win-${ARCHITECTURE}.zip wire-pod

cd ..
rm -rf tmp
rm chipper.exe
rm uninstall.exe
