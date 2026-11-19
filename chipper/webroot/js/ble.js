const vectorEpodSetup = "https://wp.pvic.xyz";
let authEl = document.getElementById("botAuth");
let statusP = document.createElement("p");
let OTAUpdating = false;
let scanFailures = 0;
const maxScanFailures = 3;

const externalSetup = document.createElement("a");
externalSetup.href = vectorEpodSetup;
externalSetup.innerHTML = vectorEpodSetup;

function showBotAuth() {
  GetLog = false;
  toggleSections("section-botauth", "icon-BotAuth");
  checkBLECapability();
}

function toggleSections(showSection, icon) {
  const sections = ["section-intents", "section-log", "section-botauth", "section-version", "section-uicustomizer"];
  sections.forEach((section) => (document.getElementById(section).style.display = "none"));
  document.getElementById(showSection).style.display = "block";
  updateColor(icon);
}

function setBLEUsable(usable) {
  document.getElementById("bleDesc").style.display = usable ? "" : "none";
  document.getElementById("oskrSection").style.display = usable ? "none" : "block";
  document.getElementById("legacySetup").style.display = usable ? "block" : "none";
}

function useLegacySetup() {
  OTAUpdating = false;
  document.getElementById("disconnectButton").innerHTML = "";
  setBLEUsable(false);
  showExternalSetupInstructions("");
  fetch("/api-ble/stop_ota").catch(() => {}).then(() => fetch("/api-ble/disconnect")).catch(() => {});
}

function checkBLECapability() {
  updateAuthel("Checking if wire-pod can use BLE directly...");
  fetch("/api-ble/init")
    .then((response) => response.text())
    .then((response) => {
      if (response.includes("success")) {
        setBLEUsable(true);
        beginBLESetup();
      } else {
        setBLEUsable(false);
        showExternalSetupInstructions(response);
      }
    }).catch(() => {
      setBLEUsable(false);
      showExternalSetupInstructions("");
    });
}

function showExternalSetupInstructions(reason = "") {
  authEl.innerHTML = `
    <p>Head to the following site on any device with Bluetooth support to set up your Vector.</p>
    <a href="${vectorEpodSetup}" target="_blank">${vectorEpodSetup}</a>
    <br>
    <small class="desc">Note: with OSKR/dev robots, it might give a warning about firmware. This can be ignored.</small>
  `;
  if (reason.includes("turned off")) {
    const p = document.createElement("p");
    p.textContent = "Bluetooth is turned off on this machine. Turn it on and retry to set up Vector directly from wire-pod.";
    authEl.prepend(p);
  }
  const retry = document.createElement("button");
  retry.className = "small-button";
  retry.textContent = "Retry BLE";
  retry.onclick = checkBLECapability;
  authEl.appendChild(document.createElement("br"));
  authEl.appendChild(retry);
}

function beginBLESetup() {
  authEl.innerHTML = `
    <p>1. Place Vector on the charger.</p>
    <p>2. Double press the button. A key should appear on screen.</p>
    <p>3. Click 'Begin Scanning' and pair with your Vector.</p>
    <button onclick="scanRobots(false)">Begin Scanning</button>
  `;
}

function reInitBLE() {
  return fetch("/api-ble/disconnect")
    .catch(() => {})
    .then(() => fetch("/api-ble/init"))
    .then((response) => response.text())
    .catch(() => "")
    .then((response) => {
      if (response.includes("success")) {
        return;
      }
      setBLEUsable(false);
      showExternalSetupInstructions(response);
      return new Promise(() => {});
    });
}

function scanRobots(returning) {
  const disconnectButtonDiv = document.getElementById("disconnectButton");
  disconnectButtonDiv.innerHTML = `
    <button class="small-button" onclick="disconnect()">Disconnect</button>
  `;
  updateAuthel("Scanning...");
  fetch("/api-ble/scan", { method: "POST", headers: { "Content-Type": "application/x-www-form-urlencoded" } })
    .then((response) => response.json())
    .then((parsed) => {
      scanFailures = 0;
      const buttonsDiv = document.createElement("div");
      parsed.forEach((robot) => {
        const button = document.createElement("button");
        button.textContent = robot.name;
        button.onclick = () => connectRobot(robot.id);
        buttonsDiv.appendChild(button);
      });

      const rescanButton = document.createElement("button");
      rescanButton.className = "small-button";
      rescanButton.textContent = "Re-scan";
      rescanButton.onclick = () => {
        updateAuthel("Reiniting BLE then scanning...");
        reInitBLE().then(() => scanRobots(false));
      };

      if (parsed.length === 0) {
        updateAuthel("No robots found. Make sure Vector is showing a key on his screen, then re-scan.");
      } else if (returning) {
        updateAuthel("Incorrect PIN was entered. Click on the robot you would like to pair with.");
      } else {
        updateAuthel("Click on the robot you would like to pair with.");
      }
      authEl.appendChild(buttonsDiv);
      authEl.appendChild(rescanButton);
    }).catch(() => {
      scanFailures++;
      if (scanFailures >= maxScanFailures) {
        scanFailures = 0;
        showBLEError("Scanning failed several times. Make sure Bluetooth is on and working on the wire-pod machine.");
        return;
      }
      updateAuthel("Error scanning. Reiniting BLE then scanning...");
      reInitBLE().then(() => scanRobots(false));
    });
}

