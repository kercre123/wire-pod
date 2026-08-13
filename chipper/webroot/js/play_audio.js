// Play Audio: file pick or high-quality mic record → 8 kHz mono WAV → /api-sdk/play_sound
// Recording captures raw PCM (not MediaRecorder/webm) for cleaner speech to Vector.

let processedAudioBlob = null;

// Recording state (raw PCM path)
let recStream = null;
let recContext = null;
let recSource = null;
let recProcessor = null;
let recMute = null;
let recChunks = null; // array of Float32Array at context sampleRate
let recSampleRate = 0;
let recActive = false;

const TARGET_RATE = 8000;

function getSerial() {
  try {
    if (typeof urlParams !== "undefined" && urlParams && urlParams.get) {
      return urlParams.get("serial") || "";
    }
  } catch (_) {}
  return new URLSearchParams(window.location.search).get("serial") || "";
}

function setStatus(msg) {
  const el = document.getElementById("audioStatus");
  if (el) el.textContent = msg || "";
}

function setString(view, offset, string) {
  for (let i = 0; i < string.length; i++) {
    view.setUint8(offset + i, string.charCodeAt(i));
  }
}

/** High-quality downsample: low-pass average box then linear interpolate */
function downsampleHQ(input, fromRate, toRate) {
  if (fromRate === toRate) return new Float32Array(input);
  // Simple anti-alias: average groups when downsampling a lot
  const ratio = fromRate / toRate;
  const newLen = Math.max(1, Math.round(input.length / ratio));
  const output = new Float32Array(newLen);

  if (ratio >= 2) {
    // Multi-tap average kernel ~ ratio samples
    const kernel = Math.max(2, Math.floor(ratio));
    for (let i = 0; i < newLen; i++) {
      const center = i * ratio;
      const start = Math.max(0, Math.floor(center - kernel / 2));
      const end = Math.min(input.length - 1, Math.ceil(center + kernel / 2));
      let sum = 0;
      let n = 0;
      for (let j = start; j <= end; j++) {
        sum += input[j];
        n++;
      }
      output[i] = n ? sum / n : 0;
    }
  } else {
    for (let i = 0; i < newLen; i++) {
      const pos = i * ratio;
      const i0 = Math.floor(pos);
      const i1 = Math.min(i0 + 1, input.length - 1);
      const f = pos - i0;
      output[i] = input[i0] * (1 - f) + input[i1] * f;
    }
  }
  return output;
}

/** Soft peak normalize toward target peak (avoid clipping / too quiet) */
function normalizeSpeech(samples, targetPeak) {
  targetPeak = targetPeak || 0.85;
  let peak = 0;
  for (let i = 0; i < samples.length; i++) {
    const a = Math.abs(samples[i]);
    if (a > peak) peak = a;
  }
  if (peak < 0.001 || peak > 0.99) {
    // silence or already hot — only scale if quiet
    if (peak >= 0.001 && peak < 0.2) {
      const g = Math.min(targetPeak / peak, 4.0);
      for (let i = 0; i < samples.length; i++) samples[i] *= g;
    }
    return samples;
  }
  if (peak < targetPeak) {
    const g = Math.min(targetPeak / peak, 3.5);
    for (let i = 0; i < samples.length; i++) samples[i] *= g;
  }
  return samples;
}

function applyShortFades(samples, fadeMs, rate) {
  const n = Math.min(Math.floor((fadeMs / 1000) * rate), Math.floor(samples.length / 10));
  for (let i = 0; i < n; i++) {
    const g = i / n;
    samples[i] *= g;
    samples[samples.length - 1 - i] *= g;
  }
}

function floatTo16BitWavBlob(float32, sampleRate) {
  const numSamples = float32.length;
  const buffer = new ArrayBuffer(44 + numSamples * 2);
  const view = new DataView(buffer);
  let offset = 0;
  setString(view, offset, "RIFF"); offset += 4;
  view.setUint32(offset, 36 + numSamples * 2, true); offset += 4;
  setString(view, offset, "WAVE"); offset += 4;
  setString(view, offset, "fmt "); offset += 4;
  view.setUint32(offset, 16, true); offset += 4;
  view.setUint16(offset, 1, true); offset += 2; // PCM
  view.setUint16(offset, 1, true); offset += 2; // mono
  view.setUint32(offset, sampleRate, true); offset += 4;
  view.setUint32(offset, sampleRate * 2, true); offset += 4;
  view.setUint16(offset, 2, true); offset += 2;
  view.setUint16(offset, 16, true); offset += 2;
  setString(view, offset, "data"); offset += 4;
  view.setUint32(offset, numSamples * 2, true); offset += 4;
  for (let i = 0; i < numSamples; i++) {
    const s = Math.max(-1, Math.min(1, float32[i]));
    view.setInt16(offset, s < 0 ? s * 0x8000 : s * 0x7fff, true);
    offset += 2;
  }
  return new Blob([buffer], { type: "audio/wav" });
}

