#!/bin/bash

echo

UNAME=$(uname -a)
CPUINFO=$(cat /proc/cpuinfo)

if [[ -f /usr/bin/apt ]]; then
    TARGET="debian"
    echo "Debian-based Linux confirmed."
elif [[ -f /usr/bin/pacman ]]; then
    TARGET="arch"
    echo "Arch Linux confirmed."
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
elif [[ "${UNAME}" == *"aarch64"* ]]; then
    ARCH="aarch64"
    echo "aarch64 architecture confirmed."
elif [[ "${UNAME}" == *"armv7l"* ]]; then
    ARCH="armv7l"
    echo "armv7l WARN: The Coqui and VOSK bindings are broken for this platform at the moment, so please choose Picovoice when the script asks."
    exit 1
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
    if [[ ${ARCH} == "x86_64" ]]; then
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
    echo "Installing required packages (ffmpeg, golang, wget, openssl, net-tools, iproute2, sox, opus)"
    if [[ ${TARGET} == "debian" ]]; then
        apt update -y
        apt install -y wget openssl net-tools libsox-dev libopus-dev make iproute2 xz-utils libopusfile-dev pkg-config gcc curl g++ unzip avahi-daemon git libasound2-dev
    elif [[ ${TARGET} == "arch" ]]; then
        pacman -Sy --noconfirm
        sudo pacman -S --noconfirm wget openssl net-tools sox opus make iproute2 opusfile curl unzip avahi git
    elif [[ ${TARGET} == "fedora" ]]; then
        dnf update
        dnf install -y wget openssl net-tools sox opus make opusfile curl unzip avahi git
    fi
    touch ./vector-cloud/packagesGotten
    echo
    echo "Installing golang binary package"
    mkdir golang
    cd golang
    if [[ ! -f /usr/local/go/bin/go ]]; then
        if [[ ${ARCH} == "x86_64" ]]; then
            wget -q --show-progress --no-check-certificate https://go.dev/dl/go1.18.2.linux-amd64.tar.gz
            rm -rf /usr/local/go && tar -C /usr/local -xzf go1.18.2.linux-amd64.tar.gz
            export PATH=$PATH:/usr/local/go/bin
        elif [[ ${ARCH} == "aarch64" ]]; then
            wget -q --show-progress --no-check-certificate https://go.dev/dl/go1.18.2.linux-arm64.tar.gz
            rm -rf /usr/local/go && tar -C /usr/local -xzf go1.18.2.linux-arm64.tar.gz
            export PATH=$PATH:/usr/local/go/bin
        elif [[ ${ARCH} == "armv7l" ]]; then
            wget -q --show-progress --no-check-certificate https://go.dev/dl/go1.18.2.linux-armv6l.tar.gz
            rm -rf /usr/local/go && tar -C /usr/local -xzf go1.18.2.linux-armv6l.tar.gz
            export PATH=$PATH:/usr/local/go/bin
        fi
    fi
    cd ..
    rm -rf golang
    echo
}

function buildCloud() {
    echo
    echo "Installing docker"
    if [[ ${TARGET} == "debian" ]]; then
        apt update -y
        apt install -y docker.io
    elif [[ ${TARGET} == "arch" ]]; then
        pacman -Sy --noconfirm
        sudo pacman -S --noconfirm docker
    fi
    systemctl start docker
    echo
    cd vector-cloud
    ./build.sh
    cd ..
    echo
    echo "./vector-cloud/build/vic-cloud built!"
    echo
}

function buildChipper() {
    echo
    cd chipper
    echo
    cd ..
}

