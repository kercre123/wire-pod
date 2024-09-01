var botStats = document.getElementById("botStats");


async function renderBatteryInfo(sdkInfo) {
  for (var i = 0; i < sdkInfo["robots"].length; i++) {
    // For each robot, we'll create a new div to hold the battery information with a class of "batteryContainer"
    var batteryContainer = document.createElement("div");
    batteryContainer.className = "batteryContainer";
    botStats.appendChild(batteryContainer);

    // Create a tooltip for the robot's serial number, with class "tooltip"

    var tooltip = document.createElement("span");
    tooltip.className = "tooltip";
    tooltip.innerHTML = `<b>${sdkInfo["robots"][i]["esn"]}</b>`;
    batteryContainer.appendChild(tooltip);

    // Create a new div to hold the battery status with a class of "batteryOutline", this will be the outline of the battery status
    var batteryOutline = document.createElement("div");
    batteryOutline.className = "batteryOutline";
    batteryContainer.appendChild(batteryOutline);

    var vectorFace = document.createElement("div");
    vectorFace.className = "vectorFace";
    vectorFace.style.backgroundImage = "url(/assets/webface.gif)"; // default loading face
    batteryOutline.appendChild(vectorFace);

    // Create the colored div that will represent the battery level, with a class of "batteryLevel"
    var batteryLevel = document.createElement("div");
    batteryLevel.className = "batteryLevel";
    batteryOutline.appendChild(batteryLevel);

    // We will manage the battery level via class names, there are only 4 levels reported (0, 1, 2, 3)

    // Get the battery status for the robot
    const batteryStatus = await getBatteryStatus(sdkInfo["robots"][i]["esn"]); // {"status":{"code":1},"battery_level":3,"battery_volts":3.9210937,"is_on_charger_platform":true}

    if (!batteryStatus) {
      batteryLevel.className = "batteryLevel batteryUnknown";
      vectorFace.style.backgroundImage = "url(/assets/wififace.gif)";
      continue;
    }
    // Add the battery voltage to the tooltip
    tooltip.innerHTML += `<br/> (${batteryStatus["battery_volts"].toFixed(2)}V)`;

    // If the robot is on the charger platform, we'll set the battery level to "charging" by adding a child div with class "charging"
    if (batteryStatus["is_on_charger_platform"]) {
      var charging = document.createElement("div");
      charging.className = "charging";
      batteryOutline.appendChild(charging);
      vectorFace.style.backgroundImage = "url(/assets/expandface.gif)";
    } else {
      vectorFace.style.backgroundImage = "url(/assets/face.gif)";
    }

    // Set the battery level based on the battery_level value and handle the rest in css
    batteryLevel.className = "batteryLevel battery" + batteryStatus["battery_level"];
  }
}

async function updateBatteryInfo(sdkInfo) {
  // Maintain the battery information for each robot but fetch the latest battery status and update the battery level
  for (var i = 0; i < sdkInfo["robots"].length; i++) {
    const batteryStatus = await getBatteryStatus(sdkInfo["robots"][i]["esn"]);
    if (!batteryStatus) {
      continue;
    }

    var batteryContainer = document.getElementsByClassName("batteryContainer")[i];
    var batteryOutline = batteryContainer.getElementsByClassName("batteryOutline")[0];
    var batteryLevel = batteryOutline.getElementsByClassName("batteryLevel")[0];
    var vectorFace = batteryOutline.getElementsByClassName("vectorFace")[0];
    var tooltip = batteryContainer.getElementsByClassName("tooltip")[0];

    if (!batteryLevel || !vectorFace) {
      continue;
    }

    // Clear tooltip, and replace serial number and the latest voltage
    tooltip.innerHTML = `<b>${sdkInfo["robots"][i]["esn"]}</b><br/> (${batteryStatus["battery_volts"].toFixed(2)}V)`;

    // Set the battery level based on the battery_level value and handle the rest in css
    const batteryLevelClass = "batteryLevel battery" + batteryStatus["battery_level"];
    if (batteryLevel.className !== batteryLevelClass) {
      batteryLevel.className = batteryLevelClass;
    }

    // Update the charging status
    if (batteryStatus["is_on_charger_platform"]) {
      if (!batteryOutline.getElementsByClassName("charging").length) {
        var charging = document.createElement("div");
        charging.className = "charging";
        batteryOutline.appendChild(charging);
        vectorFace.style.backgroundImage = "url(/assets/expandface.gif)";
      }
    } else {
      var charging = batteryOutline.getElementsByClassName("charging")[0];
      if (charging) {
        charging.remove();
        vectorFace.style.backgroundImage = "url(/assets/face.gif)";
      }
    }

  }
}

async function processBotStats() {
  try {
    // While loading, set a loading gif in a class div of "botLoader" to the botStats div
    var botLoader = document.createElement("div");
    botLoader.className = "botLoader";
    botStats.appendChild(botLoader);

    const sdkInfo = await getSDKInfo(); //{"global_guid":"tni1TRsTRTaNSapjo0Y+Sw==","robots":[{"esn":"00603f9b","ip_address":"10.42.0.248","guid":"5RlowyehhT8Qq7wEpF6JsQ==","activated":true},{"esn":"004047ef","ip_address":"10.42.0.175","guid":"ofoJZqLP3cwd9YpvXrdAfw==","activated":true}]}
    
    botLoader.remove();
    if (!sdkInfo) {
      return;
    }

    renderBatteryInfo(sdkInfo);

    setInterval(async () => {
      // Clear the botStats div to remove all battery information
      botStats.innerHTML = "";
      // Re-render the battery information
      updateBatteryInfo(sdkInfo);
    }, 3000);
  } catch {
    // Do nothing
  }
}

processBotStats();