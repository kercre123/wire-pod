// First run (#/welcome) and Add a robot (#/add).
// Connection and language are done here. Bluetooth pairing is the existing, hardware-tested flow in the
// classic UI: this page sends you there, waits for the new robot to show up, then finishes with its name.
import { html, useState, useEffect, useRef } from "../html.js";
import { Icon } from "../icons.js";
import { href, go } from "../router.js";
import { getJSON, getText, postJSON, postForm, describeError } from "../api.js";
import { useRobots, robotName, toast, refreshRobots } from "../store.js";
import { loadConfig } from "../server.js";
import { Face, Callout, pageTitle } from "../ui.js";
import { STT_LANGUAGES, PERSONALITY_PRESETS } from "../data.js";

const PAIR_URL = "/classic#bot-setup";

function Logo() {
  return html`<span class="logo" aria-hidden="true"><i></i><i></i></span>`;
}

function Shell({ steps, step, children, title }) {
  useEffect(() => pageTitle(title), [title]);
  const n = step + 1;
  return html`<div class="wiz">
    <aside class="steps"><div class="brand"><${Logo} />WirePod</div>
      <ol aria-label="Setup steps">${steps.map((s, i) => html`<li key=${s} class=${i < step ? "done" : i === step ? "on" : ""} aria-current=${i === step ? "step" : null}>
        <span class="n">${i < step ? html`<${Icon} name="check" size=${16} />` : i + 1}</span>${s}</li>`)}</ol>
      <p class="sub" style="margin-top:auto;font-size:.78rem">${steps.length > 2 ? "Adding another robot later? You’ll start at pairing." : ""}</p>
      <a class="classic" href=${href.robots} style="font-size:.78rem">Leave setup</a>
    </aside>
    <div class="wmain">
      <header class="wtop"><div class="row"><span>Step ${n} of ${steps.length} · ${steps[step]}</span><span class="brand" style="font-size:1rem"><${Logo} />WirePod</span></div>
        <div class="prog" role="progressbar" aria-label="Setup progress" aria-valuenow=${n} aria-valuemin="1" aria-valuemax=${steps.length}><i style=${`width:${(n / steps.length) * 100}%`}></i></div></header>
      <main id="main" class="wbody">${children}</main>
    </div>
  </div>`;
}

function Head({ eyebrow, title, children }) {
  const ref = useRef(null);
  useEffect(() => { if (ref.current) ref.current.focus(); }, [title]);
  return html`<div><div class="eyebrow">${eyebrow}</div><h1 tabindex="-1" ref=${ref}>${title}</h1>${children && html`<p class="sub" style="margin-top:8px;max-width:54ch">${children}</p>`}</div>`;
}

function Foot({ onBack, backLabel = "Back", next, nextLabel, disabled, busy }) {
  return html`<div class="foot">
    ${onBack ? html`<button type="button" class="btn quiet" onClick=${onBack}>${backLabel}</button>` : html`<span></span>`}
    ${next && html`<button type="button" class="btn pri" onClick=${next} disabled=${disabled || busy}>${busy ? "Working…" : nextLabel}${!busy && html`<${Icon} name="right" size=${18} />`}</button>`}
  </div>`;
}

// ---------- step: connection ----------

function ConnectionStep({ conn, setConn, onNext }) {
  const opt = (v, icon, title, desc) => html`<button type="button" aria-pressed=${conn.mode === v ? "true" : "false"} onClick=${() => setConn({ ...conn, mode: v })}>
    <span class="ico"><${Icon} name=${icon} size=${22} /></span><span><b>${title}</b><small>${desc}</small></span></button>`;
  return html`<${Head} eyebrow="Welcome" title="Let’s get your robot talking">WirePod runs on this computer and gives your Vector a voice, a memory and a personality. First, how should robots reach this computer?<//>
    <div class="choice" role="group" aria-label="Connection method">
      ${opt("ep", "robots", "Escape pod", "Recommended for regular, production robots")}
      ${opt("ip", "wifi", "IP address", "Recommended for OSKR or developer-unlocked robots")}
    </div>
    ${conn.mode === "ip" && html`<div style="max-width:240px"><label for="w-port">Port</label><input class="inp" id="w-port" inputmode="numeric" placeholder="443" value=${conn.port} onInput=${(e) => setConn({ ...conn, port: e.target.value })} />
      <div class="hint"><span>Most setups use 443.</span></div></div>`}
    <${Foot} next=${onNext} nextLabel="Continue" />`;
}