function getLanguage() {
    if [[ ${sttService} == "vosk" ]]; then
        origDir=$(pwd)
        echo
        echo "Which STT language would you like to use?"
        echo "1: English (US)"
        echo "2: Italian (IT)"
        echo "3: Spanish (ES)"
        echo "4: French (FR)"
        echo "5: German (DE)"
        echo
        read -p "Enter a number (1): " languageNum
        if [[ ! -n ${languageNum} ]]; then
            languageNum="en-US"
            if [[ ! -d vosk/models/en-US ]]; then
                cd ${origDir}
                echo "Downloading English (US) model"
                mkdir -p vosk/models/en-US
                cd vosk/models/en-US
                wget -q --show-progress --no-check-certificate https://alphacephei.com/vosk/models/vosk-model-small-en-us-0.15.zip
                unzip vosk-model-small-en-us-0.15.zip
                mv vosk-model-small-en-us-0.15 model
                rm vosk-model-small-en-us-0.15.zip
            fi
        elif [[ ${languageNum} == "1" ]]; then
            languageNum="en-US"
            if [[ ! -d vosk/models/en-US ]]; then
                cd ${origDir}
                echo "Downloading English (US) model"
                mkdir -p vosk/models/en-US
                cd vosk/models/en-US
                wget -q --show-progress --no-check-certificate https://alphacephei.com/vosk/models/vosk-model-small-en-us-0.15.zip
                unzip vosk-model-small-en-us-0.15.zip
                mv vosk-model-small-en-us-0.15 model
                rm vosk-model-small-en-us-0.15.zip
            fi
        elif [[ ${languageNum} == "2" ]]; then
            languageNum="it-IT"
            if [[ ! -d vosk/models/it-IT ]]; then
                cd ${origDir}
                echo "Downloading Italian (IT) model"
                mkdir -p vosk/models/it-IT
                cd vosk/models/it-IT
                wget -q --show-progress --no-check-certificate https://alphacephei.com/vosk/models/vosk-model-small-it-0.22.zip
                unzip vosk-model-small-it-0.22.zip
                mv vosk-model-small-it-0.22 model
                rm vosk-model-small-it-0.22.zip
            fi
        elif [[ ${languageNum} == "3" ]]; then
            languageNum="es-ES"
            if [[ ! -d vosk/models/es-ES ]]; then
                cd ${origDir}
                echo "Downloading Spanish (ES) model"
                mkdir -p vosk/models/es-ES
                cd vosk/models/es-ES
                wget -q --show-progress --no-check-certificate https://alphacephei.com/vosk/models/vosk-model-small-es-0.42.zip
                unzip vosk-model-small-es-0.42.zip
                mv vosk-model-small-es-0.42 model
                rm vosk-model-small-es-0.42.zip
            fi
        elif [[ ${languageNum} == "4" ]]; then
            languageNum="fr-FR"
            if [[ ! -d vosk/models/fr-FR ]]; then
                cd ${origDir}
                echo "Downloading French (FR) model"
                mkdir -p vosk/models/fr-FR
                cd vosk/models/fr-FR
                wget -q --show-progress --no-check-certificate https://alphacephei.com/vosk/models/vosk-model-small-fr-0.22.zip
                unzip vosk-model-small-fr-0.22.zip
                mv vosk-model-small-fr-0.22 model
                rm vosk-model-small-fr-0.22.zip
            fi
        elif [[ ${languageNum} == "5" ]]; then
            languageNum="de-DE"
            if [[ ! -d vosk/models/de-DE ]]; then
                cd ${origDir}
                echo "Downloading German (DE) model"
                mkdir -p vosk/models/de-DE
                cd vosk/models/de-DE
                wget -q --show-progress --no-check-certificate https://alphacephei.com/vosk/models/vosk-model-small-de-0.15.zip
                unzip vosk-model-small-de-0.15.zip
                mv vosk-model-small-de-0.15 model
                rm vosk-model-small-de-0.15.zip
            fi
        else
            echo
            echo "Choose a valid number, or just press enter to use the default number."
            getLanguage
        fi
        cd ${origDir}
    fi
}

