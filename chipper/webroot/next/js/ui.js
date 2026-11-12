// Shared building blocks.
import { html, useEffect, useRef, useState } from "./html.js";
import { Icon } from "./icons.js";
import { toastStore, useStore, dismissToast, confirmStore, robotName } from "./store.js";
import { eyeColor } from "./theme.js";

let uidCounter = 0;
export function useUid(prefix = "f") {
  const ref = useRef(null);
  if (ref.current === null) ref.current = `${prefix}${++uidCounter}`;
  return ref.current;
}

export function Face({ robot, size = 56 }) {
  const off = robot && robot.state === "offline";
  return html`<span class=${"face" + (off ? " off" : "")} style=${`--eye:${eyeColor(robot && robot.eye)};--fs:${size}px`} aria-hidden="true"><i></i><i></i></span>`;
}

export function StatusPill({ robot }) {
  if (robot.state === "online") return html`<span class="pill ok"><span class="dot"></span>Online</span>`;
  if (robot.state === "checking") return html`<span class="pill"><span class="dot mute"></span>Checking</span>`;
  return html`<span class="pill bad"><span class="dot bad"></span>Can’t reach</span>`;
}

export function Battery({ battery }) {
  if (!battery) return html`<div class="bat"><span class="cell none"></span><span>Battery unknown</span></div>`;
  const cls = battery.percent < 20 ? " low" : battery.percent < 50 ? " mid" : "";
  return html`<div class="bat">
    <span class=${"cell" + cls}><i style=${`width:${battery.percent}%`}></i></span>
    <b>${battery.percent}%</b>
    <span>${battery.charging ? "Charging" : battery.on_charger ? "On charger" : "On battery"}</span>
  </div>`;
}

export function Switch({ checked, onChange, label, disabled }) {
  return html`<span class="sw-wrap"><button type="button" class="sw" role="switch" aria-checked=${checked ? "true" : "false"}
    aria-label=${label} disabled=${disabled} onClick=${() => onChange(!checked)}></button></span>`;
}

// options: [[value, label], ...]
export function Segmented({ options, value, onChange, label, disabled }) {
  return html`<div class="sg" role="group" aria-label=${label}>
    ${options.map(([v, text]) => html`<button type="button" key=${v} aria-pressed=${v === value ? "true" : "false"} disabled=${disabled}
      onClick=${() => onChange(v)}>${text}</button>`)}
  </div>`;
}

export function Callout({ icon = "info", kind = "", children }) {
  return html`<div class=${"callout " + kind} role=${kind === "bad" ? "alert" : null}><${Icon} name=${icon} size=${20} /><div class="grow">${children}</div></div>`;
}

export function Loading({ children = "Loading…" }) {
  return html`<div class="loading" role="status"><span class="spinner"></span><span>${children}</span></div>`;
}

// Shown on robot sections that need the robot itself, when it can't be reached.
export function OfflineNotice({ robot, onRetry, what = "this" }) {
  const name = robotName(robot);
  return html`<div class="card offline-note">
    <${Callout} icon="info">
      <b>${name} isn’t answering.</b>
      <span>Check that it’s powered on and on the same Wi-Fi as this computer. Then try again. You can still edit ${name}’s personality and memories while it’s offline, so ${what} will be here when it’s back.</span>
    <//>
    ${onRetry && html`<button type="button" class="btn" onClick=${onRetry}><${Icon} name="refresh" size=${18} />Check again</button>`}
  </div>`;
}

// A failed action: a human sentence plus the raw error tucked away.
export function ErrorNote({ error, onRetry }) {
  if (!error || !error.message) return null;
  return html`<${Callout} icon="alert" kind="warn">
    <span>${error.message}</span>
    ${onRetry && html`<button type="button" class="btn small" onClick=${onRetry}>Try again</button>`}
    ${error.detail && error.detail !== error.message && html`<details><summary>Technical details</summary><pre>${error.detail}</pre></details>`}
  <//>`;
}

