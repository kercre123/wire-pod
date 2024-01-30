#!/bin/bash

set -e

echo

UNAME=$(uname -a)
ROOT="/root"

if [[ ${UNAME} == *"Darwin"* ]]; then
    if [[ -f /usr/local/Homebrew/bin/brew ]] || [[ -f /opt/Homebrew/bin/brew ]]; then
        TARGET="darwin"
        ROOT="$HOME"
        echo "macOS detected."
        if [[ ! -f /usr/local/go/bin/go ]]; then
	    if [[ -f /usr/local/bin/go ]]; then
		mkdir -p /usr/local/go/bin
		ln -s /usr/local/bin/go /usr/local/go/bin/go
	    else
            	echo "Go was not found. You must download it from https://go.dev/dl/ for your macOS."
            	exit 1
	    fi
        fi
    else
        echo "macOS detected, but 'brew' was not found. Install it with the following command and try running setup.sh again:"
        echo '/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"'
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
    echo "This OS is not supported. This script currently supports Linux with either apt, pacman, or dnf."
    if [[ ! "$1" == *"--bypass-target-check"* ]]; then
        echo "If you would like to get the required packages yourself, you may bypass this by running setup.sh with the --bypass-target-check flag"
        echo "The following packages are required (debian apt in this case): wget openssl net-tools libsox-dev libopus-dev make iproute2 xz-utils libopusfile-dev pkg-config gcc curl g++ unzip avahi-daemon git"
        exit 1
    fi
fi

if [[ "${UNAME}" == *"x86_64"* ]]; then
    ARCH="x86_64"
    echo "amd64 architecture confirmed."
elif [[ "${UNAME}" == *"aarch64"* ]] || [[ "${UNAME}" == *"arm64"* ]]; then
    ARCH="aarch64"
    echo "aarch64 architecture confirmed."
elif [[ "${UNAME}" == *"armv7l"* ]]; then
    ARCH="armv7l"
    echo "armv7l (32-bit) WARN: The Coqui and VOSK bindings are broken for this platform at the moment, so please choose Picovoice when the script asks. wire-pod is designed for 64-bit systems."
    STT=""
else
    echo "Your CPU architecture not supported. This script currently supports x86_64, aarch64, and armv7l."
    exit 1
fi

if [[ $EUID -ne 0 ]]; then
    echo "This script must be run as root. sudo ./setup.sh"
    exit 1
fi

if [[ ! -d ./chipper ]]; then
    echo "Script is not running in the wire-pod/ directory or chipper folder is missing. Exiting."
    exit 1
fi

if [[ $1 != "-f" ]]; then
    if [[ ${ARCH} == "x86_64" ]] && [[ ${TARGET} != "darwin" ]]; then
        CPUINFO=$(cat /proc/cpuinfo)
        if [[ "${CPUINFO}" == *"avx"* ]]; then
            echo "AVX support confirmed."
        else
            echo "This CPU does not support AVX. Text to speech performance will not be optimal."
            AVXSUPPORT="noavx"
            #echo "If you would like to bypass this, run the script like this: './setup.sh -f'"
            #exit 1
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
        sudo pacman -S --noconfirm wget openssl net-tools sox opus make iproute2 opusfile curl unzip avahi git libsodium go pkg-config
    elif [[ ${TARGET} == "fedora" ]]; then
        dnf update
        dnf install -y wget openssl net-tools sox opus make opusfile curl unzip avahi git libsodium-devel
    elif [[ ${TARGET} == "darwin" ]]; then
        sudo -u $SUDO_USER brew update
        sudo -u $SUDO_USER brew install wget pkg-config opus opusfile
    fi
    touch ./vector-cloud/packagesGotten
    echo
    echo "Installing golang binary package"
    mkdir golang
    cd golang
    if [[ ${TARGET} != "darwin" ]] && [[ ${TARGET} != "arch" ]]; then
        if [[ ! -f /usr/local/go/bin/go ]]; then
            if [[ ${ARCH} == "x86_64" ]]; then
                wget -q --show-progress --no-check-certificate https://go.dev/dl/go1.19.4.linux-amd64.tar.gz
                rm -rf /usr/local/go && tar -C /usr/local -xzf go1.19.4.linux-amd64.tar.gz
            elif [[ ${ARCH} == "aarch64" ]]; then
                wget -q --show-progress --no-check-certificate https://go.dev/dl/go1.19.4.linux-arm64.tar.gz
                rm -rf /usr/local/go && tar -C /usr/local -xzf go1.19.4.linux-arm64.tar.gz
            elif [[ ${ARCH} == "armv7l" ]]; then
                wget -q --show-progress --no-check-certificate https://go.dev/dl/go1.19.4.linux-armv6l.tar.gz
                rm -rf /usr/local/go && tar -C /usr/local -xzf go1.19.4.linux-armv6l.tar.gz
            fi
            ln -s /usr/local/go/bin/go /usr/bin/go
        fi
    else
        echo "This is a macOS or arch target, assuming Go is installed already"
	if [[ ${TARGET} == "arch" ]] && [[ ! -d /usr/local/go/bin ]]; then
		mkdir -p /usr/local/go/bin
		ln -s /usr/bin/go /usr/local/go/bin/go
	fi
    fi
    cd ..
    rm -rf golang
    echo
}

