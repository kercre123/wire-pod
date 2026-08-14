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
let recMaxTimer = null;
let recCapturedSamples = 0;
let recContextClosing = null;
let recStopping = false;

const TARGET_RATE = 8000;
const MAX_REC_SECONDS = 30;
const FADE_MS = 40;

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

/** Soft peak normalize toward target peak (avoid clipping / too quiet).
 *  Skip gain when RMS/peak sit at the noise floor so hiss is not boosted. */
function normalizeSpeech(samples, targetPeak) {
  targetPeak = targetPeak || 0.82;
  let peak = 0;
  let sumSq = 0;
  for (let i = 0; i < samples.length; i++) {
    const s = samples[i];
    const a = Math.abs(s);
    if (a > peak) peak = a;
    sumSq += s * s;
  }
  const rms = samples.length ? Math.sqrt(sumSq / samples.length) : 0;
  const NOISE_FLOOR_RMS = 0.015;
  const NOISE_FLOOR_PEAK = 0.04;
  const MAX_GAIN = 2.5;
  if (peak < NOISE_FLOOR_PEAK || rms < NOISE_FLOOR_RMS || peak >= targetPeak) {
    return samples;
  }
  const g = Math.min(targetPeak / peak, MAX_GAIN);
  if (g > 1.02) {
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
  applyShortFades(samples, FADE_MS, TARGET_RATE);
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

function micErrorMessage(err) {
  const name = (err && err.name) || "";
  if (name === "NotAllowedError" || name === "PermissionDeniedError") {
    return "Microphone permission was denied. Allow mic access for this site and try again.";
  }
  if (name === "NotFoundError" || name === "DevicesNotFoundError") {
    return "No microphone was found. Plug in a mic and try again.";
  }
  if (name === "NotReadableError" || name === "TrackStartError") {
    return "Microphone is already in use by another application.";
  }
  if (name === "OverconstrainedError" || name === "ConstraintNotSatisfiedError") {
    return "Microphone does not support the requested settings.";
  }
  if (name === "SecurityError") {
    return "Microphone requires a secure (HTTPS) connection.";
  }
  if (name === "AbortError") {
    return "Microphone request was interrupted. Try again.";
  }
  return "Could not open microphone: " + ((err && (err.message || err.name)) || "unknown error");
}

function clearRecMaxTimer() {
  if (recMaxTimer) {
    clearTimeout(recMaxTimer);
    recMaxTimer = null;
  }
}

function recSecondsSoFar() {
  const rate = recSampleRate || 48000;
  return recCapturedSamples / rate;
}

async function startRecording() {
  if (recActive || recStopping) return;
  if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
    alert("Microphone not available in this browser");
    return;
  }
  if (!getSerial()) {
    alert("No robot serial in URL. Open Vector control from the bot list.");
    return;
  }
  if (recContextClosing) {
    try { await recContextClosing; } catch (_) {}
    recContextClosing = null;
  }
  try {
    recStream = await navigator.mediaDevices.getUserMedia({
      audio: {
        channelCount: 1,
        echoCancellation: true,
        noiseSuppression: true,
        autoGainControl: true,
        sampleRate: { ideal: 48000 },
      },
    });

    // Let the browser pick the hardware rate; requested 48 kHz is only a hint
    // on getUserMedia. Read sampleRate only after the context is running.
    recContext = new (window.AudioContext || window.webkitAudioContext)({
      latencyHint: "interactive",
    });
    if (recContext.state === "suspended") {
      await recContext.resume();
    }
    recSampleRate = recContext.sampleRate;
    recSource = recContext.createMediaStreamSource(recStream);
    const bufferSize = 4096;
    recProcessor = recContext.createScriptProcessor(bufferSize, 1, 1);
    recChunks = [];
    recCapturedSamples = 0;
    recActive = true;

    recProcessor.onaudioprocess = (e) => {
      if (!recActive) return;
      const buf = e.inputBuffer;
      if (buf && buf.sampleRate) recSampleRate = buf.sampleRate;
      const input = buf.getChannelData(0);
      recChunks.push(new Float32Array(input));
      recCapturedSamples += input.length;
      if (recSecondsSoFar() >= MAX_REC_SECONDS) {
        recActive = false;
        setStatus("Reached " + MAX_REC_SECONDS + "s limit — processing…");
        stopRecording();
      }
    };

    recSource.connect(recProcessor);
    recMute = recContext.createGain();
    recMute.gain.value = 0;
    recProcessor.connect(recMute);
    recMute.connect(recContext.destination);

    clearRecMaxTimer();
    recMaxTimer = setTimeout(() => {
      if (recActive) {
        setStatus("Reached " + MAX_REC_SECONDS + "s limit — processing…");
        stopRecording();
      }
    }, MAX_REC_SECONDS * 1000);

    document.getElementById("recordButton").style.display = "none";
    document.getElementById("stopRecordButton").style.display = "inline-block";
    setStatus(
      "Recording… (" + recSampleRate + " Hz, max " + MAX_REC_SECONDS +
        "s). Speak clearly, then Stop recording, then Send."
    );
  } catch (err) {
    console.error(err);
    const msg = micErrorMessage(err);
    alert(msg);
    setStatus(msg);
    cleanupRecGraph();
  }
}

function closeRecContext(ctx) {
  if (!ctx || ctx.state === "closed") {
    return Promise.resolve();
  }
  try {
    const p = ctx.close();
    if (p && typeof p.then === "function") {
      return p.catch(() => {});
    }
  } catch (_) {}
  return Promise.resolve();
}

function cleanupRecGraph() {
  recActive = false;
  clearRecMaxTimer();
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
  const ctx = recContext;
  recProcessor = null;
  recSource = null;
  recMute = null;
  recStream = null;
  recContext = null;
  if (ctx && ctx.state !== "closed") {
    recContextClosing = closeRecContext(ctx);
  }
}

async function stopRecording() {
  if (recStopping) return;
  recStopping = true;
  if (!recActive && (!recChunks || !recChunks.length)) {
    cleanupRecGraph();
    recStopping = false;
    document.getElementById("recordButton").style.display = "inline-block";
    document.getElementById("stopRecordButton").style.display = "none";
    return;
  }
  recActive = false;
  clearRecMaxTimer();
  setStatus("Processing recording…");

  // Small delay so last audio buffers flush
  await new Promise((r) => setTimeout(r, 80));

  // Prefer the running context rate (settled after resume / first buffers)
  const rate = (recContext && recContext.sampleRate) || recSampleRate || 48000;
  recSampleRate = rate;
  const mono = mergeFloatChunks(recChunks || []);
  cleanupRecGraph();
  recChunks = null;
  recCapturedSamples = 0;

  document.getElementById("recordButton").style.display = "inline-block";
  document.getElementById("stopRecordButton").style.display = "none";

  try {
    if (!mono.length) {
      setStatus("No audio captured");
      return;
    }
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
  } finally {
    recStopping = false;
  }
}

const recordBtn = document.getElementById("recordButton");
const stopRecordBtn = document.getElementById("stopRecordButton");
if (recordBtn) recordBtn.addEventListener("click", startRecording);
if (stopRecordBtn) stopRecordBtn.addEventListener("click", stopRecording);

window.addEventListener("beforeunload", () => {
  try { cleanupRecGraph(); } catch (_) {}
});