export function Dialog({ title, onClose, children, wide, className = "" }) {
  const ref = useRef(null);
  const titleId = useUid("dlg");
  useEffect(() => {
    const d = ref.current;
    if (d && !d.open) d.showModal();
    return () => {
      if (d && d.open) d.close();
    };
  }, []);
  return html`<dialog ref=${ref} class=${(wide ? "wide " : "") + className} aria-labelledby=${titleId}
    onClose=${onClose} onClick=${(e) => { if (e.target === ref.current) onClose(); }}>
    <div class="dlg"><h2 id=${titleId}>${title}</h2>${children}</div>
  </dialog>`;
}

export function Toasts() {
  const { items } = useStore(toastStore);
  return html`<div class="toasts" role="status" aria-live="polite" aria-atomic="false">
    ${items.map((t) => html`<div class=${"toast " + (t.kind === "bad" ? "bad" : "")} key=${t.id}>
      <span>${t.message}</span>
      ${t.undo && html`<button type="button" onClick=${() => { dismissToast(t.id); t.undo(); }}>Undo</button>`}
      ${!t.undo && t.kind === "bad" && html`<button type="button" aria-label="Dismiss" onClick=${() => dismissToast(t.id)}>OK</button>`}
    </div>`)}
  </div>`;
}

export function ConfirmHost() {
  const { current } = useStore(confirmStore);
  if (!current) return null;
  const done = (v) => {
    confirmStore.set({ current: null });
    current.resolve(v);
  };
  return html`<${Dialog} title=${current.title} onClose=${() => done(false)}>
    ${current.body && html`<p class="sub">${current.body}</p>`}
    <div class="foot">
      <button type="button" class="btn quiet" onClick=${() => done(false)}>Cancel</button>
      <button type="button" class=${"btn " + (current.danger ? "danger fill" : "pri")} onClick=${() => done(true)}>${current.confirmLabel || "OK"}</button>
    </div>
  <//>`;
}

// A text input with a visible show/hide button, for API keys and passwords.
export function SecretInput({ id, value, onInput, placeholder }) {
  const [show, setShow] = useState(false);
  return html`<div class="with-btn">
    <input class="inp" id=${id} type=${show ? "text" : "password"} value=${value} placeholder=${placeholder} autocomplete="off" spellcheck="false"
      onInput=${(e) => onInput(e.target.value)} />
    <button type="button" class="btn line" aria-pressed=${show ? "true" : "false"} onClick=${() => setShow(!show)}>${show ? "Hide" : "Show"}</button>
  </div>`;
}

// Hold-to-act button: pointer down starts, up/leave/cancel stops. Space and Enter work too.
export function HoldButton({ onStart, onStop, label, children, className = "btn", disabled }) {
  const active = useRef(false);
  const start = (e) => {
    if (active.current || disabled) return;
    active.current = true;
    onStart(e);
  };
  const stop = (e) => {
    if (!active.current) return;
    active.current = false;
    onStop(e);
  };
  useEffect(() => () => { if (active.current) onStop(); }, []);
  return html`<button type="button" class=${className} aria-label=${label} disabled=${disabled}
    onPointerDown=${(e) => { e.preventDefault(); e.currentTarget.setPointerCapture && e.currentTarget.setPointerCapture(e.pointerId); start(e); }}
    onPointerUp=${stop} onPointerCancel=${stop} onLostPointerCapture=${stop} onBlur=${stop}
    onKeyDown=${(e) => { if ((e.key === " " || e.key === "Enter") && !e.repeat) { e.preventDefault(); start(e); } }}
    onKeyUp=${(e) => { if (e.key === " " || e.key === "Enter") { e.preventDefault(); stop(e); } }}
    onContextMenu=${(e) => e.preventDefault()}>${children}</button>`;
}

export function pageTitle(title) {
  document.title = title ? `${title} · WirePod` : "WirePod";
}