function getSTT() {
    echo "export DEBUG_LOGGING=true" > ./chipper/source.sh
    rm -f ./chipper/pico.key
    function sttServicePrompt() {
        echo
        echo "Which speech-to-text service would you like to use?"
        echo "1: Coqui (local, no usage collection, less accurate, a little slower)"
        echo "2: Picovoice Leopard (local, usage collected, accurate, account signup required)"
        echo "3: VOSK (local, accurate, multilanguage, fast, recommended)"
        # echo "4: Whisper (local, accurate, multilanguage, a little slower, recommended for more powerful hardware)"
        echo
        read -p "Enter a number (3): " sttServiceNum
        if [[ ! -n ${sttServiceNum} ]]; then
            sttService="vosk"
        elif [[ ${sttServiceNum} == "1" ]]; then
            if [[ ${TARGET} == "darwin" ]]; then
                echo "Coqui is not supported for macOS. Please select another option."
                sttServicePrompt
            else
            sttService="coqui"
            fi
        elif [[ ${sttServiceNum} == "2" ]]; then
            sttService="leopard"
        elif [[ ${sttServiceNum} == "3" ]]; then
            sttService="vosk"
        elif [[ ${sttServiceNum} == "4" ]]; then
            sttService="whisper"
        else
            echo
            echo "Choose a valid number, or just press enter to use the default number."
            sttServicePrompt
        fi
    }
    if [[ "$STT" == "vosk" ]]; then
        echo "Vosk config"
        sttService="vosk"
    else
        sttServicePrompt
    fi
    if [[ ${sttService} == "leopard" ]]; then
        function picoApiPrompt() {
            echo
            echo "Create an account at https://console.picovoice.ai/ and enter the Access Key it gives you."
            echo
            read -p "Enter your Access Key: " picoKey
            if [[ ! -n ${picoKey} ]]; then
                echo
                echo "You must enter a key."
                picoApiPrompt
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
            echo "Getting VOSK assets"
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
            /usr/local/go/bin/go get -u github.com/kercre123/vosk-api/go/...
            /usr/local/go/bin/go get github.com/kercre123/vosk-api
            /usr/local/go/bin/go install github.com/kercre123/vosk-api/go
            cd ${origDir}
        fi
    elif [[ ${sttService} == "whisper" ]]; then
        echo "export STT_SERVICE=whisper.cpp" >> ./chipper/source.sh
        origDir="$(pwd)"
        echo "Getting Whisper assets"
        if [[ ! -d ./whisper.cpp ]]; then
            mkdir whisper.cpp
            cd whisper.cpp
            git clone https://github.com/ggerganov/whisper.cpp.git .
            ./models/download-ggml-model.sh tiny
            cd bindings/go
            make whisper
            cd ${origDir}
        fi
    else
    echo "export STT_SERVICE=coqui" >> ./chipper/source.sh
        if [[ ! -f ./stt/completed ]]; then
            echo "Getting STT assets"
            if [[ -d /root/.coqui ]]; then
                rm -rf /root/.coqui
            fi
            origDir=$(pwd)
            mkdir /root/.coqui
            cd /root/.coqui
            if [[ ${ARCH} == "x86_64" ]]; then
                if [[ ${AVXSUPPORT} == "noavx" ]]; then
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
            fi
            cd ${origDir}/chipper
            export CGO_LDFLAGS="-L/root/.coqui/"
            export CGO_CXXFLAGS="-I/root/.coqui/"
            export LD_LIBRARY_PATH="/root/.coqui/:$LD_LIBRARY_PATH"
            /usr/local/go/bin/go get -u github.com/asticode/go-asticoqui/...
            /usr/local/go/bin/go get github.com/asticode/go-asticoqui
            /usr/local/go/bin/go install github.com/asticode/go-asticoqui
            cd ${origDir}
            mkdir -p stt
            cd stt
            function sttModelPrompt() {
                echo
                echo "Which voice model would you like to use?"
                echo "1: large_vocabulary (faster, less accurate, ~100MB)"
                echo "2: huge_vocabulary (slower, more accurate, handles faster speech better, ~900MB)"
                echo
                read -p "Enter a number (1): " sttModelNum
                if [[ ! -n ${sttModelNum} ]]; then
                    sttModel="large_vocabulary"
                elif [[ ${sttModelNum} == "1" ]]; then
                    sttModel="large_vocabulary"
                elif [[ ${sttModelNum} == "2" ]]; then
                    sttModel="huge_vocabulary"
                else
                    echo
                    echo "Choose a valid number, or just press enter to use the default number."
                    sttModelPrompt
                fi
            }
            sttModelPrompt
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
                echo "Invalid model specified"
                exit 0
            fi
            echo
            touch completed
            echo "STT assets successfully downloaded!"
            cd ..
        else
            echo "STT assets already there! If you want to redownload, use the 4th option in setup.sh."
        fi
    fi
}

function IPDNSPrompt() {
    read -p "Enter a number (3): " yn
    case $yn in
        "1") SANPrefix="IP" ;;
        "2") SANPrefix="DNS" ;;
        "3") isEscapePod="epod" ;;
        "4") noCerts="true" ;;
        "") isEscapePod="epod" ;;
        *)
            echo "Please answer with 1, 2, 3, or 4."
            IPDNSPrompt
            ;;
    esac
}

