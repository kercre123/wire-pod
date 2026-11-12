// Robot > Settings: things stored on the robot itself. Each control saves the moment you change it.
import { html, useState, useEffect, useRef } from "../html.js";
import { sdk, sdkJSON, describeError } from "../api.js";
import { robotName, toast, refreshRobots } from "../store.js";
import { Segmented, Loading, ErrorNote, OfflineNotice, Callout } from "../ui.js";
import { ACCENTS } from "../theme.js";

const VOLUME = ["Mute", "Low", "Medium low", "Medium", "Medium high", "High"];
const LOCALES = [["en-US", "English (US)"], ["en-GB", "English (UK)"], ["en-AU", "English (Australia)"], ["de-DE", "German"], ["fr-FR", "French"], ["ja-JP", "Japanese"]];
const TIME_ZONES = ["Pacific/Honolulu", "America/Juneau", "America/Los_Angeles", "America/Phoenix", "America/Denver", "America/Lima", "America/Chicago", "America/Bogota", "America/New_York", "America/Argentina/Buenos_Aires", "America/Santiago", "America/Sao_Paulo", "America/Halifax", "America/St_Johns", "GMT", "Europe/Lisbon", "Europe/London", "Europe/Paris", "Europe/Athens", "Europe/Istanbul", "Europe/Moscow", "Africa/Lagos", "Africa/Harare", "Africa/Addis_Ababa", "Asia/Dubai", "Asia/Tehran", "Asia/Karachi", "Asia/Kolkata", "Asia/Dhaka", "Asia/Bangkok", "Asia/Jakarta", "Asia/Hong_Kong", "Asia/Singapore", "Asia/Manila", "Asia/Seoul", "Asia/Tokyo", "Australia/Perth", "Australia/Darwin", "Australia/Adelaide", "Australia/Brisbane", "Australia/Sydney", "Australia/Auckland"];

