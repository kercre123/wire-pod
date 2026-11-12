// Robot > Memories: what the robot remembers (kept by wire-pod, so it works while the robot is off).
import { html, useState, useEffect, useRef } from "../html.js";
import { getJSON, postForm, describeError } from "../api.js";
import { robotName, toast } from "../store.js";
import { Icon } from "../icons.js";
import { Loading, ErrorNote } from "../ui.js";

const MAX_TEXT = 200;
const MAX_COUNT = 50;
// Deleting hides the item and offers Undo; the server delete happens when the toast is gone.
const UNDO_MS = 7200;

const fmtDate = (s) => {
  const d = new Date(s);
  return isNaN(d) ? "" : d.toLocaleDateString(undefined, { month: "short", day: "numeric" });
};

export function Memories({ robot }) {
  const name = robotName(robot);
  const esn = robot.esn;
  const [items, setItems] = useState(null);
  const [error, setError] = useState(null);
  const [hidden, setHidden] = useState([]); // ids waiting on their undo window
  const [text, setText] = useState("");
  const [addError, setAddError] = useState("");
  const [adding, setAdding] = useState(false);
  const pending = useRef(null); // { ids, all, timer }

  const load = async () => {
    try {
      setItems(await getJSON(`/api-sdk/get_robot_memories?serial=${encodeURIComponent(esn)}`));
      setError(null);
    } catch (e) {
      setError(describeError(e, name));
    }
  };
  useEffect(() => { load(); }, [esn]);

  const commit = async (keepalive = false) => {
    const p = pending.current;
    if (!p) return;
    pending.current = null;
    clearTimeout(p.timer);
    try {
      if (p.all) await postForm("/api-sdk/clear_robot_memories", { serial: esn }, { keepalive });
      else for (const id of p.ids) await postForm("/api-sdk/delete_robot_memory", { serial: esn, id }, { keepalive });
    } catch (e) {
      toast(describeError(e, name).message, { kind: "bad" });
    }
    if (!keepalive) { setHidden(pending.current ? pending.current.ids : []); load(); }
  };
  // leaving the page or closing the tab commits what was deleted
  useEffect(() => {
    const f = () => commit(true);
    window.addEventListener("pagehide", f);
    return () => { window.removeEventListener("pagehide", f); commit(true); };
  }, [esn]);

  const startDelete = (ids, all, message) => {
    if (pending.current) commit(); // an earlier delete is final once another begins
    const timer = setTimeout(() => commit(), UNDO_MS);
    pending.current = { ids, all, timer };
    setHidden(ids);
    toast(message, {
      undo: () => {
        if (!pending.current) return;
        clearTimeout(pending.current.timer);
        pending.current = null;
        setHidden([]);
      },
    });
  };

  const add = async (e) => {
    e.preventDefault();
    const t = text.trim();
    if (!t) { setAddError("Type something for " + name + " to remember."); return; }
    setAdding(true);
    setAddError("");
    try {
      await postForm("/api-sdk/add_robot_memory", { serial: esn, text: t });
      setText("");
      await load();
    } catch (err) {
      setAddError(describeError(err, name).message);
    } finally {
      setAdding(false);
    }
  };

  const visible = (items || []).filter((m) => !hidden.includes(m.id));
  const full = (items || []).length >= MAX_COUNT;
  return html`<div class="card">
    <h2>What ${name} remembers</h2>
    <p class="sub">${name} brings these up in conversation (and can add its own when AI commands are on). Only ${name} knows them.</p>
    <form class="addrow" onSubmit=${add}>
      <input class="inp" aria-label="New memory" aria-invalid=${addError ? "true" : "false"} maxlength=${MAX_TEXT} value=${text} disabled=${full}
        placeholder=${full ? `${name} can hold ${MAX_COUNT} memories. Delete one to add more.` : `Teach ${name} something, like “My dog is named Biscuit”`}
        onInput=${(e) => { setText(e.target.value); setAddError(""); }} />
      <button type="submit" class="btn pri" disabled=${adding || full}><${Icon} name="plus" size=${18} />Add</button>
    </form>
    ${addError && html`<p class="err" role="alert">${addError}</p>`}
    ${text.length > MAX_TEXT - 40 && html`<p class="hint"><span></span><span>${text.length} / ${MAX_TEXT}</span></p>`}
    ${error && html`<${ErrorNote} error=${error} onRetry=${load} />`}
    ${items === null && !error && html`<${Loading} />`}
    ${items !== null && html`<div class="list" role="list">
      ${visible.map((m) => html`<div class="item" role="listitem" key=${m.id}>
        <div class="t">${m.text}<small>Added ${fmtDate(m.created)}</small></div>
        <button type="button" class="btn icon quiet" aria-label=${`Delete memory: ${m.text}`} onClick=${() => startDelete([m.id], false, "Memory deleted")}><${Icon} name="trash" size=${18} /></button>
      </div>`)}
    </div>`}
    ${items !== null && visible.length === 0 && html`<p class="empty">Nothing yet. Add something above and ${name} will bring it up in conversation.</p>`}
    ${items !== null && html`<div class="dangerzone">
      <span class="sub" style="font-size:.78rem">${visible.length} of ${MAX_COUNT} memories</span>
      ${visible.length > 0 && html`<button type="button" class="btn danger" onClick=${() => startDelete(visible.map((m) => m.id), true, "All memories cleared")}>Clear all</button>`}
    </div>`}
  </div>`;
}