function IPPrompt() {
    if [[ ${TARGET} == "darwin" ]]; then
        IPADDRESS=$(ifconfig | grep "inet " | grep -v 127.0.0.1 | cut -d\  -f2)
    else
        IPADDRESS=$(ip -4 addr | grep $(ip addr | awk '/state UP/ {print $2}' | sed 's/://g') | grep -oP '(?<=inet\s)\d+(\.\d+){3}')
    fi
    read -p "Enter the IP address of the machine you are running this script on (${IPADDRESS}): " ipaddress
    if [[ ! -n ${ipaddress} ]]; then
        address=${IPADDRESS}
    else
        address=${ipaddress}
    fi
}

function DNSPrompt() {
    read -p "Enter the domain you would like to use: " dnsurl
    if [[ ! -n ${dnsurl} ]]; then
        echo "You must enter a domain."
        DNSPrompt
    fi
    address=${dnsurl}
}

function generateCerts() {
    echo
    echo "Creating certificates"
    echo
    echo "Would you like to use your IP address or a domain for the Subject Alt Name?"
    echo "Or would you like to use the escapepod.local certs?"
    echo
    echo "1: IP address (recommended for OSKR Vectors)"
    echo "2: Domain"
    echo "3: escapepod.local (required for regular production Vectors)"
    if [[ -d ./certs ]]; then
        echo "4: Keep certificates as is"
    fi
    IPDNSPrompt
    if [[ ${noCerts} != "true" ]]; then
        if [[ ${isEscapePod} != "epod" ]]; then
            if [[ ${SANPrefix} == "IP" ]]; then
                IPPrompt
            else
                DNSPrompt
            fi
            rm -f ./chipper/useepod
            rm -rf ./certs
            mkdir certs
            cd certs
            echo ${address} >address
            echo "Creating san config"
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
            echo "Generating key and cert"
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
    if [[ ! -n ${botAddress} ]]; then
        echo "To copy vic-cloud and server_config.json to your OSKR robot, run this script like this:"
        echo "Usage: sudo ./setup.sh scp <vector's ip> <path/to/ssh-key>"
        echo "Example: sudo ./setup.sh scp 192.168.1.150 /home/wire/id_rsa_Vector-R2D2"
        echo
        echo "If your Vector is on Wire's custom software or you have an old dev build, you can run this command without an SSH key:"
        echo "Example: sudo ./setup.sh scp 192.168.1.150"
        echo
        exit 0
    fi
    if [[ ! -f ./certs/server_config.json ]]; then
        echo "server_config.json file missing. You need to generate this file with ./setup.sh's 6th option."
        exit 0
    fi
    if [[ ! -n ${keyPath} ]]; then
        echo
        if [[ ! -f ./ssh_root_key ]]; then
            echo "Key not provided, downloading ssh_root_key..."
            wget http://wire.my.to:81/ssh_root_key
        else
            echo "Key not provided, using ./ssh_root_key (already there)..."
        fi
        chmod 600 ./ssh_root_key
        keyPath="./ssh_root_key"
    fi
    if [[ ! -f ${keyPath} ]]; then
        echo "The key that was provided was not found. Exiting."
        exit 0
    fi
    ssh -i ${keyPath} root@${botAddress} "cat /build.prop" >/tmp/sshTest 2>>/tmp/sshTest
    botBuildProp=$(cat /tmp/sshTest)
    if [[ "${botBuildProp}" == *"no mutual signature"* ]]; then
        echo
        echo "An entry must be made to the ssh config for this to work. Would you like the script to do this?"
        echo "1: Yes"
        echo "2: No (exit)"
        echo
        function rsaAddPrompt() {
            read -p "Enter a number (1): " yn
            case $yn in
                "1") echo ;;
                "2") exit 0 ;;
                "") echo ;;
                *)
                    echo "Please answer with 1 or 2."
                    rsaAddPrompt
                    ;;
            esac
        }
        rsaAddPrompt
        echo "PubkeyAcceptedKeyTypes +ssh-rsa" >>/etc/ssh/ssh_config
        botBuildProp=$(ssh -i ${keyPath} root@${botAddress} "cat /build.prop")
    fi
    if [[ ! "${botBuildProp}" == *"ro.build"* ]]; then
        echo "Unable to communicate with robot. The key may be invalid, the bot may not be unlocked, or this device and the robot are not on the same network."
        exit 0
    fi
    scp -v -i ${keyPath} root@${botAddress}:/build.prop /tmp/scpTest >/tmp/scpTest 2>>/tmp/scpTest
    scpTest=$(cat /tmp/scpTest)
    if [[ "${scpTest}" == *"sftp"* ]]; then
        oldVar="-O"
    else
        oldVar=""
    fi
    if [[ ! "${botBuildProp}" == *"ro.build"* ]]; then
        echo "Unable to communicate with robot. The key may be invalid, the bot may not be unlocked, or this device and the robot are not on the same network."
        exit 0
    fi
    ssh  -oStrictHostKeyChecking=no -i ${keyPath} root@${botAddress} "mount -o rw,remount / && mount -o rw,remount,exec /data && systemctl stop anki-robot.target && mv /anki/data/assets/cozmo_resources/config/server_config.json /anki/data/assets/cozmo_resources/config/server_config.json.bak"
    scp  -oStrictHostKeyChecking=no ${oldVar} -i ${keyPath} ./vector-cloud/build/vic-cloud root@${botAddress}:/anki/bin/
    scp  -oStrictHostKeyChecking=no ${oldVar} -i ${keyPath} ./certs/server_config.json root@${botAddress}:/anki/data/assets/cozmo_resources/config/
    scp  -oStrictHostKeyChecking=no ${oldVar} -i ${keyPath} ./vector-cloud/pod-bot-install.sh root@${botAddress}:/data/
    if [[ -f ./chipper/useepod ]]; then
        scp -oStrictHostKeyChecking=no ${oldVar} -i ${keyPath} ./chipper/epod/ep.crt root@${botAddress}:/anki/etc/wirepod-cert.crt
        scp -oStrictHostKeyChecking=no ${oldVar} -i ${keyPath} ./chipper/epod/ep.crt root@${botAddress}:/data/data/wirepod-cert.crt
    else
        scp -oStrictHostKeyChecking=no ${oldVar} -i ${keyPath} ./certs/cert.crt root@${botAddress}:/anki/etc/wirepod-cert.crt
        scp -oStrictHostKeyChecking=no ${oldVar} -i ${keyPath} ./certs/cert.crt root@${botAddress}:/data/data/wirepod-cert.crt
    fi
    ssh -oStrictHostKeyChecking=no -i ${keyPath} root@${botAddress} "chmod +rwx /anki/data/assets/cozmo_resources/config/server_config.json /anki/bin/vic-cloud /data/data/wirepod-cert.crt /anki/etc/wirepod-cert.crt /data/pod-bot-install.sh && /data/pod-bot-install.sh"
    rm -f /tmp/sshTest
    rm -f /tmp/scpTest
    echo "Vector has been reset to Onboarding mode, but no user data has actually been erased."
    echo
    echo "Everything has been copied to the bot! Use https://keriganc.com/vector-epod-setup on any device with Bluetooth to finish setting up your Vector!"
    echo
    echo "Everything is now setup! You should be ready to run chipper. sudo ./chipper/start.sh"
    echo
}

