// People a robot knows (its enrolled faces). These live on the robot, so it has to be online.
import { html, useState, useEffect, useRef } from "../html.js";
import { Icon } from "../icons.js";
import { sdk, describeError } from "../api.js";
import { robotName, toast, confirmDialog } from "../store.js";
import { Loading, ErrorNote, Dialog, Callout } from "../ui.js";

function parseFaces(text) {
  const t = text.trim();
  if (!t || t === "null") return [];
  const v = JSON.parse(t);
  return Array.isArray(v) ? v : [];
}

function RenameDialog({ face, onSave, onClose }) {
  const [v, setV] = useState(face.name);
  const ref = useRef(null);
  useEffect(() => { if (ref.current) { ref.current.focus(); ref.current.select(); } }, []);
  return html`<${Dialog} title=${`Rename ${face.name}`} onClose=${onClose}>
    <form onSubmit=${(e) => { e.preventDefault(); if (v.trim()) onSave(v.trim()); }} class="stack">
      <div><label for="face-new">New name</label><input class="inp" id="face-new" ref=${ref} value=${v} maxlength="40" onInput=${(e) => setV(e.target.value)} /></div>
      <div class="foot"><button type="button" class="btn quiet" onClick=${onClose}>Cancel</button><button type="submit" class="btn pri" disabled=${!v.trim()}>Rename</button></div>
    </form>
  <//>`;
}

export function People({ robot }) {
  const name = robotName(robot);
  const esn = robot.esn;
  const online = robot.state === "online";
  const [faces, setFaces] = useState(null);
  const [error, setError] = useState(null);
  const [newName, setNewName] = useState("");
  const [rename, setRename] = useState(null);
  const [busy, setBusy] = useState(false);

  const load = async () => {
    try {
      setFaces(parseFaces(await sdk("get_faces", esn, {}, { timeout: 20000 })));
      setError(null);
    } catch (e) {
      setError(describeError(e, name));
    }
  };
  useEffect(() => { if (online) load(); }, [esn, online]);

  const run = async (fn, message) => {
    setBusy(true);
    try { await fn(); if (message) toast(message); } catch (e) { toast(describeError(e, name).message, { kind: "bad" }); }
    setBusy(false);
  };
  const add = (e) => {
    e.preventDefault();
    const n = newName.trim();
    if (!n) return;
    run(async () => {
      await sdk("add_face", esn, { name: n });
      setNewName("");
      // the robot scans when it sees a face; check back a few times
      [4000, 12000, 25000].forEach((ms) => setTimeout(load, ms));
    }, `Now show your face to ${name}`);
  };
  const remove = async (f) => {
    const ok = await confirmDialog({ title: `Remove ${f.name}?`, body: `${name} will no longer recognize ${f.name}. To bring them back, add them again and show their face to ${name}.`, confirmLabel: "Remove", danger: true });
    if (ok) run(async () => { await sdk("delete_face", esn, { id: f.face_id }); await load(); }, `${f.name} removed`);
  };

  return html`<div class="card">
    <h2>People ${name} knows</h2>
    <p class="sub">${name} learns the faces it sees. Name them here.</p>
    ${!online && html`<${Callout} icon="info"><span>${name} has to be online to show or change the people it knows.</span><//>`}
    ${online && error && html`<${ErrorNote} error=${error} onRetry=${load} />`}
    ${online && faces === null && !error && html`<${Loading} />`}
    ${online && faces && html`<div class="list">
      ${faces.map((f) => html`<div class="item" key=${f.face_id}>
        <span class="avatar" aria-hidden="true">${(f.name || "?")[0].toUpperCase()}</span><div class="t">${f.name}</div>
        <button type="button" class="btn quiet" onClick=${() => setRename(f)} aria-label=${`Rename ${f.name}`}>Rename</button>
        <button type="button" class="btn icon quiet" aria-label=${`Remove ${f.name}`} onClick=${() => remove(f)} disabled=${busy}><${Icon} name="x" size=${18} /></button>
      </div>`)}
    </div>`}
    ${online && faces && faces.length === 0 && html`<p class="empty">No faces yet. Add a name below, then show your face to ${name}.</p>`}
    ${online && html`<form class="addrow" style="margin-top:12px" onSubmit=${add}>
      <input class="inp" aria-label="Name for a new face" placeholder="New name" value=${newName} maxlength="40" onInput=${(e) => setNewName(e.target.value)} />
      <button type="submit" class="btn" disabled=${busy || !newName.trim()}>Add face</button>
    </form>`}
    ${online && html`<p class="hint">Type a name, then show your face to ${name}.</p>`}
    ${rename && html`<${RenameDialog} face=${rename} onClose=${() => setRename(null)}
      onSave=${(n) => { const f = rename; setRename(null); run(async () => { await sdk("rename_face", esn, { id: f.face_id, oldname: f.name, newname: n }); await load(); }, "Renamed"); }} />`}
  </div>`;
}
