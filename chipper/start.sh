#!/bin/bash

UNAME=$(uname -a)
COMMIT_HASH="$(git rev-parse --short HEAD)"

if [[ $EUID -ne 0 ]]; then
    echo "This script must be run as root. sudo ./start.sh"
    exit 1
fi

if [[ -d ./chipper ]]; then
    cd chipper
fi

#if [[ ! -f ./chipper ]]; then
#   if [[ -f ./go.mod ]]; then
#     echo "You need to build chipper first. This can be done with the setup.sh script."
#   else
#     echo "You must be in the chipper directory."
#   fi
#   exit 0
#fi

if [[ ! -f ./source.sh ]]; then
    echo "You need to make a source.sh file. This can be done with the setup.sh script."
    exit 0
fi

source source.sh

# set go tags
export GOTAGS="nolibopusfile"

if [[ ${USE_INBUILT_BLE} == "true" ]]; then
    GOTAGS="${GOTAGS},inbuiltble"
fi

export GOLDFLAGS="-X 'github.com/kercre123/wire-pod/chipper/pkg/vars.CommitSHA=${COMMIT_HASH}'"

#./chipper
if [[ ${STT_SERVICE} == "leopard" ]]; then
    if [[ -f ./chipper ]]; then
        ./chipper
    else
        /usr/local/go/bin/go run -tags $GOTAGS -ldflags="${GOLDFLAGS}" cmd/leopard/main.go
    fi
    elif [[ ${STT_SERVICE} == "rhino" ]]; then
    if [[ -f ./chipper ]]; then
        ./chipper
    else
        /usr/local/go/bin/go run -tags $GOTAGS -ldflags="${GOLDFLAGS}" cmd/experimental/rhino/main.go
    fi
    elif [[ ${STT_SERVICE} == "houndify" ]]; then
    if [[ -f ./chipper ]]; then
        ./chipper
    else
        /usr/local/go/bin/go run -tags $GOTAGS -ldflags="${GOLDFLAGS}" cmd/experimental/houndify/main.go
    fi
    elif [[ ${STT_SERVICE} == "whisper" ]]; then
    if [[ -f ./chipper ]]; then
        ./chipper
    else
        /usr/local/go/bin/go run -tags $GOTAGS -ldflags="${GOLDFLAGS}" cmd/experimental/whisper/main.go
    fi
    elif [[ ${STT_SERVICE} == "whisper.cpp" ]]; then
    if [[ -f ./chipper ]]; then
        export C_INCLUDE_PATH="../whisper.cpp"
        export LIBRARY_PATH="../whisper.cpp"
        export LD_LIBRARY_PATH="$LD_LIBRARY_PATH:$(pwd)/../whisper.cpp:$(pwd)/../whisper.cpp/build:$(pwd)/../whisper.cpp/build/src"
        export CGO_LDFLAGS="-L$(pwd)/../whisper.cpp"
        export CGO_CFLAGS="-I$(pwd)/../whisper.cpp"
        ./chipper
    else
        export C_INCLUDE_PATH="../whisper.cpp"
        export LIBRARY_PATH="../whisper.cpp"
        export LD_LIBRARY_PATH="$LD_LIBRARY_PATH:$(pwd)/../whisper.cpp:$(pwd)/../whisper.cpp/build"
        export CGO_LDFLAGS="-L$(pwd)/../whisper.cpp -L$(pwd)/../whisper.cpp/build -L$(pwd)/../whisper.cpp/build/src -L$(pwd)/../whisper.cpp/build/ggml/src"
        export CGO_CFLAGS="-I$(pwd)/../whisper.cpp -I$(pwd)/../whisper.cpp/include -I$(pwd)/../whisper.cpp/ggml/include"
        if [[ ${UNAME} == *"Darwin"* ]]; then
            export GGML_METAL_PATH_RESOURCES="../whisper.cpp"
            /usr/local/go/bin/go run -tags $GOTAGS -ldflags "-extldflags '-framework Foundation -framework Metal -framework MetalKit'" cmd/experimental/whisper.cpp/main.go
        else
            /usr/local/go/bin/go run -tags $GOTAGS -ldflags="${GOLDFLAGS}" cmd/experimental/whisper.cpp/main.go
        fi
    fi
    elif [[ ${STT_SERVICE} == "vosk" ]]; then
    if [[ -f ./chipper ]]; then
        export CGO_ENABLED=1
        export CGO_CFLAGS="-I/root/.vosk/libvosk"
        export CGO_LDFLAGS="-L /root/.vosk/libvosk -lvosk -ldl -lpthread"
        export LD_LIBRARY_PATH="/root/.vosk/libvosk:$LD_LIBRARY_PATH"
        ./chipper
    else
        export CGO_ENABLED=1
        export CGO_CFLAGS="-I$HOME/.vosk/libvosk -I/root/.vosk/libvosk"
        export CGO_LDFLAGS="-L$HOME/.vosk/libvosk -L/root/.vosk/libvosk -lvosk -ldl -lpthread"
        export LD_LIBRARY_PATH="/root/.vosk/libvosk:$HOME/.vosk/libvosk:$LD_LIBRARY_PATH"
        /usr/local/go/bin/go run -tags $GOTAGS -ldflags="${GOLDFLAGS}" -exec "env DYLD_LIBRARY_PATH=$HOME/.vosk/libvosk" cmd/vosk/main.go
    fi
else
    if [[ -f ./chipper ]]; then
        export CGO_LDFLAGS="-L/root/.coqui/"
        export CGO_CXXFLAGS="-I/root/.coqui/"
        export LD_LIBRARY_PATH="/root/.coqui/:$LD_LIBRARY_PATH"
        ./chipper
    else
        export CGO_LDFLAGS="-L$HOME/.coqui/"
        export CGO_CXXFLAGS="-I$HOME/.coqui/"
        export LD_LIBRARY_PATH="$HOME/.coqui/:$LD_LIBRARY_PATH"
        /usr/local/go/bin/go run -tags $GOTAGS -ldflags="${GOLDFLAGS}" cmd/coqui/main.go
    fi
fi
