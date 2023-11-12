#!/bin/bash

set -e

export BUILDFILE="cmd/windows/main.go cmd/windows/win-initwirepod.go"

export LIBS="$(pwd)/prebuilt"

export CC=/usr/bin/x86_64-w64-mingw32-gcc
export CXX=/usr/bin/x86_64-w64-mingw32-g++

export GOOS=windows
export GOARCH=amd64
export GO_TAGS="nolibopusfile"

export CGO_ENABLED=1
export CGO_LDFLAGS="-L${LIBS}/ogg/lib -L${LIBS}/opus/lib -L${LIBS}/vosk"
export CGO_CFLAGS="-I${LIBS}/ogg/include -I${LIBS}/opus/include -I${LIBS}/vosk"

cd ..

go build \
-tags ${GO_TAGS} \
-ldflags "-H=windowsgui" \
-o windows/chipper.exe \
${BUILDFILE}

cd windows

rm -rf tmp
mkdir -p tmp/wire-pod/chipper
mkdir tmp/wire-pod/certs
mkdir -p tmp/wire-pod/vector-cloud/build

cp -r ../intent-data tmp/wire-pod/chipper/
cp ../weather-map.json tmp/wire-pod/chipper/
cp -r ../webroot tmp/wire-pod/chipper/
cp -r ../session-certs tmp/wire-pod/chipper/
cp -r ../epod tmp/wire-pod/chipper/
cp -r ../jdocs tmp/wire-pod/chipper/
cp ../stttest.pcm tmp/wire-pod/chipper/
cp ../../vector-cloud/build/vic-cloud tmp/wire-pod/vector-cloud/build/
cp ../../vector-cloud/pod-bot-install.sh tmp/wire-pod/vector-cloud/

# echo "export DEBUG_LOGGING=true" > tmp/botpack/wire-pod/chipper/source.sh
# echo "export STT_SERVICE=vosk" >> tmp/botpack/wire-pod/chipper/source.sh

cp chipper.exe tmp/wire-pod/chipper/

cp ${LIBS}/opus/bin/libopus-0.dll tmp/wire-pod/chipper/
cp ${LIBS}/ogg/bin/libogg-0.dll tmp/wire-pod/chipper/
cp ${LIBS}/vosk/* tmp/wire-pod/chipper/

cd tmp

zip -r ../wire-pod-win.zip wire-pod

cd ..
rm -rf tmp
rm chipper.exe