function showBLEError(msg) {
  updateAuthel(msg);
  const retry = document.createElement("button");
  retry.textContent = "Start over";
  retry.onclick = () => {
    updateAuthel("Starting Bluetooth...");
    reInitBLE().then(() => checkBLECapability());
  };
  authEl.appendChild(retry);
}

function disconnect() {
  authEl.innerHTML = "Disconnecting...";
  OTAUpdating = false;
  fetch("/api-ble/stop_ota").then(() => fetch("/api-ble/disconnect").then(() => checkBLECapability())).catch(() => checkBLECapability());
}

function connectRobot(id) {
  updateAuthel("Connecting to robot...");
  fetch(`/api-ble/connect?id=${id}`)
    .then((response) => response.text())
    .then((response) => {
      if (response.includes("success")) {
        createPinEntry();
      } else {
        alert("Error connecting. BLE will be reset and this will return to the first screen of setup.");
        updateAuthel("Resetting BLE...");
        setTimeout(checkBLECapability, 3000);
      }
    }).catch(() => {
      alert("Error connecting. BLE will be reset and this will return to the first screen of setup.");
      updateAuthel("Resetting BLE...");
      setTimeout(checkBLECapability, 3000);
    })
}

function createPinEntry() {
  authEl.innerHTML = `
    <p>Enter the pin shown on Vector's screen.</p>
    <input type="text" id="pinEntry" class="pin-input" placeholder="PIN" maxlength="6" inputmode="numeric" autocomplete="off">
    <br>
    <button onclick="sendPin()">Send PIN</button>
  `;
}

function sendPin() {
  const pin = document.getElementById("pinEntry").value;
  updateAuthel("Sending PIN...");
  fetch("/api-ble/send_pin?pin=" + encodeURIComponent(pin))
    .then((response) => response.text())
    .then((response) => {
      if (response.includes("incorrect pin") || response.includes("length of pin")) {
        updateAuthel("Wrong PIN... Reiniting BLE then scanning...");
        reInitBLE().then(() => scanRobots(true));
      } else {
        wifiCheck();
      }
    }).catch(() => {
      updateAuthel("Error sending PIN... Reiniting BLE then scanning...");
      reInitBLE().then(() => scanRobots(true));
    });
}

function wifiCheck() {
  fetch("/api-ble/get_wifi_status")
    .then((response) => response.text())
    .then((response) => {
      if (response === "1") {
        whatToDo();
      } else {
        scanWifi();
      }
    }).catch(() => {
      updateAuthel("Error checking Wi-Fi status... Reiniting BLE then scanning...");
      reInitBLE().then(() => scanRobots(true));
    });
}

let wifiNetworks = [];

function scanWifi() {
  updateAuthel("Scanning for Wi-Fi networks...");
  fetch("/api-ble/scan_wifi")
    .then((response) => response.json())
    .then((networks) => {
      const seen = new Set();
      wifiNetworks = networks.filter((n) => {
        if (!n.ssid || seen.has(n.ssid)) {
          return false;
        }
        seen.add(n.ssid);
        return true;
      });
      showWifiForm("");
    }).catch(() => {
      updateAuthel("Error scanning Wi-Fi networks... Reiniting BLE then scanning...");
      reInitBLE().then(() => scanRobots(true));
    });
}

function showWifiForm(selectedSsid) {
  if (wifiNetworks.length === 0) {
    updateAuthel("No Wi-Fi networks found.");
  } else {
    updateAuthel("Select a Wi-Fi network to connect Vector to.");
  }
  const select = document.createElement("select");
  wifiNetworks.forEach((network, i) => {
    const opt = document.createElement("option");
    opt.value = i;
    opt.textContent = network.ssid + (network.authtype === 0 ? " (open)" : "");
    if (network.ssid === selectedSsid) {
      opt.selected = true;
    }
    select.appendChild(opt);
  });
  const pass = document.createElement("input");
  pass.type = "text";
  pass.id = "passEntry";
  pass.placeholder = "Password";
  const connectButton = document.createElement("button");
  connectButton.textContent = "Connect to Wi-Fi";
  const rescan = document.createElement("button");
  rescan.className = "small-button";
  rescan.textContent = "Scan Again";
  rescan.onclick = scanWifi;

  const syncPass = () => {
    const network = wifiNetworks[select.value];
    pass.style.display = network && network.authtype === 0 ? "none" : "";
  };
  select.onchange = syncPass;
  syncPass();
  connectButton.onclick = () => {
    const network = wifiNetworks[select.value];
    if (!network) {
      return;
    }
    connectWifi(network.ssid, network.authtype, network.authtype === 0 ? "" : pass.value);
  };

  if (wifiNetworks.length > 0) {
    authEl.appendChild(select);
    authEl.appendChild(document.createElement("br"));
    authEl.appendChild(pass);
    authEl.appendChild(document.createElement("br"));
    authEl.appendChild(connectButton);
    authEl.appendChild(document.createElement("br"));
  }
  authEl.appendChild(rescan);
}