function mergeFloatChunks(chunks) {
  let total = 0;
  for (const c of chunks) total += c.length;
  const out = new Float32Array(total);
  let o = 0;
  for (const c of chunks) {
    out.set(c, o);
    o += c.length;
  }
  return out;
}

function toMonoFromAudioBuffer(audioBuffer) {
  const len = audioBuffer.length;
  const mono = new Float32Array(len);
  if (audioBuffer.numberOfChannels === 1) {
    mono.set(audioBuffer.getChannelData(0));
    return mono;
  }
  const ch0 = audioBuffer.getChannelData(0);
  const ch1 = audioBuffer.getChannelData(1);
  for (let i = 0; i < len; i++) {
    mono[i] = 0.5 * (ch0[i] + ch1[i]);
  }
  return mono;
}

/** Full pipeline → Vector-ready 8 kHz mono 16-bit WAV blob */
function pcmToVectorWav(floatMono, fromRate) {
  let samples = downsampleHQ(floatMono, fromRate, TARGET_RATE);
  samples = normalizeSpeech(samples, 0.82);
  applyShortFades(samples, 12, TARGET_RATE);
  return floatTo16BitWavBlob(samples, TARGET_RATE);
}

async function arrayBufferToWirePodWav(arrayBuffer) {
  const audioContext = new (window.AudioContext || window.webkitAudioContext)();
  try {
    const audioBuffer = await audioContext.decodeAudioData(arrayBuffer.slice(0));
    const mono = toMonoFromAudioBuffer(audioBuffer);
    return pcmToVectorWav(mono, audioBuffer.sampleRate);
  } finally {
    if (audioContext.close) {
      try { await audioContext.close(); } catch (_) {}
    }
  }
}

function previewBlob(blob) {
  processedAudioBlob = blob;
  const url = URL.createObjectURL(blob);
  const audioOutput = document.getElementById("audioOutput");
  if (audioOutput) {
    audioOutput.src = url;
  }
  const btn = document.getElementById("uploadButton");
  if (btn) btn.style.display = "inline-block";
}

async function uploadAudio(blob, { silent } = {}) {
  if (!blob) {
    if (!silent) alert("No audio to send.");
    return false;
  }
  const esn = getSerial();
  if (!esn) {
    if (!silent) alert("No robot serial in URL. Open Vector control from the bot list.");
    return false;
  }
  const formData = new FormData();
  formData.append("sound", blob, "processed.wav");
  try {
    const response = await fetch("/api-sdk/play_sound?serial=" + encodeURIComponent(esn), {
      method: "POST",
      body: formData,
    });
    if (!response.ok) {
      throw new Error("HTTP " + response.status + " " + response.statusText);
    }
    try { await response.json(); } catch (_) {}
    if (!silent) setStatus("Sent to Vector");
    return true;
  } catch (error) {
    console.error("Error sending audio:", error);
    if (!silent) setStatus("Send failed: " + error.message);
    return false;
  }
}

// --- File ---
const fileInputEl = document.getElementById("fileInput");
if (fileInputEl) {
  fileInputEl.addEventListener("change", async () => {
    const fileInput = document.getElementById("fileInput");
    if (!fileInput.files.length) {
      alert("Please select an audio file");
      return;
    }
    setStatus("Processing file…");
    try {
      const arrayBuffer = await fileInput.files[0].arrayBuffer();
      const blob = await arrayBufferToWirePodWav(arrayBuffer);
      previewBlob(blob);
      setStatus("Ready — press Send when ready");
    } catch (error) {
      console.error(error);
      setStatus("Could not process file");
      alert("Error processing the file: " + error.message);
    }
  });
}

const uploadBtn = document.getElementById("uploadButton");
if (uploadBtn) {
  uploadBtn.addEventListener("click", async () => {
    setStatus("Sending…");
    await uploadAudio(processedAudioBlob);
  });
}

