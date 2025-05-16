#!/usr/bin/env bash
set -euo pipefail

# Global Vars
UNAME="$(uname -a)"
COMMIT_HASH="$(git rev-parse --short HEAD || true)"  # 'true' ensures no error if .git not present

# Ensure script is run as root
if [[ $EUID -ne 0 ]]; then
  echo "ERROR: This script must be run as root. Try: sudo ./start.sh"
  exit 1
fi

# Move into chipper/ directory if it exists
if [[ -d "./chipper" ]]; then
  cd chipper
fi

# Ensure source.sh file exists
if [[ ! -f "./source.sh" ]]; then
  echo "ERROR: source.sh not found. Please create one with the setup.sh script."
  exit 1
fi

# Source environment variables
# shellcheck disable=SC1091
source "./source.sh"

# Ensure STT_SERVICE is set
if [[ -z "${STT_SERVICE:-}" ]]; then
  echo "ERROR: STT_SERVICE is not defined. Please set it in source.sh or your environment."
  exit 1
fi

# Default GOTAGS
export GOTAGS="nolibopusfile"

# Add inbuiltble tag if requested
if [[ "${USE_INBUILT_BLE:-}" == "true" ]]; then
  GOTAGS="${GOTAGS},inbuiltble"
fi

# Link commit hash into build flags
export GOLDFLAGS="-X 'github.com/kercre123/wire-pod/chipper/pkg/vars.CommitSHA=${COMMIT_HASH}'"

# Helper function to run Chipper or go build

run_chipper() {
  local main_file="$1"

  if [[ -f ./chipper ]]; then
    # If compiled binary exists, run it
    ./chipper
  else
    # Otherwise, run via go
    /usr/local/go/bin/go run -tags "${GOTAGS}" -ldflags="${GOLDFLAGS}" "${main_file}"
  fi
}


# Case statement to handle STT_SERVICE

case "${STT_SERVICE}" in
  leopard)
    run_chipper "cmd/leopard/main.go"
    ;;
  rhino)
    run_chipper "cmd/experimental/rhino/main.go"
    ;;
  houndify)
    run_chipper "cmd/experimental/houndify/main.go"
    ;;
  whisper)
    run_chipper "cmd/experimental/whisper/main.go"
    ;;
  whisper.cpp)
    # Exports specific to whisper.cpp
    export C_INCLUDE_PATH="../whisper.cpp"
    export LIBRARY_PATH="../whisper.cpp"
    export LD_LIBRARY_PATH="${LD_LIBRARY_PATH:-}:$(pwd)/../whisper.cpp:$(pwd)/../whisper.cpp/build"
    export CGO_LDFLAGS="-L$(pwd)/../whisper.cpp -L$(pwd)/../whisper.cpp/build -L$(pwd)/../whisper.cpp/build/src -L$(pwd)/../whisper.cpp/build/ggml/src"
    export CGO_CFLAGS="-I$(pwd)/../whisper.cpp -I$(pwd)/../whisper.cpp/include -I$(pwd)/../whisper.cpp/ggml/include"

    # For macOS Metal support:
    if [[ "${UNAME}" == *"Darwin"* ]]; then
      export GGML_METAL_PATH_RESOURCES="../whisper.cpp"
      if [[ -f ./chipper ]]; then
        ./chipper
      else
        /usr/local/go/bin/go run -tags "${GOTAGS}" \
          -ldflags "-extldflags '-framework Foundation -framework Metal -framework MetalKit'" \
          cmd/experimental/whisper.cpp/main.go
      fi
    else
      run_chipper "cmd/experimental/whisper.cpp/main.go"
    fi
    ;;
  vosk)
    # Exports specific to vosk
    export CGO_ENABLED=1
    export CGO_CFLAGS="-I${HOME:-/root}/.vosk/libvosk -I/root/.vosk/libvosk"
    export CGO_LDFLAGS="-L${HOME:-/root}/.vosk/libvosk -L/root/.vosk/libvosk -lvosk -ldl -lpthread"
    export LD_LIBRARY_PATH="/root/.vosk/libvosk:${HOME:-/root}/.vosk/libvosk:${LD_LIBRARY_PATH:-}"

    if [[ -f ./chipper ]]; then
      ./chipper
    else
      /usr/local/go/bin/go run -tags "${GOTAGS}" \
        -ldflags="${GOLDFLAGS}" \
        -exec "env DYLD_LIBRARY_PATH=${HOME:-/root}/.vosk/libvosk" \
        cmd/vosk/main.go
    fi
    ;;
  coqui|*)
    # Default to coqui if STT_SERVICE is 'coqui' or anything else
    # Exports specific to coqui
    export CGO_LDFLAGS="-L${HOME:-/root}/.coqui/"
    export CGO_CXXFLAGS="-I${HOME:-/root}/.coqui/"
    export LD_LIBRARY_PATH="${HOME:-/root}/.coqui/:${LD_LIBRARY_PATH:-}"

    run_chipper "cmd/coqui/main.go"
    ;;
esac

exit 0
