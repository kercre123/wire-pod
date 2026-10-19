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
  checkProxyMode();
}

function toggleSections(showSection, icon) {
  const sections = ["section-intents", "section-log", "section-botauth", "section-version", "section-uicustomizer"];
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
  fetch("/api-ble/disconnect").then(() => fetch("/api-ble/init").catch(() => {
    showExternalSetupInstructions();
  })).catch(() => fetch("/api-ble/init"));
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
    }).catch(() => {
      updateAuthel("Error scanning. Reiniting BLE then scanning...");
      reInitBLE().then(() => scanRobots(false));
    });
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
        alert("Error connecting. WirePod will restart and this will return to the first screen of setup.");
        updateAuthel("Waiting for WirePod to restart...");
        setTimeout(checkBLECapability, 3000);
      }
    }).catch(() => {
      alert("Error connecting. WirePod will restart and this will return to the first screen of setup.");
      updateAuthel("Waiting for WirePod to restart...");
      setTimeout(checkBLECapability, 3000);
    })
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
    }).catch(() => {
      updateAuthel("Error scanning Wi-Fi networks... Reiniting BLE then scanning...");
      reInitBLE().then(() => scanRobots(true));
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
    }).catch(() => {
      updateAuthel("Error connecting to Wi-Fi... Reiniting BLE then scanning...");
      reInitBLE().then(() => scanRobots(true));
    });
}

