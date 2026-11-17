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

// in-page replacement for window.confirm. resolves true on confirm, false on cancel
function confirmDialog(message, confirmText, cancelText) {
  return new Promise((resolve) => {
    const previousFocus = document.activeElement;

    const overlay = document.createElement("div");
    overlay.className = "dialogOverlay";

    const box = document.createElement("div");
    box.className = "dialogBox";
    box.setAttribute("role", "alertdialog");
    box.setAttribute("aria-modal", "true");

    const msg = document.createElement("p");
    msg.className = "dialogMessage";
    msg.textContent = message;

    const buttons = document.createElement("div");
    buttons.className = "dialogButtons";

    const cancelBtn = document.createElement("button");
    cancelBtn.type = "button";
    cancelBtn.className = "dialogBtn";
    cancelBtn.textContent = cancelText || "Cancel";

    const confirmBtn = document.createElement("button");
    confirmBtn.type = "button";
    confirmBtn.className = "dialogBtn dialogBtnDanger";
    confirmBtn.textContent = confirmText || "Confirm";

    buttons.appendChild(cancelBtn);
    buttons.appendChild(confirmBtn);
    box.appendChild(msg);
    box.appendChild(buttons);
    overlay.appendChild(box);
    document.body.appendChild(overlay);
    cancelBtn.focus();

    function close(result) {
      document.removeEventListener("keydown", onKey, true);
      overlay.remove();
      if (previousFocus && previousFocus.focus) {
        previousFocus.focus();
      }
      resolve(result);
    }
    function onKey(e) {
      if (e.key === "Escape") {
        e.preventDefault();
        close(false);
      } else if (e.key === "Tab") {
        // keep focus inside the dialog
        e.preventDefault();
        (document.activeElement === cancelBtn ? confirmBtn : cancelBtn).focus();
      }
    }
    document.addEventListener("keydown", onKey, true);
    cancelBtn.onclick = () => close(false);
    confirmBtn.onclick = () => close(true);
    overlay.onclick = (e) => {
      if (e.target === overlay) {
        close(false);
      }
    };
  });
}

function closeBotMenus() {
  document.querySelectorAll(".botMenuButtons.open").forEach((m) => m.classList.remove("open"));
}
// click anywhere else, or press Escape, to close any open bot menu
document.addEventListener("click", closeBotMenus);
document.addEventListener("keydown", (e) => {
  if (e.key === "Escape") {
    closeBotMenus();
  }
});

// icon-only button for the bot menu
function makeMenuAction(className, title, icon, onAction) {
  const btn = document.createElement("div");
  btn.className = "botMenuAction " + className;
  btn.title = title;
  btn.innerHTML = '<i class="fa-solid ' + icon + '"></i>';
  btn.onclick = function(e) {
    e.stopPropagation();
    // actions need a reachable bot. the chip's "active" class is kept up to date by updateBotChip
    const chip = btn.closest(".botChip");
    if (chip && !chip.classList.contains("active")) {
      return;
    }
    closeBotMenus();
    onAction();
  };
  return btn;
}

// same call as sendForm('/api-sdk/cloud_intent?intent=...') on the settings page,
// but with this chip's serial instead of the page-level esn
function sendBotIntent(esn, intent) {
  fetch("/api-sdk/cloud_intent?intent=" + encodeURIComponent(intent) + "&serial=" + encodeURIComponent(esn), {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" }
  }).then((response) => response.text()).then((data) => {
    if (data.startsWith("error:")) {
      alert("Unable to send " + intent + ": " + data.replace(/^error:\s*/, ""));
    }
  }).catch((error) => {
    console.error("Unable to send " + intent + ":", error);
    alert("Unable to send " + intent + ": " + error.message);
  });
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

  const menuButton = document.createElement("div");
  menuButton.className = "botMenuButton";
  menuButton.title = "Menu";
  menuButton.innerHTML = '<i class="fa-solid fa-ellipsis"></i>';
  chip.appendChild(menuButton);

  const menuButtons = document.createElement("div");
  menuButtons.className = "botMenuButtons";
  // clicks inside the menu must not open the bot's settings page
  menuButtons.onclick = function(e) {
    e.stopPropagation();
  };
  chip.appendChild(menuButtons);

  menuButton.onclick = function(e) {
    e.stopPropagation();
    const wasOpen = menuButtons.classList.contains("open");
    closeBotMenus();
    if (!wasOpen) {
      menuButtons.classList.add("open");
    }
  };

  menuButtons.appendChild(makeMenuAction("botSleepBtn", "Go to sleep", "fa-moon", function() {
    sendBotIntent(bot.esn, "intent_system_sleep");
  }));
  menuButtons.appendChild(makeMenuAction("botHomeBtn", "Go home", "fa-house", function() {
    sendBotIntent(bot.esn, "intent_system_charger");
  }));

  menuButtons.appendChild(makeMenuAction("botExploreBtn", "Explore", "fa-compass", function() {
    sendBotIntent(bot.esn, "explore_start");
  }));
  menuButtons.appendChild(makeMenuAction("botDanceBtn", "Dance", "fa-music", function() {
    sendBotIntent(bot.esn, "intent_imperative_dance");
  }));
  menuButtons.appendChild(makeMenuAction("botFetchCubeBtn", "Fetch cube", "fa-cube", function() {
    sendBotIntent(bot.esn, "intent_imperative_fetchcube");
  }));

  const botDeleteBtn = document.createElement("div");
  botDeleteBtn.className = 'botDeleteBtn';
  botDeleteBtn.title = "Delete";
  botDeleteBtn.innerHTML = '<i class="fa-solid fa-trash-can"></i>';
  botDeleteBtn.onclick = function(e) {
    e.stopPropagation();
    closeBotMenus();
    confirmDialog("Are you sure you want to delete " + bot.esn + "?\nYou will need to re-activate to be added back", "Delete", "Cancel").then((confirmed) => {
      if (!confirmed) {
        return;
      }
      return fetch("/delete_bot?serial=" + encodeURIComponent(bot.esn), {
        method: "POST"
      }).then((response) => {
        if (!response.ok) {
          throw new Error("bad response: " + response.status);
        }
        return response.text();
      }).then((data) => {
        if (data.trim() === "success") {
          chip.remove();
          delete botChips[bot.esn];
        } else {
          alert("Failed to delete bot: " + data.replace(/^error:\s*/, ""));
        }
      }).catch((error) => {
        console.error("Unable to delete bot:", error);
        alert("Unable to delete bot: " + error.message);
      });
    });
  }
  menuButtons.appendChild(botDeleteBtn);

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
