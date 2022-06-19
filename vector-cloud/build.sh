#!/bin/bash
UNAME=$(uname -a)
if [[ "${UNAME}" == *"armv7l"* ]]; then
	mkdir -p build
	echo "Building vic-cloud (direct because host arch is armv7l)..."
  	/usr/local/go/bin/go build  \
	-tags nolibopusfile,vicos \
	--trimpath \
	-ldflags '-w -s -linkmode internal -extldflags "-static" -r /anki/lib' \
	-o build/vic-cloud \
	cloud/main.go
else
	echo "Building vic-cloud (docker)..."
	make docker-builder
	make vic-cloud
fi
