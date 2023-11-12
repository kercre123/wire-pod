#!/bin/bash

set -e

if [[ ! -d vic-toolchain ]]; then
	git clone https://github.com/kercre123/vic-toolchain
fi

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

upx vbuild/chipper
fi

cd vbuild

rm -rf tmp/botpack
mkdir -p tmp/botpack/wire-pod/chipper
mkdir tmp/botpack/wire-pod/certs
mkdir -p tmp/botpack/wire-pod/vector-cloud/build
mkdir tmp/botpack/wire-pod/lib

cp -r ../intent-data tmp/botpack/wire-pod/chipper/
cp ../weather-map.json tmp/botpack/wire-pod/chipper/
cp -r ../webroot tmp/botpack/wire-pod/chipper/
cp -r ../session-certs tmp/botpack/wire-pod/chipper/
cp -r ../jdocs tmp/botpack/wire-pod/chipper/
cp ../stttest.pcm tmp/botpack/wire-pod/chipper/
cp ../../vector-cloud/build/vic-cloud tmp/botpack/wire-pod/vector-cloud/build/

echo "export DEBUG_LOOGING=true" > tmp/botpack/wire-pod/chipper/source.sh
echo "export STT_SERVICE=vosk" >> tmp/botpack/wire-pod/chipper/source.sh
echo "export VOSK_WITH_GRAMMER=true" >> tmp/botpack/wire-pod/chipper/source.sh
echo "export JDOCS_ENABLE_PINGER=false" >> tmp/botpack/wire-pod/chipper/source.sh
echo "export WEBSERVER_PORT=8081" >> tmp/botpack/wire-pod/chipper/source.sh

cp custom-start.sh tmp/botpack/wire-pod/chipper/start.sh
cp chipper tmp/botpack/wire-pod/chipper/

cp built-libs/opus/libopus.so.0 tmp/botpack/wire-pod/lib/
cp built-libs/ogg/libogg.so.0 tmp/botpack/wire-pod/lib/
cp built-libs/vosk/libvosk.so tmp/botpack/wire-pod/lib/

tar -czvf botpack.tar.gz -C tmp/botpack .

rm -rf chipper tmp 
