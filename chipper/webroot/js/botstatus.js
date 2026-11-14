var botStats = document.getElementById("botStats");
const botStatusPollMs = 5000;
const botChips = {};

const holidayLightGifs = [
  "faceholidaylights.gif",
  "faceholidaylights1.gif",
  "faceholidaylights2.gif",
  "faceholidaylights3.gif"
];
const holidayGif = holidayLightGifs[Math.floor(Math.random() * holidayLightGifs.length)];

function activeFace() {
  const now = new Date();
  const month = now.getMonth();
  const day = now.getDate();
  if ((month == 11 && day == 31) || (month == 0 && day == 1)) {
    return "facehnyfirework.gif";
  }
  if (month == 11 && (day == 24 || day == 25)) {
    return holidayGif;
  }
  if (month == 0 || month == 11) {
    return "facesnowglobe.gif";
  }
  return "face.gif";
}

function lastSeenText(secs) {
  if (secs < 0) {
    return "not seen yet";
  }
  if (secs < 60) {
    return "seen " + secs + "s ago";
  }
  if (secs < 3600) {
    return "seen " + Math.floor(secs / 60) + "m ago";
  }
  if (secs < 86400) {
    return "seen " + Math.floor(secs / 3600) + "h ago";
  }
  return "seen " + Math.floor(secs / 86400) + "d ago";
}

function makeBotChip(bot) {
  const chip = document.createElement("div");
  chip.className = "botChip";
  chip.onclick = function() {
    if (!chip.classList.contains("active")) {
      return;
    }
    window.location.href = "/sdkapp/settings.html?serial=" + encodeURIComponent(bot.esn);
  };

  const face = document.createElement("div");
  face.className = "vectorFace";
  chip.appendChild(face);

  const text = document.createElement("div");
  text.className = "botChipText";
  const name = document.createElement("span");
  name.className = "botChipName";
  const statusText = document.createElement("span");
  statusText.className = "botChipStatus";
  text.appendChild(name);
  text.appendChild(statusText);
  chip.appendChild(text);

  botStats.appendChild(chip);
  return { chip, face, name, statusText, face_url: "" };
}

function updateBotChip(c, bot) {
  c.name.textContent = bot.name || bot.esn;
  c.chip.title = "ESN: " + bot.esn + (bot.ip_address ? "\nIP: " + bot.ip_address : "");
  c.chip.classList.toggle("active", bot.active);
  c.statusText.textContent = bot.active ? "active" : lastSeenText(bot.seconds_since_seen);
  const faceUrl = "url(/assets/" + (bot.active ? activeFace() : "wififace.gif") + ")";
  if (c.face_url !== faceUrl) {
    c.face.style.backgroundImage = faceUrl;
    c.face_url = faceUrl;
  }
}

async function fetchActiveBots() {
  const response = await fetch("/api-sdk/get_active_bots", {
    signal: AbortSignal.timeout(botStatusPollMs)
  });
  if (!response.ok) {
    throw new Error("bad response: " + response.status);
  }
  return await response.json();
}

async function renderBotStatus(onlySerial) {
  if (!botStats) {
    return;
  }
  const loader = document.createElement("div");
  loader.className = "botLoader";
  botStats.appendChild(loader);

  async function poll() {
    try {
      let bots = await fetchActiveBots();
      if (onlySerial) {
        bots = bots.filter((b) => b.esn.toLowerCase() === onlySerial.toLowerCase());
      }
      loader.remove();
      if (bots.length === 0) {
        botStats.style.display = "none";
      } else {
        botStats.style.display = "";
      }
      const seen = new Set();
      for (const bot of bots) {
        seen.add(bot.esn);
        if (!botChips[bot.esn]) {
          botChips[bot.esn] = makeBotChip(bot);
        }
        updateBotChip(botChips[bot.esn], bot);
      }
      for (const esn of Object.keys(botChips)) {
        if (!seen.has(esn)) {
          botChips[esn].chip.remove();
          delete botChips[esn];
        }
      }
    } catch (error) {
      console.error("Unable to get bot status:", error);
    }
    setTimeout(poll, botStatusPollMs);
  }
  poll();
}

function processBotStats() {
  renderBotStatus();
}
