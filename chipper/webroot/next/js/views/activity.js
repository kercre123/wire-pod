// Robot > Activity: conversations, lifetime stats, mood, photos and a log filtered to this robot.
import { html, useState, useEffect, useRef } from "../html.js";
import { Icon } from "../icons.js";
import { getJSON, getText, sdk, sdkJSON, describeError } from "../api.js";
import { robotName, toast, confirmDialog, refreshRobots, cacheGet, cachePut } from "../store.js";
import { Loading, ErrorNote, OfflineNotice, Dialog, Switch } from "../ui.js";
import { linesFor } from "../logs.js";
import { href } from "../router.js";

// ---------- conversations ----------

function Conversations({ robot }) {
  const name = robotName(robot);
  const [lines, setLines] = useState(() => cacheGet(`chats:${robot.esn}`) || null);
  useEffect(() => { if (lines) cachePut(`chats:${robot.esn}`, lines); }, [lines]);
  const [error, setError] = useState(null);
  const [showLog, setShowLog] = useState(false);
  const reads = useRef(0); // only the newest read may apply
  const load = async () => {
    const id = ++reads.current;
    try { const v = await getJSON(`/api-sdk/get_robot_chats?serial=${encodeURIComponent(robot.esn)}`); if (id !== reads.current) return; setLines(v); setError(null); }
    catch (e) { setError(describeError(e, name)); }
  };
  useEffect(() => {
    load();
    const t = setInterval(() => { if (!document.hidden) load(); }, 10000);
    return () => clearInterval(t);
  }, [robot.esn]);
  const deleteAll = async () => {
    const ok = await confirmDialog({ title: "Delete saved conversations?", body: "This removes the saved conversations of every robot on this WirePod, and what they remember from them in chat. It can’t be undone. Memories are kept.", confirmLabel: "Delete", danger: true });
    if (!ok) return;
    try { await getText("/api/delete_chats"); toast("Saved conversations deleted"); load(); } catch (e) { toast(describeError(e).message, { kind: "bad" }); }
  };
  return html`<div class="card">
    <h2 class="h3">Recent conversations</h2>
    <p class="sub">Saved on this computer while “Save conversations” is on. <a href=${href.settings("ai")}>Change in Settings</a></p>
    ${error && html`<${ErrorNote} error=${error} onRetry=${load} />`}
    ${lines === null && !error && html`<${Loading} />`}
    ${lines && lines.length === 0 && html`<p class="empty">Nothing saved yet. Conversations appear here after ${name} answers a question with an AI provider and saving is on.</p>`}
    ${lines && lines.length > 0 && html`<div class="chat" role="log" aria-label=${`Recent conversation with ${name}`}>
      ${lines.map((l, i) => html`<div key=${i} class=${"bubble " + (l.role === "user" ? "u" : "r")}><span class="sr-only">${l.role === "user" ? "You: " : `${name}: `}</span>${l.content}</div>`)}
    </div>`}
    <div class="dangerzone" style="margin-top:12px">
      <button type="button" class="btn quiet" aria-expanded=${showLog ? "true" : "false"} onClick=${() => setShowLog(!showLog)}>${showLog ? "Hide" : "View"} technical log <${Icon} name=${showLog ? "down" : "right"} size=${16} /></button>
      <button type="button" class="btn danger" onClick=${deleteAll}>Delete saved conversations</button>
    </div>
    ${showLog && html`<${TechLog} esn=${robot.esn} label=${name} />`}
  </div>`;
}

