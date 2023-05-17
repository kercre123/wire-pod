const vectorEpodSetup = "https://keriganc.com/vector-wirepod-setup"

var authEl = document.getElementById("botAuth")
var statusP = document.createElement("p")
var externalSetup = document.createElement("a")
externalSetup.href = vectorEpodSetup
externalSetup.innerHTML = vectorEpodSetup

function showBotAuth() {
    GetLog = false
    document.getElementById("section-intents").style.display = "none";
    document.getElementById("section-language").style.display = "none";
    document.getElementById("section-log").style.display = "none";
    document.getElementById("section-botauth").style.display = "block";
    updateColor("icon-BotAuth");
    checkBLECapability()

}

function checkBLECapability() {
    // fetch("/api-ble/init")
    //     .then(response => response.text())
    //     .then((response) => {
    //         if (response.includes("success")) {
    //             BeginBLESetup()
    //         } else if (response.includes("error")) {
    //             statusP.innerHTML = "Error initializing bluetooth on the device running wire-pod. Use the following site on any machine with Bluetooth for setup:"
    //             authEl.innerHTML = ""
    //             authEl.appendChild(statusP)
    //             authEl.appendChild(document.createElement("br"))
    //             authEl.appendChild(externalSetup)
    //         }
    //     })
    authEl.innerHTML = ""
    m1 = document.createElement("p")
    m2 = document.createElement("a")
    m3 = document.createElement("small")
    m1.innerHTML = "Head to the following site on any device with Bluetooth support to set up your Vector."
    m2.text = "https://keriganc.com/vector-wirepod-setup"
    m2.href = "https://keriganc.com/vector-wirepod-setup"
    m2.target = "_blank"
    m3.innerHTML = "Note: if you have an OSKR/dev-unlocked robot, follow the instructions in the section below this one BEFORE using the web setup."
    m1.class = "center"
    m2.class = "center"
    m3.class = "center"
    authEl.appendChild(m1)
    //authEl.appendChild(document.createElement("br"))
    authEl.appendChild(m2)
    authEl.appendChild(document.createElement("br"))
    authEl.appendChild(m3)
}

function BeginBLESetup() {
    authEl.innerHTML = ""
    m1 = document.createElement("p")
    m1.innerHTML = "1. Place Vector on the charger."
    m2 = document.createElement("p")
    m2.innerHTML = "2. Double press the button. A key should appear on screen."
    m3 = document.createElement("p")
    m3.innerHTML = "3. Click 'Begin Scanning' and pair with your Vector."
    button = document.createElement("button")
    button.innerHTML = "Begin Scanning"
    button.onclick = function(){ScanRobots(false)}
    authEl.appendChild(m1)
    authEl.appendChild(m2)
    authEl.appendChild(m3)
    authEl.appendChild(button)

}

var Scanning = false
var IsScanning = false