function connectWifi(ssid, authtype, password) {
  updateAuthel("Connecting Vector to Wi-Fi...");
  fetch(`/api-ble/connect_wifi?ssid=${encodeURIComponent(ssid)}&password=${encodeURIComponent(password)}&authType=${encodeURIComponent(authtype)}`)
    .then((response) => response.text())
    .then((response) => {
      if (!response.includes("255")) {
        alert("Error connecting, likely incorrect password");
        showWifiForm(ssid);
      } else {
        whatToDo();
      }
    }).catch(() => {
      updateAuthel("Error connecting to Wi-Fi... Reiniting BLE then scanning...");
      reInitBLE().then(() => scanRobots(true));
    });
}

function whatToDo() {
  fetch("/api-ble/get_robot_status")
    .then((response) => response.text())
    .then((response) => {
      switch (response) {
        case "in_recovery_prod":
          doOTA("http://anki2.ca:81/otas/Anki/prod/2.0.1.6086ep.ota");
          break;
        case "in_recovery_dev":
          doOTA("http://ota.pvic.xyz/vic/latest/dev.ota");
          break;
        case "in_firmware_nonep":
          showRecoveryInstructions();
          break;
        case "in_firmware_dev":
          showDevSSHSetup();
          break;
        case "in_firmware_dev_ready":
        case "in_firmware_ep":
          showAuthButton();
          break;
        default:
          showBLEError("Couldn't get the robot's status (" + response + ").");
      }
    }).catch(() => {
      showBLEError("Couldn't get the robot's status.");
    });
}

function showRecoveryInstructions() {
  authEl.innerHTML = `
    <p>1. Place Vector on the charger.</p>
    <p>2. Hold the button for 15 seconds. He will turn off - keep holding it until he turns back on.</p>
    <p>3. Click 'Begin Scanning' and pair with your Vector.</p>
    <button onclick="scanRobots(false)">Begin Scanning</button>
  `;
  alert("Your bot is not on the correct firmware for wire-pod. Follow the directions to put him in recovery mode.");
}

let devSSHIP = "";

function showDevSSHSetup() {
  updateAuthel("This is a dev robot, so wire-pod needs to set it up over SSH first.");
  const go = document.createElement("button");
  go.textContent = "Set up bot";
  go.onclick = devSSHSetup;
  const skip = document.createElement("button");
  skip.className = "small-button";
  skip.textContent = "Already set up, skip";
  skip.onclick = showAuthButton;
  authEl.appendChild(go);
  authEl.appendChild(document.createElement("br"));
  authEl.appendChild(skip);
}

function devSSHSetup() {
  updateAuthel("Disconnecting BLE and starting SSH setup...");
  fetch("/api-ble/dev_ssh_setup", { method: "POST" })
    .then((response) => response.text())
    .then((response) => {
      if (!response.startsWith("running")) {
        alert(response);
        showDevSSHSetup();
        return;
      }
      devSSHIP = response.slice("running".length).trim();
      document.getElementById("disconnectButton").innerHTML = "";
      pollDevSSHSetup();
    }).catch(() => {
      showBLEError("Error starting SSH setup.");
    });
}

function showDevSSHKeyPrompt() {
  updateAuthel("The default dev SSH key didn't work on this robot. Upload the robot's SSH key to continue.");
  const keyInput = document.createElement("input");
  keyInput.type = "file";
  const go = document.createElement("button");
  go.textContent = "Set up bot";
  go.onclick = () => devSSHSetupWithKey(keyInput.files[0]);
  authEl.appendChild(keyInput);
  authEl.appendChild(document.createElement("br"));
  authEl.appendChild(go);
}

function devSSHSetupWithKey(key) {
  if (!key) {
    alert("Select the robot's SSH key first.");
    return;
  }
  updateAuthel("Starting SSH setup...");
  const formData = new FormData();
  formData.append("ip", devSSHIP);
  formData.append("key", key);
  fetch("/api-ssh/setup", { method: "POST", body: formData })
    .then((response) => response.text())
    .then((response) => {
      if (!response.includes("running")) {
        alert(response);
        showDevSSHKeyPrompt();
        return;
      }
      pollDevSSHSetup();
    }).catch(() => {
      showBLEError("Error starting SSH setup.");
    });
}