// esn: only that robot's lines; without it, everything. Also used by Settings > Server.
export function TechLog({ esn, label }) {
  const [all, setAll] = useState(false);
  const [text, setText] = useState(null);
  const box = useRef(null);
  useEffect(() => {
    let alive = true, timer;
    const run = async () => {
      if (!document.hidden) {
        try { const t = await getText(all ? "/api/get_debug_logs" : "/api/get_logs", { timeout: 6000 }); if (alive) setText(t); } catch (e) { if (alive) setText(""); }
      }
      if (alive) timer = setTimeout(run, 1500);
    };
    run();
    return () => { alive = false; clearTimeout(timer); };
  }, [all]);
  const shown = text === null ? null : all || !esn ? text : linesFor(text, esn);
  useEffect(() => { if (box.current) box.current.scrollTop = box.current.scrollHeight; }, [shown]);
  return html`<div style="margin-top:10px">
    <div class="opt flush" style="margin-bottom:8px"><div>${esn ? "Show everything" : "Show debug lines"}<small>${esn ? `All robots and debug lines, not just ${label}` : "Include the detailed lines that are normally hidden"}</small></div><${Switch} checked=${all} onChange=${setAll} label=${esn ? "Show everything" : "Show debug lines"} /></div>
    <pre class="log" ref=${box} tabindex="0" aria-label="Log">${shown === null ? "Loading…" : shown || (esn ? "Nothing logged for this robot yet. Say something to it and this updates by itself." : "Nothing logged yet. Say something to a robot and this updates by itself.")}</pre>
  </div>`;
}

// ---------- stats ----------

function Stats({ robot, photoCount }) {
  const name = robotName(robot);
  const [stats, setStats] = useState(() => cacheGet(`stats:${robot.esn}`) || null);
  useEffect(() => { if (stats) cachePut(`stats:${robot.esn}`, stats); }, [stats]);
  const [error, setError] = useState(null);
  const reads = useRef(0);
  const load = async () => {
    const id = ++reads.current;
    try { const v = await sdkJSON("get_robot_stats", robot.esn, {}, { timeout: 20000 }); if (id !== reads.current) return; setStats(v); setError(null); }
    catch (e) { setError(describeError(e, name)); }
  };
  useEffect(() => { load(); }, [robot.esn]);
  const n = (k) => (stats && typeof stats[k] === "number" ? stats[k] : 0);
  const tiles = stats && [
    [(n("Alive.seconds") / 86400).toFixed(n("Alive.seconds") >= 864000 ? 0 : 1), "Days alive"],
    [n("BStat.ReactedToTriggerWord"), "Times heard “Hey Vector”"],
    [n("FeatureType.Utility"), "Utility features used"],
    [Math.round(n("Pet.ms") / 60000), "Minutes petted"],
    [(n("Stim.CumlPosDelta") / 10000).toFixed(1), "Meters moved"],
    photoCount !== null && [photoCount, "Photos"],
  ].filter(Boolean);
  return html`<div class="card">
    <div class="top"><h2 class="h3">Lifetime</h2><button type="button" class="btn quiet small" onClick=${load}><${Icon} name="refresh" size=${16} />Refresh</button></div>
    ${error && html`<${ErrorNote} error=${error} onRetry=${load} />`}
    ${!stats && !error && html`<${Loading} />`}
    ${tiles && html`<div class="stats" style="margin-top:8px">${tiles.map(([v, l]) => html`<div key=${l}><b>${v}</b><span>${l}</span></div>`)}</div>`}
  </div>`;
}

// ---------- mood (stimulation) ----------

