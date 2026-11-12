// Robot > Personality: the name and personality wire-pod gives the AI for this robot.
import { html, useState, useEffect, useRef } from "../html.js";
import { postForm, describeError } from "../api.js";
import { refreshRobots, robotName, toast } from "../store.js";
import { PERSONALITY_PRESETS } from "../data.js";

const NAME_MAX = 32;
const PERSONALITY_MAX = 1000;

export function Personality({ robot }) {
  const name = robotName(robot);
  const [form, setForm] = useState({ name: robot.name || "", personality: robot.personality || "" });
  const saved = useRef({ name: robot.name || "", personality: robot.personality || "" });
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const dirty = form.name !== saved.current.name || form.personality !== saved.current.personality;

  // warn before closing the tab with unsaved edits
  useEffect(() => {
    if (!dirty) return;
    const f = (e) => { e.preventDefault(); e.returnValue = ""; };
    window.addEventListener("beforeunload", f);
    return () => window.removeEventListener("beforeunload", f);
  }, [dirty]);

  const save = async (e) => {
    e && e.preventDefault();
    setBusy(true);
    setError("");
    try {
      await postForm("/api-sdk/set_robot_identity", { serial: robot.esn, name: form.name, personality: form.personality });
      saved.current = { name: form.name.trim().replace(/\s+/g, " "), personality: form.personality.trim() };
      setForm({ ...saved.current });
      toast("Saved");
      refreshRobots();
    } catch (err) {
      setError(describeError(err, name).message);
    } finally {
      setBusy(false);
    }
  };

  const usePreset = (text) => {
    const before = form.personality;
    setForm({ ...form, personality: text });
    if (before.trim() && before !== text) {
      toast("Personality replaced", { undo: () => setForm((f) => ({ ...f, personality: before })) });
    }
  };

  const shownName = form.name.trim() || "Vector";
  const told = [
    form.name.trim() && `Your name is ${form.name.trim().replace(/\s+/g, " ")}. This name overrides any other name given in the instructions below.`,
    form.personality.trim() && `Your personality and identity: ${form.personality.trim()}`,
  ].filter(Boolean).join(" ");

  return html`<form onSubmit=${save} class="stack">
    <div class="grid2">
      <div class="card"><div class="stack">
        <div>
          <label for="p-name">Name</label>
          <input class="inp" id="p-name" value=${form.name} maxlength=${NAME_MAX} autocomplete="off" aria-describedby="p-name-hint"
            onInput=${(e) => setForm({ ...form, name: e.target.value })} />
          <div class="hint" id="p-name-hint"><span>Shown in WirePod and used by the AI. It does not rename the robot itself.</span><span>${form.name.length} / ${NAME_MAX}</span></div>
        </div>
        <div>
          <label for="p-pers">Personality</label>
          <textarea class="inp" id="p-pers" rows="5" maxlength=${PERSONALITY_MAX} aria-describedby="p-pers-hint"
            placeholder="Example: Grumpy but secretly kind, loves astronomy, speaks like a pirate"
            value=${form.personality} onInput=${(e) => setForm({ ...form, personality: e.target.value })}></textarea>
          <div class="hint" id="p-pers-hint"><span>How ${shownName} talks and what it cares about.</span><span>${form.personality.length} / ${PERSONALITY_MAX}</span></div>
        </div>
        <div role="group" aria-labelledby="p-start">
          <span class="lbl" id="p-start">Start from</span>
          <div class="chips">${PERSONALITY_PRESETS.map(([label, text]) => html`<button type="button" class="chip" key=${label} onClick=${() => usePreset(text)}>${label}</button>`)}</div>
        </div>
      </div></div>
      <div class="card">
        <h2 class="h3">What the AI is told</h2>
        <p class="sub">This is added in front of your AI instructions every time ${shownName} answers.</p>
        <div class="promptview" aria-live="polite">${told || "Nothing yet. Without a name or personality, the AI uses only your general instructions."}</div>
        <p class="hint" style="margin-top:12px">${shownName}’s memories are added too. The AI itself is set in Settings.</p>
      </div>
    </div>
    ${error && html`<p class="err" role="alert">${error}</p>`}
    ${dirty && html`<div class="savebar" role="region" aria-label="Unsaved changes">
      <span><span class="dot warn"></span>Unsaved changes</span>
      <div class="b">
        <button type="button" class="btn quiet" onClick=${() => { setForm({ ...saved.current }); setError(""); }} disabled=${busy}>Discard</button>
        <button type="submit" class="btn pri" disabled=${busy}>${busy ? "Saving…" : "Save changes"}</button>
      </div>
    </div>`}
  </form>`;
}