// ---------- step: language (applies language and connection) ----------

function LanguageStep({ conn, onBack, onDone }) {
  const [lang, setLang] = useState("en-US");
  const [state, setState] = useState({ busy: false, text: "", error: "" });
  useEffect(() => { getJSON("/api/get_stt_info").then((s) => s.language && setLang(s.language), () => {}); }, []);

  const poll = async () => {
    for (let i = 0; i < 1800; i++) {
      await new Promise((r) => setTimeout(r, 600));
      const t = await getText("/api/get_download_status");
      if (t.includes("success")) return;
      if (t.includes("error")) throw new Error(t);
      setState({ busy: true, text: t.includes("not downloading") ? "Starting the download…" : t, error: "" });
    }
    throw new Error("The download is taking too long. Check your internet connection and try again.");
  };
  const run = async () => {
    const port = conn.port.trim() || "443";
    if (conn.mode === "ip" && !/^\d+$/.test(port)) { setState({ busy: false, text: "", error: "The port must be a number." }); return; }
    setState({ busy: true, text: "Setting your language…", error: "" });
    try {
      const t = await postJSON("/api/set_stt_info", { language: lang }, { timeout: 60000 });
      if (t.includes("downloading")) { setState({ busy: true, text: "Downloading the speech model…", error: "" }); await poll(); }
      setState({ busy: true, text: "Choosing how robots connect…", error: "" });
      const r = await getText(conn.mode === "ep" ? "/api-chipper/use_ep" : `/api-chipper/use_ip?port=${encodeURIComponent(port)}`, { timeout: 60000 });
      if (/^\s*error/i.test(r)) throw new Error(r.replace(/^\s*error:\s*/i, ""));
      await loadConfig();
      onDone();
    } catch (e) {
      setState({ busy: false, text: "", error: e.message && !e.kind ? e.message : describeError(e).message });
    }
  };
  return html`<${Head} eyebrow="Step 2" title="Which language will you speak?">WirePod downloads a speech model for it. You can change this later in Settings.<//>
    <div><label for="w-lang">Language</label>
      <select class="inp" id="w-lang" value=${lang} disabled=${state.busy} onChange=${(e) => setLang(e.target.value)}>${STT_LANGUAGES.map(([c, l]) => html`<option value=${c} selected=${c === lang}>${l}</option>`)}</select></div>
    <${Callout}><span>The speech model is about 50 MB. It downloads now, so this step can take a minute.</span><//>
    <div role="status" aria-live="polite">${state.busy && html`<div class="loading"><span class="spinner"></span><span>${state.text}</span></div>`}${state.error && html`<p class="err" role="alert">${state.error}</p>`}</div>
    <${Foot} onBack=${state.busy ? null : onBack} next=${run} nextLabel=${state.error ? "Try again" : "Continue"} busy=${state.busy} />`;
}

// ---------- step: pair ----------