function setupSystemd() {
    if [[ ${TARGET} == "darwin" ]]; then
        echo "This cannot be done on macOS."
        exit 1
    fi
    if [[ ! -f ./chipper/source.sh ]]; then
        echo "You need to make a source.sh file. This can be done with the setup.sh script, option 6."
        exit 1
    fi
    source ./chipper/source.sh
    echo "[Unit]" >wire-pod.service
    echo "Description=Wire Escape Pod (coqui)" >>wire-pod.service
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
    if [[ ${STT_SERVICE} == "leopard" ]]; then
        echo "wire-pod.service created, building chipper with Picovoice STT service..."
        /usr/local/go/bin/go build cmd/leopard/main.go
    elif [[ ${STT_SERVICE} == "vosk" ]]; then
        echo "wire-pod.service created, building chipper with VOSK STT service..."
        export CGO_ENABLED=1
        export CGO_CFLAGS="-I$HOME/.vosk/libvosk"
        export CGO_LDFLAGS="-L $HOME/.vosk/libvosk -lvosk -ldl -lpthread"
        export LD_LIBRARY_PATH="$HOME/.vosk/libvosk:$LD_LIBRARY_PATH"
        /usr/local/go/bin/go build cmd/vosk/main.go
    else
        echo "wire-pod.service created, building chipper with Coqui STT service..."
        export CGO_LDFLAGS="-L$HOME/.coqui/"
        export CGO_CXXFLAGS="-I$HOME/.coqui/"
        export LD_LIBRARY_PATH="$HOME/.coqui/:$LD_LIBRARY_PATH"
        /usr/local/go/bin/go build cmd/coqui/main.go
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
    echo "systemd service has been installed and enabled! The service is called wire-pod.service"
    echo
    echo "To start the service, run: 'systemctl start wire-pod'"
    echo "Then, to see logs, run 'journalctl -fe | grep start.sh'"
}

