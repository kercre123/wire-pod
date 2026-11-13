// Robot > Controls: camera, drive, say something, play a sound, quick actions, mirror mode.
// Driving and speaking need "behavior control"; this screen takes it when you first need it and gives it
// back when you leave.
import { html, useState, useEffect, useRef } from "../html.js";
import { Icon } from "../icons.js";
import { sdk, postMultipart, describeError } from "../api.js";
import { robotName, toast, refreshRobots } from "../store.js";
import { Switch, HoldButton, OfflineNotice, Callout } from "../ui.js";
import { toVectorWav } from "../audio.js";
import { QUICK, runQuick } from "../quick.js";

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

// wheel speeds for a set of held keys (w a s d)
export function wheelsFor(k) {
  const f = k.w && !k.s, b = k.s && !k.w, l = k.a && !k.d, r = k.d && !k.a;
  if (f && l) return [100, 190];
  if (f && r) return [190, 100];
  if (b && l) return [-100, -190];
  if (b && r) return [-190, -100];
  if (f) return [140, 140];
  if (b) return [-150, -150];
  if (l) return [-150, 150];
  if (r) return [150, -150];
  return [0, 0];
}

export function Controls({ robot }) {
  const name = robotName(robot);
  const esn = robot.esn;
  const [controlled, setControlled] = useState(false);
  const [cam, setCam] = useState(false);
  const [camPhase, setCamPhase] = useState("off"); // off, loading, live, error, timeout
  const camImg = useRef(null);
  const [mirror, setMirror] = useState(false);
  const [kb, setKb] = useState(false);
  const [say, setSay] = useState("");
  const [sound, setSound] = useState(null); // { blob, url, label }
  const [soundBusy, setSoundBusy] = useState(false);
  const ctl = useRef({ controlled: false, mirror: false, cam: false });
  const alive = useRef(true); // false once this screen is left, so a late "take control" answer can hand control straight back
  useEffect(() => { alive.current = true; return () => { alive.current = false; }; }, []);
  const held = useRef(0); // bumps on every start and stop, so a late start can tell it was let go
  const online = robot.state === "online";

  const fail = (e) => { if (alive.current) toast(describeError(e, name).message, { kind: "bad" }); };

  const ensureControl = async () => {
    if (ctl.current.controlled) return;
    await sdk("assume_behavior_control", esn, { priority: "high" });
    if (!alive.current) { sdk("release_behavior_control", esn).catch(() => {}); throw new Error("left"); }
    ctl.current.controlled = true;
    setControlled(true);
    await sleep(1200); // the robot grants control a moment later
  };
  const release = async () => {
    const c = ctl.current;
    if (c.mirror) { c.mirror = false; setMirror(false); sdk("mirror_mode", esn, { enable: "false" }).catch(() => {}); }
    if (c.controlled) {
      c.controlled = false;
      setControlled(false);
      setKb(false);
      await sdk("release_behavior_control", esn); // awaited so a quick action can tell control is really back
    }
  };
  const stopCam = (phase = "off") => {
    if (ctl.current.cam) { ctl.current.cam = false; sdk("stop_cam_stream", esn).catch(() => {}); }
    setCam(false);
    setCamPhase(typeof phase === "string" ? phase : "off");
  };
  // the stream can take around ten seconds to wake up; give up after twenty rather than staying black
  useEffect(() => {
    if (camPhase !== "loading") return;
    const t = setTimeout(() => {
      const img = camImg.current;
      if (img && img.naturalWidth > 0) setCamPhase("live"); // frames are arriving even though no load event fired
      else stopCam("timeout");
    }, 20000);
    return () => clearTimeout(t);
  }, [camPhase]);

  // leaving this screen hands everything back
  useEffect(() => () => {
    const c = ctl.current;
    if (c.cam) sdk("stop_cam_stream", esn).catch(() => {});
    if (c.mirror) sdk("mirror_mode", esn, { enable: "false" }).catch(() => {});
    if (c.controlled) {
      sdk("move_wheels", esn, { lw: 0, rw: 0 }).catch(() => {});
      sdk("release_behavior_control", esn).catch(() => {});
    }
  }, [esn]);

  // the robot dropped off the network: nothing here can work, so reset
  useEffect(() => { if (!online) { ctl.current = { controlled: false, mirror: false, cam: false }; setControlled(false); setMirror(false); setCam(false); setCamPhase("off"); setKb(false); } }, [online]);

  const drive = {
    start: async (lw, rw) => {
      const id = ++held.current;
      try {
        await ensureControl();
        if (held.current === id) await sdk("move_wheels", esn, { lw, rw });
      } catch (e) { fail(e); }
    },
    stop: async () => {
      held.current++;
      try { await sdk("move_wheels", esn, { lw: 0, rw: 0 }); } catch (e) { /* the robot may already be gone */ }
    },
  };
  const moveAxis = (kind) => ({
    start: async (speed) => {
      const id = ++held.current;
      try {
        await ensureControl();
        if (held.current === id) await sdk(kind, esn, { speed });
      } catch (e) { fail(e); }
    },
    stop: async () => { held.current++; try { await sdk(kind, esn, { speed: 0 }); } catch (e) { /* ignore */ } },
  });
  const lift = moveAxis("move_lift");
  const head = moveAxis("move_head");

  // keyboard driving: w a s d wheels, r f lift, t g head
  useEffect(() => {
    if (!kb) return;
    const keys = {};
    let lastWheels = "0,0", lastLift = 0, lastHead = 0;
    const apply = () => {
      const w = wheelsFor(keys).join(",");
      if (w !== lastWheels) { lastWheels = w; const [lw, rw] = w.split(","); sdk("move_wheels", esn, { lw, rw }).catch(() => {}); }
      const l = keys.r && !keys.f ? 2 : keys.f && !keys.r ? -2 : 0;
      if (l !== lastLift) { lastLift = l; sdk("move_lift", esn, { speed: l }).catch(() => {}); }
      const h = keys.t && !keys.g ? 2 : keys.g && !keys.t ? -2 : 0;
      if (h !== lastHead) { lastHead = h; sdk("move_head", esn, { speed: h }).catch(() => {}); }
    };
    const typing = (e) => /^(INPUT|TEXTAREA|SELECT)$/.test(e.target.tagName) || e.target.isContentEditable;
    const down = (e) => { if (typing(e) || e.metaKey || e.ctrlKey || e.altKey) return; const k = e.key.toLowerCase(); if ("wasdrftg".includes(k) && k.length === 1) { e.preventDefault(); if (!keys[k]) { keys[k] = true; apply(); } } };
    const up = (e) => { const k = e.key.toLowerCase(); if (keys[k]) { keys[k] = false; apply(); } };
    const blur = () => { Object.keys(keys).forEach((k) => (keys[k] = false)); apply(); };
    document.addEventListener("keydown", down);
    document.addEventListener("keyup", up);
    window.addEventListener("blur", blur);
    return () => { document.removeEventListener("keydown", down); document.removeEventListener("keyup", up); window.removeEventListener("blur", blur); blur(); };
  }, [kb]);

  if (robot.state === "offline") return html`<${OfflineNotice} robot=${robot} onRetry=${refreshRobots} what="these controls" />`;

  const toggleControl = async (on) => {
    try {
      if (on) await ensureControl();
      else await release();
    } catch (e) { fail(e); }
  };
  const toggleMirror = async (on) => {
    try {
      if (on) await ensureControl();
      await sdk("mirror_mode", esn, { enable: on ? "true" : "false" });
      ctl.current.mirror = on;
      setMirror(on);
    } catch (e) { fail(e); }
  };
  const startCam = () => { ctl.current.cam = true; setCam(true); setCamPhase("loading"); };
  const doSay = async (e) => {
    e.preventDefault();
    const t = say.trim();
    if (!t) return;
    if ([...t].length >= 600) { toast("That’s too long to say. Keep it under 600 characters.", { kind: "bad" }); return; }
    try {
      await ensureControl();
      await sdk("say_text", esn, { text: t });
    } catch (err) { fail(err); }
  };
  const quick = (q) => runQuick(q, { esn, name, controlled: ctl.current.controlled, release, toast });
  const heyVector = async () => {
    try {
      await sdk("trigger_wake_word", esn);
      toast(`Said “Hey Vector” to ${name}`);
    } catch (e) {
      toast("Couldn’t trigger “Hey Vector”. It needs developer firmware or 2.0.1.6085ep or later.", { kind: "bad" });
    }
  };
  const chooseSound = async (e) => {
    const file = e.target.files && e.target.files[0];
    if (!file) return;
    setSoundBusy(true);
    try {
      const blob = await toVectorWav(file);
      if (sound) URL.revokeObjectURL(sound.url);
      setSound({ blob, url: URL.createObjectURL(blob), label: file.name });
    } catch (err) {
      toast("Couldn’t read that audio file. Try a WAV file.", { kind: "bad" });
    } finally { setSoundBusy(false); e.target.value = ""; }
  };
  const playSound = async () => {
    const fd = new FormData();
    fd.append("sound", sound.blob, "sound.wav");
    try {
      await postMultipart(`/api-sdk/play_sound?serial=${encodeURIComponent(esn)}`, fd, { timeout: 120000 });
      toast(`Sent to ${name}`);
    } catch (err) { fail(err); }
  };

  // a plain function, not a component, so re-rendering never remounts a button that is being held
  const pad = (label, icon, lw, rw) => html`<${HoldButton} label=${label} onStart=${() => drive.start(lw, rw)} onStop=${drive.stop} className=""><${Icon} name=${icon} size=${22} /><//>`;
  return html`<div class="grid2">
    <div class="stack">
      <div class="card">
        <div class="top"><h2 class="h3">Camera</h2>
          ${cam ? html`<button type="button" class="btn" onClick=${() => stopCam()}><${Icon} name="stop" size=${18} />Stop camera</button>`
                : html`<button type="button" class="btn" onClick=${startCam}><${Icon} name="camera" size=${18} />Start camera</button>`}</div>
        <div class="cam" style="margin-top:12px">
          ${cam ? html`<img ref=${camImg} alt=${`Live view from ${name}’s camera`} src=${`/cam-stream?serial=${encodeURIComponent(esn)}`} onLoad=${() => setCamPhase("live")} onError=${() => stopCam("error")} style=${camPhase === "live" ? "" : "visibility:hidden;position:absolute"} />`
            : null}
          ${camPhase === "loading" && html`<div role="status" aria-live="polite" class="cam-note"><span class="spinner"></span><span>Waiting for ${name}’s camera. This can take about 10 seconds.</span></div>`}
          ${!cam && html`<${Icon} name="camera" size=${30} /><span role=${camPhase === "off" ? null : "alert"}>${camPhase === "error" ? `The camera didn’t start. Check that ${name} is on and try again.` : camPhase === "timeout" ? `${name}’s camera didn’t respond. Check that ${name} is on and on the same network, then try again.` : "Camera is off"}</span>`}
        </div>
      </div>
      <div class="card">
        <h2 class="h3">Quick actions</h2><p class="sub">Send ${name} off to do something.</p>
        <div class="chips">
          ${QUICK.map((q) => html`<button type="button" class="chip" key=${q.label} onClick=${() => quick(q)}>${q.icon && html`<${Icon} name=${q.icon} size=${16} />`}${q.label}</button>`)}
          <button type="button" class="chip" onClick=${heyVector}>Say “Hey Vector”</button>
        </div>
      </div>
      <div class="card">
        <h2 class="h3">Say something</h2>
        <form class="addrow" style="margin-top:8px" onSubmit=${doSay}>
          <input class="inp" id="say-text" aria-label=${`What ${name} should say`} placeholder=${`Type what ${name} should say`} value=${say} maxlength="590" onInput=${(e) => setSay(e.target.value)} />
          <button type="submit" class="btn pri" disabled=${!say.trim()}><${Icon} name="send" size=${18} />Say it</button>
        </form>
        <div class="opt" style="margin-top:6px"><div>Play a sound<small>${sound ? sound.label : "Choose an audio file (WAV works best)"}</small></div>
          <div class="chips">
            <label class="btn" style="margin:0;cursor:pointer">${soundBusy ? "Converting…" : "Choose file"}<input type="file" accept="audio/*,.wav" class="sr-only" onChange=${chooseSound} /></label>
            ${sound && html`<button type="button" class="btn pri" onClick=${playSound}><${Icon} name="play" size=${18} />Play on ${name}</button>`}
          </div></div>
        ${sound && html`<audio controls src=${sound.url} style="width:100%;margin-top:8px"></audio>`}
      </div>
    </div>
    <div class="stack">
      <div class="card">
        <h2 class="h3">Drive</h2>
        <div class="opt" style="padding-top:6px"><div id="ctl-take">Take control<small>${controlled ? `You’re in control. It’s given back when you leave this page.` : "Taken automatically when you drive or speak"}</small></div>
          <${Switch} checked=${controlled} onChange=${toggleControl} label="Take control" /></div>
        <p class="sub" style="font-size:.78rem;margin:6px 0 10px">Hold a button to move. Let go to stop.</p>
        <div class="dpad">
          <span></span>${pad("Forward", "up", 140, 140)}<span></span>
          ${pad("Turn left", "left", -150, 150)}
          <button type="button" class="c" aria-label="Stop" onClick=${drive.stop}><${Icon} name="x" size=${20} /></button>
          ${pad("Turn right", "right", 150, -150)}
          <span></span>${pad("Back", "down", -150, -150)}<span></span>
        </div>
        <div class="holdrow" style="margin-top:12px"><span class="lbl" style="margin:0">Lift</span><div class="chips">
          <${HoldButton} label="Lift up" className="btn" onStart=${() => lift.start(2)} onStop=${lift.stop}><${Icon} name="up" size=${18} />Up<//>
          <${HoldButton} label="Lift down" className="btn" onStart=${() => lift.start(-2)} onStop=${lift.stop}><${Icon} name="down" size=${18} />Down<//></div></div>
        <div class="holdrow"><span class="lbl" style="margin:0">Head</span><div class="chips">
          <${HoldButton} label="Head up" className="btn" onStart=${() => head.start(2)} onStop=${head.stop}><${Icon} name="up" size=${18} />Up<//>
          <${HoldButton} label="Head down" className="btn" onStart=${() => head.start(-2)} onStop=${head.stop}><${Icon} name="down" size=${18} />Down<//></div></div>
        <div class="opt"><div>Keyboard control<small><span class="kbd">W</span><span class="kbd">A</span><span class="kbd">S</span><span class="kbd">D</span> drive, <span class="kbd">R</span><span class="kbd">F</span> lift, <span class="kbd">T</span><span class="kbd">G</span> head</small></div>
          <${Switch} checked=${kb} onChange=${async (on) => { if (on) { try { await ensureControl(); setKb(true); } catch (e) { fail(e); } } else setKb(false); }} label="Keyboard control" /></div>
        <div class="opt"><div>Mirror mode<small>${name} shows what its camera sees</small></div>
          <${Switch} checked=${mirror} onChange=${toggleMirror} label="Mirror mode" /></div>
      </div>
    </div>
  </div>`;
}
