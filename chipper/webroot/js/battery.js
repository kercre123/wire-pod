var botStats = document.getElementById("botStats");

function getBatteryPercentage(voltage) {
  let percentage;
  const maxVoltage = 4.1; // Maximum voltage for the battery
  const midVoltage = 3.85; // Mid voltage for the battery
  const minVoltage = 3.5; // Minimum voltage for the battery

  if (voltage >= maxVoltage) {
      percentage = 100; // Fully charged
  } else if (voltage >= midVoltage) {
      // Fast drop from 100% to 80%
      let scaledVoltage = (voltage - midVoltage) / (maxVoltage - midVoltage);
      percentage = 80 + 20 * Math.log10(1 + scaledVoltage * 9); // Adjust factor to make the curve steeper
  } else if (voltage >= minVoltage) {
      // Gradual drop off from 80% to 0%
      let scaledVoltage = (voltage - minVoltage) / (midVoltage - minVoltage);
      percentage = 80 * Math.log10(1 + scaledVoltage * 9); // Adjust factor for the curve
  } else if (!voltage) {
      // if the bot is turned on whilst off the charger, the get_battery response doesn't include voltage. 
      // assume a reasonable battery percentage
      percentage = 70;
  } else {
      percentage = 0; // At or below 3.5V, considered empty
  }

  return Math.max(0, Math.min(100, Math.round(percentage))); // Ensure percentage is within 0-100%
}


async function updateBatteryInfo(serial, i) {
  var batteryContainer = document.getElementsByClassName("batteryContainer")[i];
  if (!batteryContainer) {
    return;
  }
  var batteryOutline = batteryContainer.getElementsByClassName("batteryOutline")[0];
  var batteryLevel = batteryOutline.getElementsByClassName("batteryLevel")[0];
  var chargeTimeRemaining = batteryOutline.getElementsByClassName("chargeTimeRemaining")[0];
  var vectorFace = batteryContainer.getElementsByClassName("vectorFace")[0];
  var tooltip = batteryContainer.getElementsByClassName("tooltip")[0];

  if (!batteryLevel || !vectorFace) {
    return;
  }

  let batteryStatus;

  try {
  // Maintain the battery information for each robot but fetch the latest battery status and update the battery level
    batteryStatus = await getBatteryStatus(serial);
  } catch {
    // Do nothing
  }
  if (!batteryStatus) {
    if (charging) {
      charging.remove();
    }
    batteryLevel.className = "batteryLevel batteryUnknown";
    vectorFace.style.backgroundImage = "url(/assets/wififace.gif)";
    tooltip.innerHTML = `<b>${serial}</b><br/>??%<br/> (Unable to connect)`;
    setTimeout(async () => {
      // Re-render the battery information
      updateBatteryInfo(serial, i);
    }, 6000);
    return;
  }

  let batteryPercentage = getBatteryPercentage(batteryStatus["battery_volts"]);
  if (batteryStatus["battery_level"] === 2) {
    // If the battery level is 2, we'll update the colors to reflect the battery level
    if (batteryPercentage < 20) {
      batteryStatus["battery_level"] = 0;
    } else if (batteryPercentage < 50) {
      batteryStatus["battery_level"] = 1;
    }
  } else if (batteryStatus["battery_level"] === 1 && !batteryStatus["is_on_charger_platform"]) {
    // Cap the battery level at 15% if the battery level is 1 and not charging
    vectorFace.style.backgroundImage = "url(/assets/homeface.png)";
    batteryPercentage = Math.min(15, batteryPercentage);
    batteryStatus["battery_level"] = 0; // Set color to red
  }


  // Set the battery level based on the battery_level value and handle the rest in css
  const batteryLevelClass = "batteryLevel battery" + batteryStatus["battery_level"];
  if (batteryLevel.className !== batteryLevelClass) {
    batteryLevel.className = batteryLevelClass;
  }
  
  // Update the battery level
  batteryLevel.style.width = batteryPercentage + "%";

  // Clear tooltip, and replace serial number and the latest voltage
  tooltip.innerHTML = `<b>${serial}</b><br/>~${batteryPercentage}%<br/> (${batteryStatus["battery_volts"].toFixed(2)}V)`;

  // Update the charging status
  if (batteryStatus["is_on_charger_platform"]) {
    if (!batteryOutline.getElementsByClassName("charging").length) {
      var charging = document.createElement("div");
      charging.className = "charging";
      batteryOutline.appendChild(charging);
      vectorFace.style.backgroundImage = "url(/assets/expandface.gif)";
    }

    chargeTimeRemaining.style.display = "block";
    if (batteryStatus["suggested_charger_sec"]) {
      chargeTimeRemaining.innerHTML = `~${Math.round(batteryStatus["suggested_charger_sec"])}s`;
    } else if (batteryStatus["is_charging"]) {
      chargeTimeRemaining.innerHTML = "";
    }else {
      chargeTimeRemaining.innerHTML = "Full";
      // assume 100% if Full
      batteryLevel.style.width = "100%";
      vectorFace.style.backgroundImage = "url(/assets/face.gif)";
    }
  } else {
    var charging = batteryOutline.getElementsByClassName("charging")[0];
    if (charging) {
      charging.remove();
      vectorFace.style.backgroundImage = "url(/assets/facegaze.gif)";
    }
    chargeTimeRemaining.style.display = "none";
  }

  

  setTimeout(async () => {
    // Re-render the battery information
    updateBatteryInfo(serial, i);
  }, 3000);
}

