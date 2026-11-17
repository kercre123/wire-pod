#!/bin/bash
# Builds wire-pod from this checkout the same way as the official .deb
# (kercre123/WirePod, debian/pod), but with the Wyoming STT provider instead of
# the built-in Vosk.
#
#   wyoming-build/build.sh [workdir]
#
# Output in <workdir>/out: the wire-pod-wyoming binary plus webroot,
# intent-data and weather-map.json for /etc/wire-pod. Needs root (installs
# pkg-config and libopus-dev); Go is downloaded into <workdir>/go.
set -euo pipefail

WORK="${1:-/opt/build}"
POD_REF="${POD_REF:-f111378}"
HERE="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$HERE/.." && pwd)"

DEBIAN_FRONTEND=noninteractive apt-get install -y -qq --no-install-recommends \
    git curl ca-certificates gcc libc6-dev pkg-config libopus-dev binutils >/dev/null

mkdir -p "$WORK" && cd "$WORK"
rm -rf WirePod
git clone -q --filter=blob:none https://github.com/kercre123/WirePod.git
git -C WirePod checkout -q "$POD_REF"
git -C WirePod apply "$HERE/WirePod.patch"

GO_VERSION="$(tr -d '[:space:]' < "$ROOT/GO_VERSION")"
if ! go/bin/go version 2>/dev/null | grep -q "go${GO_VERSION} "; then
    rm -rf go
    curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" | tar xz
fi
export PATH="$WORK/go/bin:$PATH" GOPATH="$WORK/gopath" GOCACHE="$WORK/gocache" \
    GOTOOLCHAIN=local CGO_ENABLED=1

# build the packaged main against this checkout instead of the published module
(cd WirePod && go mod edit -replace "github.com/kercre123/wire-pod/chipper=$ROOT/chipper")

SHA="$(git -C "$ROOT" rev-parse --short HEAD 2>/dev/null || echo unknown)-wyoming"
rm -rf out && mkdir -p out
(cd WirePod/debian && go build -mod=mod -tags nolibopusfile,inbuiltble \
    -ldflags "-w -s -X github.com/kercre123/wire-pod/chipper/pkg/vars.CommitSHA=$SHA" \
    -o "$WORK/out/wire-pod-wyoming" ./pod/)
cp -r "$ROOT/chipper/webroot" "$ROOT/chipper/intent-data" "$ROOT/chipper/weather-map.json" out/

echo "built $WORK/out/wire-pod-wyoming ($SHA)"
# a binary built on a newer distro may not start on an older one
echo "needs glibc: $(objdump -T out/wire-pod-wyoming | grep -o 'GLIBC_[0-9.]*' | sort -Vu | tail -1)"
if ldd out/wire-pod-wyoming | grep -q vosk; then
    echo "WARNING: libvosk is linked"
fi