function getSTT() {
    rm -f ./chipper/pico.key
    function sttServicePrompt() {
        echo
        echo "Which speech-to-text service would you like to use?"
        echo "1: Coqui (local, no usage collection, less accurate, a little slower)"
        echo "2: Picovoice Leopard (local, usage collected, accurate, account signup required)"
        echo "3: VOSK (local, accurate, multilanguage, fast, recommended)"
        echo
        read -p "Enter a number (3): " sttServiceNum
        if [[ ! -n ${sttServiceNum} ]]; then
            sttService="vosk"
        elif [[ ${sttServiceNum} == "1" ]]; then
            sttService="coqui"
        elif [[ ${sttServiceNum} == "2" ]]; then
            sttService="leopard"
        elif [[ ${sttServiceNum} == "3" ]]; then
            sttService="vosk"
        else
            echo
            echo "Choose a valid number, or just press enter to use the default number."
            sttServicePrompt
        fi
    }
    sttServicePrompt
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
        echo ${picoKey} > ./chipper/pico.key
    elif [[ ${sttService} == "vosk" ]]; then
        origDir=$(pwd)
        if [[ ! -f ./vosk/completed ]]; then
            echo "Getting VOSK assets"
            rm -fr /root/.vosk
            mkdir /root/.vosk
            cd /root/.vosk
            if [[ ${ARCH} == "x86_64" ]]; then
                VOSK_DIR="vosk-linux-x86_64-0.3.43"
            elif [[ ${ARCH} == "aarch64" ]]; then
                VOSK_DIR="vosk-linux-aarch64-0.3.43"
            elif [[ ${ARCH} == "armv7l" ]]; then
                VOSK_DIR="vosk-linux-armv7l-0.3.43"
            fi
            VOSK_ARCHIVE="$VOSK_DIR.zip"
            wget -q --show-progress --no-check-certificate "https://github.com/alphacep/vosk-api/releases/download/v0.3.43/$VOSK_ARCHIVE"
            unzip "$VOSK_ARCHIVE"
            mv "$VOSK_DIR" libvosk
            rm -fr "$VOSK_ARCHIVE"

            cd ${origDir}/chipper
            export CGO_ENABLED=1
            export CGO_CFLAGS="-I/root/.vosk/libvosk"
            export CGO_LDFLAGS="-L /root/.vosk/libvosk -lvosk -ldl -lpthread"
            export LD_LIBRARY_PATH="$HOME/.vosk/libvosk:$LD_LIBRARY_PATH"
            /usr/local/go/bin/go get -u github.com/alphacep/vosk-api/go/...
            /usr/local/go/bin/go get github.com/alphacep/vosk-api
            /usr/local/go/bin/go install github.com/alphacep/vosk-api/go
            cd ${origDir}
        fi
    else
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
            export CGO_LDFLAGS="-L$HOME/.coqui/"
            export CGO_CXXFLAGS="-I$HOME/.coqui/"
            export LD_LIBRARY_PATH="$HOME/.coqui/:$LD_LIBRARY_PATH"
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
    IPADDRESS=$(ip -4 addr | grep $(ip addr | awk '/state UP/ {print $2}' | sed 's/://g') | grep -oP '(?<=inet\s)\d+(\.\d+){3}')
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

function makeSource() {
    if [[ -f ./chipper/source.sh ]]; then
        echo "Found an existing source.sh, exporting"
        cd chipper
        source source.sh
        cd ..
        SOURCEEXPORTED="true"
    fi
    if [[ ! -f ./certs/address ]] && [[ ! -f ./chipper/useepod ]]; then
        echo "You need to generate certs first!"
        exit 0
    fi
    cd chipper
    if [[ ! -f ./useepod ]]; then
        read -p "What port would you like to use? (443): " portPrompt
        if [[ -n ${portPrompt} ]]; then
            port=${portPrompt}
        else
            port="443"
        fi
        if netstat -pln | grep :${port}; then
            echo
            netstat -pln | grep :${port}
            echo
            echo "Something may be using port ${port}. Make sure that port is free before you start chipper."
        fi
    else
        echo
        echo "Using port 443 for chipper because escapepod.local is being used."
        port="443"
        echo
    fi
    read -p "What port would you like to use for the HTTP webserver used for custom intents, bot configuration, etc? (8080): " webportPrompt
    if [[ -n ${webportPrompt} ]]; then
        webport=${webportPrompt}
    else
        webport="8080"
    fi
    echo
    function weatherPrompt() {
        echo "Would you like to setup weather commands? This involves creating a free account at one of the weather providers' websites and putting in your API key."
        echo "Otherwise, placeholder values will be used."
        echo
        echo "1: Yes, and I want to use weatherapi.com"
        echo "2: Yes, and I want to use openweathermap.org (with forecast support)"
        echo "3: No"
        if [[ ${SOURCEEXPORTED} == "true" ]]; then
            echo "4: Do not change weather configuration"
        fi
        read -p "Enter a number (3): " yn
        case $yn in
            "1") weatherSetup="true" weatherProvider="weatherapi.com" ;;
            "2") weatherSetup="true" weatherProvider="openweathermap.org" ;;
            "3") weatherSetup="false" ;;
            "4") weatherSetup="true" noChangeWeather="true" ;;
            "") weatherSetup="false" ;;
            *)
                echo "Please answer with 1 or 2."
                weatherPrompt
                ;;
        esac
    }
    weatherPrompt
    if [[ ${weatherSetup} == "true" ]]; then
        if [[ ! ${noChangeWeather} == "true" ]]; then
            function weatherKeyPrompt() {
                echo
                echo "Create an account at https://$weatherProvider and enter the API key it gives you."
                echo "If you have changed your mind, enter Q to continue without weather commands."
                echo
                read -p "Enter your API key: " weatherAPI
                if [[ ! -n ${weatherAPI} ]]; then
                    echo "You must enter an API key. If you have changed your mind, you may also enter Q to continue without weather commands."
                    weatherKeyPrompt
                fi
                if [[ ${weatherAPI} == "Q" ]]; then
                    weatherSetup="false"
                fi
            }
            weatherKeyPrompt
            function weatherUnitPrompt() {
                echo "What temperature unit would you like to use?"
                echo
                echo "1: Fahrenheit"
                echo "2: Celsius"
                read -p "Enter a number (1): " yn
                case $yn in
                    "1") weatherUnit="F" ;;
                    "2") weatherUnit="C" ;;
                    "") weatherUnit="F" ;;
                    *)
                        echo "Please answer with 1 or 2."
                        weatherUnitPrompt
                        ;;
                esac
            }
            weatherUnitPrompt
        else
            if [[ $WEATHERAPI_ENABLED == "true" ]]; then
                weatherUnit="$WEATHERAPI_UNIT"
                weatherProvider="$WEATHERAPI_PROVIDER"
                weatherAPI="$WEATHERAPI_KEY"
            else
                weatherSetup="false"
            fi
        fi
    fi
    function knowledgePrompt() {
        echo
        echo "Would you like to setup knowledge graph (I have a question) commands?"
        echo "Houndify: Same service the official server uses. You must create a free account at https://www.houndify.com/signup and putting in your Client Key and Client ID."
        echo "OpenAI: May not provide accurate results, but may be faster and more interesting. You must create a free account at https://beta.openai.com/signup and enter an API key. You only have a trial."
        echo "This is not required, and if you choose 3 then placeholder values will be used. And if you change your mind later, just run ./setup.sh again with the 6th option."
        echo
        echo "1: Yes, with Houndify"
        echo "2: Yes, with OpenAI"
        echo "3: No"
        if [[ ${SOURCEEXPORTED} == "true" ]]; then
            echo "4: Do not change knowledgegraph configuration"
        fi
        read -p "Enter a number (3): " yn
        case $yn in
            "1") knowledgeSetup="true" knowledgeProvider="houndify" ;;
            "2") knowledgeSetup="true" knowledgeProvider="openai" ;;
            "3") knowledgeSetup="false" ;;
            "4") knowledgeSetup="true" noChangeKnowledge="true" ;;
            "") knowledgeSetup="false" ;;
            *)
                echo "Please answer with 1 or 2."
                knowledgePrompt
                ;;
        esac
    }
    knowledgePrompt
    if [[ ${knowledgeSetup} == "true" ]]; then
        if [[ ! ${noChangeKnowledge} == "true" ]]; then
            function houndifyIDPrompt() {
				knowledgeIntent="false"
                echo
                echo "Create an account at https://www.houndify.com/signup and enter the Client ID (not Key) it gives you."
                echo "If you have changed your mind, enter Q to continue without knowledge graph commands."
                echo
                read -p "Enter your Client ID: " knowledgeID
                if [[ ! -n ${knowledgeID} ]]; then
                    echo "You must enter a Houndify Client ID. If you have changed your mind, you may also enter Q to continue without knowledgegraph commands."
                    houndifyIDPrompt
                fi
                if [[ ${knowledgeID} == "Q" ]]; then
                    knowledgeSetup="false"
                fi
            }
            function openAIPrompt() {
                echo
                echo "Create an account at https://beta.openai.com/signup, generate an API key, and enter it here."
                echo "If you have changed your mind, enter Q to continue without knowledge graph commands."
                if [[ $SOURCEEXPORTED == "true" ]] && ! [[ $KNOWLEDGE_KEY == "" ]]; then
                    echo "Press enter to use the key you entered the last time you set this up."
                fi
                echo
                read -p "Enter your API key: " knowledgeKey
                if [[ ! -n ${knowledgeKey} ]]; then
                    if [[ $KNOWLEDGE_KEY == "" ]]; then
                        echo "You must enter an OpenAI API key. If you have changed your mind, you may also enter Q to continue without knowledgegraph commands."
                        openAIPrompt
                    else
                        knowledgeKey="$KNOWLEDGE_KEY"
                    fi
                fi
                if [[ ${knowledgeKey} == "Q" ]]; then
                    knowledgeSetup="false"
                fi
                echo
                echo "Would you like to use the intent graph feature? This is a feature which was introduced by DDL in firmware 1.8. If the speech does not match an intent, it will feed the information to OpenAI and the bot will respond with that response instead."
                echo
                echo "1: Yes"
                echo "2: No"
                echo
                read -p "Enter a number (1): " yn
                case $yn in
                    "1") knowledgeIntent="true" ;;
                    "2") knowledgeIntent="false" ;;
                    "") knowledgeIntent="true" ;;
                    *) knowledgeIntent="true" ;;
                esac
				echo
            }
            function houndifyKeyPrompt() {
                echo
                echo "Now enter the Houndify Client Key (not ID)."
                echo
                read -p "Enter your Client Key: " knowledgeKey
                if [[ ! -n ${knowledgeKey} ]]; then
                    echo "You must enter a Houndify Client Key."
                    houndifyKeyPrompt
                fi
                if [[ ${knowledgeKey} == "Q" ]]; then
                    knowledgeSetup="false"
                fi
            }
            if [[ ${knowledgeProvider} == "houndify" ]]; then
                houndifyIDPrompt
            else
                openAIPrompt
            fi
            if [[ ${knowledgeSetup} == "true" ]] && [[ ${knowledgeProvider} == "houndify" ]]; then
                houndifyKeyPrompt
            fi
        else
            if [[ "$KNOWLEDGE_ENABLED" == "true" ]]; then
                knowledgeKey="$KNOWLEDGE_KEY"
                knowledgeID="$KNOWLEDGE_ID"
                knowledgeProvider="$KNOWLEDGE_PROVIDER"
            else
                if [[ "$HOUNDIFY_ENABLED" == "true" ]]; then
                    knowledgeKey="$HOUNDIFY_CLIENT_KEY"
                    knowledgeID="$HOUNDIFY_CLIENT_ID"
                    knowledgeProvider="houndify"
                else
                    knowledgeSetup="false"
                fi
            fi
        fi
    fi
    echo "export DDL_RPC_PORT=${port}" >source.sh
    if [[ ! -f ./useepod ]]; then
        echo 'export DDL_RPC_TLS_CERTIFICATE=$(cat ../certs/cert.crt)' >>source.sh
        echo 'export DDL_RPC_TLS_KEY=$(cat ../certs/cert.key)' >>source.sh
    else
        echo 'export DDL_RPC_TLS_CERTIFICATE=$(cat ./epod/ep.crt)' >>source.sh
        echo 'export DDL_RPC_TLS_KEY=$(cat ./epod/ep.key)' >>source.sh
    fi
    echo "export DDL_RPC_CLIENT_AUTHENTICATION=NoClientCert" >>source.sh
    if [[ ${weatherSetup} == "true" ]]; then
        echo "export WEATHERAPI_ENABLED=true" >>source.sh
        echo "export WEATHERAPI_PROVIDER=$weatherProvider" >>source.sh
        echo "export WEATHERAPI_KEY=${weatherAPI}" >>source.sh
        echo "export WEATHERAPI_UNIT=${weatherUnit}" >>source.sh
    else
        echo "export WEATHERAPI_ENABLED=false" >>source.sh
    fi
    if [[ ${knowledgeSetup} == "true" ]]; then
        echo "export KNOWLEDGE_ENABLED=true" >>source.sh
        echo "export KNOWLEDGE_INTENT_GRAPH=${knowledgeIntent}" >> source.sh
        if [[ ${knowledgeProvider} == "houndify" ]]; then
            echo "export KNOWLEDGE_PROVIDER=houndify" >> source.sh
            echo "export KNOWLEDGE_KEY=${knowledgeKey}" >>source.sh
            echo "export KNOWLEDGE_ID=${knowledgeID}" >>source.sh
        else
            echo "export KNOWLEDGE_PROVIDER=openai" >> source.sh
            echo "export KNOWLEDGE_KEY=${knowledgeKey}" >> source.sh
        fi
    else
        echo "export KNOWLEDGE_ENABLED=false" >>source.sh
    fi
    echo "export WEBSERVER_PORT=${webport}" >>source.sh
    if [[ -f ./pico.key ]]; then
        picoKey=$(cat ./pico.key)
        echo "export STT_SERVICE=leopard" >>source.sh
        echo "export PICOVOICE_APIKEY=${picoKey}" >> source.sh
    elif [[ ${sttService} == "vosk" ]]; then
        echo "export STT_SERVICE=vosk" >>source.sh
        echo "export STT_LANGUAGE=${languageNum}" >>source.sh
    else
        echo "export STT_SERVICE=coqui" >>source.sh
    fi
    echo "export DEBUG_LOGGING=true" >>source.sh
    echo "export WIREPOD_EX_TMP_PATH=/tmp" >>source.sh
    echo "export WIREPOD_EX_DATA_PATH=./data" >>source.sh
    echo "export WIREPOD_EX_NVM_PATH=./nvm" >>source.sh
    cd ..
    echo
    echo "Created source.sh file!"
    echo
    if [[ ! -f ./chipper/useepod ]]; then
        cd certs
        echo "Creating server_config.json for robot"
        echo '{"jdocs": "REPLACEME", "tms": "REPLACEME", "chipper": "REPLACEME", "check": "REPLACECONN/ok:80", "logfiles": "s3://anki-device-logs-prod/victor", "appkey": "oDoa0quieSeir6goowai7f"}' >server_config.json
        address=$(cat address)
        sed -i "s/REPLACEME/${address}:${port}/g" server_config.json
        sed -i "s/REPLACECONN/${address}/g" server_config.json
        cd ..
    else
        mkdir -p certs
        cd certs
        echo "Creating server_config.json for robot"
        echo '{"jdocs": "escapepod.local:443", "tms": "escapepod.local:443", "chipper": "escapepod.local:443", "check": "escapepod.local/ok:80", "logfiles": "s3://anki-device-logs-prod/victor", "appkey": "oDoa0quieSeir6goowai7f"}' >server_config.json
        cd ..
    fi
    echo "Created!"
    echo
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
    if [[ ${TARGET} == "macos" ]]; then
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
    echo >>wire-pod.service
    echo "[Service]" >>wire-pod.service
    echo "Type=simple" >>wire-pod.service
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
    if [[ ${TARGET} == "macos" ]]; then
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

