FROM golang:1.18.4-buster

COPY ./sources.list /etc/apt/sources.list

RUN dpkg --add-architecture armel

RUN apt-get -o Acquire::Check-Valid-Until=false -o Acquire::http::Dl-Limit=500 update -y && apt -o Acquire::Check-Valid-Until=false install -y \
	g++-arm-linux-gnueabi \
	gcc-arm-linux-gnueabi \
	libc6-dev-armhf-cross \
	libopus-dev:armel \
	libogg-dev:armel \
	android-liblog-dev:armel android-liblog:armel \
	upx

ENV GOPATH=/go
ENV GOOS=linux
ENV GOARCH=arm
ENV GOARM=7
ENV CGO_ENABLED=1
ENV CGO_LDFLAGS="-L/anki/lib"
ENV CGO_CFLAGS="-I/anki/lib"
ENV CC=arm-linux-gnueabi-gcc
ENV PKG_CONFIG_PATH=/usr/lib/arm-linux-gnueabi/pkgconfig
ENV CXX="arm-linux-gnueabi-g++-8"
ENV GOCACHE=/tmp