function Mood({ robot }) {
  const name = robotName(robot);
  const [points, setPoints] = useState([]);
  const [failed, setFailed] = useState(false);
  useEffect(() => {
    let alive = true, timer, started = false;
    const poll = async () => {
      if (!alive) return;
      if (!document.hidden) {
        try {
          const t = await sdk("get_stim_status", robot.esn, {}, { timeout: 5000 });
          const v = parseFloat(t);
          if (!isNaN(v) && alive) setPoints((p) => [...p, v].slice(-60));
          setFailed(false);
        } catch (e) { if (alive) setFailed(true); }
      }
      if (alive) timer = setTimeout(poll, 500);
    };
    sdk("begin_event_stream", robot.esn).then(() => { started = true; poll(); }, () => alive && setFailed(true));
    return () => { alive = false; clearTimeout(timer); sdk("stop_event_stream", robot.esn).catch(() => {}); };
  }, [robot.esn]);
  const last = points.length ? points[points.length - 1] : null;
  const path = points.map((v, i) => `${(i / Math.max(1, 59)) * 240},${54 - Math.max(0, Math.min(1, v)) * 48}`).join(" ");
  return html`<div class="card">
    <h2 class="h3">Mood</h2><p class="sub">How stimulated ${name} is right now, over the last half minute.</p>
    ${failed && points.length === 0 && html`<p class="empty">Couldn’t read ${name}’s mood just now.</p>`}
    ${points.length > 0 && html`<svg viewBox="0 0 240 56" width="100%" height="64" role="img" aria-label=${`Mood is ${Math.round((last || 0) * 100)} percent`} preserveAspectRatio="none">
      <polyline points=${path} fill="none" stroke="var(--accent)" stroke-width="2.5" stroke-linejoin="round" stroke-linecap="round" vector-effect="non-scaling-stroke" /></svg>
      <p class="hint"><span>Calm</span><span><b>${Math.round((last || 0) * 100)}%</b></span><span>Excited</span></p>`}
    ${points.length === 0 && !failed && html`<${Loading}>Listening…<//>`}
  </div>`;
}

// ---------- photos ----------

function Photos({ robot, onCount }) {
  const name = robotName(robot);
  const esn = robot.esn;
  const [ids, setIds] = useState(() => cacheGet(`photos:${esn}`) || null);
  useEffect(() => { if (ids) cachePut(`photos:${esn}`, ids); }, [ids]);
  const [error, setError] = useState(null);
  const [open, setOpen] = useState(null);
  const reads = useRef(0);
  const load = async () => {
    const id = ++reads.current;
    try {
      const t = await sdk("get_image_ids", esn, {}, { timeout: 25000 });
      const v = t.trim() === "null" || !t.trim() ? [] : JSON.parse(t);
      if (id !== reads.current) return;
      setIds(v); onCount(v.length); setError(null);
    } catch (e) { setError(describeError(e, name)); }
  };
  useEffect(() => { load(); }, [esn]);
  const del = async (id) => {
    const ok = await confirmDialog({ title: "Delete this photo?", body: `It’s removed from ${name} and can’t be brought back.`, confirmLabel: "Delete", danger: true });
    if (!ok) return;
    try { await sdk("delete_image", esn, { id }); setOpen(null); toast("Photo deleted"); load(); } catch (e) { toast(describeError(e, name).message, { kind: "bad" }); }
  };
  const src = (kind, id) => `/api-sdk/${kind}?serial=${encodeURIComponent(esn)}&id=${id}`;
  return html`<div class="card">
    <div class="top"><h2 class="h3">Photos</h2><button type="button" class="btn quiet small" onClick=${load}><${Icon} name="refresh" size=${16} />Refresh</button></div>
    ${error && html`<${ErrorNote} error=${error} onRetry=${load} />`}
    ${ids === null && !error && html`<${Loading} />`}
    ${ids && ids.length === 0 && html`<p class="empty">No photos yet. Tell ${name} to take a photo, then refresh.</p>`}
    ${ids && ids.length > 0 && html`<div class="photos" style="margin-top:10px">${ids.map((id) => html`<button type="button" key=${id} aria-label=${`Open photo ${id}`} onClick=${() => setOpen(id)}><img src=${src("get_image_thumb", id)} alt="" loading="lazy" /></button>`)}</div>`}
    ${open !== null && html`<${Dialog} title=${`Photo ${open}`} className="photo" onClose=${() => setOpen(null)}>
      <img src=${src("get_image", open)} alt=${`Photo ${open} taken by ${name}`} />
      <div class="foot"><a class="btn" href=${src("get_image", open)} target="_blank" rel="noopener">Open original</a><button type="button" class="btn danger" onClick=${() => del(open)}>Delete</button><button type="button" class="btn pri" onClick=${() => setOpen(null)}>Close</button></div>
    <//>`}
  </div>`;
}

export function Activity({ robot }) {
  const [photoCount, setPhotoCount] = useState(null);
  const online = robot.state === "online";
  return html`<div class="grid2">
    <div class="stack"><${Conversations} robot=${robot} /></div>
    <div class="stack">
      ${online ? html`<${Stats} robot=${robot} photoCount=${photoCount} /><${Mood} robot=${robot} /><${Photos} robot=${robot} onCount=${setPhotoCount} />`
        : html`<${OfflineNotice} robot=${robot} onRetry=${refreshRobots} what="stats, mood and photos" />`}
    </div>
  </div>`;
}
