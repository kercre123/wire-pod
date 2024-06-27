const vectorEpodSetup = "https://wpsetup.keriganc.com";
let authEl = document.getElementById("botAuth");
let statusP = document.createElement("p");
let OTAUpdating = false;

const externalSetup = document.createElement("a");
externalSetup.href = vectorEpodSetup;
externalSetup.innerHTML = vectorEpodSetup;

function showBotAuth() {
  GetLog = false;
  toggleSections("section-botauth", "icon-BotAuth");
  checkBLECapability();
}

function toggleSections(showSection, icon) {
  const sections = ["section-intents", "section-language", "section-log", "section-botauth", "section-version", "section-uicustomizer"];
  sections.forEach((section) => (document.getElementById(section).style.display = "none"));
  document.getElementById(showSection).style.display = "block";
  updateColor(icon);
}

function checkBLECapability() {
  updateAuthel("Checking if wire-pod can use BLE directly...");
  fetch("/api-ble/init")
    .then((response) => response.text())
    .then((response) => {
      if (response.includes("success")) {
        beginBLESetup();
      } else {
        showExternalSetupInstructions();
      }
    });
}

function showExternalSetupInstructions() {
  authEl.innerHTML = `
    <p>Head to the following site on any device with Bluetooth support to set up your Vector.</p>
    <a href="${vectorEpodSetup}" target="_blank">${vectorEpodSetup}</a>
    <br>
    <small class="desc">Note: with OSKR/dev robots, it might give a warning about firmware. This can be ignored.</small>
  `;
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
  fetch("/api-ble/disconnect").then(() => fetch("/api-ble/init"));
}

function scanRobots(returning) {
  const disconnectButtonDiv = document.getElementById("disconnectButton");
  disconnectButtonDiv.innerHTML = `
    <button onclick="disconnect()">Disconnect</button>
  `;
  updateAuthel("Scanning...");
  fetch("/api-ble/scan", { method: "POST", headers: { "Content-Type": "application/x-www-form-urlencoded" } })
    .then((response) => response.json())
    .then((parsed) => {
      authEl.innerHTML = returning ? "<p>Incorrect PIN was entered, scanning again...</p>" : "";
      authEl.innerHTML += "<small>Scanning...</small>";

      const buttonsDiv = document.createElement("div");
      parsed.forEach((robot) => {
        const button = document.createElement("button");
        button.innerHTML = robot.name;
        button.onclick = () => connectRobot(robot.id);
        buttonsDiv.appendChild(button);
      });

      const rescanButton = document.createElement("button");
      rescanButton.innerHTML = "Re-scan";
      rescanButton.onclick = () => {
        updateAuthel("Reiniting BLE then scanning...");
        reInitBLE().then(() => scanRobots(false));
      };

      updateAuthel("Click on the robot you would like to pair with.");
      authEl.appendChild(rescanButton);
      authEl.appendChild(buttonsDiv);
    });
}

function disconnect() {
  authEl.innerHTML = "Disconnecting...";
  OTAUpdating = false;
  fetch("/api-ble/stop_ota").then(() => fetch("/api-ble/disconnect").then(() => checkBLECapability()));
}

function connectRobot(id) {
  updateAuthel("Connecting to robot...");
  fetch(`/api-ble/connect?id=${id}`)
    .then((response) => response.text())
    .then((response) => {
      if (response.includes("success")) {
        createPinEntry();
      } else {
        alert("Error connecting. WirePod will restart and this will return to the first screen of setup.");
        updateAuthel("Waiting for WirePod to restart...");
        setTimeout(checkBLECapability, 3000);
      }
    });
}

function createPinEntry() {
  authEl.innerHTML = `
    <p>Enter the pin shown on Vector's screen.</p>
    <input type="text" id="pinEntry" placeholder="Enter PIN here" maxlength="6">
    <br>
    <button onclick="sendPin()">Send PIN</button>
  `;
}

function sendPin() {
  const pin = document.getElementById("pinEntry").value;
  updateAuthel("Sending PIN...");
  fetch(`/api-ble/send_pin?pin=${pin}`)
    .then((response) => response.text())
    .then((response) => {
      if (response.includes("incorrect pin") || response.includes("length of pin")) {
        updateAuthel("Wrong PIN... Reiniting BLE then scanning...");
        reInitBLE().then(() => scanRobots(true));
      } else {
        wifiCheck();
      }
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
    });
}

