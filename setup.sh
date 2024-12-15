#!/usr/bin/env bash

set -euo pipefail

echo

# Check for required tools upfront
for cmd in uname cut grep ip awk mkdir rm tar ln cat readlink openssl wget unzip curl git systemctl; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
        echo "Error: required command '$cmd' not found."
        exit 1
    fi
done

UNAME=$(uname -a)
ROOT="/root"
TARGET=""
ARCH=""
SANPrefix=""
address=""
isEscapePod=""
noCerts=""
AVXSUPPORT=""
SUDO_USER="${SUDO_USER:-$(whoami)}"

# Check for --non-interactive flag
NON_INTERACTIVE="${NON_INTERACTIVE:-false}"
if [[ $# -gt 0 && $1 == "--non-interactive" ]]; then
    NON_INTERACTIVE="true"
    # Shift out the --non-interactive argument
    shift
fi

# Detect OS/Target
if [[ ${UNAME} == *"Darwin"* ]]; then
    # macOS
    if [[ -f /usr/local/Homebrew/bin/brew ]] || [[ -f /opt/Homebrew/bin/brew ]]; then
        TARGET="darwin"
        ROOT="$HOME"
        echo "macOS detected."
        if [[ ! -f /usr/local/go/bin/go ]]; then
            if [[ -f /usr/local/bin/go ]]; then
                mkdir -p /usr/local/go/bin
                if [[ ! -f /usr/local/go/bin/go ]]; then
                    ln -s /usr/local/bin/go /usr/local/go/bin/go
                fi
            else
                echo "Go was not found. You must download it from https://go.dev/dl/ for your macOS."
                exit 1
            fi
        fi
    else
        echo "macOS detected, but 'brew' was not found."
        echo 'Install with: /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"'
        exit 1
    fi
elif [[ -f /usr/bin/apt ]]; then
    TARGET="debian"
    echo "Debian-based Linux detected."
elif [[ -f /usr/bin/pacman ]]; then
    TARGET="arch"
    echo "Arch Linux detected."
elif [[ -f /usr/bin/dnf ]]; then
    TARGET="fedora"
    echo "Fedora/openSUSE detected."
else
    echo "This OS is not supported. This script currently supports Linux with apt, pacman, or dnf, and macOS."
    if [[ $# -lt 1 || "$1" != "--bypass-target-check" ]]; then
        echo "If you would like to get the required packages yourself, run setup.sh with --bypass-target-check"
        echo "Required packages (for Debian/apt): wget openssl net-tools libsox-dev libopus-dev make iproute2 xz-utils libopusfile-dev pkg-config gcc curl g++ unzip avahi-daemon git"
        exit 1
    fi
fi

# Detect Architecture
if [[ "${UNAME}" == *"x86_64"* ]]; then
    ARCH="x86_64"
    echo "amd64 architecture confirmed."
elif [[ "${UNAME}" == *"aarch64"* ]] || [[ "${UNAME}" == *"arm64"* ]]; then
    ARCH="aarch64"
    echo "aarch64 architecture confirmed."
elif [[ "${UNAME}" == *"armv7l"* ]]; then
    ARCH="armv7l"
    echo "armv7l (32-bit) architecture detected."
    echo "WARN: Coqui and VOSK bindings may be broken for this platform."
else
    echo "Your CPU architecture is not supported. Supported: x86_64, aarch64, armv7l."
    exit 1
fi

if [[ $EUID -ne 0 ]]; then
    echo "This script must be run as root. Try: sudo ./setup.sh"
    exit 1
fi

if [[ ! -d ./chipper ]]; then
    echo "Script must run in wire-pod/ directory with 'chipper/' folder present."
    exit 1
fi

# Check for AVX if x86_64 and not Darwin
if [[ $# -eq 0 || ( $1 != "-f" && $1 != "scp" ) ]]; then
    if [[ ${ARCH} == "x86_64" && ${TARGET} != "darwin" && ${NON_INTERACTIVE} != "true" ]]; then
        if [[ -f /proc/cpuinfo ]]; then
            CPUINFO=$(cat /proc/cpuinfo)
            if [[ "${CPUINFO}" == *"avx"* ]]; then
                echo "AVX support confirmed."
            else
                echo "This CPU does not support AVX. Performance may not be optimal."
                AVXSUPPORT="noavx"
            fi
        fi
    fi
fi

echo "Checks have passed!"
echo

function getPackages() {
    echo "Installing required packages"
    if [[ ${TARGET} == "debian" ]]; then
        apt update -y
        apt install -y wget openssl net-tools libsox-dev libopus-dev make iproute2 xz-utils libopusfile-dev pkg-config gcc curl g++ unzip avahi-daemon git libasound2-dev libsodium-dev
    elif [[ ${TARGET} == "arch" ]]; then
        pacman -Sy --noconfirm
        pacman -S --noconfirm wget openssl net-tools sox opus make iproute2 opusfile curl unzip avahi git libsodium go pkg-config
    elif [[ ${TARGET} == "fedora" ]]; then
        dnf update -y
        dnf install -y wget openssl net-tools sox opus make opusfile curl unzip avahi git libsodium-devel
    elif [[ ${TARGET} == "darwin" ]]; then
        if ! command -v brew >/dev/null 2>&1; then
            echo "Homebrew not found. Please install it and rerun."
            exit 1
        fi
        sudo -u "$SUDO_USER" brew update
        sudo -u "$SUDO_USER" brew install wget pkg-config opus opusfile
    fi

    mkdir -p ./vector-cloud
    touch ./vector-cloud/packagesGotten
    echo
    echo "Installing golang binary package"
    mkdir -p golang
    pushd golang >/dev/null
    if [[ ${TARGET} != "darwin" && ${TARGET} != "arch" ]]; then
        if [[ ! -f /usr/local/go/bin/go ]]; then
            if [[ ${ARCH} == "x86_64" ]]; then
                wget -q --show-progress --no-check-certificate https://go.dev/dl/go1.22.4.linux-amd64.tar.gz
                rm -rf /usr/local/go && tar -C /usr/local -xzf go1.22.4.linux-amd64.tar.gz
            elif [[ ${ARCH} == "aarch64" ]]; then
                wget -q --show-progress --no-check-certificate https://go.dev/dl/go1.22.4.linux-arm64.tar.gz
                rm -rf /usr/local/go && tar -C /usr/local -xzf go1.22.4.linux-arm64.tar.gz
            elif [[ ${ARCH} == "armv7l" ]]; then
                wget -q --show-progress --no-check-certificate https://go.dev/dl/go1.22.4.linux-armv6l.tar.gz
                rm -rf /usr/local/go && tar -C /usr/local -xzf go1.22.4.linux-armv6l.tar.gz
            fi
            if [[ ! -f /usr/bin/go && ! -e /usr/bin/go ]]; then
                ln -s /usr/local/go/bin/go /usr/bin/go
            fi
        fi
    else
        echo "This is macOS or Arch; assuming Go is already installed."
        if [[ ${TARGET} == "arch" && ! -d /usr/local/go/bin ]]; then
            mkdir -p /usr/local/go/bin
            ln -s /usr/bin/go /usr/local/go/bin/go
        fi
    fi
    popd >/dev/null
    rm -rf golang
    echo
}

function sttServicePrompt() {
    if [[ ${NON_INTERACTIVE} == "true" ]]; then
        # Default to vosk if not specified
        # If STT env var is set, use that instead
        local defaultSTT="${STT:-vosk}"
        echo "${defaultSTT}"
        return
    fi

    echo
    echo "Which speech-to-text service would you like to use?"
    echo "1: Coqui (local)"
    echo "2: Picovoice Leopard (local, requires API key)"
    echo "3: VOSK (local, recommended default)"
    echo "4: Whisper (local, accurate, slower)"

    read -p "Enter a number (3): " sttServiceNum
    if [[ -z ${sttServiceNum} ]]; then
        echo "vosk"
    elif [[ ${sttServiceNum} == "1" ]]; then
        if [[ ${TARGET} == "darwin" ]]; then
            echo "Coqui is not supported on macOS. Defaulting to vosk."
            echo "vosk"
        else
            echo "coqui"
        fi
    elif [[ ${sttServiceNum} == "2" ]]; then
        echo "leopard"
    elif [[ ${sttServiceNum} == "3" ]]; then
        echo "vosk"
    elif [[ ${sttServiceNum} == "4" ]]; then
        echo "whisper"
    else
        echo "Invalid choice. Defaulting to vosk."
        echo "vosk"
    fi
}

function getSTT() {
    echo "export DEBUG_LOGGING=true" > ./chipper/source.sh
    rm -f ./chipper/pico.key

    if [[ -n "${STT:-}" ]]; then
        sttService="${STT}"
    else
        sttService=$(sttServicePrompt)
    fi

    if [[ ${sttService} == "leopard" ]]; then
        function picoApiPrompt() {
            if [[ ${NON_INTERACTIVE} == "true" ]]; then
                if [[ -z "${PICOVOICE_APIKEY:-}" ]]; then
                    echo "Leopard chosen non-interactively but PICOVOICE_APIKEY not set. Exiting."
                    exit 1
                fi
                picoKey="${PICOVOICE_APIKEY}"
            else
                echo
                echo "Create an account at https://console.picovoice.ai/ and enter the Access Key:"
                read -p "Enter your Access Key: " picoKey
                if [[ -z ${picoKey} ]]; then
                    echo "You must enter a key."
                    picoApiPrompt
                fi
            fi
        }
        picoApiPrompt
        echo "export STT_SERVICE=leopard" >> ./chipper/source.sh
        echo "export PICOVOICE_APIKEY=${picoKey}" >> ./chipper/source.sh
        echo "export PICOVOICE_APIKEY=${picoKey}" > ./chipper/pico.key
    elif [[ ${sttService} == "vosk" ]]; then
        echo "export STT_SERVICE=vosk" >> ./chipper/source.sh
        origDir="$(pwd)"
        if [[ ! -f ./vosk/completed ]]; then
            echo "Getting VOSK assets..."
            rm -fr ${ROOT}/.vosk
            mkdir ${ROOT}/.vosk
            cd ${ROOT}/.vosk
            VOSK_VER="0.3.45"
            if [[ ${TARGET} == "darwin" ]]; then
                VOSK_VER="0.3.42"
                VOSK_DIR="vosk-osx-${VOSK_VER}"
            elif [[ ${ARCH} == "x86_64" ]]; then
                VOSK_DIR="vosk-linux-x86_64-${VOSK_VER}"
            elif [[ ${ARCH} == "aarch64" ]]; then
                VOSK_DIR="vosk-linux-aarch64-${VOSK_VER}"
            elif [[ ${ARCH} == "armv7l" ]]; then
                VOSK_DIR="vosk-linux-armv7l-${VOSK_VER}"
            else
                echo "No suitable VOSK binary found for this architecture."
                exit 1
            fi
            VOSK_ARCHIVE="$VOSK_DIR.zip"
            wget -q --show-progress --no-check-certificate "https://github.com/alphacep/vosk-api/releases/download/v${VOSK_VER}/${VOSK_ARCHIVE}"
            unzip "$VOSK_ARCHIVE"
            mv "$VOSK_DIR" libvosk
            rm -fr "$VOSK_ARCHIVE"
            
            cd ${origDir}/chipper
            export CGO_ENABLED=1
            export CGO_CFLAGS="-I${ROOT}/.vosk/libvosk"
            export CGO_LDFLAGS="-L ${ROOT}/.vosk/libvosk -lvosk -ldl -lpthread"
            export LD_LIBRARY_PATH="${ROOT}/.vosk/libvosk:$LD_LIBRARY_PATH"
            /usr/local/go/bin/go get -u github.com/kercre123/vosk-api/go/... || true
            /usr/local/go/bin/go get github.com/kercre123/vosk-api || true
            /usr/local/go/bin/go install github.com/kercre123/vosk-api/go || true
            cd ${origDir}
            mkdir -p ./vosk
            touch ./vosk/completed
        else
            echo "VOSK assets already present."
        fi
    elif [[ ${sttService} == "whisper" ]]; then
        echo "export STT_SERVICE=whisper.cpp" >> ./chipper/source.sh
        origDir="$(pwd)"
        echo "Getting Whisper assets..."
        if [[ ! -d ./whisper.cpp ]]; then
            mkdir whisper.cpp
            cd whisper.cpp
            git clone https://github.com/ggerganov/whisper.cpp.git .
        else
            cd whisper.cpp
        fi
        function whichWhisperModel() {
            local availableModels="tiny base small medium large-v3 large-v3-q5_0"
            if [[ ${NON_INTERACTIVE} == "true" ]]; then
                # Default to tiny if not set
                local wmodel="${WHISPER_MODEL:-tiny}"
                if [[ ! " $availableModels " == *" ${wmodel} "* ]]; then
                    wmodel="tiny"
                fi
                echo $wmodel
                return
            fi

            echo
            echo "Which Whisper model would you like to use?"
            echo "Options: $availableModels"
            echo "(tiny is recommended)"
            echo
            read -p "Enter preferred model: " whispermodel
            if [[ -z ${whispermodel} ]]; then
                whispermodel="tiny"
            fi
            if [[ ! " $availableModels " == *" ${whispermodel} "* ]]; then
                echo "Invalid model. Defaulting to tiny."
                whispermodel="tiny"
            fi
            echo $whispermodel
        }
        chosenModel=$(whichWhisperModel)
        ./models/download-ggml-model.sh "$chosenModel"
        cd bindings/go
        make whisper
        cd ${origDir}
        echo "export WHISPER_MODEL=$chosenModel" >> ./chipper/source.sh
    else
        # coqui
        echo "export STT_SERVICE=coqui" >> ./chipper/source.sh
        if [[ ! -f ./stt/completed ]]; then
            echo "Getting STT assets..."
            if [[ -d /root/.coqui ]]; then
                rm -rf /root/.coqui
            fi
            origDir=$(pwd)
            mkdir /root/.coqui
            cd /root/.coqui
            if [[ ${ARCH} == "x86_64" ]]; then
                if [[ ${AVXSUPPORT:-} == "noavx" ]]; then
                    wget -q --show-progress --no-check-certificate https://wire.my.to/noavx-coqui/native_client.tflite.Linux.tar.xz
                else
                    wget -q --show-progress --no-check-certificate https://github.com/coqui-ai/STT/releases/download/v1.3.0/native_client.tflite.Linux.tar.xz
                fi
                tar -xf native_client.tflite.Linux.tar.xz
                rm -f ./native_client.tflite.Linux.tar.xz
            elif [[ ${ARCH} == "aarch64" ]]; then
                wget -q --show-progress --no-check-certificate https://github.com/coqui-ai/STT/releases/download/v1.3.0/native_client.tflite.linux.aarch64.tar.xz
                tar -xf native_client.tflite.linux.aarch64.tar.xz
                rm -f ./native_client.tflite.linux.aarch64.tar.xz
            elif [[ ${ARCH} == "armv7l" ]]; then
                wget -q --show-progress --no-check-certificate https://github.com/coqui-ai/STT/releases/download/v1.3.0/native_client.tflite.linux.armv7.tar.xz
                tar -xf native_client.tflite.linux.armv7.tar.xz
                rm -f ./native_client.tflite.linux.armv7.tar.xz
            else
                echo "No suitable Coqui binary for this architecture."
                exit 1
            fi
            cd ${origDir}/chipper
            export CGO_LDFLAGS="-L/root/.coqui/"
            export CGO_CXXFLAGS="-I/root/.coqui/"
            export LD_LIBRARY_PATH="/root/.coqui/:$LD_LIBRARY_PATH"
            /usr/local/go/bin/go get -u github.com/asticode/go-asticoqui/... || true
            /usr/local/go/bin/go get github.com/asticode/go-asticoqui || true
            /usr/local/go/bin/go install github.com/asticode/go-asticoqui || true
            cd ${origDir}
            mkdir -p stt
            cd stt

            function sttModelPrompt() {
                if [[ ${NON_INTERACTIVE} == "true" ]]; then
                    # Default to large_vocabulary
                    echo "large_vocabulary"
                    return
                fi
                echo
                echo "Which voice model would you like to use?"
                echo "1: large_vocabulary (~100MB)"
                echo "2: huge_vocabulary (~900MB)"
                echo
                read -p "Enter a number (1): " sttModelNum
                if [[ -z ${sttModelNum} ]]; then
                    echo "large_vocabulary"
                elif [[ ${sttModelNum} == "1" ]]; then
                    echo "large_vocabulary"
                elif [[ ${sttModelNum} == "2" ]]; then
                    echo "huge_vocabulary"
                else
                    echo "Invalid. Using default large_vocabulary."
                    echo "large_vocabulary"
                fi
            }
            sttModel=$(sttModelPrompt)
            if [[ -f model.scorer ]]; then
                rm -rf ./*
            fi
            if [[ ${sttModel} == "large_vocabulary" ]]; then
                echo "Getting STT model..."
                wget -O model.tflite -q --show-progress --no-check-certificate https://coqui.gateway.scarf.sh/english/coqui/v1.0.0-large-vocab/model.tflite
                echo "Getting STT scorer..."
                wget -O model.scorer -q --show-progress --no-check-certificate https://coqui.gateway.scarf.sh/english/coqui/v1.0.0-large-vocab/large_vocabulary.scorer
            elif [[ ${sttModel} == "huge_vocabulary" ]]; then
                echo "Getting STT model..."
                wget -O model.tflite -q --show-progress --no-check-certificate https://coqui.gateway.scarf.sh/english/coqui/v1.0.0-huge-vocab/model.tflite
                echo "Getting STT scorer..."
                wget -O model.scorer -q --show-progress --no-check-certificate https://coqui.gateway.scarf.sh/english/coqui/v1.0.0-huge-vocab/huge-vocabulary.scorer
            else
                echo "Invalid model specified, defaulting to large_vocabulary."
                wget -O model.tflite -q --show-progress --no-check-certificate https://coqui.gateway.scarf.sh/english/coqui/v1.0.0-large-vocab/model.tflite
                wget -O model.scorer -q --show-progress --no-check-certificate https://coqui.gateway.scarf.sh/english/coqui/v1.0.0-large-vocab/large_vocabulary.scorer
            fi
            echo
            touch completed
            echo "STT assets successfully downloaded!"
            cd ..
        else
            echo "STT assets already present."
        fi
    fi
}

function IPDNSPrompt() {
    if [[ ${NON_INTERACTIVE} == "true" ]]; then
        local defaultChoice="${SAN_CHOICE:-epod}"
        # Valid choices: IP, DNS, epod, nocerts
        case $defaultChoice in
            IP)
                SANPrefix="IP"
                ;;
            DNS)
                SANPrefix="DNS"
                ;;
            epod)
                isEscapePod="epod"
                ;;
            nocerts)
                noCerts="true"
                ;;
            *)
                isEscapePod="epod"
                ;;
        esac
        return
    fi

    echo
    echo "Would you like to use:"
    echo "1: IP address (recommended for OSKR Vectors)"
    echo "2: Domain"
    echo "3: escapepod.local (for production Vectors)"
    if [[ -d ./certs ]]; then
        echo "4: Keep certificates as is (no cert generation)"
    fi
    read -p "Enter a number (3): " yn
    if [[ -z ${yn} ]]; then
        isEscapePod="epod"
    elif [[ ${yn} == "1" ]]; then
        SANPrefix="IP"
    elif [[ ${yn} == "2" ]]; then
        SANPrefix="DNS"
    elif [[ ${yn} == "3" ]]; then
        isEscapePod="epod"
    elif [[ ${yn} == "4" ]]; then
        noCerts="true"
    else
        echo "Invalid choice. Using escapepod.local."
        isEscapePod="epod"
    fi
}

function IPPrompt() {
    if [[ ${TARGET} == "darwin" ]]; then
        IPADDRESS=$(ifconfig | grep "inet " | grep -v 127.0.0.1 | cut -d\  -f2 || true)
    else
        defInt=$(ip addr | awk '/state UP/ {print $2}' | sed 's/://g' | head -n 1)
        IPADDRESS=$(ip -4 addr show $defInt | grep -oP '(?<=inet\s)\d+(\.\d+){3}' || true)
    fi
    if [[ -z ${IPADDRESS} ]]; then
        IPADDRESS="192.168.1.100"
    fi

    if [[ ${NON_INTERACTIVE} == "true" ]]; then
        address="${IPADDRESS}"
        return
    fi

    read -p "Enter the IP address of this machine (${IPADDRESS}): " ipaddress
    if [[ -z ${ipaddress} ]]; then
        address=${IPADDRESS}
    else
        address=${ipaddress}
    fi
}

function DNSPrompt() {
    if [[ ${NON_INTERACTIVE} == "true" ]]; then
        # Default domain if not provided by env
        address="${DOMAIN_NAME:-example.com}"
        if [[ -z ${address} ]]; then
            address="example.com"
        fi
        return
    fi

    read -p "Enter the domain you would like to use: " dnsurl
    if [[ -z ${dnsurl} ]]; then
        echo "You must enter a domain."
        DNSPrompt
    fi
    address=${dnsurl}
}

function createServerConfig() {
    # Create a basic server_config.json
    if [[ ! -f ./chipper/source.sh ]]; then
        echo "Error: source.sh not found. Cannot create server_config.json."
        exit 1
    fi
    source ./chipper/source.sh
    if [[ ! -d ./certs ]]; then
        mkdir ./certs
    fi
    local finalAddress="${address}"
    if [[ ${isEscapePod:-} == "epod" ]]; then
        finalAddress="escapepod.local"
    fi

    if [[ -z "${STT_SERVICE:-}" ]]; then
        echo "STT_SERVICE not set. Cannot create server_config.json."
        exit 1
    fi

    cat > ./certs/server_config.json <<EOF
{
  "server_url": "https://${finalAddress}",
  "stt_service": "${STT_SERVICE}"
}
EOF
    echo "server_config.json created at ./certs/server_config.json"
}

function generateCerts() {
    echo
    echo "Creating certificates"
    echo
    IPDNSPrompt
    if [[ ${noCerts:-} != "true" ]]; then
        if [[ ${isEscapePod:-} != "epod" ]]; then
            if [[ ${SANPrefix:-} == "IP" ]]; then
                IPPrompt
            elif [[ ${SANPrefix:-} == "DNS" ]]; then
                DNSPrompt
            else
                # Default to escapepod if not chosen
                isEscapePod="epod"
            fi
            rm -f ./chipper/useepod || true
            rm -rf ./certs
            mkdir certs
            cd certs
            echo "${address}" >address
            echo "[req]" >san.conf
            echo "default_bits  = 4096" >>san.conf
            echo "default_md = sha256" >>san.conf
            echo "distinguished_name = req_distinguished_name" >>san.conf
            echo "x509_extensions = v3_req" >>san.conf
            echo "prompt = no" >>san.conf
            echo "[req_distinguished_name]" >>san.conf
            echo "C = US" >>san.conf
            echo "ST = VA" >>san.conf
            echo "L = SomeCity" >>san.conf
            echo "O = MyCompany" >>san.conf
            echo "OU = MyDivision" >>san.conf
            echo "CN = ${address}" >>san.conf
            echo "[v3_req]" >>san.conf
            echo "keyUsage = nonRepudiation, digitalSignature, keyEncipherment" >>san.conf
            echo "extendedKeyUsage = serverAuth" >>san.conf
            echo "subjectAltName = @alt_names" >>san.conf
            echo "[alt_names]" >>san.conf
            echo "${SANPrefix}.1 = ${address}" >>san.conf
            openssl req -x509 -nodes -days 730 -newkey rsa:2048 -keyout cert.key -out cert.crt -config san.conf
            echo
            echo "Certificates generated!"
            echo
            cd ..
        else
            echo
            echo "escapepod.local chosen."
            touch chipper/useepod
        fi
    fi
}

function scpToBot() {
    local botAddress="${botAddress:-}"
    local keyPath="${keyPath:-}"
    if [[ -z ${botAddress} ]]; then
        echo "Usage: sudo ./setup.sh scp <vector's ip> <path/to/ssh-key>"
        echo "Or without SSH key if your Vector uses Wire's custom software:"
        echo "sudo ./setup.sh scp <vector's ip>"
        exit 1
    fi
    if [[ ! -f ./certs/server_config.json ]]; then
        echo "server_config.json missing. Generate it first."
        exit 1
    fi
    if [[ -z ${keyPath} ]]; then
        if [[ ! -f ./ssh_root_key ]]; then
            echo "No key provided, downloading ssh_root_key..."
            wget http://wire.my.to:81/ssh_root_key
        else
            echo "No key provided, using existing ./ssh_root_key..."
        fi
        chmod 600 ./ssh_root_key
        keyPath="./ssh_root_key"
    fi
    if [[ ! -f ${keyPath} ]]; then
        echo "Provided key not found."
        exit 1
    fi
    ssh -oStrictHostKeyChecking=no -i "${keyPath}" root@"${botAddress}" "cat /build.prop" >/tmp/sshTest 2>>/tmp/sshTest || true
    botBuildProp=$(cat /tmp/sshTest || true)
    if [[ "${botBuildProp}" == *"no mutual signature"* ]]; then
        echo
        echo "SSH config needs updating (PubkeyAcceptedKeyTypes +ssh-rsa). Attempting to fix."
        echo "PubkeyAcceptedKeyTypes +ssh-rsa" >>/etc/ssh/ssh_config
        botBuildProp=$(ssh -oStrictHostKeyChecking=no -i "${keyPath}" root@"${botAddress}" "cat /build.prop" || true)
    fi
    if [[ ! "${botBuildProp}" == *"ro.build"* ]]; then
        echo "Unable to communicate with robot. Check network or keys."
        exit 1
    fi
    scpTest=$(scp -v -oStrictHostKeyChecking=no -i "${keyPath}" root@"${botAddress}":/build.prop /tmp/scpTest 2>&1 || true)
    if [[ "${scpTest}" == *"sftp"* ]]; then
        oldVar="-O"
    else
        oldVar=""
    fi
    if [[ ! "${botBuildProp}" == *"ro.build"* ]]; then
        echo "Unable to communicate with robot."
        exit 1
    fi
    ssh -oStrictHostKeyChecking=no -i "${keyPath}" root@"${botAddress}" "mount -o rw,remount / && mount -o rw,remount,exec /data && systemctl stop anki-robot.target && mv /anki/data/assets/cozmo_resources/config/server_config.json /anki/data/assets/cozmo_resources/config/server_config.json.bak"
    scp -oStrictHostKeyChecking=no ${oldVar} -i "${keyPath}" ./vector-cloud/build/vic-cloud root@"${botAddress}":/anki/bin/
    scp -oStrictHostKeyChecking=no ${oldVar} -i "${keyPath}" ./certs/server_config.json root@"${botAddress}":/anki/data/assets/cozmo_resources/config/
    scp -oStrictHostKeyChecking=no ${oldVar} -i "${keyPath}" ./vector-cloud/pod-bot-install.sh root@"${botAddress}":/data/
    if [[ -f ./chipper/useepod ]]; then
        scp -oStrictHostKeyChecking=no ${oldVar} -i "${keyPath}" ./chipper/epod/ep.crt root@"${botAddress}":/anki/etc/wirepod-cert.crt
        scp -oStrictHostKeyChecking=no ${oldVar} -i "${keyPath}" ./chipper/epod/ep.crt root@"${botAddress}":/data/data/wirepod-cert.crt
    else
        scp -oStrictHostKeyChecking=no ${oldVar} -i "${keyPath}" ./certs/cert.crt root@"${botAddress}":/anki/etc/wirepod-cert.crt
        scp -oStrictHostKeyChecking=no ${oldVar} -i "${keyPath}" ./certs/cert.crt root@"${botAddress}":/data/data/wirepod-cert.crt
    fi
    ssh -oStrictHostKeyChecking=no -i "${keyPath}" root@"${botAddress}" "chmod +rwx /anki/data/assets/cozmo_resources/config/server_config.json /anki/bin/vic-cloud /data/data/wirepod-cert.crt /anki/etc/wirepod-cert.crt /data/pod-bot-install.sh && /data/pod-bot-install.sh"
    rm -f /tmp/sshTest /tmp/scpTest
    echo "Vector has been reset to Onboarding mode. No user data erased."
    echo "Setup complete. Run: sudo ./chipper/start.sh"
    echo
}

function setupSystemd() {
    if [[ ${TARGET} == "darwin" ]]; then
        echo "Systemd setup not supported on macOS."
        exit 1
    fi
    if [[ ! -f ./chipper/source.sh ]]; then
        echo "source.sh not found. Run setup options to create it first."
        exit 1
    fi
    source ./chipper/source.sh
    echo "[Unit]" >wire-pod.service
    echo "Description=Wire Escape Pod" >>wire-pod.service
    echo "StartLimitIntervalSec=500" >>wire-pod.service
    echo "StartLimitBurst=5" >>wire-pod.service
    echo >>wire-pod.service
    echo "[Service]" >>wire-pod.service
    echo "Type=simple" >>wire-pod.service
    echo "Restart=on-failure" >>wire-pod.service
    echo "RestartSec=5s" >>wire-pod.service
    echo "WorkingDirectory=$(readlink -f ./chipper)" >>wire-pod.service
    echo "ExecStart=$(readlink -f ./chipper/start.sh)" >>wire-pod.service
    echo >>wire-pod.service
    echo "[Install]" >>wire-pod.service
    echo "WantedBy=multi-user.target" >>wire-pod.service
    cat wire-pod.service
    echo
    cd chipper
    export GOTAGS="nolibopusfile"
    if [[ ${USE_INBUILT_BLE:-} == "true" ]]; then
        GOTAGS="nolibopusfile,inbuiltble"
    fi
    COMMIT_HASH="$(git rev-parse --short HEAD)"
    export GOLDFLAGS="-X 'github.com/kercre123/wire-pod/chipper/pkg/vars.CommitSHA=${COMMIT_HASH}'"
    if [[ ${STT_SERVICE} == "leopard" ]]; then
        echo "Building chipper with Picovoice STT..."
        /usr/local/go/bin/go build -tags $GOTAGS -ldflags="${GOLDFLAGS}" cmd/leopard/main.go
    elif [[ ${STT_SERVICE} == "vosk" ]]; then
        echo "Building chipper with VOSK STT..."
        export CGO_ENABLED=1
        export CGO_CFLAGS="-I${ROOT}/.vosk/libvosk"
        export CGO_LDFLAGS="-L ${ROOT}/.vosk/libvosk -lvosk -ldl -lpthread"
        export LD_LIBRARY_PATH="${ROOT}/.vosk/libvosk:$LD_LIBRARY_PATH"
        /usr/local/go/bin/go build -tags $GOTAGS -ldflags="${GOLDFLAGS}" cmd/vosk/main.go
    elif [[ ${STT_SERVICE} == "whisper.cpp" ]]; then
        echo "Building chipper with Whisper.CPP STT..."
        export CGO_ENABLED=1
        export C_INCLUDE_PATH="../whisper.cpp"
        export LIBRARY_PATH="../whisper.cpp"
        export LD_LIBRARY_PATH="$LD_LIBRARY_PATH:$(pwd)/../whisper.cpp"
        export CGO_LDFLAGS="-L$(pwd)/../whisper.cpp"
        export CGO_CFLAGS="-I$(pwd)/../whisper.cpp"
        /usr/local/go/bin/go build -tags $GOTAGS -ldflags="${GOLDFLAGS}" cmd/experimental/whisper.cpp/main.go
    else
        echo "Building chipper with Coqui STT..."
        export CGO_LDFLAGS="-L/root/.coqui/"
        export CGO_CXXFLAGS="-I/root/.coqui/"
        export LD_LIBRARY_PATH="/root/.coqui/:$LD_LIBRARY_PATH"
        /usr/local/go/bin/go build -tags $GOTAGS -ldflags="${GOLDFLAGS}" cmd/coqui/main.go
    fi
    sync
    mv main chipper
    echo
    echo "./chipper/chipper has been built!"
    cd ..
    mv wire-pod.service /lib/systemd/system/
    systemctl daemon-reload
    systemctl enable wire-pod
    echo
    echo "systemd service installed and enabled (wire-pod.service)!"
    echo "To start the service: 'systemctl start wire-pod'"
    echo "To see logs: 'journalctl -fe | grep start.sh'"
}

function disableSystemd() {
    if [[ ${TARGET} == "darwin" ]]; then
        echo "This cannot be done on macOS."
        exit 1
    fi
    echo
    echo "Disabling wire-pod.service"
    systemctl stop wire-pod.service || true
    systemctl disable wire-pod.service || true
    rm -f ./chipper/chipper
    rm -f /lib/systemd/system/wire-pod.service
    systemctl daemon-reload
    echo
    echo "wire-pod.service has been removed and disabled."
}

function fullSetup() {
    getPackages
    getSTT
    generateCerts
    createServerConfig
    echo
    echo "Full setup complete! You can now run: sudo ./chipper/start.sh"
}

function justBuildVicCloud() {
    echo "This option (build vic-cloud) is not fully implemented."
    echo "Please add logic if needed."
}

function justBuildChipper() {
    getPackages
    getSTT
    echo "Building chipper..."
    setupSystemd
}

function justGetSTT() {
    getPackages
    getSTT
    echo "STT assets acquired."
}

function justGenerateCerts() {
    generateCerts
    if [[ ${noCerts:-} != "true" && ${isEscapePod:-} != "epod" ]]; then
        createServerConfig
    fi
    echo "Certificates generated and server_config.json created if applicable."
}

function createWirePodConfig() {
    echo "Reconfiguring wire-pod..."
    getSTT
    if [[ -d ./certs ]]; then
        createServerConfig
    fi
    echo "Configuration updated!"
}

function defaultLaunch() {
    fullSetup
}

# Handle arguments
if [[ $# -gt 0 ]]; then
    arg="$1"
    case "$arg" in
        scp)
            botAddress=${2:-}
            keyPath=${3:-}
            scpToBot
            exit 0
            ;;
        daemon-enable)
            getPackages
            setupSystemd
            exit 0
            ;;
        daemon-disable)
            disableSystemd
            exit 0
            ;;
        -f)
            if [[ ${2:-} == "scp" ]]; then
                botAddress=$3
                keyPath=$4
                scpToBot
                exit 0
            else
                echo "Unknown option after -f"
                exit 1
            fi
            ;;
        --bypass-target-check)
            # If bypass-target-check is used standalone, run default launch
            defaultLaunch
            exit 0
            ;;
        *)
            echo "Unknown argument provided."
            exit 1
            ;;
    esac
else
    # No arguments: Show a menu if interactive, else run full setup
    if [[ ${NON_INTERACTIVE} == "true" ]]; then
        defaultLaunch
    else
        echo "What would you like to do?"
        echo "1: Full Setup (recommended)"
        echo "2: Just build vic-cloud"
        echo "3: Just build chipper"
        echo "4: Just get STT assets"
        echo "5: Just generate certs"
        echo "6: Create wire-pod config file (API keys, STT selection)"
        echo "(Press enter for the default: Full Setup)"
        echo
        read -p "Enter a number (1): " choice
        if [[ -z ${choice} ]]; then
            choice="1"
        fi
        case $choice in
            1) fullSetup ;;
            2) justBuildVicCloud ;;
            3) justBuildChipper ;;
            4) justGetSTT ;;
            5) justGenerateCerts ;;
            6) createWirePodConfig ;;
            *) echo "Invalid choice, running full setup by default." ; fullSetup ;;
        esac
    fi
fi
