#!/bin/bash
set -e

echo

if [[ $EUID -ne 0 ]]; then
    echo "This script must be run as root. Try 'sudo ./build.sh'."
    exit 1
fi

if [[ ! -d ./chipper ]]; then
    echo "Script is not running in the wire-pod/ directory or chipper folder is missing. Exiting."
    exit 1
fi

if [[ ! -f ./chipper/source.sh ]]; then
    echo "You need to make a source.sh file. This can be done with the setup.sh script, option 6."
    exit 1
fi
source ./chipper/source.sh

cd chipper

if [[ $STT_SERVICE == "leopard" ]]; then
    echo "building chipper with Picovoice STT service..."
    /usr/local/go/bin/go build cmd/leopard/main.go
elif [[ $STT_SERVICE == "vosk" ]]; then
    echo "building chipper with VOSK STT service..."
    export CGO_ENABLED=1
    export CGO_CFLAGS="-I$HOME/.vosk/libvosk"
    export CGO_LDFLAGS="-L $HOME/.vosk/libvosk -lvosk -ldl -lpthread"
    export LD_LIBRARY_PATH="$HOME/.vosk/libvosk:$LD_LIBRARY_PATH"
    /usr/local/go/bin/go build cmd/vosk/main.go
else
    echo "building chipper with Coqui STT service..."
    export CGO_LDFLAGS="-L$HOME/.coqui/"
    export CGO_CXXFLAGS="-I$HOME/.coqui/"
    export LD_LIBRARY_PATH="$HOME/.coqui/:$LD_LIBRARY_PATH"
    /usr/local/go/bin/go build cmd/coqui/main.go
fi

mv main chipper

echo "./chipper/chipper has been built!"