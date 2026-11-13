// Vector plays 8 kHz mono 16-bit WAV. This converts any audio file the browser can decode
// (same method as the classic UI) and returns a WAV Blob.

export async function toVectorWav(file) {
  const Ctx = window.AudioContext || window.webkitAudioContext;
  const ctx = new Ctx();
  try {
    const decoded = await ctx.decodeAudioData(await file.arrayBuffer());
    // mix down to mono
    const len = decoded.length;
    const mono = new Float32Array(len);
    for (let c = 0; c < decoded.numberOfChannels; c++) {
      const ch = decoded.getChannelData(c);
      for (let i = 0; i < len; i++) mono[i] += ch[i] / decoded.numberOfChannels;
    }
    // resample to 8000 Hz (linear)
    const rate = 8000;
    const newLen = Math.round((len * rate) / decoded.sampleRate);
    const out = new Float32Array(newLen);
    for (let i = 0; i < newLen; i++) {
      const pos = (i * decoded.sampleRate) / rate;
      const i0 = Math.floor(pos);
      const i1 = Math.min(i0 + 1, len - 1);
      const f = pos - i0;
      out[i] = mono[i0] * (1 - f) + mono[i1] * f;
    }
    return encodeWav(out, rate);
  } finally {
    if (ctx.close) ctx.close();
  }
}

function encodeWav(samples, rate) {
  const buffer = new ArrayBuffer(44 + samples.length * 2);
  const v = new DataView(buffer);
  const str = (o, s) => { for (let i = 0; i < s.length; i++) v.setUint8(o + i, s.charCodeAt(i)); };
  str(0, "RIFF"); v.setUint32(4, 36 + samples.length * 2, true); str(8, "WAVE"); str(12, "fmt ");
  v.setUint32(16, 16, true); v.setUint16(20, 1, true); v.setUint16(22, 1, true);
  v.setUint32(24, rate, true); v.setUint32(28, rate * 2, true); v.setUint16(32, 2, true); v.setUint16(34, 16, true);
  str(36, "data"); v.setUint32(40, samples.length * 2, true);
  for (let i = 0; i < samples.length; i++) {
    const s = Math.max(-1, Math.min(1, samples[i]));
    v.setInt16(44 + i * 2, s < 0 ? s * 0x8000 : s * 0x7fff, true);
  }
  return new Blob([buffer], { type: "audio/wav" });
}
