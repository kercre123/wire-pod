# syntax=docker/dockerfile:1.6

FROM --platform=$BUILDPLATFORM golang:bookworm AS builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG VOSK_VERSION=0.3.45
ARG COMMIT_SHA=unknown

ENV DEBIAN_FRONTEND=noninteractive \
    CGO_ENABLED=1

RUN apt-get update \ 
    && apt-get install -y --no-install-recommends \ 
        build-essential \ 
        ca-certificates \ 
        curl \
        git \
        libasound2-dev \
        libopus-dev \
        libopusfile-dev \
        libsox-dev \
        libsodium-dev \
        pkg-config \
        unzip \
        wget \
        gcc-aarch64-linux-gnu \
        g++-aarch64-linux-gnu \
        gcc-arm-linux-gnueabihf \
        g++-arm-linux-gnueabihf \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src

COPY chipper/go.mod chipper/go.sum ./chipper/

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd chipper && go mod download

COPY . .

RUN find . -type f -name '*.sh' -exec sed -i 's/\r$//' {} +

RUN set -eux; \
    case "${TARGETARCH}" in \
        amd64) VOSK_PKG="vosk-linux-x86_64-${VOSK_VERSION}.zip" ;; \
        arm64) VOSK_PKG="vosk-linux-aarch64-${VOSK_VERSION}.zip" ;; \
        arm) VOSK_PKG="vosk-linux-armv7l-${VOSK_VERSION}.zip" ;; \
        *) echo "Unsupported architecture: ${TARGETARCH}" >&2; exit 1 ;; \
    esac; \
    mkdir -p /opt/vosk; \
    curl -fsSL -o /tmp/vosk.zip "https://github.com/alphacep/vosk-api/releases/download/v${VOSK_VERSION}/${VOSK_PKG}"; \
    unzip /tmp/vosk.zip -d /tmp; \
    VOSK_DIR="$(find /tmp -maxdepth 1 -type d -name 'vosk*' | head -n1)"; \
    mv "${VOSK_DIR}" /opt/vosk/libvosk; \
    rm -rf /tmp/vosk.zip /tmp/vosk-*

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    set -eux; \
    BUILD_COMMIT="${COMMIT_SHA}"; \
    if [ "${BUILD_COMMIT}" = "unknown" ] || [ -z "${BUILD_COMMIT}" ]; then \
        BUILD_COMMIT=$(git rev-parse --short HEAD || echo "dev"); \
    fi; \
    GOOS_VALUE="${TARGETOS}"; \
    GOARCH_VALUE="${TARGETARCH}"; \
    if [ -z "${GOOS_VALUE}" ]; then \
        GOOS_VALUE=$(go env GOOS); \
    fi; \
    if [ -z "${GOARCH_VALUE}" ]; then \
        GOARCH_VALUE=$(go env GOARCH); \
    fi; \
    if [ "${TARGETARCH}" = "arm" ]; then \
        GOARM_VALUE="${TARGETVARIANT#v}"; \
        if [ -z "${GOARM_VALUE}" ]; then GOARM_VALUE=7; fi; \
        export GOARM=${GOARM_VALUE}; \
    fi; \
    mkdir -p /build; \
    cd /src/chipper; \
    CC_VALUE=""; \
    CXX_VALUE=""; \
    if [ "${GOARCH_VALUE}" = "arm64" ]; then \
        CC_VALUE="aarch64-linux-gnu-gcc"; \
        CXX_VALUE="aarch64-linux-gnu-g++"; \
    elif [ "${GOARCH_VALUE}" = "arm" ]; then \
        CC_VALUE="arm-linux-gnueabihf-gcc"; \
        CXX_VALUE="arm-linux-gnueabihf-g++"; \
    fi; \
    if [ -n "${CC_VALUE}" ]; then \
        export CC=${CC_VALUE}; \
    fi; \
    if [ -n "${CXX_VALUE}" ]; then \
        export CXX=${CXX_VALUE}; \
    fi; \
    GOOS=${GOOS_VALUE} GOARCH=${GOARCH_VALUE} \
    CGO_CFLAGS="-I/opt/vosk/libvosk" \
    CGO_LDFLAGS="-L/opt/vosk/libvosk -lvosk -ldl -lpthread" \
    go build -tags "nolibopusfile" -ldflags "-s -w -X github.com/kercre123/wire-pod/chipper/pkg/vars.CommitSHA=${BUILD_COMMIT}" \
        -o /build/chipper ./cmd/vosk; \
    echo "${BUILD_COMMIT}" >/build/.wirepod-version


FROM ubuntu:22.04 AS runtime

ARG COMMIT_SHA=unknown

ENV DEBIAN_FRONTEND=noninteractive \
    WIREPOD_DATA_DIR=/data \
    LD_LIBRARY_PATH=/opt/vosk/libvosk

RUN apt-get update \ 
    && apt-get install -y --no-install-recommends \ 
        avahi-daemon \ 
        avahi-utils \ 
        bash \ 
        ca-certificates \ 
        curl \ 
        git \ 
        iproute2 \ 
        libasound2 \ 
        libopus0 \ 
        libopusfile0 \ 
        libsodium23 \ 
        libsox3 \ 
        tzdata \ 
        unzip \ 
        wget \ 
    && rm -rf /var/lib/apt/lists/*

WORKDIR /opt/wire-pod

COPY --from=builder /opt/vosk/libvosk /opt/vosk/libvosk
COPY --from=builder /src /opt/wire-pod
COPY --from=builder /build/chipper /opt/wire-pod/chipper/chipper
COPY --from=builder /build/.wirepod-version /opt/wire-pod/.wirepod-version

RUN rm -rf .git && \
    chmod +x \
        /opt/wire-pod/setup.sh \
        /opt/wire-pod/update.sh \
        /opt/wire-pod/chipper/start.sh \
        /opt/wire-pod/docker/entrypoint.sh

VOLUME ["/data"]

EXPOSE 80 443 8080 8084

LABEL org.opencontainers.image.revision="${COMMIT_SHA}"

ENTRYPOINT ["/opt/wire-pod/docker/entrypoint.sh"]
CMD ["/opt/wire-pod/chipper/start.sh"]
