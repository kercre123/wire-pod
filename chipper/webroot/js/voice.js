// "Voice Services" section of the setup page: external Wyoming speech-to-text
// (e.g. wyoming-faster-whisper) and text-to-speech (e.g. wyoming-piper).

const WYOMING_DEFAULT_STT_PORT = 10300;
const WYOMING_DEFAULT_TTS_PORT = 10200;

// voices returned by the last "Test Connection & Load Voices"
let wyomingVoices = [];
// STT language from "Set Language", used to filter voices (e.g. "pl")
let wyomingLanguage = "";

function showVoice() {
  toggleVisibility(["section-weather", "section-restart", "section-language", "section-kg", "section-voice"], "section-voice", "icon-Voice");
  updateVoiceSettings();
}

function checkVoice() {
  getE("wyomingSttInput").style.display = getE("wyomingSttEnable").checked ? "block" : "none";
  getE("wyomingTtsInput").style.display = getE("wyomingTtsEnable").checked ? "block" : "none";
}

function updateVoiceSettings() {
  fetch("/api/get_stt_info")
    .then((response) => response.json())
    .then((stt) => {
      wyomingLanguage = (stt.language || "").split("-")[0].toLowerCase();
      if (!getE("wyomingTtsPreviewText").value) {
        getE("wyomingTtsPreviewText").value = wyomingLanguage === "pl"
          ? "Cześć, jestem Vector. Jak mogę pomóc?"
          : "Hi, I'm Vector. How can I help?";
      }
    });
  fetch("/api/get_wyoming")
    .then((response) => response.json())
    .then((cfg) => {
      getE("wyomingSttEnable").checked = cfg.stt.enable;
      getE("wyomingSttHost").value = cfg.stt.host || "";
      getE("wyomingSttPort").value = cfg.stt.port || WYOMING_DEFAULT_STT_PORT;
      getE("wyomingTtsEnable").checked = cfg.tts.enable;
      getE("wyomingTtsHost").value = cfg.tts.host || "";
      getE("wyomingTtsPort").value = cfg.tts.port || WYOMING_DEFAULT_TTS_PORT;
      getE("wyomingTtsVoice").value = cfg.tts.voice || "";
      getE("wyomingTtsVolume").value = cfg.tts.volume || 100;
      getE("wyomingTtsVolumeValue").textContent = getE("wyomingTtsVolume").value + "%";
      checkVoice();
    });
}

function wyomingServer(kind) {
  const prefix = kind === "stt" ? "wyomingStt" : "wyomingTts";
  return {
    host: getE(prefix + "Host").value.trim(),
    port: parseInt(getE(prefix + "Port").value, 10) || 0,
  };
}

function testWyoming(kind) {
  const statusId = kind === "stt" ? "wyomingSttStatus" : "wyomingTtsStatus";
  const server = wyomingServer(kind);
  if (!server.host || !server.port) {
    displayMessage(statusId, "Enter host and port first.");
    return;
  }
  displayMessage(statusId, "Connecting...");
  fetch(`/api/wyoming_describe?host=${encodeURIComponent(server.host)}&port=${server.port}`)
    .then((response) => {
      if (!response.ok) {
        return response.text().then((text) => { throw new Error(text); });
      }
      return response.json();
    })
    .then((info) => {
      if (kind === "stt") {
        const asr = info.asr || [];
        if (asr.length === 0) {
          displayMessage(statusId, "Connected, but this server offers no speech-to-text.");
          return;
        }
        const models = asr.flatMap((p) => (p.models || []).map((m) => m.name));
        displayMessage(statusId, `Connected: ${asr.map((p) => p.name + " " + (p.version || "")).join(", ")} - models: ${models.join(", ")}`);
      } else {
        const tts = info.tts || [];
        if (tts.length === 0) {
          displayMessage(statusId, "Connected, but this server offers no text-to-speech.");
          return;
        }
        wyomingVoices = tts.flatMap((p) => (p.voices || []).filter((v) => v.installed !== false));
        fillVoiceList();
        const own = voicesForLanguage().length;
        displayMessage(statusId, `Connected: ${tts.map((p) => p.name + " " + (p.version || "")).join(", ")} - ${wyomingVoices.length} voices` +
          (wyomingLanguage ? `, ${own} for language "${wyomingLanguage}"` : "") + ". Pick one in the Voice field.");
      }
    })
    .catch((err) => displayMessage(statusId, err.message));
}

function voicesForLanguage() {
  if (!wyomingLanguage) {
    return wyomingVoices;
  }
  return wyomingVoices.filter((v) => (v.languages || []).some((l) => l.toLowerCase().startsWith(wyomingLanguage)));
}

function fillVoiceList() {
  const list = getE("wyomingTtsVoiceList");
  list.innerHTML = "";
  const voices = getE("wyomingTtsAllVoices").checked ? wyomingVoices : voicesForLanguage();
  voices.forEach((v) => {
    const option = document.createElement("option");
    option.value = v.name;
    option.label = (v.languages || []).join(", ");
    list.appendChild(option);
  });
}

function previewWyomingVoice() {
  const server = wyomingServer("tts");
  const text = getE("wyomingTtsPreviewText").value.trim();
  if (!server.host || !server.port || !text) {
    displayMessage("wyomingTtsPreviewStatus", "Enter host, port and a test sentence first.");
    return;
  }
  displayMessage("wyomingTtsPreviewStatus", "Synthesizing...");
  const params = new URLSearchParams({
    host: server.host,
    port: server.port,
    voice: getE("wyomingTtsVoice").value.trim(),
    text: text,
  });
  fetch("/api/wyoming_tts_preview?" + params.toString())
    .then((response) => {
      if (!response.ok) {
        return response.text().then((t) => { throw new Error(t); });
      }
      return response.blob();
    })
    .then((blob) => {
      const audio = getE("wyomingTtsAudio");
      audio.src = URL.createObjectURL(blob);
      audio.style.display = "block";
      audio.volume = (parseInt(getE("wyomingTtsVolume").value, 10) || 100) / 100;
      audio.play();
      displayMessage("wyomingTtsPreviewStatus", "This is how the robot will sound (16 kHz mono).");
    })
    .catch((err) => displayMessage("wyomingTtsPreviewStatus", err.message));
}

function sendVoiceSettings() {
  const stt = wyomingServer("stt");
  const tts = wyomingServer("tts");
  const data = {
    stt: {
      enable: getE("wyomingSttEnable").checked,
      host: stt.host,
      port: stt.port,
    },
    tts: {
      enable: getE("wyomingTtsEnable").checked,
      host: tts.host,
      port: tts.port,
      voice: getE("wyomingTtsVoice").value.trim(),
      volume: parseInt(getE("wyomingTtsVolume").value, 10) || 100,
    },
  };
  displayMessage("voiceSettingsStatus", "Saving...");
  fetch("/api/set_wyoming", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  })
    .then((response) => response.text())
    .then((response) => displayMessage("voiceSettingsStatus", response));
}