function disableSystemd() {
    if [[ ${TARGET} == "darwin" ]]; then
        echo "This cannot be done on macOS."
        exit 1
    fi
    echo
    echo "Disabling wire-pod.service"
    systemctl stop wire-pod.service
    systemctl disable wire-pod.service
    rm ./chipper/chipper
    rm -f /lib/systemd/system/wire-pod.service
    systemctl daemon-reload
    echo
    echo "wire-pod.service has been removed and disabled."
}

function defaultLaunch() {
    echo
    getPackages
    getSTT
    echo
    echo "wire-pod is ready to run! You are ready to move to the next step and run sudo ./chipper/start.sh"
}

if [[ $1 == "scp" ]]; then
    botAddress=$2
    keyPath=$3
    scpToBot
    exit 0
fi

if [[ $1 == "daemon-enable" ]]; then
    setupSystemd
    exit 0
fi

if [[ $1 == "daemon-disable" ]]; then
    disableSystemd
    exit 0
fi

if [[ $1 == "-f" ]] && [[ $2 == "scp" ]]; then
    botAddress=$3
    keyPath=$4
    scpToBot
    exit 0
fi

# echo "What would you like to do?"
# echo "1: Full Setup (recommended) (builds chipper, gets STT stuff, generates certs, creates source.sh file, and creates server_config.json for your bot"
# echo "2: Just build vic-cloud"
# echo "3: Just build chipper"
# echo "4: Just get STT assets"
# echo "5: Just generate certs"
# echo "6: Create wire-pod config file (change/add API keys)"
# echo "(NOTE: You can just press enter without entering a number to select the default, recommended option)"
# echo
defaultLaunch