export function RobotSettings({ robot }) {
  const name = robotName(robot);
  const esn = robot.esn;
  const [s, setS] = useState(null);
  const [error, setError] = useState(null);
  const [custom, setCustom] = useState(false);
  const [hue, setHue] = useState(180);
  const [sat, setSat] = useState(90);
  const [city, setCity] = useState("");
  const [volumeDraft, setVolumeDraft] = useState(null);
  const online = robot.state === "online";

  const load = async () => {
    setError(null);
    try {
      const v = await sdkJSON("get_sdk_settings", esn, {}, { timeout: 25000 });
      setS(v);
      setCity(v.default_location || "");
      if (v.custom_eye_color) { setHue(Math.round((v.custom_eye_color.hue || 0) * 360)); setSat(Math.round((v.custom_eye_color.saturation || 0) * 100)); }
    } catch (e) {
      setError(describeError(e, name));
    }
  };
  useEffect(() => { if (online) load(); }, [esn, online]);

  if (robot.state === "offline") return html`<${OfflineNotice} robot=${robot} onRetry=${refreshRobots} what="these settings" />`;
  if (!s) return error ? html`<${ErrorNote} error=${error} onRetry=${load} />` : html`<${Loading}>Reading ${name}’s settings…<//>`;

  // change(patch, send): show the change at once, send it, and put it back if the robot says no
  const change = async (patch, send, message = "Saved") => {
    const before = s;
    setS({ ...s, ...patch });
    try {
      await send();
      toast(message);
      if ("eye_color" in patch || "custom_eye_color" in patch) refreshRobots();
    } catch (e) {
      setS(before);
      toast(describeError(e, name).message, { kind: "bad" });
    }
  };

  const eyeIdx = s.custom_eye_color && s.custom_eye_color.enabled ? -1 : s.eye_color;
  const vol = volumeDraft !== null ? volumeDraft : s.master_volume;
  const locales = LOCALES.some(([c]) => c === s.locale) || !s.locale ? LOCALES : [...LOCALES, [s.locale, s.locale]];
  const zones = TIME_ZONES.includes(s.time_zone) || !s.time_zone ? TIME_ZONES : [...TIME_ZONES, s.time_zone];
  const customColor = `hsl(${hue} ${sat}% 50%)`;

  return html`<div class="stack">
    <p class="sub" style="font-size:.8125rem;margin-bottom:-6px">Changes here save instantly.</p>
    <div class="row2" style="align-items:start">
      <div class="stack">
        <div class="card">
          <h2 class="h3">Sound & look</h2><p class="sub">How ${name} sounds and what its eyes look like.</p>
          <label for="rs-vol">Volume: <span>${VOLUME[vol] || vol}</span></label>
          <input class="range" id="rs-vol" type="range" min="0" max="5" step="1" value=${vol} aria-valuetext=${VOLUME[vol]}
            onInput=${(e) => setVolumeDraft(+e.target.value)}
            onChange=${(e) => { const v = +e.target.value; setVolumeDraft(null); change({ master_volume: v }, () => sdk("volume", esn, { volume: v })); }} />
          <div class="rl" aria-hidden="true"><span>Mute</span><span>Medium</span><span>High</span></div>
          <span class="lbl" style="margin-top:14px" id="rs-eyes">Eye color</span>
          <div class="swatches" role="group" aria-labelledby="rs-eyes">
            ${ACCENTS.map((a, i) => html`<button type="button" key=${a.id} style=${`--c:${a.swatch}`} aria-label=${a.name} aria-pressed=${eyeIdx === i ? "true" : "false"}
              onClick=${() => { setCustom(false); change({ eye_color: i, custom_eye_color: { ...(s.custom_eye_color || {}), enabled: false } }, () => sdk("eye_color", esn, { color: i })); }}></button>`)}
            <button type="button" class="rainbow" aria-label="Custom color" aria-pressed=${eyeIdx === -1 || custom ? "true" : "false"} aria-expanded=${custom ? "true" : "false"} onClick=${() => setCustom(!custom)}></button>
          </div>
          ${custom && html`<div class="colorpick">
            <div><label for="rs-hue">Color</label><input class="hue" id="rs-hue" type="range" min="0" max="360" value=${hue} onInput=${(e) => setHue(+e.target.value)} /></div>
            <div><label for="rs-sat">Intensity</label><input class="range" id="rs-sat" type="range" min="0" max="100" value=${sat} onInput=${(e) => setSat(+e.target.value)} /></div>
            <div class="opt flush"><span class="face" style=${`--eye:${customColor};--fs:36px`} aria-hidden="true"><i></i><i></i></span>
              <button type="button" class="btn pri" onClick=${() => change({ custom_eye_color: { enabled: true, hue: hue / 360, saturation: sat / 100 } }, () => sdk("custom_eye_color", esn, { hue: (hue / 360).toFixed(3), sat: (sat / 100).toFixed(3) }))}>Use this color</button></div>
          </div>`}
        </div>
        <div class="card">
          <h2 class="h3">Button & Alexa</h2><p class="sub">What the button on ${name}’s back does.</p>
          <div class="opt"><div id="rs-btn">Button opens</div>
            <${Segmented} label="Button opens" options=${[[0, "Hey Vector"], [1, "Alexa"]]} value=${s.button_wakeword ? 1 : 0}
              onChange=${(v) => change({ button_wakeword: v }, () => sdk(v ? "button_alexa" : "button_hey_vector", esn))} /></div>
          <div class="opt"><div>Alexa<small>WirePod can’t tell whether ${name} is signed in.</small></div>
            <div class="chips">
              <button type="button" class="btn" onClick=${() => change({}, () => sdk("alexa_sign_in", esn), "Alexa sign-in started")}>Sign in</button>
              <button type="button" class="btn line" onClick=${() => change({}, () => sdk("alexa_sign_out", esn), "Signed out of Alexa")}>Sign out</button>
            </div></div>
        </div>
      </div>
      <div class="stack">
        <div class="card">
          <h2 class="h3">Language & time</h2><p class="sub">Used for answers, timers and the clock.</p>
          <div class="stack">
            <div><label for="rs-loc">Language</label>
              <select class="inp" id="rs-loc" value=${s.locale} onChange=${(e) => { const v = e.target.value; change({ locale: v }, () => sdk("locale", esn, { locale: v })); }}>
                ${locales.map(([c, l]) => html`<option value=${c} selected=${c === s.locale}>${l}</option>`)}</select></div>
            <div class="opt flush"><div>Clock</div><${Segmented} label="Clock" options=${[[false, "12 hour"], [true, "24 hour"]]} value=${!!s.clock_24_hour}
              onChange=${(v) => change({ clock_24_hour: v }, () => sdk(v ? "time_format_24" : "time_format_12", esn))} /></div>
            <div class="opt flush"><div>Units</div><${Segmented} label="Temperature units" options=${[[true, "°F"], [false, "°C"]]} value=${!!s.temp_is_fahrenheit}
              onChange=${(v) => change({ temp_is_fahrenheit: v }, () => sdk(v ? "temp_f" : "temp_c", esn))} /></div>
            <div><label for="rs-tz">Time zone</label>
              <select class="inp" id="rs-tz" onChange=${(e) => { const v = e.target.value; change({ time_zone: v }, () => sdk("timezone", esn, { timezone: v })); }}>
                ${zones.map((z) => html`<option value=${z} selected=${z === s.time_zone}>${z}</option>`)}</select></div>
            <form onSubmit=${(e) => { e.preventDefault(); const v = city.trim(); if (!v) { toast("Location can’t be blank.", { kind: "bad" }); return; } change({ default_location: v }, () => sdk("location", esn, { location: v })); }}>
              <label for="rs-city">Location for weather</label>
              <div class="with-btn"><input class="inp" id="rs-city" value=${city} onInput=${(e) => setCity(e.target.value)} placeholder="City, State" /><button type="submit" class="btn">Save</button></div>
            </form>
          </div>
        </div>
        <div class="card">
          <div class="top"><h2 class="h3">Voice & AI for ${name}</h2><span class="pill plan">Planned</span></div>
          <p class="sub">${name} uses the server’s AI and voice settings.</p>
          <button type="button" class="btn line" disabled>Use something different for ${name}</button>
        </div>
      </div>
    </div>
  </div>`;
}