function pollDevSSHSetup() {
  const interval = setInterval(() => {
    fetch("/api-ssh/get_setup_status")
      .then((response) => response.text())
      .then((response) => {
        if (response.includes("done")) {
          clearInterval(interval);
          showRepairAfterSSH();
        } else if (response.includes("ssh key rejected")) {
          clearInterval(interval);
          showDevSSHKeyPrompt();
        } else if (response.includes("error")) {
          clearInterval(interval);
          let msg = response;
          if (response.includes("no route to host") || response.includes("i/o timeout")) {
            msg = "wire-pod couldn't reach the robot over SSH. Make sure it's on the same network as wire-pod.";
          }
          showBLEError("SSH setup failed: " + msg);
        } else if (!response.includes("not running")) {
          updateAuthel(response);
        }
      });
  }, 1000);
}

function showRepairAfterSSH() {
  authEl.innerHTML = `
    <p>Phase 1 of setup completed successfully. Vector will restart into onboarding mode, and will <b>need to be paired again</b> to finish setup.</p>
    <p>1. Wait for Vector's face to come back.</p>
    <p>2. Double press the button. A key should appear on screen.</p>
    <p>3. Click 'Begin Scanning' and pair with your Vector.</p>
    <button onclick="updateAuthel('Starting Bluetooth...'); reInitBLE().then(() => scanRobots(false))">Begin Scanning</button>
  `;
}

function showAuthButton() {
  authEl.innerHTML = `<button onclick="doAuth()">AUTHENTICATE</button>`;
}

function doOTA(url) {
  updateAuthel("Starting OTA update...");
  fetch("/api-ble/start_ota?url=" + encodeURIComponent(url))
    .then((response) => response.text())
    .then((response) => {
      if (response.includes("success")) {
        OTAUpdating = true;
        const interval = setInterval(() => {
          fetch("/api-ble/get_ota_status")
            .then((otaResponse) => otaResponse.text())
            .then((otaResponse) => {
              updateAuthel(otaResponse);
              if (otaResponse.includes("complete")) {
                alert("The OTA update is complete. When the bot reboots, follow the steps to re-pair the bot with wire-pod. wire-pod will then authenticate the robot and setup will be complete.");
                OTAUpdating = false;
                clearInterval(interval);
                checkBLECapability();
              } else if (otaResponse.includes("Error")) {
                OTAUpdating = false;
                clearInterval(interval);
                showBLEError(otaResponse);
              } else if (otaResponse.includes("stopped") || !OTAUpdating) {
                clearInterval(interval);
              }
            }).catch(() => {});
        }, 2000);
      } else {
        whatToDo();
      }
    }).catch(() => {
      showBLEError("Error starting the OTA update.");
    });
}

function updateAuthel(update) {
  authEl.innerHTML = "";
  const p = document.createElement("p");
  p.textContent = update;
  authEl.appendChild(p);
}

function doAuth() {
  updateAuthel("Authenticating your Vector...");
  fetch("/api-ble/do_auth")
    .then((response) => response.text())
    .then((response) => {
      if (response.includes("error")) {
        showAuthError();
      } else {
        doOnboard();
      }
    }).catch(() => {
      showAuthError();
    });
}

function showAuthError() {
  updateAuthel("Authentication failure. Try again in ~15 seconds. If it happens again, check the troubleshooting guide:");
  const troubleshootingLink = document.createElement("a");
  troubleshootingLink.href = "https://github.com/kercre123/wire-pod/wiki/Troubleshooting#error-logging-in-the-bot-is-likely-unable-to-communicate-with-your-wire-pod-instance";
  troubleshootingLink.target = "_blank";
  troubleshootingLink.innerText = "https://github.com/kercre123/wire-pod/wiki/Troubleshooting";
  authEl.appendChild(document.createElement("br"));
  authEl.appendChild(troubleshootingLink);
}

function doOnboard() {
  updateAuthel("Authentication was successful! Taking Vector out of onboarding mode...");
  fetch("/api-ble/onboard?with_anim=false").then((response) => response.text()).then((response) => {
    if (response.includes("error")) {
      showBLEError("Vector was authenticated, but taking him out of onboarding mode failed: " + response);
      return;
    }
    document.getElementById("disconnectButton").innerHTML = "";
    updateAuthel("Vector is now fully set up! Use the Bot Settings tab to further configure your bot.");
    const back = document.createElement("button");
    back.className = "small-button";
    back.textContent = "Set up another bot";
    back.onclick = checkBLECapability;
    authEl.appendChild(back);
  }).catch(() => {
    showBLEError("Vector was authenticated, but taking him out of onboarding mode failed.");
  });
}
