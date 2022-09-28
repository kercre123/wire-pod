#!/bin/bash
UNAME=$(uname -a)
if [[ "${UNAME}" == *"armv7l"* ]]; then
	mkdir -p build
	echo "Building boot-anim (direct because host arch is armv7l)..."
  	/usr/local/go/bin/go build  \
	-tags nolibopusfile,vicos \
	--trimpath \
	-ldflags '-w -s -linkmode internal -extldflags "-static" -r /anki/lib' \
	-o build/boot-anim \
	boot-anim/raw.go
else
	echo "Building boot-anim (docker)..."
	make docker-builder
	make boot-anim
fi