function scanWifi() {
  authEl.innerHTML = "Scanning for Wi-Fi networks...";
  fetch("/api-ble/scan_wifi")
    .then((response) => response.json())
    .then((networks) => {
      authEl.innerHTML = `
        <p>Select a Wi-Fi network to connect Vector to.</p>
        <button onclick="scanWifi()">Scan Again</button>
        <br>
        ${networks
          .map(
            (network) =>
              network.ssid && `<button onclick="createWiFiPassEntry('${network.ssid}', '${network.authtype}')">${network.ssid}</button>`
          )
          .join("")}
      `;
    });
}

function createWiFiPassEntry(ssid, authtype) {
  authEl.innerHTML = `
    <button onclick="scanWifi()">Scan Again</button>
    <p>Enter the password for ${ssid}</p>
    <input type="text" id="passEntry" placeholder="Password">
    <br>
    <button onclick="connectWifi('${ssid}', '${authtype}')">Connect to Wi-Fi</button>
  `;
}

function connectWifi(ssid, authtype) {
  const password = document.getElementById("passEntry").value;
  authEl.innerHTML = "Connecting Vector to Wi-Fi...";
  fetch(`/api-ble/connect_wifi?ssid=${ssid}&password=${password}&authType=${authtype}`)
    .then((response) => response.text())
    .then((response) => {
      if (!response.includes("255")) {
        alert("Error connecting, likely incorrect password");
        createWiFiPassEntry(ssid, authtype);
      } else {
        whatToDo();
      }
    });
}

function checkFirmware() {
  fetch("/api-ble/get_firmware")
    .then((response) => response.text())
    .then((response) => {
      const splitFirmware = response.split("-");
      console.log(splitFirmware);
    });
}

function whatToDo() {
  fetch("/api-ble/get_robot_status")
    .then((response) => response.text())
    .then((response) => {
      switch (response) {
        case "in_recovery_prod":
          doOTA("local");
          break;
        case "in_recovery_dev":
          doOTA("http://wpsetup.keriganc.com:81/1.6.0.3331.ota");
          break;
        case "in_firmware_nonep":
          showRecoveryInstructions();
          break;
        case "in_firmware_dev":
          showDevWarning();
          break;
        case "in_firmware_ep":
          showAuthButton();
          break;
      }
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

function showDevWarning() {
  alert("Your bot is a dev robot. Make sure you have done the 'Configure an OSKR/dev-unlocked robot' section before authentication. If you did already, you can ignore this warning.");
  showAuthButton();
}

function showAuthButton() {
  authEl.innerHTML = `<button onclick="doAuth()">AUTHENTICATE</button>`;
}

function doOTA(url) {
  updateAuthel("Starting OTA update...");
  fetch(`/api-ble/start_ota?url=${url}`)
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
              } else if (otaResponse.includes("stopped") || !OTAUpdating) {
                clearInterval(interval);
              }
            });
        }, 2000);
      } else {
        whatToDo();
      }
    });
}

function updateAuthel(update) {
  authEl.innerHTML = `<p>${update}</p>`;
}

function doAuth() {
  updateAuthel("Authenticating your Vector...");
  fetch("/api-ble/do_auth")
    .then((response) => response.text())
    .then((response) => {
      if (response.includes("error")) {
        showAuthError();
      } else {
        showWakeOptions();
      }
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

function showWakeOptions() {
  updateAuthel("Authentication was successful! How would you like to wake Vector up?");
  authEl.innerHTML += `
    <button onclick="doOnboard(true)">Wake with wake-up animation (recommended)</button>
    <br>
    <button onclick="doOnboard(false)">Wake immediately, without wake-up animation</button>
  `;
}

function doOnboard(withAnim) {
  updateAuthel("Onboarding robot...");
  fetch(`/api-ble/onboard?with_anim=${withAnim}`).then(() => {
    fetch("/api-ble/disconnect");
    updateAuthel("Vector is now fully set up! Use the Bot Settings tab to further configure your bot.");
    const disconnectButtonDiv = document.getElementById("disconnectButton");
    disconnectButtonDiv.innerHTML = `
      <button onclick="checkBLECapability()">Return to pair instructions</button>
    `;
  });
}