// --- High-quality Record (raw PCM, full session, then convert + auto-send) ---

async function startRecording() {
  if (recActive) return;
  if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
    alert("Microphone not available in this browser");
    return;
  }
  if (!getSerial()) {
    alert("No robot serial in URL. Open Vector control from the bot list.");
    return;
  }
  try {
    recStream = await navigator.mediaDevices.getUserMedia({
      audio: {
        channelCount: 1,
        echoCancellation: true,
        noiseSuppression: true,
        autoGainControl: true,
        // Prefer highest rate the device allows; we downsample carefully later
        sampleRate: { ideal: 48000 },
      },
    });

    recContext = new (window.AudioContext || window.webkitAudioContext)({
      sampleRate: 48000, // request 48k if browser allows
      latencyHint: "interactive",
    });
    if (recContext.state === "suspended") {
      await recContext.resume();
    }
    recSampleRate = recContext.sampleRate;
    recSource = recContext.createMediaStreamSource(recStream);
    // Larger buffer = fewer callbacks, less chance of dropouts while recording
    const bufferSize = 4096;
    recProcessor = recContext.createScriptProcessor(bufferSize, 1, 1);
    recChunks = [];
    recActive = true;

    recProcessor.onaudioprocess = (e) => {
      if (!recActive) return;
      const input = e.inputBuffer.getChannelData(0);
      // copy — input buffer is reused
      recChunks.push(new Float32Array(input));
    };

    recSource.connect(recProcessor);
    recMute = recContext.createGain();
    recMute.gain.value = 0;
    recProcessor.connect(recMute);
    recMute.connect(recContext.destination);

    document.getElementById("recordButton").style.display = "none";
    document.getElementById("stopRecordButton").style.display = "inline-block";
    setStatus(
      "Recording… (" + recSampleRate + " Hz capture). Speak clearly, then Stop recording, then Send."
    );
  } catch (err) {
    console.error(err);
    alert("Could not open microphone: " + err.message);
    setStatus("Mic permission denied or unavailable");
    cleanupRecGraph();
  }
}

function cleanupRecGraph() {
  recActive = false;
  try {
    if (recProcessor) {
      recProcessor.onaudioprocess = null;
      recProcessor.disconnect();
    }
  } catch (_) {}
  try { if (recSource) recSource.disconnect(); } catch (_) {}
  try { if (recMute) recMute.disconnect(); } catch (_) {}
  try {
    if (recStream) recStream.getTracks().forEach((t) => t.stop());
  } catch (_) {}
  try { if (recContext) recContext.close(); } catch (_) {}
  recProcessor = null;
  recSource = null;
  recMute = null;
  recStream = null;
  recContext = null;
}

async function stopRecording() {
  if (!recActive && (!recChunks || !recChunks.length)) {
    cleanupRecGraph();
    document.getElementById("recordButton").style.display = "inline-block";
    document.getElementById("stopRecordButton").style.display = "none";
    return;
  }
  recActive = false;
  setStatus("Processing recording…");

  // Small delay so last audio buffers flush
  await new Promise((r) => setTimeout(r, 80));

  const rate = recSampleRate || 48000;
  const mono = mergeFloatChunks(recChunks || []);
  cleanupRecGraph();
  recChunks = null;

  document.getElementById("recordButton").style.display = "inline-block";
  document.getElementById("stopRecordButton").style.display = "none";

  if (!mono.length) {
    setStatus("No audio captured");
    return;
  }

  try {
    const blob = pcmToVectorWav(mono, rate);
    // Load into player for optional manual listen — do NOT auto-play
    previewBlob(blob);
    const sec = (mono.length / rate).toFixed(1);
    setStatus(
      "Recorded " + sec + "s @ " + rate + " Hz → 8 kHz. Press Send when ready."
    );
  } catch (err) {
    console.error(err);
    setStatus("Process failed: " + err.message);
  }
}

const recordBtn = document.getElementById("recordButton");
const stopRecordBtn = document.getElementById("stopRecordButton");
if (recordBtn) recordBtn.addEventListener("click", startRecording);
if (stopRecordBtn) stopRecordBtn.addEventListener("click", stopRecording);

window.addEventListener("beforeunload", () => {
  try { cleanupRecGraph(); } catch (_) {}
});