function PairStep({ adding, found, onManual, onBack }) {
  const [clicked, setClicked] = useState(false);
  const [hint, setHint] = useState(false);
  return html`<${Head} eyebrow=${adding ? "Add a robot" : "Step 3"} title="Pair your robot">Pairing uses Bluetooth, so it happens in WirePod’s original setup page. It opens in a new tab. Come back here when it says Vector is set up.<//>
    <ol class="ordered">
      <li>Put Vector on his charger.</li>
      <li>Double-press his back button. A key appears on his screen.</li>
      <li>In the setup page, choose <b>Begin Scanning</b>, pick your Vector, then enter the code he shows.</li>
      <li>Follow the steps there for Wi-Fi and signing in. If it offers a software update, let it finish (about 5 minutes).</li>
    </ol>
    <div class="chips"><a class="btn pri" href=${PAIR_URL} target="_blank" rel="noopener" onClick=${() => setClicked(true)}><${Icon} name="bt" size=${18} />Open Bluetooth setup<${Icon} name="external" size=${16} /></a></div>
    <p class="sub" style="font-size:.8125rem">Have a developer (OSKR) robot? The same page has a section to set it up over SSH first.</p>
    <div role="status" aria-live="polite">
      ${found ? html`<${Callout}><span><b>Found ${robotName(found)}.</b> Pairing worked.</span><//>`
        : clicked ? html`<div class="loading"><span class="spinner"></span><span>Waiting for your robot to finish pairing…</span></div>` : null}
      ${hint && !found && html`<${Callout} icon="alert" kind="warn"><span>No new robot yet. If the setup page says Vector is set up, give it a few seconds, then try again.</span><//>`}
    </div>
    <div class="foot">
      <button type="button" class="btn quiet" onClick=${onBack}>${adding ? "Cancel" : "Skip for now"}</button>
      <button type="button" class="btn" onClick=${() => { if (found) onManual(); else { setHint(true); refreshRobots(); } }}>I’ve finished pairing<${Icon} name="right" size=${18} /></button>
    </div>`;
}

// ---------- step: name ----------

function NameStep({ robot, onDone }) {
  const [name, setName] = useState(robot.name || "");
  const [pers, setPers] = useState(robot.personality || "");
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState("");
  const finish = async (skip) => {
    if (!skip && (name.trim() || pers.trim())) {
      setBusy(true);
      try { await postForm("/api-sdk/set_robot_identity", { serial: robot.esn, name, personality: pers }); refreshRobots(); }
      catch (e) { setErr(describeError(e).message); setBusy(false); return; }
    }
    onDone();
  };
  return html`<${Head} eyebrow="Last step" title="Meet your robot">Give Vector a name and a personality. You can change both any time.<//>
    <div class="rhead" style="align-items:flex-start"><${Face} robot=${robot} size=${72} />
      <div class="who" style="gap:14px"><div><label for="w-name">Name</label><input class="inp" id="w-name" maxlength="32" value=${name} placeholder="Pip" onInput=${(e) => setName(e.target.value)} /></div></div></div>
    <div><label for="w-pers">Personality</label><textarea class="inp" id="w-pers" rows="3" maxlength="1000" value=${pers} placeholder="Example: Curious and upbeat. Loves a bad pun." onInput=${(e) => setPers(e.target.value)}></textarea></div>
    <div class="chips" role="group" aria-label="Start from">${PERSONALITY_PRESETS.map(([l, t]) => html`<button type="button" class="chip" key=${l} onClick=${() => setPers(t)}>${l}</button>`)}</div>
    ${err && html`<p class="err" role="alert">${err}</p>`}
    <div class="foot"><button type="button" class="btn quiet" onClick=${() => finish(true)} disabled=${busy}>Skip</button>
      <button type="button" class="btn pri" onClick=${() => finish(false)} disabled=${busy}>${busy ? "Saving…" : "Start talking"}</button></div>`;
}

// ---------- the flow ----------

export function Wizard({ first }) {
  const adding = !first;
  const steps = first ? ["Connection", "Language", "Pair your robot", "Name & personality"] : ["Pair your robot", "Name & personality"];
  const [step, setStep] = useState(0);
  const [conn, setConn] = useState({ mode: "ep", port: "" });
  const { loaded, robots } = useRobots();
  const before = useRef(null);
  if (before.current === null && loaded) before.current = new Set(robots.map((r) => r.esn));
  const found = before.current ? robots.find((r) => !before.current.has(r.esn)) : null;
  const pairIdx = first ? 2 : 0;
  const nameIdx = pairIdx + 1;

  // when a new robot appears while pairing, move on to naming it
  useEffect(() => { if (found && step === pairIdx) setStep(nameIdx); }, [found && found.esn, step]);

  const leave = () => go(href.robots);
  let body;
  const kind = steps[step];
  if (kind === "Connection") body = html`<${ConnectionStep} conn=${conn} setConn=${setConn} onNext=${() => setStep(1)} />`;
  else if (kind === "Language") body = html`<${LanguageStep} conn=${conn} onBack=${() => setStep(0)} onDone=${() => setStep(2)} />`;
  else if (kind === "Pair your robot") body = html`<${PairStep} adding=${adding} found=${found} onBack=${leave} onManual=${() => setStep(nameIdx)} />`;
  else body = found ? html`<${NameStep} robot=${found} onDone=${() => { toast(`${robotName(found)} is ready`); go(href.robot(found.esn)); }} />`
    : html`<${Head} eyebrow="Last step" title="Waiting for your robot">No new robot has shown up yet.<//><div class="foot"><button type="button" class="btn quiet" onClick=${() => setStep(pairIdx)}>Back</button><a class="btn" href=${href.robots}>Go to your robots</a></div>`;
  return html`<${Shell} steps=${steps} step=${step} title=${kind}>${body}<//>`;
}