function ScanRobots(returning) {
    disconnectButtonDiv = document.getElementById("disconnectButton")
    disconnectButton = document.createElement("button")
    disconnectButton.onclick = function(){Disconnect()}
    disconnectButton.innerHTML = "Disconnect"
    disconnectButtonDiv.appendChild(disconnectButton)
    Scanning = true
    authEl.innerHTML = ""
    statusDiv = document.createElement("div")
    buttonsDiv = document.createElement("div")
    buttonsDiv.class = "center"
    statusDiv.class = "center"
    if (returning) {
        incorrectPin = document.createElement("p")
        incorrectPin.innerHTML = "Incorrect PIN was entered, scanning again"
        statusDiv.appendChild(incorrectPin)
    }
    scanNotice = document.createElement("small")
    scanNotice.innerHTML = "Scanning..."
    statusDiv.appendChild(scanNotice)
    authEl.appendChild(statusDiv)
    IsScanning = true
    let xhr = new XMLHttpRequest();
    xhr.open("POST", "/api-ble/scan");
    xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
    xhr.send();
    xhr.onload = function() {
            response = xhr.response
            IsScanning = false
            if (!Scanning) {
                clearInterval(interval)
                return
            }
            console.log(response)
            parsed = JSON.parse(response)
            buttonsDiv.innerHTML = ""
            authEl.innerHTML = ""
            for (var i = 0; i < parsed.length; i++) {
                button = document.createElement("button")
                id = parsed[i]["id"]
                button.innerHTML = parsed[i]["name"]
                button.onclick = function(){Scanning = false; ConnectRobot(id);}
                buttonsDiv.appendChild(button)
            }
            authEl.appendChild(buttonsDiv)
        }
    interval = setInterval(function(){
        if (!Scanning) {
            clearInterval(interval)
            return
        }
        scanNotice.innerHTML = "Scanning..."
        statusDiv.innerHTML = ""
        statusDiv.appendChild(scanNotice)
        IsScanning = true
        let xhr = new XMLHttpRequest();
        xhr.open("POST", "/api-ble/scan");
        xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
        xhr.send();
        xhr.onload = function() {
            response = xhr.response
            IsScanning = false
            if (!Scanning) {
                clearInterval(interval)
                return
            }
            console.log(response)
            parsed = JSON.parse(response)
            buttonsDiv.innerHTML = ""
            authEl.innerHTML = ""
            for (var i = 0; i < parsed.length; i++) {
                button = document.createElement("button")
                id = parsed[i]["id"]
                button.innerHTML = parsed[i]["name"]
                button.onclick = (function(id) {
                    return function() {
                        Scanning = false;
                        ConnectRobotBuffer(id);
                    };
                  })(id);
                buttonsDiv.appendChild(button)
            }
            authEl.appendChild(buttonsDiv)
        }
    }, 5000)
}

function ConnectRobotBuffer(id) {
    authEl.innerHTML = ""
    statusP.innerHTML = "Connecting to robot..."
    authEl.appendChild(statusP)
    // if scanning, dont make connection request
    if (IsScanning) {
        console.log("Scan request being made, wait to connect robot...")
        inte = setInterval(function(){
            if (!IsScanning) {
                setTimeout(function(){
                    clearInterval(inte)
                    console.log("connecting robot...")
                    ConnectRobot(id)
                }, 1000)
            }
        }, 500)
    }
}

function Disconnect() {
    disconnectButtonDiv = document.getElementById("disconnectButton")
    disconnectButtonDiv.innerHTML = ""
    authEl.innerHTML = ""
    statusP.innerHTML = "Disconnecting..."
    authEl.appendChild(statusP)
    fetch("/api-ble/disconnect")
    .then((response) => {
    setTimeout(function(){
        checkBLECapability();
    }, 2000)
})
}

function ConnectRobot(id) {
    fetch("/api-ble/connect?id=" + id)
    .then(response => response.text())
    .then((response) => {
        if (response.includes("success")) {
            statusP.innerHTML = "Connected to robot! Loading pin screen..."
            authEl.innerHTML = ""
            authEl.appendChild(statusP)
            CreatePinEntry()
            return
        }
    })
}

function validateInput(input) {
    return input.value.length <= 6 && /^\d+$/.test(input.value);
  }

function CreatePinEntry() {
    authEl.innerHTML = ""
    statusDiv = document.createElement("div")
    statusP.innerHTML = "Enter the pin shown on Vector's screen."
    statusDiv.appendChild(statusP)
    authEl.appendChild(statusDiv)
    pinEntry = document.createElement("input")
    pinEntry.type = "text"
    pinEntry.id = "pinEntry"
    pinEntry.name = "pinEntry"
    pinEntry.placeholder = "Enter PIN here"
    pinEntry.setAttribute("type", "text");
    pinEntry.setAttribute("maxlength", "6");
    pinEntry.setAttribute("oninput", "validateInput(this)");
    button = document.createElement("button")
    button.onclick = function(){SendPin()}
    button.innerHTML = "Send PIN"
    authEl.appendChild(pinEntry)
    authEl.appendChild(document.createElement("br"))
    authEl.appendChild(button)
    return
}

function SendPin() {
    pin = document.getElementById("pinEntry").value
    authEl.innerHTML = ""
    fetch("/api-ble/send_pin?pin=" + pin)
    .then(response => response.text())
    .then((response) => {
        if (response.includes("incorrect pin")) {
            ScanRobots(true)
        } else {
            // create auth button
            WifiCheck()
        }
        return
    })
}

