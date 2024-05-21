#!/bin/bash

UNAME=$(uname -a)

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

# sets env vars: (i.e. STT_SERVICE)
source source.sh

# set go tags
export GOTAGS="nolibopusfile"

if [[ ${USE_INBUILT_BLE} == "true" ]]; then
    GOTAGS="${GOTAGS},inbuiltble"
fi

#./chipper
if [[ ${STT_SERVICE} == "leopard" ]]; then
    if [[ -f ./chipper ]]; then
        ./chipper
    else
        /usr/local/go/bin/go run -tags $GOTAGS cmd/leopard/main.go
    fi
elif [[ ${STT_SERVICE} == "rhino" ]]; then
    if [[ -f ./chipper ]]; then
        ./chipper
    else
        /usr/local/go/bin/go run -tags $GOTAGS cmd/experimental/rhino/main.go
    fi
elif [[ ${STT_SERVICE} == "houndify" ]]; then
    if [[ -f ./chipper ]]; then
        ./chipper
    else
        /usr/local/go/bin/go run -tags $GOTAGS cmd/experimental/houndify/main.go
    fi
elif [[ ${STT_SERVICE} == "whisper" ]]; then
    if [[ -f ./chipper ]]; then
        ./chipper
    else
        /usr/local/go/bin/go run -tags $GOTAGS cmd/experimental/whisper/main.go
    fi
elif [[ ${STT_SERVICE} == "whisper.cpp" ]]; then
    if [[ -f ./chipper ]]; then
        export C_INCLUDE_PATH="../whisper.cpp"
        export LIBRARY_PATH="../whisper.cpp"
        export LD_LIBRARY_PATH="$LD_LIBRARY_PATH:$(pwd)/../whisper.cpp"
        export CGO_LDFLAGS="-L$(pwd)/../whisper.cpp"
        export CGO_CFLAGS="-I$(pwd)/../whisper.cpp"
        ./chipper
    else
        export C_INCLUDE_PATH="../whisper.cpp"
        export LIBRARY_PATH="../whisper.cpp"
        export LD_LIBRARY_PATH="$LD_LIBRARY_PATH:$(pwd)/../whisper.cpp"
        export CGO_LDFLAGS="-L$(pwd)/../whisper.cpp"
        export CGO_CFLAGS="-I$(pwd)/../whisper.cpp"
        if [[ ${UNAME} == *"Darwin"* ]]; then
            export GGML_METAL_PATH_RESOURCES="../whisper.cpp"
            /usr/local/go/bin/go run -tags $GOTAGS -ldflags "-extldflags '-framework Foundation -framework Metal -framework MetalKit'" cmd/experimental/whisper.cpp/main.go
        else
            /usr/local/go/bin/go run -tags $GOTAGS cmd/experimental/whisper.cpp/main.go
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
        export CGO_CFLAGS="-I/root/.vosk/libvosk"
        export CGO_LDFLAGS="-L /root/.vosk/libvosk -lvosk -ldl -lpthread"
        export LD_LIBRARY_PATH="/root/.vosk/libvosk:$LD_LIBRARY_PATH"
        if [[ ${DO_GO_BUILD} == "true" ]]; then
            /usr/local/go/bin/go build -tags $GOTAGS -o chipper cmd/vosk/main.go
        else
            /usr/local/go/bin/go run -tags $GOTAGS -exec "env DYLD_LIBRARY_PATH=$HOME/.vosk/libvosk" cmd/vosk/main.go
        fi
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
        /usr/local/go/bin/go run -tags $GOTAGS cmd/coqui/main.go
    fi
fi
