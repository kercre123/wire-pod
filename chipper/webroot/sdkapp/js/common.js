// The purpose of this file is to provide common functions that are used by multiple pages (home, bot settings, etc).


async function getSDKInfo() {
  try {
    var response = await fetch("/api-sdk/get_sdk_info");
    console.log(response)
    if (!response.ok) {
      return undefined;
    }
    var data = await response.json();
    return data;
  } catch (error) {
    console.error('Unable to get SDK info:', error);
    throw error;
  }
}


function escapeHTML(text) {
  var div = document.createElement("div");
  div.textContent = text;
  return div.innerHTML;
}

// Nickname if one is set, otherwise the serial number.
function robotDisplayName(robot) {
  return robot["nickname"] ? robot["nickname"] : robot["esn"];
}

async function setRobotNickname(serial, nickname) {
  var response = await fetch("/api/set_nickname", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ esn: serial, nickname: nickname }),
  });
  if (!response.ok) {
    throw new Error(await response.text());
  }
}


async function getBatteryStatus(serial) {
  if (!serial) {
    return 'Serial number is required';
  }

  try {
    var response = await fetch("/api-sdk/get_battery?serial=" + serial, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      signal: AbortSignal.timeout(15000)
    });
    if (!response.ok) {
      return undefined;
    }
    var data = await response.json(); // {"status":{"code":1},"battery_level":3,"battery_volts":3.9210937,"is_on_charger_platform":true}
    return data;
  } catch (error) {
    console.error('Unable to get battery status:', error);
    throw error;
  }
}