function WifiCheck() {
    fetch("/api-ble/get_wifi_status")
    .then(response => response.text())
    .then((response) => {
        console.log(response)
        if (response == "1") {
            DoAuth()
        } else {
            ScanWifi()
        }
        return
    })
}

//var parsedScan

function ScanWifi() {
    authEl.innerHTML = ""
    statusP.innerHTML = "Scanning for Wi-Fi networks..."
    authEl.appendChild(statusP)
        var xhr = new XMLHttpRequest();
        xhr.open("GET", "/api-ble/scan_wifi", true);
        xhr.onreadystatechange = function() {
          if (this.readyState === XMLHttpRequest.DONE && this.status === 200) {
            authEl.innerHTML = ""
            // create scan again button
            var scanAgain = document.createElement("button")
            scanAgain.innerHTML = "Scan Again"
            scanAgain.onclick = function(){ScanWifi()}
            authEl.appendChild(scanAgain)
            authEl.appendChild(document.createElement("br"))
            // add network buttons
            var networks = JSON.parse(this.responseText);
            for (var i = 0; i < networks.length; i++) {
              var ssid = networks[i].ssid;
              if (ssid != "") {
              var authtype = networks[i].authtype;
              var btn = document.createElement("button");
              btn.innerHTML = ssid;
              btn.onclick = (function(ssid, authtype) {
                return function() {
                    CreateWiFiPassEntry(ssid, authtype);
                };
              })(ssid, authtype);
              authEl.appendChild(btn);
            }
            }
          }
        };
        xhr.send();
      }

function CreateWiFiPassEntry(ssid, authtype) {
    console.log(ssid)
    console.log(authtype)
    authEl.innerHTML = ""
    againButton = document.createElement("button")
    againButton.innerHTML = "Scan Again"
    againButton.onclick = function(){ScanWifi()}
    authEl.appendChild
    statusDiv = document.createElement("div")
    statusP.innerHTML = "Enter the password for " + ssid
    statusDiv.appendChild(statusP)
    authEl.appendChild(statusDiv)
    pinEntry = document.createElement("input")
    pinEntry.type = "text"
    pinEntry.id = "passEntry"
    pinEntry.name = "passEntry"
    pinEntry.placeholder = "Password"
    button = document.createElement("button")
    button.onclick = function(){ConnectWifi(ssid, authtype)}
    button.innerHTML = "Connect to Wi-Fi"
    authEl.appendChild(pinEntry)
    authEl.appendChild(document.createElement("br"))
    authEl.appendChild(button)
    return
}

function ConnectWifi(ssid, authtype) {
    password = document.getElementById("passEntry").value
    authEl.innerHTML = ""
    passP = document.createElement("p")
    passP.innerHTML = "Connecting Vector to Wi-Fi..."
    authEl.appendChild(passP)
    fetch("/api-ble/connect_wifi?ssid=" + ssid + "&password=" + password + "&authType=" + authtype)
    .then(response => response.text())
    .then((response) => {
        if (!response.includes("255")) {
            alert("Error connecting, likely incorrect password")
            CreateWiFiPassEntry(ssid, authtype)
        } else {
            authEl.innerHTML = ""
            button = document.createElement("button")
            button.innerHTML = "Click to authenticate"
            button.onclick = function(){DoAuth()}
            authEl.appendChild(button)
        }
    })
}

function CheckFirmware() {
    fetch("/api-ble/get_firmware")
    .then(response => response.text())
    .then((response) => {
        let splitFirmware = response.split("-")
        console.log(splitFirmware)
    })
}

function DoAuth() {
    authEl.innerHTML = ""
    authP = document.createElement("p")
    authP.innerHTML = "Authenticating your Vector..."
    authEl.appendChild(authP)
    fetch("/api-ble/do_auth")
    .then(response => response.text())
    .then((response) => {
        console.log(response)
        authEl.innerHTML = ""
        authP.innerHTML = "Authentication complete!"
        authEl.appendChild(authP)
        fetch("/api-ble/disconnect")
        disconnectButtonDiv = document.getElementById("disconnectButton")
        disconnectButtonDiv.innerHTML = ""
        disconnectButton = document.createElement("button")
        disconnectButton.onclick = function(){checkBLECapability()}
        disconnectButton.innerHTML = "Back to setup"
        disconnectButtonDiv.appendChild(disconnectButton)
    })
}