function firstPrompt() {
    read -p "Enter a number (1): " yn
    case $yn in
        "1")
            echo
            getPackages
            getSTT
            getLanguage
            generateCerts
            buildChipper
            makeSource
            echo -e "\033[33m\033[1mEverything is set up! wire-pod is ready to start!\033[0m"
            echo
            if [[ -f ./chipper/useepod ]]; then
                echo "You chose to use escapepod.local, so you do not need to run any SCP commands for a prod bot to use this. You just need to put on an official escape pod OTA. The instructions can be found in the root of this repo."
                echo
            fi
            ;;
        "2")
            echo
            getPackages
            buildCloud
            ;;
        "3")
            echo
            getPackages
            buildChipper
            ;;
        "4")
            echo
            rm -f ./stt/completed
            getSTT
            getLanguage
            ;;
        "5")
            echo
            getPackages
            generateCerts
            ;;
        "6")
            echo
            makeSource
            ;;
        "")
            echo
            getPackages
            getSTT
            getLanguage
            generateCerts
            buildChipper
            makeSource
            echo -e "\033[33m\033[1mEverything is set up! wire-pod is ready to start!\033[0m"
            # echo "Everything is done! To copy everything needed to your bot, run this script like this:"
            # echo "Usage: sudo ./setup.sh scp <vector's ip> <path/to/ssh-key>"
            # echo "Example: sudo ./setup.sh scp 192.168.1.150 /home/wire/id_rsa_Vector-R2D2"
            # echo
            # echo "If your Vector is on Wire's custom software or you have an old dev build, you can run this command without an SSH key:"
            # echo "Example: sudo ./setup.sh scp 192.168.1.150"
            echo
            # if [[ -f ./chipper/useepod ]]; then
            # 	echo "You chose to use escapepod.local, so you do not need to run any SCP commands for a prod bot to use this. You just need to put on an official escape pod OTA. The instructions can be found in the root of this repo."
            # 	echo
            # fi
            ;;
        *)
            echo "Please answer with 1, 2, 3, 4, 5, 6, or just press enter with no input for 1."
            firstPrompt
            ;;
    esac
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

echo "What would you like to do?"
echo "1: Full Setup (recommended) (builds chipper, gets STT stuff, generates certs, creates source.sh file, and creates server_config.json for your bot"
echo "2: Just build vic-cloud"
echo "3: Just build chipper"
echo "4: Just get STT assets"
echo "5: Just generate certs"
echo "6: Create wire-pod config file (change/add API keys)"
echo "(NOTE: You can just press enter without entering a number to select the default, recommended option)"
echo
firstPrompt