async function renderBatteryInfo(serial, i = 0) {
  // For each robot, we'll create a new div to hold the battery information with a class of "batteryContainer"
  var batteryContainer = document.createElement("div");
  batteryContainer.className = "batteryContainer";
  botStats.appendChild(batteryContainer);
  batteryContainer.onclick = function() {
    window.location.href = "/sdkapp/settings.html?serial=" + serial;
  };

  // Create a tooltip for the robot's serial number, with class "tooltip"

  var tooltip = document.createElement("span");
  tooltip.className = "tooltip";
  tooltip.innerHTML = `<b>${serial}</b>`;
  batteryContainer.appendChild(tooltip);

  // Create a new div to hold the battery status with a class of "batteryOutline", this will be the outline of the battery status
  var batteryOutline = document.createElement("div");
  batteryOutline.className = "batteryOutline";
  batteryContainer.appendChild(batteryOutline);

  var vectorFace = document.createElement("div");
  vectorFace.className = "vectorFace";
  vectorFace.style.backgroundImage = "url(/assets/webface.gif)"; // default loading face
  batteryContainer.appendChild(vectorFace);

  // Create the colored div that will represent the battery level, with a class of "batteryLevel"
  var batteryLevel = document.createElement("div");
  batteryLevel.className = "batteryLevel";
  batteryOutline.appendChild(batteryLevel);

  var chargeTimeRemaining = document.createElement("div");
  chargeTimeRemaining.className = "chargeTimeRemaining";
  batteryOutline.appendChild(chargeTimeRemaining);

  // We will manage the battery level via class names, there are only 4 levels reported (0, 1, 2, 3)

  // Get the battery status for the robot
  let batteryStatus;

  try {
  // Maintain the battery information for each robot but fetch the latest battery status and update the battery level
    batteryStatus = await getBatteryStatus(serial);
  } catch {
    // Do nothing
  }
  if (!batteryStatus) {
    batteryLevel.className = "batteryLevel batteryUnknown";
    vectorFace.style.backgroundImage = "url(/assets/wififace.gif)";
    tooltip.innerHTML = `<b>${serial}</b><br/>??<br/> (Unable to connect)`;
    setTimeout(async () => {
      // Re-render the battery information
      updateBatteryInfo(serial, i);
    }, 6000);
    return;
  }

  // If the robot is on the charger platform, we'll set the battery level to "charging" by adding a child div with class "charging"
  if (batteryStatus["is_on_charger_platform"]) {
    var charging = document.createElement("div");
    charging.className = "charging";
    batteryOutline.appendChild(charging);
    vectorFace.style.backgroundImage = "url(/assets/expandface.gif)";
  } else {
    vectorFace.style.backgroundImage = "url(/assets/facegaze.gif)";
  }

  // Set the battery level based on the battery_level value and handle the rest in css
  batteryLevel.className = "batteryLevel battery" + batteryStatus["battery_level"];
  const batteryPercentage = getBatteryPercentage(batteryStatus["battery_volts"]);
  batteryLevel.style.width = batteryPercentage + "%";

  tooltip.innerHTML += `<br/>~${batteryPercentage}%`;

  // Add the battery voltage to the tooltip
  tooltip.innerHTML += `<br/> (${batteryStatus["battery_volts"].toFixed(2)}V)`;

  setTimeout(async () => {
    // Re-render the battery information
    updateBatteryInfo(serial, i);
  }, 3000);
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
      botStats.style.display = "none";
      return;
    }

    for (var i = 0; i < sdkInfo["robots"].length; i++) {
      const serial = sdkInfo["robots"][i]["esn"];

      renderBatteryInfo(serial, i);
    }
  } catch {
    // Do nothing
  }
}