function checkFirmware() {
  fetch("/api-ble/get_firmware")
    .then((response) => response.text())
    .then((response) => {
      const splitFirmware = response.split("-");
      console.log(splitFirmware);
    }).catch(() => {
      updateAuthel("Error getting firmware version... Reiniting BLE then scanning...");
      reInitBLE().then(() => scanRobots(true));
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

// --- Robot LAN addresses (proxy mode only) ---------------------------------
//
// Gated on server.proxy_mode via the existing /api/get_config, which already
// returns the whole APIConfig, so no extra endpoint is needed for the gate.
// A default install never sets the flag and never sees any of this.

function checkProxyMode() {
  fetch("/api/get_config")
    .then((response) => response.json())
    .then((config) => {
      const enabled = config && config.server && config.server.proxy_mode === true;
      document.getElementById("proxyRobotSection").style.display = enabled ? "block" : "none";
      if (enabled) {
        loadProxyRobots();
      }
    })
    .catch(() => {
      // A failed gate check must leave the section hidden rather than
      // exposing controls that would not apply.
      document.getElementById("proxyRobotSection").style.display = "none";
    });
}

// `note` carries a message from a preceding save. Without it the refresh this
// function performs would immediately wipe that confirmation off the screen.
function loadProxyRobots(note) {
  const list = document.getElementById("proxyRobotList");
  const status = document.getElementById("proxyRobotStatus");
  // Deliberately no "Loading..." placeholder when the table is already on
  // screen. Emptying the list first collapsed it to a single line and dropped
  // everything below it up the page, then put it back when the fetch landed
  // ~10ms later -- a jolt that reads as a full page reload. Build the
  // replacement off-DOM and swap it in one go, so the height never changes.
  if (!list.hasChildNodes()) {
    list.innerHTML = "<p>Loading robots...</p>";
  }

  fetch("/api/get_robots")
    .then((response) => response.json())
    .then((robots) => {
      if (!robots || robots.length === 0) {
        list.innerHTML = "<p>No robots known to wire-pod yet. Authenticate a robot first.</p>";
      } else {
        const parts = document.createDocumentFragment();
        parts.appendChild(buildProxyRobotTable(robots));

        // One Save under the whole table, centred like the rest of the page.
        const saveRow = document.createElement("div");
        const save = document.createElement("button");
        save.type = "button";
        save.innerHTML = "Save";
        save.onclick = saveProxyRobotAddresses;
        saveRow.appendChild(save);
        parts.appendChild(saveRow);

        list.replaceChildren(parts);
        prefillSSHIp(robots);
      }
      // The list is rebuilt identically when nothing has changed, and on a LAN
      // the fetch is too quick to see, so without a timestamp the button looks
      // dead and the serial looks hard-coded. Both are the same missing signal.
      const count = robots ? robots.length : 0;
      status.innerHTML =
        (note ? "<p>" + note + "</p>" : "") +
        "<small class=\"desc\">Checked " + new Date().toLocaleTimeString() +
        " — " + count + (count === 1 ? " robot" : " robots") +
        " known to wire-pod.</small>";
    })
    .catch((err) => {
      list.innerHTML = "";
      status.innerHTML = "<p>Error loading robots: " + err + "</p>";
    });
}

// The OSKR flow below asks for the same address this section configures. Fill
// it in from what is already known rather than asking twice.
//
// Only ever from configured_address: effective_target falls back to the
// observed peer address, which behind a proxy is the proxy itself -- prefilling
// that would point an SSH setup at the wrong host entirely. The field stays
// editable because a robot being set up for the first time is not yet known.
function prefillSSHIp(robots) {
  const sshIp = document.getElementById("sshIp");
  const hint = document.getElementById("sshIpHint");
  if (!sshIp || sshIp.value) {
    return;
  }
  const known = robots.find((robot) => robot.configured_address);
  if (!known) {
    return;
  }
  sshIp.value = known.configured_address;
  if (hint) {
    hint.innerHTML =
      "<small class=\"desc\">Filled in from the LAN address set above for " +
      known.esn + ". Change it if you are setting up a different robot.</small>";
  }
}

// One row per robot. A table because the interesting comparison is down the
// columns -- every robot is seen at the same proxy address, and that repetition
// is only visible when the values line up under each other.
function buildProxyRobotTable(robots) {
  const table = document.createElement("table");
  table.className = "table";

  const head = table.createTHead().insertRow();
  ["Robot", "Seen at", "Dialling", "LAN address"].forEach((text) => {
    const th = document.createElement("th");
    th.innerHTML = text;
    head.appendChild(th);
  });

  const body = table.createTBody();
  robots.forEach((robot) => body.appendChild(buildProxyRobotRow(robot)));
  return table;
}

// One line per robot. The proxy address is explained once in the section
// description, so it is not repeated under every row -- that sub-line was what
// made the rows double height.
function buildProxyRobotRow(robot) {
  const row = document.createElement("tr");

  const esn = row.insertCell();
  esn.innerHTML =
    "<b>" + robot.esn + "</b>" +
    (robot.activated ? "" : " <small class=\"desc\">(not authenticated)</small>");

  row.insertCell().innerHTML = robot.observed_address;
  row.insertCell().innerHTML = robot.effective_target;

  const input = document.createElement("input");
  input.type = "text";
  input.id = "lanAddr-" + robot.esn;
  input.placeholder = "blank to clear";
  input.value = robot.configured_address || "";
  // The original is kept on the element so Save can write only what changed,
  // rather than re-POSTing every row on every press.
  input.dataset.esn = robot.esn;
  input.dataset.original = robot.configured_address || "";
  input.onkeydown = (e) => {
    if (e.key === "Enter") {
      e.preventDefault();
      saveProxyRobotAddresses();
    }
  };
  row.insertCell().appendChild(input);

  return row;
}

// A single Save for the whole table, so several robots can be corrected in one
// pass. Only changed rows are sent.
//
// The API takes one robot per call, so this fans out and uses allSettled rather
// than all: a partial failure must still report which rows did save, or the
// operator cannot tell what state they are in.
function saveProxyRobotAddresses() {
  const status = document.getElementById("proxyRobotStatus");
  const inputs = Array.from(
    document.querySelectorAll("#proxyRobotList input[data-esn]")
  );
  const changed = inputs.filter((i) => i.value.trim() !== i.dataset.original);

  if (changed.length === 0) {
    status.innerHTML = "<p>No changes to save.</p>";
    return;
  }

  Promise.allSettled(
    changed.map((input) =>
      fetch("/api/set_robot_lan_address", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ esn: input.dataset.esn, address: input.value.trim() }),
      }).then(async (response) => {
        const text = (await response.text()).trim();
        if (!response.ok) {
          throw new Error(input.dataset.esn + ": " + text);
        }
        return input.dataset.esn;
      })
    )
  ).then((results) => {
    const saved = results.filter((r) => r.status === "fulfilled").map((r) => r.value);
    const failed = results.filter((r) => r.status === "rejected").map((r) => r.reason.message);
    let note = "";
    if (saved.length) {
      note += "Saved " + saved.join(", ") + ".";
    }
    if (failed.length) {
      note += (note ? " " : "") + "Failed — " + failed.join("; ");
    }
    // Handed to the refresh rather than set here, or the refresh that follows
    // would overwrite it before it could be read.
    loadProxyRobots(note);
  });
}
